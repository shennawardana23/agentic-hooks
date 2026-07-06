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
