package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func writeRegistryFixture(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "agents.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
	return path
}

const validRegistry = `
- name: example-agent
  description: A remote example agent.
  card_url: http://localhost:9003
- name: another-agent
  description: Another remote agent.
  card_url: http://localhost:9004
`

const malformedRegistry = `
- name: [this is not a valid scalar for name
`

func TestLoadRegistry_ParsesValidEntries(t *testing.T) {
	dir := t.TempDir()
	path := writeRegistryFixture(t, dir, validRegistry)

	entries, err := LoadRegistry(path)
	if err != nil {
		t.Fatalf("LoadRegistry() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Name != "example-agent" {
		t.Errorf("entries[0].Name = %q, want %q", entries[0].Name, "example-agent")
	}
	if entries[0].Description != "A remote example agent." {
		t.Errorf("entries[0].Description = %q, want %q", entries[0].Description, "A remote example agent.")
	}
	if entries[0].CardURL != "http://localhost:9003" {
		t.Errorf("entries[0].CardURL = %q, want %q", entries[0].CardURL, "http://localhost:9003")
	}
	if entries[1].Name != "another-agent" {
		t.Errorf("entries[1].Name = %q, want %q", entries[1].Name, "another-agent")
	}
}

func TestLoadRegistry_MissingFileReturnsError(t *testing.T) {
	_, err := LoadRegistry(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatal("LoadRegistry() error = nil, want non-nil for a missing file")
	}
}

func TestLoadRegistry_MalformedYAMLReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := writeRegistryFixture(t, dir, malformedRegistry)

	_, err := LoadRegistry(path)
	if err == nil {
		t.Fatal("LoadRegistry() error = nil, want non-nil for malformed YAML")
	}
}

func TestLoadRegistry_EmptyFileReturnsEmptySlice(t *testing.T) {
	dir := t.TempDir()
	path := writeRegistryFixture(t, dir, "")

	entries, err := LoadRegistry(path)
	if err != nil {
		t.Fatalf("LoadRegistry() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}
