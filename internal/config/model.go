// Package config loads and validates rule configuration.
package config

// RuleSet represents the top-level YAML configuration.
type RuleSet struct {
	Rules                []Rule     `yaml:"rules"`
	Include              StringList `yaml:"include,omitempty"`
	Exclude              StringList `yaml:"exclude,omitempty"`
	FailOn               *string    `yaml:"failOn,omitempty"`
	Concurrency          *int       `yaml:"concurrency,omitempty"`
	Baseline             *string    `yaml:"baseline,omitempty"`
	ConsoleColorsEnabled *bool      `yaml:"consoleColorsEnabled,omitempty"`
	IgnoreFilesEnabled   *bool      `yaml:"ignoreFilesEnabled,omitempty"`
	IgnoreFiles          StringList `yaml:"ignoreFiles,omitempty"`
}
