package mcpserver_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestServeStdio_RealBinaryOverStdio builds the actual agentic-hooks binary
// and drives it as a real MCP server over stdio, using the MCP Go SDK's own
// client — the same way an external agent host (Claude Code, Cursor, etc.)
// would. This exercises the real wire transport, unlike the handler-level
// unit tests in server_test.go, which call the tool functions directly.
func TestServeStdio_RealBinaryOverStdio(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("filepath.Abs error = %v", err)
	}

	binPath := filepath.Join(t.TempDir(), "agentic-hooks-test-bin")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/agentic-hooks")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build error = %v\noutput:\n%s", err, out)
	}

	knowledgeDir := t.TempDir()
	conceptPath := filepath.Join(knowledgeDir, "solid", "single-responsibility.md")
	if err := os.MkdirAll(filepath.Dir(conceptPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	content := `---
type: principle
title: Single Responsibility Principle
tags: [solid]
---

A component should have one reason to change.
`
	if err := os.WriteFile(conceptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "agentic-hooks-test-client", Version: "v0.0.0"}, nil)
	transport := &mcp.CommandTransport{
		Command: exec.Command(binPath, "serve", "--knowledge-dir", knowledgeDir),
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("client.Connect error = %v", err)
	}
	defer session.Close()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_knowledge",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool(list_knowledge) error = %v", err)
	}
	if res.IsError {
		t.Fatalf("list_knowledge returned an error result: %+v", res.Content)
	}

	getRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_knowledge",
		Arguments: map[string]any{"id": "solid/single-responsibility"},
	})
	if err != nil {
		t.Fatalf("CallTool(get_knowledge) error = %v", err)
	}
	if getRes.IsError {
		t.Fatalf("get_knowledge returned an error result: %+v", getRes.Content)
	}
}
