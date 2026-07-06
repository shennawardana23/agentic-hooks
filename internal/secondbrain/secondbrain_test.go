package secondbrain

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", full, err)
	}
}

const validConcept = `---
type: principle
title: Single Responsibility Principle
description: Each component has one reason to change.
tags: [solid, architecture]
timestamp: 2026-07-02
---

A component should have one, and only one, reason to change.
`

const missingTypeConcept = `---
title: Broken Concept
---

This file is missing the required type field.
`

func TestLoad_ParsesValidConceptWithPathAsID(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "solid/single-responsibility.md", validConcept)

	brain, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got, err := brain.Get("solid/single-responsibility")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Type != "principle" {
		t.Errorf("Type = %q, want %q", got.Type, "principle")
	}
	if got.Title != "Single Responsibility Principle" {
		t.Errorf("Title = %q, want %q", got.Title, "Single Responsibility Principle")
	}
	if len(got.Tags) != 2 || got.Tags[0] != "solid" {
		t.Errorf("Tags = %v, want [solid architecture]", got.Tags)
	}
	if got.Body == "" {
		t.Error("Body is empty, want the concept text")
	}
}

func TestLoad_SkipsFileMissingRequiredType(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "broken.md", missingTypeConcept)
	writeFixture(t, dir, "solid/single-responsibility.md", validConcept)

	brain, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, err := brain.Get("broken"); err == nil {
		t.Error("Get(\"broken\") = nil error, want error because the file should have been skipped")
	}
	if _, err := brain.Get("solid/single-responsibility"); err != nil {
		t.Errorf("Get() error = %v, want the valid concept to still load", err)
	}
}

func TestBrain_ListFiltersByTag(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "solid/single-responsibility.md", validConcept)
	writeFixture(t, dir, "go/error-handling.md", `---
type: convention
title: Error Handling
tags: [go]
---

Wrap errors with context.
`)

	brain, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := brain.List("", "solid")
	if len(got) != 1 || got[0].ID != "solid/single-responsibility" {
		t.Errorf("List(tag=solid) = %v, want just the SOLID concept", got)
	}
}

func TestBrain_QueryMatchesTitleAndBody(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "solid/single-responsibility.md", validConcept)

	brain, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := brain.Query("reason to change")
	if len(got) != 1 {
		t.Fatalf("Query() returned %d concepts, want 1", len(got))
	}
}
