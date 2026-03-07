package ignore_test

import (
	"testing"

	"github.com/iyaki/reglint/internal/ignore"
)

func TestMatcherRespectsOrderingAndNegation(t *testing.T) {
	rules := []ignore.IgnoreRule{
		{BaseDir: ".", Pattern: "*.txt", Negated: false},
		{BaseDir: ".", Pattern: "keep.txt", Negated: true},
	}
	matcher := ignore.NewMatcher(rules)

	ignored, err := matcher.Ignored("keep.txt", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ignored {
		t.Fatal("expected keep.txt to be unignored")
	}

	ignored, err = matcher.Ignored("skip.txt", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ignored {
		t.Fatal("expected skip.txt to be ignored")
	}
}
