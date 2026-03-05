# Analyze Command

Status: Implemented

## Overview

### Purpose

- Run regex scans using a YAML rules config and selected output formats.
- Parse flags, load rules, build a `ScanRequest`, and render outputs.
- Provide deterministic exit codes for CI usage.

### Goals

- Stable, well-documented flags and exit codes.
- Clear validation errors and deterministic behavior.
- Support console, JSON, and SARIF output selection.
- Provide the `analyse` alias with identical behavior.

### Non-Goals

- Interactive UI or TUI.
- Editing or generating rules files (see `specs/cli-init.md`).
- Remote or daemonized scanning.

### Scope

- Command syntax, flags, defaults, and validation rules.
- Mapping flags to `ScanRequest` and output writers.

## Architecture

### Module/package layout (tree format)

```
cmd/
  regex-checker/
    main.go
internal/
  cli/
    analyze.go
  config/
  rules/
  scan/
  output/
```

### Component diagram (ASCII)

```
[Analyze Command] -> [Config Loader] -> [Scan Service] -> [Output Writers]
```

### Data flow summary

1. Parse flags and positional paths.
2. Load and validate the rules config.
3. Resolve include/exclude lists and defaults.
4. Build `ScanRequest`.
5. Run scan and render requested outputs.
6. Exit with a deterministic code.

## Data model

### Core Entities

CLIConfig

- Definition: Parsed CLI inputs and resolved defaults for the analyze command.
- Fields:
  - `configPath` (string, required)
  - `roots` (list of string, required)
  - `formats` (list of string, required): `console|json|sarif`
  - `outJSON` (string, optional)
  - `outSARIF` (string, optional)
  - `include` (list of string, optional)
  - `exclude` (list of string, optional)
  - `concurrency` (int, required)
  - `maxFileSizeBytes` (int64, required)
  - `failOnSeverity` (string, optional)

Reference struct (Go):

```go
type CLIConfig struct {
    ConfigPath       string
    Roots            []string
    Formats          []string
    OutJSON          string
    OutSARIF         string
    Include          []string
    Exclude          []string
    Concurrency      int
    MaxFileSizeBytes int64
    FailOnSeverity   string
}
```

### Relationships

- `CLIConfig` maps directly to `ScanRequest` in `specs/data-model.md`.
- Severity values align with `specs/regex-rules.md`.

## Workflows

### Parse flags and args (happy path)

1. Parse flags.
2. Collect positional arguments as `roots`.
3. If no roots provided, set `roots = ["."]`.
4. Resolve the config path (default `regex-rules.yaml`).
5. Validate flags and formats (see Validation).
6. Load and compile rules from the config file.
7. Apply CLI overrides and build `ScanRequest`.
8. Run scan and render outputs.

### Validation and errors

- `--config` defaults to `regex-rules.yaml` in the current directory. If the file is missing or unreadable, print an error and exit with code 1.
- `--format` must include only `console`, `json`, or `sarif`.
- `--concurrency` must be a positive integer.
- `--max-file-size` must be a positive integer.
- `--fail-on` must be one of `error|warning|notice|info` when set.
- If multiple formats are requested:
  - `json` requires `--out-json`.
  - `sarif` requires `--out-sarif`.
- Any validation failure prints a single error message and exits with code 1.

### Exit codes

- `0`: scan completed and no match at or above `--fail-on` (or `--fail-on` unset).
- `2`: scan completed and has matches at or above `--fail-on`.
- `1`: configuration or runtime error.

## Configuration

### Command syntax

```
regex-checker analyze [flags] [path ...]
regex-checker analyse [flags] [path ...]
```

### Flags

| Flag              | Type   | Required | Default            | Purpose                                       |
| ----------------- | ------ | -------- | ------------------ | --------------------------------------------- |
| `--config`        | string | no       | `regex-rules.yaml` | Path to YAML rules config file.               |
| `--format`        | string | no       | `console`          | Comma-separated list of `console,json,sarif`. |
| `--out-json`      | string | no       | none               | Output path for JSON results.                 |
| `--out-sarif`     | string | no       | none               | Output path for SARIF results.                |
| `--include`       | string | no       | none               | Repeatable include glob for all rules.        |
| `--exclude`       | string | no       | none               | Repeatable exclude glob for all rules.        |
| `--concurrency`   | int    | no       | `GOMAXPROCS`       | Worker count.                                 |
| `--max-file-size` | int    | no       | `5242880`          | Skip files larger than N bytes.               |
| `--fail-on`       | string | no       | none               | Fail if matches at or above severity.         |

### Precedence

- If `--include` is provided, it replaces the RuleSet `include` list from `specs/configuration.md`.
- If `--exclude` is provided, it replaces the RuleSet `exclude` list from `specs/configuration.md`.
- If `--include` is not provided, use RuleSet `include` or default to `**/*`.
- If `--exclude` is not provided, use RuleSet `exclude` or default to `**/.git/**`, `**/node_modules/**`, `**/vendor/**`.
- If `--fail-on` is provided, it overrides RuleSet `failOn`.
- If `--fail-on` is not provided, use RuleSet `failOn` when set; otherwise unset.

### Output behavior

- `console` always writes to stdout.
- If `json` is the only format and `--out-json` is unset, write JSON to stdout.
- If `sarif` is the only format and `--out-sarif` is unset, write SARIF to stdout.
- If multiple formats are requested, JSON and SARIF must have explicit output paths.

## Verifications

- `regex-checker analyze --config rules.yaml` scans current directory and exits with code 0/2.
- `regex-checker analyze --config rules.yaml --format json` writes JSON to stdout.
- `regex-checker analyze --config rules.yaml --format console,json --out-json /tmp/scan.json` writes console to stdout and JSON to file.
- Invalid `--fail-on` value exits with code 1 and prints an error.

## Appendices

### Examples

```
regex-checker analyze --config configs/example.rules.yaml
regex-checker analyse --config configs/example.rules.yaml --format json --out-json /tmp/scan.json
regex-checker analyze --config configs/example.rules.yaml --format sarif --out-sarif /tmp/scan.sarif
```
