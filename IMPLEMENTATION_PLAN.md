# Implementation Plan (e2e-tests)

**Status:** E2E spec and reference integration tests exist; compiled-binary e2e harness now includes build-once execution, typed scenarios, deterministic assertion catalog, ordering guarantees, replay diagnostics, complete smoke scenario coverage (`E2E-SMOKE-001..006`), first ten full-tier scenarios (`E2E-FULL-001..010`), and a local smoke make target (`make test-e2e-smoke`), while the remaining full-tier scenarios, `make test-e2e`, and CI tier gates remain outstanding (Phases 15-17 complete; Phases 18-19 in progress; 3/6 phases complete).
**Last Updated:** 2026-03-11
**Primary Specs:** `specs/e2e-test-suite.md` (related: `specs/testing-and-validations.md`, `specs/cli-analyze.md`, `specs/cli-init.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md`, `specs/git-integration.md`, `specs/core-architecture.md`)

## Quick Reference

| System / Subsystem                                                         | Specs                                                                                 | Modules / Packages                                                                               | Artifacts                                                        | Status                                                                                                                             |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| Canonical e2e scenario catalog and tier policy                             | `specs/e2e-test-suite.md`, `specs/testing-and-validations.md`, `specs/cli-analyze.md` | `specs/`                                                                                         | Scenario IDs `E2E-SMOKE-*`, `E2E-FULL-*`                         | ✅ Implemented (spec-only)                                                                                                         |
| Existing command-level behavior coverage (in-process, not compiled binary) | `specs/cli.md`, `specs/cli-analyze.md`, `specs/cli-init.md`                           | `cmd/reglint/main_test.go`, `internal/cli/*_test.go`                                             | CLI contract tests for baseline, git, format, help, exit codes   | ✅ Implemented                                                                                                                     |
| File-handling and deterministic ordering reference tests                   | `specs/testing-and-validations.md`, `specs/e2e-test-suite.md`                         | `internal/scan/engine_test.go`, `internal/scan/ignore_test.go`, `internal/output/golden_test.go` | Binary/oversized/unreadable handling tests, golden outputs       | ✅ Implemented                                                                                                                     |
| Compiled-binary e2e scenario harness                                       | `specs/e2e-test-suite.md`                                                             | `cmd/reglint/e2e_harness_internal_test.go`, `cmd/reglint/e2e_harness_test.go`                    | Build-once harness + typed assertion engine + replay diagnostics | Partial (foundation + assertion engine + `E2E-SMOKE-001/002/003/004/005/006` + `E2E-FULL-001/002/003/004/005/006/007/008/009/010`) |
| PR smoke and nightly/manual full make targets                              | `specs/e2e-test-suite.md`, `specs/testing-and-validations.md`                         | `Makefile`                                                                                       | `make test-e2e-smoke`, `make test-e2e`                           | Partial (`make test-e2e-smoke` implemented; `make test-e2e` missing)                                                               |
| CI gate policy for e2e tiers                                               | `specs/e2e-test-suite.md`, `specs/testing-and-validations.md`                         | `.github/workflows/*.yml`                                                                        | PR smoke e2e job, nightly scheduled full e2e job                 | Missing                                                                                                                            |
| Scenario-specific fixture workspaces for path/permission/git edge cases    | `specs/e2e-test-suite.md`                                                             | `testdata/fixtures/`, `testdata/rules/`, `testdata/baseline/`, `testdata/golden/`                | Stable fixture matrix for 21 scenarios                           | Partial (base fixtures exist; dedicated e2e matrix missing)                                                                        |

## Phase 15: Scope lock and stale-plan reset

**Goal:** Confirm e2e requirements, verify real code gaps, and replace stale plan scope.
**Status:** Complete
**Paths:** `specs/e2e-test-suite.md`, `specs/testing-and-validations.md`, `specs/cli-analyze.md`, `specs/README.md`, `Makefile`, `.github/workflows/*.yml`, `cmd/reglint/main_test.go`, `IMPLEMENTATION_PLAN.md`
**Reference pattern:** `cmd/reglint/main_test.go`, `internal/output/golden_test.go`

### 15.1 Spec and history verification

