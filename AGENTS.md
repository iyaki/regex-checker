# Agent Guidelines

## Spec-First Workflow

- Read `specs/README.md` before any feature work.
- Assume specs describe intent, not implementation.
- Verify reality in the codebase before claiming something exists.
- Implement to spec patterns and data shapes; update specs only when asked.
- When Writting specs, **NEVER** follow Test Driven Development practices. Write the spec first and stop.

## Testing and Quality Gates

- Follow Test Driven Development practices: write failing tests before implementation.
- Local suite: `bash scripts/quality.sh all`.
- Targeted runs:
  - `bash scripts/quality.sh lint|test|coverage|security|arch`.
- Coverage gate: default min 90% (`COVERAGE_MIN` override).
- Run core tests with `go test ./...`.
- Execute mutation testing with `bash scripts/quality.sh mutation` ONLY in final stages of the task development. **NEVER** execute mutation testing during the Test Driven Development process.

## Build and Run

- Build the CLI binary: `go build -o bin/reglint ./cmd/reglint`.
- Run from source (no build): `go run ./cmd/reglint <command> [flags]`.
- Example with test data (config + fixture): `go run ./cmd/reglint analyze --config testdata/rules/example.yaml testdata/fixtures`.
- Example with failing config (failOn): `go run ./cmd/reglint analyze --config testdata/rules/fail.yaml testdata/fixtures`.

## Tooling Expectations

- Go version: 1.25 (see `go.mod`).
- Mutation testing tool: `go-mutesting`.
- Lint and security via `golangci-lint`, `govulncheck`, `go-arch-lint`, `go-fmt`.

## Implementation Guidance

- Keep scans deterministic and reproducible.
- Skip binary/oversized files per spec; record skipped file stats.
- Treat match text as sensitive; avoid logging it in console.
- When multiple code paths do similar work with small variations, consolidate into shared services with request structs.
