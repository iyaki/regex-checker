# Implementation Plan (ignore-files)

**Status:** Ignore files support in progress (2/6 phases complete)
**Last Updated:** 2026-03-07
**Primary Specs:** `specs/ignore-files.md`, `specs/configuration.md`, `specs/cli-analyze.md`, `specs/data-model.md`

## Quick Reference

| System / Subsystem              | Specs                                                                     | Modules / Packages                                                               | Artifacts | Status |
| ------------------------------- | ------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | --------- | ------ |
| CLI analyze flags + config      | `specs/ignore-files.md`, `specs/cli-analyze.md`, `specs/configuration.md` | `internal/cli/analyze.go`, `internal/cli/help.go`, `internal/config/model.go`    | N/A       | —      |
| RuleSet mapping + scan request  | `specs/ignore-files.md`, `specs/data-model.md`                            | `internal/config/rules.go`, `internal/scan/model.go`                             | N/A       | —      |
| Ignore settings + matcher       | `specs/ignore-files.md`                                                   | `internal/ignore/*`                                                              | N/A       | —      |
| Scan entry collection + filters | `specs/ignore-files.md`, `specs/data-model.md`                            | `internal/scan/engine.go`                                                        | N/A       | —      |
| Tests (ignore behavior)         | `specs/testing-and-validations.md`, `specs/ignore-files.md`               | `internal/scan/*_test.go`, `internal/cli/*_test.go`, `internal/config/*_test.go` | N/A       | —      |

## Phase 9: Ignore settings + config schema

**Goal:** Add ignore files settings to RuleSet and CLI configuration handling.
**Status:** Complete
**Paths:** `internal/config/model.go`, `internal/config/rules.go`, `internal/cli/analyze.go`, `internal/cli/help.go`
**Reference pattern:** `internal/config/model.go`

### 9.1 RuleSet schema additions

- [x] Add `ignoreFilesEnabled` and `ignoreFiles` to `config.RuleSet`.
- [x] Validate ignore file names (non-empty, no path separators, no duplicates).
- [x] Extend `rules.RuleSet` to carry ignore settings for scan requests.

### 9.2 Analyze CLI overrides

- [x] Add `--no-ignore-files` flag to analyze config parsing.
- [x] Include `--no-ignore-files` in analyze help output.

**Definition of Done**

- RuleSet YAML accepts `ignoreFilesEnabled` and `ignoreFiles`.
- `reglint analyze --help` shows `--no-ignore-files`.
- Invalid ignore file names or duplicates error during config load.

**Risks/Dependencies**

- Ensure config validation errors are surfaced before scanning begins.

## Phase 10: Ignore settings resolution

**Goal:** Resolve effective ignore settings with default, RuleSet, and CLI overrides.
**Status:** Complete
**Paths:** `internal/cli/analyze.go`, `internal/rules/model.go`, `internal/scan/model.go`
**Reference pattern:** `internal/cli/analyze.go`

### 10.1 Effective settings model

- [x] Add `IgnoreSettings` to scan request model.
- [x] Provide defaults (`enabled=true`, `files=['.ignore', '.reglintignore']`).

### 10.2 Precedence logic

- [x] Apply RuleSet overrides for `ignoreFilesEnabled` and `ignoreFiles`.
- [x] Apply `--no-ignore-files` to disable ignore support.

**Definition of Done**

- Scan requests carry resolved ignore settings.
- CLI flags override config as per spec precedence.

**Risks/Dependencies**

- Must preserve existing include/exclude override semantics.

## Phase 11: Ignore file loader + parser

**Goal:** Load, parse, and normalize ignore rules in deterministic order.
**Status:** Complete
**Paths:** `internal/ignore/loader.go`, `internal/ignore/parser.go`, `internal/ignore/matcher.go`
**Reference pattern:** `internal/scan/engine.go`

### 11.1 Ignore loader

- [x] Walk directories in lexical order per scan root.
- [x] Discover ignore files using ordered `IgnoreSettings.Files`.
- [x] Normalize line endings and read UTF-8 text.

### 11.2 Ignore parser

- [x] Parse blank lines and comments with escaped `#` and `!`.
- [x] Support negated rules, anchored `/`, trailing `/` directory-only.
- [x] Build `IgnoreRule` with base dir, source path, line number, pattern.

**Definition of Done**

- Parser reports invalid patterns with `<source>:<line>`.
- Loader preserves deterministic rule order.

**Risks/Dependencies**

- Must keep matching rules independent of OS path separator.

## Phase 12: Ignore matcher + path filtering

**Goal:** Apply ignore rules during scan entry collection.
**Status:** Not started
**Paths:** `internal/scan/engine.go`, `internal/ignore/matcher.go`
**Reference pattern:** `internal/scan/engine.go`

