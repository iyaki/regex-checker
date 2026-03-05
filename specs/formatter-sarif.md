# SARIF Formatter

Status: Implemented

## Overview

### Purpose

- Provide SARIF 2.1.0 output for CI and code scanning systems.
- Preserve rule metadata and locations in SARIF-compliant structures.
- Follow shared formatter guidelines in `specs/formatter.md`.

### Goals

- Produce a minimal, valid SARIF log with deterministic ordering.
- Map scan results to SARIF results with correct severity and locations.
- Use column units aligned with rune-based column reporting.

### Non-Goals

- Rich SARIF features (code flows, fixes, partial fingerprints).
- Embedding file contents or snippets.
- Multi-run SARIF logs.

### Scope

- Render a `ScanResult` into a single-run SARIF log.
- Output to file or stdout depending on CLI flags.

## Architecture

### Module/package layout (tree format)

```
internal/
  output/
    sarif.go
```

### Component diagram (ASCII)

```
[ScanResult] -> [SARIF Formatter] -> stdout | file
```

### Data flow summary

1. Receive `ScanResult` from the scan service.
2. Build SARIF rule descriptors from the compiled rules.
3. Sort matches deterministically (see Ordering).
4. Convert matches to SARIF results and locations.
5. Write SARIF log to stdout or `--out-sarif`.

## Data model

### Core Entities

SarifLog

- Definition: Root SARIF log object.
- Fields:
  - `version` (string, required): `2.1.0`.
  - `$schema` (string, required): `https://json.schemastore.org/sarif-2.1.0.json`.
  - `runs` (array of SarifRun, required): Exactly one run.

SarifRun

- Definition: Single scan run.
- Fields:
  - `tool` (SarifTool, required)
  - `results` (array of SarifResult, required)
  - `columnKind` (string, required): `unicodeCodePoints`.

SarifTool

- Definition: Tool metadata.
- Fields:
  - `driver` (SarifToolComponent, required)

SarifToolComponent

- Definition: Tool driver descriptor.
- Fields:
  - `name` (string, required): `regex-checker`.
  - `rules` (array of SarifRuleDescriptor, required)

SarifRuleDescriptor

- Definition: Rule metadata for SARIF.
- Fields:
  - `id` (string, required): Rule id (see Rule id mapping).
  - `shortDescription.text` (string, required): Rule message template.

SarifResult

- Definition: A single SARIF result.
- Fields:
  - `ruleId` (string, required): Rule id (see Rule id mapping).
  - `level` (string, required): `error|warning|note` (see Severity mapping).
  - `message.text` (string, required): Interpolated match message.
  - `locations` (array of SarifLocation, required): Exactly one location.

SarifLocation

- Definition: Physical location in a file.
- Fields:
  - `physicalLocation.artifactLocation.uri` (string, required): File path URI.
  - `physicalLocation.region.startLine` (int, required)
  - `physicalLocation.region.startColumn` (int, required)
  - `physicalLocation.region.endLine` (int, required)
  - `physicalLocation.region.endColumn` (int, required, exclusive)

### Relationships

- `SarifResult` maps to `Match` in `specs/data-model.md`.
- `SarifRuleDescriptor` maps to a compiled rule from the ruleset.

### Persistence Notes

- SARIF output is written to stdout or a file; no persistence beyond that.

## Workflows

### Render SARIF (happy path)

1. Build rule descriptors with deterministic rule ids.
2. Sort matches (Ordering).
3. For each match, create a SARIF result with one location.
4. Write the SARIF log.

### No matches

- `runs[0].results` is an empty array.

### Error cases

- If `--format` includes `sarif` and `--out-sarif` is set, write to that file.
- If `--format` is only `sarif` and `--out-sarif` is not set, write to stdout.
- If `--format` includes `sarif` and another formatter, `--out-sarif` is required; otherwise return an error.
- Any file write or marshal error aborts the run with exit code 1.

## APIs

- Internal writer interface only. No network APIs.
- Suggested signature: `WriteSARIF(result ScanResult, out io.Writer) error`.

## Client SDK Design

- No client SDK. Formatter is internal only.

## Configuration

- CLI flag: `--format sarif`.
- Optional output path: `--out-sarif <path>`.

## Permissions

- No permissions or authentication.

## Security Considerations

- SARIF output contains interpolated messages that may include sensitive data.
- Do not embed file contents or snippets in SARIF output.

## Dependencies

- `github.com/owenrumney/go-sarif/v2/sarif` for SARIF generation.

## Verifications

- SARIF validates against the SARIF 2.1.0 schema.
- `run.columnKind` is set to `unicodeCodePoints` and columns align with rune indices.
- Ordering is deterministic across runs with identical inputs.

## Appendices

### Rule id mapping

- Rule ids are derived from rule order in the rules file.
- Format: `RC` + zero-padded 4-digit index (1-based). Example: `RC0001`.

### Severity mapping

| Rule severity | SARIF level |
| ------------- | ----------- |
| `error`       | `error`     |
| `warning`     | `warning`   |
| `notice`      | `note`      |
| `info`        | `note`      |

### Location mapping

- `startLine` and `startColumn` come from `Match.line` and `Match.column`.
- `endLine` equals `startLine` (single-line matches only).
- `endColumn` is `startColumn + runeLength(matchText)` and is exclusive.
- `artifactLocation.uri` is the match `filePath` with path separators normalized to `/`.

### Ordering

- Sort by `filePath` (ascending, byte-wise).
- Within a file, sort by `line` then `column`.
- If ties remain, sort by severity order `error > warning > notice > info`.
- If ties remain, sort by `message` (ascending, byte-wise).

### Output example

```json
{
	"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
	"version": "2.1.0",
	"runs": [
		{
			"tool": {
				"driver": {
					"name": "regex-checker",
					"rules": [
						{
							"id": "RC0001",
							"shortDescription": {
								"text": "Avoid hardcoded token: $1"
							}
						}
					]
				}
			},
			"columnKind": "unicodeCodePoints",
			"results": [
				{
					"ruleId": "RC0001",
					"level": "error",
					"message": {
						"text": "Avoid hardcoded token: abc123"
					},
					"locations": [
						{
							"physicalLocation": {
								"artifactLocation": {
									"uri": "src/auth/token.go"
								},
								"region": {
									"startLine": 12,
									"startColumn": 5,
									"endLine": 12,
									"endColumn": 18
								}
							}
						}
					]
				}
			]
		}
	]
}
```
