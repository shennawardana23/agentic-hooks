# Design: Agent-to-Agent Communication via ADK's Native A2A Primitives

| | |
|---|---|
| **Date** | 2026-07-06 |
| **Status** | Approved (pending final spec review) |
| **Session** | v7, same day as the agent-policy-layer implementation session |

## Context

`SESSION_HANDOFF.md`'s "Answered, 2026-07-06" entry records the user
confirming this is the same initiative previously raised twice as a
"massive audit of Provider Fallback + a Skills system" — reframed as
"Intelligent Communication Through Tool Calling": agents should see each
other's public manifests and decide who to delegate to via a communication
tool, instead of a fixed orchestration graph. This **reopens** the
previously locked decision "no network A2A, in-process ADK delegation
only" (see `MEMORY.md`) — user-confirmed reopening, not a silent scope
change.

Two sibling repos were named as prior art:
`archpublicwebsite-agentic` and `go-adk-q`. Both were checked directly
(not assumed) before this design was written — see "Prior art, verified"
below, since an earlier draft of this session's own scoping questions
contained a wrong assumption about them that was caught and corrected
before this design was finalized.

## Prior art, verified (not assumed)

**ADK v2 (`google.golang.org/adk/v2@v2.0.0`, this project's exact pinned
dependency) already ships native A2A support**, verified via `go doc` and
direct module-cache source reads (dispatched to the `adk-api-verifier`
subagent, cross-checked inline):

- `tool/skilltoolset` — single-agent local capability loading
  (`list_skills`/`load_skill`, SKILL.md-format frontmatter per
  `agentskills.io/specification#frontmatter`). **Does not solve
  delegation** — its verbs are "load instructions for myself," not "call
  another agent and get a result." Ruled out.
- `agent/remoteagent` + `server/adka2a` (no `/v2` suffix) — both marked
  `// Deprecated: Use .../v2 instead` in their package doc comments.
- `agent/remoteagent/v2` + `server/adka2a/v2` — the current, non-deprecated
  primitives. `remoteagent.NewA2A(A2AConfig) (agent.Agent, error)` wraps a
  real A2A-protocol remote agent (via `github.com/a2aproject/a2a-go/v2
  v2.3.1`, a real open-protocol SDK, not ADK-proprietary) so it implements
  ADK's own `agent.Agent` interface — confirmed droppable directly into an
  `llmagent.Config`'s `SubAgents`/`Tools`. `AgentCardProvider` re-resolves
  the remote agent's `AgentCard` manifest fresh on every invocation, not
  once at startup. Deep test coverage: `a2a_agent_test.go` (1437 lines),
  `a2a_e2e_test.go` (1276 lines) with real recorded HTTP round-trips
  (`.httprr` fixtures) covering streaming and mid-response error scenarios.
- `tool/agenttool` — `agenttool.New(agent.Agent, *Config) tool.Tool` wraps
  *any* `agent.Agent` (including a `remoteagent.NewA2A` result) as a
  callable `tool.Tool`, with the framework handling session management and
  result summarization (`Config.SkipSummarization`, default `false`).
- **No discovery-of-unknown-agents registry exists anywhere in ADK.**
  Registration is always URL-driven — you must already know the endpoint.
  This project still needs to supply that piece (a static config file, see
  below).

**Sibling repos, checked directly:**
- `archpublicwebsite-agentic`: has a **real, currently-implemented A2A
  server** — `internal/a2a/a2a.go` (`Server`/`New`/`Start`/`Stop`, no
  client-side code anywhere in the repo — server-only), backing
  `pwdev-mcp`'s `pwdev_orchestrator` agent, per **ADR-0004** (Accepted,
  2026-05-13): separate port 9003, agent card at the standard
  `/.well-known/agent.json` A2A discovery path. Pinned to
  `google.golang.org/adk v1.2.0` (pre-v2) and `github.com/a2aproject/a2a-go
  v0.3.13` — the **same major line as ADK v2's *deprecated*
  `remoteagent`/`adka2a` pair**, not the `v2` line this design builds on.
  Verified via `lsof`/`ps`: **not currently running** — no live target
  today, code + an Accepted architectural decision only.
- `go-adk-q`: `a2aproject/a2a-go` appears only in `go.mod`/`go.sum` as a
  transitive dependency of ADK itself — **no actual A2A usage** anywhere in
  the repo's source. Its `skills/` taxonomy and `model/failover` are
  unrelated to this design (not pulled in).

**Consequence for scope:** wiring this repo's client to the real
`pwdev-mcp` server is **not** part of this design. That would require
either upgrading `pwdev-mcp` to `a2a-go/v2`, or a separate, explicit
protocol-compatibility verification between `a2a-go v0.3.x` and
`a2a-go/v2` — neither is assumed here, and both are out of scope for this
repo's own work. This design produces a mechanism verified against a local
test fixture; pointing it at a real remote agent (`pwdev-mcp` or otherwise)
is a deliberately separate future task, gated on that other repo's own
upgrade decision.

## Decision

Build **Approach B**: reuse `agenttool` to auto-generate one tool per
configured remote agent, rather than hand-rolling a single generic
`call_agent(name, task)` tool with a custom invocation loop. Rationale:
`agenttool` + `remoteagent/v2` already implement session handling,
summarization, and error propagation with deep test coverage; wrapping N
registry entries into N `agenttool`-backed tools is a thin, low-risk layer
on top. A hand-rolled dynamic dispatch tool (Approach A) would duplicate
that machinery for no material benefit at this repo's current scale (a
handful of known remote agents via static config, not floating discovery
of arbitrary unknown agents).

