package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunShowsHelpWhenNoArgs(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := run([]string{}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("expected usage help, got %q", output.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	code := run([]string{"bogus"}, &output)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	if output.String() != "Unknown command: bogus\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
}
