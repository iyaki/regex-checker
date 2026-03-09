//nolint:testpackage
package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/iyaki/reglint/internal/output"
	"github.com/iyaki/reglint/internal/rules"
	"github.com/iyaki/reglint/internal/scan"
)

func lockAnalyzeOutput(t *testing.T) {
	t.Helper()

	cwdMutex.Lock()
	t.Cleanup(func() {
		cwdMutex.Unlock()
	})
}

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
	if !strings.Contains(buffer.String(), "RegLint") {
		t.Fatalf("unexpected stdout output: %q", buffer.String())
	}
}

func TestRenderOutputsWritesJSONFile(t *testing.T) {
	t.Parallel()
	lockAnalyzeOutput(t)

	path := filepath.Join(t.TempDir(), "scan.json")
	cfg := Config{Formats: []string{"json"}, OutJSON: path}
	buffer := &bytes.Buffer{}

	if err := renderOutputs(
		cfg.Formats,
		sampleRules(),
		cfg,
		output.ConsoleColorSettings{},
		scan.Result{},
		buffer,
	); err != nil {
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
	lockAnalyzeOutput(t)

	path := filepath.Join(t.TempDir(), "scan.sarif")
	cfg := Config{Formats: []string{"sarif"}, OutSARIF: path}
	buffer := &bytes.Buffer{}

	if err := renderOutputs(
		cfg.Formats,
		sampleRules(),
		cfg,
		output.ConsoleColorSettings{},
		scan.Result{},
		buffer,
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buffer.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", buffer.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read sarif output: %v", err)
	}
	if !strings.Contains(string(data), "RegLint") {
		t.Fatalf("expected sarif output, got %q", string(data))
	}
}

func TestRenderOutputsWritesConsole(t *testing.T) {
	t.Parallel()
	lockAnalyzeOutput(t)

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

	if err := renderOutputs(cfg.Formats, sampleRules(), cfg, output.ConsoleColorSettings{}, result, buffer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buffer.String(), "Summary:") {
		t.Fatalf("expected summary output, got %q", buffer.String())
	}
}

