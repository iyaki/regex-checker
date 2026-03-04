package cli_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/iyaki/regex-checker/internal/cli"
)

func TestParseAnalyzeDefaults(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	got, err := cli.ParseAnalyzeArgs([]string{"--config", configPath})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertDefaultConfig(t, got, configPath)
}

func TestParseAnalyzeFormatsDedupAndTrim(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	got, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "json, json"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{"json"}
	if len(got.Formats) != len(want) {
		t.Fatalf("expected formats %v, got %v", want, got.Formats)
	}
	for i := range want {
		if got.Formats[i] != want[i] {
			t.Fatalf("expected formats %v, got %v", want, got.Formats)
		}
	}
}

func TestParseAnalyzeIncludeExclude(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	got, err := cli.ParseAnalyzeArgs([]string{
		"--config", configPath,
		"--include", "**/*.go",
		"--include", "**/*.md",
		"--exclude", "**/vendor/**",
		"--exclude", "**/.git/**",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(got.Include) != 2 || got.Include[0] != "**/*.go" || got.Include[1] != "**/*.md" {
		t.Fatalf("unexpected include list: %v", got.Include)
	}
	if len(got.Exclude) != 2 || got.Exclude[0] != "**/vendor/**" || got.Exclude[1] != "**/.git/**" {
		t.Fatalf("unexpected exclude list: %v", got.Exclude)
	}
}

func TestParseAnalyzeInvalidFormat(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "bogus"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeInvalidFailOn(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--fail-on", "fatal"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeRequiresOutPathForMultiFormat(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "console,json"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeAllowsSingleJsonToStdout(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "json"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestParseAnalyzeAllowsSingleSarifToStdout(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "sarif"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func assertDefaultConfig(t *testing.T, got cli.Config, configPath string) {
	t.Helper()

	if got.ConfigPath != configPath {
		t.Fatalf("expected config path %q, got %q", configPath, got.ConfigPath)
	}
	if len(got.Roots) != 1 || got.Roots[0] != "." {
		t.Fatalf("expected default root '.', got %v", got.Roots)
	}
	if len(got.Formats) != 1 || got.Formats[0] != "console" {
		t.Fatalf("expected default format [console], got %v", got.Formats)
	}
	if got.Concurrency != runtime.GOMAXPROCS(0) {
		t.Fatalf("expected concurrency %d, got %d", runtime.GOMAXPROCS(0), got.Concurrency)
	}
	if got.MaxFileSizeBytes != 5242880 {
		t.Fatalf("expected max file size 5242880, got %d", got.MaxFileSizeBytes)
	}
	if got.OutJSON != "" || got.OutSARIF != "" {
		t.Fatalf("expected empty output paths, got out-json=%q out-sarif=%q", got.OutJSON, got.OutSARIF)
	}
}

func TestParseAnalyzeRejectsEmptyFormatValue(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", ""})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeRequiresConfigFile(t *testing.T) {
	t.Parallel()

	missingPath := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", missingPath})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeRejectsZeroConcurrency(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--concurrency", "0"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeRejectsZeroMaxFileSize(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--max-file-size", "0"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeRejectsUnwritableOutJSON(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputDir := readOnlyDir(t)
	outputPath := filepath.Join(outputDir, "scan.json")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "json", "--out-json", outputPath})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeRejectsUnwritableOutSARIF(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputDir := readOnlyDir(t)
	outputPath := filepath.Join(outputDir, "scan.sarif")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "sarif", "--out-sarif", outputPath})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeAllowsWritableOutJSONFile(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := writableFile(t, "scan.json")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "json", "--out-json", outputPath})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestParseAnalyzeRejectsReadOnlyOutJSONFile(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := readOnlyFile(t, "scan.json")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "json", "--out-json", outputPath})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeRejectsOutJSONWithMissingParent(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := filepath.Join(t.TempDir(), "missing", "scan.json")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "json", "--out-json", outputPath})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeRejectsOutJSONDirectory(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := t.TempDir()

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "json", "--out-json", outputPath})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeAllowsWritableOutSARIFFile(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := writableFile(t, "scan.sarif")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "sarif", "--out-sarif", outputPath})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestParseAnalyzeRejectsOutSARIFDirectory(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := t.TempDir()

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "sarif", "--out-sarif", outputPath})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseAnalyzeAllowsNewOutJSONInWritableDir(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := filepath.Join(t.TempDir(), "new.json")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "json", "--out-json", outputPath})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestParseAnalyzeAllowsNewOutSARIFInWritableDir(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := filepath.Join(t.TempDir(), "new.sarif")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "sarif", "--out-sarif", outputPath})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestParseAnalyzeRejectsReadOnlyOutSARIFFile(t *testing.T) {
	t.Parallel()

	configPath := writeTempConfig(t)
	outputPath := readOnlyFile(t, "scan.sarif")

	_, err := cli.ParseAnalyzeArgs([]string{"--config", configPath, "--format", "sarif", "--out-sarif", outputPath})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func writeTempConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(path, []byte("rules: []"), 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	return path
}

func readOnlyDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("failed to set read-only permissions: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})

	return dir
}

func writableFile(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	return path
}

func readOnlyFile(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatalf("failed to set read-only permissions: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(path, 0o600)
	})

	return path
}
