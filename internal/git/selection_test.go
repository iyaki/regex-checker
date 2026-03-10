//nolint:testpackage
package git

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSelectCandidateFilesModeOffIsNoOp(t *testing.T) {
	runCommandCalls := 0
	setSelectionRunCommandHook(t, func(string, ...string) (string, error) {
		runCommandCalls++

		return "", nil
	})

	files, err := SelectCandidateFiles(CandidateSelectionRequest{Mode: "off", WorkingDir: "."})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files, got %v", files)
	}
	if runCommandCalls != 0 {
		t.Fatalf("expected no git command calls, got %d", runCommandCalls)
	}
}

func TestSelectCandidateFilesStagedUsesCachedDiffAndNormalizesPaths(t *testing.T) {
	var gotDir string
	var gotArgs []string
	setSelectionRunCommandHook(t, func(dir string, args ...string) (string, error) {
		gotDir = dir
		gotArgs = append([]string{}, args...)

		return "./pkg/beta.go\x00pkg/alpha.go\x00pkg\\gamma.go\x00pkg/alpha.go\x00", nil
	})

	files, err := SelectCandidateFiles(CandidateSelectionRequest{Mode: "staged", WorkingDir: "/tmp/repo"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotDir != "/tmp/repo" {
		t.Fatalf("expected command dir %q, got %q", "/tmp/repo", gotDir)
	}
	wantArgs := []string{"diff", "--cached", "--name-only", "--diff-filter=ACMR", "-z"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("expected args %v, got %v", wantArgs, gotArgs)
	}
	wantFiles := []string{"pkg/alpha.go", "pkg/beta.go", "pkg/gamma.go"}
	if !reflect.DeepEqual(files, wantFiles) {
		t.Fatalf("expected files %v, got %v", wantFiles, files)
	}
}

func TestSelectCandidateFilesDiffRequiresTarget(t *testing.T) {
	setSelectionRunCommandHook(t, func(string, ...string) (string, error) {
		t.Fatal("expected git command not to be invoked")

		return "", nil
	})

	_, err := SelectCandidateFiles(CandidateSelectionRequest{Mode: "diff", WorkingDir: "."})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "git mode diff requires diff target" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSelectCandidateFilesDiffUsesTarget(t *testing.T) {
	var gotArgs []string
	setSelectionRunCommandHook(t, func(_ string, args ...string) (string, error) {
		gotArgs = append([]string{}, args...)

		return "pkg/zeta.go\npkg/alpha.go\n", nil
	})

	files, err := SelectCandidateFiles(CandidateSelectionRequest{
		Mode:       "diff",
		DiffTarget: "HEAD~1..HEAD",
		WorkingDir: "/repo",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	wantArgs := []string{"diff", "--name-only", "--diff-filter=ACMR", "-z", "HEAD~1..HEAD"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("expected args %v, got %v", wantArgs, gotArgs)
	}
	wantFiles := []string{"pkg/alpha.go", "pkg/zeta.go"}
	if !reflect.DeepEqual(files, wantFiles) {
		t.Fatalf("expected files %v, got %v", wantFiles, files)
	}
}

func TestSelectCandidateFilesRejectsPathOutsideRepositoryRoot(t *testing.T) {
	setSelectionRunCommandHook(t, func(string, ...string) (string, error) {
		return "../outside.go\x00", nil
	})

	_, err := SelectCandidateFiles(CandidateSelectionRequest{Mode: "staged", WorkingDir: "."})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "non-root-relative") {
		t.Fatalf("expected root-relative validation error, got %v", err)
	}
}

func TestSelectCandidateFilesReturnsModeSpecificCommandErrors(t *testing.T) {
	setSelectionRunCommandHook(t, func(string, ...string) (string, error) {
		return "", errors.New("fatal")
	})

	_, err := SelectCandidateFiles(CandidateSelectionRequest{Mode: "staged", WorkingDir: "."})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "git mode staged failed to resolve changed files" {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = SelectCandidateFiles(CandidateSelectionRequest{
		Mode:       "diff",
		DiffTarget: "HEAD",
		WorkingDir: ".",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "git mode diff failed to resolve changed files for target \"HEAD\"" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func setSelectionRunCommandHook(t *testing.T, runCommandHook func(string, ...string) (string, error)) {
	t.Helper()

	originalRunCommand := runCommand
	runCommand = runCommandHook
	t.Cleanup(func() {
		runCommand = originalRunCommand
	})
}
