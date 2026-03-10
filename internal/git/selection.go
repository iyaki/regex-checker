package git

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
)

const windowsAbsoluteMinLength = 3

// CandidateSelectionRequest defines inputs for selecting Git-scoped candidates.
type CandidateSelectionRequest struct {
	Mode       string
	DiffTarget string
	WorkingDir string
}

// SelectCandidateFiles resolves deterministic root-relative candidate files.
func SelectCandidateFiles(request CandidateSelectionRequest) ([]string, error) {
	mode := strings.TrimSpace(request.Mode)
	if mode == "" || mode == "off" {
		return nil, nil
	}

	var (
		output string
		err    error
	)

	switch mode {
	case "staged":
		output, err = runCommand(request.WorkingDir, "diff", "--cached", "--name-only", "--diff-filter=ACMR", "-z")
		if err != nil {
			return nil, errors.New("git mode staged failed to resolve changed files")
		}
	case "diff":
		target := strings.TrimSpace(request.DiffTarget)
		if target == "" {
			return nil, errors.New("git mode diff requires diff target")
		}

		output, err = runCommand(request.WorkingDir, "diff", "--name-only", "--diff-filter=ACMR", "-z", target)
		if err != nil {
			return nil, fmt.Errorf("git mode diff failed to resolve changed files for target %q", target)
		}
	default:
		return nil, fmt.Errorf("git mode %s is not supported", mode)
	}

	return normalizeCandidateFiles(output)
}

func normalizeCandidateFiles(output string) ([]string, error) {
	rawPaths := splitGitNameOnlyOutput(output)
	files := make([]string, 0, len(rawPaths))
	seen := make(map[string]struct{}, len(rawPaths))

	for _, rawPath := range rawPaths {
		normalized, err := normalizeRelativePath(rawPath)
		if err != nil {
			return nil, err
		}
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}
		files = append(files, normalized)
	}

	sort.Strings(files)

	return files, nil
}

func splitGitNameOnlyOutput(output string) []string {
	if strings.Contains(output, "\x00") {
		return strings.Split(output, "\x00")
	}

	return strings.Split(output, "\n")
}

func normalizeRelativePath(rawPath string) (string, error) {
	if rawPath == "" {
		return "", nil
	}

	normalized := strings.ReplaceAll(rawPath, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = path.Clean(normalized)

	if normalized == "." || normalized == "" {
		return "", nil
	}
	if path.IsAbs(normalized) || isWindowsAbs(normalized) {
		return "", fmt.Errorf("git returned non-root-relative path %q", rawPath)
	}
	if normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", fmt.Errorf("git returned non-root-relative path %q", rawPath)
	}

	return normalized, nil
}

func isWindowsAbs(value string) bool {
	if len(value) < windowsAbsoluteMinLength {
		return false
	}
	drive := value[0]
	if !((drive >= 'a' && drive <= 'z') || (drive >= 'A' && drive <= 'Z')) {
		return false
	}

	return value[1] == ':' && value[2] == '/'
}
