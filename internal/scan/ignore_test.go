//nolint:testpackage
package scan

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestCollectEntriesAppliesIgnoreRules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFileWithContent(t, filepath.Join(root, ".ignore"), "skip.txt\n")
	writeFileWithContent(t, filepath.Join(root, "keep.txt"), "keep")
	writeFileWithContent(t, filepath.Join(root, "skip.txt"), "skip")

	request := Request{
		Roots:            []string{root},
		Include:          []string{"**/*.txt"},
		Exclude:          nil,
		Ignore:           IgnoreSettings{Enabled: true, Files: []string{".ignore"}},
		MaxFileSizeBytes: 1024,
		Concurrency:      1,
	}

	entries, skipped, _, _, err := collectScanEntries(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skipped != 0 {
		t.Fatalf("expected 0 skipped files, got %d", skipped)
	}

	paths := entryPaths(entries)
	if !reflect.DeepEqual(paths, []string{"keep.txt"}) {
		t.Fatalf("expected keep.txt only, got %v", paths)
	}
}

func TestCollectEntriesAllowsIgnoreNegation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFileWithContent(t, filepath.Join(root, ".ignore"), "generated/**\n!generated/keep.txt\n")
	writeFileWithContent(t, filepath.Join(root, "generated", "keep.txt"), "keep")
	writeFileWithContent(t, filepath.Join(root, "generated", "skip.txt"), "skip")

	request := Request{
		Roots:            []string{root},
		Include:          []string{"**/*.txt"},
		Exclude:          nil,
		Ignore:           IgnoreSettings{Enabled: true, Files: []string{".ignore"}},
		MaxFileSizeBytes: 1024,
		Concurrency:      1,
	}

	entries, err := collectEntriesForTest(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := entryPaths(entries)
	if !reflect.DeepEqual(paths, []string{"generated/keep.txt"}) {
		t.Fatalf("expected keep.txt only, got %v", paths)
	}
}

func TestCollectEntriesReturnsIgnoreLoadErrors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFileWithContent(t, filepath.Join(root, ".ignore"), "[\n")

	request := Request{
		Roots:            []string{root},
		Include:          []string{"**/*.txt"},
		Exclude:          nil,
		Ignore:           IgnoreSettings{Enabled: true, Files: []string{".ignore"}},
		MaxFileSizeBytes: 1024,
		Concurrency:      1,
	}

	_, err := collectEntriesForTest(request)
	if err == nil {
		t.Fatal("expected error for invalid ignore pattern")
	}
}

func TestCollectEntriesIgnoreNegationDoesNotOverrideExclude(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFileWithContent(t, filepath.Join(root, ".ignore"), "!generated/keep.txt\n")
	writeFileWithContent(t, filepath.Join(root, "generated", "keep.txt"), "keep")

	request := Request{
		Roots:            []string{root},
		Include:          []string{"**/*.txt"},
		Exclude:          []string{"generated/**"},
		Ignore:           IgnoreSettings{Enabled: true, Files: []string{".ignore"}},
		MaxFileSizeBytes: 1024,
		Concurrency:      1,
	}

	entries, err := collectEntriesForTest(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no entries, got %d", len(entries))
	}
}

func TestLoadIgnoreRulesDisabledReturnsNil(t *testing.T) {
	t.Parallel()

	request := Request{
		Roots:            []string{t.TempDir()},
		Include:          []string{"**/*"},
		Exclude:          nil,
		Ignore:           IgnoreSettings{Enabled: false, Files: []string{".ignore"}},
		MaxFileSizeBytes: 1024,
		Concurrency:      1,
	}

	rulesByRoot, err := loadIgnoreRules(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rulesByRoot != nil {
		t.Fatal("expected nil rules when ignore disabled")
	}
}

func entryPaths(entries []fileEntry) []string {
	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.relPath
	}

	return paths
}

func collectEntriesForTest(request Request) ([]fileEntry, error) {
	entries, skipped, include, exclude, err := collectScanEntries(request)
	_ = skipped
	_ = include
	_ = exclude

	return entries, err
}
