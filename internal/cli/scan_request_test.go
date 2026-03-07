package cli_test

import (
	"testing"

	"github.com/iyaki/reglint/internal/cli"
	"github.com/iyaki/reglint/internal/config"
	"github.com/iyaki/reglint/internal/rules"
)

func TestBuildScanRequestOverridesIncludeExcludeAndFailOn(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		Include:            []string{"src/**"},
		Exclude:            []string{"vendor/**"},
		FailOn:             stringPtr("warning"),
		IgnoreFilesEnabled: boolPtr(true),
		Rules: []config.Rule{
			{
				Message: "hello",
				Regex:   "world",
			},
		},
	}
	cfg := cli.Config{
		Roots:            []string{"."},
		Include:          []string{"**/*.go"},
		Exclude:          []string{"**/generated/**"},
		FailOnSeverity:   "error",
		NoIgnoreFiles:    true,
		Concurrency:      3,
		MaxFileSizeBytes: 99,
	}

	request, failOn := cli.BuildScanRequest(cfg, ruleSet)

	assertEqualString(t, "fail-on severity", failOn, "error")
	assertStringSlice(t, "include override", request.Include, []string{"**/*.go"})
	assertStringSlice(t, "exclude override", request.Exclude, []string{"**/generated/**"})
	assertRuleCount(t, request.Rules, 1)
	assertStringSlice(t, "rule paths override", request.Rules[0].Paths, []string{"**/*.go"})
	assertStringSlice(t, "rule exclude override", request.Rules[0].Exclude, []string{"**/generated/**"})
	assertEqualString(t, "ruleset include unchanged", ruleSet.Include[0], "src/**")
	assertEqualString(t, "ruleset exclude unchanged", ruleSet.Exclude[0], "vendor/**")
}

func TestBuildScanRequestUsesRuleSetDefaultsWithoutOverrides(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		Rules: []config.Rule{{Message: "hello", Regex: "world"}},
	}
	cfg := cli.Config{
		Roots:            []string{"./root"},
		Concurrency:      2,
		MaxFileSizeBytes: 10,
	}

	request, failOn := cli.BuildScanRequest(cfg, ruleSet)

	expectedExclude := []string{"**/.git/**", "**/node_modules/**", "**/vendor/**"}
	assertEqualString(t, "fail-on severity", failOn, "")
	assertStringSlice(t, "default include", request.Include, []string{"**/*"})
	assertStringSlice(t, "default exclude", request.Exclude, expectedExclude)
	assertRuleCount(t, request.Rules, 1)
	assertStringSlice(t, "default rule paths", request.Rules[0].Paths, []string{"**/*"})
	assertStringSlice(t, "default rule exclude", request.Rules[0].Exclude, expectedExclude)
}

func TestBuildScanRequestUsesRuleSetConcurrencyWhenNotSetInCLI(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		Concurrency: intPtr(7),
		Rules:       []config.Rule{{Message: "hello", Regex: "world"}},
	}
	cfg := cli.Config{
		Roots:            []string{"./root"},
		Concurrency:      2,
		MaxFileSizeBytes: 10,
		ConcurrencySet:   false,
	}

	request, _ := cli.BuildScanRequest(cfg, ruleSet)

	if request.Concurrency != 7 {
		t.Fatalf("expected concurrency 7, got %d", request.Concurrency)
	}
}

func TestBuildScanRequestUsesCLIConcurrencyWhenSet(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		Concurrency: intPtr(7),
		Rules:       []config.Rule{{Message: "hello", Regex: "world"}},
	}
	cfg := cli.Config{
		Roots:            []string{"./root"},
		Concurrency:      3,
		MaxFileSizeBytes: 10,
		ConcurrencySet:   true,
	}

	request, _ := cli.BuildScanRequest(cfg, ruleSet)

	if request.Concurrency != 3 {
		t.Fatalf("expected concurrency 3, got %d", request.Concurrency)
	}
}

func TestBuildScanRequestResolvesIgnoreSettings(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		IgnoreFilesEnabled: boolPtr(false),
		IgnoreFiles:        []string{".customignore"},
		Rules:              []config.Rule{{Message: "hello", Regex: "world"}},
	}
	cfg := cli.Config{
		Roots:            []string{"./root"},
		Concurrency:      1,
		MaxFileSizeBytes: 10,
		NoIgnoreFiles:    true,
	}

	request, _ := cli.BuildScanRequest(cfg, ruleSet)

	if request.Ignore.Enabled {
		t.Fatal("expected ignore files to be disabled")
	}
	assertStringSlice(t, "ignore files", request.Ignore.Files, []string{".customignore"})
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func assertRuleCount(t *testing.T, rules []rules.Rule, expected int) {
	t.Helper()

	if len(rules) != expected {
		t.Fatalf("expected %d rule(s), got %d", expected, len(rules))
	}
}

func assertStringSlice(t *testing.T, label string, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %s %v, got %v", label, want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %s %v, got %v", label, want, got)
		}
	}
}

func assertEqualString(t *testing.T, label, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("expected %s %q, got %q", label, want, got)
	}
}
