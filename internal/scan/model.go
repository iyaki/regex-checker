// Package scan defines scan request and result models.
package scan

import "github.com/iyaki/regex-checker/internal/rules"

// Request defines the input to the scan service.
type Request struct {
	Roots            []string
	Rules            []rules.Rule
	Include          []string
	Exclude          []string
	MaxFileSizeBytes int64
	Concurrency      int
}

// Match represents a single rule match.
type Match struct {
	Message   string
	Severity  string
	FilePath  string
	Root      string `json:"-"`
	Line      int
	Column    int
	MatchText string
	RuleIndex int `json:"-"`
}

// Stats captures aggregate scan statistics.
type Stats struct {
	FilesScanned int
	FilesSkipped int
	Matches      int
	DurationMs   int64
}

// Result aggregates matches and stats.
type Result struct {
	Matches []Match
	Stats   Stats
}
