package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentic-hooks/internal/secondbrain"
)

func testBrainForReview(t *testing.T) *secondbrain.Brain {
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

func TestBuildReviewPrompt_IncludesRelevantConcepts(t *testing.T) {
	brain := testBrainForReview(t)
	diff := "func DoManyThings() { /* violates solid: validates input, writes to disk, sends email */ }"

	prompt := BuildReviewPrompt(diff, brain, nil)

	if !strings.Contains(prompt, "Single Responsibility Principle") {
		t.Errorf("prompt = %q, want it to include the matched concept title", prompt)
	}
	if !strings.Contains(prompt, diff) {
		t.Error("prompt does not include the diff text")
	}
}

func TestBuildReviewPrompt_NoMatchStillIncludesDiff(t *testing.T) {
	brain := testBrainForReview(t)
	diff := "func Add(a, b int) int { return a + b }"

	prompt := BuildReviewPrompt(diff, brain, nil)

	if !strings.Contains(prompt, diff) {
		t.Error("prompt does not include the diff text even with no matched concepts")
	}
}
