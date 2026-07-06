# SYSTEM.md — Architecture Overview

Source of truth for architecture claims: [docs/superpowers/specs/2026-07-02-second-brain-orchestration-design.md](docs/superpowers/specs/2026-07-02-second-brain-orchestration-design.md).
This file is a living summary of it — if the two disagree, the design spec
plus actual source code wins, and this file should be corrected to match.
Deep rationale ("why ADK not Genkit", "why HITL is CLI-level") lives in
[docs/explanation/](docs/explanation/), not here — this file describes
structure, not justification.

## Two entry points, one shared component

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

Both subcommands are Cobra commands defined in `cmd/agentic-hooks/` (`run.go`,
`serve.go`, `main.go`). Neither path talks to the other at runtime — they
share code (`internal/secondbrain`), not a running process.

## Component responsibilities

### `cmd/agentic-hooks` (CLI entrypoint)
- `main.go` — root Cobra command, `version` subcommand, wires `run` and `serve`.
- `run.go` — builds the full ADK pipeline (Search + self-correcting loop),
  streams every event live tagged `[author]`, then runs the HITL
  approve/reject prompt and writes a feedback record.
- `serve.go` — loads the Second Brain, builds the MCP server, runs it over
  `mcp.StdioTransport{}`. Blocks on stdio; prints nothing by design.

### `internal/agent` (ADK runtime)
- `search.go` — one-shot MCP-client sub-agent; its `Toolsets` point at
  whatever `--search-mcp-server`/`--search-mcp-server-args` configures
  (nothing hardcoded).
- `generator.go` — drafts/revises a text answer.
- `review.go` — critiques a draft against the Second Brain
  (`matchConceptsInDiff` — title/tag substring match against the diff
  text, not `secondbrain.Query`, which matches the opposite direction),
  calls ADK's `exitlooptool` on `APPROVE`.
- `loop.go` — `NewSelfCorrectingLoop` wraps Generator+Review in
  `google.golang.org/adk/v2/agent/workflowagents/loopagent`, bounded by
  `--max-iterations` (default 4).
- `root.go` — `NewRootAgent` delegates to Search then the loop via ADK's
  built-in sub-agent transfer (in-process, no A2A protocol).

### `internal/mcpserver` (MCP server)
- `server.go` — `NewServer` registers two tools via `mcp.AddTool`:
  `list_knowledge` (optional `type`/`tag` filter) and `get_knowledge` (by
  id). Both are thin wrappers over `internal/secondbrain`, no protocol
  logic of their own. No HITL gate on this path — read-only by design.

### `internal/secondbrain` (knowledge base loader)
- Walks a directory of OKF-frontmatter Markdown files at startup.
  `Load(dir)` returns a `*Brain` exposing `List(type, tag)`, `Get(id)`,
  and `Query(topic)` (substring/tag match). One package backs both the
  Review sub-agent (direct function call) and the MCP server (wrapped as
  tools) — see [docs/reference/second-brain-frontmatter.md](docs/reference/second-brain-frontmatter.md)
  for the exact frontmatter schema and [SKILL.md](SKILL.md) for authoring
  guidance.
- Malformed frontmatter in one file is logged and skipped at load time,
  not fatal to the whole directory scan.

### `internal/feedback` (HITL audit log)
- `Append(dir, record)` writes one JSON line per HITL decision to
  `<feedback-dir>/feedback.jsonl`: `{timestamp, task, transcript,
  approved, reason}`. Append-only, unconditional — every run writes a
  record whether approved or rejected. Durable input for a future offline
  eval/training step, not a live training loop.

## Data flow

**`agentic-hooks run "<task>"`**
1. CLI starts the ADK root agent in-process.
2. Root agent delegates to Search if external lookup is needed (MCP
   client call via `mcptoolset`).
3. Root agent delegates to the self-correcting loop: Generator drafts,
   Review critiques against the Second Brain, looping until `APPROVE`
   (via `exit_loop`) or `--max-iterations` is hit.
4. CLI streams every sub-agent's text live, then prompts
   `Approve? [y/N]:`. Fail-closed: anything but literal `y`/`Y` is a
   reject.
5. `internal/feedback.Append` writes the decision unconditionally.
6. On approve, the final transcript prints to stdout; on reject, nothing
   is returned as final output.

**`agentic-hooks serve`**
1. CLI loads the Second Brain and starts the MCP server over stdio.
2. External agent host calls `list_knowledge` / `get_knowledge`.
3. Server calls `internal/secondbrain` directly, returns results.
4. No ADK involvement on this path at all — `serve` and `run` never share
   a live process.

## Error handling

- **MCP client connection failure** (Search's external server
  unreachable): Search returns a typed error to the root agent, which
  surfaces "search unavailable" rather than failing the whole run —
  Review can proceed using only the Second Brain.
- **Second Brain parse failure**: logged and skipped at load time, not
  fatal to the directory scan.
- **HITL no response / anything but `y`/`Y`**: treated as reject
  (fail-closed).
- **MCP server tool errors** (`serve` mode): returned as standard MCP
  error responses; the process stays up.

## Diagrams

- [docs/diagrams/architecture-overview.d2](docs/diagrams/architecture-overview.d2)
  (rendered [.svg](docs/diagrams/architecture-overview.svg)) — static
  component diagram.
- [docs/diagrams/run-sequence.mmd](docs/diagrams/run-sequence.mmd),
  [docs/diagrams/serve-sequence.mmd](docs/diagrams/serve-sequence.mmd) —
  Mermaid sequence diagrams for each entry point.
- [docs/diagrams/loop-state-machine.mmd](docs/diagrams/loop-state-machine.mmd) —
  Generator↔Review convergence loop as a state machine.
- Walkthrough tying these together: [docs/explanation/architecture-overview.md](docs/explanation/architecture-overview.md).
