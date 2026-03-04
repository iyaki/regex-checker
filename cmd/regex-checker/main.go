// Command regex-checker provides the CLI entrypoint.
package main

import (
	"io"
	"os"

	"github.com/iyaki/regex-checker/internal/cli"
)

func main() {
	code := run(os.Args[1:], os.Stdout)
	os.Exit(code)
}

func run(args []string, out io.Writer) int {
	handlers := map[string]cli.Handler{}

	return cli.Run(args, handlers, out)
}
