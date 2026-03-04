package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iyaki/regex-checker/internal/cli"
)

func TestRunRoutesAnalyzeAlias(t *testing.T) {
	t.Parallel()

	var called int
	var gotArgs []string
	handler := func(args []string, out *bytes.Buffer) int {
		called++
		gotArgs = append([]string{}, args...)
		_, _ = out.WriteString("ok")

		return 0
	}

	handlers := map[string]cli.Handler{
		"analyze": func(args []string, out *bytes.Buffer) int {
			return handler(args, out)
		},
	}

	var output bytes.Buffer
	code := cli.Run([]string{"analyse", "./path"}, handlers, &output)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	} else if called != 1 {
		t.Fatalf("expected handler called once, got %d", called)
	} else if len(gotArgs) != 1 || gotArgs[0] != "./path" {
		t.Fatalf("expected args [./path], got %v", gotArgs)
	}
}

func TestRunShowsHelpWhenNoCommand(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := cli.Run([]string{}, map[string]cli.Handler{}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	text := output.String()
	if !strings.Contains(text, "Usage:") {
		t.Fatalf("expected help output to include Usage, got %q", text)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := cli.Run([]string{"bogus"}, map[string]cli.Handler{}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	got := output.String()
	want := "Unknown command: bogus\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
