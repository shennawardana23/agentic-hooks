# Agent-to-Agent Communication Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let the root agent delegate to external agents over the real A2A
protocol, driven by a static YAML registry, using ADK's own
`remoteagent/v2` + `agenttool` primitives — additive only, zero behavior
change when `--agents-config` is omitted.

**Architecture:** `internal/agent/registry.go` loads a YAML list of
`{name, description, card_url}` entries. `internal/agent/agentcomm.go`
turns each entry into a `tool.Tool` via
`agenttool.New(remoteagent.NewA2A(...), nil)`. `NewRootAgent` gains a 4th
parameter, `agentTools []tool.Tool`, merged into `llmagent.Config.Tools`.
`run.go` gains an optional `--agents-config` flag; omitted, the whole path
is a no-op and today's exact behavior is reproduced.

**Tech Stack:** Go, `google.golang.org/adk/v2@v2.0.0`
(`agent/remoteagent/v2`, `tool/agenttool`), `github.com/a2aproject/a2a-go/v2@v2.3.1`
(pulled in transitively — pin via `go get .../a2a-go/v2@v2.3.1` +
`go mod tidy` in Task 2), `go.yaml.in/yaml/v3` (already a dependency).

## Global Constraints

- Every task ends with `go build ./...`, `go vet ./...`, `go test ./...`
  all clean before moving to the next task.
- No `git commit` unless the user explicitly asks in a given message
  (standing repo-wide instruction — see `CLAUDE.md`/`POLICY.md`).
- Additive only: `NewRootAgent`'s and `newRunCmd`'s existing behavior with
  no `agentTools`/no `--agents-config` must be byte-for-byte the same as
  today. Every existing test must still pass unmodified except where a
  call site's signature literally changed (root_test.go).
- One-concept-per-file convention already established in
  `internal/agent/` (`search.go`, `review.go`, `root.go`, `loop.go`) — new
  code goes in new files `registry.go` and `agentcomm.go`, not appended to
  existing files.
- Per-entry registry failures are warnings, never fatal; only a missing or
  unparseable registry *file* is fatal (matches
  `docs/superpowers/specs/2026-07-06-agent-to-agent-communication-design.md`'s
  Error Handling section).

