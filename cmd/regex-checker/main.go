// Command regex-checker provides the CLI entrypoint.
package main

import (
	"io"
	"os"

	"github.com/iyaki/regex-checker/internal/cli"
)

var (
	args                   = os.Args
	outputWriter io.Writer = os.Stdout
	exitFunc               = os.Exit
)

func main() {
	code := run(args[1:], outputWriter)
	exitFunc(code)
}

func run(args []string, out io.Writer) int {
	handlers := map[string]cli.Handler{
		"analyze": cli.HandleAnalyze,
	}

	return cli.Run(args, handlers, out)
}
