# Implementation Plan (cli-analyze)

**Status:** Core implementation missing; tooling baseline present (1/5 phases complete)
**Last Updated:** 2026-03-04
**Primary Specs:** `specs/cli-analyze.md`, `specs/cli.md`, `specs/core-architecture.md`, `specs/data-model.md`, `specs/configuration.md`, `specs/regex-rules.md`, `specs/formatter-console.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md`, `specs/testing-and-validations.md`

## Quick Reference

| System / Subsystem     | Specs                                                                               | Modules / Packages                                     | Web Packages | Migrations / Artifacts                                                                  |
| ---------------------- | ----------------------------------------------------------------------------------- | ------------------------------------------------------ | ------------ | --------------------------------------------------------------------------------------- |
| CLI analyze command    | `specs/cli-analyze.md`, `specs/cli.md`                                              | `cmd/regex-checker/main.go`, `internal/cli/analyze.go` | N/A          | N/A                                                                                     |
| Config + rules loading | `specs/configuration.md`, `specs/regex-rules.md`                                    | `internal/config/`, `internal/rules/`                  | N/A          | N/A                                                                                     |
| Scan engine + service  | `specs/core-architecture.md`, `specs/data-model.md`                                 | `internal/scan/`, `internal/io/`                       | N/A          | N/A                                                                                     |
| Output formatters      | `specs/formatter-console.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md` | `internal/output/`                                     | N/A          | N/A                                                                                     |
| Tooling baseline       | `specs/testing-and-validations.md`                                                  | N/A                                                    | N/A          | ✅ `go.mod`, `scripts/quality.sh`, `lefthook.yml`, `.golangci.yml`, `.go-arch-lint.yml` |
| Testing + validation   | `specs/testing-and-validations.md`                                                  | `internal/**`, `testdata/`                             | N/A          | ✅ `.github/workflows/quality.yml`, `.github/workflows/quality-local.yml`               |

## Phase 1: CLI analyze entrypoint + routing

**Goal:** Implement CLI entrypoint, command routing, and alias handling.
**Status:** Complete
**Paths:** `cmd/regex-checker/`, `internal/cli/`
**Reference patterns:** `specs/cli.md`, `specs/cli-analyze.md`

### 1.1 CLI command structure

- [x] Create CLI entrypoint at `cmd/regex-checker/main.go` with subcommand parsing.
- [x] Implement `analyze` handler and `analyse` alias routing.
- [x] Print help and exit code `1` when no command is provided.
- [x] Print a single error message and exit code `1` for unknown commands.

**Definition of Done**

- `regex-checker analyze` and `regex-checker analyse` route to the same handler.
- Help/unknown command behaviors align with `specs/cli.md`.

**Risks/Dependencies**

- Requires structure alignment with core packages from `specs/core-architecture.md`.

## Phase 2: Analyze flag parsing + validation

**Goal:** Parse analyze flags, validate inputs, and map to `CLIConfig`.
**Status:** Complete
**Paths:** `internal/cli/analyze.go`
**Reference patterns:** `specs/cli-analyze.md`

### 2.1 Flag parsing + defaults

- [x] Parse flags: `--config`, `--format`, `--out-json`, `--out-sarif`, `--include`, `--exclude`, `--concurrency`, `--max-file-size`, `--fail-on`.
- [x] Default roots to `.` when no positional paths are provided.
- [x] Default `--config` to `regex-rules.yaml` in CWD.
- [x] Default formats to `console`.
- [x] Default `--concurrency` to `GOMAXPROCS` and `--max-file-size` to `5242880`.

### 2.2 Validation rules

- [x] Validate `--config` file existence and readability.
- [x] Validate `--format` values (`console|json|sarif`) and de-duplicate.
- [x] Validate `--concurrency` and `--max-file-size` are positive integers.
- [x] Validate `--fail-on` values (`error|warning|notice|info`).
- [x] Enforce output path requirements for multi-format output.
- [x] Enforce stdout behavior when only JSON or SARIF requested without output path.

**Definition of Done**

- `CLIConfig` fully maps to `ScanRequest` inputs and output requirements.
- Validation errors return exit code `1` with a single error message.

**Risks/Dependencies**

- Requires config loader and rules compiler for `--config` validation.

