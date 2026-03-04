//nolint:testpackage
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleAnalyzeMissingConfig(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := HandleAnalyze([]string{"--config", filepath.Join(t.TempDir(), "missing.yaml")}, &output)

	if code != exitCodeError {
		t.Fatalf("expected exit code %d, got %d", exitCodeError, code)
	}
	if !strings.Contains(output.String(), "config file not found") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestHandleAnalyzeFailOnThreshold(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "token=abc")
	configPath := writeConfig(t, sampleConfig())

	var output bytes.Buffer
	code := HandleAnalyze([]string{
		"--config", configPath,
		"--fail-on", "warning",
		rootDir,
	}, &output)

	if code != exitCodeFailOn {
		t.Fatalf("expected exit code %d, got %d", exitCodeFailOn, code)
	}
	if !strings.Contains(output.String(), "Summary:") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestHandleAnalyzeNoMatches(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "clean")
	configPath := writeConfig(t, sampleConfig())

	var output bytes.Buffer
	code := HandleAnalyze([]string{
		"--config", configPath,
		rootDir,
	}, &output)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output.String(), "No matches found.") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestSeverityRankUnknown(t *testing.T) {
	t.Parallel()

	if severityRank("bogus") != severityRankUnknown {
		t.Fatalf("unexpected severity rank")
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "rules.yaml")
	if err := os.WriteFile(path, []byte(contents), defaultFileMode); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return path
}

func writeFile(t *testing.T, dir, name, contents string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), defaultFileMode); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

func sampleConfig() string {
	return "rules:\n  - message: \"Found token $0\"\n    regex: \"token=[a-z]+\"\n    severity: \"error\"\n"
}
