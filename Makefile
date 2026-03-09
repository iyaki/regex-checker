

.PHONY: help quality all format lint test coverage test-coverage test-race mutation test-mutation \
	security arch build run analyze-example analyze-fail

GLOBAL_TARGETS ?=
BUILD_OUT ?= bin/reglint
ARGS ?=

help:
	@printf "%s\n" \
	"Common targets:" \
	"  make quality         Run full quality checks" \
	"  make format          Run gofmt on tracked Go files" \
	"  make lint            Run golangci-lint" \
	"  make test            Run tests with coverage gate" \
	"  make coverage        Run coverage gate only" \
	"  make mutation        Run mutation testing (final stage)" \
	"  make security        Run govulncheck and gosec" \
	"  make arch            Run go-arch-lint" \
	"  make build           Build CLI binary to $(BUILD_OUT)" \
	"  make run ARGS='...'  Run CLI from source" \
	"  make analyze-example Analyze test fixtures with example config" \
	"  make analyze-fail    Analyze test fixtures with failOn config"

quality: test lint test-race test-coverage test-mutation security arch

format:
	gofmt -w $$(git ls-files '*.go')

lint:
	golangci-lint run

test:
	go test ./...

coverage: test-coverage

test-coverage:
	@coverprofile="$$(mktemp -t quality-cover.XXXXXX)"; \
	go test -coverprofile="$$coverprofile" -covermode=atomic -coverpkg=./... ./...; \
	total="$$(go tool cover -func="$$coverprofile" | awk '/^total:/{gsub(/%/,"",$$3); print $$3}')"; \
	rm -f "$$coverprofile"; \
	if ! awk -v total="$$total" -v minimum="90" 'BEGIN {exit !(total >= minimum)}'; then \
		echo "Coverage $${total}% is below required 90%." >&2; \
		exit 1; \
	fi

test-race:
	go test -race ./...

test-mutation:
	gremlins unleash $(ARGS)

mutation: test-mutation

security:
	govulncheck ./...
	gosec ./...

arch:
	go-arch-lint check

build:
	go build -o $(BUILD_OUT) ./cmd/reglint

run:
	go run ./cmd/reglint $(ARGS)

analyze-example:
	go run ./cmd/reglint analyze --config testdata/rules/example.yaml testdata/fixtures

analyze-fail:
	go run ./cmd/reglint analyze --config testdata/rules/fail.yaml testdata/fixtures
