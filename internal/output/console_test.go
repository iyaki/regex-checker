//nolint:testpackage
package output

import (
	"bytes"
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

	assertConsoleOutput(t, buffer.String(), expectedGroupedOutput())
}

func expectedGroupedOutput() string {
	return "a/file.go\n" +
		"  ERROR 2:5 Error msg\n" +
		"  WARN  2:5 Alpha warn\n" +
		"  WARN  2:5 Zulu warn\n" +
		"  INFO  10:1 Info msg\n\n" +
		"b/file.go\n" +
		"  WARN  2:3 Warn msg\n\n" +
		"Summary: files=2 skipped=1 matches=5 durationMs=12\n"
}

func assertConsoleOutput(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected console output:\n%s", got)
	}
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

func TestSeverityLabelUnknownUppercase(t *testing.T) {
	t.Parallel()

	if severityLabel("custom") != "CUSTOM" {
		t.Fatalf("unexpected label for custom severity")
	}
}
