# Second Brain Orchestration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `agentic-hooks`, a single Go binary with a `run <task>` subcommand (ADK-orchestrated Search + Review agents over a local knowledge base, gated by human approval) and a `serve` subcommand (MCP server exposing that same knowledge base to external coding agents).

**Architecture:** One Go module. `internal/secondbrain` parses a directory of OKF-frontmatter Markdown files into queryable `Concept`s. `internal/mcpserver` wraps that package as MCP tools. `internal/agent` wraps it as an ADK Review sub-agent, paired with a Search sub-agent that calls out to external MCP servers, coordinated by a root `LlmAgent`. `cmd/agentic-hooks` is the cobra CLI wiring both paths.

**Tech Stack:** Go, `google.golang.org/adk/v2` (agent runtime), `github.com/modelcontextprotocol/go-sdk` (MCP client + server), `github.com/spf13/cobra` (CLI), `gopkg.in/yaml.v3` (frontmatter parsing).

## Global Constraints

- Go version: bump `go.mod` to current stable (verify with `go version` on
  the machine executing this plan — do not assume a specific number is
  still current).
- No database. Second Brain is local `.md` files only.
- No Genkit dependency in this plan — Genkit's only designed role (offline
  eval, spec §4.5) is an explicit placeholder, not implemented here. Do not
  add `firebase/genkit` to `go.mod` as part of this plan.
- MCP: `github.com/modelcontextprotocol/go-sdk` only, for both client and
  server roles. Do not add `mark3labs/mcp-go` or any Genkit MCP plugin.
- **No git commits.** The user has a standing instruction for this project:
  do not run `git commit` without their explicit go-ahead in that specific
  moment. Every task below ends with "stage changes, do not commit" instead
  of the usual commit step. If executing this plan autonomously, stop and
  ask before committing anything.
- ADK Go v2's programmatic run/session API (as opposed to its CLI/web
  launcher) was **not found in public docs** as of this plan's writing —
  only `full.NewLauncher()` (a CLI/web launcher) is documented. Task 7
  contains an explicit `go doc` verification step before that task's
  invocation code is trusted. Do not skip it.
- OKF frontmatter fields are exactly: `type` (required), `title`,
  `description`, `resource`, `tags`, `timestamp` (recommended). The
  concept's ID is its file path with `.md` stripped — never add a
  separate `id` field.

---

## File Structure

```
go.mod                              modify — bump Go version, add deps
cmd/agentic-hooks/main.go           create — cobra root command, version subcommand
cmd/agentic-hooks/serve.go          create — `serve` subcommand
cmd/agentic-hooks/run.go            create — `run` subcommand
internal/secondbrain/secondbrain.go create — OKF frontmatter parsing, Brain type
internal/secondbrain/secondbrain_test.go  create
internal/mcpserver/server.go        create — MCP server, list_knowledge/get_knowledge tools
internal/mcpserver/server_test.go   create
internal/agent/search.go            create — Search sub-agent (MCP client toolset)
internal/agent/review.go            create — Review sub-agent (Second Brain + HITL)
internal/agent/review_test.go       create
internal/agent/root.go              create — root coordinator agent
```

---

### Task 1: Project scaffold + CLI skeleton

**Files:**
- Modify: `go.mod`
- Create: `cmd/agentic-hooks/main.go`
- Test: `cmd/agentic-hooks/main_test.go`

**Interfaces:**
- Produces: a `main()` binary with cobra root command `agentic-hooks`,
  subcommands `run`, `serve`, `version` (only `version` implemented in this
  task; `run`/`serve` are registered as empty commands so `--help` lists
  them, real logic added in Tasks 4 and 7).

- [ ] **Step 1: Bump go.mod and add dependencies**

Run:
```bash
go mod edit -go=1.26
go get google.golang.org/adk/v2@latest
go get github.com/modelcontextprotocol/go-sdk@latest
go get github.com/spf13/cobra@latest
go get gopkg.in/yaml.v3@latest
go mod tidy
```
Expected: `go.mod` now declares `go 1.26` (or whatever `go version` reports
as current stable — adjust the `-go=` flag to match) and lists all four
dependencies.

