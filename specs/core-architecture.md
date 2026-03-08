# Core Architecture

## Overview

### Purpose

- Define the core architectural building blocks for the RegLint CLI.
- Establish clear module boundaries, data flow, and extension points.

### Goals

- Keep the scanning engine deterministic and reproducible across platforms.
- Make output generation pluggable across multiple formats.
- Separate configuration, scanning, and rendering concerns.
- Ensure the core can be reused by future integrations (e.g., reviewdog wrapper).

### Non-Goals

- No long-running daemon or network service.
- No persistence layer or database.
- No UI beyond CLI output.

### Scope

- Local filesystem scanning using YAML rules.
- CLI execution path only.

## Architecture

### Module/package layout (tree format)

```
cmd/
  reglint/
    main.go
internal/
  config/
    loader.go
  rules/
    model.go
  scan/
    service.go
    engine.go
    match.go
  output/
    console.go
    json.go
    sarif.go
  io/
    fs.go
```

### Component diagram (ASCII)

```
[CLI]
  | flags, args
  v
[Config Loader] -> [Rule Compiler]
  |                     |
  v                     v
[Scan Service] -> [Scan Engine] -> [File Walker]
  |
  v
[Output Writers] -> console | json | sarif
```

### Data flow summary

1. CLI parses flags and resolves scan roots.
2. YAML config is loaded and validated.
3. Rules compile to RE2 regexes with normalized severity.
4. ScanService builds a request and starts the scan.
5. Engine walks files, filters by globs, and matches rules.
6. Results are aggregated and rendered by output writers.

## Data model

### Core Entities

ScanRequest

- Definition: Input to the scan service.
- Fields: roots, rules, include/exclude, maxFileSizeBytes, concurrency, output formats.

Match

- Definition: Single rule match with location and severity.
- Fields: message, severity, filePath, line, column, matchText.

ScanResult

- Definition: Aggregated output for JSON/SARIF writers.
- Fields: matches, stats.

### Relationships

- CLI builds ScanRequest.
- ScanService produces ScanResult.
- Output writers consume ScanResult.

### Persistence Notes

- No persistence.

## Workflows

### CLI run (happy path)

1. User runs `reglint --rules <file> [path ...]`.
2. Config loader validates YAML and compiles rules.
3. ScanService scans files with engine.
4. Output writers render console and optional JSON/SARIF.
5. Exit code reflects `--fail-on` threshold.

### Error cases

- Invalid YAML or regex: fail fast with exit code 1.
- Output write failure: exit code 1.
- File read error: record skipped and continue.

## APIs

- No network APIs. CLI is the public interface.

## Client SDK Design

- Not exposed publicly. Internal scanning service may be extracted later.

## Configuration

- RuleSet schema and global defaults are defined in `specs/configuration.md`.
- Rule schema, message templates, and severity mapping are defined in `specs/regex-rules.md`.

## Permissions

- No auth or roles.

## Security Considerations

- RE2 regex prevents catastrophic backtracking.
- Skip large or binary files to avoid resource abuse.
- Treat match text as sensitive in logs.

## Dependencies

- `gopkg.in/yaml.v3`
- `github.com/bmatcuk/doublestar/v4`
- `github.com/owenrumney/go-sarif/v2/sarif`
- `golang.org/x/sync/errgroup`

## Open Questions / Risks

- Should we standardize output ordering at the engine or writer layer?
- Do we need optional redaction of match text for sensitive scans?

## Verifications

- Code builds and `make test` passes.
- CLI scan produces console, JSON, and SARIF outputs.

## Appendices

- See `specs/configuration.md` and `specs/regex-rules.md` for configuration details and `specs/data-model.md` for output schemas.