## Phase 3: Config loading + scan request construction

**Goal:** Load rules config, apply precedence, build `ScanRequest`.
**Status:** In progress
**Paths:** `internal/config/`, `internal/rules/`, `internal/scan/`
**Reference patterns:** `specs/configuration.md`, `specs/regex-rules.md`, `specs/data-model.md`

### 3.1 Config loader + rule compilation

- [x] Implement YAML parsing for RuleSet and validation of required fields.
- [x] Compile regex rules with RE2 and normalize severity defaults.
- [x] Implement message interpolation rules (`$0`, `$1`, `$$`).
- [x] Enforce RuleSet defaults for `include`, `exclude`, `failOn`, `concurrency`.

### 3.2 CLI precedence integration

- [x] Apply CLI overrides for `include`, `exclude`, and `failOn`.
- [x] Resolve effective include/exclude lists per rule.
- [x] Build `ScanRequest` with roots, rules, include/exclude, max size, concurrency.

**Definition of Done**

- `ScanRequest` conforms to `specs/data-model.md` and uses precedence rules.
- Invalid config or regex compilation exits with code `1`.

**Risks/Dependencies**

- Must align with formatter expectations for rule ordering and severity mapping.

## Phase 4: Scan engine + output writers

**Goal:** Implement scanning behavior and output rendering for console/JSON/SARIF.
**Status:** In progress
**Paths:** `internal/scan/`, `internal/output/`, `internal/io/`
**Reference patterns:** `specs/core-architecture.md`, `specs/formatter-console.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md`

### 4.1 Scan engine

- [x] Walk filesystem roots and apply include/exclude glob filtering.
- [x] Skip files exceeding `maxFileSizeBytes` or binary detection (per core spec).
- [x] Capture matches with 1-based line/column rune indices.
- [x] Add scan result data models (`Match`, `ScanStats`, `ScanResult`) in `internal/scan`.
- [x] Aggregate `ScanResult` with deterministic ordering.

### 4.2 Console output

- [x] Implement console formatter with grouping by file and summary line.
- [x] Ensure `No matches found.` output when no matches.

### 4.3 JSON output

- [x] Implement JSON formatter with `schemaVersion = 1` and stable ordering.
- [x] Write to stdout or `--out-json` per rules.

### 4.4 SARIF output

- [x] Implement SARIF formatter using rule ids `RC0001`+.
- [x] Map severities and set `columnKind` to `unicodeCodePoints`.
- [x] Compute `endColumn` as exclusive with rune length of `matchText`.

**Definition of Done**

- Outputs match formatter specs and are deterministic across runs.
- Exit code is `2` when matches meet or exceed `--fail-on` threshold.

**Risks/Dependencies**

- SARIF requires `github.com/owenrumney/go-sarif/v2/sarif`.

## Phase 5: Tests + validation coverage

**Goal:** Implement unit/integration/golden tests and quality tooling checks.
**Status:** Partially started (tooling present; tests missing)
**Paths:** `internal/**`, `testdata/`, `scripts/quality.sh`, `.github/workflows/`
**Reference patterns:** `specs/testing-and-validations.md`

### 5.1 Unit tests

- [ ] Config loader defaulting and validation.
- [x] Rule compiler regex compilation and message interpolation.
- [x] Path filtering include/exclude behavior.
- [ ] Scan engine line/column mapping and match aggregation.
- [x] CLI routing for `analyze` command.

### 5.2 Integration + golden tests

- [x] CLI analyze happy path with fixture rules and sample files.
- [x] Exit code behavior for invalid config and `failOn` threshold.
- [x] JSON/SARIF output validation and deterministic ordering.
- [ ] Golden snapshots for console/JSON/SARIF outputs.

### 5.3 Quality tooling baseline

- [x] Quality runner present (`scripts/quality.sh`).
- [x] Lint + architecture config present (`.golangci.yml`, `.go-arch-lint.yml`).
- [x] Git hooks config present (`lefthook.yml`).
- [x] CI workflows present (`.github/workflows/quality.yml`, `.github/workflows/quality-local.yml`).

**Definition of Done**

- `go test ./...` passes with coverage/mutation thresholds per spec.
- `scripts/quality.sh all` succeeds in local environment.