- [ ] **Step 2: Write the failing test**

```go
// cmd/agentic-hooks/main_test.go
package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_HelpListsSubcommands(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"run", "serve", "version"} {
		if !strings.Contains(out, want) {
			t.Errorf("help output missing subcommand %q\noutput:\n%s", want, out)
		}
	}
}

func TestVersionCommand_PrintsVersion(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(buf.String(), "agentic-hooks") {
		t.Errorf("version output = %q, want it to mention agentic-hooks", buf.String())
	}
}
```

- [ ] **Step 2b: Run test to verify it fails**

Run: `go test ./cmd/agentic-hooks/... -run TestRootCommand_HelpListsSubcommands -v`
Expected: FAIL — `undefined: newRootCmd`

- [ ] **Step 3: Write minimal implementation**

```go
// cmd/agentic-hooks/main.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "agentic-hooks",
		Short: "Second Brain orchestration CLI",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the agentic-hooks version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "agentic-hooks dev")
			return err
		},
	}

	runCmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run the Search/Review agent pipeline on a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the Second Brain as an MCP server over stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}

	root.AddCommand(versionCmd, runCmd, serveCmd)
	return root
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/agentic-hooks/... -v`
Expected: PASS (both `TestRootCommand_HelpListsSubcommands` and
`TestVersionCommand_PrintsVersion`)

- [ ] **Step 5: Stage changes (do not commit)**

```bash
git add go.mod go.sum cmd/agentic-hooks/main.go cmd/agentic-hooks/main_test.go
```
Do not run `git commit` — see Global Constraints.

---

### Task 2: Second Brain — OKF frontmatter parsing and query

**Files:**
- Create: `internal/secondbrain/secondbrain.go`
- Test: `internal/secondbrain/secondbrain_test.go`

**Interfaces:**
- Produces:
  - `type Concept struct { ID, Type, Title, Description, Resource, Timestamp string; Tags []string; Body string }`
  - `func Load(dir string) (*Brain, error)`
  - `func (b *Brain) List(typeFilter string, tagFilter string) []Concept`
  - `func (b *Brain) Get(id string) (Concept, error)`
  - `func (b *Brain) Query(topic string) []Concept`

- [ ] **Step 1: Write the failing tests**

