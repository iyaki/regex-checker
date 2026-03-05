//nolint:testpackage
package output

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/iyaki/regex-checker/internal/rules"
	"github.com/iyaki/regex-checker/internal/scan"
)

func TestWriteConsoleNoMatches(t *testing.T) {
	t.Parallel()

	result := scan.Result{
		Matches: nil,
		Stats: scan.Stats{
			FilesScanned: 0,
			FilesSkipped: 0,
			Matches:      0,
			DurationMs:   0,
		},
	}

	var buffer bytes.Buffer
	if err := WriteConsole(result, &buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "No matches found.\n" +
		"Summary: files=0 skipped=0 matches=0 durationMs=0\n"
	if buffer.String() != expected {
		t.Fatalf("unexpected console output:\n%s", buffer.String())
	}
}

func TestWriteConsoleOrdersAndGroupsMatches(t *testing.T) {
	t.Parallel()

	result := scan.Result{
		Matches: []scan.Match{
			{
				Message:  "Warn msg",
				Severity: "warning",
				FilePath: "b/file.go",
				Line:     2,
				Column:   3,
			},
			{
				Message:  "Info msg",
				Severity: "info",
				FilePath: "a/file.go",
				Line:     10,
				Column:   1,
			},
			{
				Message:  "Error msg",
				Severity: "error",
				FilePath: "a/file.go",
				Line:     2,
				Column:   5,
			},
			{
				Message:  "Zulu warn",
				Severity: "warning",
				FilePath: "a/file.go",
				Line:     2,
				Column:   5,
			},
			{
				Message:  "Alpha warn",
				Severity: "warning",
				FilePath: "a/file.go",
				Line:     2,
				Column:   5,
			},
		},
		Stats: scan.Stats{
			FilesScanned: 2,
			FilesSkipped: 1,
			Matches:      5,
			DurationMs:   12,
		},
	}

	var buffer bytes.Buffer
	if err := WriteConsole(result, &buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertConsoleOutput(t, buffer.String(), expectedGroupedOutput(t))
}

func TestFormatConsoleMatchLine(t *testing.T) {
	t.Parallel()

	match := scan.Match{
		Message:  "Error msg",
		Severity: "error",
		FilePath: "a/file.go",
		Line:     2,
		Column:   5,
	}

	line, err := formatConsoleMatchLine(match)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	absPath, err := expectedAbsolutePathLine(match.FilePath, match.Line)
	if err != nil {
		t.Fatalf("failed to build absolute path: %v", err)
	}

	expected := "- ERROR 2:5 Error msg\n" +
		"  " + absPath + "\n"
	if line != expected {
		t.Fatalf("unexpected match line: %s", line)
	}
}

func TestWriteConsoleUsesScanRootForAbsolutePath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	scanRoot := filepath.Join(root, "fixtures")
	if err := os.MkdirAll(scanRoot, 0o755); err != nil {
		t.Fatalf("failed to create fixtures dir: %v", err)
	}
	filePath := filepath.Join(scanRoot, "sample.txt")
	if err := os.WriteFile(filePath, []byte("token=abc"), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	request := scan.Request{
		Roots: []string{scanRoot},
		Rules: []rules.Rule{
			{
				Message:  "Found $0",
				Regex:    "token=[a-z]+",
				Severity: "error",
				Paths:    []string{"**/*"},
			},
		},
		Include:          []string{"**/*"},
		Exclude:          nil,
		MaxFileSizeBytes: 1024,
		Concurrency:      1,
	}

	result, err := scan.Run(request)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Matches))
	}

	var buffer bytes.Buffer
	if err := WriteConsole(result, &buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedLine := fmt.Sprintf("%s:1", filePath)
	if !strings.Contains(buffer.String(), expectedLine) {
		t.Fatalf("expected absolute path %s in output, got:\n%s", expectedLine, buffer.String())
	}
}

func TestFormatConsoleMatchLineReturnsErrorWhenCwdMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("current directory removal is restricted on windows")
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected getwd error: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("unexpected chdir restore error: %v", err)
		}
	}()

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("unexpected chdir error: %v", err)
	}
	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("unexpected remove error: %v", err)
	}

	_, err = formatConsoleMatchLine(scan.Match{FilePath: "relative/file.go", Line: 1})
	if err == nil {
		t.Fatalf("expected error with missing cwd")
	}
}

func expectedGroupedOutput(t *testing.T) string {
	t.Helper()

	fileA2 := expectedAbsolutePathLineHelper(t, "a/file.go", 2)
	fileA10 := expectedAbsolutePathLineHelper(t, "a/file.go", 10)
	fileB2 := expectedAbsolutePathLineHelper(t, "b/file.go", 2)

	return "a/file.go\n" +
		"- ERROR 2:5 Error msg\n" +
		"  " + fileA2 + "\n\n" +
		"- WARN  2:5 Alpha warn\n" +
		"  " + fileA2 + "\n\n" +
		"- WARN  2:5 Zulu warn\n" +
		"  " + fileA2 + "\n\n" +
		"- INFO  10:1 Info msg\n" +
		"  " + fileA10 + "\n\n" +
		"b/file.go\n" +
		"- WARN  2:3 Warn msg\n" +
		"  " + fileB2 + "\n\n" +
		"Summary: files=2 skipped=1 matches=5 durationMs=12\n"
}

func assertConsoleOutput(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected console output:\n%s", got)
	}
}

func expectedAbsolutePathLine(filePath string, line int) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", absPath, line), nil
}

func expectedAbsolutePathLineHelper(t *testing.T, filePath string, line int) string {
	t.Helper()

	absPath, err := expectedAbsolutePathLine(filePath, line)
	if err != nil {
		t.Fatalf("failed to build absolute path: %v", err)
	}

	return absPath
}

func TestSeverityRankKnownValues(t *testing.T) {
	t.Parallel()

	values := []string{"error", "warning", "notice", "info"}
	for _, value := range values {
		if severityRank(value) == severityRankUnknown {
			t.Fatalf("expected severity %s to have known rank", value)
		}
	}
}

func TestSeverityRankUnknownValue(t *testing.T) {
	t.Parallel()

	if severityRank("custom") != severityRankUnknown {
		t.Fatalf("expected unknown severity rank")
	}
}

func TestSeverityLabelUnknownUppercase(t *testing.T) {
	t.Parallel()

	if severityLabel("custom") != "CUSTOM" {
		t.Fatalf("unexpected label for custom severity")
	}
}
