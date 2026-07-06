# Architecture overview

`agentic-hooks` is one Go binary with two independent entry points that
share one component. This document walks through both paths and how they
connect. For exhaustive flag/schema listings, see the
[reference docs](../reference/); for step-by-step tasks, see the
[how-to guides](../how-to/).

![Architecture overview](../diagrams/architecture-overview.svg)

## Two entry points, one shared component

**`agentic-hooks run "<task>"`** starts the ADK Go v2 runtime in-process: a
Root agent delegates to a Search sub-agent (an MCP client that can query an
externally configured MCP server for supporting context) and to a
self-correcting Generator↔Review loop that drafts and critiques an answer
against the Second Brain. The Review agent reads the Second Brain directly
via a Go function call — no protocol overhead in-process. A converged
verdict is gated by a CLI-level human approve/reject prompt before it's
treated as final output, and every decision (approved or rejected) is
appended to `feedback/feedback.jsonl`.

```mermaid
sequenceDiagram
    participant U as Human
    participant R as Root Agent
    participant G as Generator
    participant V as Review
    participant B as Second Brain
    participant F as feedback.jsonl

    U->>R: run "<task>"
    R->>G: draft answer
    G-->>R: draft v1
    R->>V: review draft v1
    V->>B: match concepts (title/tag in diff)
    V-->>R: CHANGES_REQUESTED + reasons
    R->>G: revise (sees v1 verdict in history)
    G-->>R: draft v2
    R->>V: review draft v2
    V-->>R: APPROVE (exit_loop)
    R->>U: verdict + Approve? [y/N]
    U-->>R: y
    R->>F: append {task, transcript, approved, reason}
    R->>U: final transcript
```

**`agentic-hooks serve`** starts an MCP server over stdio, independent of
the ADK runtime entirely — no agent pipeline runs on this path. It exposes
the same Second Brain as two read-only MCP tools (`list_knowledge`,
`get_knowledge`) so any external MCP-compatible agent host (Claude Code,
Cursor, etc.) can query it directly.

```mermaid
sequenceDiagram
    participant E as External Agent Host
    participant M as MCP Server (stdio)
    participant B as Second Brain

    E->>M: initialize
    M-->>E: capabilities (tools: list_knowledge, get_knowledge)
    E->>M: tools/list
    M-->>E: [list_knowledge, get_knowledge]
    E->>M: tools/call list_knowledge {tag: "solid"}
    M->>B: List(type, tag)
    B-->>M: []Concept
    M-->>E: {concepts: [...]}
    E->>M: tools/call get_knowledge {id: "solid/single-responsibility"}
    M->>B: Get(id)
    alt concept found
        B-->>M: Concept
        M-->>E: {id, title, body}
    else concept missing
        B-->>M: error
        M-->>E: MCP error response (process stays up)
    end
```

## Why one shared component instead of two copies

Both entry points read the same `internal/secondbrain` package — a
directory of OKF-frontmatter Markdown files under `knowledge/`, walked and
parsed once per process start. The Review sub-agent calls it directly as a
Go function; the MCP server wraps it as two tool handlers. There is
exactly one source of truth for the Second Brain's content and query
logic, regardless of which entry point is driving a given process. See
[ADR 0004](../adr/0004-second-brain-as-markdown-not-database.md) for why
this is Markdown files rather than a database.

## Why the two entry points don't share a process

`run` and `serve` are separate CLI subcommands, each starting a fresh
process — they are never both active in the same binary invocation. This
matters for one specific wiring detail: when `run` needs a Search MCP
server to point its client at, `--search-mcp-server` can point at this
project's own `serve` subcommand as a valid stand-in (see the
[quick-start tutorial](../tutorials/first-run.md)), which means one
`run` invocation spawns a second, separate `serve` process as a
subprocess. The two entry points are architecturally independent even
when one is used to bootstrap the other.

## Further reading

- [Why ADK Go, not Genkit](why-adk-not-genkit.md) — the runtime choice.
- [Self-correcting loop](self-correcting-loop.md) — how Generator↔Review
  convergence works.
- [HITL design](hitl-design.md) — why approval is a CLI prompt.
- [ADRs](../adr/) — the full set of locked architectural decisions.
