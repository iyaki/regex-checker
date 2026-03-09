# Implementation Plan (baseline)

**Status:** Baseline feature set is partially implemented in runtime code (Discovery complete, 6/8 phases complete; Phase 7 in progress)
**Last Updated:** 2026-03-09
**Primary Specs:** `specs/cli-analyze-baseline.md`, `specs/cli-analyze.md`, `specs/configuration.md` (related: `specs/testing-and-validations.md`, `specs/cli-help.md`, `specs/cli.md`, `specs/core-architecture.md`, `specs/formatter.md`)

## Quick Reference

| System / Subsystem                                                         | Specs                                                                                              | Modules / Packages                                                                                             | Artifacts                         | Status                                                      |
| -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | --------------------------------- | ----------------------------------------------------------- |
| Baseline domain model, loader, comparator, writer                          | `specs/cli-analyze-baseline.md`, `specs/core-architecture.md`                                      | `internal/baseline/*`                                                                                          | `testdata/baseline/*.json`        | ✅ Implemented (model + loader + compare + writer complete) |
| Analyze baseline flags and control flow (`--baseline`, `--write-baseline`) | `specs/cli-analyze.md`, `specs/cli.md`                                                             | `internal/cli/analyze.go`, `cmd/reglint/main.go`                                                               | CLI integration tests             | ✅ Implemented (flags + precedence + compare/write flow)    |
| RuleSet baseline config field and propagation                              | `specs/configuration.md`, `specs/cli-analyze.md`                                                   | `internal/config/model.go`, `internal/config/loader.go`, `internal/config/rules.go`, `internal/rules/model.go` | `testdata/rules/*.yaml`           | ✅ Implemented (field + validation + propagation complete)  |
| Help output coverage for baseline flags                                    | `specs/cli-help.md`, `specs/cli-analyze.md`                                                        | `internal/cli/help.go`, `internal/cli/cli_test.go`                                                             | Help snapshot assertions in tests | ✅ Implemented (baseline flags + alias help coverage)       |
| Scan + formatter deterministic pipeline dependency                         | `specs/data-model.md`, `specs/formatter.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md` | `internal/scan/engine.go`, `internal/output/*.go`                                                              | `testdata/golden/*`               | ✅ Implemented (reusable baseline dependency)               |
| Existing analyze routing, fail-on behavior, and output selection           | `specs/cli-analyze.md`                                                                             | `internal/cli/analyze.go`, `internal/cli/analyze_output_test.go`, `cmd/reglint/main_test.go`                   | Existing CLI tests                | ✅ Implemented (baseline extension point)                   |
| Ignore-file precedence pattern (config -> CLI override)                    | `specs/ignore-files.md`                                                                            | `internal/cli/analyze.go`, `internal/scan/ignore_rules.go`, `internal/ignore/*`                                | Ignore behavior tests             | ✅ Implemented (reference pattern for precedence design)    |

## Phase 1: Scope verification and plan reset

**Goal:** Lock the baseline scope and replace stale planning content with verified gaps.
**Status:** Complete
**Paths:** `specs/cli-analyze-baseline.md`, `specs/cli-analyze.md`, `specs/configuration.md`, `specs/testing-and-validations.md`, `IMPLEMENTATION_PLAN.md`
**Reference pattern:** `internal/cli/analyze.go` (existing precedence and output pipeline shape)

### 1.1 Spec and git history verification

- [x] Verified baseline specs are present and linked from `specs/README.md`.
- [x] Verified recent baseline scope update via commit `fd4ae9f`.
- [x] Verified baseline-related spec files changed together (`cli`, `analyze`, `help`, `configuration`, `testing`, `core-architecture`, `formatter`).

### 1.2 Gap confirmation and stale-plan replacement

- [x] Verified previous `IMPLEMENTATION_PLAN.md` was for `ansi-colors` and out of scope for baseline.
- [x] Verified no production baseline implementation exists in `internal/**` or `cmd/**`.
- [x] Regenerated plan to baseline scope and current code reality.

