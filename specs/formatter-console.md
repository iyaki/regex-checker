# Console Formatter

Status: Proposed

## Overview

### Purpose

- Provide human-readable scan results on stdout for local usage.
- Keep output stable and grep-friendly across platforms.
- Follow shared formatter guidelines in `specs/formatter.md`.

### Goals

- Deterministic ordering of files and matches.
- Clear severity labeling and location formatting.
- Single-line summary with scan stats.

### Non-Goals

- Colored or interactive output.
- Machine-optimized formats (see JSON/SARIF formatter specs).
- Streaming output per file while scanning (render after scan completes).

### Scope

- Render a `ScanResult` to stdout when `--format console` is selected.
- Use only data already present in `ScanResult`.

## Architecture

### Module/package layout (tree format)

```
internal/
  output/
    console.go
```

### Component diagram (ASCII)

```
[ScanResult] -> [Console Formatter] -> stdout
```

### Data flow summary

1. Receive `ScanResult` from the scan service.
2. Sort matches deterministically (see Ordering).
3. Group matches by relative `filePath`.
4. Render each match as a two-line block.
5. Render the summary line.

## Data model

### Core Entities

ConsoleMatchLine

- Definition: A rendered line for a single match.
- Fields:
  - `filePath` (string): Relative file path used for ordering.
  - `severityLabel` (string): One of `ERROR|WARN|NOTICE|INFO`.
  - `line` (int): 1-based line number.
  - `column` (int): 1-based column number (rune index).
  - `message` (string): Interpolated match message.
  - `absolutePath` (string): Absolute file path printed on the line below the finding as `<absolutePath>:<line>`.

ConsoleSummary

- Definition: Summary of the scan.
- Fields:
  - `filesScanned` (int)
  - `filesSkipped` (int)
  - `matches` (int)
  - `durationMs` (int64)

### Relationships

- `ConsoleMatchLine` derives from `Match` in `specs/data-model.md`.
- `ConsoleSummary` derives from `ScanStats` in `specs/data-model.md`.

### Persistence Notes

- No persistence. Output is written to stdout.

## Workflows

### Render matches (happy path)

1. Sort matches (Ordering).
2. For each `filePath`, print the file header line.
3. For each match, print a bullet-prefixed finding line.
4. Print the absolute path line with `:<line>` on the next line.
5. Print a blank line between findings.
6. After all findings, print the summary line.

### No matches

1. Print `No matches found.`
2. Print the summary line.

## APIs

- Internal writer interface only. No network APIs.
- Suggested signature: `WriteConsole(result ScanResult, out io.Writer) error`.

## Client SDK Design

- No client SDK. Formatter is internal only.

## Configuration

- CLI flag: `--format console` (default).
- Output destination: stdout only.

## Permissions

- No permissions or authentication.

## Security Considerations

- Output contains interpolated messages that may include sensitive data.
- Do not emit raw `matchText` in console output.
- Output includes absolute file paths, which may reveal local filesystem layout.

## Dependencies

- Standard library only.

## Verifications

- Console output is deterministic across runs with identical inputs.
- `No matches found.` appears when `matches == 0`.
- Summary line includes files scanned, skipped, total matches, and duration.
- During tests/QA, validate that every `absolutePath` refers to an existing file; the CLI does not validate at runtime.

## Appendices

### Ordering

- Sort by `filePath` (ascending, byte-wise).
- Within a file, sort by `line` then `column`.
- If ties remain, sort by severity order `error > warning > notice > info`.
- If ties remain, sort by `message` (ascending, byte-wise).

### Output format

```
- ERROR 12:5 Avoid hardcoded token: abc123
  /abs/path/to/file.ext:12

- WARN  42:1 Unexpected debug flag
  /abs/path/to/file.ext:42

- NOTICE 3:9 Use of TODO comment
  /abs/another/file.go:3

Summary: files=10 skipped=1 matches=3 durationMs=120
```
