# Implementation Plan (git-integration)

**Status:** Git integration implementation in progress (Phases 9-10 complete; Phase 11 not started, 2/6 phases complete)
**Last Updated:** 2026-03-10
**Primary Specs:** `specs/git-integration.md` (related: `specs/cli-analyze.md`, `specs/configuration.md`, `specs/data-model.md`, `specs/ignore-files.md`, `specs/testing-and-validations.md`, `specs/core-architecture.md`)

## Quick Reference

| System / Subsystem                                                                                       | Specs                                                                                              | Modules / Packages                                                                                             | Artifacts                                                   | Status                                                                                             |
| -------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| RuleSet Git schema and validation (`git.mode`, `git.diff`, `git.addedLinesOnly`, `git.gitignoreEnabled`) | `specs/configuration.md`, `specs/git-integration.md`, `specs/testing-and-validations.md`           | `internal/config/model.go`, `internal/config/loader.go`, `internal/config/rules.go`, `internal/rules/model.go` | `testdata/rules/*.yaml`, `internal/config/*_test.go`        | ✅ Implemented (schema, shared conversion/defaults, cross-field validation, and fixtures complete) |
| Analyze Git flags and effective settings resolution                                                      | `specs/cli-analyze.md`, `specs/cli-help.md`, `specs/git-integration.md`                            | `internal/cli/analyze.go`, `internal/cli/help.go`, `internal/cli/cli.go`                                       | `internal/cli/*_test.go`, `cmd/reglint/main_test.go`        | Not implemented (no Git flags/help entries)                                                        |
| Git adapter and hook provider                                                                            | `specs/git-integration.md`, `specs/core-architecture.md`                                           | `internal/git/*`, `internal/hooks/*`                                                                           | package tests under `internal/git` and `internal/hooks`     | Not implemented (`internal/git` and `internal/hooks` missing)                                      |
| Scan request Git constraints and line/file scoping                                                       | `specs/data-model.md`, `specs/cli-analyze.md`, `specs/git-integration.md`                          | `internal/scan/model.go`, `internal/scan/engine.go`                                                            | `internal/scan/*_test.go`                                   | Not implemented (no Git fields in `scan.Request`)                                                  |
| Ignore-file engine foundation (`.ignore`, `.reglintignore`)                                              | `specs/ignore-files.md`                                                                            | `internal/ignore/*`, `internal/scan/ignore_rules.go`                                                           | `internal/ignore/*_test.go`, `internal/scan/ignore_test.go` | ✅ Implemented (reusable precedence foundation for Git mode)                                       |
| Deterministic scan ordering and formatter contracts                                                      | `specs/data-model.md`, `specs/formatter.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md` | `internal/scan/engine.go`, `internal/output/*`                                                                 | golden/output tests                                         | ✅ Implemented                                                                                     |
| Baseline compare/write behavior (related analyze flow)                                                   | `specs/cli-analyze-baseline.md`, `specs/cli-analyze.md`                                            | `internal/baseline/*`, `internal/cli/analyze.go`                                                               | `testdata/baseline/*`, baseline tests                       | ✅ Implemented                                                                                     |
| Git-focused docs and fixtures                                                                            | `specs/cli-analyze.md`, `specs/testing-and-validations.md`                                         | `README.md`, `testdata/rules/*`, `testdata/fixtures/*`                                                         | git-mode fixtures/examples                                  | Not implemented                                                                                    |

## Phase 9: Scope lock and stale-plan reset

**Goal:** Confirm current Git-integration gaps and replace stale plan content.
**Status:** Complete
**Paths:** `specs/git-integration.md`, `specs/cli-analyze.md`, `specs/configuration.md`, `specs/data-model.md`, `specs/ignore-files.md`, `specs/testing-and-validations.md`, `specs/core-architecture.md`, `IMPLEMENTATION_PLAN.md`
**Reference pattern:** `internal/cli/analyze.go`, `internal/scan/ignore_rules.go`

### 9.1 Spec and history verification

- [x] Verified Git integration spec exists and is linked from `specs/README.md`.
- [x] Verified scope-defining spec commit `0476869` updates Git integration plus related specs together.
- [x] Verified Git scope includes cross-domain changes in CLI, config, data model, ignore precedence, and testing.

### 9.2 Code reality and plan reset

