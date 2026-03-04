// Package config loads and validates rule configuration.
package config

// RuleSet represents the top-level YAML configuration.
type RuleSet struct {
	Rules       []Rule   `yaml:"rules"`
	Include     []string `yaml:"include,omitempty"`
	Exclude     []string `yaml:"exclude,omitempty"`
	FailOn      *string  `yaml:"failOn,omitempty"`
	Concurrency *int     `yaml:"concurrency,omitempty"`
}
