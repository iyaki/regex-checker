package scan

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/iyaki/regex-checker/internal/rules"
)

const (
	binaryProbeSize = 8000
	nullByte        = 0x00
)

const (
	captureIndexPairSize = 2
)

func collectFiles(roots []string, include []string, exclude []string, maxFileSizeBytes int64) ([]string, int, error) {
	var files []string
	skipped := 0

	for _, root := range roots {
		root = filepath.Clean(root)
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			relPath = filepath.ToSlash(relPath)
			if relPath == "." {
				return nil
			}

			selected, fileSkipped, err := evaluateFile(path, relPath, entry, include, exclude, maxFileSizeBytes)
			if err != nil {
				return err
			}
			if fileSkipped {
				skipped++

				return nil
			}
			if selected {
				files = append(files, relPath)
			}

			return nil
		}); err != nil {
			return nil, 0, err
		}
	}

	sort.Strings(files)

	return files, skipped, nil
}

type compiledRule struct {
	rule    rules.Rule
	regex   *regexp.Regexp
	paths   []string
	exclude []string
}

type fileEntry struct {
	root    string
	relPath string
}

// Run executes a scan request and returns the aggregated result.
func Run(request Request) (Result, error) {
	start := time.Now()

	entries, skipped, include, exclude, err := collectScanEntries(request)
	if err != nil {
		return Result{}, err
	}

	compiled, err := compileRules(request.Rules, include, exclude)
	if err != nil {
		return Result{}, err
	}

	matches, filesScanned, filesSkipped, err := scanEntries(entries, compiled)
	if err != nil {
		return Result{}, err
	}
	filesSkipped += skipped

	sortMatches(matches)

	result := Result{
		Matches: matches,
		Stats: Stats{
			FilesScanned: filesScanned,
			FilesSkipped: filesSkipped,
			Matches:      len(matches),
			DurationMs:   time.Since(start).Milliseconds(),
		},
	}

	return result, nil
}

func collectScanEntries(request Request) ([]fileEntry, int, []string, []string, error) {
	include := normalizePatterns(request.Include)
	exclude := normalizePatterns(request.Exclude)
	if len(include) == 0 {
		return nil, 0, nil, nil, errors.New("include patterns required")
	}

	entries, skipped, err := collectEntries(request.Roots, include, exclude, request.MaxFileSizeBytes)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	return entries, skipped, include, exclude, nil
}

func scanEntries(entries []fileEntry, compiled []compiledRule) ([]Match, int, int, error) {
	var matches []Match
	filesScanned := 0
	filesSkipped := 0

	for _, entry := range entries {
		fullPath := filepath.Join(entry.root, filepath.FromSlash(entry.relPath))
		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			filesSkipped++

			continue
		}
		filesScanned++
		content := string(contentBytes)

		for _, rule := range compiled {
			match, err := matchesPath(entry.relPath, rule.paths, rule.exclude)
			if err != nil {
				return nil, 0, 0, err
			}
			if !match {
				continue
			}

			indices := rule.regex.FindAllStringSubmatchIndex(content, -1)
			for _, index := range indices {
				captures := buildCaptures(content, index)
				line, column := lineColumnFromIndex(content, index[0])
				message := rules.InterpolateMessage(rule.rule.Message, captures)

				matches = append(matches, Match{
					Message:   message,
					Severity:  rule.rule.Severity,
					FilePath:  entry.relPath,
					Line:      line,
					Column:    column,
					MatchText: captures[0],
				})
			}
		}
	}

	return matches, filesScanned, filesSkipped, nil
}

func collectEntries(
	roots []string,
	include []string,
	exclude []string,
	maxFileSizeBytes int64,
) ([]fileEntry, int, error) {
	entries := make([]fileEntry, 0)
	skipped := 0

	for _, root := range roots {
		files, fileSkipped, err := collectFiles([]string{root}, include, exclude, maxFileSizeBytes)
		if err != nil {
			return nil, 0, err
		}
		skipped += fileSkipped
		for _, relPath := range files {
			entries = append(entries, fileEntry{root: root, relPath: relPath})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].relPath != entries[j].relPath {
			return entries[i].relPath < entries[j].relPath
		}

		return entries[i].root < entries[j].root
	})

	return entries, skipped, nil
}

