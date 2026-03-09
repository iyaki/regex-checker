package config

import (
	"runtime"

	"github.com/iyaki/reglint/internal/rules"
)

// Rule represents a single regex rule entry.
type Rule struct {
	Message  string     `yaml:"message"`
	Regex    string     `yaml:"regex"`
	Severity string     `yaml:"severity,omitempty"`
	Paths    StringList `yaml:"paths,omitempty"`
	Exclude  StringList `yaml:"exclude,omitempty"`
}

func (r Rule) toRulesRule() rules.Rule {
	severity := r.Severity
	if severity == "" {
		severity = "warning"
	}

	paths := append([]string{}, []string(r.Paths)...)
	if len(paths) == 0 {
		paths = []string{"**/*"}
	}

	exclude := append([]string{}, []string(r.Exclude)...)

	return rules.Rule{
		Message:  r.Message,
		Regex:    r.Regex,
		Severity: severity,
		Paths:    paths,
		Exclude:  exclude,
	}
}

// ToRules converts the parsed config into shared rule models.
func (ruleSet RuleSet) ToRules() rules.RuleSet {
	var baseline *string
	if ruleSet.Baseline != nil {
		baselineValue := *ruleSet.Baseline
		baseline = &baselineValue
	}

	converted := rules.RuleSet{
		Include:              append([]string{}, []string(ruleSet.Include)...),
		Exclude:              append([]string{}, []string(ruleSet.Exclude)...),
		FailOn:               ruleSet.FailOn,
		Concurrency:          ruleSet.Concurrency,
		Baseline:             baseline,
		ConsoleColorsEnabled: ruleSet.ConsoleColorsEnabled,
		IgnoreFilesEnabled:   ruleSet.IgnoreFilesEnabled,
		IgnoreFiles:          append([]string{}, []string(ruleSet.IgnoreFiles)...),
	}
	if len(converted.Include) == 0 {
		converted.Include = []string{"**/*"}
	}
	if len(converted.Exclude) == 0 {
		converted.Exclude = []string{"**/.git/**", "**/node_modules/**", "**/vendor/**"}
	}
	if converted.Concurrency == nil {
		concurrency := runtime.GOMAXPROCS(0)
		converted.Concurrency = &concurrency
	}

	if len(ruleSet.Rules) > 0 {
		converted.Rules = make([]rules.Rule, len(ruleSet.Rules))
		for i, rule := range ruleSet.Rules {
			convertedRule := rule.toRulesRule()
			convertedRule.Index = i
			if len(rule.Paths) == 0 {
				convertedRule.Paths = append([]string{}, converted.Include...)
			}
			if len(rule.Exclude) == 0 {
				convertedRule.Exclude = append([]string{}, converted.Exclude...)
			}
			converted.Rules[i] = convertedRule
		}
	}

	return converted
}