- [x] Verified `internal/git` and `internal/hooks` packages do not exist.
- [x] Verified `internal/cli/analyze.go` and `internal/cli/help.go` do not expose Git flags/settings.
- [x] Verified `internal/config/*` and `internal/rules/model.go` do not model RuleSet `git` settings.
- [x] Verified `internal/scan/model.go` has no Git scope fields and `internal/scan/engine.go` has no Git hook path.
- [x] Replaced baseline-focused plan with this Git-integration gap plan.

**Definition of Done**

- Verification evidence is captured in the Verification Log.
- Plan scope now matches `git-integration` and current codebase reality.

**Risks/Dependencies**

- Existing specs describe Git behavior as required while code lacks implementation; this is a release-risk if not tracked explicitly.

## Phase 10: RuleSet and shared model contracts

**Goal:** Add Git settings to configuration and shared models with spec-aligned validation.
**Status:** Complete
**Paths:** `internal/config/model.go`, `internal/config/loader.go`, `internal/config/rules.go`, `internal/rules/model.go`, `internal/config/*_test.go`, `internal/rules/*_test.go`, `testdata/rules/*.yaml`
**Reference pattern:** `internal/config/model.go` and `internal/config/loader.go` existing validation style for `baseline`, `failOn`, `ignoreFiles*`

### 10.1 RuleSet schema updates

- [x] Add `git` settings struct to `config.RuleSet` with `mode`, `diff`, `addedLinesOnly`, `gitignoreEnabled`.
- [x] Add corresponding fields in `rules.RuleSet` and copy-safe conversion in `config.RuleSet.ToRules()`.
- [x] Preserve defaults expected by spec (`mode=off`, `diff` unset, `addedLinesOnly=false`, `gitignoreEnabled=true`).

### 10.2 Validation and fixture coverage

- [x] Add cross-field validation (`git.diff` valid only in `mode=diff`; `mode=diff` requires `git.diff`; `addedLinesOnly` only with `staged|diff`).
- [x] Add loader tests for valid/invalid Git config combinations and error messages.
- [x] Add sample rules fixtures for Git settings combinations under `testdata/rules/`.

**Definition of Done**

- `go test ./internal/config ./internal/rules` passes.
- RuleSet conversion exposes effective Git settings for CLI/runtime use.

**Risks/Dependencies**

- Cross-field validation can drift from CLI validation if error semantics are not centralized.

## Phase 11: Analyze CLI flags and settings precedence

**Goal:** Add Git CLI flags, help output, and effective settings resolution.
**Status:** Not started
**Paths:** `internal/cli/analyze.go`, `internal/cli/help.go`, `internal/cli/cli_test.go`, `internal/cli/analyze_test.go`, `internal/cli/scan_request_test.go`, `cmd/reglint/main_test.go`
**Reference pattern:** baseline precedence flow in `internal/cli/analyze.go` (`resolveBaselinePaths`, `prepareAnalyzeConfig`)

### 11.1 Flag parsing and help exposure

- [ ] Add `--git-mode`, `--git-diff`, `--git-added-lines-only`, and `--no-gitignore` to analyze flag parsing.
- [ ] Extend `cli.Config` with Git fields required by `specs/cli-analyze.md`.
- [ ] Update help output snapshots to include all Git flags.

### 11.2 Effective settings and validation

- [ ] Implement precedence `defaults -> RuleSet git.* -> CLI`, with `--git-diff` forcing effective `diff` mode.
- [ ] Enforce CLI validation rules (single error message, exit code 1 via existing error path).
- [ ] Thread effective Git settings into scan request assembly.

**Definition of Done**

- `go test ./internal/cli ./cmd/reglint` passes with Git-flag cases.
- Analyze help for `analyze` and `analyse` includes Git flags and defaults.

**Risks/Dependencies**

- Precedence interactions with existing baseline and ignore flags can introduce subtle regressions.

## Phase 12: Git adapter and hook infrastructure

**Goal:** Introduce Git runtime services and deterministic hook contracts.
**Status:** Not started
**Paths:** `internal/git/*.go`, `internal/hooks/*.go`, `internal/cli/analyze.go`, `internal/git/*_test.go`, `internal/hooks/*_test.go`
**Reference pattern:** deterministic root-scoped rule loading in `internal/scan/ignore_rules.go`

### 12.1 Git adapter services

- [ ] Implement Git capability checks (binary availability and repository context) gated by mode.
- [ ] Implement staged file selection and diff-target file selection with normalized root-relative slash paths.
- [ ] Implement added-line extraction (`addedLinesByFile`) from Git diff/staging output.

### 12.2 Hook model integration

