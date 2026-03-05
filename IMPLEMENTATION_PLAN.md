# Implementation Plan (formatter)

**Status:** Formatter scope partially complete (SARIF aligned; core registry missing; console/JSON output schema mismatch)
**Last Updated:** 2026-03-05
**Primary Specs:** `specs/formatter.md`, `specs/formatter-console.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md`, `specs/data-model.md`, `specs/testing-and-validations.md`, `specs/cli-analyze.md`

## Quick Reference

| System / Subsystem           | Specs                                                                                                 | Modules / Packages                                                  | Web Packages | Migrations / Artifacts |
| ---------------------------- | ----------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------ | ---------------------- |
| Formatter core + registry    | `specs/formatter.md`                                                                                  | ✅ `internal/output/formatter.go`, ✅ `internal/output/registry.go` | N/A          | N/A                    |
| Console formatter            | `specs/formatter-console.md`                                                                          | ✅ `internal/output/console.go`                                     | N/A          | N/A                    |
| JSON formatter               | `specs/formatter-json.md`                                                                             | ✅ `internal/output/json.go`                                        | N/A          | N/A                    |
| SARIF formatter              | `specs/formatter-sarif.md`                                                                            | ✅ `internal/output/sarif.go`                                       | N/A          | N/A                    |
| File URI helper (non-spec)   | N/A                                                                                                   | ✅ `internal/output/file_uri.go`                                    | N/A          | N/A                    |
| Formatter tests + golden     | `specs/testing-and-validations.md`                                                                    | ✅ `internal/output/*_test.go`                                      | N/A          | ✅ `testdata/golden/*` |
| Formatter data model link    | `specs/data-model.md`                                                                                 | ✅ `internal/scan/model.go`                                         | N/A          | N/A                    |
| CLI output wiring (relation) | `specs/cli-analyze.md`, `specs/cli.md`                                                                | ✅ `internal/cli/analyze.go`, `internal/cli/cli.go`                 | N/A          | N/A                    |
| Scan engine data source      | `specs/core-architecture.md`, `specs/regex-rules.md`, `specs/configuration.md`, `specs/data-model.md` | ✅ `internal/scan/*`, `internal/rules/*`, `internal/config/*`       | N/A          | N/A                    |

## Phase 1: Formatter core contract + registry

**Goal:** Implement shared formatter interface and registry described in `specs/formatter.md`.
**Status:** Complete
**Paths:** `internal/output/formatter.go`, `internal/output/registry.go`, `internal/cli/analyze.go`
**Reference patterns:** `specs/formatter.md`

### 1.1 Formatter interface

- [x] Add `Formatter` interface with `Name() string` and `Write(result scan.Result, out io.Writer) error`.
- [x] Ensure format identifiers are lowercase and stable (`console`, `json`, `sarif`).
- [x] Keep formatters stateless and only write to the provided writer.

### 1.2 Registry + CLI integration

- [x] Add registry map for formatters and resolve requested formats from `--format`.
- [x] Return a single error for unknown formats (preserve CLI behavior in `internal/cli/analyze.go`).
- [x] Ensure CLI uses registry lookup instead of direct `switch` over format strings.

**Definition of Done**

- Formatter interface and registry exist under `internal/output/` and are wired through CLI format resolution.
- Format resolution returns deterministic errors for unknown formats.
- Unit tests cover registry resolution and duplicate format handling.

**Risks/Dependencies**

- CLI output flow currently bypasses a registry; updates must preserve existing validation and output rules.

## Phase 2: Console formatter alignment

**Goal:** Align console output with `specs/formatter-console.md` (two-line entries, absolute path line, no file URIs).
**Status:** Partial
**Paths:** `internal/output/console.go`, `internal/output/console_test.go`, `testdata/golden/console.txt`
**Reference patterns:** `specs/formatter-console.md`

### 2.1 Output shape + content

- [x] Sort matches deterministically by file path, line, column, severity, message.
- [x] Group matches by file and render a summary line.
- [x] Print `No matches found.` when there are zero matches.
- [x] Avoid emitting raw `matchText` in console output.
- [x] Render a two-line match block with a bullet-prefixed line plus an absolute path line on the next line (per spec).
- [x] Emit `<absolutePath>:<line>` on the second line (no `file://` URI).

### 2.2 File path handling

- [x] Shared file URI helper exists for absolute paths with line suffix (non-spec behavior).
- [x] Replace file URI usage with absolute path rendering per spec or update spec (if instructed).

**Definition of Done**

