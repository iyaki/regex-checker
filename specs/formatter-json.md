# JSON Formatter

Status: Partial

## Overview

### Purpose

- Provide machine-readable scan results for CI and automation.
- Preserve full match detail and scan stats.
- Follow shared formatter guidelines in `specs/formatter.md`.

### Goals

- Stable schema with explicit field types.
- Deterministic ordering of matches.
- JSON output that round-trips into `ScanResult` without loss.

### Non-Goals

- Streaming output during scanning.
- Compatibility with SARIF (see SARIF formatter spec).

### Scope

- Render a `ScanResult` into a JSON document.
- Output to file or stdout depending on CLI flags.

## Architecture

### Module/package layout (tree format)

```
internal/
  output/
    json.go
```

### Component diagram (ASCII)

```
[ScanResult] -> [JSON Formatter] -> stdout | file
```

### Data flow summary

1. Receive `ScanResult` from the scan service.
2. Sort matches deterministically (see Ordering).
3. Build JSON document from matches + stats.
4. Write JSON to stdout or `--out-json`.

## Data model

### Core Entities

JSONResult

- Definition: The JSON root object written by the formatter.
- Fields:
  - `schemaVersion` (int, required): Current version, `1`.
  - `matches` (array of JSONMatch, required)
  - `stats` (JSONStats, required)

JSONMatch

- Definition: A single match in JSON output.
- Fields:
  - `message` (string, required)
  - `severity` (string, required): `error|warning|notice|info`
  - `filePath` (string, required)
  - `absolutePath` (string, required): Absolute file path with `:<line>` appended.
  - `line` (int, required, 1-based)
  - `column` (int, required, 1-based, rune index)
  - `matchText` (string, required): Full matched substring (`$0`).

JSONStats

- Definition: Scan statistics.
- Fields:
  - `filesScanned` (int, required)
  - `filesSkipped` (int, required)
  - `matches` (int, required)
  - `durationMs` (int64, required)

### Relationships

- JSON entities map 1:1 to `ScanResult`, `Match`, and `ScanStats` from `specs/data-model.md`.
- Severity values follow `specs/regex-rules.md`.

### Persistence Notes

- JSON output is written to stdout or a file; no persistence beyond that.

## Workflows

### Render JSON (happy path)

1. Sort matches (Ordering).
2. Build `JSONResult` with `schemaVersion = 1`.
3. Marshal to JSON with stable key order.
4. Write output.

### No matches

- `matches` is an empty array.
- `stats.matches` is `0`.

### Error cases

- If `--format` includes `json` and `--out-json` is set, write to that file.
- If `--format` is only `json` and `--out-json` is not set, write to stdout.
- If `--format` includes `json` and another formatter, `--out-json` is required; otherwise return an error.
- Any file write or marshal error aborts the run with exit code 1.

## APIs

- Internal writer interface only. No network APIs.
- Suggested signature: `WriteJSON(result ScanResult, out io.Writer) error`.

## Client SDK Design

- No client SDK. Formatter is internal only.

## Configuration

- CLI flag: `--format json`.
- Optional output path: `--out-json <path>`.

## Permissions

- No permissions or authentication.

## Security Considerations

- JSON includes `matchText`, which may contain sensitive data.
- JSON includes absolute file paths, which may reveal local filesystem layout.
- If output is written to disk, follow least-privilege filesystem permissions.

## Dependencies

- Standard library JSON encoder.

## Verifications

- JSON validates against the schema described in this spec.
- Ordering is deterministic across runs with identical inputs.
- When `--format` includes `json` with another formatter and `--out-json` is missing, the command fails.
- During tests/QA, validate that every `absolutePath` refers to an existing file; the CLI does not validate at runtime.

## Appendices

### Ordering

- Sort by `filePath` (ascending, byte-wise).
- Within a file, sort by `line` then `column`.
- If ties remain, sort by severity order `error > warning > notice > info`.
- If ties remain, sort by `message` (ascending, byte-wise).

### Output example

```json
{
	"schemaVersion": 1,
	"matches": [
		{
			"message": "Avoid hardcoded token: abc123",
			"severity": "error",
			"filePath": "src/auth/token.go",
			"absolutePath": "/abs/src/auth/token.go:12",
			"line": 12,
			"column": 5,
			"matchText": "token=abc123"
		}
	],
	"stats": {
		"filesScanned": 10,
		"filesSkipped": 1,
		"matches": 1,
		"durationMs": 120
	}
}
```
