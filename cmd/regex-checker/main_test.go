package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/regex-checker/internal/cli"
	"github.com/iyaki/regex-checker/internal/scan"
)

func TestRunShowsHelpWhenNoArgs(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := run([]string{}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("expected usage help, got %q", output.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := run([]string{"bogus"}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	if output.String() != "Unknown command: bogus\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestRunRoutesAnalyze(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := run([]string{"analyze"}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1 for missing config, got %d", code)
	}
	if !strings.Contains(output.String(), "config file not found: regex-rules.yaml") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestRunRoutesAnalyseAlias(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := run([]string{"analyse"}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1 for missing config, got %d", code)
	}
	if !strings.Contains(output.String(), "config file not found: regex-rules.yaml") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestRunAnalyzeWritesJSONToStdout(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	configDir := t.TempDir()
	writeFixture(t, rootDir, "sample.txt", "token=abc")
	configPath := writeRuleConfig(t, configDir, "")

	var output bytes.Buffer
	code := run([]string{"analyze", "--config", configPath, "--format", "json", rootDir}, &output)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	var got jsonResult
	if err := json.Unmarshal(output.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse json output: %v", err)
	}
	if got.SchemaVersion != 1 {
		t.Fatalf("unexpected schema version: %d", got.SchemaVersion)
	}
	if len(got.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got.Matches))
	}
	match := got.Matches[0]
	if match.FilePath != "sample.txt" {
		t.Fatalf("unexpected match file: %s", match.FilePath)
	}
	if match.Severity != "error" {
		t.Fatalf("unexpected match severity: %s", match.Severity)
	}
	if got.Stats.Matches != 1 {
		t.Fatalf("unexpected stats matches: %d", got.Stats.Matches)
	}
}

func TestRunAnalyzeWritesJSONFileForMultiFormat(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	configDir := t.TempDir()
	writeFixture(t, rootDir, "sample.txt", "token=abc")
	configPath := writeRuleConfig(t, configDir, "")
	jsonPath := filepath.Join(t.TempDir(), "scan.json")

	var output bytes.Buffer
	code := run([]string{
		"analyze",
		"--config", configPath,
		"--format", "console,json",
		"--out-json", jsonPath,
		rootDir,
	}, &output)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output.String(), "Summary:") {
		t.Fatalf("expected console output summary, got %q", output.String())
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read json output: %v", err)
	}
	var got jsonResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to parse json file: %v", err)
	}
	if len(got.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got.Matches))
	}
}

func TestRunAnalyzeExitCodeFailOn(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	configDir := t.TempDir()
	writeFixture(t, rootDir, "sample.txt", "token=abc")
	configPath := writeRuleConfig(t, configDir, "")

	var output bytes.Buffer
	code := run([]string{
		"analyze",
		"--config", configPath,
		"--fail-on", "warning",
		rootDir,
	}, &output)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}

func TestRunUsesProvidedOutputWriter(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := cli.Run([]string{}, map[string]cli.Handler{}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("expected usage help, got %q", output.String())
	}
}

func TestMainExitsWithRunCode(t *testing.T) {
	var output bytes.Buffer
	var exitCode int
	var exited bool

	originalArgs := args
	originalOutput := outputWriter
	originalExit := exitFunc
	defer func() {
		args = originalArgs
		outputWriter = originalOutput
		exitFunc = originalExit
	}()

	args = []string{"regex-checker"}
	outputWriter = &output
	exitFunc = func(code int) {
		exited = true
		exitCode = code
	}

	main()

	if !exited {
		t.Fatal("expected main to call exit")
	}
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("expected usage help, got %q", output.String())
	}
}

type jsonResult struct {
	SchemaVersion int          `json:"schemaVersion"`
	Matches       []scan.Match `json:"matches"`
	Stats         scan.Stats   `json:"stats"`
}

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	return path
}

func writeRuleConfig(t *testing.T, dir, failOn string) string {
	t.Helper()

	config := "rules:\n  - message: \"Found token $0\"\n    regex: \"token=[a-z]+\"\n    severity: \"error\"\n"
	if failOn != "" {
		config = "failOn: \"" + failOn + "\"\n" + config
	}
	path := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return path
}
