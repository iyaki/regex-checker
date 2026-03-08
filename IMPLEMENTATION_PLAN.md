# Implementation Plan (ansi-colors)

**Status:** ANSI color scope is partially implemented; analyze color resolution is complete and console ANSI rendering remains (3/6 phases complete)
**Last Updated:** 2026-03-08
**Primary Specs:** `specs/formatter-console.md`, `specs/configuration.md`, `specs/cli-analyze.md` (related: `specs/formatter.md`, `specs/testing-and-validations.md`)

## Quick Reference

| System / Subsystem                                       | Specs                                                 | Modules / Packages                                                                                             | Artifacts                                                     | Status                                                                |
| -------------------------------------------------------- | ----------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- | --------------------------------------------------------------------- |
| Console formatter baseline (ordering, grouping, summary) | `specs/formatter-console.md`, `specs/formatter.md`    | `internal/output/console.go`, `internal/output/formatter.go`, `internal/output/registry.go`                    | `testdata/golden/console.txt`                                 | ✅ Implemented (plain output only)                                    |
| RuleSet color config schema                              | `specs/configuration.md`                              | `internal/config/model.go`, `internal/config/loader.go`, `internal/config/rules.go`, `internal/rules/model.go` | `internal/cli/init.go`, `testdata/rules/*.yaml`               | ✅ Implemented                                                        |
| Analyze color resolution and env precedence              | `specs/cli-analyze.md`, `specs/formatter-console.md`  | `internal/cli/analyze.go`, `internal/cli/help.go`                                                              | N/A                                                           | ✅ Implemented                                                        |
| ANSI severity rendering in console output                | `specs/formatter-console.md`                          | `internal/output/console.go`                                                                                   | `testdata/golden/console.txt` (or dedicated color fixtures)   | Missing                                                               |
| Non-console formatter behavior (must stay ANSI-free)     | `specs/formatter-json.md`, `specs/formatter-sarif.md` | `internal/output/json.go`, `internal/output/sarif.go`                                                          | `testdata/golden/output.json`, `testdata/golden/output.sarif` | ✅ Implemented                                                        |
| Verification and regression coverage                     | `specs/testing-and-validations.md`                    | `internal/output/*_test.go`, `internal/cli/*_test.go`, `internal/config/*_test.go`                             | `Makefile`, `testdata/golden/*`                               | Partial (CLI precedence coverage added; ANSI rendering tests missing) |

## Phase 1: Scope verification and delta lock

**Goal:** Lock exact ansi-colors requirements and confirmed gaps before code changes.
**Status:** Complete
**Paths:** `specs/formatter-console.md`, `specs/configuration.md`, `specs/cli-analyze.md`, `internal/output/console.go`, `internal/cli/analyze.go`, `internal/config/*.go`, `internal/rules/model.go`
**Reference pattern:** `internal/output/json.go` (deterministic formatting pipeline)

### 1.1 Spec and history verification

- [x] Verified ansi-colors requirements exist in formatter/config/analyze specs.
- [x] Verified recent scope change commit (`8fefbff`) and affected spec files.
- [ ] Clarify ambiguous empty env-var bullet in `specs/formatter-console.md` before implementation.

### 1.2 Code gap verification

- [x] Verified no production usage of `NO_COLOR` in `internal/**` runtime code.
- [x] Verified no `consoleColorsEnabled` fields in config/rules models.
- [x] Verified console formatter emits plain severity labels without ANSI SGR codes.
- [x] Verified JSON/SARIF formatters are unaffected by ANSI-color scope.

**Definition of Done**

- Verification commands and file reads are logged in the Verification Log.
- Gap list is actionable and scoped to config + CLI + output + tests.

**Risks/Dependencies**

- Spec ambiguity at `specs/formatter-console.md` around a blank env-var name may cause implementation churn if unresolved.

## Phase 2: RuleSet schema and model propagation

**Goal:** Implement `consoleColorsEnabled` in configuration and propagate it into runtime rule models.
**Status:** Complete
**Paths:** `internal/config/model.go`, `internal/config/loader.go`, `internal/config/rules.go`, `internal/rules/model.go`, `internal/config/loader_test.go`, `internal/config/rules_interpolate_coverage_test.go`
**Reference pattern:** `internal/config/model.go` + `internal/config/rules.go` (existing global fields such as `concurrency`, `failOn`)

### 2.1 Schema fields and defaults

- [x] Add `consoleColorsEnabled` to `config.RuleSet`.
- [x] Add `ConsoleColorsEnabled` to `rules.RuleSet`.
- [x] Propagate value through `RuleSet.ToRules()`.
- [x] Ensure default behavior is `true` when unset (resolved at runtime boundary).

### 2.2 Validation and tests

- [x] Add/adjust loader tests for boolean acceptance/rejection semantics.
- [x] Verify YAML type errors are surfaced as config parse/validation errors.
- [x] Existing loader validation/test scaffolding is present and can be extended.

**Definition of Done**

