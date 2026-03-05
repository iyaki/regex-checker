# Regex Rules

Status: Implemented

## Overview

### Purpose

- Define the regex rule schema used by regex-checker.
- Capture rule-level defaults and validation.
- Document message templates and regex constraints.

### Goals

- Keep rule definitions stable across versions.
- Ensure predictable matching behavior and severity semantics.

### Non-Goals

- Define CLI flags or output formats (see other specs).
- Support regex features beyond Go RE2.

## Rule Schema

Rule

- Definition: A single regex rule entry within a RuleSet.
- Fields:
  - `message` (string, required): Annotation message. Supports capture group interpolation (see Message templates).
  - `regex` (string, required): RE2 regex pattern.
  - `severity` (string, optional): One of `error|warning|notice|info`. Default: `warning`.
  - `paths` (list of string, optional): Inclusion globs for file selection. Default: inherits RuleSet `include` when set; otherwise `**/*`.
  - `exclude` (list of string, optional): Exclusion globs for this rule. Overrides RuleSet `exclude` for that rule.

## Defaults

- `severity`: `warning` if missing.
- `paths`: inherits RuleSet `include` when set; otherwise `**/*`.
- `exclude`: inherits RuleSet `exclude` when set; otherwise none.

## Severity mapping

| Severity  | JSON      | SARIF level |
| --------- | --------- | ----------- |
| `error`   | `error`   | `error`     |
| `warning` | `warning` | `warning`   |
| `notice`  | `notice`  | `note`      |
| `info`    | `info`    | `note`      |

## Path filtering

- Paths use doublestar-style `**` globs.
- Rule `paths` overrides RuleSet `include`.
- Rule `exclude` overrides RuleSet `exclude` for that rule.

## Message templates

- Messages may interpolate regex capture groups using `$1`, `$2`, ... based on the rule's `regex`.
- `$0` refers to the full match.
- Missing groups resolve to an empty string.

### Formatting rules

- Only numeric group references are supported (`$0`, `$1`, ...). Named groups are not substituted.
- Use `$$` to emit a literal `$` in the message.
- A `$` not followed by a digit or another `$` is treated as a literal `$`.
- Group indices follow the regex capture order (left to right).

## Regex syntax

- Regex patterns use Go RE2 syntax (no lookbehind, no PCRE-only features).
- Patterns are compiled and validated once during configuration loading.

## Validation

- Each rule requires `message` and `regex`.
- `severity` must be one of the allowed values when set.
- `regex` patterns must compile using RE2.

## YAML example

```yaml
rules:
  - message: "Avoid hardcoded token: $1"
    regex: "token\s*[:=]\s*([A-Za-z0-9_-]+)"
    severity: "error"
    paths:
      - "src/**"
    exclude:
      - "src/vendor/**"
```
