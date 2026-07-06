package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptForApprovalAndRecordFeedback_ParsesApproval(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"lowercase y approves", "y\n", true},
		{"uppercase Y approves", "Y\n", true},
		{"n rejects", "n\n", false},
		{"empty line rejects (fail-closed)", "\n", false},
		{"arbitrary text rejects (fail-closed)", "yes\n", false},
		{"EOF with no trailing newline rejects", "y", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			in := bufio.NewReader(strings.NewReader(tt.input + "some reason\n"))
			feedbackDir := t.TempDir()

			got := promptForApprovalAndRecordFeedback(out, in, feedbackDir, "task", "result")
			if got != tt.want {
				t.Errorf("promptForApprovalAndRecordFeedback() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptForApprovalAndRecordFeedback_WarnsOnFeedbackWriteFailureButKeepsApproval(t *testing.T) {
	out := &bytes.Buffer{}
	in := bufio.NewReader(strings.NewReader("y\nreason\n"))

	// feedback.Append does os.MkdirAll(feedbackDir, ...); pointing feedbackDir
	// through an existing regular file makes MkdirAll fail, forcing the
	// write-failure path without needing filesystem permission tricks.
	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	feedbackDir := filepath.Join(blocker, "feedback")

	got := promptForApprovalAndRecordFeedback(out, in, feedbackDir, "task", "result")

	if !got {
		t.Error("promptForApprovalAndRecordFeedback() = false, want true — the human's approval must stand even if the feedback write fails")
	}
	if !strings.Contains(out.String(), "warning: failed to write feedback record") {
		t.Errorf("output = %q, want a warning about the failed feedback write", out.String())
	}
}
