package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"agentic-hooks/internal/secondbrain"
)

func testBrain(t *testing.T) *secondbrain.Brain {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "solid", "single-responsibility.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	content := `---
type: principle
title: Single Responsibility Principle
tags: [solid]
---

A component should have one reason to change.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	brain, err := secondbrain.Load(dir)
	if err != nil {
		t.Fatalf("secondbrain.Load error = %v", err)
	}
	return brain
}

func TestListKnowledge_ReturnsAllConceptsWhenNoFilter(t *testing.T) {
	brain := testBrain(t)
	_, out, err := listKnowledgeHandler(brain)(context.Background(), nil, ListKnowledgeInput{})
	if err != nil {
		t.Fatalf("listKnowledgeHandler error = %v", err)
	}
	if len(out.Concepts) != 1 || out.Concepts[0].ID != "solid/single-responsibility" {
		t.Errorf("Concepts = %v, want one concept with id solid/single-responsibility", out.Concepts)
	}
}

func TestListKnowledge_SurfacesSkippedFilesAsWarnings(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.md"), []byte("---\ntitle: no type field\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	brain, err := secondbrain.Load(dir)
	if err != nil {
		t.Fatalf("secondbrain.Load error = %v", err)
	}

	_, out, err := listKnowledgeHandler(brain)(context.Background(), nil, ListKnowledgeInput{})
	if err != nil {
		t.Fatalf("listKnowledgeHandler error = %v", err)
	}
	if len(out.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1 entry for the skipped file", out.Warnings)
	}
}

func TestGetKnowledge_ReturnsBodyForKnownID(t *testing.T) {
	brain := testBrain(t)
	_, out, err := getKnowledgeHandler(brain)(context.Background(), nil, GetKnowledgeInput{ID: "solid/single-responsibility"})
	if err != nil {
		t.Fatalf("getKnowledgeHandler error = %v", err)
	}
	if out.Body == "" {
		t.Error("Body is empty, want the concept text")
	}
}

func TestGetKnowledge_ErrorsForUnknownID(t *testing.T) {
	brain := testBrain(t)
	_, _, err := getKnowledgeHandler(brain)(context.Background(), nil, GetKnowledgeInput{ID: "does/not-exist"})
	if err == nil {
		t.Error("getKnowledgeHandler error = nil, want error for unknown id")
	}
}

func TestGetAgentPolicy_ReturnsRealPolicyFileContent(t *testing.T) {
	policyDir := t.TempDir()
	policyPath := filepath.Join(policyDir, "POLICY.md")
	want := "# Agent Collaboration Policy\n\nTest content.\n"
	if err := os.WriteFile(policyPath, []byte(want), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	got, out, err := getAgentPolicyHandler(policyPath)(context.Background(), nil, GetAgentPolicyInput{})
	if err != nil {
		t.Fatalf("getAgentPolicyHandler error = %v", err)
	}
	if got != nil {
		t.Errorf("CallToolResult = %v, want nil (matches list_knowledge/get_knowledge convention)", got)
	}
	if out.Content != want {
		t.Errorf("Content = %q, want %q", out.Content, want)
	}
}

func TestGetAgentPolicy_ErrorsWhenFileMissing(t *testing.T) {
	_, _, err := getAgentPolicyHandler(filepath.Join(t.TempDir(), "does-not-exist.md"))(context.Background(), nil, GetAgentPolicyInput{})
	if err == nil {
		t.Fatal("getAgentPolicyHandler error = nil, want an error for a missing policy file")
	}
}