```go
// internal/secondbrain/secondbrain_test.go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/secondbrain/... -v`
Expected: FAIL — build error, `undefined: Load` (package doesn't exist yet)

- [ ] **Step 3: Write minimal implementation**

```go
// internal/secondbrain/secondbrain.go
package secondbrain

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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
			return nil
		}
		concepts = append(concepts, concept)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("secondbrain: load %s: %w", dir, err)
	}

	return &Brain{concepts: concepts}, nil
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/secondbrain/... -v`
Expected: PASS on all four tests.

- [ ] **Step 5: Stage changes (do not commit)**

```bash
git add internal/secondbrain/secondbrain.go internal/secondbrain/secondbrain_test.go
```

---

### Task 3: MCP server exposing the Second Brain

**Files:**
- Create: `internal/mcpserver/server.go`
- Test: `internal/mcpserver/server_test.go`

**Interfaces:**
- Consumes: `secondbrain.Brain` (`List`, `Get`) from Task 2.
- Produces: `func NewServer(brain *secondbrain.Brain) *mcp.Server` — a
  configured MCP server with `list_knowledge` and `get_knowledge` tools
  registered, ready for `server.Run(ctx, transport)`.

- [ ] **Step 1: Write the failing tests**

Test the tool handler functions directly (plain Go function calls) rather
than through the MCP wire transport — this exercises the actual business
logic without depending on unverified transport testing helpers.

```go
// internal/mcpserver/server_test.go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/mcpserver/... -v`
Expected: FAIL — `undefined: listKnowledgeHandler` (package doesn't exist yet)

- [ ] **Step 3: Write minimal implementation**

```go
// internal/mcpserver/server.go
package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"agentic-hooks/internal/secondbrain"
)

type ListKnowledgeInput struct {
	Type string `json:"type,omitempty" jsonschema:"filter by concept type, e.g. principle"`
	Tag  string `json:"tag,omitempty" jsonschema:"filter by tag, e.g. solid"`
}

type ConceptSummary struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type ListKnowledgeOutput struct {
	Concepts []ConceptSummary `json:"concepts"`
}

type GetKnowledgeInput struct {
	ID string `json:"id" jsonschema:"the concept id, i.e. its file path without .md"`
}

type GetKnowledgeOutput struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func listKnowledgeHandler(brain *secondbrain.Brain) func(context.Context, *mcp.CallToolRequest, ListKnowledgeInput) (*mcp.CallToolResult, ListKnowledgeOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListKnowledgeInput) (*mcp.CallToolResult, ListKnowledgeOutput, error) {
		concepts := brain.List(input.Type, input.Tag)
		out := ListKnowledgeOutput{Concepts: make([]ConceptSummary, 0, len(concepts))}
		for _, c := range concepts {
			out.Concepts = append(out.Concepts, ConceptSummary{
				ID: c.ID, Type: c.Type, Title: c.Title, Description: c.Description, Tags: c.Tags,
			})
		}
		return nil, out, nil
	}
}

func getKnowledgeHandler(brain *secondbrain.Brain) func(context.Context, *mcp.CallToolRequest, GetKnowledgeInput) (*mcp.CallToolResult, GetKnowledgeOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetKnowledgeInput) (*mcp.CallToolResult, GetKnowledgeOutput, error) {
		concept, err := brain.Get(input.ID)
		if err != nil {
			return nil, GetKnowledgeOutput{}, fmt.Errorf("mcpserver: %w", err)
		}
		return nil, GetKnowledgeOutput{ID: concept.ID, Title: concept.Title, Body: concept.Body}, nil
	}
}

func NewServer(brain *secondbrain.Brain) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "agentic-hooks-secondbrain", Version: "v0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_knowledge",
		Description: "List Second Brain concepts, optionally filtered by type or tag",
	}, listKnowledgeHandler(brain))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_knowledge",
		Description: "Get the full content of a Second Brain concept by id",
	}, getKnowledgeHandler(brain))

	return server
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/mcpserver/... -v`
Expected: PASS on all three tests.

- [ ] **Step 5: Stage changes (do not commit)**

```bash
git add internal/mcpserver/server.go internal/mcpserver/server_test.go
```

---

### Task 4: CLI `serve` subcommand

**Files:**
- Modify: `cmd/agentic-hooks/main.go` (replace the `serveCmd` stub from Task 1)
- Create: `cmd/agentic-hooks/serve.go`

**Interfaces:**
- Consumes: `secondbrain.Load` (Task 2), `mcpserver.NewServer` (Task 3).
- Produces: `func newServeCmd() *cobra.Command`, wired into the root command.

- [ ] **Step 1: Write the failing test**

```go
// cmd/agentic-hooks/serve_test.go
package main

import "testing"

func TestServeCmd_RequiresKnowledgeDirFlag(t *testing.T) {
	cmd := newServeCmd()
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("Execute() error = nil, want error because --knowledge-dir is required")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/agentic-hooks/... -run TestServeCmd_RequiresKnowledgeDirFlag -v`
Expected: FAIL — `undefined: newServeCmd`

- [ ] **Step 3: Write minimal implementation**

```go
// cmd/agentic-hooks/serve.go
package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"agentic-hooks/internal/mcpserver"
	"agentic-hooks/internal/secondbrain"
)

func newServeCmd() *cobra.Command {
	var knowledgeDir string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the Second Brain as an MCP server over stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			brain, err := secondbrain.Load(knowledgeDir)
			if err != nil {
				return err
			}
			server := mcpserver.NewServer(brain)
			return server.Run(context.Background(), &mcp.StdioTransport{})
		},
	}

	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", "", "path to the Second Brain knowledge directory (required)")
	cmd.MarkFlagRequired("knowledge-dir")

	return cmd
}
```

Then in `cmd/agentic-hooks/main.go`, replace the inline `serveCmd` stub:

```go
	root.AddCommand(versionCmd, runCmd, newServeCmd())
```

(remove the old `serveCmd := &cobra.Command{...}` block and its reference
from the `AddCommand` call)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/agentic-hooks/... -v`
Expected: PASS on all tests (Task 1's tests plus the new one).

- [ ] **Step 5: Stage changes (do not commit)**

```bash
git add cmd/agentic-hooks/main.go cmd/agentic-hooks/serve.go cmd/agentic-hooks/serve_test.go
```

---

### Task 5: Search sub-agent

**Files:**
- Create: `internal/agent/search.go`

**Interfaces:**
- Consumes: `mcptoolset.New` from `google.golang.org/adk/v2/tool/mcptoolset`.
- Produces: `func NewSearchAgent(mcpCommand string, mcpArgs []string, model model.LLM) (agent.Agent, error)`

**Note:** this task has no dedicated unit test — `LlmAgent` construction is
config-object assembly with no branching logic of our own to test in
isolation, and exercising it live requires a real model + MCP server
(covered by the Task 7 manual smoke test instead). If `llmagent.New`
returns a validation error for a config field this plan gets wrong,
Step 2 below will surface it immediately.

- [ ] **Step 1: Write the implementation**

```go
// internal/agent/search.go
package agent

import (
	"os/exec"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/mcptoolset"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewSearchAgent(mcpCommand string, mcpArgs []string, m model.LLM) (agent.Agent, error) {
	toolset, err := mcptoolset.New(mcptoolset.Config{
		Transport: &mcp.CommandTransport{Command: exec.Command(mcpCommand, mcpArgs...)},
	})
	if err != nil {
		return nil, err
	}

	return llmagent.New(llmagent.Config{
		Name:        "search",
		Model:       m,
		Description: "Looks up external information via a configured MCP tool server.",
		Instruction: "Use the available tools to find information relevant to the task. Summarize findings concisely.",
		Toolsets:    []tool.Set{toolset},
	})
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/agent/...`
Expected: builds cleanly. If `llmagent.Config`, `mcptoolset.Config`, or
`tool.Set` field names differ from what's above, this will fail with a
clear compiler error naming the actual field — fix the struct literal to
match and re-run.

- [ ] **Step 3: Stage changes (do not commit)**

```bash
git add internal/agent/search.go
```

---

### Task 6: Review sub-agent with HITL confirmation

**Files:**
- Create: `internal/agent/review.go`
- Test: `internal/agent/review_test.go`

**Interfaces:**
- Consumes: `secondbrain.Brain.Query` (Task 2).
- Produces:
  - `type StructuralFacts struct{}` (empty marker type for the deferred
    tree-sitter seam — see spec §4.2/§7)
  - `func BuildReviewPrompt(diff string, brain *secondbrain.Brain, facts *StructuralFacts) string`
    — the part of Review's logic that's pure and unit-testable without a
    live model or ADK runtime: it gathers the relevant Second Brain
    concepts for the diff and assembles the prompt text.
  - `func NewReviewAgent(m model.LLM, brain *secondbrain.Brain) (agent.Agent, error)`
    — the ADK-facing wrapper, covered by the Task 7 manual smoke test
    rather than a unit test, same reasoning as Task 5.

- [ ] **Step 1: Write the failing test**

```go
// internal/agent/review_test.go
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
	diff := "func DoManyThings() { /* validates input, writes to disk, sends email, all in one reason to change */ }"

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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/agent/... -run TestBuildReviewPrompt -v`
Expected: FAIL — `undefined: BuildReviewPrompt`

- [ ] **Step 3: Write minimal implementation**

```go
// internal/agent/review.go
package agent

import (
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"

	"agentic-hooks/internal/secondbrain"
)

// StructuralFacts is a reserved seam for a future tree-sitter pre-filter
// stage (deferred — see design spec §7). It carries no fields yet.
type StructuralFacts struct{}

func BuildReviewPrompt(diff string, brain *secondbrain.Brain, facts *StructuralFacts) string {
	var sb strings.Builder
	sb.WriteString("Review the following code change against these Second Brain principles:\n\n")

	matched := brain.Query(diff)
	if len(matched) == 0 {
		sb.WriteString("(no specific Second Brain concept matched this diff by keyword — review using general SOLID and clean-code judgment)\n\n")
	}
	for _, c := range matched {
		fmt.Fprintf(&sb, "- %s: %s\n", c.Title, c.Body)
	}

	sb.WriteString("\nCode change:\n")
	sb.WriteString(diff)
	sb.WriteString("\n\nProduce a verdict: APPROVE or CHANGES_REQUESTED, with reasoning tied to the principles above.")

	return sb.String()
}

func NewReviewAgent(m model.LLM, brain *secondbrain.Brain) (agent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "review",
		Model:       m,
		Description: "Reviews a code diff against the Second Brain's coding principles.",
		Instruction: "You review code changes for adherence to the provided principles. Always end with an explicit APPROVE or CHANGES_REQUESTED verdict.",
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/... -v`
Expected: PASS on both `TestBuildReviewPrompt_*` tests.

- [ ] **Step 5: Stage changes (do not commit)**

```bash
git add internal/agent/review.go internal/agent/review_test.go
```

---

### Task 7: Root agent, HITL wiring, and `run` subcommand

**This task starts with a mandatory verification step — do not skip it.**
Prior research for this plan found ADK Go v2's *programmatic* run/session
API undocumented publicly; only the CLI/web launcher (`full.NewLauncher()`)
was confirmed. The code below is written against the field names/methods
that *are* confirmed (`llmagent.Config.SubAgents`, `ctx.RequestConfirmation`,
`ctx.ToolConfirmation`) plus a best-effort invocation sketch for the parts
that are not. Treat anything marked `// UNVERIFIED` as a hypothesis to
check against the real installed module, not as fact.

**Files:**
- Create: `internal/agent/root.go`
- Modify: `cmd/agentic-hooks/main.go` (replace the `runCmd` stub from Task 1)
- Create: `cmd/agentic-hooks/run.go`

**Interfaces:**
- Consumes: `NewSearchAgent` (Task 5), `NewReviewAgent` (Task 6).
- Produces: `func NewRootAgent(search, review agent.Agent, m model.LLM) (agent.Agent, error)`,
  `func newRunCmd() *cobra.Command`.

- [ ] **Step 1: Verify the programmatic invocation API before writing run.go**

Run:
```bash
go doc google.golang.org/adk/v2
go doc google.golang.org/adk/v2/agent
```
Read the output for a type that runs an `agent.Agent` and returns a result
synchronously (likely named something like `Runner`, `Session`, or
similar). Compare it against the `// UNVERIFIED` block in Step 3 below.
If the real API differs, rewrite that block to match what `go doc` shows
— do not proceed with the unverified guess if a real API is found.

- [ ] **Step 2: Write root.go using confirmed APIs**

```go
// internal/agent/root.go
package agent

import (
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
)

func NewRootAgent(search, review agent.Agent, m model.LLM) (agent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "root",
		Model:       m,
		Description: "Coordinates Search and Review sub-agents to handle a task.",
		Instruction: "Delegate lookups to the search agent and code review to the review agent as needed. Always surface the review agent's final verdict to the user.",
		SubAgents:   []agent.Agent{search, review},
	})
}
```

- [ ] **Step 3: Write run.go — invocation code, with the unverified part flagged**

```go
// cmd/agentic-hooks/run.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agentic-hooks/internal/agent"
	"agentic-hooks/internal/secondbrain"
)

func newRunCmd() *cobra.Command {
	var knowledgeDir string
	var mcpCommand string

	cmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run the Search/Review agent pipeline on a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task := args[0]

			brain, err := secondbrain.Load(knowledgeDir)
			if err != nil {
				return err
			}

			m, err := newDefaultModel() // UNVERIFIED: confirm the real model
			// constructor (e.g. a Gemini client) via `go doc
			// google.golang.org/adk/v2/model` before trusting this call.
			if err != nil {
				return err
			}

			search, err := agent.NewSearchAgent(mcpCommand, nil, m)
			if err != nil {
				return err
			}
			review, err := agent.NewReviewAgent(m, brain)
			if err != nil {
				return err
			}
			root, err := agent.NewRootAgent(search, review, m)
			if err != nil {
				return err
			}

			// UNVERIFIED: the real invocation likely looks like creating a
			// Runner/Session against `root` and calling something like
			// runner.Run(ctx, sessionID, task) to get a response, per Step 1's
			// `go doc` output. Replace this block once confirmed:
			result, err := runRootAgent(context.Background(), root, task)
			if err != nil {
				return err
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Fprintf(cmd.OutOrStdout(), "Verdict:\n%s\n\nApprove? [y/N]: ", result)
			line, _ := reader.ReadString('\n')
			if line != "y\n" && line != "Y\n" {
				fmt.Fprintln(cmd.OutOrStdout(), "Rejected — no output returned as final.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), result)
			return nil
		},
	}

	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", "", "path to the Second Brain knowledge directory (required)")
	cmd.Flags().StringVar(&mcpCommand, "search-mcp-server", "", "command to launch the Search agent's MCP server (required)")
	cmd.MarkFlagRequired("knowledge-dir")
	cmd.MarkFlagRequired("search-mcp-server")

	return cmd
}
```

Then in `cmd/agentic-hooks/main.go`, replace the inline `runCmd` stub:

```go
	root.AddCommand(newRunCmd(), newServeCmd())
```

(remove the old `runCmd := &cobra.Command{...}` block, the old `serveCmd`
reference if not already removed in Task 4, and the stale `AddCommand`
call referencing them)

- [ ] **Step 4: Build and fix against the real API**

Run: `go build ./...`
Expected: this WILL fail on `newDefaultModel` and `runRootAgent` — those
are intentionally left undefined, standing in for the API confirmed in
Step 1. Replace both with the real constructor/invocation calls found via
`go doc`, then re-run `go build ./...` until it passes.

- [ ] **Step 5: Manual smoke test (not part of the fast unit suite)**

```bash
mkdir -p /tmp/agentic-hooks-knowledge/solid
cat > /tmp/agentic-hooks-knowledge/solid/single-responsibility.md <<'EOF'
---
type: principle
title: Single Responsibility Principle
tags: [solid]
---

A component should have one reason to change.
EOF

go run ./cmd/agentic-hooks run "review: func DoEverything() { /* ... */ }" \
  --knowledge-dir /tmp/agentic-hooks-knowledge \
  --search-mcp-server /path/to/some/mcp/server
```
Expected: a verdict prints, followed by an approve/reject prompt. Confirm
both the approve and reject paths behave as described in spec §5.

- [ ] **Step 6: Stage changes (do not commit)**

```bash
git add internal/agent/root.go cmd/agentic-hooks/main.go cmd/agentic-hooks/run.go
```

---

## Self-Review Notes

- **Spec coverage**: §3 architecture (Task 1, 5-7), §4.1 CLI (Task 1, 4, 7),
  §4.2 ADK runtime incl. HITL (Task 5-7), §4.3 Second Brain (Task 2), §4.4
  MCP server (Task 3), §5 data flow (Task 4, 7), §6 error handling (Task 2's
  skip-on-parse-failure, Task 3's typed error on unknown id) are all
  covered. §4.5 (offline eval) and §7 (deferred items) are intentionally
  NOT covered — out of scope for this plan per the spec.
- **Type consistency checked**: `secondbrain.Concept`, `*secondbrain.Brain`,
  `agent.Agent`, `model.LLM` names are used identically across Tasks 2-7.
- **Known gap, called out rather than hidden**: Task 7's model constructor
  and root-agent invocation are explicitly unverified and gated behind a
  `go doc` check — this is a real open risk in the plan, not an oversight.

---

**Plan complete and saved to `docs/superpowers/plans/2026-07-02-second-brain-orchestration.md`.**