**Definition of Done**

- Verification artifacts captured in the Verification Log.
- Plan reflects actual repository state instead of prior ANSI-color completion state.

**Risks/Dependencies**

- Several specs claim baseline-ready behavior while runtime code does not; this increases risk of false confidence unless tracked explicitly.

## Phase 2: RuleSet schema and baseline path propagation

**Goal:** Add baseline path support in config models and enforce schema validation.
**Status:** Complete
**Paths:** `internal/config/model.go`, `internal/config/loader.go`, `internal/config/loader_test.go`, `internal/config/rules.go`, `internal/rules/model.go`
**Reference pattern:** `internal/config/model.go` and `internal/config/rules.go` existing global field propagation (`failOn`, `concurrency`, `consoleColorsEnabled`)

### 2.1 RuleSet schema updates

- [x] Add `baseline` to `config.RuleSet` and `rules.RuleSet`.
- [x] Keep field optional but reject empty/whitespace values when provided.
- [x] Propagate value through `RuleSet.ToRules()` with existing copy semantics.

### 2.2 Validation and conversion tests

- [x] Add loader tests for valid baseline path values.
- [x] Add loader tests for invalid baseline path shapes (empty/non-string).
- [x] Add conversion tests confirming baseline propagation and immutability.

**Definition of Done**

- `go test ./internal/config ./internal/rules` passes.
- Baseline field is available in effective rules model used by analyze runtime.
- Files touched are limited to config/rules schema + tests.

**Risks/Dependencies**

- Validation behavior must stay compatible with existing YAML parsing error style (single actionable error).

## Phase 3: Baseline package implementation (`internal/baseline`)

**Goal:** Implement deterministic baseline load/validate/compare/write services.
**Status:** Complete
**Paths:** `internal/baseline/model.go`, `internal/baseline/loader.go`, `internal/baseline/compare.go`, `internal/baseline/writer.go`, `internal/baseline/*_test.go`
**Reference pattern:** `internal/scan/engine.go` deterministic ordering (`sortMatches`) and `internal/output/json.go` stable canonical output style

### 3.1 Data model and validation

- [x] Define `BaselineEntry`, `BaselineDocument`, `BaselineComparison`, and generation result structs.
- [x] Enforce `schemaVersion == 1`, required `entries`, and positive `count`.
- [x] Enforce unique `(filePath, message)` keys and relative normalized `filePath` without traversal.

### 3.2 Comparison service

- [x] Implement suppression by `(filePath, message)` with count decrement semantics.
- [x] Return regression matches only, plus `suppressedCount` and `improvementsCount`.
- [x] Keep comparison deterministic for equivalent inputs.

### 3.3 Baseline writer service

- [x] Aggregate full matches into canonical `(filePath, message)` counts.
- [x] Write canonical JSON (`schemaVersion=1`, sorted by `filePath`, then `message`).
- [x] Overwrite existing target file in write mode.

**Definition of Done**

- `go test ./internal/baseline` passes.
- Loader and writer return deterministic output and single clear errors.
- New package is isolated and reusable from `internal/cli/analyze.go`.

**Risks/Dependencies**

- Message-based keys can be volatile when interpolation changes; tests must pin current behavior.

## Phase 4: Analyze CLI flags, precedence, and path resolution

**Goal:** Add baseline flags and effective path resolution behavior to analyze command.
**Status:** Complete
**Paths:** `internal/cli/analyze.go`, `internal/cli/analyze_test.go`, `internal/cli/scan_request_test.go`, `internal/cli/analyze_handle_test.go`
**Reference pattern:** existing precedence helpers in `internal/cli/analyze.go` (`resolveFailOn`, ignore settings resolution)

### 4.1 Flag parsing and validation

- [x] Add `--baseline` (string) and `--write-baseline` (bool) to `ParseAnalyzeArgs`.
- [x] Extend `Config` with baseline fields.
- [x] Enforce `--write-baseline` requires an effective baseline path.