- [x] Verified `specs/e2e-test-suite.md` exists and is indexed in `specs/README.md`.
- [x] Verified e2e scope was introduced by spec commit `19a474b` touching `specs/e2e-test-suite.md`, `specs/testing-and-validations.md`, `specs/cli-analyze.md`, and `specs/README.md`.
- [x] Verified e2e scope is cross-domain (analyze/init contracts, formatter outputs, git/ignore behavior, file handling, CI policy).

### 15.2 Code reality and plan reset

- [x] Verified current repo has no dedicated compiled-binary e2e harness files (`glob **/*e2e*_test.go` and `glob **/e2e/**` returned none).
- [x] Verified `Makefile` does not provide `test-e2e-smoke` or `test-e2e` targets.
- [x] Verified workflows do not include e2e smoke/full jobs or nightly `schedule` trigger.
- [x] Verified prior `IMPLEMENTATION_PLAN.md` was scoped to `git-integration`, not `e2e-tests`.

**Definition of Done**

- Gap evidence is captured in Verification Log entries with exact commands.
- Plan scope now matches `e2e-tests` and current repository reality.

**Risks/Dependencies**

- Spec requires black-box compiled-binary execution, while existing coverage is mostly in-process; this can hide packaging/process-boundary regressions.

## Phase 16: Compiled-binary harness foundation

**Goal:** Create deterministic e2e scenario runner that executes the built `reglint` binary and reports replayable diagnostics.
**Status:** Complete
**Paths:** `cmd/reglint/e2e_harness_internal_test.go`, `cmd/reglint/e2e_harness_test.go`, `testdata/fixtures/**`, `testdata/rules/**`, `testdata/baseline/**`
**Reference pattern:** helper/test setup style in `cmd/reglint/main_test.go`

### 16.1 Binary execution harness

- [x] Verified build entrypoint already exists via `make build` (`Makefile` -> `bin/reglint`).
- [x] Add harness that builds binary once per run and executes per-scenario commands via process boundary.
- [x] Capture and assert exit code, stdout, stderr, and output artifacts per scenario.
- [x] Emit scenario ID + fixture path + replay command on every failure.

### 16.2 Scenario model and assertion engine

- [x] Introduce typed scenario definition aligned to spec fields (`id`, `tier`, `fixture`, `command`, `env`, `expectedExit`, `assertions`).
- [x] Implement deterministic assertion types: stdout/stderr contains/not-contains/regex, file exists/not-exists, JSON/SARIF field equality.
- [x] Enforce deterministic scenario ordering and stable test logs.

Progress note:

- Harness now supports `e2EScenario` with `assertions`, deterministic `sortE2EScenarios` ordering, reusable assertion evaluators for contains/not-contains/regex/file exists-file not exists/JSON field/SARIF field checks, and structured failure diagnostics with replay commands.

**Definition of Done**

- `make build` succeeds.
- `go test ./cmd/reglint -run TestE2E` passes for harness-only tests.
- Files touched: e2e harness tests under `cmd/reglint/` plus required test helpers.

**Risks/Dependencies**

- Cross-platform command quoting, temp-path normalization, and binary path resolution can make output assertions flaky if not normalized.

## Phase 17: Smoke tier implementation (PR gate)

**Goal:** Implement PR-blocking smoke scenarios `E2E-SMOKE-001..006` using the compiled binary harness.
**Status:** Complete
**Paths:** `cmd/reglint/` (new smoke scenario tests), `testdata/fixtures/**`, `testdata/e2e-fixtures/**`, `testdata/rules/**`
**Reference pattern:** behavior assertions in `cmd/reglint/main_test.go`

### 17.1 Core smoke flows

- [x] Verified in-process references exist for happy-path/fail-path/NO_COLOR flows in `cmd/reglint/main_test.go`.
- [x] Implement `E2E-SMOKE-001` analyze happy path (exit `0`, deterministic summary contract).
- [x] Implement `E2E-SMOKE-002` invalid config path/content (single actionable error, exit `1`).
- [x] Implement `E2E-SMOKE-003` fail-on threshold exceeded (exit `2`).
- [x] Implement `E2E-SMOKE-004` no-findings scenario (exit `0`).
- [x] Implement `E2E-SMOKE-005` `NO_COLOR=1` disables ANSI output.

