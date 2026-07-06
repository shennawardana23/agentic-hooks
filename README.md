# agentic-hooks

Second Brain orchestration CLI, built on Go ADK v2. A Search agent and a
Review agent run in a loop against a Markdown-based "Second Brain" of
engineering principles, with a human-in-the-loop (HITL) approval gate before
any output is treated as final. The same Second Brain is also exposed as an
MCP server, so external agent hosts (Claude Code, Cursor, MCP Inspector,
etc.) can query it directly over stdio.

## How it fits together

- **Second Brain** (`internal/secondbrain`) — loads a directory of
  frontmatter-tagged Markdown files (`type`, `title`, `tags`) into an
  in-memory set of concepts.
- **MCP server** (`internal/mcpserver`) — serves the Second Brain over
  stdio via `list_knowledge` / `get_knowledge` tools, using the real MCP Go
  SDK.
- **Review agent** (`internal/agent/review.go`) — takes a diff, matches it
  against Second Brain concepts, and returns an `APPROVE` /
  `CHANGES_REQUESTED` verdict.
- **Generator agent** (`internal/agent/generator.go`) — drafts/revises an
  answer, incorporating `CHANGES_REQUESTED` feedback from the Review agent.
- **Search agent** (`internal/agent/search.go`) — an MCP client sub-agent
  that can query the Second Brain MCP server (or any other MCP server)
  for supporting context.
- **Root orchestration + HITL** (`cmd/agentic-hooks/run.go`) — wires the
  above into a loop, then prompts for human approval before printing a
  final transcript.
- **Feedback log** (`internal/feedback`) — every `run` (approved or
  rejected) is appended to `feedback/feedback.jsonl` as an audit trail.

## Requirements

- Go 1.25+
- `GEMINI_API_KEY` (falls back to `GOOGLE_API_KEY` if that's what you have
  set) — only needed for `make dev` and `make eval`. Not needed to build,
  vet, test, or run the MCP server standalone.

## Quick start

```bash
make build              # bin/agentic-hooks
make server             # start the MCP server over stdio
make dev TASK="review: func DoEverything() {...}"
```

`make dev` accepts three ways to supply a task:

```bash
make dev FILE=path/to/code.go        # review a real file
make dev TASK="review: ..."          # review inline text
make dev                             # prompts interactively
```

Run `make help` for the full target list.

## Testing

See [TESTING.md](TESTING.md) for the full guide (automated suite, manual
MCP server check, full end-to-end run, and MCP Inspector usage). Short
version:

```bash
go build ./...
go vet ./...
go test ./...
```

## Benchmarking and evaluation

- `make bench` — Go-native micro-benchmarks (`go test -bench`), no API
  cost. Covers Second Brain loading, review-prompt construction, and MCP
  handler performance.
- `make eval` — golden-set evaluation of the Review agent against a real
  Gemini model. Costs API calls, opt-in only (`AGENTIC_HOOKS_EVAL=1`),
  never runs as part of `make check`/`go test ./...`.

## Project layout

```
cmd/agentic-hooks/     CLI entrypoint (serve, run, version)
internal/agent/        Generator, Review, Search agents + eval harness
internal/mcpserver/    MCP server exposing the Second Brain
internal/secondbrain/  Markdown knowledge-base loader
internal/feedback/     Append-only run audit log
knowledge/             Second Brain content (Markdown, frontmatter-tagged)
docs/                  Design specs and plans
```