### 4.2 Effective baseline path precedence

- [x] Apply precedence `--baseline` > RuleSet `baseline` > unset.
- [x] Resolve RuleSet baseline relative to config file directory.
- [x] Resolve CLI baseline relative to current working directory.

**Definition of Done**

- `go test ./internal/cli -run Baseline` (or equivalent targeted selection) passes.
- Errors remain single-line and exit through existing `exitCodeError` path.
- No behavior regression for non-baseline analyze runs.

**Risks/Dependencies**

- Path resolution differences between config-relative and cwd-relative modes can cause subtle CI regressions if under-tested.

## Phase 5: Analyze runtime integration (compare/write modes + exit semantics)

**Goal:** Integrate baseline logic into scan execution without changing formatter contracts.
**Status:** Complete
**Paths:** `internal/cli/analyze.go`, `cmd/reglint/main_test.go`, `internal/cli/analyze_output_test.go`, `internal/output/*.go` (verification only)
**Reference pattern:** current `runAnalyze -> renderOutputs -> fail-on` flow in `internal/cli/analyze.go`

### 5.1 Baseline compare mode

- [x] Load baseline when active and `--write-baseline` is false.
- [x] Filter full scan matches to regression-only result before output rendering.
- [x] Evaluate `--fail-on` against regression matches only.

### 5.2 Baseline write mode

- [x] Skip suppression when `--write-baseline` is set.
- [x] Generate baseline from full findings and overwrite target file.
- [x] Return exit code `0` on successful baseline write regardless of matches/`--fail-on`.

### 5.3 Formatter contract guardrails

- [x] Keep console/json/sarif schemas unchanged.
- [x] Ensure `stats.matches` reflects effective result in compare mode.
- [x] Ensure write mode still renders full findings.

**Definition of Done**

- `go test ./internal/cli ./cmd/reglint` passes.
- Manual CLI checks for equal/increase/decrease baseline counts match spec behavior.
- Formatter outputs remain stable except expected match-count filtering.

**Risks/Dependencies**

- Exit code overrides in write mode can be accidentally broken by existing fail-on path if integration is not carefully ordered.

## Phase 6: Help text, docs, and fixture alignment

**Goal:** Align CLI help and user-facing docs with new baseline capabilities.
**Status:** Complete
**Paths:** `internal/cli/help.go`, `internal/cli/cli_test.go`, `README.md`, `testdata/rules/*.yaml`, `testdata/baseline/*.json`
**Reference pattern:** existing deterministic help-output snapshots in `internal/cli/cli_test.go`

### 6.1 Analyze help output

- [x] Add `--baseline` and `--write-baseline` to analyze help topic output.
- [x] Update strict help snapshot tests.
- [x] Keep alias behavior (`analyse`) unchanged.

### 6.2 Docs and sample artifacts

- [x] Add baseline usage examples to README.
- [x] Add baseline JSON fixtures for valid and invalid cases.
- [x] Keep examples aligned with actual command behavior and exit codes.

**Definition of Done**

- `go test ./internal/cli` passes with updated help snapshots.
- README examples are executable against local fixtures.
- Baseline fixture files are deterministic and minimal.

**Risks/Dependencies**

- Help snapshot tests are strict; minor formatting drift can cause broad failures.

## Phase 7: End-to-end tests and regression coverage

**Goal:** Add baseline-focused unit/integration coverage across config, CLI, and command entrypoints.
**Status:** Complete
**Paths:** `internal/baseline/*_test.go`, `internal/config/loader_test.go`, `internal/cli/*_test.go`, `cmd/reglint/main_test.go`, `testdata/*`
**Reference pattern:** current command-level tests in `cmd/reglint/main_test.go` and fixture-driven tests in `internal/cli/analyze_handle_test.go`

### 7.1 Baseline behavior matrix

- [x] Equal-count suppression yields zero regressions.
- [x] Increased count yields only excess regressions.
- [x] Decreased count yields no regressions and non-failing behavior for that key.

