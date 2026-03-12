# Implementation Plan (gitignore)

**Status:** Partially Implemented (14/27 verified checklist items); core ignore pipeline and Git-mode precedence are in place, but `.gitignore` default behavior in `git-mode=off` is not yet verified/aligned with latest spec intent.
**Last Updated:** 2026-03-11
**Primary Specs:** `specs/ignore-files.md` (related: `specs/git-integration.md`, `specs/cli-analyze.md`, `specs/configuration.md`, `specs/testing-and-validations.md`, `specs/core-architecture.md`, `specs/data-model.md`)

## Quick Reference

| System / Subsystem                              | Specs                                                         | Modules / Packages                                                                                                      | Artifacts                                                                           | Status                                            |
| ----------------------------------------------- | ------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ------------------------------------------------- |
| RuleSet and defaults for ignore behavior        | `specs/ignore-files.md`, `specs/configuration.md`             | `internal/config/model.go`, `internal/config/loader.go`, `internal/config/rules.go`, `internal/rules/model.go`          | `RuleSet.ignoreFilesEnabled`, `RuleSet.ignoreFiles`, `RuleSet.git.gitignoreEnabled` | ✅ Implemented                                    |
| CLI controls for ignore and gitignore           | `specs/ignore-files.md`, `specs/cli-analyze.md`               | `internal/cli/analyze.go`, `internal/cli/help.go`                                                                       | `--no-ignore-files`, `--no-gitignore`                                               | ✅ Implemented                                    |
| Ignore parsing/loading/matching engine          | `specs/ignore-files.md`                                       | `internal/ignore/loader.go`, `internal/ignore/parser.go`, `internal/ignore/matcher.go`, `internal/scan/ignore_rules.go` | Deterministic ordered rule list with source+line metadata                           | ✅ Implemented                                    |
| Scan-order precedence include/exclude -> ignore | `specs/ignore-files.md`, `specs/git-integration.md`           | `internal/scan/engine.go`                                                                                               | `evaluateFile(...)`, `collectScanEntries(...)` ordering contract                    | ✅ Implemented                                    |
| Git hook augmentation for `.gitignore`          | `specs/git-integration.md`, `specs/cli-analyze.md`            | `internal/git/hook_provider.go`, `internal/hooks/scan_hooks.go`, `internal/cli/analyze.go`                              | `.gitignore` injected ahead of `.ignore/.reglintignore` for Git-enabled runs        | ✅ Implemented                                    |
| `.gitignore` default in non-Git mode            | `specs/ignore-files.md` (`e6c8a35`)                           | `internal/cli/analyze.go`, `internal/git/hook_provider.go`                                                              | Behavior when `--git-mode=off`                                                      | [ ] Gap (not currently confirmed in runtime path) |
| Regression and e2e coverage for precedence      | `specs/testing-and-validations.md`, `specs/e2e-test-suite.md` | `internal/scan/ignore_test.go`, `cmd/reglint/main_test.go`, `cmd/reglint/e2e_harness_*_test.go`                         | `E2E-FULL-014` and staged-mode precedence tests                                     | ✅ Implemented                                    |

## Phase 21: Scope reset and spec delta confirmation

**Goal:** Re-scope planning from stale e2e plan to current `gitignore` scope and confirm latest spec deltas.
**Status:** Complete
**Paths:** `specs/README.md`, `specs/ignore-files.md`, `specs/git-integration.md`, `specs/cli-analyze.md`, `specs/configuration.md`, `specs/testing-and-validations.md`, `IMPLEMENTATION_PLAN.md`
**Reference pattern:** `specs/ignore-files.md`, `specs/git-integration.md`

### 21.1 Spec and history audit

- [x] Verified `specs/ignore-files.md` is indexed from `specs/README.md`.
- [x] Verified latest scope-specific spec commit is `e6c8a35` (`[specs] gitignore support enabled by default`).
- [x] Verified scope crosses config, CLI, hooks, scan filtering, tests, and docs.

### 21.2 Stale-plan replacement

- [x] Verified previous `IMPLEMENTATION_PLAN.md` tracked `e2e-tests`, not `gitignore`.
- [x] Replaced plan structure and checklist with gitignore-specific gaps and verified evidence.

**Definition of Done**

- Plan scope matches `gitignore` and references latest related spec commits.
- Verification log includes exact history and code-search commands used.

**Risks/Dependencies**

- Spec wording changed recently; implementation assumptions must follow `e6c8a35`, not older behavior.

## Phase 22: Confirmed existing implementation coverage