**Deviation from the approved design spec, noted up front:** the spec's
sketch of `BuildAgentTools` includes a trailing `err error` return. Nothing
in the spec's own Error Handling section ever produces a non-nil value for
it (per-entry failures are warnings; there is no registry-wide construction
failure once `LoadRegistry` has already succeeded). Task 4 below drops that
return value — `func BuildAgentTools(entries []RegistryEntry) (tools []tool.Tool, warnings []string)`
— to avoid an always-nil `error` a reviewer would flag as dead. This is the
only knowing deviation from the spec in this plan; everything else matches
it exactly, including the parts of the spec that were re-verified and
found to need correction (see Task 4's well-known-path note).

---

### Task 1: Document the reopened architecture decision

**Files:**
- Create: `docs/adr/0010-network-a2a-via-adk-remoteagent.md`
- Modify: `docs/adr/0001-adk-go-v2-as-sole-orchestrator.md` (Consequences section)
- Modify: `MEMORY.md` (Decisions already locked)

**Interfaces:** none — documentation only, no code dependency for later tasks.

- [ ] **Step 1: Write the new ADR**

Create `docs/adr/0010-network-a2a-via-adk-remoteagent.md`:

```markdown
# ADR-0010: Network A2A via ADK's `remoteagent/v2` primitives

## Status
Accepted

## Context
ADR-0001 recorded "no network-based A2A this iteration" as a consequence
of choosing ADK Go v2 as sole orchestrator — at the time, in-process
delegation covered every sub-agent this project had (Search, Generator,
Review). The user has since asked for "Intelligent Communication Through
Tool Calling": the root agent should be able to see other agents' public
manifests and decide, at runtime, whether to delegate to them — instead of
a fixed orchestration graph. This is a deliberate, user-confirmed
reopening of that consequence, not a silent scope change (see
`SESSION_HANDOFF.md`'s 2026-07-06 "Answered" entry and
`docs/superpowers/specs/2026-07-06-agent-to-agent-communication-design.md`
for the full design).

ADK Go v2 (this project's exact pinned dependency, `v2.0.0`) already ships
non-deprecated network A2A primitives: `agent/remoteagent/v2.NewA2A` wraps
a real A2A-protocol remote agent (via `github.com/a2aproject/a2a-go/v2`) as
a normal `agent.Agent`; `tool/agenttool.New` wraps any `agent.Agent` as a
callable `tool.Tool`. No hand-rolled A2A client is needed.

## Decision
Add network A2A support as an **additive, opt-in** capability: a static
YAML registry (`--agents-config`, no default) is loaded into
`agenttool`-wrapped `remoteagent.NewA2A` tools and merged into the root
agent's `Tools`. In-process delegation (Search, Generator/Review loop via
`SubAgents`) is completely untouched — this is a second, independent
delegation mechanism (tool-calling to remote agents), not a replacement.

## Consequences
- ADR-0001's "no network-based A2A this iteration" consequence is
  superseded by this ADR for the tool-calling delegation path specifically.
  In-process `SubAgents` delegation (Search, loop) is unaffected and remains
  the primary orchestration mechanism.
- A new dependency, `github.com/a2aproject/a2a-go/v2`, enters `go.mod`
  (already an indirect dependency of `google.golang.org/adk/v2`; this pins
  it as direct once `internal/agent/agentcomm.go` imports `remoteagent/v2`).
- No discovery-of-unknown-agents mechanism exists or is added — the
  registry is a static, explicitly-configured YAML file. Wiring a specific
  real remote agent (e.g. a sibling repo's A2A server) in as a live
  registry entry is out of scope for this decision; see the design spec's
  "Explicitly out of scope" section.
```

- [ ] **Step 2: Cross-reference from ADR-0001**

In `docs/adr/0001-adk-go-v2-as-sole-orchestrator.md`, find this line in the
Consequences section:

```markdown
- Sub-agent delegation (Search, Generator, Review) is in-process ADK
  delegation only — no network-based A2A this iteration.
```

Replace it with:

```markdown
- Sub-agent delegation (Search, Generator, Review) is in-process ADK
  delegation only — no network-based A2A this iteration. **Superseded
  2026-07-06** for the tool-calling delegation path specifically: see
  [ADR-0010](0010-network-a2a-via-adk-remoteagent.md). In-process
  `SubAgents` delegation remains unchanged and is still the primary
  orchestration mechanism.
```

- [ ] **Step 3: Update MEMORY.md's locked-decisions list**

In `MEMORY.md`, find the ADR-0001 bullet:

```markdown
- **Runtime**: ADK Go v2 (`google.golang.org/adk/v2`) is the sole
  request-path orchestrator. Genkit is never added to the request path.
  ([docs/adr/0001](docs/adr/0001-adk-go-v2-as-sole-orchestrator.md))
```

Add a new bullet immediately after it:

```markdown
- **Network A2A (opt-in, additive)**: the root agent can delegate to
  external agents over the real A2A protocol via `agent/remoteagent/v2` +
  `tool/agenttool`, driven by a static YAML registry
  (`--agents-config`, no default — omitted means zero behavior change).
  This supersedes ADR-0001's original "no network A2A" consequence for
  the tool-calling path only; in-process `SubAgents` delegation is
  unaffected.
  ([docs/adr/0010](docs/adr/0010-network-a2a-via-adk-remoteagent.md))
```

- [ ] **Step 4: Verify no other doc references the stale claim**

Run: `grep -rn "no network.A2A\|no network-based A2A" --include='*.md' .`
Expected: only the now-annotated line in `docs/adr/0001-...md` (with its
new "Superseded" sentence) and this plan file itself. If any other doc
(e.g. `docs/explanation/architecture-overview.md`) asserts "no network
A2A" as current fact, note it here — do not edit files not in this task's
Files list without confirming the doc's scope first.

- [ ] **Step 5: Commit**

```bash
git add docs/adr/0010-network-a2a-via-adk-remoteagent.md docs/adr/0001-adk-go-v2-as-sole-orchestrator.md MEMORY.md
git commit -m "docs: add ADR-0010 superseding ADR-0001's no-network-A2A consequence"
```

(Only run this commit if the user has asked for commits this session —
otherwise leave staged/modified per the standing no-commit instruction.)

---

### Task 2: `RegistryEntry` + `LoadRegistry`

**Files:**
- Create: `internal/agent/registry.go`
- Create: `internal/agent/registry_test.go`

**Interfaces:**
- Produces: `type RegistryEntry struct { Name, Description, CardURL string }`
  (yaml-tagged `name`/`description`/`card_url`); `func LoadRegistry(path string) ([]RegistryEntry, error)`.
  Task 4 (`agentcomm.go`) and Task 6 (`run.go`) both consume these exact
  names/types.

- [ ] **Step 1: Write the failing tests**

Create `internal/agent/registry_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/agent/... -run TestLoadRegistry -v`
Expected: FAIL — `LoadRegistry` (and `RegistryEntry`) undefined.

- [ ] **Step 3: Write the implementation**

Create `internal/agent/registry.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/... -run TestLoadRegistry -v`
Expected: PASS (all 4 subtests).

- [ ] **Step 5: Commit**

```bash
git add internal/agent/registry.go internal/agent/registry_test.go
git commit -m "feat: add YAML registry loader for remote agent config"
```

---

### Task 3: Pin `a2a-go/v2` as a direct dependency

**Files:**
- Modify: `go.mod`, `go.sum`

**Interfaces:** none — this task only makes the module available for
Task 4 to import. No Go source changes.

- [ ] **Step 1: Fetch the pinned version**

Run: `go get github.com/a2aproject/a2a-go/v2@v2.3.1`
Expected output includes a line like:
`go: added github.com/a2aproject/a2a-go/v2 v2.3.1`
(it may also report a toolchain bump, e.g. `go1.21.13 => go1.25.0` — accept
it; `go.mod` already declares `go 1.25.0` per `SESSION_HANDOFF.md`, re-verify
with `go version` if this looks stale).

- [ ] **Step 2: Verify the module resolves cleanly**

Run: `go build ./...`
Expected: clean build, no new compile errors (nothing imports the new
module yet, so this just proves `go.mod`/`go.sum` are consistent).

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: pin github.com/a2aproject/a2a-go/v2 v2.3.1"
```

---

### Task 4: `BuildAgentTools`

**Files:**
- Create: `internal/agent/agentcomm.go`
- Create: `internal/agent/agentcomm_test.go`

**Interfaces:**
- Consumes: `RegistryEntry{Name, Description, CardURL}` from Task 2.
- Produces: `func BuildAgentTools(entries []RegistryEntry) (tools []tool.Tool, warnings []string)`.
  Task 5 (`root.go`) and Task 6 (`run.go`) both consume this exact name,
  signature (no `error` return — see the Global Constraints deviation
  note), and the fact that `tools`/`warnings` can each independently be
  nil/empty.

**Verified facts this task depends on** (re-verified against
`google.golang.org/adk/v2@v2.0.0` and `github.com/a2aproject/a2a-go/v2@v2.3.1`
source directly, not assumed from the design spec — the spec had one wrong
detail, corrected here):

- `remoteagent.NewAgentCardProvider(source string, ...)` (package
  `google.golang.org/adk/v2/agent/remoteagent/v2`) treats `source` as a
  **bare base URL**. It internally joins `/.well-known/agent-card.json`
  (`agentcard.DefaultResolver`, in
  `github.com/a2aproject/a2a-go/v2@v2.3.1/a2aclient/agentcard/resolver.go`).
  **The design spec's draft path, `/.well-known/agent.json`, is wrong** —
  do not use it anywhere (test fixtures included, Task 4 Step 1 below uses
  the corrected path).
- No `Validate()` call exists anywhere between `NewAgentCardProvider` and
  `NewA2A` — the only real constraint is enforced later, at first
  invocation, by `a2aclient.Factory.CreateFromCard`: `SupportedInterfaces`
  must be non-empty and contain a transport the client supports
  (JSON-RPC by default).
- Card/tool construction is pure Go object construction — no network call
  happens until the tool is actually invoked (`AgentCardProvider` defers
  the fetch). This is why `agentcomm_test.go`'s unit-level tests (Step 1)
  need no real server.

- [ ] **Step 1: Write the failing unit tests (no network)**

Create `internal/agent/agentcomm_test.go`:

```go
package agent

import (
	"testing"
)

func TestBuildAgentTools_WrapsValidEntriesAsTools(t *testing.T) {
	entries := []RegistryEntry{
		{Name: "example-agent", Description: "An example.", CardURL: "http://localhost:9003"},
	}

	tools, warnings := BuildAgentTools(entries)

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1", len(tools))
	}
	if got := tools[0].Name(); got != "example-agent" {
		t.Errorf("tools[0].Name() = %q, want %q", got, "example-agent")
	}
}

