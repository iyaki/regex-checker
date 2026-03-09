# Data Model

Status: Implemented

## Overview

### Purpose

- Define the shared entities exchanged between the CLI, scan service, and formatters.
- Keep entity fields consistent across output formats.

### Goals

- Single source of truth for core scan entities.
- Stable field names and types across formats.

### Non-Goals

- Define scanning behavior (see scan engine spec).
- Define CLI flags or configuration schema (see other specs).

## Entities

### ScanRequest

- Definition: Input to the scan service.
- Fields:
  - `roots` (list of string, required): Root paths to scan.
  - `rules` (list of Rule, required): Rules compiled from `specs/regex-rules.md`.
  - `include` (list of string, required): Effective include globs.
  - `exclude` (list of string, required): Effective exclude globs.
  - `maxFileSizeBytes` (int64, required): Skip files larger than this limit.
  - `concurrency` (int, required): Worker count for scanning.
  - `git` (GitSelectionRequest, optional): Optional Git-backed scope constraints.

### GitSelectionRequest

- Definition: Optional Git-specific scope constraints attached to `ScanRequest`.
- Fields:
  - `mode` (string, required when set): `off|staged|diff`
  - `diffTarget` (string, optional): Effective diff target for `mode=diff`.
  - `candidateFiles` (list of string, optional): Root-relative candidate files from Git selection.
  - `addedLinesByFile` (map string -> set of int, optional): 1-based added lines for each file.
  - `gitignoreEnabled` (bool, required when set): Apply `.gitignore` filtering when true.

### Match

- Definition: A single rule match with location and severity.
- Fields:
  - `message` (string, required): Interpolated message.
  - `severity` (string, required): `error|warning|notice|info`.
  - `filePath` (string, required): Path relative to scan root.
  - `line` (int, required): 1-based line number.
  - `column` (int, required): 1-based column number (rune index).
  - `matchText` (string, required): Full match text (`$0`).

### ScanStats

- Definition: Aggregated scan statistics.
- Fields:
  - `filesScanned` (int, required)
  - `filesSkipped` (int, required)
  - `matches` (int, required)
  - `durationMs` (int64, required)

### ScanResult

- Definition: Aggregated output for formatters.
- Fields:
  - `matches` (list of Match, required)
  - `stats` (ScanStats, required)

## Relationships

- CLI builds `ScanRequest`.
- Scan service produces `ScanResult`.
- Formatters consume `ScanResult`.
- Severity values align with `specs/regex-rules.md`.
- Git selection constraints narrow eligible files/lines before final match reporting.

## Notes

- Location units are rune-based and 1-based; end columns are defined by the scan engine spec.
- Formatters may require access to compiled rules from the scan pipeline; this entity model defines only match output.
- `Match` and `ScanResult` schemas remain unchanged when Git integration is used.
