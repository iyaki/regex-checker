# Configuration

Status: Implemented

## Overview

### Purpose

- Define the YAML schema for regex-checker rules.
- Capture defaults and validation rules in one place.
- Document global scan controls that apply across rules.

### Goals

- Keep the rules config easy to read and stable across versions.
- Ensure deterministic defaults for scanning behavior.

### Non-Goals

- Define CLI flags or output formats (see other specs).

## Schema

RuleSet

- Definition: Top-level YAML configuration file.
- Fields:
  - `rules` (list of Rule, required): Rule schema, defaults, and validation are defined in `specs/regex-rules.md`.
  - `include` (list of string, optional): Global include globs. Overridden by per-rule `paths`.
  - `exclude` (list of string, optional): Global exclude globs. Overridden by per-rule `exclude`.
  - `failOn` (string, optional): One of `error|warning|notice|info`. Causes non-zero exit status at or above this severity.
  - `concurrency` (int, optional): Worker count for scanning.

## YAML example (with globals)

```yaml
include:
  - "src/**"
  - "lib/**"
exclude:
  - "**/generated/**"
failOn: "error"
concurrency: 8
rules:
  - message: "This is an error message"
    regex: "regex1"
    severity: "error"
    paths:
      - "src/**/*.js"
      - "lib/**/*.js"
  - message: "This is a warning message"
    regex: "regex2"
    severity: "warning"
    exclude:
      - "lib/vendor/**"
```

## Defaults

- RuleSet `include`: `**/*` if missing.
- RuleSet `exclude`: `**/.git/**`, `**/node_modules/**`, `**/vendor/**` if missing.
- `failOn`: unset (no failure threshold) if missing.
- `concurrency`: `GOMAXPROCS` if missing.

## Validation

- YAML must parse successfully.
- `rules` is required.
- `failOn` must be one of the allowed values when set.
- `concurrency` must be a positive integer when set.
- Rules are validated per `specs/regex-rules.md`.

## Notes

- Rule schema, defaults, and path override behavior are defined in `specs/regex-rules.md`.