func TestBuildAgentTools_SkipsEmptyNameWithWarning(t *testing.T) {
	entries := []RegistryEntry{
		{Name: "", Description: "no name", CardURL: "http://localhost:9003"},
		{Name: "valid-agent", Description: "fine", CardURL: "http://localhost:9004"},
	}

	tools, warnings := BuildAgentTools(entries)

	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1 (only the valid entry)", len(tools))
	}
	if len(warnings) != 1 {
		t.Fatalf("len(warnings) = %d, want 1", len(warnings))
	}
}

func TestBuildAgentTools_SkipsEmptyCardURLWithWarning(t *testing.T) {
	entries := []RegistryEntry{
		{Name: "no-url-agent", Description: "missing url", CardURL: ""},
	}

	tools, warnings := BuildAgentTools(entries)

	if len(tools) != 0 {
		t.Errorf("len(tools) = %d, want 0", len(tools))
	}
	if len(warnings) != 1 {
		t.Fatalf("len(warnings) = %d, want 1", len(warnings))
	}
}

func TestBuildAgentTools_SkipsUnparseableCardURLWithWarning(t *testing.T) {
	entries := []RegistryEntry{
		{Name: "bad-url-agent", Description: "bad url", CardURL: "http://[::1]:namedport"},
	}

	tools, warnings := BuildAgentTools(entries)

	if len(tools) != 0 {
		t.Errorf("len(tools) = %d, want 0", len(tools))
	}
	if len(warnings) != 1 {
		t.Fatalf("len(warnings) = %d, want 1", len(warnings))
	}
}

