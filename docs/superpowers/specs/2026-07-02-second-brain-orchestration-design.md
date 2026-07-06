# Second Brain Orchestration — Design Spec v0.1

Status: draft, pending user review
Date: 2026-07-02
Diagram: https://claude.ai/code/artifact/063c19ed-c306-4cff-b389-1d0a3af59c3a

## 1. Purpose

`agentic-hooks` is a single Go binary with two jobs:

1. Run a small multi-agent pipeline (Search + Review) that can look things up and
   review code against a personal "Second Brain" of coding principles (SOLID,
   clean code, Go style, project conventions).
2. Expose that Second Brain as an MCP server so any external coding agent
   (Claude Code, Cursor, etc.) can query it directly, independent of the
   pipeline above.

This spec covers both, since they share one component (the Second Brain) and
one process (the CLI binary). Everything else from the original brainstorm
(visual canvas, additional agents, true network A2A, RAG/semantic search,
tracing dashboard) is explicitly deferred — see §7.

## 2. Non-goals (this iteration)

- No network-based A2A (agent cards, remote agent servers). Sub-agents are
  in-process ADK delegation only.
- No database. No semantic/vector search. No RAG pipeline.
- No frontend. No visual flow canvas.
- No image-generation agent, no fetch-agent separate from search.
- No live/inline Genkit involvement — Genkit only runs offline, after the
  fact, never in the request path.

## 3. Architecture

One Go binary, `agentic-hooks`, built primarily on **ADK Go v2**
(`google.golang.org/adk/v2`). Two entry points (cobra subcommands), sharing
one Second Brain component:

```
agentic-hooks run "<task>"        agentic-hooks serve
        |                                |
        v                                v
  ADK Runtime (in-process)        MCP Server (stdio)
  root agent -> Search sub-agent  github.com/modelcontextprotocol/go-sdk
             -> Review sub-agent  exposes Second Brain as MCP tools
        |                                |
        v                                v
     Second Brain (shared) -----------------
     directory of OKF-frontmatter Markdown files
```

**Why ADK Go, not Genkit, as the runtime.** Both are Google Go frameworks but
serve different purposes: Genkit is single-flow/prototyping oriented; ADK Go
v2 has a graph-based workflow engine, built-in sub-agent delegation, native
MCP client support (`adk/tool/mcptoolset`, built on
`modelcontextprotocol/go-sdk`), and first-class HITL confirmation
(`tool.Context.RequestConfirmation()` / `.ToolConfirmation()`, plus a
`RequireConfirmation` flag directly on `mcptoolset.Config`). Using both
frameworks for the same
responsibility (model calls, orchestration) would mean two tracing
pipelines and two config surfaces for no capability gained — confirmed
against ADK Go's `model.LLM` interface, which technically *could* wrap
Genkit's `ai.Generate()`, but there's no reason to.

**Genkit's actual role**: offline evaluation only. ADK session traces are
exported as a dataset (`{testCaseId, input, output, context, traceIds}`,
Genkit's documented "raw evaluation" format for externally-produced output)
and scored via `genkit eval:run` + `DefineEvaluator` (LLM-as-judge). This
runs after a session completes, never inside it.

**MCP is bidirectional, one library.** `github.com/modelcontextprotocol/go-sdk`
is used both ways: as a *client* inside the ADK Search sub-agent
(`mcptoolset.New()`, consuming external MCP servers — e.g. a docs-lookup
server, configurable, not hardcoded to one vendor) and as a *server* in
`agentic-hooks serve` (`mcp.NewServer` + `mcp.AddTool`, exposing Second Brain
queries to external agents). One dependency, both directions — no need for
Genkit's own MCP plugin (`GenkitMCPServer`, which sits on a different
underlying library, `mark3labs/mcp-go`) anywhere in this project.

## 4. Components