func compileRules(ruleList []rules.Rule, defaultInclude []string, defaultExclude []string) ([]compiledRule, error) {
	compiled := make([]compiledRule, len(ruleList))
	for i, rule := range ruleList {
		if strings.TrimSpace(rule.Regex) == "" {
			return nil, errors.New("rule regex must not be empty")
		}
		regex, err := regexp.Compile(rule.Regex)
		if err != nil {
			return nil, err
		}
		paths := normalizePatterns(rule.Paths)
		if len(paths) == 0 {
			paths = append([]string{}, defaultInclude...)
		}
		exclude := normalizePatterns(rule.Exclude)
		if len(exclude) == 0 {
			exclude = append([]string{}, defaultExclude...)
		}
		compiled[i] = compiledRule{
			rule:    rule,
			regex:   regex,
			paths:   paths,
			exclude: exclude,
		}
	}

	return compiled, nil
}

func buildCaptures(content string, index []int) []string {
	count := len(index) / captureIndexPairSize
	captures := make([]string, count)
	for i := 0; i < count; i++ {
		start := index[i*2]
		end := index[i*2+1]
		if start == -1 || end == -1 {
			captures[i] = ""

			continue
		}
		captures[i] = content[start:end]
	}

	return captures
}

func lineColumnFromIndex(content string, byteIndex int) (int, int) {
	line := 1
	column := 1
	for idx, runeValue := range content {
		if idx >= byteIndex {
			break
		}
		if runeValue == '\n' {
			line++
			column = 1

			continue
		}
		column++
	}

	return line, column
}

func sortMatches(matches []Match) {
	sort.Slice(matches, func(i, j int) bool {
		left := matches[i]
		right := matches[j]
		if left.FilePath != right.FilePath {
			return left.FilePath < right.FilePath
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Column != right.Column {
			return left.Column < right.Column
		}
		if left.Severity != right.Severity {
			return severityRank(left.Severity) < severityRank(right.Severity)
		}

		return left.Message < right.Message
	})
}

const (
	severityRankError = iota
	severityRankWarning
	severityRankNotice
	severityRankInfo
	severityRankUnknown
)

func severityRank(value string) int {
	switch value {
	case "error":
		return severityRankError
	case "warning":
		return severityRankWarning
	case "notice":
		return severityRankNotice
	case "info":
		return severityRankInfo
	default:
		return severityRankUnknown
	}
}

func matchesPath(path string, include []string, exclude []string) (bool, error) {
	if len(include) == 0 {
		return false, errors.New("include patterns required")
	}

	for _, pattern := range exclude {
		ok, err := doublestar.Match(pattern, path)
		if err != nil {
			return false, err
		}
		if ok {
			return false, nil
		}
	}

	for _, pattern := range include {
		ok, err := doublestar.Match(pattern, path)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}

func normalizePatterns(patterns []string) []string {
	if len(patterns) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func evaluateFile(
	path string,
	relPath string,
	entry os.DirEntry,
	include []string,
	exclude []string,
	maxFileSizeBytes int64,
) (bool, bool, error) {
	match, err := matchesPath(relPath, include, exclude)
	if err != nil {
		return false, false, err
	}
	if !match {
		return false, false, nil
	}

	skip, err := shouldSkipFile(path, entry, maxFileSizeBytes)
	if err != nil {
		return false, false, err
	}
	if skip {
		return false, true, nil
	}

	return true, false, nil
}

func shouldSkipFile(path string, entry os.DirEntry, maxFileSizeBytes int64) (bool, error) {
	info, err := entry.Info()
	if err != nil {
		return false, err
	}
	if maxFileSizeBytes > 0 && info.Size() > maxFileSizeBytes {
		return true, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}

	buffer := make([]byte, binaryProbeSize)
	read, readErr := file.Read(buffer)
	closeErr := file.Close()
	if closeErr != nil {
		return false, closeErr
	}
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return false, readErr
	}
	for _, value := range buffer[:read] {
		if value == nullByte {
			return true, nil
		}
	}

	return false, nil
}