func TestBuildAgentTools_EmptyRegistryReturnsEmptyNoWarnings(t *testing.T) {
	tools, warnings := BuildAgentTools(nil)
	if len(tools) != 0 || len(warnings) != 0 {
		t.Errorf("BuildAgentTools(nil) = (%v, %v), want (empty, empty)", tools, warnings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/agent/... -run TestBuildAgentTools -v`
Expected: FAIL — `BuildAgentTools` undefined.

- [ ] **Step 3: Write the implementation**

Create `internal/agent/agentcomm.go`:

```go
package agent

import (
	"fmt"
	"net/url"

	remoteagent "google.golang.org/adk/v2/agent/remoteagent/v2"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/agenttool"
)

// BuildAgentTools turns each valid RegistryEntry into a callable tool.Tool
// backed by a real A2A remote agent (agenttool.New wrapping
// remoteagent.NewA2A). A per-entry validation or construction failure is
// skipped with a warning, not fatal — the rest of the registry still
// loads. Card fetching itself is deferred to first invocation (see
// remoteagent.AgentCardProvider) — this function makes no network calls.
func BuildAgentTools(entries []RegistryEntry) (tools []tool.Tool, warnings []string) {
	for _, e := range entries {
		if e.Name == "" {
			warnings = append(warnings, fmt.Sprintf("registry entry skipped: empty name (card_url=%q)", e.CardURL))
			continue
		}
		if e.CardURL == "" {
			warnings = append(warnings, fmt.Sprintf("registry entry %q skipped: empty card_url", e.Name))
			continue
		}
		if _, err := url.Parse(e.CardURL); err != nil {
			warnings = append(warnings, fmt.Sprintf("registry entry %q skipped: invalid card_url %q: %v", e.Name, e.CardURL, err))
			continue
		}

		remote, err := remoteagent.NewA2A(remoteagent.A2AConfig{
			Name:              e.Name,
			Description:       e.Description,
			AgentCardProvider: remoteagent.NewAgentCardProvider(e.CardURL),
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("registry entry %q skipped: %v", e.Name, err))
			continue
		}

		tools = append(tools, agenttool.New(remote, nil))
	}
	return tools, warnings
}
```

- [ ] **Step 4: Run unit tests to verify they pass**

Run: `go test ./internal/agent/... -run TestBuildAgentTools -v`
Expected: PASS (all 5 subtests).

- [ ] **Step 5: Write the failing live wire-protocol test**

This is the one test in this plan that proves the whole mechanism actually
works end to end over real A2A wire format — a real `httptest.Server`
serving a real `a2a.AgentCard` + a real JSON-RPC `SendMessage` responder
(via `a2asrv`'s own public server stack, not a hand-rolled JSON encoder),
and a real ADK runner driving the resulting `agenttool`.

**This exact composition (`AgentCardProvider` + `agentcard.DefaultResolver`
HTTP fetch, driven end-to-end through a real runner) has no precedent in
ADK's own test suite** — `a2a_agent_test.go`/`a2a_e2e_test.go` always set
`A2AConfig.AgentCard` directly, bypassing the resolver. Each piece
(`AgentCardProvider`'s resolver, and `agenttool`+`remoteagent` given a
static `AgentCard`) is independently tested upstream; this test is new
ground for their composition. If it fails in a way that looks like an ADK
bug rather than a mistake in this test, stop and dispatch the
`adk-api-verifier` subagent rather than guessing further.

Append to `internal/agent/agentcomm_test.go`:

```go
import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"
)

// echoExecutor is a minimal a2asrv.AgentExecutor: it always replies with a
// single fixed text message, regardless of the incoming request. Enough to
// prove one full round trip through the real A2A wire protocol.
type echoExecutor struct{}

func (echoExecutor) Execute(ctx context.Context, ec *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		yield(a2a.NewMessage(a2a.MessageRoleAgent, a2a.NewTextPart("hello from remote")), nil)
	}
}

func (echoExecutor) Cancel(ctx context.Context, ec *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {}
}

// runnableTool is the unexported surface agenttool's returned tool.Tool
// actually implements (tool.Tool itself only has Name/Description/
// IsLongRunning — no call method). Declared locally: Go interface
// satisfaction is structural, so this works across package boundaries.
type runnableTool interface {
	Run(ctx agent.Context, args any) (map[string]any, error)
}

func TestBuildAgentTools_InvokesRealAgentOverA2AWireProtocol(t *testing.T) {
	mux := http.NewServeMux()

	var cardURL string
	mux.HandleFunc("/.well-known/agent-card.json", func(w http.ResponseWriter, r *http.Request) {
		card := a2a.AgentCard{
			Name: "remote-echo",
			SupportedInterfaces: []*a2a.AgentInterface{
				a2a.NewAgentInterface(cardURL, a2a.TransportProtocolJSONRPC),
			},
			Capabilities: a2a.AgentCapabilities{Streaming: true},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := jsonEncode(w, card); err != nil {
			t.Errorf("encode agent card: %v", err)
		}
	})
	mux.Handle("/", a2asrv.NewJSONRPCHandler(a2asrv.NewHandler(echoExecutor{})))

	server := httptest.NewServer(mux)
	defer server.Close()
	cardURL = server.URL

	tools, warnings := BuildAgentTools([]RegistryEntry{
		{Name: "remote-echo", Description: "A test remote agent.", CardURL: server.URL},
	})
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1", len(tools))
	}
	remoteTool, ok := tools[0].(runnableTool)
	if !ok {
		t.Fatalf("tools[0] (%T) does not implement runnableTool", tools[0])
	}

	var gotResult map[string]any
	var runErr error
	harness, err := agent.New(agent.Config{
		Name: "harness",
		Run: func(ic agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				toolCtx := agent.NewToolContext(ic, "", &session.EventActions{}, nil)
				gotResult, runErr = remoteTool.Run(toolCtx, map[string]any{"request": "hi"})
				yield(&session.Event{}, nil)
			}
		},
	})
	if err != nil {
		t.Fatalf("agent.New(harness) error = %v", err)
	}

	ctx := context.Background()
	sessionService := session.InMemoryService()
	sess, err := sessionService.Create(ctx, &session.CreateRequest{AppName: "test", UserID: "u"})
	if err != nil {
		t.Fatalf("sessionService.Create() error = %v", err)
	}
	r, err := runner.New(runner.Config{AppName: "test", Agent: harness, SessionService: sessionService})
	if err != nil {
		t.Fatalf("runner.New() error = %v", err)
	}

	message := genai.NewContentFromText("go", genai.RoleUser)
	for _, runErr2 := range r.Run(ctx, "u", sess.Session.ID(), message, agent.RunConfig{
		StreamingMode: agent.StreamingModeNone,
	}) {
		if runErr2 != nil {
			t.Fatalf("runner.Run() error = %v", runErr2)
		}
	}

	if runErr != nil {
		t.Fatalf("remoteTool.Run() error = %v", runErr)
	}
	if gotResult == nil {
		t.Fatal("remoteTool.Run() returned nil result")
	}
}
```

Note: `jsonEncode` above is a one-line local helper
(`json.NewEncoder(w).Encode(v)`) — add it (or inline
`json.NewEncoder(w).Encode(card)` directly and drop the helper) along with
the `encoding/json` import when writing this file; it's omitted from the
snippet above only to keep the import block focused on ADK/A2A types.

- [ ] **Step 6: Run the test, expect compile or runtime friction**

Run: `go test ./internal/agent/... -run TestBuildAgentTools_InvokesRealAgentOverA2AWireProtocol -v`

This is genuinely new-ground composition (see the note in Step 5) — expect
at least one iteration here. Likely friction points, in order of
likelihood, and how to resolve each without guessing:

1. **Compile error on `a2a.NewMessage`/`a2a.NewTextPart`/`a2a.MessageRoleAgent`
   signatures.** Run `go doc github.com/a2aproject/a2a-go/v2/a2a.NewMessage`
   and `go doc .../a2a.NewTextPart` to get the exact signatures — do not
   guess a fix.
2. **`a2asrv.NewHandler`/`a2asrv.AgentExecutor` requires a third method**
   (e.g. `Cleanup`) that this plan's `echoExecutor` doesn't implement —
   the compiler error will name the missing method exactly; add a no-op
   implementation with that exact signature (check
   `go doc github.com/a2aproject/a2a-go/v2/a2asrv.AgentExecutor` for the
   real interface if so).
3. **The client sends `SendStreamingMessage` (SSE) instead of `SendMessage`**
   — this happens if `agent.RunConfig{StreamingMode: ...}` isn't
   `agent.StreamingModeNone` by the time it reaches the remote-agent call.
   Confirm the `agent.RunConfig{StreamingMode: agent.StreamingModeNone}`
   passed to `r.Run(...)` above is actually the one in effect (it should
   propagate to the tool call transitively — if it doesn't, that's a real
   finding, not a test bug; dispatch `adk-api-verifier` to check how
   `agenttool`/`remoteagent` derive the streaming mode for the sub-call).
4. **`agent.NewToolContext`'s `actions` param panics or behaves
   unexpectedly with an empty `&session.EventActions{}`** — check
   `go doc google.golang.org/adk/v2/agent.NewToolContext`'s doc comment
   (already confirms a nil `actions` is fine, allocating fresh — try `nil`
   instead of `&session.EventActions{}` if the literal struct causes
   issues).

Iterate until this test passes for real — do not weaken the assertions or
skip the test to get a green build. If truly stuck after a genuine attempt
at each of the above, stop and dispatch `adk-api-verifier` with the exact
compiler/runtime error rather than continuing to guess.

- [ ] **Step 7: Run the full package test suite**

Run: `go test ./internal/agent/... -v`
Expected: PASS, all tests including the pre-existing ones (this proves the
new files didn't break anything already in the package).

- [ ] **Step 8: Commit**

```bash
git add internal/agent/agentcomm.go internal/agent/agentcomm_test.go
git commit -m "feat: build agenttool-wrapped remote-agent tools from the registry"
```

---

### Task 5: Wire `agentTools` into `NewRootAgent`

**Files:**
- Modify: `internal/agent/root.go`
- Modify: `internal/agent/root_test.go`

**Interfaces:**
- Consumes: `tool.Tool` (stdlib ADK type, `google.golang.org/adk/v2/tool`).
- Produces: `func NewRootAgent(search, loop agent.Agent, m model.LLM, agentTools []tool.Tool) (agent.Agent, error)`
  — Task 6 (`run.go`) calls this with the 4th argument from Task 4's
  `BuildAgentTools` output (or `nil` when `--agents-config` is absent).

- [ ] **Step 1: Update the existing test's call site (it will not compile otherwise)**

In `internal/agent/root_test.go`, find:

```go
	root, err := NewRootAgent(search, loop, stubModel{})
