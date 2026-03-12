# RegLint

RegLint is a regex-based linter that scans codebases using YAML-defined rules and emits console, JSON, or SARIF outputs for local and CI workflows.

## Quickstart

```bash
reglint init
reglint analyze --config reglint-rules.yaml
```

## Development Checks

For contributor workflows, run the e2e tier that matches your intent:

- `make test-e2e-smoke` runs the fast PR-blocking smoke scenarios.
- `make test-e2e` runs the full nightly/manual matrix (smoke + full scenarios).

For complete local validation, run:

```bash
make quality
```

## CLI

```bash
reglint <command> [flags]

Commands:
  analyze (alias: analyse)
  init
```

## Config

The default rules file name is `reglint-rules.yaml`.

```bash
reglint init --out configs/reglint-rules.yaml
reglint analyze --config configs/reglint-rules.yaml
```

Console output uses ANSI severity colors by default. You can disable colors in the rules file or for a single run with `NO_COLOR`.

```yaml
consoleColorsEnabled: false
rules:
  - message: "Found token $0"
    regex: "token=[a-z]+"
    severity: "error"
```

```bash
NO_COLOR=1 reglint analyze --config reglint-rules.yaml --format console
```

## Baseline

Use baseline compare mode to suppress known findings keyed by `(filePath, message)` with count-based tolerance.

```bash
reglint analyze --config testdata/rules/fail.yaml --baseline testdata/baseline/valid-equal.json testdata/fixtures
```

This command exits `0` because the fixture match is already covered by the baseline.

You can also set a default baseline in config and run without `--baseline`:

```bash
reglint analyze --config testdata/rules/baseline.yaml testdata/fixtures
```

To regenerate a baseline from full current findings:

```bash
reglint analyze --config testdata/rules/fail.yaml --baseline testdata/baseline/generated.json --write-baseline testdata/fixtures
```

This command exits `0` on successful write, even when matches exist.

To see baseline validation behavior with invalid data:

```bash
reglint analyze --config testdata/rules/fail.yaml --baseline testdata/baseline/invalid-duplicate-keys.json testdata/fixtures
```

This command exits `1` with a single baseline validation error.

## Git-Scoped Scans

Git mode is optional and defaults to `off`, so regular scans do not require Git:

```bash
reglint analyze --config reglint-rules.yaml
```

Ignore files are enabled by default in all scan modes (`off`, `staged`, and `diff`).
When enabled, RegLint evaluates `.gitignore` first, then `.ignore`, then `.reglintignore`.

Scan only staged files:

```bash
reglint analyze --config reglint-rules.yaml --git-mode staged
```

Scan files from a diff target:

```bash
reglint analyze --config reglint-rules.yaml --git-mode diff --git-diff HEAD~1..HEAD
```

Passing `--git-diff` without `--git-mode` implies `--git-mode diff`:

```bash
reglint analyze --config reglint-rules.yaml --git-diff HEAD~1..HEAD
```

Report only matches on added lines in the selected Git scope:

```bash
reglint analyze --config reglint-rules.yaml --git-mode diff --git-diff HEAD~1..HEAD --git-added-lines-only
```

Disable only `.gitignore` filtering for one run:

```bash
reglint analyze --config reglint-rules.yaml --git-mode staged --no-gitignore
```

Disable all ignore-file processing for one run (`.gitignore`, `.ignore`, and `.reglintignore`):

```bash
reglint analyze --config reglint-rules.yaml --no-ignore-files
```

Expected exit behavior in Git-enabled runs:

- Missing Git executable, non-repository context, or invalid diff target exits `1` with a single error message.
- Empty changed-file sets are valid and complete successfully.
- Successful Git-scoped scans follow normal exit rules (`0` or `2` based on `--fail-on`).

## Outputs

- `console` writes to stdout.
- `json` writes to stdout when it is the only format; otherwise set `--out-json`.
- `sarif` writes to stdout when it is the only format; otherwise set `--out-sarif`.

```bash
reglint analyze --config reglint-rules.yaml --format console,json --out-json /tmp/scan.json
reglint analyze --config reglint-rules.yaml --format sarif --out-sarif /tmp/scan.sarif
```
