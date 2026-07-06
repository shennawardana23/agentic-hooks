package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_HelpListsSubcommands(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"run", "serve", "version"} {
		if !strings.Contains(out, want) {
			t.Errorf("help output missing subcommand %q\noutput:\n%s", want, out)
		}
	}
}

func TestVersionCommand_PrintsVersion(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(buf.String(), "agentic-hooks") {
		t.Errorf("version output = %q, want it to mention agentic-hooks", buf.String())
	}
}