func TestRenderOutputsRejectsUnknownFormat(t *testing.T) {
	t.Parallel()
	lockAnalyzeOutput(t)

	cfg := Config{Formats: []string{"bogus"}}

	if err := renderOutputs(
		cfg.Formats,
		sampleRules(),
		cfg,
		output.ConsoleColorSettings{},
		scan.Result{},
		&bytes.Buffer{},
	); err == nil {
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

	result, failOn, formats, ruleset, cfg, consoleColors, err := runAnalyze([]string{"--config", configPath})
	_ = result
	_ = failOn
	_ = formats
	_ = ruleset
	_ = cfg
	_ = consoleColors
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunAnalyzeReturnsConfigLoadError(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfigFile(t, "rules: []")
	if err := os.Chmod(configPath, 0o000); err != nil {
		t.Fatalf("failed to set permissions: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(configPath, 0o600)
	})

	result, failOn, formats, ruleset, cfg, consoleColors, err := runAnalyze([]string{"--config", configPath})
	_ = result
	_ = failOn
	_ = formats
	_ = ruleset
	_ = cfg
	_ = consoleColors
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunAnalyzeResolvesRuleSetBaselinePathFromConfigDirectory(t *testing.T) {
	t.Parallel()
	lockAnalyzeOutput(t)

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "clean")

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "rules.yaml")
	config := "baseline: baselines/current.json\n" + sampleConfig()
	if err := os.WriteFile(configPath, []byte(config), defaultFileMode); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	baselineDir := filepath.Join(configDir, "baselines")
	if err := os.MkdirAll(baselineDir, 0o700); err != nil {
		t.Fatalf("failed to create baseline directory: %v", err)
	}
	baselinePath := filepath.Join(baselineDir, "current.json")
	if err := os.WriteFile(baselinePath, []byte(`{"schemaVersion":1,"entries":[]}`), defaultFileMode); err != nil {
		t.Fatalf("failed to write baseline: %v", err)
	}

	result, failOn, formats, ruleset, cfg, consoleColors, err := runAnalyze([]string{"--config", configPath, rootDir})
	_ = result
	_ = failOn
	_ = formats
	_ = ruleset
	_ = consoleColors
	assertNoError(t, err)

	want := filepath.Join(configDir, "baselines", "current.json")
	assertConfigStringFieldValue(t, cfg, "RuleSetBaselinePath", want)
	assertConfigStringFieldValue(t, cfg, "EffectiveBaselinePath", want)
	assertConfigStringFieldValue(t, cfg, "BaselinePath", "")
}

func TestRunAnalyzeResolvesCLIBaselinePathFromWorkingDirectory(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "clean")
	configPath := writeConfig(t, sampleConfig())

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to read cwd: %v", err)
	}
	baselineArg := filepath.Join("relative", "baseline.json")
	if err := os.MkdirAll(filepath.Join(cwd, "relative"), 0o700); err != nil {
		t.Fatalf("failed to create baseline directory: %v", err)
	}
	baselinePayload := []byte(`{"schemaVersion":1,"entries":[]}`)
	if err := os.WriteFile(
		filepath.Join(cwd, baselineArg),
		baselinePayload,
		defaultFileMode,
	); err != nil {
		t.Fatalf("failed to write baseline: %v", err)
	}

	result, failOn, formats, ruleset, cfg, consoleColors, err := runAnalyze([]string{
		"--config", configPath,
		"--baseline", baselineArg,
		rootDir,
	})
	_ = result
	_ = failOn
	_ = formats
	_ = ruleset
	_ = consoleColors
	assertNoError(t, err)

	want := filepath.Join(cwd, baselineArg)
	assertConfigStringFieldValue(t, cfg, "BaselinePath", want)
	assertConfigStringFieldValue(t, cfg, "EffectiveBaselinePath", want)
	assertConfigStringFieldValue(t, cfg, "RuleSetBaselinePath", "")
}

func TestRunAnalyzePrefersCLIBaselineOverRuleSetBaseline(t *testing.T) {
	t.Parallel()
	setAnalyzeCwd(t)

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "clean")

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "rules.yaml")
	config := "baseline: config/baseline.json\n" + sampleConfig()
	if err := os.WriteFile(configPath, []byte(config), defaultFileMode); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to read cwd: %v", err)
	}
	cliBaseline := filepath.Join("cli", "baseline.json")
	if err := os.MkdirAll(filepath.Join(cwd, "cli"), 0o700); err != nil {
		t.Fatalf("failed to create baseline directory: %v", err)
	}
	baselinePayload := []byte(`{"schemaVersion":1,"entries":[]}`)
	if err := os.WriteFile(
		filepath.Join(cwd, cliBaseline),
		baselinePayload,
		defaultFileMode,
	); err != nil {
		t.Fatalf("failed to write baseline: %v", err)
	}

	result, failOn, formats, ruleset, cfg, consoleColors, err := runAnalyze([]string{
		"--config", configPath,
		"--baseline", cliBaseline,
		rootDir,
	})
	_ = result
	_ = failOn
	_ = formats
	_ = ruleset
	_ = consoleColors
	assertNoError(t, err)

	wantCLI := filepath.Join(cwd, cliBaseline)
	wantRuleSet := filepath.Join(configDir, "config", "baseline.json")
	assertConfigStringFieldValue(t, cfg, "BaselinePath", wantCLI)
	assertConfigStringFieldValue(t, cfg, "RuleSetBaselinePath", wantRuleSet)
	assertConfigStringFieldValue(t, cfg, "EffectiveBaselinePath", wantCLI)
}

