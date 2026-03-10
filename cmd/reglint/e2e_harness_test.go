package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EHarnessRunExecutesCompiledBinary(t *testing.T) {
	harness := newE2EHarness(t)

	workspace := t.TempDir()
	configPath := filepath.Join(workspace, "rules.yaml")

	result, err := harness.run(workspace, []string{"init", "--out", configPath}, nil)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stdout=%q stderr=%q", result.ExitCode, result.Stdout, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "Wrote default config to "+configPath) {
		t.Fatalf("expected init success output, got %q", result.Stdout)
	}
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file at %s: %v", configPath, err)
	}
}

func TestE2EHarnessBuildsBinaryOnlyOnce(t *testing.T) {
	harness1 := newE2EHarness(t)
	workspace1 := t.TempDir()
	path1 := filepath.Join(workspace1, "rules.yaml")

	result1, err := harness1.run(workspace1, []string{"init", "--out", path1}, nil)
	if err != nil {
		t.Fatalf("first run returned error: %v", err)
	}
	if result1.ExitCode != 0 {
		t.Fatalf("expected first exit code 0, got %d", result1.ExitCode)
	}

	buildCountAfterFirst := e2eBinaryBuildInvocations()

	harness2 := newE2EHarness(t)
	workspace2 := t.TempDir()
	path2 := filepath.Join(workspace2, "rules.yaml")

	result2, err := harness2.run(workspace2, []string{"init", "--out", path2}, nil)
	if err != nil {
		t.Fatalf("second run returned error: %v", err)
	}
	if result2.ExitCode != 0 {
		t.Fatalf("expected second exit code 0, got %d", result2.ExitCode)
	}

	buildCountAfterSecond := e2eBinaryBuildInvocations()
	if buildCountAfterSecond != buildCountAfterFirst {
		t.Fatalf("expected binary build count to stay at %d, got %d", buildCountAfterFirst, buildCountAfterSecond)
	}
}
