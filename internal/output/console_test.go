//nolint:testpackage
package output

import (
	"bytes"
	"os"
	"runtime"
	"testing"

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

	fileURI, err := fileURIWithLine(match.FilePath, match.Line)
	if err != nil {
		t.Fatalf("failed to build file uri: %v", err)
	}

	line, err := formatConsoleMatchLine(match)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "  ERROR 2:5 Error msg " + fileURI + "\n"
	if line != expected {
		t.Fatalf("unexpected match line: %s", line)
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

	fileA2 := expectedFileURI(t, "a/file.go", 2)
	fileA10 := expectedFileURI(t, "a/file.go", 10)
	fileB2 := expectedFileURI(t, "b/file.go", 2)

	return "a/file.go\n" +
		"  ERROR 2:5 Error msg " + fileA2 + "\n" +
		"  WARN  2:5 Alpha warn " + fileA2 + "\n" +
		"  WARN  2:5 Zulu warn " + fileA2 + "\n" +
		"  INFO  10:1 Info msg " + fileA10 + "\n\n" +
		"b/file.go\n" +
		"  WARN  2:3 Warn msg " + fileB2 + "\n\n" +
		"Summary: files=2 skipped=1 matches=5 durationMs=12\n"
}

func assertConsoleOutput(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected console output:\n%s", got)
	}
}

func expectedFileURI(t *testing.T, filePath string, line int) string {
	t.Helper()

	uri, err := fileURIWithLine(filePath, line)
	if err != nil {
		t.Fatalf("failed to build file uri: %v", err)
	}

	return uri
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
