package secondbrain

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Concept struct {
	ID          string
	Type        string
	Title       string
	Description string
	Resource    string
	Tags        []string
	Timestamp   string
	Body        string
}

type Brain struct {
	concepts []Concept
	skipped  []string
}

type frontmatter struct {
	Type        string   `yaml:"type"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Resource    string   `yaml:"resource"`
	Tags        []string `yaml:"tags"`
	Timestamp   string   `yaml:"timestamp"`
}

func parseConcept(id, content string) (Concept, error) {
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) < 3 {
		return Concept{}, fmt.Errorf("secondbrain: %s: missing --- frontmatter delimiters", id)
	}

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return Concept{}, fmt.Errorf("secondbrain: %s: invalid frontmatter: %w", id, err)
	}
	if fm.Type == "" {
		return Concept{}, fmt.Errorf("secondbrain: %s: missing required field 'type'", id)
	}

	return Concept{
		ID:          id,
		Type:        fm.Type,
		Title:       fm.Title,
		Description: fm.Description,
		Resource:    fm.Resource,
		Tags:        fm.Tags,
		Timestamp:   fm.Timestamp,
		Body:        strings.TrimSpace(parts[2]),
	}, nil
}

func Load(dir string) (*Brain, error) {
	var concepts []Concept
	var skipped []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		id := strings.TrimSuffix(filepath.ToSlash(rel), ".md")

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		concept, err := parseConcept(id, string(data))
		if err != nil {
			log.Printf("secondbrain: skipping %s: %v", path, err)
			skipped = append(skipped, fmt.Sprintf("%s: %v", rel, err))
			return nil
		}
		concepts = append(concepts, concept)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("secondbrain: load %s: %w", dir, err)
	}

	return &Brain{concepts: concepts, skipped: skipped}, nil
}

// SkippedFiles returns one "path: reason" string per concept file Load
// couldn't parse. Callers that expose the Brain over a network boundary
// (e.g. the MCP server's list_knowledge tool) should surface these —
// log.Printf alone is invisible to a remote caller.
func (b *Brain) SkippedFiles() []string {
	return b.skipped
}

func (b *Brain) List(typeFilter, tagFilter string) []Concept {
	var out []Concept
	for _, c := range b.concepts {
		if typeFilter != "" && c.Type != typeFilter {
			continue
		}
		if tagFilter != "" && !containsTag(c.Tags, tagFilter) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (b *Brain) Get(id string) (Concept, error) {
	for _, c := range b.concepts {
		if c.ID == id {
			return c, nil
		}
	}
	return Concept{}, fmt.Errorf("secondbrain: no concept with id %q", id)
}

func (b *Brain) Query(topic string) []Concept {
	topic = strings.ToLower(topic)
	var out []Concept
	for _, c := range b.concepts {
		haystack := strings.ToLower(c.Title + " " + c.Description + " " + c.Body + " " + strings.Join(c.Tags, " "))
		if strings.Contains(haystack, topic) {
			out = append(out, c)
		}
	}
	return out
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
