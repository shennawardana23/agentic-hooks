package agent

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

// RegistryEntry describes one remote agent this project's root agent can
// delegate to over A2A, loaded from a static YAML file (see LoadRegistry).
type RegistryEntry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	CardURL     string `yaml:"card_url"`
}

// LoadRegistry parses path as a YAML list of RegistryEntry. A missing or
// unparseable file is a fatal error — this is a config-file-not-found/
// malformed problem, not a per-entry validation concern (that's
// BuildAgentTools's job; see agentcomm.go). An empty file yields an empty,
// non-nil-error slice.
func LoadRegistry(path string) ([]RegistryEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("agent: load registry %s: %w", path, err)
	}

	var entries []RegistryEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("agent: parse registry %s: %w", path, err)
	}
	return entries, nil
}