func TestRunAnalyzeWriteBaselineRequiresEffectiveBaselinePath(t *testing.T) {
	t.Parallel()
	lockAnalyzeOutput(t)

	rootDir := t.TempDir()
	writeFile(t, rootDir, "sample.txt", "clean")
	configPath := writeConfig(t, sampleConfig())

	result, failOn, formats, ruleset, cfg, consoleColors, err := runAnalyze([]string{
		"--config", configPath,
		"--write-baseline",
		rootDir,
	})
	_ = result
	_ = failOn
	_ = formats
	_ = ruleset
	_ = cfg
	_ = consoleColors
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--write-baseline requires an effective baseline path") {
		t.Fatalf("expected write-baseline path error, got %v", err)
	}
}

type errorFormatter struct {
	name string
}

func (e errorFormatter) Name() string {
	return e.name
}

func (e errorFormatter) Write(scan.Result, io.Writer) error {
	return errors.New("write failed")
}

func TestRunAnalyzeShortFlags(t *testing.T) {
	cwdMutex.Lock()
	defer cwdMutex.Unlock()

	config := "rules:\n  - message: 'hello'\n    regex: 'world'\n"
	configPath := writeTempConfigFile(t, config)

	current, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to read cwd: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change cwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(current)
	})

	result, failOn, formats, ruleset, cfg, consoleColors, err := runAnalyze([]string{"-c", configPath, "-f", "console"})
	assertNoError(t, err)
	assertFailOnEmpty(t, failOn)
	assertSingleConsoleFormat(t, formats)
	assertRuleSetSize(t, ruleset, 1)
	assertConfigPathValue(t, cfg.ConfigPath, configPath)
	assertDefaultConsoleColors(t, consoleColors)
	assertFilesScannedNonNegative(t, result.Stats.FilesScanned)
}