**Goal:** Document what is already implemented to avoid duplicate work.
**Status:** Mostly Complete
**Paths:** `internal/config/*.go`, `internal/cli/analyze.go`, `internal/cli/help.go`, `internal/ignore/*.go`, `internal/scan/*.go`, `internal/git/*.go`, `internal/hooks/*.go`, `cmd/reglint/*test.go`
**Reference pattern:** `internal/scan/ignore_test.go`, `cmd/reglint/main_test.go`

### 22.1 Config and CLI surfaces

- [x] Verified config model includes `ignoreFilesEnabled`, `ignoreFiles`, and `git.gitignoreEnabled`.
- [x] Verified defaults map to ignore enabled + `.ignore/.reglintignore`, and `git.gitignoreEnabled=true`.
- [x] Verified CLI exposes `--no-ignore-files` and `--no-gitignore` in parse + help output.

### 22.2 Runtime ignore/precedence pipeline

- [x] Verified scan path order is include -> exclude -> ignore -> file-size/binary checks.
- [x] Verified ignore loader order is deterministic by directory, file list order, then line.
- [x] Verified Git hook augmentation prepends `.gitignore` before `.ignore/.reglintignore` in Git-enabled runs.
- [x] Verified conflict priority behavior is achievable via merged order + last-match-wins matcher.

### 22.3 Existing test evidence

- [x] Verified unit coverage for parser/matcher/loader and ignore validation errors.
- [x] Verified integration coverage for staged-mode precedence (`.reglintignore > .ignore > .gitignore`).
- [x] Verified e2e coverage for precedence in `E2E-FULL-014`.
- [ ] No direct test currently verifies `.gitignore` default filtering when `--git-mode=off`.

**Definition of Done**

- Existing behavior inventory is linked to concrete files/tests.
- Completed items are marked only where code evidence exists.

**Risks/Dependencies**

- Permission-based test cases are skipped on Windows in some suites, limiting cross-platform confidence for unreadable-file edges.

## Phase 23: Align runtime with default `.gitignore` across scan modes

**Goal:** Close the gap between current implementation and spec intent from `e6c8a35`.
**Status:** Not Started
**Paths:** `internal/cli/analyze.go`, `internal/git/hook_provider.go`, `internal/hooks/scan_hooks.go`, `internal/scan/ignore_rules.go`, `internal/scan/ignore_test.go`, `cmd/reglint/main_test.go`, `cmd/reglint/e2e_harness_*_test.go`
**Reference pattern:** `internal/scan/ignore_test.go`, `cmd/reglint/main_test.go:841`, `cmd/reglint/e2e_harness_internal_test.go:664`

### 23.1 Runtime behavior updates

- [ ] Ensure `.gitignore` is applied by default when `--git-mode=off` (including non-repo scans).
- [ ] Preserve no-Git dependency in `git-mode=off` while enabling `.gitignore` matching.
- [ ] Ensure `--no-gitignore` disables `.gitignore` in both Git and non-Git modes.
- [ ] Ensure RuleSet `git.gitignoreEnabled: false` disables `.gitignore` in both Git and non-Git modes.
- [ ] Preserve `--no-ignore-files` as highest-precedence global ignore disable.

### 23.2 Regression and contract tests

- [ ] Add/extend CLI test coverage for mode-off default `.gitignore` filtering.
- [ ] Add/extend CLI test coverage for mode-off `--no-gitignore` override behavior.
- [ ] Add/extend config-driven test for `git.gitignoreEnabled: false` in mode-off execution.
- [ ] Add/extend e2e scenario(s) to cover non-Git default `.gitignore` behavior at process boundary.

**Definition of Done**

- Targeted commands pass: `go test ./internal/scan ./internal/cli ./internal/git ./internal/hooks`.
- Process-level contracts pass: `go test ./cmd/reglint -run 'TestRunAnalyzeGitModeOff|TestE2EFull014|TestE2E.*Gitignore'`.
- If e2e scenario catalog changes, `make test-e2e` passes.

**Risks/Dependencies**

- Current `.gitignore` wiring is coupled to Git hooks; broadening to mode-off may require refactoring to avoid duplicate or inconsistent ignore augmentation.

## Phase 24: Documentation and verification evidence alignment

**Goal:** Keep user/developer docs and verification logs consistent with final behavior.
**Status:** Not Started
**Paths:** `README.md`, `internal/cli/help.go`, `IMPLEMENTATION_PLAN.md` (spec files are references unless explicitly requested to edit)
**Reference pattern:** `README.md:90`, `internal/cli/cli_test.go:177`

### 24.1 Documentation consistency

- [ ] Update README wording/examples to reflect that `.gitignore` is default across scan modes.
- [ ] Confirm help text and examples remain consistent for `--no-gitignore` and `--no-ignore-files`.

