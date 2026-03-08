#!/usr/bin/env bash

set -euo pipefail

mode="${1:-all}"
global_targets="$(echo "${2:-}" | tr -d '"')"

require_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "Missing required tool: $1" >&2
		exit 1
	fi
}

has_go_files() {
	[ -n "$(git ls-files '*.go')" ]
}

run_go_fmt() {
	require_cmd gofmt
	git ls-files -z '*.go' >/tmp/quality-go-files
	if [ -s /tmp/quality-go-files ]; then
		xargs -0 gofmt -w </tmp/quality-go-files
	fi
}

run_golangci() {
	require_cmd golangci-lint
	if ! has_go_files; then
		echo "No Go files found; skipping golangci-lint."
		return 0
	fi
	golangci-lint run
}

run_test_coverage() {
	require_cmd go
	if ! has_go_files; then
		echo "No Go files found; skipping test coverage."
		return 0
	fi

	local coverprofile
	coverprofile="$(mktemp -t quality-cover.XXXXXX)"

	local packages
	packages="$(go list ./...)"
	packages="$(printf '%s\n' "$packages" | grep -v '/integration' | grep -v '/testdata' || true)"
	if [ -z "$packages" ]; then
		echo "No Go packages found after exclusions; skipping test coverage."
		rm -f "$coverprofile"
		return 0
	fi

	go test -covermode=atomic -coverprofile="$coverprofile" $packages

	local total
	total="$(go tool cover -func="$coverprofile" | awk '/^total:/{gsub(/%/,"",$3); print $3}')"
	rm -f "$coverprofile"

	if [ -z "$total" ]; then
		echo "Unable to determine total coverage." >&2
		exit 1
	fi

	local minimum
	minimum="${COVERAGE_MIN:-90}"
	if ! awk -v total="$total" -v minimum="$minimum" 'BEGIN {exit !(total >= minimum)}'; then
		echo "Coverage ${total}% is below required ${minimum}%." >&2
		exit 1
	fi
}

run_mutation_testing() {
	require_cmd go-mutesting
	if ! has_go_files; then
		echo "No Go files found; skipping mutation testing."
		return 0
	fi

	if [ -n "${SKIP_MUTATION_TESTING:-}" ]; then
		echo "Skipping mutation testing (SKIP_MUTATION_TESTING set)."
		return 0
	fi

	local output_file
	output_file="$(mktemp -t quality-mutesting.XXXXXX)"

	local targets
	targets="${global_targets:-}"
	if [ -z "$targets" ]; then
		targets="${MUTATION_TARGETS:-./...}"
	fi

	if ! go-mutesting $targets | tee "$output_file"; then
		rm -f "$output_file" report.json go-mutesting-report.html
		exit 1
	fi

	local score
	score="$(awk '/^The mutation score is/{score=$5} END {print score}' "$output_file")"
	rm -f "$output_file" report.json go-mutesting-report.html

	if [ -z "$score" ]; then
		echo "Unable to determine mutation score." >&2
		exit 1
	fi

	local minimum
	minimum="${MUTATION_SCORE_MIN:-0.8}"
	if ! awk -v score="$score" -v minimum="$minimum" 'BEGIN {exit !(score >= minimum)}'; then
		echo "Mutation score ${score} is below required ${minimum}." >&2
		exit 1
	fi
}

run_test() {
	run_test_coverage
}

run_govulncheck() {
	require_cmd govulncheck
	if ! has_go_files; then
		echo "No Go files found; skipping govulncheck."
		return 0
	fi
	govulncheck ./...
}

run_gosec() {
	require_cmd gosec
	if ! has_go_files; then
		echo "No Go files found; skipping gosec."
		return 0
	fi
	gosec ./...
}

run_go_arch_lint() {
	require_cmd go-arch-lint
	if ! has_go_files; then
		echo "No Go files found; skipping go-arch-lint."
		return 0
	fi
	go-arch-lint check
}

case "$mode" in
gofmt)
	run_go_fmt
	;;
golangci | golangci-lint)
	run_golangci
	;;
govulncheck)
	run_govulncheck
	;;
gosec)
	run_gosec
	;;
go-arch-lint | arch)
	run_go_arch_lint
	;;
format)
	run_go_fmt
	;;
lint)
	run_golangci
	;;
test)
	run_test
	;;
coverage)
	run_test_coverage
	;;
mutation)
	run_mutation_testing
	;;
security)
	run_govulncheck
	run_gosec
	;;
all)
	run_go_fmt
	run_golangci
	run_test
	run_govulncheck
	run_gosec
	run_go_arch_lint
	;;
*)
	echo "Usage: $0 [all|format|lint|test|coverage|mutation|security|arch|gofmt|golangci|govulncheck|gosec|go-arch-lint]" >&2
	exit 1
	;;
esac