```

Replace with:

```go
	root, err := NewRootAgent(search, loop, stubModel{}, nil)
```

- [ ] **Step 2: Write the new failing test for agentTools merging**

Append to `internal/agent/root_test.go`:

```go
// fakeTool is the minimal tool.Tool implementation needed to prove
// NewRootAgent merges a supplied []tool.Tool into the agent's Tools
// without needing a real agenttool/remoteagent round trip (that's
// agentcomm_test.go's job) — this test is only about root.go's wiring.
type fakeTool struct{ name string }

func (f fakeTool) Name() string        { return f.name }
func (f fakeTool) Description() string { return "a fake tool for testing" }
func (f fakeTool) IsLongRunning() bool { return false }

func TestNewRootAgent_AcceptsNonNilAgentToolsWithoutError(t *testing.T) {
	brain := testBrainForReview(t)

	gen, err := NewGeneratorAgent(stubModel{})
	if err != nil {
		t.Fatalf("NewGeneratorAgent error = %v", err)
	}
	review, err := NewReviewAgent(stubModel{}, brain)
	if err != nil {
		t.Fatalf("NewReviewAgent error = %v", err)
	}
	loop, err := NewSelfCorrectingLoop(gen, review, 4)
	if err != nil {
		t.Fatalf("NewSelfCorrectingLoop error = %v", err)
	}
	search, err := NewGeneratorAgent(stubModel{})
	if err != nil {
		t.Fatalf("NewGeneratorAgent (search stand-in) error = %v", err)
	}

	root, err := NewRootAgent(search, loop, stubModel{}, []tool.Tool{fakeTool{name: "remote-echo"}})
	if err != nil {
		t.Fatalf("NewRootAgent() error = %v", err)
	}

	// Construction succeeding with a non-nil agentTools slice, and
	// SubAgents staying at exactly 2 (search + loop, untouched by the new
	// param), is what this task's change is responsible for — the tool
	// actually being invokable end-to-end is proven in agentcomm_test.go.
	sub := root.SubAgents()
	if len(sub) != 2 {
		t.Fatalf("SubAgents() len = %d, want 2 (agentTools must not affect SubAgents)", len(sub))
	}
}
```

Add `"google.golang.org/adk/v2/tool"` to `root_test.go`'s import block.

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/agent/... -run TestNewRootAgent -v`
Expected: FAIL — `NewRootAgent` still takes 3 args, compile error.