### 17.2 Path edge and determinism

- [x] Verified path-with-space coverage currently exists only in formatter-level tests (`internal/output/file_uri_test.go`), not process-level CLI.
- [x] Implement `E2E-SMOKE-006` path containing spaces with correct path reporting.
- [x] Ensure smoke scenarios remain deterministic and non-flaky over repeated runs.

**Definition of Done**

- `make test-e2e-smoke` passes locally.
- Smoke failures print scenario IDs and replay commands.
- Files touched: smoke scenario tests and any required fixture additions.

**Risks/Dependencies**

- Platform-dependent path escaping can break path-with-space assertions if fixture setup is not normalized.

## Phase 18: Full matrix implementation (nightly/manual)

**Goal:** Implement full matrix scenarios `E2E-FULL-001..015` for baseline, formatters, git scope, file handling, ignore precedence, and ordering.
**Status:** In progress
**Paths:** `cmd/reglint/` (new full scenario tests), `testdata/baseline/**`, `testdata/rules/**`, `testdata/fixtures/**`, `testdata/golden/**`
**Reference pattern:** `cmd/reglint/main_test.go`, `internal/scan/engine_test.go`, `internal/scan/ignore_test.go`, `internal/output/golden_test.go`

### 18.1 Baseline and formatter scenarios

- [x] Verified baseline compare/write/precedence and JSON/SARIF contracts already have in-process reference coverage.
- [x] Implement compiled-binary scenario `E2E-FULL-001` (baseline compare mode suppresses non-regressions).
- [x] Implement compiled-binary scenario `E2E-FULL-002` (baseline generation mode overwrites target and exits `0`).
- [x] Implement compiled-binary scenario `E2E-FULL-003` (baseline path precedence uses `--baseline` over RuleSet `baseline`).
- [x] Implement compiled-binary scenario `E2E-FULL-004` (JSON-only format writes to stdout when `--out-json` is unset).
- [x] Implement compiled-binary scenario `E2E-FULL-005` (SARIF-only format writes to stdout when `--out-sarif` is unset).
- [x] Implement compiled-binary scenario `E2E-FULL-006` (multi-format runs require explicit `--out-json`/`--out-sarif` output paths).

### 18.2 Git mode scenarios

- [x] Verified in-process references exist for `git-mode off|staged|diff|added-lines|outside-repo|invalid-target` in `cmd/reglint/main_test.go`.
- [x] Implement compiled-binary scenario `E2E-FULL-007` (`--git-mode off` works when Git executable is unavailable).
- [x] Implement compiled-binary scenario `E2E-FULL-008` (`--git-mode staged` scans only staged files).
- [x] Implement compiled-binary scenario `E2E-FULL-009` (`--git-mode diff --git-diff <target>` scans diff-selected files).
- [x] Implement compiled-binary scenario `E2E-FULL-010` (`--git-added-lines-only` reports only matches on added lines).
- [ ] Implement compiled-binary scenario `E2E-FULL-011` (Git-enabled run outside repo exits `1` with a single error).

### 18.3 File handling, ignore precedence, and ordering

- [x] Verified lower-level references for binary/oversized/unreadable handling and deterministic ordering exist in scan/output tests.
- [ ] Implement compiled-binary scenarios `E2E-FULL-012..015` including `.reglintignore > .ignore > .gitignore` precedence.

**Definition of Done**

- `make test-e2e` passes locally with all 15 full scenario IDs.
- Repeated local full runs over identical fixture states produce identical ordering assertions.
- Files touched: full scenario tests and fixture additions/updates.

**Risks/Dependencies**

- Git and permission-dependent scenarios can vary by OS; fixtures and assertions must avoid platform-specific nondeterminism.

## Phase 19: Developer tooling and CI gate wiring

**Goal:** Add local make targets and CI enforcement for smoke/full e2e tiers.
**Status:** In progress (`make test-e2e-smoke` added; `make test-e2e` and CI jobs pending)
**Paths:** `Makefile`, `.github/workflows/quality.yml`, `.github/workflows/security.yml` (or new e2e workflow files)
**Reference pattern:** existing job structure in `.github/workflows/quality.yml`