### 7.2 Precedence and validation matrix

- [x] RuleSet baseline path works when CLI flag is unset.
- [x] CLI baseline overrides RuleSet baseline.
- [x] Invalid baseline JSON/schema/duplicates fail with exit code `1` and one error message.

### 7.3 Write-mode matrix

- [x] `--write-baseline` requires effective path.
- [x] Existing baseline content is ignored during write mode.
- [x] Successful write mode exits `0` even with failing matches.

**Definition of Done**

- `go test ./...` passes with new baseline cases.
- Added tests explicitly cover all verification bullets in `specs/cli-analyze-baseline.md`.
- No skipped or flaky baseline tests introduced.

**Risks/Dependencies**

- Command-level tests may require careful cwd and temp-file isolation to avoid cross-test interference.

## Phase 8: Final quality gates and release readiness

**Goal:** Validate baseline implementation against repository quality gates and manual behavior checks.
**Status:** Not started
**Paths:** repository-wide (`internal/**`, `cmd/**`, `testdata/**`, `README.md`, `Makefile`)
**Reference pattern:** `specs/testing-and-validations.md`

### 8.1 Automated gates

- [ ] `go test ./...`
- [ ] `make test`
- [ ] `make lint`
- [ ] `make quality`
- [ ] `make mutation` (final stage only, per project guidance)

### 8.2 Manual verification commands

- [ ] `reglint analyze --config <rules> --baseline <file> <path>` (compare mode).
- [ ] `reglint analyze --config <rules> --baseline <file> --write-baseline <path>` (write mode).
- [ ] `reglint analyze --help` includes baseline flags.
- [ ] JSON/SARIF outputs remain ANSI-free and schema-stable after baseline filtering.

**Definition of Done**

- All quality gates pass.
- Verification Log is updated with command outcomes.
- Remaining effort is reduced to `None`.

**Risks/Dependencies**

- Mutation testing runtime can be long; run only after implementation stabilizes.

## Verification Log