**Additive only.** `root.go`'s existing `SubAgents: []agent.Agent{search,
loop}` is untouched. This design adds a 4th parameter to `NewRootAgent`
and one new optional CLI flag; omitting the flag reproduces today's exact
behavior.

## Architecture & components

Two new files in `internal/agent/`, matching the package's existing
one-concept-per-file convention (`search.go`, `review.go`, `root.go`,
`loop.go`):

- **`registry.go`**
  ```go
  type RegistryEntry struct {
      Name        string `yaml:"name"`
      Description string `yaml:"description"`
      CardURL     string `yaml:"card_url"`
  }
  func LoadRegistry(path string) ([]RegistryEntry, error)
  ```
  Parses a small YAML file (reuses `go.yaml.in/yaml/v3`, already a
  dependency after this session's dependency-bump work). Structurally
  mirrors `secondbrain.Load`.

- **`agentcomm.go`**
  ```go
  func BuildAgentTools(entries []RegistryEntry) (tools []tool.Tool, warnings []string, err error)
  ```
  For each entry: `remoteagent.NewA2A(remoteagent.A2AConfig{Name:
  entry.Name, Description: entry.Description, AgentCardProvider:
  remoteagent.NewAgentCardProvider(entry.CardURL)})`, then
  `agenttool.New(remoteAgent, nil)` (default summarization — per your
  answer, final-answer-text only, not raw transcript). A per-entry
  construction failure is skipped with a warning appended, not fatal (see
  Error handling) — the rest of the registry still loads.

**Wiring:**
- `NewRootAgent(search, loop agent.Agent, m model.LLM, agentTools
  []tool.Tool) (agent.Agent, error)` — `agentTools` merges into
  `llmagent.Config.Tools`; `nil`/empty reproduces today's exact
  construction.
- `run.go` gains `--agents-config <path>` (optional, no default — matches
  your answer, not an always-on default like `--policy-file`). When set:
  `LoadRegistry` → `BuildAgentTools` → pass into `NewRootAgent`. Any
  warnings print to `cmd.OutOrStdout()` (consistent with the
  feedback-write-failure warning pattern already in `run.go`). When unset:
  `agentTools` is `nil`, zero behavior change.

## Data flow

**Startup** (only when `--agents-config` is given): registry loads, tools
are constructed — this is pure Go object construction, **no network
calls** yet. `AgentCardProvider` defers the actual card fetch to
invocation time (verified: cards re-resolve fresh per call, not cached at
startup — a remote agent's declared capabilities can change between calls
without redeploying the caller).

**Runtime**: root's LLM sees the new tools alongside its existing ones, by
name and description — no special-cased discovery step; this is the same
mechanism root already uses to decide whether to use any tool. If it picks
one: `agenttool`'s existing framework machinery fetches that entry's fresh
`AgentCard`, sends the task over real A2A JSON-RPC to `CardURL`, gets a
response, summarizes it, and returns final text into root's conversation —
indistinguishable to root from any other tool call. No new runtime code
path is introduced by this design beyond tool construction at startup;
everything at call time is existing, already-tested ADK machinery.

## Error handling

- **Registry load failure** (`--agents-config` given, file missing or the
  YAML itself doesn't parse): `run` fails fast with a clear local error —
  same pattern as `secondbrain.Load`'s existing failure path in `run.go`.
  No API call wasted on a bad config. `LoadRegistry` only validates that
  the file is well-formed YAML; it does not inspect individual entries.
- **Per-entry semantic failure** (a syntactically valid entry with an
  empty `name`/`card_url`, or a `card_url` that fails `url.Parse`):
  `BuildAgentTools` skips that one entry with a warning, not fatal —
  directly reuses the convention this session already shipped for
  `secondbrain.Load`'s malformed-concept-file handling (log + a returned
  warning list; one bad entry doesn't take down the rest). This is the
  only place entry-level validation happens — `LoadRegistry` deliberately
  does not duplicate it.
- **Runtime call failure** (remote agent unreachable, A2A protocol error
  mid-call): no new handling required. `remoteagent/v2` already has real
  e2e error-scenario test coverage (e.g.
  `TestA2ASingleHopFinalResponse_llm_mid-response_error`) for exactly this
  case; the error surfaces to root's LLM as a normal failed-tool-call,
  same as any other tool today.

## Testing

- `internal/agent/registry_test.go` — valid/malformed YAML fixtures,
  mirrors `secondbrain_test.go`'s style (`t.TempDir()` + `writeFixture`
  helper pattern).
- `internal/agent/agentcomm_test.go` — `BuildAgentTools` exercised against
  a local `httptest.Server` fake A2A endpoint (serving a real `AgentCard`
  JSON at `/.well-known/agent.json` plus a minimal JSON-RPC task
  responder) — a genuine wire-protocol round-trip, hermetic and CI-safe,
  no dependency on any sibling repo's uptime or `a2a-go` version.
- `cmd/agentic-hooks/run_test.go` — `--agents-config` is optional; root's
  construction is unaffected when the flag is absent.
- A `docs/how-to/` addition documenting how to point `--agents-config` at
  a *real* remote agent is explicitly deferred — not part of this design's
  deliverable, since no compatible live target currently exists (see
  "Consequence for scope" above).

## Explicitly out of scope (deferred, not forgotten)

- Wiring the real `pwdev-mcp` server in as a live registry entry — blocked
  on that repo's own `a2a-go`/ADK version decision, not this repo's work.
- Any discovery-of-unknown-agents mechanism (a registry service, capability
  search) — ADK itself has no such primitive; today's static config file
  is a deliberate, minimal starting point per YAGNI, not a placeholder for
  something bigger assumed-but-unbuilt.
- Migrating `search`/`loop` into the same dynamic-tool mechanism — this
  design is purely additive; the existing, already-verified production
  delegation path is untouched.
- Exposing `agentic-hooks` itself as an A2A server (so other agents could
  call *it*) — this design only builds the client side.