### 19.1 Local command targets

- [x] Add `make test-e2e-smoke` for PR-required tier.
- [ ] Add `make test-e2e` for full matrix tier.
- [ ] Keep target behavior deterministic and independent from unrelated quality jobs.

### 19.2 CI policy enforcement

- [x] Verified current workflows have no e2e test jobs and no nightly `schedule` trigger.
- [ ] Add PR smoke e2e job (blocking gate).
- [ ] Add nightly/manual full e2e job.
- [ ] Ensure CI outputs include scenario-level diagnostics and replay commands.

**Definition of Done**

- PR workflow runs smoke e2e and fails on scenario failures.
- Nightly/manual workflow runs full matrix deterministically.
- Files touched: `Makefile` and workflow YAMLs.

**Risks/Dependencies**

- CI runtime budget and Git tool availability may require job tuning (caching, selective fixture setup).

## Phase 20: Verification evidence and documentation alignment

**Goal:** Produce reproducible verification evidence and align user/developer docs with implemented e2e commands.
**Status:** Not started
**Paths:** `README.md`, `Makefile`, `.github/workflows/*.yml`, `cmd/reglint/` e2e tests, `IMPLEMENTATION_PLAN.md`
**Reference pattern:** verification sections in `specs/e2e-test-suite.md` and `specs/testing-and-validations.md`

### 20.1 Verification runs and logging

- [ ] Run and record: `make build`, `make test-e2e-smoke`, `make test-e2e`, and `make quality`.
- [ ] Record scenario-level pass/fail evidence and deterministic rerun checks.
- [ ] Document files touched for each verification/fix cycle.

### 20.2 Documentation touchpoints

- [x] Verified e2e command expectations are already documented in specs.
- [ ] Update `README.md` with e2e smoke/full commands and usage expectations (if implementation scope includes docs update).
- [ ] Ensure any contributor guidance references the exact make target names.

**Definition of Done**

- Verification log contains exact commands and outcomes for all required e2e gates.
- Docs and runnable commands are consistent with implemented targets.

**Risks/Dependencies**

- Command-name drift between docs and Make/workflow wiring can cause false CI or onboarding failures.

## Verification Log