- 2026-03-09: `Read specs/README.md` - baseline spec index confirmed; tests run: none (planning mode); bug fixes discovered: none; files touched: `specs/README.md`.
- 2026-03-09: `Read specs/cli-analyze-baseline.md` - baseline requirements and data model captured; tests run: none; bug fixes discovered: none; files touched: `specs/cli-analyze-baseline.md`.
- 2026-03-09: `Read specs/cli-analyze.md` - baseline flags, precedence, and exit semantics captured; tests run: none; bug fixes discovered: none; files touched: `specs/cli-analyze.md`.
- 2026-03-09: `Read specs/configuration.md` - verified expected RuleSet `baseline` field; tests run: none; bug fixes discovered: none; files touched: `specs/configuration.md`.
- 2026-03-09: `Read specs/testing-and-validations.md` - baseline validation/test obligations captured; tests run: none; bug fixes discovered: none; files touched: `specs/testing-and-validations.md`.
- 2026-03-09: `git log --oneline --decorate -n 20 -- specs/cli-analyze-baseline.md` - baseline spec introduced in `fd4ae9f`; tests run: none; bug fixes discovered: none; files touched: `specs/cli-analyze-baseline.md`.
- 2026-03-09: `git log --oneline --decorate -n 20 -- specs/cli-analyze.md` and `git log --oneline --decorate -n 20 -- specs/configuration.md` - baseline-related spec updates confirmed in `fd4ae9f`; tests run: none; bug fixes discovered: none; files touched: `specs/cli-analyze.md`, `specs/configuration.md`.
- 2026-03-09: `git show --name-only --oneline fd4ae9f` - confirmed baseline scope touched CLI/analyze/help/config/testing/core architecture specs; tests run: none; bug fixes discovered: none; files touched: `specs/*.md` set listed by command.
- 2026-03-09: `glob internal/baseline/**/*.go` - no baseline package files found; tests run: none; bug fixes discovered: none; files touched: none.
- 2026-03-09: `grep "baseline|write-baseline|Baseline" internal/*.go cmd/*.go` - no runtime baseline references found; tests run: none; bug fixes discovered: none; files touched: none.
- 2026-03-09: `Read internal/cli/analyze.go` - verified no baseline flags or runtime compare/write flow; tests run: none; bug fixes discovered: none; files touched: `internal/cli/analyze.go`.
- 2026-03-09: `Read internal/config/model.go`, `internal/config/loader.go`, `internal/config/rules.go`, `internal/rules/model.go` - verified `baseline` field not implemented; tests run: none; bug fixes discovered: none; files touched: listed files.
- 2026-03-09: `Read internal/cli/help.go` and `internal/cli/cli_test.go` - verified analyze help output omits baseline flags despite spec expectations; tests run: none; bug fixes discovered: none; files touched: `internal/cli/help.go`, `internal/cli/cli_test.go`.
- 2026-03-09: `Read internal/output/console.go`, `internal/output/json.go`, `internal/output/sarif.go`, `internal/scan/engine.go` - verified deterministic scan/output path exists and is reusable for baseline integration; tests run: none; bug fixes discovered: none; files touched: listed files.
- 2026-03-09: `git status --short` - working tree verified before plan rewrite; tests run: none; bug fixes discovered: none; files touched: none.
- 2026-03-09: `Plan-only update` - replaced stale ansi-colors plan with baseline gap plan; tests run: none; bug fixes discovered: stale/out-of-sync implementation plan corrected; files touched: `IMPLEMENTATION_PLAN.md`.
- 2026-03-09: go test ./internal/baseline - pass.
- 2026-03-09: go test ./internal/config ./internal/rules - pass.
- 2026-03-09: make lint - pass.
- 2026-03-09: make arch - pass.
- 2026-03-09: git show --name-only --oneline 2589fe9 - baseline model + loader + loader tests committed with arch component registration.
- 2026-03-09: Update IMPLEMENTATION_PLAN.md - marked Phase 3 as in progress with 3.1 complete and refreshed remaining effort.
- 2026-03-09: go test ./internal/baseline - pass (writer service tests included).
- 2026-03-09: make lint - pass.
- 2026-03-09: make test-coverage - pass.
- 2026-03-09: make mutation ARGS="--diff HEAD" - pass (test efficacy 88.24%, mutator coverage 100%).
- 2026-03-09: git commit -m "Add deterministic baseline writer generation" -- internal/baseline/writer.go internal/baseline/writer_test.go - success.
- 2026-03-09: go test ./internal/config ./internal/rules - pass.
- 2026-03-09: make lint - pass.
- 2026-03-09: git commit -m "Add baseline path support to rules configuration" -- internal/config/model.go internal/config/rules.go internal/config/loader.go internal/config/loader_test.go internal/rules/model.go - success.
- 2026-03-09: go test ./internal/baseline - pass.
- 2026-03-09: make lint - pass.
- 2026-03-09: git commit -m "Add deterministic baseline comparison service" -- internal/baseline/compare.go internal/baseline/compare_test.go - success.
- 2026-03-09: Update IMPLEMENTATION_PLAN.md - marked Phase 3 complete and refreshed remaining effort.
- 2026-03-09: go test ./internal/cli -run "TestHandleAnalyzeBaselineSuppressionAffectsFailOn|TestHandleAnalyzeBaselineCompareReportsOnlyRegressions|TestHandleAnalyzeWriteBaselineIgnoresExistingContentAndReturnsZero" - fail (baseline compare/write behavior not yet integrated).
- 2026-03-09: go test ./internal/cli -run "TestHandleAnalyzeBaselineSuppressionAffectsFailOn|TestHandleAnalyzeBaselineCompareReportsOnlyRegressions|TestHandleAnalyzeWriteBaselineIgnoresExistingContentAndReturnsZero" - pass.
- 2026-03-09: go test ./internal/cli - pass.
- 2026-03-09: go test ./cmd/reglint - pass.
- 2026-03-09: go test ./... - pass.
- 2026-03-09: make test - pass.
- 2026-03-09: make lint - fail (cyclomatic complexity and long-line violations).
- 2026-03-09: make lint - pass.
- 2026-03-09: git commit -m "Integrate baseline compare and write behavior into analyze" -- internal/cli/analyze.go internal/cli/analyze_handle_test.go internal/cli/analyze_output_test.go - success.
- 2026-03-09: Update IMPLEMENTATION_PLAN.md - marked Phases 4 and 5 complete and refreshed remaining effort.
- 2026-03-09: go test ./internal/cli -run "TestRunShowsHelpForAnalyzeFlag|TestRunShowsHelpForAnalyseFlag" - fail (analyze help output missing baseline flags before implementation).
- 2026-03-09: go test ./internal/cli -run "TestRunShowsHelpForAnalyzeFlag|TestRunShowsHelpForAnalyseFlag" - pass.
- 2026-03-09: go test ./internal/cli - pass.
- 2026-03-09: go test ./cmd/reglint - pass.
- 2026-03-09: git commit -m "Add baseline flags to analyze help output" -- internal/cli/help.go internal/cli/cli_test.go - success.
- 2026-03-09: Update IMPLEMENTATION_PLAN.md - marked Phase 6.1 complete and Phase 6 as in progress.
- 2026-03-09: go test ./cmd/reglint -run "TestRunAnalyzeUsesBaselineFixtureForCompareMode|TestRunAnalyzeUsesRuleSetBaselineFixture|TestRunAnalyzeRejectsInvalidBaselineFixture" - fail (baseline fixtures not present yet).
- 2026-03-09: go test ./cmd/reglint -run "TestRunAnalyzeUsesBaselineFixtureForCompareMode|TestRunAnalyzeUsesRuleSetBaselineFixture|TestRunAnalyzeRejectsInvalidBaselineFixture" - pass.
- 2026-03-09: go test ./cmd/reglint - pass.
- 2026-03-09: go test ./internal/cli -run "Baseline|TestRunAnalyzeResolvesRuleSetBaselinePathFromConfigDirectory|TestRunAnalyzePrefersCLIBaselineOverRuleSetBaseline" - pass.
- 2026-03-09: go test ./... - pass.
- 2026-03-09: git commit -m "Add baseline fixtures and README usage examples" -- README.md cmd/reglint/main_test.go testdata/baseline/invalid-duplicate-keys.json testdata/baseline/valid-equal.json testdata/rules/baseline.yaml - success.
- 2026-03-09: edit IMPLEMENTATION_PLAN.md - marked Phase 6 complete and updated remaining effort.
- 2026-03-09: go test ./cmd/reglint -run "TestRunAnalyzeBaselineIncreaseReportsOnlyExcessRegressions|TestRunAnalyzeBaselineDecreaseDoesNotFailOnSuppressedKey" - pass.
- 2026-03-09: go test ./cmd/reglint - pass.
- 2026-03-09: git commit -m "Add baseline count-delta analyze integration tests" -- cmd/reglint/main_test.go - fail (pre-commit lint: lll in cmd/reglint/main_test.go).
- 2026-03-09: git commit -m "Add baseline count-delta analyze integration tests" -- cmd/reglint/main_test.go - success.
- 2026-03-09: Update IMPLEMENTATION_PLAN.md - marked Phase 7 as in progress and completed 7.1 baseline behavior matrix.
- 2026-03-09: go test ./cmd/reglint -run "TestRunAnalyzeCLIBaselineOverridesRuleSetBaseline|TestRunAnalyzeRejectsInvalidBaselineJSONWithSingleErrorMessage|TestRunAnalyzeRejectsInvalidBaselineSchemaVersionWithSingleErrorMessage|TestRunAnalyzeRejectsInvalidBaselineFixture" - pass.
- 2026-03-09: go test ./cmd/reglint - pass.
- 2026-03-09: go test ./... - pass.
- 2026-03-09: make lint - pass.
- 2026-03-09: git commit -m "Add baseline precedence and validation integration coverage" -- cmd/reglint/main_test.go - success.
- 2026-03-09: go test ./cmd/reglint -run "TestRunAnalyzeWriteBaselineRequiresEffectiveBaselinePath|TestRunAnalyzeWriteBaselineIgnoresExistingBaselineContent|TestRunAnalyzeWriteBaselineExitsZeroWithFailOnMatches" - pass.
- 2026-03-09: go test ./cmd/reglint - pass.
- 2026-03-09: go test ./... - pass.
- 2026-03-09: make lint - fail (cyclomatic complexity in new write-mode integration test).
- 2026-03-09: make lint - pass.
- 2026-03-09: git commit -m "Add write-mode baseline integration coverage" -- cmd/reglint/main_test.go - success.

