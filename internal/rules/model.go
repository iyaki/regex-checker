// Package rules defines shared rule data models.
package rules

import "strings"

const digitBase = 10

// RuleSet represents the top-level rules configuration.
type RuleSet struct {
	Rules                []Rule
	Include              []string
	Exclude              []string
	FailOn               *string
	Concurrency          *int
	Baseline             *string
	ConsoleColorsEnabled *bool
	IgnoreFilesEnabled   *bool
	IgnoreFiles          []string
}

// Rule represents a single regex rule entry.
type Rule struct {
	Message  string
	Regex    string
	Severity string
	Paths    []string
	Exclude  []string
	Index    int
}

// InterpolateMessage replaces $0, $1, ... with regex capture groups.
// $$ emits a literal $. Missing captures resolve to empty string.
func InterpolateMessage(message string, captures []string) string {
	if message == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(message))
	for i := 0; i < len(message); i++ {
		if message[i] != '$' {
			builder.WriteByte(message[i])

			continue
		}
		if i+1 >= len(message) {
			builder.WriteByte('$')

			continue
		}

		next := message[i+1]
		if next == '$' {
			builder.WriteByte('$')
			i++

			continue
		}
		if next < '0' || next > '9' {
			builder.WriteByte('$')

			continue
		}

		index, consumed := parseDigits(message[i+1:])
		i += consumed

		if index < len(captures) {
			builder.WriteString(captures[index])
		}
	}

	return builder.String()
}

func parseDigits(value string) (int, int) {
	index := 0
	consumed := 0
	for consumed < len(value) {
		digit := value[consumed]
		if digit < '0' || digit > '9' {
			break
		}
		index = index*digitBase + int(digit-'0')
		consumed++
	}

	return index, consumed
}