- [ ] **Step 4: Update the implementation**

Replace `internal/agent/root.go` in full:

```go
package agent

import (
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
)

// NewRootAgent coordinates a one-shot Search lookup with the self-correcting
// generator/review loop (see NewSelfCorrectingLoop). loop is itself an
// agent.Agent — LoopAgent implements the same interface as any other
// sub-agent — so root can delegate to it exactly like search.
//
// agentTools is an optional set of additional tools (typically
// agenttool-wrapped remote agents built by BuildAgentTools) the root LLM
// can choose to call. nil/empty reproduces the exact behavior of the
// pre-agentTools NewRootAgent — this parameter is purely additive.
func NewRootAgent(search, loop agent.Agent, m model.LLM, agentTools []tool.Tool) (agent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "root",
		Model:       m,
		Description: "Coordinates Search and the self-correcting generate/review loop to handle a task.",
		Instruction: "Delegate lookups to the search agent first if the task needs " +
			"external information, then delegate to the loop agent to draft and " +
			"refine a final answer. Always surface the loop's final approved " +
			"answer and the review agent's verdict to the user.",
		SubAgents: []agent.Agent{search, loop},
		Tools:     agentTools,
	})
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/agent/... -v`
Expected: PASS — all tests in the package, including the updated
`TestNewRootAgent_WrapsSearchAndLoop` and the new
`TestNewRootAgent_AcceptsNonNilAgentToolsWithoutError`.