- [ ] Add hook contracts for capability checks, candidate scoping, ignore augmentation, and post-match filtering.
- [ ] Register Git hooks only when Git mode is enabled; keep `mode=off` path no-op.
- [ ] Ensure hook execution order is deterministic and failures are fatal only in Git-enabled modes.

**Definition of Done**

- `go test ./internal/git ./internal/hooks` passes.
- Hook behavior matches `specs/git-integration.md` contracts and error semantics.

**Risks/Dependencies**

- Git command output parsing must stay deterministic across supported platforms.

## Phase 13: Scan engine and ignore precedence integration

**Goal:** Apply Git-selected file/line constraints while preserving current deterministic scan behavior.
**Status:** Not started
**Paths:** `internal/scan/model.go`, `internal/scan/engine.go`, `internal/scan/ignore_rules.go`, `internal/ignore/*`, `internal/scan/*_test.go`, `internal/ignore/*_test.go`
**Reference pattern:** candidate collection + deterministic sorting in `internal/scan/engine.go`

### 13.1 Candidate selection and ignore precedence

- [ ] Extend `scan.Request` with optional Git selection constraints per `specs/data-model.md`.
- [ ] Apply Git candidate selection before include/exclude and ignore evaluation.
- [ ] Implement `.gitignore` augmentation when enabled, with precedence `.gitignore` < `.ignore` < `.reglintignore`.

### 13.2 Added-lines-only filtering and deterministic results

- [ ] Apply added-lines filtering only for `mode=staged|diff` when enabled.
- [ ] Ensure files without added lines report zero matches in added-lines mode.
- [ ] Preserve deterministic ordering and unchanged formatter schemas.

**Definition of Done**

- `go test ./internal/scan ./internal/ignore` passes with Git precedence and added-line cases.
- Repeated runs with same repo state yield stable file and match ordering.

**Risks/Dependencies**

- Path normalization and root-relative mapping errors can silently drop or misattribute matches.

## Phase 14: Integration verification, docs, and quality gates

**Goal:** Close remaining gaps with end-to-end coverage, docs alignment, and quality gate evidence.
**Status:** Not started
**Paths:** `cmd/reglint/main_test.go`, `internal/cli/*_test.go`, `internal/scan/*_test.go`, `testdata/rules/*`, `testdata/fixtures/*`, `README.md`, `Makefile`
**Reference pattern:** existing command-level integration harness in `cmd/reglint/main_test.go`

### 14.1 Git behavior matrix coverage

- [ ] Add integration tests for `git-mode=off|staged|diff` success/error paths.
- [ ] Add tests for `--git-diff` implied mode, invalid diff targets, and missing Git binary handling.
- [ ] Add tests for added-lines-only output behavior and ignore precedence conflicts.

### 14.2 Docs and final quality checks

- [ ] Add README examples for Git mode usage and expected exit behavior.
- [ ] Ensure analyze help/output tests remain deterministic after Git flag additions.
- [ ] Run and log `go test ./...`, `make test`, `make lint`, and `make quality`.

**Definition of Done**

- All spec verification bullets for Git integration are covered by tests and/or reproducible commands.
- Quality gates pass and Verification Log records outcomes.

**Risks/Dependencies**

- Git-dependent tests can be flaky without strict temp-repo setup and deterministic commit/diff fixtures.

## Verification Log

