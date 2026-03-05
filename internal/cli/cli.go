// Package cli provides command routing for the CLI.
package cli

import (
	"bytes"
	"fmt"
	"io"
)

// Handler handles a CLI subcommand.
type Handler func(args []string, out *bytes.Buffer) int

// Run routes CLI args to the matching handler.
func Run(args []string, handlers map[string]Handler, out io.Writer) int {
	if len(args) == 0 {
		writeHelp(out)

		return 1
	}
	if isHelpArg(args[0]) {
		writeHelp(out)

		return 0
	}

	command := args[0]
	if command == "analyse" {
		command = "analyze"
	}

	commandArgs := args[1:]

	handler, ok := handlers[command]
	if !ok {
		_, _ = fmt.Fprintf(out, "Unknown command: %s\n", args[0])

		return 1
	}

	buffer := &bytes.Buffer{}
	code := handler(commandArgs, buffer)
	_, _ = out.Write(buffer.Bytes())

	return code
}

func isHelpArg(arg string) bool {
	return arg == "--help" || arg == "-h"
}

func writeHelp(out io.Writer) {
	_, _ = fmt.Fprintln(out, "Usage:")
	_, _ = fmt.Fprintln(out, "  reglint <command> [flags]")
	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintln(out, "Commands:")
	_, _ = fmt.Fprintln(out, "  analyze (alias: analyse)")
	_, _ = fmt.Fprintln(out, "  init")
	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintln(out, "Flags:")
	_, _ = fmt.Fprintln(out, "  -h, --help bool (default false)  Print help and exit.")
}