### 4.1 CLI (`cmd/agentic-hooks`)
Cobra-based, two subcommands:
- `run <task>`: starts the ADK runtime in-process, streams the result to
  stdout, no network hop.
- `serve`: starts the MCP server over stdio, meant to be spawned by an
  external agent host (Claude Code, Cursor, etc.), not run interactively.

### 4.2 ADK Runtime (`internal/agent`)
- **Root agent**: receives the task, delegates to Search and/or Review via
  ADK's built-in sub-agent transfer (in-process, no A2A protocol).
- **Search sub-agent**: has one or more MCP-client tools attached via its
  `Toolsets []tool.Set` field (`mcptoolset.New(mcptoolset.Config{...})`),
  pointed at externally configured MCP servers (config-driven, so the actual
  server — docs lookup, web search, etc. — is swappable without code
  changes).
- **Review sub-agent**: reads the Second Brain directly via a Go function
  call (`internal/secondbrain.Query(...)`), no protocol overhead in-process.
  Produces a verdict. MVP review is LLM-over-diff-text only, no AST/structural
  parsing (see §7 — tree-sitter deferred). To keep that door open without a
  later refactor, the review entry point takes an optional structural-facts
  parameter now: `Review(ctx, diff, structuralFacts)`, where
  `structuralFacts` is `nil` in the MVP and would be populated by a future
  tree-sitter pre-filter stage.
- **Root agent**: a coordinator `LlmAgent` with `SubAgents: []agent.Agent{search, review}`
  in its `llmagent.Config` — ADK auto-generates the delegation tools
  (`request_task_search`, `request_task_review`), no manual transfer logic
  needed.
- **HITL gate**: before a Review verdict is returned to the caller, the
  tool checks `ctx.ToolConfirmation()`; if `nil`, it calls
  `ctx.RequestConfirmation(hint, payload)` and pauses. The CLI surfaces
  this as an approve/reject prompt; on the next invocation
  `ctx.ToolConfirmation()` returns the human's decision. Rejected verdicts
  are discarded, not returned. Rationale: no agent output that could
  influence a code decision should reach the user as final without a
  human checkpoint — this is close to free given ADK's built-in primitive,
  so it ships in the MVP rather than being deferred.

### 4.3 Second Brain (`internal/secondbrain`)
A directory of Markdown files, each one "concept" in **OKF** terms
(Open Knowledge Format — https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md).
Each file has OKF frontmatter plus a freeform Markdown body — OKF supplies
the envelope (`type`/`title`/`description`/`tags`/`timestamp`, all real
spec fields), the body is the actual coding-principle content:

```markdown
---
type: principle
title: Single Responsibility Principle
description: Each component has one reason to change.
tags: [solid, architecture]
timestamp: 2026-07-02
---

A component should have one, and only one, reason to change. In review,
flag functions/types that mix unrelated concerns (e.g. business logic and
I/O, or validation and persistence) even if each concern is individually
small.
```

Per spec, the file's path (minus `.md`) is the concept's identifier — no
separate `id` field. Example layout:

```
knowledge/
  solid/single-responsibility.md
  solid/open-closed.md
  clean-code/no-comments-unless-non-obvious.md
  go/effective-go-naming.md
  go/error-handling.md
```

