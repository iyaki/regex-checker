package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// StringList enforces lists of strings in YAML.
type StringList []string

// UnmarshalYAML ensures the value is a list of string scalars.
func (s *StringList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == 0 {
		return nil
	}
	if value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
		return nil
	}
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("expected a list of strings")
	}

	items := make([]string, len(value.Content))
	for i, node := range value.Content {
		if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
			return fmt.Errorf("expected a list of strings")
		}
		items[i] = node.Value
	}

	*s = items

	return nil
}
