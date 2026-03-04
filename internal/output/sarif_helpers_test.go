//nolint:testpackage
package output

import (
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	t.Parallel()

	if got := normalizePath("./path/to/file.go"); got != "path/to/file.go" {
		t.Fatalf("unexpected path: %s", got)
	}
	if got := normalizePath(filepath.Join("path", "to", "file.go")); got != "path/to/file.go" {
		t.Fatalf("unexpected path: %s", got)
	}
}

func TestRuleIDForIndex(t *testing.T) {
	t.Parallel()

	if got := ruleIDForIndex(0); got != "RC0001" {
		t.Fatalf("unexpected rule id: %s", got)
	}
	if got := ruleIDForIndex(-1); got != "RC0000" {
		t.Fatalf("unexpected rule id: %s", got)
	}
}

func TestFormatRuleIndexClamp(t *testing.T) {
	t.Parallel()

	if got := formatRuleIndex(10000); got != "9999" {
		t.Fatalf("unexpected clamped index: %s", got)
	}
}

func TestSarifLevelMapping(t *testing.T) {
	t.Parallel()

	if sarifLevel("notice") != "note" {
		t.Fatal("expected notice to map to note")
	}
	if sarifLevel("info") != "note" {
		t.Fatal("expected info to map to note")
	}
	if sarifLevel("warning") != "warning" {
		t.Fatal("expected warning to map to warning")
	}
}

func TestMatchTextRuneLength(t *testing.T) {
	t.Parallel()

	if matchTextRuneLength("") != 0 {
		t.Fatal("expected empty match length to be 0")
	}
	if matchTextRuneLength("aß") != 2 {
		t.Fatal("expected rune length 2")
	}
}

func TestFormatRuleIndexPadsEachRange(t *testing.T) {
	t.Parallel()

	if got := formatRuleIndex(0); got != "0000" {
		t.Fatalf("unexpected padding for 0: %s", got)
	}
	if got := formatRuleIndex(9); got != "0009" {
		t.Fatalf("unexpected padding for 9: %s", got)
	}
	if got := formatRuleIndex(10); got != "0010" {
		t.Fatalf("unexpected padding for 10: %s", got)
	}
	if got := formatRuleIndex(99); got != "0099" {
		t.Fatalf("unexpected padding for 99: %s", got)
	}
	if got := formatRuleIndex(100); got != "0100" {
		t.Fatalf("unexpected padding for 100: %s", got)
	}
}