- 2026-03-09: Read `specs/README.md` - confirmed `Git Integration` is indexed; tests run: none (planning mode); bug fixes discovered: none; files touched: `specs/README.md`.
- 2026-03-09: Read `specs/git-integration.md` - captured hook contracts, settings, precedence, and verification matrix; tests run: none; bug fixes discovered: none; files touched: `specs/git-integration.md`.
- 2026-03-09: Read `specs/cli-analyze.md`, `specs/configuration.md`, `specs/data-model.md`, `specs/ignore-files.md`, `specs/testing-and-validations.md`, `specs/core-architecture.md` - confirmed cross-domain Git requirements; tests run: none; bug fixes discovered: none; files touched: listed spec files.
- 2026-03-09: `git log --oneline --decorate -n 30 -- specs/git-integration.md` - confirmed latest Git scope commit `0476869`; tests run: none; bug fixes discovered: none; files touched: `specs/git-integration.md`.
- 2026-03-09: `git show --name-only --oneline 0476869` - verified related specs changed together (`cli`, `configuration`, `data-model`, `ignore-files`, `testing`, `core-architecture`); tests run: none; bug fixes discovered: none; files touched: spec files listed by command.
- 2026-03-09: `glob internal/git/**/*.go` and `glob internal/hooks/**/*.go` - no Git integration packages found; tests run: none; bug fixes discovered: none; files touched: none.
- 2026-03-09: `grep "git-mode|git\.mode|--git-diff|--no-gitignore" --include "*.go"` - no runtime Git support found in Go sources; tests run: none; bug fixes discovered: none; files touched: none.
- 2026-03-09: Read `internal/cli/analyze.go` and `internal/cli/help.go` - verified baseline/ignore features exist and Git flags are absent; tests run: none; bug fixes discovered: none; files touched: `internal/cli/analyze.go`, `internal/cli/help.go`.
- 2026-03-09: Read `internal/config/model.go`, `internal/config/loader.go`, `internal/config/rules.go`, `internal/rules/model.go` - verified RuleSet has no Git settings model/validation; tests run: none; bug fixes discovered: none; files touched: listed config/rules files.
- 2026-03-09: Read `internal/scan/model.go`, `internal/scan/engine.go`, `internal/scan/ignore_rules.go` - verified no Git request constraints/hooks and confirmed deterministic scan/ignore foundation exists; tests run: none; bug fixes discovered: none; files touched: listed scan files.
- 2026-03-09: `glob testdata/**/*git*` and read `testdata/rules/*.yaml` - verified no Git fixtures currently exist; tests run: none; bug fixes discovered: none; files touched: `testdata/rules/example.yaml`, `testdata/rules/fail.yaml`, `testdata/rules/baseline.yaml`.
- 2026-03-09: Plan-only update - replaced stale baseline-scoped plan with Git-integration plan reflecting current gaps; tests run: none; bug fixes discovered: stale plan corrected; files touched: `IMPLEMENTATION_PLAN.md`.
- 2026-03-10: `go test ./internal/config ./internal/rules` - passed.
- 2026-03-10: `go clean -testcache && go test -coverprofile=/tmp/no-cache-cover.out -covermode=atomic -coverpkg=./... ./... && go tool cover -func=/tmp/no-cache-cover.out` - passed with total coverage 94.5%.
- 2026-03-10: `git commit -m "Add Git settings to RuleSet models and conversions"` - committed Phase 10.1 model updates as `b68d18d`.
- 2026-03-10: `go test ./internal/config -run "TestLoadRuleSet(RejectsInvalidGitMode|RejectsEmptyGitDiff|RejectsGitDiffOutsideDiffMode|RejectsGitModeDiffWithoutDiff|RejectsGitAddedLinesOnlyWithoutGitMode|AllowsGitAddedLinesOnlyWithStagedMode)"` - failed first (expected for RED), then passed after implementing validation.
- 2026-03-10: `go test ./internal/config ./internal/rules` - passed with Phase 10.2 changes.
- 2026-03-10: `make lint` - passed after refactoring validation helper complexity and test line-length issues.
- 2026-03-10: `go test ./...` - passed.
- 2026-03-10: `git commit -m "Add Git config cross-field validation and fixtures"` - committed Phase 10.2 updates as `03447f4`.

## Summary

| Phase                                                       | Status      |
| ----------------------------------------------------------- | ----------- |
| Phase 9: Scope lock and stale-plan reset                    | Complete    |
| Phase 10: RuleSet and shared model contracts                | Complete    |
| Phase 11: Analyze CLI flags and settings precedence         | Not started |
| Phase 12: Git adapter and hook infrastructure               | Not started |
| Phase 13: Scan engine and ignore precedence integration     | Not started |
| Phase 14: Integration verification, docs, and quality gates | Not started |

**Remaining effort:** Implement Phases 11-14 (CLI flags/precedence, adapter/hooks, scan integration, integration coverage, and docs/quality verification).

## Known Existing Work

- `internal/ignore/loader.go`, `internal/ignore/parser.go`, `internal/ignore/matcher.go`, and `internal/scan/ignore_rules.go` already provide deterministic `.ignore`/`.reglintignore` support and are the reference for Git-ignore precedence integration.
- `internal/scan/engine.go` already provides deterministic file and match ordering and should remain the ordering reference after Git candidate filtering is added.
- `internal/cli/analyze.go` already has robust precedence patterns (baseline path resolution and ignore toggles) that should be reused for Git settings precedence.
- `internal/cli/help.go` and `internal/cli/cli_test.go` already enforce strict help snapshots and should be extended (not replaced) for Git flags.
- `internal/baseline/*` and `internal/output/*` already implement baseline compare/write behavior and stable formatter schemas that Git integration must not alter.

## Manual Deployment Tasks

None