func TestWriteJSONFileFailsOnDirectory(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	if err := writeJSONFile(path, scan.Result{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteJSONFileFailsOnMissingParent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "scan.json")
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

func TestWriteSARIFFileFailsOnMissingParent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "scan.sarif")
	if err := writeSARIFFile(path, scan.Result{}, sampleRules()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteJSONFileFailsOnReadOnlyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("failed to set permissions: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})

	path := filepath.Join(dir, "scan.json")
	if err := writeJSONFile(path, scan.Result{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteSARIFFileFailsOnReadOnlyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("failed to set permissions: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})

	path := filepath.Join(dir, "scan.sarif")
	if err := writeSARIFFile(path, scan.Result{}, sampleRules()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteJSONFileFailsOnWriteError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "scan.json")
	if err := writeJSONFile(path, scan.Result{Matches: []scan.Match{{FilePath: "", Line: 1}}}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteJSONFileFailsOnCreateError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "scan.json")
	if err := writeJSONFile(path, scan.Result{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWriteJSONPropagatesWriteError(t *testing.T) {
	t.Parallel()

	result := scan.Result{Matches: []scan.Match{{FilePath: "file.txt", Line: 1, Column: 1}}}
	if err := output.WriteJSON(result, failingWriter{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteSARIFPropagatesWriteError(t *testing.T) {
	t.Parallel()

	result := scan.Result{Matches: []scan.Match{{FilePath: "file.txt", Line: 1, Column: 1}}}
	if err := output.WriteSARIF(result, sampleRules(), failingWriter{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteJSONFileFailsOnReadOnlyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "scan.json")
	if err := os.WriteFile(path, []byte("data"), 0o400); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := writeJSONFile(path, scan.Result{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteSARIFFileFailsOnReadOnlyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "scan.sarif")
	if err := os.WriteFile(path, []byte("data"), 0o400); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := writeSARIFFile(path, scan.Result{}, sampleRules()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

type captureFormatter struct {
	name    string
	written bool
}

func (c *captureFormatter) Name() string {
	return c.name
}

func (c *captureFormatter) Write(scan.Result, io.Writer) error {
	c.written = true

	return nil
}

type mockFormatter struct {
	name string
}

func (m mockFormatter) Name() string {
	return m.name
}

func (m mockFormatter) Write(scan.Result, io.Writer) error {
	return nil
}

func TestRenderOutputsRequiresRegistrySetup(t *testing.T) {
	t.Parallel()

	registry, err := output.NewRegistry(mockFormatter{name: ""})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if registry != nil {
		t.Fatal("expected nil registry")
	}
}

func TestRenderOutputsPropagatesFormatterError(t *testing.T) {
	t.Parallel()

	registry, err := output.NewRegistry(errorFormatter{name: "console"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	formatter, err := registry.ResolveName("console")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := Config{Formats: []string{"console"}}
	if err := renderFormat(formatter, cfg, sampleRules(), scan.Result{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRenderOutputsSkipsUnknownFormatWrite(t *testing.T) {
	t.Parallel()
	lockAnalyzeOutput(t)

	formatter := &captureFormatter{name: "console"}
	registry, err := output.NewRegistry(formatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = registry.ResolveName("console")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := Config{Formats: []string{"bogus"}}
	if err := renderOutputs(
		cfg.Formats,
		sampleRules(),
		cfg,
		output.ConsoleColorSettings{},
		scan.Result{},
		&bytes.Buffer{},
	); err == nil {
		t.Fatal("expected error, got nil")
	}
	if formatter.written {
		t.Fatal("expected formatter not to be written")
	}
}

type nilFormatter struct{}

func (nilFormatter) Name() string {
	return "console"
}

func (nilFormatter) Write(scan.Result, io.Writer) error {
	return nil
}

func TestRenderOutputsReturnsErrorWhenResolveFails(t *testing.T) {
	lockAnalyzeOutput(t)

	outputRegistry = func([]rules.Rule, output.ConsoleColorSettings) (*output.Registry, error) {
		return output.NewRegistry(nilFormatter{})
	}
	t.Cleanup(func() {
		outputRegistry = defaultOutputRegistry
	})

	if err := renderOutputs(
		[]string{"missing"},
		sampleRules(),
		Config{},
		output.ConsoleColorSettings{},
		scan.Result{},
		&bytes.Buffer{},
	); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRenderOutputsReturnsErrorWhenFormatterFails(t *testing.T) {
	lockAnalyzeOutput(t)

	currentRegistry := outputRegistry
	outputRegistry = func([]rules.Rule, output.ConsoleColorSettings) (*output.Registry, error) {
		return output.NewRegistry(errorFormatter{name: "console"})
	}
	t.Cleanup(func() {
		outputRegistry = currentRegistry
	})

	if err := renderOutputs(
		[]string{"console"},
		sampleRules(),
		Config{},
		output.ConsoleColorSettings{},
		scan.Result{},
		&bytes.Buffer{},
	); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRenderOutputsReturnsErrorWhenRegistryFails(t *testing.T) {
	lockAnalyzeOutput(t)

	outputRegistry = func([]rules.Rule, output.ConsoleColorSettings) (*output.Registry, error) {
		return nil, errors.New("registry failed")
	}
	t.Cleanup(func() {
		outputRegistry = defaultOutputRegistry
	})

	if err := renderOutputs(
		[]string{"console"},
		sampleRules(),
		Config{},
		output.ConsoleColorSettings{},
		scan.Result{},
		&bytes.Buffer{},
	); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildEffectiveRulesSkipsOverridesWithoutCLIInput(t *testing.T) {
	t.Parallel()

	effective := rules.RuleSet{
		Include: []string{"**/*.go"},
		Exclude: []string{"**/vendor/**"},
		Rules: []rules.Rule{
			{
				Message: "rule",
				Regex:   "token",
				Paths:   []string{"custom-path"},
				Exclude: []string{"custom-exclude"},
			},
		},
	}

	got := buildEffectiveRules(Config{}, effective)
	if len(got) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got))
	}
	if len(got[0].Paths) != 1 || got[0].Paths[0] != "custom-path" {
		t.Fatalf("expected paths to remain custom, got %v", got[0].Paths)
	}
	if len(got[0].Exclude) != 1 || got[0].Exclude[0] != "custom-exclude" {
		t.Fatalf("expected exclude to remain custom, got %v", got[0].Exclude)
	}
}

func TestResolveConcurrencyUsesConfigWhenRulesetMissing(t *testing.T) {
	t.Parallel()

	got := resolveConcurrency(Config{Concurrency: 3}, nil)
	if got != 3 {
		t.Fatalf("expected concurrency 3, got %d", got)
	}
}

func TestRenderOutputsReturnsErrorWhenRegistrySetupFails(t *testing.T) {
	t.Parallel()

	if _, err := output.NewRegistry(nilFormatter{}, nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDefaultOutputRegistryUsesConsoleColorSettings(t *testing.T) {
	t.Parallel()
	lockAnalyzeOutput(t)

	settings := output.ConsoleColorSettings{Enabled: false, Source: output.ConsoleColorSourceConfig}
	registry, err := defaultOutputRegistry(sampleRules(), settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	formatter, err := registry.ResolveName("console")
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}

	consoleFormatter, ok := formatter.(output.ConsoleFormatter)
	if !ok {
		t.Fatalf("expected console formatter type, got %T", formatter)
	}
	if consoleFormatter.ColorSettings != settings {
		t.Fatalf("expected settings %+v, got %+v", settings, consoleFormatter.ColorSettings)
	}
}

func TestRenderOutputsPassesConsoleColorSettingsToRegistry(t *testing.T) {
	lockAnalyzeOutput(t)

	currentRegistry := outputRegistry
	captured := output.ConsoleColorSettings{}
	outputRegistry = func(ruleSet []rules.Rule, settings output.ConsoleColorSettings) (*output.Registry, error) {
		captured = settings

		return output.NewRegistry(
			output.ConsoleFormatter{ColorSettings: settings},
			output.JSONFormatter{},
			output.SARIFFormatter{Rules: ruleSet},
		)
	}
	t.Cleanup(func() {
		outputRegistry = currentRegistry
	})

	expected := output.ConsoleColorSettings{Enabled: false, Source: output.ConsoleColorSourceEnv}
	cfg := Config{Formats: []string{"json"}}
	if err := renderOutputs([]string{"json"}, sampleRules(), cfg, expected, scan.Result{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured != expected {
		t.Fatalf("expected settings %+v, got %+v", expected, captured)
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

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func assertFailOnEmpty(t *testing.T, failOn string) {
	t.Helper()

	if failOn != "" {
		t.Fatalf("expected empty fail-on, got %q", failOn)
	}
}

func assertSingleConsoleFormat(t *testing.T, formats []string) {
	t.Helper()

	if len(formats) != 1 || formats[0] != "console" {
		t.Fatalf("expected formats [console], got %v", formats)
	}
}

func assertRuleSetSize(t *testing.T, ruleSet []rules.Rule, expected int) {
	t.Helper()

	if len(ruleSet) != expected {
		t.Fatalf("expected %d rule, got %d", expected, len(ruleSet))
	}
}

func assertConfigPathValue(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("expected config path %q, got %q", want, got)
	}
}

func assertDefaultConsoleColors(t *testing.T, settings output.ConsoleColorSettings) {
	t.Helper()

	if settings.Source != output.ConsoleColorSourceDefault {
		t.Fatalf("expected default console color source, got %q", settings.Source)
	}
	if !settings.Enabled {
		t.Fatal("expected default console colors enabled")
	}
}

func assertFilesScannedNonNegative(t *testing.T, filesScanned int) {
	t.Helper()

	if filesScanned < 0 {
		t.Fatalf("unexpected files scanned: %d", filesScanned)
	}
}

func assertConfigStringFieldValue(t *testing.T, cfg Config, fieldName, want string) {
	t.Helper()

	value := reflect.ValueOf(cfg)
	field := value.FieldByName(fieldName)
	if !field.IsValid() {
		t.Fatalf("expected config to include %s field", fieldName)
	}
	if field.Kind() != reflect.String {
		t.Fatalf("expected config field %s to be string, got %s", fieldName, field.Kind())
	}

	if field.String() != want {
		t.Fatalf("expected config field %s=%q, got %q", fieldName, want, field.String())
	}
}
