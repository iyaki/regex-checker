#!/usr/bin/env bash

set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
	echo "Go is required to install tooling." >&2
	exit 1
fi

export GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
mkdir -p "$GOBIN"

curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.10.1
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/fe3dback/go-arch-lint@latest
go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@latest
go install github.com/evilmartians/lefthook/v2@v2.1.2
