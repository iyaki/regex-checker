package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadRuleSet reads and validates a YAML rules configuration.
func LoadRuleSet(path string) (RuleSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RuleSet{}, fmt.Errorf("read config: %w", err)
	}

	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return RuleSet{}, fmt.Errorf("parse config: %w", err)
	}

	if err := validateRuleSet(ruleSet); err != nil {
		return RuleSet{}, err
	}

	return ruleSet, nil
}

func validateRuleSet(ruleSet RuleSet) error {
	if err := validateRuleSetFields(ruleSet); err != nil {
		return err
	}

	return validateRules(ruleSet.Rules)
}

func validateRuleSetFields(ruleSet RuleSet) error {
	if len(ruleSet.Rules) == 0 {
		return fmt.Errorf("rules must be a non-empty list")
	}
	if ruleSet.FailOn != nil && !isValidSeverity(*ruleSet.FailOn) {
		return fmt.Errorf("invalid failOn value: %s", *ruleSet.FailOn)
	}
	if ruleSet.Concurrency != nil && *ruleSet.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive")
	}

	return nil
}

func validateRules(rules []Rule) error {
	for i, rule := range rules {
		if err := validateRule(rule, i); err != nil {
			return err
		}
	}

	return nil
}

func validateRule(rule Rule, index int) error {
	if strings.TrimSpace(rule.Message) == "" {
		return fmt.Errorf("rule %d message is required", index+1)
	}
	if strings.TrimSpace(rule.Regex) == "" {
		return fmt.Errorf("rule %d regex is required", index+1)
	}
	if _, err := regexp.Compile(rule.Regex); err != nil {
		return fmt.Errorf("rule %d regex is invalid: %w", index+1, err)
	}
	if rule.Severity != "" && !isValidSeverity(rule.Severity) {
		return fmt.Errorf("rule %d has invalid severity: %s", index+1, rule.Severity)
	}

	return nil
}

func isValidSeverity(value string) bool {
	switch value {
	case "error", "warning", "notice", "info":
		return true
	default:
		return false
	}
}
