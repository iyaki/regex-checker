// Package output provides scan result formatters.
package output

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/iyaki/regex-checker/internal/scan"
)

// WriteConsole renders a scan result to the provided writer.
func WriteConsole(result scan.Result, out io.Writer) error {
	matches := append([]scan.Match{}, result.Matches...)
	sort.Slice(matches, func(i, j int) bool {
		left := matches[i]
		right := matches[j]
		if left.FilePath != right.FilePath {
			return left.FilePath < right.FilePath
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Column != right.Column {
			return left.Column < right.Column
		}
		if left.Severity != right.Severity {
			return severityRank(left.Severity) < severityRank(right.Severity)
		}

		return left.Message < right.Message
	})

	var builder strings.Builder
	if err := appendConsoleMatches(&builder, matches); err != nil {
		return err
	}

	builder.WriteString(fmt.Sprintf("Summary: files=%d skipped=%d matches=%d durationMs=%d\n",
		result.Stats.FilesScanned,
		result.Stats.FilesSkipped,
		result.Stats.Matches,
		result.Stats.DurationMs,
	))

	_, err := io.WriteString(out, builder.String())

	return err
}

func appendConsoleMatches(builder *strings.Builder, matches []scan.Match) error {
	if len(matches) == 0 {
		builder.WriteString("No matches found.\n")

		return nil
	}

	currentFile := ""
	for _, match := range matches {
		if match.FilePath != currentFile {
			if currentFile != "" {
				builder.WriteString("\n")
			}
			currentFile = match.FilePath
			builder.WriteString(match.FilePath)
			builder.WriteString("\n")
		}
		line, err := formatConsoleMatchLine(match)
		if err != nil {
			return err
		}
		builder.WriteString(line)
	}
	builder.WriteString("\n")

	return nil
}

func formatConsoleMatchLine(match scan.Match) (string, error) {
	fileURI, err := fileURIWithLine(match.FilePath, match.Line)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"  %-5s %d:%d %s %s\n",
		severityLabel(match.Severity),
		match.Line,
		match.Column,
		match.Message,
		fileURI,
	), nil
}

const (
	severityRankError = iota
	severityRankWarning
	severityRankNotice
	severityRankInfo
	severityRankUnknown
)

func severityRank(value string) int {
	switch value {
	case "error":
		return severityRankError
	case "warning":
		return severityRankWarning
	case "notice":
		return severityRankNotice
	case "info":
		return severityRankInfo
	default:
		return severityRankUnknown
	}
}

func severityLabel(value string) string {
	switch value {
	case "error":
		return "ERROR"
	case "warning":
		return "WARN"
	case "notice":
		return "NOTICE"
	case "info":
		return "INFO"
	default:
		return strings.ToUpper(value)
	}
}