### 24.2 Final verification evidence

- [ ] Run and record targeted tests for new gitignore behavior and overrides.
- [ ] Run `make test` (and `make quality` if cross-cutting behavior changes) and log outcomes.

**Definition of Done**

- Verification log records exact commands and pass/fail results.
- Plan status and remaining effort reflect real repository state.

**Risks/Dependencies**

- README/help drift can leave behavior correct in code but unclear for users and CI contributors.

## Verification Log

- 2026-03-11: `git log --oneline --decorate -n 30 -- specs/ignore-files.md specs/git-integration.md specs/cli-analyze.md specs/configuration.md specs/testing-and-validations.md specs/README.md` - confirmed latest gitignore-related spec commit is `e6c8a35`; tests run: none (planning mode); bug fixes discovered: none; files touched: listed spec files.
- 2026-03-11: `git show --name-only --oneline e6c8a35` and `git show --stat --oneline e6c8a35` - verified spec delta limited to `specs/ignore-files.md` and `specs/git-integration.md`; tests run: none; bug fixes discovered: none; files touched: `specs/ignore-files.md`, `specs/git-integration.md`.
- 2026-03-11: `git show e6c8a35 -- specs/ignore-files.md specs/git-integration.md` - confirmed intent changed to default `.gitignore` behavior across scans; tests run: none; bug fixes discovered: behavior gap identified (not fixed in planning mode); files touched: spec files only.
- 2026-03-11: `grep "no-gitignore|gitignoreEnabled|ignoreFilesEnabled|ignoreFiles|no-ignore-files|reglintignore|\.gitignore"` across `*.go` - mapped implementation entry points in CLI/config/scan/git/hooks/tests; tests run: none; bug fixes discovered: none; files touched: search-only across `internal/*` and `cmd/reglint/*`.
- 2026-03-11: code read audit for `internal/cli/analyze.go`, `internal/scan/engine.go`, `internal/scan/ignore_rules.go`, `internal/git/hook_provider.go`, `internal/hooks/scan_hooks.go`, `internal/ignore/{loader,parser,matcher}.go` - verified current wiring applies `.gitignore` via Git hook augmentation, not standalone mode-off path; tests run: none; bug fixes discovered: scope gap logged for Phase 23; files touched: none.
- 2026-03-11: test read audit for `internal/scan/ignore_test.go`, `internal/git/hook_provider_test.go`, `internal/cli/{analyze_test.go,scan_request_test.go,analyze_output_test.go}`, `cmd/reglint/main_test.go`, `cmd/reglint/e2e_harness_*_test.go` - verified staged-mode precedence coverage and lack of explicit mode-off default `.gitignore` contract test; tests run: none; bug fixes discovered: none; files touched: none.
- 2026-03-11: `git status --short` - verified clean tree before plan rewrite; tests run: none; bug fixes discovered: none; files touched: none.
- 2026-03-11: Updated `IMPLEMENTATION_PLAN.md` for `gitignore` scope - replaced stale e2e plan with phase-based gitignore gap plan; tests run: none (plan-only update); bug fixes discovered: none; files touched: `IMPLEMENTATION_PLAN.md`.

## Summary

| Phase                                                               | Status          |
| ------------------------------------------------------------------- | --------------- |
| Phase 21: Scope reset and spec delta confirmation                   | Complete        |
| Phase 22: Confirmed existing implementation coverage                | Mostly Complete |
| Phase 23: Align runtime with default `.gitignore` across scan modes | Not Started     |
| Phase 24: Documentation and verification evidence alignment         | Not Started     |

**Remaining effort:** Implement and verify Phase 23 runtime/test updates (default mode-off `.gitignore` behavior + overrides), then complete Phase 24 docs and final quality evidence.

## Known Existing Work

- `internal/config` already supports `ignoreFilesEnabled`, `ignoreFiles`, and `git.gitignoreEnabled` with validation/default propagation.
- `internal/cli/analyze.go` already exposes and parses `--no-ignore-files` and `--no-gitignore`, and merges ignore settings into `scan.Request`.
- `internal/ignore` already provides deterministic loader/parser/matcher behavior with source+line error metadata.
- `internal/git/hook_provider.go` + `internal/hooks/scan_hooks.go` already support deterministic `.gitignore` augmentation for Git-enabled runs.
- `internal/scan/engine.go` already enforces include/exclude before ignore matching and keeps deterministic file/match ordering.
- `cmd/reglint/main_test.go` and `cmd/reglint/e2e_harness_*_test.go` already cover staged-mode precedence (`.reglintignore > .ignore > .gitignore`) and replayable process-level assertions.

## Manual Deployment Tasks

None