`internal/secondbrain` is a plain Go package: walks the directory, parses
YAML frontmatter + Markdown body, no external SDK (none exists for OKF in
Go — it's a documented plaintext format, not a wire protocol). Exposes:
- `List(filter tags/type) []Concept`
- `Get(id string) (Concept, error)`
- `Query(topic string) []Concept` (substring/tag match for MVP — no
  semantic/vector search this iteration, see §7)

This one package backs both consumers: the Review sub-agent (direct call)
and the MCP server (wrapped as tools).

### 4.4 MCP Server (`internal/mcpserver`)
Built on `modelcontextprotocol/go-sdk`. Exposes two tools backed by
`internal/secondbrain`:
- `list_knowledge` (optional tag/type filter)
- `get_knowledge` (by id/path)

No HITL here — this path only answers read queries, it never produces a
verdict or takes an action.

### 4.5 Offline Eval (`internal/eval`, deferred to a later phase)
Placeholder package boundary only in this iteration — exports ADK session
traces to a Genkit raw-eval dataset. Not required for the MVP to be useful;
tracked in §7.

## 5. Data flow

**`agentic-hooks run "review this diff"`**
1. CLI starts ADK root agent in-process.
2. Root agent delegates to Search sub-agent if external lookup is needed
   (MCP client call via `mcptoolset`).
3. Root agent delegates to Review sub-agent, which calls
   `secondbrain.Query()` directly (function call, no protocol).
4. Review produces a verdict; ADK raises `RequestInputEvent`.
5. CLI prompts for approve/reject. On approve, verdict streams to stdout.
   On reject, nothing is returned as final — session ends.
6. Session trace persists to disk for later offline eval (§4.5).

**`agentic-hooks serve`**
1. CLI starts the MCP server over stdio.
2. External agent host calls `list_knowledge` / `get_knowledge`.
3. Server calls `internal/secondbrain` directly, returns results.
4. No ADK involvement on this path at all.

## 6. Error handling

- **MCP client connection failure** (Search sub-agent's external server
  unreachable): Search sub-agent returns a typed error to the root agent,
  which surfaces "search unavailable" in the response rather than failing
  the whole run — Review can still proceed using only the Second Brain.
- **Second Brain parse failure** (malformed frontmatter in one file):
  logged and skipped at load time, not fatal to the whole directory scan —
  one bad file shouldn't take down the knowledge base.
- **HITL timeout/no response**: treated as reject (fail closed, not open —
  consistent with the zero-trust rationale in §4.2).
- **MCP server tool errors** (`serve` mode): returned as standard MCP error
  responses to the calling agent, process stays up.

## 7. Deferred (explicit roadmap, not silently dropped)

Each of these is a candidate for its own future design spec, not built now:

- Frontend visual canvas (Vue + VueFlow + Pinia) to observe/edit runs.
- True network A2A (agent cards, remote agent servers) if agents are ever
  built by other teams/languages.
- Additional agents (Fetch, Image-gen) beyond Search/Review.
- Semantic/vector search over the Second Brain (current `Query` is
  substring/tag match only).
- Wired-up offline eval loop (`internal/eval` is a package boundary only
  right now, not implemented).
- Token-bucket/rate limiting, cost tracking, full observability/tracing
  dashboard.
- Tree-sitter-based structural/AST pre-filtering ahead of the Review agent
  (function length, nesting depth, SRP-by-structure checks). MVP stays
  LLM-over-diff-text only; `Review()`'s `structuralFacts` param (§4.2) is
  the seam this plugs into later, no refactor needed when it lands. Go
  tree-sitter bindings are CGo-based (`tree-sitter/go-tree-sitter` or
  `smacker/go-tree-sitter`) — verify current state before implementing,
  don't assume a pure-Go option exists.

## 8. Testing approach

- `internal/secondbrain`: table-driven unit tests over a fixture directory
  of `.md` files (valid, missing-required-field, malformed-YAML cases).
- `internal/agent`: ADK sub-agent delegation tested with a fake/mock MCP
  server for Search, and a fixture Second Brain for Review — no live model
  calls in unit tests.
- `internal/mcpserver`: integration test using the MCP Go SDK's in-memory
  client/server pair against a fixture Second Brain.
- End-to-end: one smoke test invoking `run` against a fixture task with a
  scripted HITL approval, run manually/CI, not part of the fast unit suite.

## 9. Open items for user review

- Exact MCP server(s) the Search sub-agent points at by default (none
  hardcoded per earlier decision — needs at least one configured for local
  dev/testing).
- `go.mod` currently declares `go 1.21.13`; latest stable is 1.26.4 — bump
  before implementation starts.