- `go test ./internal/config` passes.
- `consoleColorsEnabled` is available on the effective rules model used by analyze.
- Files touched are limited to config/rules model + tests.

**Risks/Dependencies**

- Backward compatibility for existing configs must be preserved (field remains optional).

## Phase 3: Analyze command color resolution

**Goal:** Resolve effective console color settings with config + environment precedence and wire to output.
**Status:** Complete
**Paths:** `internal/cli/analyze.go`, `internal/cli/scan_request_test.go`, `internal/cli/analyze_test.go`, `internal/cli/analyze_output_test.go`
**Reference pattern:** `internal/cli/analyze.go` precedence helpers (`resolveFailOn`, ignore-files resolution)

### 3.1 Effective setting resolution

- [x] Introduce runtime color setting in analyze execution path.
- [x] Apply precedence: default `true` -> RuleSet `consoleColorsEnabled` -> `NO_COLOR` non-empty forces `false`.
- [x] Record setting source if needed for parity with spec data model (`default|config|env`).

### 3.2 Output-path integration

- [x] Pass resolved color setting only to console formatter path.
- [x] Keep JSON/SARIF output paths and payloads unchanged.
- [x] Existing output routing by formatter name is in place and tested.

**Definition of Done**

- `go test ./internal/cli` passes with new precedence tests.
- Manual checks validate `NO_COLOR=1` disables ANSI in `--format console`.
- Files touched are limited to analyze runtime + tests.

**Risks/Dependencies**

- Env-dependent tests must isolate process env to avoid cross-test leakage.

## Phase 4: Console ANSI rendering

**Goal:** Add deterministic ANSI severity highlighting to console output when enabled.
**Status:** In progress
**Paths:** `internal/output/console.go`, `internal/output/console_test.go`, `internal/output/golden_test.go`
**Reference pattern:** `internal/output/json.go` (stable sorting and conversion pipeline)

### 4.1 Formatter API adjustments

- [x] Introduce `ConsoleColorSettings` in output layer (enabled flag, optional source metadata).
- [x] Wire settings into `ConsoleFormatter` without breaking registry contract.

### 4.2 Rendering semantics

- [ ] Apply fixed mapping: `ERROR=31`, `WARN=33`, `NOTICE=36`, `INFO=34`.
- [ ] Wrap only severity label segment and always reset with `\x1b[0m`.
- [ ] Emit byte-identical plain output when colors are disabled.
- [x] Current deterministic ordering/grouping behavior is already implemented and must be preserved.

**Definition of Done**

- `go test ./internal/output` passes.
- Console output contains ANSI SGR only when enabled.
- No ANSI sequences appear in disabled mode snapshots/assertions.

**Risks/Dependencies**

- Golden snapshot strategy must avoid brittle absolute-path assertions while validating color codes.

## Phase 5: Tests, fixtures, and docs alignment

**Goal:** Add explicit ansi-colors coverage and keep docs/examples aligned.
**Status:** Not started
**Paths:** `internal/output/*_test.go`, `internal/cli/*_test.go`, `internal/config/*_test.go`, `testdata/golden/*`, `testdata/rules/*`, `README.md`
**Reference pattern:** `internal/output/golden_test.go`

### 5.1 Automated coverage

- [ ] Add console tests for enabled ANSI emission and reset behavior.
- [ ] Add console tests for config-disabled mode (no ANSI).
- [ ] Add CLI tests for `NO_COLOR` precedence over config.
- [ ] Add config tests for `consoleColorsEnabled` parse/validation behavior.
- [x] Baseline tests for output/cli/config suites already exist.

### 5.2 Fixture and documentation updates

- [ ] Decide whether to use additional color golden files or targeted string assertions.
- [ ] Update sample config/docs to mention `consoleColorsEnabled` and `NO_COLOR` behavior.
- [ ] Keep quick examples consistent with actual CLI behavior.

**Definition of Done**

- `go test ./internal/output ./internal/cli ./internal/config` passes.
- README/examples and test fixtures reflect implemented behavior.
- Files touched are confined to tests/fixtures/docs relevant to ansi-colors.

**Risks/Dependencies**

- Documentation can drift if runtime precedence changes during implementation.

## Phase 6: Final verification and quality gates

**Goal:** Verify end-to-end behavior and clear quality gates for merge readiness.
**Status:** Not started
**Paths:** repository-wide (`internal/**`, `testdata/**`, `README.md`, `Makefile`)
**Reference pattern:** `specs/testing-and-validations.md`

### 6.1 Verification commands

- [ ] `go test ./...`
- [ ] `make test`
- [ ] `make lint`
- [ ] `make quality`
- [ ] Manual CLI checks for color enabled/disabled behavior.

### 6.2 Regression checks

- [ ] Confirm JSON and SARIF outputs remain ANSI-free.
- [ ] Confirm console ordering/summary remains deterministic.
- [x] Baseline package tests currently pass before ansi-colors changes.

**Definition of Done**

