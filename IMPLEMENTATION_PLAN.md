# Implementation Plan (cli-help)

**Status:** Root help flag handling complete; analyze/init help pending
**Last Updated:** 2026-03-05
**Primary Specs:** `specs/cli-help.md`, `specs/cli.md`, `specs/cli-analyze.md`, `specs/cli-init.md`

## Quick Reference

| System / Subsystem           | Specs                                                   | Modules / Packages                                                               | Web Packages | Migrations / Artifacts |
| ---------------------------- | ------------------------------------------------------- | -------------------------------------------------------------------------------- | ------------ | ---------------------- |
| CLI routing baseline         | `specs/cli.md`                                          | `cmd/reglint/main.go` ✅, `internal/cli/cli.go` ✅                               | N/A          | N/A                    |
| CLI help routing             | `specs/cli-help.md`, `specs/cli.md`                     | `cmd/reglint/main.go`, `internal/cli/cli.go`                                     | N/A          | N/A                    |
| Help topics + renderer       | `specs/cli-help.md`                                     | `internal/cli/help.go` (planned)                                                 | N/A          | N/A                    |
| Analyze command help surface | `specs/cli-help.md`, `specs/cli-analyze.md`             | `internal/cli/analyze.go`                                                        | N/A          | N/A                    |
| Init command help surface    | `specs/cli-help.md`, `specs/cli-init.md`                | `internal/cli/init.go`                                                           | N/A          | N/A                    |
| CLI help tests               | `specs/cli-help.md`, `specs/testing-and-validations.md` | `internal/cli/cli_test.go`, `cmd/reglint/main_test.go`, `internal/cli/*_test.go` | N/A          | N/A                    |

## Phase 9: Help detection + routing

**Goal:** Detect `--help`/`-h` at root and subcommand level, short-circuit handlers, and exit `0`.
**Status:** In progress (baseline no-args help and unknown command behavior verified)
**Paths:** `cmd/reglint/main.go`, `internal/cli/cli.go`
**Reference patterns:** `internal/cli/cli.go`, `specs/cli-help.md` (workflows, exit codes)

### 9.1 Root help detection

- [x] Detect `reglint --help` or `reglint -h` before routing.
- [x] Emit root help output with Usage/Commands/Flags sections in required order.
- [x] Exit `0` for help requests.
- [x] Baseline no-args help prints Usage/Commands (missing Flags section and help flag handling).

### 9.2 Subcommand help detection

- [ ] For `analyze`/`analyse`, detect `--help`/`-h` in remaining args and short-circuit.
- [ ] For `init`, detect `--help`/`-h` in remaining args and short-circuit.
- [x] Preserve existing unknown command behavior (`Unknown command: <name>` and exit `1`) even when `--help` is present.

**Definition of Done**

- Help detection occurs before config loading or scans.
- Root help exits `0` without side effects.
- Unknown command handling remains unchanged.

**Risks/Dependencies**

- Must avoid invoking flag parsing for help requests to keep filesystem-independent behavior.

## Phase 10: Help topics + rendering

**Goal:** Implement structured help topics and renderer aligned with `HelpTopic` and formatting rules.
**Status:** Not started
**Paths:** `internal/cli/help.go` (new), `internal/cli/cli.go`
**Reference patterns:** `internal/cli/cli.go`, `specs/cli-help.md` (data model, formatting rules)

### 10.1 Help data model

- [ ] Define `HelpTopic` and `HelpFlag` structs in `internal/cli/help.go`.
- [ ] Encode topics for `root`, `analyze`, and `init`.
- [ ] Ensure `analyse` alias maps to the `analyze` topic.

### 10.2 Help rendering

- [ ] Render `Usage:` section and usage lines in order.
- [ ] Render `Commands:` section for root help.
- [ ] Render `Flags:` section with single-line format and `none` defaults when unset.
- [ ] Omit short flag prefix when missing while keeping alignment.

**Definition of Done**

- Output matches ordering and formatting rules in `specs/cli-help.md` examples.
- Help output uses stdout writer passed into CLI routing.

**Risks/Dependencies**

- Must stay in sync with analyze/init flag defaults and descriptions from specs.

## Phase 11: Help flag wiring + tests

**Goal:** Wire help detection into analyze/init argument parsing and add tests that lock output.
**Status:** Not started
**Paths:** `internal/cli/analyze.go`, `internal/cli/init.go`, `internal/cli/cli_test.go`, `cmd/reglint/main_test.go`
**Reference patterns:** `internal/cli/cli_test.go`, `specs/cli-help.md` verifications

### 11.1 Analyze help path

- [ ] Ensure `reglint analyze --help` and `reglint analyse -h` exit `0` and do not load config or scan.
- [ ] Include analyze flags from `specs/cli-analyze.md` plus `-h/--help`.

### 11.2 Init help path

- [ ] Ensure `reglint init --help` exits `0` and does not write files.
- [ ] Include init flags from `specs/cli-init.md` plus `-h/--help`.

### 11.3 CLI help tests

- [ ] Add tests for root help output and exit code `0`.
- [ ] Add tests for analyze/analyse help output and exit code `0`.
- [ ] Add tests for init help output and exit code `0`.
- [ ] Add tests for `reglint bogus --help` to ensure it still exits `1` and prints only the unknown command error.

**Definition of Done**

- Help paths are covered by unit tests matching exact output structure.
- No tests rely on filesystem state for help output.

**Risks/Dependencies**

- Tests must avoid reading config files or writing output files when help is requested.

## Verification Log

- 2026-03-05: `git log -n 10 -- specs/cli-help.md specs/cli.md specs/cli-analyze.md specs/cli-init.md` - reviewed recent CLI spec changes.
- 2026-03-05: Read `specs/cli-help.md`, `specs/cli.md`, `specs/cli-analyze.md`, `specs/cli-init.md` - captured help flag requirements and defaults.
- 2026-03-05: Read `cmd/reglint/main.go`, `internal/cli/cli.go` - verified routing has no `--help` handling and no-args help prints Usage/Commands only.
- 2026-03-05: Read `internal/cli/analyze.go`, `internal/cli/init.go` - verified flag parsing runs before any help short-circuit and can touch filesystem.
- 2026-03-05: Read `internal/cli/cli_test.go`, `cmd/reglint/main_test.go` - confirmed tests cover empty-args help and unknown command only.
- 2026-03-05: `go test ./internal/cli -run TestRunShowsHelpForRootFlag` - passed.

## Summary

| Phase                              | Status      |
| ---------------------------------- | ----------- |
| Phase 9: Help detection + routing  | In progress |
| Phase 10: Help topics + rendering  | Not started |
| Phase 11: Help flag wiring + tests | Not started |

**Remaining effort:** Implement help topic model + renderer, add analyze/init help detection, and add help-specific tests.

## Known Existing Work

- Root help output now includes `--help`/`-h` handling and `Flags:` section in `internal/cli/cli.go`.
- CLI routing and analyze/init handlers are implemented under `cmd/reglint/main.go` and `internal/cli/` with existing tests for routing and init/analyze behaviors.

## Manual Deployment Tasks

None.
