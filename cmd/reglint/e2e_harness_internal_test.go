package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type e2EHarness struct {
	binaryPath string
}

type e2EProcessResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

var (
	e2eBinaryBuildOnce sync.Once
	e2eBinaryPath      string
	e2eBinaryBuildErr  error
	e2eBinaryBuilds    atomic.Int32
)

func newE2EHarness(t *testing.T) *e2EHarness {
	t.Helper()

	binaryPath, err := ensureE2EBinaryBuilt()
	if err != nil {
		t.Fatalf("build e2e binary: %v", err)
	}

	return &e2EHarness{binaryPath: binaryPath}
}

func (h *e2EHarness) run(workDir string, args []string, env map[string]string) (e2EProcessResult, error) {
	cmd := exec.Command(h.binaryPath, args...)
	cmd.Dir = workDir
	cmd.Env = mergeEnv(env)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := e2EProcessResult{
		ExitCode: 0,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()

		return result, nil
	}

	return e2EProcessResult{}, fmt.Errorf("execute binary %q: %w", h.binaryPath, err)
}

func e2eBinaryBuildInvocations() int {
	return int(e2eBinaryBuilds.Load())
}

func ensureE2EBinaryBuilt() (string, error) {
	e2eBinaryBuildOnce.Do(func() {
		e2eBinaryBuilds.Add(1)

		moduleRoot, err := findModuleRoot()
		if err != nil {
			e2eBinaryBuildErr = err

			return
		}

		outDir, err := os.MkdirTemp("", "reglint-e2e-bin-")
		if err != nil {
			e2eBinaryBuildErr = fmt.Errorf("create temp e2e build directory: %w", err)

			return
		}

		binaryName := "reglint-e2e"
		if runtime.GOOS == "windows" {
			binaryName += ".exe"
		}

		binaryPath := filepath.Join(outDir, binaryName)

		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/reglint")
		cmd.Dir = moduleRoot

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			e2eBinaryBuildErr = fmt.Errorf(
				"go build ./cmd/reglint failed: %w; stdout=%q stderr=%q",
				err,
				strings.TrimSpace(stdout.String()),
				strings.TrimSpace(stderr.String()),
			)

			return
		}

		e2eBinaryPath = binaryPath
	})

	if e2eBinaryBuildErr != nil {
		return "", e2eBinaryBuildErr
	}

	return e2eBinaryPath, nil
}

func findModuleRoot() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("resolve current file path")
	}

	dir := filepath.Dir(currentFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", currentFile)
		}

		dir = parent
	}
}

func mergeEnv(overrides map[string]string) []string {
	if len(overrides) == 0 {
		return os.Environ()
	}

	merged := map[string]string{}
	for _, pair := range os.Environ() {
		idx := strings.IndexByte(pair, '=')
		if idx < 0 {
			continue
		}

		merged[pair[:idx]] = pair[idx+1:]
	}

	for key, value := range overrides {
		merged[key] = value
	}

	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, key+"="+merged[key])
	}

	return env
}