- All commands above pass.
- Verification Log includes final command outputs/results.
- No unrelated subsystem regressions are introduced.

**Risks/Dependencies**

- Mutation testing (`make mutation`) is expensive; run only at final stage per project guidance.

## Verification Log

- 2026-03-08: `Read specs/README.md` - confirmed primary and related specs for ansi-colors scope; files touched: `specs/README.md`.
- 2026-03-08: `Read specs/formatter-console.md` - captured ANSI mapping, reset, and precedence requirements; noted ambiguous blank env-var bullet; files touched: `specs/formatter-console.md`.
- 2026-03-08: `Read specs/configuration.md` - confirmed `consoleColorsEnabled` schema/default requirements; files touched: `specs/configuration.md`.
- 2026-03-08: `Read specs/cli-analyze.md` - confirmed `NO_COLOR` precedence and console-only scope; files touched: `specs/cli-analyze.md`.
- 2026-03-08: `git log --oneline -- specs` - identified recent specs history and ansi-colors commit (`8fefbff`).
- 2026-03-08: `git log --oneline -- specs/formatter-console.md` - verified scope-specific spec history.
- 2026-03-08: `git log --oneline -- specs/configuration.md` - verified scope-specific spec history.
- 2026-03-08: `git log --oneline -- specs/cli-analyze.md` - verified scope-specific spec history.
- 2026-03-08: `git show --name-only --oneline 8fefbff` - confirmed impacted specs: `specs/cli-analyze.md`, `specs/configuration.md`, `specs/formatter-console.md`.
- 2026-03-08: `grep "NO_COLOR|consoleColorsEnabled|ConsoleColorsEnabled|\\x1b\\[" (internal, *.go)` - no production implementation found for ansi-colors settings.
- 2026-03-08: `Read internal/output/console.go` - verified deterministic plain-text severity rendering with no ANSI sequences.
- 2026-03-08: `Read internal/cli/analyze.go` - verified no runtime resolution for `consoleColorsEnabled` or `NO_COLOR`.
- 2026-03-08: `Read internal/config/model.go` and `internal/rules/model.go` - verified missing color fields in RuleSet models.
- 2026-03-08: `go test ./internal/output ./internal/cli ./internal/config` - pass (cached).
- 2026-03-08: `Plan-only gap analysis` - bug fixes discovered: none; files touched: `IMPLEMENTATION_PLAN.md`.
- 2026-03-08: go test ./internal/config - pass; added coverage for consoleColorsEnabled parsing and RuleSet model propagation.
- 2026-03-08: go test ./internal/cli ./internal/output ./internal/rules - pass; no regressions from RuleSet model changes.
- 2026-03-08: make lint - pass.
- 2026-03-08: make arch - pass.
- 2026-03-08: make test-coverage - pass; coverage gate satisfied (>90%).
- 2026-03-08: git commit -m "Add consoleColorsEnabled to ruleset models" - success (commit `3d205ca`).
- 2026-03-08: Read specs/README.md, specs/cli-analyze.md, specs/formatter-console.md, specs/configuration.md - confirmed Phase 3 precedence and console-only ANSI scope.
- 2026-03-08: Read internal/cli/analyze.go and internal/output/console.go - confirmed analyze output routing and console formatter wiring points.
- 2026-03-08: go test ./internal/cli ./internal/output ./internal/config - pass.
- 2026-03-08: go test ./... - pass.
- 2026-03-08: NO_COLOR=1 go run ./cmd/reglint analyze --config testdata/rules/example.yaml --format console testdata/fixtures - pass (verified no ANSI sequences in stdout).
- 2026-03-08: make test - pass.
- 2026-03-08: make lint - pass.
- 2026-03-08: git commit -m "Resolve analyze console color precedence" - success (commit `171c57d`).

## Summary

| Phase                                         | Status      |
| --------------------------------------------- | ----------- |
| Phase 1: Scope verification and delta lock    | Complete    |
| Phase 2: RuleSet schema and model propagation | Complete    |
| Phase 3: Analyze command color resolution     | Complete    |
| Phase 4: Console ANSI rendering               | In progress |
| Phase 5: Tests, fixtures, and docs alignment  | Not started |
| Phase 6: Final verification and quality gates | Not started |

**Remaining effort:** Implement Phases 4-6; primary gaps are ANSI severity rendering, dedicated output color tests, and docs/fixtures alignment.

## Known Existing Work

- Console formatter already provides deterministic ordering/grouping and summary output in `internal/output/console.go`.
- Formatter registry + routing is established in `internal/output/registry.go` and `internal/cli/analyze.go`.
- Analyze runtime resolves console color precedence (`default -> config -> NO_COLOR`) and passes effective settings into the console formatter path.
- JSON and SARIF formatters already avoid ANSI concerns (`internal/output/json.go`, `internal/output/sarif.go`).
- Baseline output/CLI/config tests and golden tests already exist and can be extended (`internal/output/golden_test.go`, `testdata/golden/*`).

## Manual Deployment Tasks

None.