- [ ] **Step 6: Commit**

```bash
git add internal/agent/root.go internal/agent/root_test.go
git commit -m "feat: accept optional agentTools in NewRootAgent"
```

---

### Task 6: `--agents-config` CLI flag

**Files:**
- Modify: `cmd/agentic-hooks/run.go`
- Modify: `cmd/agentic-hooks/run_test.go`

**Interfaces:**
- Consumes: `myagent.LoadRegistry(path string) ([]RegistryEntry, error)`
  (Task 2), `myagent.BuildAgentTools(entries []RegistryEntry) ([]tool.Tool, []string)`
  (Task 4), `myagent.NewRootAgent(search, loop agent.Agent, m model.LLM, agentTools []tool.Tool) (agent.Agent, error)`
  (Task 5).
- Produces: `func loadAgentTools(path string, out io.Writer) ([]tool.Tool, error)`
  — a new unexported helper, extracted so this logic is unit-testable the
  same way `promptForApprovalAndRecordFeedback` already is in this file.

- [ ] **Step 1: Write the failing tests**

Append to `cmd/agentic-hooks/run_test.go`:

```go
func TestLoadAgentTools_EmptyPathIsNoOp(t *testing.T) {
	out := &bytes.Buffer{}
	tools, err := loadAgentTools("", out)
	if err != nil {
		t.Fatalf("loadAgentTools(\"\") error = %v", err)
	}
	if tools != nil {
		t.Errorf("loadAgentTools(\"\") tools = %v, want nil", tools)
	}
	if out.String() != "" {
		t.Errorf("loadAgentTools(\"\") wrote output %q, want none", out.String())
	}
}

func TestLoadAgentTools_MissingFileReturnsError(t *testing.T) {
	out := &bytes.Buffer{}
	_, err := loadAgentTools(filepath.Join(t.TempDir(), "nope.yaml"), out)
	if err == nil {
		t.Fatal("loadAgentTools() error = nil, want non-nil for a missing file")
	}
}

func TestLoadAgentTools_ValidRegistryReturnsTools(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agents.yaml")
	content := "- name: example-agent\n  description: An example.\n  card_url: http://localhost:9003\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out := &bytes.Buffer{}
	tools, err := loadAgentTools(path, out)
	if err != nil {
		t.Fatalf("loadAgentTools() error = %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1", len(tools))
	}
	if out.String() != "" {
		t.Errorf("loadAgentTools() wrote output %q, want none (no warnings expected)", out.String())
	}
}

func TestLoadAgentTools_WarnsOnInvalidEntryButReturnsToolsForRest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agents.yaml")
	content := "- name: \"\"\n  description: missing name\n  card_url: http://localhost:9003\n" +
		"- name: valid-agent\n  description: fine\n  card_url: http://localhost:9004\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out := &bytes.Buffer{}
	tools, err := loadAgentTools(path, out)
	if err != nil {
		t.Fatalf("loadAgentTools() error = %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1 (only the valid entry)", len(tools))
	}
	if !strings.Contains(out.String(), "warning:") {
		t.Errorf("output = %q, want a warning about the skipped entry", out.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/agentic-hooks/... -run TestLoadAgentTools -v`
Expected: FAIL — `loadAgentTools` undefined.

- [ ] **Step 3: Write the implementation**

In `cmd/agentic-hooks/run.go`, add `"google.golang.org/adk/v2/tool"` to
the import block, then add this function (place it near
`promptForApprovalAndRecordFeedback`, same file):

```go
// loadAgentTools loads the optional --agents-config registry and builds
// the resulting tool.Tool set, printing any per-entry warnings to out. An
// empty path is a no-op (nil tools, no error, no output) — this is the
// default, --agents-config-absent behavior, and must reproduce today's
// exact root-agent construction.
func loadAgentTools(path string, out io.Writer) ([]tool.Tool, error) {
	if path == "" {
		return nil, nil
	}

	entries, err := myagent.LoadRegistry(path)
	if err != nil {
		return nil, err
	}

	tools, warnings := myagent.BuildAgentTools(entries)
	for _, w := range warnings {
		fmt.Fprintf(out, "warning: %s\n", w)
	}
	return tools, nil
}
```

Then in `newRunCmd()`'s `RunE`, find:

```go
			root, err := myagent.NewRootAgent(search, loop, m)
			if err != nil {
				return err
			}
```

