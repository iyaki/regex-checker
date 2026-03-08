# Testing and Validations

## Overview

### Purpose

- Define validation rules for configuration, rules, and CLI inputs.
- Establish the automated test strategy for correctness and regression safety.
- Keep outputs consistent across console, JSON, and SARIF formats.

### Goals

- Fail fast on invalid configuration or rule definitions with clear errors.
- Ensure validation happens before any scan starts.
- Provide deterministic outputs so tests and CI are stable.
- Cover core workflows and edge cases with unit and integration tests.

### Non-Goals

- Performance benchmarking or load testing.
- Formal verification or proof of correctness.
- Defining CI pipelines beyond required test commands.

### Scope

- CLI-only execution (no services or APIs).
- Validation and testing of local scans, outputs, and exit codes.
- Local quality tooling (linting, security, architecture guardrails) and git hooks.

## Quality Tooling

### Tooling set

- `golangci-lint` (lint aggregator)
- `govulncheck` (Go vulnerability scanner)
- `go-arch-lint` (architecture guardrails)
- `gofmt` (formatting)
- `gremlins` (mutation testing)
- `lefthook` (git hooks runner)

### Local usage

Full suite:

```bash
make quality
```

Install git hooks:

```bash
lefthook install
```

Targeted runs:

```bash
make lint
make test
make test-race
make coverage
make mutation
make security
make arch
```

### Configuration files

- `/.golangci.yml` - lint configuration
- `/.go-arch-lint.yml` - architecture rules
- `/.tool-versions` - pinned golangci-lint version
- `/lefthook.yml` - git hook configuration

### CI coverage

- `/.github/workflows/quality.yml` runs lint, and architecture checks.
- `/.github/workflows/quality.yml` runs mutation testing via `make mutation`.

## Validation Rules

### RuleSet (configuration)

- YAML must parse successfully.
- `rules` is required and must be a non-empty list.
- `include` and `exclude` must be lists of strings when set.
- `failOn` must be one of `error|warning|notice|info` when set.
- `concurrency` must be a positive integer when set.

### Rule (per entry)

- `message` is required and must be non-empty.
- `regex` is required and must compile with RE2.
- `severity` must be one of `error|warning|notice|info` when set.
- `paths` and `exclude` must be lists of strings when set.

### CLI (analyze)

- Unknown commands exit with code `1` and show a single error message.
- `--config` (`-c`) must point to a readable file.
- `--format` (`-f`) must be one of `console|json|sarif` (per formatter specs).
- Output paths must be writable when `--out-*` flags are set.

### Runtime scan

- Invalid YAML or regex patterns exit with code `1`.
- File read errors are recorded and scanning continues.
- Matches at or above `failOn` return exit code `2`.
- Output writer errors exit with code `1`.

## Test Strategy

### Unit tests

- Config loader: YAML parsing, defaults, and validation errors.
- Rule compiler: regex compilation, severity normalization, message templates.
- Path filtering: include/exclude behavior with doublestar globs.
- Scan engine: line/column mapping and match aggregation.

### Integration tests

- CLI analyze happy path with fixture rules and sample files.
- Exit code behavior for invalid config, invalid regex, and `failOn` threshold.
- Output writers produce valid JSON and SARIF.

### Golden tests

- Snapshot console/JSON/SARIF outputs for fixture directories.
- Keep outputs deterministic by sorting matches by file path, then line, then rule index.

### Regression tests

- Binary and large file skipping behavior.
- Regex capture group interpolation (`$0`, `$1`, `$$`).
- Mixed include/exclude rules and per-rule overrides.

### Mutation testing

- Enforced via `gremlins` on local quality runs, pre-commit hooks, and CI.
- Minimum mutation score default: `0.8` (override with `MUTATION_SCORE_MIN`).
- Optional minimum mutant coverage default: `0` (override with `MUTATION_COVERAGE_MIN`).
- Optional diff target can be passed with `make mutation ARGS='--diff <target>'`.

### Coverage

- Minimum line coverage default: `90%` (override with `COVERAGE_MIN`).

## Test Data

- Use a top-level `testdata/` directory with:
  - `rules/` for sample RuleSet files.
  - `fixtures/` for code samples and edge cases.
  - `golden/` for expected outputs.

## Test Execution

- `make test` runs unit and integration tests.
- Optional: `UPDATE_GOLDEN=1 make test` to refresh golden files.
- Unit tests must enforce line coverage > 90% (exclude integration and testdata packages).
- Mutation testing must meet the minimum mutation score.

## Verifications

- `make test` passes.
- `reglint analyze --config testdata/rules/example.yaml ./testdata/fixtures` exits with `0` when below `failOn`.
- `reglint analyze --config testdata/rules/fail.yaml ./testdata/fixtures` exits with `2` when above `failOn`.