- 2026-03-10: `Read IMPLEMENTATION_PLAN.md` - confirmed existing plan was scoped to `git-integration`, not `e2e-tests`; tests run: none (planning mode); bug fixes discovered: stale scope mismatch; files touched: `IMPLEMENTATION_PLAN.md`.
- 2026-03-10: `Read specs/README.md` and `Read specs/e2e-test-suite.md` - confirmed canonical e2e spec is indexed and defines 6 smoke + 15 full scenarios; tests run: none; bug fixes discovered: none; files touched: `specs/README.md`, `specs/e2e-test-suite.md`.
- 2026-03-10: `Read specs/testing-and-validations.md` and `Read specs/cli-analyze.md` - confirmed e2e command/CI policy requirements and scenario cross-references; tests run: none; bug fixes discovered: none; files touched: `specs/testing-and-validations.md`, `specs/cli-analyze.md`.
- 2026-03-10: `git log --oneline --decorate -n 30 -- specs/e2e-test-suite.md` - confirmed latest e2e spec commit is `19a474b`; tests run: none; bug fixes discovered: none; files touched: `specs/e2e-test-suite.md`.
- 2026-03-10: `git show --name-only --oneline 19a474b` and `git show --stat --oneline 19a474b` - verified related spec updates landed together (`README`, `cli-analyze`, `testing-and-validations`, `e2e-test-suite`); tests run: none; bug fixes discovered: none; files touched: listed spec files.
- 2026-03-10: `glob **/*e2e*`, `glob **/*e2e*_test.go`, `glob **/e2e/**` - verified no dedicated e2e harness files/directories exist; tests run: none; bug fixes discovered: none; files touched: none.
- 2026-03-10: `Read Makefile` and `grep "make test-e2e-smoke|make test-e2e"` - verified required e2e make targets are missing; tests run: none; bug fixes discovered: none; files touched: `Makefile`.
- 2026-03-10: `Read .github/workflows/quality.yml`, `Read .github/workflows/security.yml`, and `grep "schedule:|test-e2e|smoke|nightly" .github/workflows/*.yml` - verified workflows have no e2e jobs and no nightly schedule trigger; tests run: none; bug fixes discovered: none; files touched: `.github/workflows/quality.yml`, `.github/workflows/security.yml`.
- 2026-03-10: `Read cmd/reglint/main_test.go` and `grep "NO_COLOR|--git-mode|--write-baseline" cmd/reglint/main_test.go` - verified broad in-process command-level coverage exists for many e2e behaviors but not compiled-binary scenario harness; tests run: none; bug fixes discovered: none; files touched: `cmd/reglint/main_test.go`.
- 2026-03-10: `Read internal/scan/engine_test.go`, `Read internal/cli/analyze_output_test.go`, and `Read internal/output/golden_test.go` - verified lower-level references exist for file handling, deterministic ordering, and formatter contracts; tests run: none; bug fixes discovered: none; files touched: listed test files.
- 2026-03-10: Plan-only update - replaced stale scope with this `e2e-tests` implementation plan and current verified gaps; tests run: none; bug fixes discovered: stale planning scope corrected; files touched: `IMPLEMENTATION_PLAN.md`.
- 2026-03-10: go test ./cmd/reglint -run TestE2EHarness - failed with undefined harness symbols before implementation (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2EHarness - passed after adding compiled-binary harness foundation.
- 2026-03-10: make build && go test ./cmd/reglint -run TestE2E - passed (build succeeds and harness-focused tests pass).
- 2026-03-10: gremlins unleash --diff HEAD - passed on rerun after a transient mutation-tool panic during commit hook.
- 2026-03-10: go test ./cmd/reglint - passed full `cmd/reglint` package tests after harness addition.
- 2026-03-10: `go test ./cmd/reglint -run TestE2EHarness` - failed first with undefined `e2EScenario`/diagnostic helper symbols (expected RED stage for scenario-diagnostics task).
- 2026-03-10: `go test ./cmd/reglint -run TestE2EHarness` - passed after introducing typed scenario metadata, reusable harness assertions, and scenario replay diagnostics.
- 2026-03-10: `make build && go test ./cmd/reglint -run TestE2E` - passed with compiled binary build and e2e-focused tests green.
- 2026-03-10: `go test ./cmd/reglint` - passed full `cmd/reglint` package after diagnostics/assertion helper changes.
- 2026-03-10: `go test ./cmd/reglint -run TestE2EHarness` - failed first with undefined assertion catalog and ordering symbols (`e2EAssertion*`, `sortE2EScenarios`) for expected RED stage.
- 2026-03-10: `go test ./cmd/reglint -run TestE2EHarness` - passed after implementing typed assertion catalog (contains/not-contains/regex/file-exists/file-not-exists/json-field/sarif-field) and deterministic scenario ordering helpers.
- 2026-03-10: `make build && go test ./cmd/reglint -run TestE2E` - passed with compiled binary build and e2e-focused suite after assertion-engine changes.
- 2026-03-10: `go test ./cmd/reglint` - passed full `cmd/reglint` package after assertion and ordering implementation.
- 2026-03-10: go test ./cmd/reglint -run TestE2ESmoke001AnalyzeHappyPathDeterministicSummary - failed with undefined `newE2ESmoke001Scenario` before implementation (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2ESmoke001AnalyzeHappyPathDeterministicSummary - passed after implementing `E2E-SMOKE-001` compiled-binary smoke scenario.
- 2026-03-10: make build && go test ./cmd/reglint -run TestE2E - passed with smoke scenario `E2E-SMOKE-001` included in harness-focused suite.
- 2026-03-10: go test ./cmd/reglint -run TestE2ESmoke002InvalidConfigSingleActionableError - failed with undefined `newE2ESmoke002Scenario` before implementation (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2ESmoke002InvalidConfigSingleActionableError - passed after implementing `E2E-SMOKE-002` compiled-binary smoke scenario.
- 2026-03-10: make build && go test ./cmd/reglint -run TestE2E - passed with smoke scenarios `E2E-SMOKE-001` and `E2E-SMOKE-002` included in harness-focused suite.
- 2026-03-10: go test ./cmd/reglint - passed full `cmd/reglint` package after `E2E-SMOKE-002` addition.
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke003FailOnThresholdExceeded` - failed with undefined `newE2ESmoke003Scenario` before implementation (expected RED stage).
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke003FailOnThresholdExceeded` - passed after implementing `E2E-SMOKE-003` compiled-binary smoke scenario.
- 2026-03-10: `make build && go test ./cmd/reglint -run TestE2E` - passed with smoke scenarios `E2E-SMOKE-001..003` included in harness-focused suite.
- 2026-03-10: `go test ./cmd/reglint` - passed full `cmd/reglint` package after `E2E-SMOKE-003` addition.
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke004NoFindingsExitZero` - failed with undefined `newE2ESmoke004Scenario` before implementation (expected RED stage).
- 2026-03-10: `make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint` - failed after initial fixture placement under `testdata/fixtures/no-findings` because `E2E-SMOKE-001/003` summary file counts changed from `1` to `2`.
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke004NoFindingsExitZero` - passed after implementing `E2E-SMOKE-004` and moving no-findings data to `testdata/e2e-fixtures/no-findings`.
- 2026-03-10: `make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint` - passed with smoke scenarios `E2E-SMOKE-001..004` and full `cmd/reglint` package tests green.
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke005NoColorDisablesANSIOutput` - failed first with undefined `newE2ESmoke005Scenario` before implementation (expected RED stage).
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke005NoColorDisablesANSIOutput` - passed after implementing `E2E-SMOKE-005` compiled-binary smoke scenario with `NO_COLOR=1` assertion.
- 2026-03-10: `make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint` - passed with smoke scenarios `E2E-SMOKE-001..005` and full `cmd/reglint` package tests green.
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke006PathWithSpacesCorrectPathReporting` - failed first with undefined `newE2ESmoke006Scenario` before implementation (expected RED stage).
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke006PathWithSpacesCorrectPathReporting` - failed with missing fixture path `testdata/e2e-fixtures/path with spaces` before fixture creation.
- 2026-03-10: `go test ./cmd/reglint -run TestE2ESmoke006PathWithSpacesCorrectPathReporting` - passed after implementing `E2E-SMOKE-006` and adding path-with-spaces fixture.
- 2026-03-10: `make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint -run 'TestE2ESmoke00(1|2|3|4|5|6)' -count=3 && go test ./cmd/reglint` - passed; smoke scenarios `E2E-SMOKE-001..006` deterministic over repeated runs and full `cmd/reglint` package tests green.
- 2026-03-10: `make test-e2e-smoke` - passed (`go test -count=1 ./cmd/reglint -run '^TestE2ESmoke'` returned `ok`).
- 2026-03-10: `make build` - passed (`go build -o bin/reglint ./cmd/reglint`).

## Summary

| Phase                                                       | Status      |
| ----------------------------------------------------------- | ----------- |
| Phase 15: Scope lock and stale-plan reset                   | Complete    |
| Phase 16: Compiled-binary harness foundation                | Complete    |
| Phase 17: Smoke tier implementation (PR gate)               | Complete    |
| Phase 18: Full matrix implementation (nightly/manual)       | In progress |
| Phase 19: Developer tooling and CI gate wiring              | In progress |
| Phase 20: Verification evidence and documentation alignment | Not started |

**Remaining effort:** Wire 5 remaining full-tier scenario IDs on top of the completed smoke/harness assertion model, add `make test-e2e`, add PR/nightly CI gates, and capture full execution evidence.

## Known Existing Work

- `cmd/reglint/main_test.go` already provides extensive command-level behavior assertions for baseline, git mode, formatter outputs, help, and exit codes; reuse these as scenario assertion references.
- `internal/cli/analyze_output_test.go` already validates JSON/SARIF output path and multi-format constraints needed by full-tier formatter scenarios.
- `internal/scan/engine_test.go` already validates binary/oversized skipping, unreadable-file handling, and added-lines behavior; these are strong references for full-tier edge scenarios.
- `internal/scan/ignore_test.go` already covers ignore precedence and git candidate-scope ordering; use this as canonical precedence behavior.
- `internal/output/golden_test.go` and `testdata/golden/*` already enforce deterministic formatter output ordering.
- `Makefile` now provides `make test-e2e-smoke` for compiled-binary smoke scenarios alongside existing `make build`, `make test`, and `make quality` targets.
- `cmd/reglint/e2e_harness_internal_test.go` and `cmd/reglint/e2e_harness_test.go` now provide a compiled-binary build-once harness with typed scenario metadata, deterministic assertion catalog, scenario ordering helpers, and replayable failure diagnostics.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2ESmoke001Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-SMOKE-001` as a compiled-binary smoke contract.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2ESmoke002Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-SMOKE-002` for invalid-config single-error process-level behavior.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2ESmoke003Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-SMOKE-003` for fail-on threshold exit-code behavior.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2ESmoke004Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-SMOKE-004` for the process-level no-findings exit-zero contract using dedicated fixtures under `testdata/e2e-fixtures/no-findings`.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2ESmoke005Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-SMOKE-005` for `NO_COLOR=1` ANSI-disable process-level behavior.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2ESmoke006Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-SMOKE-006` for path-with-spaces process-level path reporting using fixture `testdata/e2e-fixtures/path with spaces/sample file.txt`.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull001Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-001` for baseline-compare suppression behavior at the compiled-binary process boundary.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull002Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-002` for baseline generation overwrite + exit-zero behavior at the compiled-binary process boundary.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull003Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-003` for baseline path precedence behavior where `--baseline` overrides RuleSet `baseline` at the compiled-binary process boundary.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull004Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-004` for JSON-only stdout behavior when `--out-json` is unset.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull005Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-005` for SARIF-only stdout behavior when `--out-sarif` is unset.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull006Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-006` for process-level multi-format output-path validation requiring explicit `--out-json`/`--out-sarif` in combined format runs.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull007Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-007` for process-level `--git-mode off` behavior when Git executable is unavailable.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull008Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-008` for process-level `--git-mode staged` staged-file-only behavior in a temporary Git repository fixture.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull009Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-009` for process-level `--git-mode diff --git-diff HEAD` diff-selected-file behavior in a temporary Git repository fixture.
- `cmd/reglint/e2e_harness_internal_test.go` now defines `newE2EFull010Scenario(...)`, and `cmd/reglint/e2e_harness_test.go` executes `E2E-FULL-010` for process-level `--git-added-lines-only` staged added-line filtering behavior in a temporary Git repository fixture.

## Manual Deployment Tasks

None

## Verification Log Addendum

- 2026-03-10: go test ./cmd/reglint -run TestE2EFull001BaselineCompareSuppressesNonRegressions - failed with undefined `newE2EFull001Scenario` (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull001BaselineCompareSuppressesNonRegressions - passed after implementing `E2E-FULL-001` compiled-binary baseline-compare scenario.
- 2026-03-10: go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001` and full `cmd/reglint` package tests green.
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull002BaselineWriteOverwritesTargetAndExitsZero - failed with undefined `newE2EFull002Scenario` (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull002BaselineWriteOverwritesTargetAndExitsZero - passed after implementing `E2E-FULL-002` compiled-binary baseline-write scenario.
- 2026-03-10: make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..002` and full `cmd/reglint` package tests green.
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull003BaselinePathPrecedenceCLIOverridesRuleSet - failed with undefined `newE2EFull003Scenario` (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull003BaselinePathPrecedenceCLIOverridesRuleSet - passed after implementing `E2E-FULL-003` compiled-binary baseline precedence scenario.
- 2026-03-10: make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..003` and full `cmd/reglint` package tests green.
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull004JSONOnlyFormatWritesToStdoutWhenOutPathUnset - failed with undefined `newE2EFull004Scenario` (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull004JSONOnlyFormatWritesToStdoutWhenOutPathUnset - passed after implementing `E2E-FULL-004` compiled-binary JSON-only stdout scenario.
- 2026-03-10: make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..004` and full `cmd/reglint` package tests green.
- 2026-03-10: Updated IMPLEMENTATION_PLAN.md - marked `E2E-FULL-004` complete and refreshed remaining full-tier effort.
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull005SARIFOnlyFormatWritesToStdoutWhenOutPathUnset - failed with undefined `newE2EFull005Scenario` (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull005SARIFOnlyFormatWritesToStdoutWhenOutPathUnset - passed after implementing `E2E-FULL-005` compiled-binary SARIF-only stdout scenario.
- 2026-03-10: go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..005` and full `cmd/reglint` package tests green.
- 2026-03-10: Updated IMPLEMENTATION_PLAN.md - marked `E2E-FULL-005` complete and refreshed remaining full-tier effort.
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull006MultiFormatRequiresExplicitOutputPaths - failed with undefined `newE2EFull006Scenario` (expected RED stage).
- 2026-03-10: go test ./cmd/reglint -run TestE2EFull006MultiFormatRequiresExplicitOutputPaths - passed after implementing `E2E-FULL-006` compiled-binary multi-format output-path validation scenario.
- 2026-03-10: make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..006` and full `cmd/reglint` package tests green.
- 2026-03-10: Updated IMPLEMENTATION_PLAN.md - marked `E2E-FULL-006` complete and refreshed remaining full-tier effort.
- 2026-03-11: go test ./cmd/reglint -run TestE2EFull007GitModeOffWorksWhenGitExecutableUnavailable - failed with undefined `newE2EFull007Scenario` (expected RED stage).
- 2026-03-11: go test ./cmd/reglint -run TestE2EFull007GitModeOffWorksWhenGitExecutableUnavailable - passed after implementing `E2E-FULL-007` compiled-binary git-mode-off scenario.
- 2026-03-11: make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..007` and full `cmd/reglint` package tests green.
- 2026-03-11: Updated IMPLEMENTATION_PLAN.md - marked `E2E-FULL-007` complete and refreshed remaining full-tier effort.
- 2026-03-11: go test ./cmd/reglint -run TestE2EFull008GitModeStagedScansOnlyStagedFiles - failed with undefined `newE2EFull008Scenario` (expected RED stage).
- 2026-03-11: go test ./cmd/reglint -run TestE2EFull008GitModeStagedScansOnlyStagedFiles - passed after implementing `E2E-FULL-008` compiled-binary git-mode-staged scenario.
- 2026-03-11: make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..008` and full `cmd/reglint` package tests green.
- 2026-03-11: Updated IMPLEMENTATION_PLAN.md - marked `E2E-FULL-008` complete and refreshed remaining full-tier effort.
- 2026-03-11: go test ./cmd/reglint -run TestE2EFull009GitModeDiffScansOnlyDiffSelectedFiles - failed with undefined `newE2EFull009Scenario` (expected RED stage).
- 2026-03-11: go test ./cmd/reglint -run TestE2EFull009GitModeDiffScansOnlyDiffSelectedFiles - passed after implementing `E2E-FULL-009` compiled-binary git-mode-diff scenario.
- 2026-03-11: make build && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..009` and full `cmd/reglint` package tests green.
- 2026-03-11: Updated IMPLEMENTATION_PLAN.md - marked `E2E-FULL-009` complete and refreshed remaining full-tier effort.
- 2026-03-11: go test ./cmd/reglint -run TestE2EFull010GitAddedLinesOnlyReportsOnlyMatchesOnAddedLines - failed with undefined `newE2EFull010Scenario` (expected RED stage).
- 2026-03-11: go test ./cmd/reglint -run TestE2EFull010GitAddedLinesOnlyReportsOnlyMatchesOnAddedLines - passed after implementing `E2E-FULL-010` compiled-binary git added-lines-only scenario.
- 2026-03-11: make build && go test ./cmd/reglint -run TestE2EFull010GitAddedLinesOnlyReportsOnlyMatchesOnAddedLines && go test ./cmd/reglint -run TestE2E && go test ./cmd/reglint - passed with `E2E-FULL-001..010` and full `cmd/reglint` package tests green.
- 2026-03-11: Updated IMPLEMENTATION_PLAN.md - marked `E2E-FULL-010` complete and refreshed remaining full-tier effort.