Replace with:

```go
			agentTools, err := loadAgentTools(agentsConfigPath, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			root, err := myagent.NewRootAgent(search, loop, m, agentTools)
			if err != nil {
				return err
			}
```

And in `newRunCmd()`'s var block, find:

```go
	var knowledgeDir string
	var mcpCommand string
	var mcpArgs []string
	var maxIterations uint
	var feedbackDir string
```

Replace with:

```go
	var knowledgeDir string
	var mcpCommand string
	var mcpArgs []string
	var maxIterations uint
	var feedbackDir string
	var agentsConfigPath string
```

And after the existing `cmd.Flags().StringVar(&feedbackDir, ...)` line,
add:

```go
	cmd.Flags().StringVar(&agentsConfigPath, "agents-config", "", "optional path to a YAML registry of remote agents to expose as tools (see internal/agent.RegistryEntry)")
```

Do **not** add `cmd.MarkFlagRequired("agents-config")` — it must stay
optional, matching the design's "no default like `--policy-file`, matches
your answer" decision.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/agentic-hooks/... -v`
Expected: PASS — all tests in the package, including the new
`TestLoadAgentTools_*` tests and every pre-existing test in this file
(`TestPromptForApprovalAndRecordFeedback_*`, etc.) unmodified and still
green.

- [ ] **Step 5: Commit**

```bash
git add cmd/agentic-hooks/run.go cmd/agentic-hooks/run_test.go
git commit -m "feat: add optional --agents-config flag wiring remote agent tools"
```

---

### Task 7: Full-repo verification and doc sync

**Files:**
- Modify: `SESSION_HANDOFF.md` (append a new dated entry)
- Modify: `docs/reference/cli.md` (document the new flag, if this file
  enumerates `run`'s flags — check before assuming)

**Interfaces:** none — this is the final verification + documentation
task, no new code.

- [ ] **Step 1: Full clean build/vet/test**

Run: `make check` (equivalent to `go vet ./... && go test ./... -v && go build ...`)
Expected: all green — every test in the repo, not just the new ones. If
anything outside `internal/agent`/`cmd/agentic-hooks` fails, stop; that's
a regression this plan's changes are not supposed to cause (e.g. an
import cycle from the new `google.golang.org/adk/v2/tool` import, or a
`go.sum` inconsistency from Task 3).

- [ ] **Step 2: Confirm the additive-only guarantee by inspection**

Run: `git diff --stat` (or `git diff` in full) and confirm:
- `internal/agent/root.go`'s diff is exactly the signature/field addition
  from Task 5, no other logic changed.
- `cmd/agentic-hooks/run.go`'s diff is exactly the flag + `loadAgentTools`
  call from Task 6, no other logic changed.
- No existing test's assertions were weakened to make it pass.

- [ ] **Step 3: Update `docs/reference/cli.md` if it documents `run`'s flags**

Run: `grep -n "search-mcp-server\|feedback-dir" docs/reference/cli.md`
If this file lists `run`'s existing flags (likely, given the doc overhaul
mentioned in `SESSION_HANDOFF.md`), add a row/entry for `--agents-config`
consistent with that file's existing format for the other optional flags.
If the grep finds nothing, this file doesn't enumerate flags this way —
skip this step and note that in the SESSION_HANDOFF entry instead.

- [ ] **Step 4: Append a dated entry to SESSION_HANDOFF.md**

Add a new `## 2026-07-06 — Agent-to-agent communication: implemented (vN
session, same day)` section (pick the next session letter/number per this
file's existing convention) summarizing: what was built (registry.go,
agentcomm.go, root.go's 4th param, `--agents-config`), the one corrected
spec detail (`/.well-known/agent-card.json`, not `/.well-known/agent.json`),
the dropped `error` return from `BuildAgentTools` vs. the spec's sketch,
the real test counts (before/after), and that ADR-0010 now supersedes
ADR-0001's "no network A2A" consequence. Mirror the terse, evidence-cited
style of this file's existing entries — no marketing language, state
what was actually run and observed (per `POLICY.md` 08.1: never claim a
test passed without having run it this session).

- [ ] **Step 5: Update the "Outstanding tasks" list**

In `SESSION_HANDOFF.md`'s "Outstanding tasks, all in one place" section,
mark item 1 ("Start a fresh `superpowers:brainstorming` cycle for
agent-to-agent communication") as done, pointing at this plan and the new
dated entry — it's no longer outstanding once this task completes.

- [ ] **Step 6: Commit**

```bash
git add SESSION_HANDOFF.md docs/reference/cli.md
git commit -m "docs: record agent-to-agent communication implementation in session handoff"
```

(Only run Task 1's and this task's commits if the user has asked for
commits this session — otherwise leave everything staged/modified per the
standing no-commit instruction, and say so explicitly in the final report
to the user.)