**Risks/Dependencies**

- Requires test fixtures under `testdata/` and stable output ordering.

## Verification Log

- 2026-03-04: `go test ./internal/scan -run TestRunCapturesRuneLineAndColumn` - pass.
- 2026-03-04: `go test ./internal/scan` - pass.
- 2026-03-04: `go test ./...` - pass.
- 2026-03-04: `bash scripts/quality.sh coverage` - pass.
- 2026-03-04: `go test ./internal/output -run TestWriteJSONNoMatches` - pass.
- 2026-03-04: `go test ./internal/output` - pass.
- 2026-03-04: `go test ./cmd/regex-checker` - pass.
- 2026-03-04: `go test ./internal/cli` - pass.
- 2026-03-04: `go test ./internal/output` - pass.
- 2026-03-04: `go test ./internal/scan -run TestCollectFilesFiltersByIncludeExclude` - pass.
- 2026-03-04: `go test ./internal/scan` - pass.
- 2026-03-04: `go test ./...` - pass.
- 2026-03-04: `go test ./internal/cli -run TestBuildScanRequest` - pass.
- 2026-03-04: Read `specs/cli-analyze.md` - documented flags, validations, exit codes, and output rules.
- 2026-03-04: Read `specs/cli.md` - confirmed CLI command structure and alias requirements.
- 2026-03-04: Read `specs/core-architecture.md` - captured module layout and scan/output flow.
- 2026-03-04: Read `specs/data-model.md` - confirmed `ScanRequest`, `Match`, `ScanResult` fields.
- 2026-03-04: Read formatter specs (`specs/formatter-console.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md`) - captured ordering/output requirements.
- 2026-03-04: `git log -n 10 -- specs` - reviewed recent spec changes for analyze scope.
- 2026-03-04: `glob **/*.go` - no Go source files found.
- 2026-03-04: `glob internal/**`, `glob cmd/**` - no implementation directories found.
- 2026-03-04: `go test ./internal/rules` - pass.
- 2026-03-04: Read quality/tooling configs (`scripts/quality.sh`, `.golangci.yml`, `.go-arch-lint.yml`, `lefthook.yml`, `.github/workflows/quality*.yml`) - tooling baseline present.
- 2026-03-04: `go test ./internal/cli` - pass.
- 2026-03-04: `go test ./internal/cli` - pass (added analyze routing coverage).
- 2026-03-04: `go test ./...` - pass.
- 2026-03-04: `bash scripts/quality.sh all` - pass.
- 2026-03-04: `go test ./internal/cli` - pass.
- 2026-03-04: `bash scripts/quality.sh all` - pass.
- 2026-03-04: `go test ./internal/config` - pass.
- 2026-03-04: `go test ./...` - pass.
- 2026-03-04: `go test ./internal/config -run TestRuleSetToRulesDefaultsRulePathsAndExclude` - pass.
- 2026-03-04: `go test ./internal/config -run TestRuleSetToRulesDefaultsConcurrency` - pass.
- 2026-03-04: `go test ./internal/scan -run TestScanModelsHoldFields` - pass.
- 2026-03-04: `go test ./internal/output` - pass.
- 2026-03-04: `bash scripts/quality.sh all` - pass.

## Summary

| Phase                                  | Status                                                         |
| -------------------------------------- | -------------------------------------------------------------- |
| Phase 1: CLI entrypoint + routing      | Complete                                                       |
| Phase 2: Flag parsing + validation     | Complete                                                       |
| Phase 3: Config loading + scan request | In progress                                                    |
| Phase 4: Scan engine + output writers  | In progress                                                    |
| Phase 5: Tests + validation coverage   | Partially started (tooling baseline + CLI analyze integration) |

**Remaining effort:** Complete remaining scan engine tests, golden outputs, and finalize Phase 3/4 status.

## Known Existing Work

- Tooling baseline present: `go.mod`, `scripts/quality.sh`, `.golangci.yml`, `.go-arch-lint.yml`, `lefthook.yml`, `.github/workflows/quality.yml`, `.github/workflows/quality-local.yml`.
- CLI entrypoint + routing scaffolding added in `cmd/regex-checker/` and `internal/cli/`.

## Manual Deployment Tasks

None.