- Console output matches spec format including absolute path line and spacing.
- Golden tests updated to reflect console output format.
- Unit tests validate deterministic ordering and `No matches found.` behavior.

**Risks/Dependencies**

- Output changes require updating golden snapshots and any downstream consumers.

## Phase 3: JSON formatter alignment

**Goal:** Align JSON schema with `specs/formatter-json.md` (absolutePath field, no fileUri).
**Status:** Partial
**Paths:** `internal/output/json.go`, `internal/output/json_test.go`, `testdata/golden/output.json`
**Reference patterns:** `specs/formatter-json.md`

### 3.1 JSON schema compliance

- [x] Emit `schemaVersion = 1` with `matches` array and `stats` object.
- [x] Ensure deterministic ordering of matches.
- [x] Write empty `matches` array when no matches exist.
- [ ] Replace `fileUri` with `absolutePath` (`<abs-path>:<line>`) per spec.
- [ ] Keep `matchText` present in JSON output.

### 3.2 Output destination rules (CLI integration)

- [x] CLI enforces `--out-json` when multiple formats are selected.
- [x] JSON-only output uses stdout when `--out-json` is unset.

**Definition of Done**

- JSON output matches spec schema fields and naming.
- Golden JSON snapshot updated to spec shape.
- JSON unit tests updated to validate `absolutePath`.

**Risks/Dependencies**

- Schema change impacts consumers relying on `fileUri`.

## Phase 4: SARIF formatter verification

**Goal:** Confirm SARIF output adheres to `specs/formatter-sarif.md`.
**Status:** Complete
**Paths:** `internal/output/sarif.go`, `internal/output/sarif_test.go`, `testdata/golden/output.sarif`
**Reference patterns:** `specs/formatter-sarif.md`

### 4.1 SARIF rule + result mapping

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

- 2026-03-05: `git log -n 20 --oneline -- specs/formatter.md specs/formatter-console.md specs/formatter-json.md specs/formatter-sarif.md` - reviewed formatter spec history.
- 2026-03-05: Read `specs/formatter.md`, `specs/formatter-console.md`, `specs/formatter-json.md`, `specs/formatter-sarif.md` - captured formatter requirements.
- 2026-03-05: Read `specs/data-model.md`, `specs/testing-and-validations.md`, `specs/cli-analyze.md`, `specs/cli.md` - confirmed cross-cutting requirements.
- 2026-03-05: Listed `internal/output/*` and `internal/cli/*` - confirmed formatter and CLI file inventory.
- 2026-03-05: Read `internal/output/console.go`, `internal/output/json.go`, `internal/output/sarif.go`, `internal/output/file_uri.go` - verified formatter implementations.
- 2026-03-05: Read `internal/output/golden_test.go` and `testdata/golden/console.txt`, `testdata/golden/output.json`, `testdata/golden/output.sarif` - verified golden output expectations.
- 2026-03-05: Read `internal/cli/analyze.go` - verified CLI output wiring and output path validation.
- 2026-03-05: `go test ./internal/output -run TestRegistry` - passed.
- 2026-03-05: `go test ./internal/cli -run TestParseAnalyzeInvalidFormat` - passed.
- 2026-03-05: `go test ./internal/output -run TestWriteConsoleNoMatches` - passed.
- 2026-03-05: `go test ./internal/output -run TestWriteConsoleOrdersAndGroupsMatches` - passed.
- 2026-03-05: `go test ./internal/output -run TestFormatConsoleMatchLine` - passed.
- 2026-03-05: `go test ./internal/output -run TestFormatConsoleMatchLineReturnsErrorWhenCwdMissing` - passed.
- 2026-03-05: `go test ./internal/output -run TestGoldenConsoleOutput` - passed.

## Summary

| Phase                                 | Status   |
| ------------------------------------- | -------- |
| Phase 1: Formatter core + registry    | Complete |
| Phase 2: Console formatter alignment  | Complete |
| Phase 3: JSON formatter alignment     | Partial  |
| Phase 4: SARIF formatter verification | Complete |

**Remaining effort:** Align JSON outputs with spec schema.

## Known Existing Work

- Console, JSON, and SARIF formatter implementations exist under `internal/output/` with deterministic ordering.
- File URI helper exists at `internal/output/file_uri.go` and is used by console/JSON outputs (non-spec behavior).
- Formatter unit tests and golden snapshots exist under `internal/output/*_test.go` and `testdata/golden/`.
- SARIF formatter uses `github.com/owenrumney/go-sarif/v2/sarif` and sets `columnKind` to `unicodeCodePoints`.

## Manual Deployment Tasks

None.