### 12.1 Matcher semantics

- [ ] Implement ordered rule evaluation (last match wins).
- [ ] Ensure negation un-ignores only when include/exclude allowed.

### 12.2 File selection precedence

- [ ] Apply include -> exclude -> ignore evaluation order per spec.
- [ ] Ensure ignored files are excluded from scan entries.

**Definition of Done**

- Ignore matcher yields deterministic results for same inputs.
- Scan results exclude ignored paths while preserving include/exclude logic.

**Risks/Dependencies**

- Shared usage of doublestar must match scan glob semantics.

## Phase 13: Error handling + stats

**Goal:** Surface ignore errors and track skipped file counts consistently.
**Status:** Not started
**Paths:** `internal/scan/engine.go`, `internal/ignore/*`
**Reference pattern:** `internal/scan/engine.go`

### 13.1 Loader errors

- [ ] Exit analyze with code 1 on ignore file read error.
- [ ] Bubble invalid ignore pattern errors with source/line.

### 13.2 Skipped stats integration

- [ ] Count ignored files as skipped.
- [ ] Ensure existing binary/size skip counts remain correct.

**Definition of Done**

- `filesSkipped` includes ignored + unreadable/large/binary files.
- Errors stop the run with a clear message.

**Risks/Dependencies**

- Avoid leaking ignore file contents in error output.

## Phase 14: Tests + fixtures

**Goal:** Add automated coverage for ignore files behavior.
**Status:** Not started
**Paths:** `internal/scan/engine_test.go`, `internal/cli/*_test.go`, `internal/config/*_test.go`, `testdata/fixtures`
**Reference pattern:** `internal/scan/engine_test.go`

### 14.1 Scan behavior tests

- [ ] Root `.ignore` excludes matching files.
- [ ] Nested `.reglintignore` + `!` negation re-includes paths.
- [ ] `--no-ignore-files` scans ignored files.
- [ ] Invalid ignore pattern errors with `<source>:<line>`.
- [ ] Deterministic ordering with same inputs.

### 14.2 Config + CLI tests

- [ ] Validate ignore file name constraints in config load.
- [ ] Ensure CLI flag precedence disables ignore support.

**Definition of Done**

- Tests cover all spec verifications.
- Coverage stays above 90% for touched packages.

**Risks/Dependencies**

- File system ordering tests must account for deterministic sorting.

## Verification Log

- 2026-03-07: Read `specs/README.md` - verified spec index and scope references.
- 2026-03-07: Read `specs/ignore-files.md` - documented ignore file requirements and data model.
- 2026-03-07: Read `specs/configuration.md`, `specs/cli-analyze.md`, `specs/data-model.md`, `specs/testing-and-validations.md` - confirmed related schema and validation expectations.
- 2026-03-07: Read `internal/scan/engine.go` - verified include/exclude filtering exists, no ignore support yet.
- 2026-03-07: Read `internal/scan/model.go` - scan request had no ignore settings before Phase 10 work.
- 2026-03-07: Read `internal/config/model.go` and `internal/config/rules.go` - ignore settings already present in RuleSet.
- 2026-03-07: Read `internal/cli/analyze.go` and `internal/cli/help.go` - ignore flags were already present in CLI.
- 2026-03-07: Searched `internal/**/ignore*` - ignore package does not exist yet.
- 2026-03-07: `git log --oneline -- specs` - reviewed recent spec change history.
- 2026-03-07: `git log -n 10 --oneline -- specs/ignore-files.md` - reviewed ignore spec change history.
- 2026-03-07: `go test ./internal/config ./internal/cli` - ok
- 2026-03-07: `go test ./internal/cli -run TestBuildScanRequestResolvesIgnoreSettings` - ok
- 2026-03-07: `go test ./internal/ignore` - ok
- 2026-03-07: `go test ./internal/ignore` - ok

## Summary

| Phase                                | Status      |
| ------------------------------------ | ----------- |
| Phase 9: Ignore settings + config    | Complete    |
| Phase 10: Ignore settings resolution | Complete    |
| Phase 11: Ignore loader + parser     | Complete    |
| Phase 12: Ignore matcher + filtering | Not started |
| Phase 13: Error handling + stats     | Not started |
| Phase 14: Tests + fixtures           | Not started |

**Remaining effort:** Remaining phases pending; ignore-files feature still in progress.

## Known Existing Work

- Include/exclude filtering is implemented in `internal/scan/engine.go` via `matchesPath` and `evaluateFile`.
- Default include/exclude and concurrency rules are handled in `internal/config/rules.go`.
- Scan request model is in `internal/scan/model.go` with stats tracking for skipped files.
- Binary/oversize skip handling and skipped-file counting are implemented in `internal/scan/engine.go`.

## Manual Deployment Tasks

None.
