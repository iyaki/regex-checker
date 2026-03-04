//nolint:testpackage
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/regex-checker/internal/rules"
	"github.com/iyaki/regex-checker/internal/scan"
)

func TestWriteJSONOutputRequiresPathForMultipleFormats(t *testing.T) {
	t.Parallel()

	cfg := Config{Formats: []string{"console", "json"}}

	if err := writeJSONOutput(cfg, scan.Result{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteJSONOutputToStdout(t *testing.T) {
	t.Parallel()

	cfg := Config{Formats: []string{"json"}}
	buffer := &bytes.Buffer{}

	if err := writeJSONOutput(cfg, scan.Result{}, buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buffer.String(), "schemaVersion") {
		t.Fatalf("unexpected stdout output: %q", buffer.String())
	}
}

func TestWriteSARIFOutputRequiresPathForMultipleFormats(t *testing.T) {
	t.Parallel()

	cfg := Config{Formats: []string{"console", "sarif"}}

	if err := writeSARIFOutput(cfg, scan.Result{}, sampleRules(), &bytes.Buffer{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteSARIFOutputToStdout(t *testing.T) {
	t.Parallel()

	cfg := Config{Formats: []string{"sarif"}}
	buffer := &bytes.Buffer{}

	if err := writeSARIFOutput(cfg, scan.Result{}, sampleRules(), buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buffer.String(), "regex-checker") {
		t.Fatalf("unexpected stdout output: %q", buffer.String())
	}
}

func TestRenderOutputsWritesJSONFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "scan.json")
	cfg := Config{Formats: []string{"json"}, OutJSON: path}
	buffer := &bytes.Buffer{}

	if err := renderOutputs(cfg.Formats, sampleRules(), cfg, scan.Result{}, buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buffer.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", buffer.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read json output: %v", err)
	}
	if !strings.Contains(string(data), "schemaVersion") {
		t.Fatalf("expected json output, got %q", string(data))
	}
}

func TestRenderOutputsWritesSARIFFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "scan.sarif")
	cfg := Config{Formats: []string{"sarif"}, OutSARIF: path}
	buffer := &bytes.Buffer{}

	if err := renderOutputs(cfg.Formats, sampleRules(), cfg, scan.Result{}, buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buffer.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", buffer.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read sarif output: %v", err)
	}
	if !strings.Contains(string(data), "regex-checker") {
		t.Fatalf("expected sarif output, got %q", string(data))
	}
}

func TestRenderOutputsWritesConsole(t *testing.T) {
	t.Parallel()

	result := scan.Result{
		Matches: []scan.Match{{Message: "msg", Severity: "error", FilePath: "file.txt", Line: 1, Column: 1}},
		Stats: scan.Stats{
			FilesScanned: 1,
			FilesSkipped: 0,
			Matches:      1,
			DurationMs:   2,
		},
	}
	cfg := Config{Formats: []string{"console"}}
	buffer := &bytes.Buffer{}

	if err := renderOutputs(cfg.Formats, sampleRules(), cfg, result, buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buffer.String(), "Summary:") {
		t.Fatalf("expected summary output, got %q", buffer.String())
	}
}

func TestRenderOutputsRejectsUnknownFormat(t *testing.T) {
	t.Parallel()

	cfg := Config{Formats: []string{"bogus"}}

	if err := renderOutputs(cfg.Formats, sampleRules(), cfg, scan.Result{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExitCodeForFailOn(t *testing.T) {
	t.Parallel()

	matches := []scan.Match{{Severity: "warning"}}
	if exitCodeForFailOn(matches, "warning") != exitCodeFailOn {
		t.Fatalf("expected fail-on exit code")
	}
	if exitCodeForFailOn(matches, "error") != 0 {
		t.Fatalf("expected success exit code")
	}
}

func TestRunAnalyzeReturnsScanError(t *testing.T) {
	t.Parallel()

	config := "include:\n  - ''\nrules:\n  - message: 'hello'\n    regex: 'world'\n"
	configPath := writeTempConfigFile(t, config)

	result, failOn, formats, ruleset, cfg, err := runAnalyze([]string{"--config", configPath})
	_ = result
	_ = failOn
	_ = formats
	_ = ruleset
	_ = cfg
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteJSONFileFailsOnDirectory(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	if err := writeJSONFile(path, scan.Result{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteSARIFFileFailsOnDirectory(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	if err := writeSARIFFile(path, scan.Result{}, sampleRules()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func sampleRules() []rules.Rule {
	return []rules.Rule{
		{
			Message:  "rule",
			Regex:    "token",
			Severity: "warning",
			Index:    0,
		},
	}
}

func writeTempConfigFile(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "rules.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return path
}