## Summary

| Phase                                                                       | Status      |
| --------------------------------------------------------------------------- | ----------- |
| Phase 1: Scope verification and plan reset                                  | Complete    |
| Phase 2: RuleSet schema and baseline path propagation                       | Complete    |
| Phase 3: Baseline package implementation (`internal/baseline`)              | Complete    |
| Phase 4: Analyze CLI flags, precedence, and path resolution                 | Complete    |
| Phase 5: Analyze runtime integration (compare/write modes + exit semantics) | Complete    |
| Phase 6: Help text, docs, and fixture alignment                             | Complete    |
| Phase 7: End-to-end tests and regression coverage                           | Complete    |
| Phase 8: Final quality gates and release readiness                          | Not started |

**Remaining effort:** Complete Phase 8.

## Known Existing Work

- `internal/scan/engine.go` already provides deterministic match sorting and stable stats aggregation.
- `internal/cli/analyze.go` now integrates baseline path resolution, compare-mode suppression, and write-mode generation while preserving formatter contracts.
- `internal/output/console.go`, `internal/output/json.go`, and `internal/output/sarif.go` already provide deterministic formatter behavior and should remain schema-compatible.
- `internal/cli/help.go` and `internal/cli/cli_test.go` now include deterministic analyze help coverage for `--baseline`/`--write-baseline`, including the `analyse` alias help path.
- `internal/config/loader.go` + `internal/config/loader_test.go` already provide strong schema validation scaffolding for additional RuleSet fields.
- `internal/config/model.go`, `internal/config/rules.go`, and `internal/rules/model.go` now support RuleSet `baseline` propagation with copy-safe conversion.
- `cmd/reglint/main_test.go` already contains command-level integration tests that can be extended for baseline behavior.
- `internal/baseline/model.go` and `internal/baseline/loader.go` now provide baseline document structures and strict load/validation behavior.
- `internal/baseline/compare.go` now provides deterministic suppression-by-count comparison with regression-only outputs and improvement/suppression counts.
- `internal/baseline/writer.go` now provides deterministic baseline generation and canonical JSON overwrite behavior.
- `README.md` now documents baseline compare/write usage and expected exit-code behavior with executable fixture examples.
- `testdata/baseline/valid-equal.json`, `testdata/baseline/invalid-duplicate-keys.json`, and `testdata/rules/baseline.yaml` now provide deterministic fixtures for baseline compare, RuleSet baseline resolution, and validation-failure scenarios.
- `cmd/reglint/main_test.go` now includes baseline increase/decrease integration coverage to verify regression-only excess reporting and non-failing decrease behavior.
- `cmd/reglint/main_test.go` now includes baseline CLI-overrides-RuleSet precedence coverage and invalid baseline JSON/schema/duplicate validation coverage with single-line error assertions.
- `cmd/reglint/main_test.go` now includes write-mode integration coverage for missing effective path validation, ignore-existing-baseline overwrite behavior, and guaranteed exit code `0` with `--fail-on` in write mode.

## Manual Deployment Tasks

None
