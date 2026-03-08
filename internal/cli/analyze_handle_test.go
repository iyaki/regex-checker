//nolint:testpackage
package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/iyaki/reglint/internal/output"
	"github.com/iyaki/reglint/internal/rules"
)

var cwdMutex sync.Mutex

func TestHandleAnalyzeMissingConfig(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

	var output bytes.Buffer
	code := HandleAnalyze([]string{"--config", filepath.Join(t.TempDir(), "missing.yaml")}, &output)

	if code != exitCodeError {
		t.Fatalf("expected exit code %d, got %d", exitCodeError, code)
	}
	if !strings.Contains(output.String(), "config file not found") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestHandleAnalyzeSurfaceRenderErrors(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

	config := "rules:\n  - message: 'hello'\n    regex: 'world'\n"
	configPath := writeConfig(t, config)

	var output bytes.Buffer
	code := HandleAnalyze([]string{"--config", configPath, "--format", "json", "--out-json", t.TempDir()}, &output)

	if code != exitCodeError {
		t.Fatalf("expected exit code %d, got %d", exitCodeError, code)
	}
	if !strings.Contains(output.String(), "output path is a directory") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestHandleAnalyzeSurfaceRegistryErrors(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

	currentRegistry := outputRegistry
	outputRegistry = func([]rules.Rule, output.ConsoleColorSettings) (*output.Registry, error) {
		return nil, errors.New("registry failed")
	}
	t.Cleanup(func() {
		outputRegistry = currentRegistry
	})

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "clean")
	configPath := writeConfig(t, sampleConfig())

	var output bytes.Buffer
	code := HandleAnalyze([]string{
		"--config", configPath,
		rootDir,
	}, &output)

	if code != exitCodeError {
		t.Fatalf("expected exit code %d, got %d", exitCodeError, code)
	}
	if !strings.Contains(output.String(), "registry failed") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestHandleAnalyzeFailOnThreshold(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

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

func TestExitCodeFailOnConstant(t *testing.T) {
	t.Parallel()

	if exitCodeFailOn != 2 {
		t.Fatalf("expected exitCodeFailOn to be 2, got %d", exitCodeFailOn)
	}
}

func TestHandleAnalyzeNoMatches(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

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

func TestHandleAnalyzeReturnsZeroWhenFailOnUnset(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "token=abc")
	configPath := writeConfig(t, sampleConfig())

	var output bytes.Buffer
	code := HandleAnalyze([]string{
		"--config", configPath,
		rootDir,
	}, &output)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output.String(), "Summary:") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestHandleAnalyzeAcceptsShortFlags(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "clean")
	configPath := writeConfig(t, sampleConfig())

	var output bytes.Buffer
	code := HandleAnalyze([]string{
		"-c", configPath,
		"-f", "console",
		rootDir,
	}, &output)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output.String(), "No matches found.") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestHandleAnalyzeNoColorEnvOverridesConfigEnabledColors(t *testing.T) {
	setAnalyzeCwd(t)
	t.Setenv("NO_COLOR", "1")

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "token=abc")
	configPath := writeConfig(t, configWithConsoleColorsEnabled(true))

	var output bytes.Buffer
	code := HandleAnalyze([]string{
		"--config", configPath,
		rootDir,
	}, &output)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if strings.Contains(output.String(), "\x1b[") {
		t.Fatalf("expected NO_COLOR to disable ANSI output, got: %q", output.String())
	}
	if !strings.Contains(output.String(), "- ERROR 1:1 Found token token=abc") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestHandleAnalyzeConfigEnabledColorsWithoutNoColorEnv(t *testing.T) {
	setAnalyzeCwd(t)
	t.Setenv("NO_COLOR", "")

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "token=abc")
	configPath := writeConfig(t, configWithConsoleColorsEnabled(true))

	var output bytes.Buffer
	code := HandleAnalyze([]string{
		"--config", configPath,
		rootDir,
	}, &output)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output.String(), "\x1b[31mERROR\x1b[0m") {
		t.Fatalf("expected ANSI-colored error label, got: %q", output.String())
	}
}

func TestHandleAnalyzeReturnsErrorWhenFormatsInvalid(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

	configPath := writeConfig(t, sampleConfig())

	var output bytes.Buffer
	code := HandleAnalyze([]string{
		"--config", configPath,
		"--format", "bogus",
	}, &output)

	if code != exitCodeError {
		t.Fatalf("expected exit code %d, got %d", exitCodeError, code)
	}
	if !strings.Contains(output.String(), "invalid format") {
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

func configWithConsoleColorsEnabled(enabled bool) string {
	return "consoleColorsEnabled: " + fmt.Sprintf("%t", enabled) + "\n" + sampleConfig()
}

func setAnalyzeCwd(t *testing.T) {
	t.Helper()

	cwdMutex.Lock()
	t.Cleanup(func() {
		cwdMutex.Unlock()
	})

	currentRegistry := outputRegistry
	current, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to read cwd: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change cwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(current)
		outputRegistry = currentRegistry
	})
}
