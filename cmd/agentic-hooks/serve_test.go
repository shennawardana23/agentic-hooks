package main

import "testing"

func TestServeCmd_RequiresKnowledgeDirFlag(t *testing.T) {
	cmd := newServeCmd()
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("Execute() error = nil, want error because --knowledge-dir is required")
	}
}
