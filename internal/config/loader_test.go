package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/regex-checker/internal/config"
)

func TestLoadRuleSetRejectsEmptyRules(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules: []\n")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "rules must be") {
		t.Fatalf("expected rules error, got %v", err)
	}
}

func TestLoadRuleSetRejectsInvalidYAML(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules: [")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRuleSetParsesConfig(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules:\n  - message: 'hello'\n    regex: 'world'\n")

	rules, err := config.LoadRuleSet(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(rules.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules.Rules))
	}
	if rules.Rules[0].Message != "hello" || rules.Rules[0].Regex != "world" {
		t.Fatalf("unexpected rule contents: %+v", rules.Rules[0])
	}
}

func TestLoadRuleSetAllowsFailOnAndConcurrency(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "failOn: 'warning'\nconcurrency: 2\nrules:\n  - message: 'hello'\n    regex: 'world'\n")

	_, err := config.LoadRuleSet(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestLoadRuleSetAllowsRuleSeverity(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules:\n  - message: 'hello'\n    regex: 'world'\n    severity: 'info'\n")

	_, err := config.LoadRuleSet(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestLoadRuleSetRejectsRuleMissingMessage(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules:\n  - regex: 'world'\n")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRuleSetRejectsRuleMissingRegex(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules:\n  - message: 'hello'\n")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRuleSetRejectsInvalidFailOn(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "failOn: 'fatal'\nrules:\n  - message: 'hello'\n    regex: 'world'\n")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRuleSetRejectsNonPositiveConcurrency(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "concurrency: 0\nrules:\n  - message: 'hello'\n    regex: 'world'\n")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRuleSetRejectsInvalidRuleSeverity(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules:\n  - message: 'hello'\n    regex: 'world'\n    severity: 'fatal'\n")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRuleSetRejectsInvalidRegex(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules:\n  - message: 'hello'\n    regex: '(['\n")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRuleSetRejectsBlankRuleMessage(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "rules:\n  - message: '  '\n    regex: 'world'\n")

	_, err := config.LoadRuleSet(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRuleSetToRulesCopiesSlices(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		Include: []string{"src/**"},
		Exclude: []string{"vendor/**"},
		Rules: []config.Rule{
			{
				Message:  "hello",
				Regex:    "world",
				Severity: "warning",
				Paths:    []string{"src/**"},
				Exclude:  []string{"src/vendor/**"},
			},
		},
	}

	converted := ruleSet.ToRules()
	converted.Include[0] = "changed/**"
	converted.Exclude[0] = "changed/**"
	converted.Rules[0].Paths[0] = "changed/**"
	converted.Rules[0].Exclude[0] = "changed/**"

	if ruleSet.Include[0] == "changed/**" || ruleSet.Exclude[0] == "changed/**" {
		t.Fatal("expected include/exclude to be copied")
	}
	if ruleSet.Rules[0].Paths[0] == "changed/**" || ruleSet.Rules[0].Exclude[0] == "changed/**" {
		t.Fatal("expected rule paths/exclude to be copied")
	}
}

func TestRuleSetToRulesDefaultsSeverity(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		Rules: []config.Rule{
			{
				Message: "hello",
				Regex:   "world",
			},
		},
	}

	converted := ruleSet.ToRules()
	if converted.Rules[0].Severity != "warning" {
		t.Fatalf("expected default severity warning, got %q", converted.Rules[0].Severity)
	}
}

func TestRuleSetToRulesDefaultsRulePathsAndExclude(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		Rules: []config.Rule{
			{
				Message: "hello",
				Regex:   "world",
			},
		},
	}

	converted := ruleSet.ToRules()
	if len(converted.Rules[0].Paths) != 1 || converted.Rules[0].Paths[0] != "**/*" {
		t.Fatalf("expected rule paths to default to ruleset include, got %v", converted.Rules[0].Paths)
	}
	if len(converted.Rules[0].Exclude) != 3 {
		t.Fatalf("expected rule exclude to default, got %v", converted.Rules[0].Exclude)
	}
}

func TestRuleSetToRulesDefaultsConcurrency(t *testing.T) {
	t.Parallel()

	ruleSet := config.RuleSet{
		Rules: []config.Rule{{Message: "hello", Regex: "world"}},
	}

	converted := ruleSet.ToRules()
	if converted.Concurrency == nil {
		t.Fatal("expected concurrency default, got nil")
	}
}

func writeConfigFile(t *testing.T, contents string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return path
}
