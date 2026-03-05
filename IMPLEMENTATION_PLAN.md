# Implementation Plan (formatter)

**Status:** Formatter scope partially complete (Console/JSON missing file URIs; SARIF complete)
**Last Updated:** 2026-03-05
**Primary Specs:** `specs/formatter-console.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md`, `specs/data-model.md`, `specs/testing-and-validations.md`

## Quick Reference

| System / Subsystem           | Specs                                                                                                 | Modules / Packages                                            | Web Packages | Migrations / Artifacts |
| ---------------------------- | ----------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- | ------------ | ---------------------- |
| Console formatter            | `specs/formatter-console.md`                                                                          | ✅ `internal/output/console.go`                               | N/A          | N/A                    |
| JSON formatter               | `specs/formatter-json.md`                                                                             | ✅ `internal/output/json.go`                                  | N/A          | N/A                    |
| SARIF formatter              | `specs/formatter-sarif.md`                                                                            | ✅ `internal/output/sarif.go`                                 | N/A          | N/A                    |
| Formatter tests + golden     | `specs/testing-and-validations.md`                                                                    | ✅ `internal/output/*_test.go`                                | N/A          | ✅ `testdata/golden/*` |
| Formatter data model link    | `specs/data-model.md`                                                                                 | ✅ `internal/scan/model.go`                                   | N/A          | N/A                    |
| CLI output wiring (relation) | `specs/cli-analyze.md`, `specs/cli.md`                                                                | ✅ `internal/cli/analyze.go`, `internal/cli/cli.go`           | N/A          | N/A                    |
| Scan engine data source      | `specs/core-architecture.md`, `specs/regex-rules.md`, `specs/configuration.md`, `specs/data-model.md` | ✅ `internal/scan/*`, `internal/rules/*`, `internal/config/*` | N/A          | N/A                    |

## Phase 1: Console formatter

**Goal:** Provide deterministic console output grouped by file with summary line.
**Status:** Partial
**Paths:** `internal/output/console.go`, `internal/output/console_test.go`, `testdata/golden/console.txt`
**Reference patterns:** `specs/formatter-console.md`

### 1.1 Console output structure

- [x] Sort matches deterministically by file path, line, column, severity, message.
- [x] Group matches by file and render a summary line.
- [x] Print `No matches found.` when there are zero matches.
- [x] Render file URI suffix (`file://<abs-path>:<line>`) per match line.
- [x] URL-encode file URI paths and handle Windows drive letter format.
- [x] Avoid emitting raw `matchText` in console output.

### 1.2 File URI generation (shared helper)

- [x] Add shared helper for absolute, URL-encoded file URIs with line suffix.
- [x] Ensure Windows drive letters render as `file:///C:/...:line`.
- [x] Reuse helper in console formatter.

**Definition of Done**

- Console output matches spec formatting with file URI suffixes.
- Deterministic ordering verified by unit and golden tests.

**Risks/Dependencies**

- Requires absolute path resolution for file URIs.
- Console output change impacts golden snapshots.

## Phase 2: JSON formatter

**Goal:** Provide stable JSON output with schema version, matches, and stats.
**Status:** Partial
**Paths:** `internal/output/json.go`, `internal/output/json_test.go`, `testdata/golden/output.json`
**Reference patterns:** `specs/formatter-json.md`

### 2.1 JSON output schema

- [x] Emit `schemaVersion = 1` with `matches` array and `stats` object.
- [x] Ensure deterministic ordering of matches.
- [x] Write empty `matches` array when no matches exist.
- [x] Add `fileUri` field to JSON matches (`file://<abs-path>:<line>`).
- [x] URL-encode file URI paths and handle Windows drive letter format.

### 2.2 File URI generation (shared helper)

- [x] Reuse shared file URI helper for JSON match entries.
- [x] Update JSON unit/golden tests to include file URIs.

**Definition of Done**

- JSON output matches spec schema including `fileUri` and deterministic ordering.
- Golden JSON output updated to include file URIs.

**Risks/Dependencies**

- Requires absolute path resolution for file URIs.
- JSON schema change impacts integration tests and golden snapshots.

## Phase 3: SARIF formatter

**Goal:** Provide SARIF 2.1.0 output with deterministic ordering and rule metadata.
**Status:** Complete
**Paths:** `internal/output/sarif.go`, `internal/output/sarif_test.go`, `testdata/golden/output.sarif`
**Reference patterns:** `specs/formatter-sarif.md`

### 3.1 SARIF rule + result mapping

- [x] Emit SARIF log with `version = 2.1.0`, `$schema`, single run, `columnKind = unicodeCodePoints`.
- [x] Map severities to SARIF levels (`error|warning|note`).
- [x] Map start line/column and end column using rune length.
- [x] Use normalized path (`/`) for `artifactLocation.uri` and deterministic ordering.
- [x] Confirm rule id mapping uses rule order and 1-based index with `RC0001` format.

**Definition of Done**

- SARIF output validates against schema and matches spec mapping rules.
- Golden SARIF snapshot reflects rule id format and location mapping.

**Risks/Dependencies**

- Rule index defaulting in scan engine must align with 1-based rule id mapping.

## Verification Log

- 2026-03-05: `git log -n 5 -- specs/formatter-console.md specs/formatter-json.md specs/formatter-sarif.md` - confirmed formatter spec updates.
- 2026-03-05: Read `specs/formatter-console.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md`, `specs/testing-and-validations.md` - captured formatter requirements.
- 2026-03-05: Read `internal/output/console.go`, `internal/output/json.go`, `internal/output/sarif.go` - verified formatter implementations.
- 2026-03-05: Read `internal/output/*_test.go`, `testdata/golden/*` - verified formatter tests and golden snapshots.
- 2026-03-05: Read `internal/scan/model.go` - confirmed match fields used by formatters.
- 2026-03-05: `go test ./internal/output -run "TestWriteConsoleOrdersAndGroupsMatches|TestWriteJSONOrdersMatches"` - passed.
- 2026-03-05: `UPDATE_GOLDEN=1 go test ./internal/output -run TestGolden` - passed (updated golden snapshots).

## Summary

| Phase                      | Status   |
| -------------------------- | -------- |
| Phase 1: Console formatter | Complete |
| Phase 2: JSON formatter    | Complete |
| Phase 3: SARIF formatter   | Complete |

**Remaining effort:** None.

## Known Existing Work

- Console/JSON/SARIF formatter implementations exist under `internal/output/` with deterministic ordering.
- Formatter unit tests and golden snapshots exist under `internal/output/*_test.go` and `testdata/golden/`.
- SARIF formatter uses `github.com/owenrumney/go-sarif/v2/sarif` and sets `columnKind` to `unicodeCodePoints`.

## Manual Deployment Tasks

None.
