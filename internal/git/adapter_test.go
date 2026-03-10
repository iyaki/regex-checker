//nolint:testpackage
package git

import (
	"errors"
	"reflect"
	"testing"
)

func TestCheckCapabilitiesModeOffIsNoOp(t *testing.T) {
	lookPathCalls := 0
	runCommandCalls := 0
	setCapabilityTestHooks(
		t,
		func(string) (string, error) {
			lookPathCalls++

			return "/usr/bin/git", nil
		},
		func(string, ...string) (string, error) {
			runCommandCalls++

			return "true\n", nil
		},
	)

	err := CheckCapabilities(CapabilityRequest{Mode: "off", WorkingDir: "."})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if lookPathCalls != 0 {
		t.Fatalf("expected no git lookup, got %d calls", lookPathCalls)
	}
	if runCommandCalls != 0 {
		t.Fatalf("expected no git command calls, got %d", runCommandCalls)
	}
}

func TestCheckCapabilitiesRequiresGitExecutable(t *testing.T) {
	setCapabilityTestHooks(
		t,
		func(string) (string, error) {
			return "", errors.New("missing")
		},
		func(string, ...string) (string, error) {
			t.Fatal("expected git command not to be invoked")

			return "", nil
		},
	)

	err := CheckCapabilities(CapabilityRequest{Mode: "staged", WorkingDir: "."})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "git mode staged requires git executable" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckCapabilitiesRequiresRepositoryContext(t *testing.T) {
	setCapabilityTestHooks(
		t,
		func(string) (string, error) {
			return "/usr/bin/git", nil
		},
		func(string, ...string) (string, error) {
			return "", errors.New("fatal")
		},
	)

	err := CheckCapabilities(CapabilityRequest{Mode: "diff", WorkingDir: "."})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "git mode diff requires a git repository" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckCapabilitiesRequiresWorkTreeOutput(t *testing.T) {
	setCapabilityTestHooks(
		t,
		func(string) (string, error) {
			return "/usr/bin/git", nil
		},
		func(string, ...string) (string, error) {
			return "false\n", nil
		},
	)

	err := CheckCapabilities(CapabilityRequest{Mode: "staged", WorkingDir: "."})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "git mode staged requires a git repository" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckCapabilitiesUsesRequestedWorkingDir(t *testing.T) {
	var gotDir string
	var gotArgs []string
	setCapabilityTestHooks(
		t,
		func(string) (string, error) {
			return "/usr/bin/git", nil
		},
		func(dir string, args ...string) (string, error) {
			gotDir = dir
			gotArgs = append([]string{}, args...)

			return "true\n", nil
		},
	)

	err := CheckCapabilities(CapabilityRequest{Mode: "staged", WorkingDir: "/tmp/repo"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotDir != "/tmp/repo" {
		t.Fatalf("expected command dir %q, got %q", "/tmp/repo", gotDir)
	}
	wantArgs := []string{"rev-parse", "--is-inside-work-tree"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("expected args %v, got %v", wantArgs, gotArgs)
	}
}

func setCapabilityTestHooks(
	t *testing.T,
	lookPathHook func(string) (string, error),
	runCommandHook func(string, ...string) (string, error),
) {
	t.Helper()

	originalLookPath := lookPath
	originalRunCommand := runCommand
	lookPath = lookPathHook
	runCommand = runCommandHook
	t.Cleanup(func() {
		lookPath = originalLookPath
		runCommand = originalRunCommand
	})
}
