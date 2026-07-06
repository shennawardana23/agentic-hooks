# AGENTS.md — Start Here for Any Agent

This is the entry point for any coding agent (Claude Code, Cursor, or this
project's own Search/Review agents) picking up `agentic-hooks` cold. Read
this before touching code.

## What this project is

A single Go binary, `agentic-hooks`, with two jobs:

1. `run <task>` — an ADK Go v2 multi-agent pipeline: a Search sub-agent
   (MCP client) feeds a self-correcting Generator↔Review loop, gated by a
   human approve/reject checkpoint (HITL).
2. `serve` — exposes the same "Second Brain" knowledge base as an MCP
   server over stdio, for external agent hosts to query directly.

Full architecture: [SYSTEM.md](SYSTEM.md). Product framing and personas:
[PRODUCT.md](PRODUCT.md). Durable decisions: [MEMORY.md](MEMORY.md).
Standing principles: [DESIGN.md](DESIGN.md). Forward-looking roadmap:
[PLAN.md](PLAN.md). How the knowledge base itself works:
[SKILL.md](SKILL.md). Dated changelog: [APPEND_SYSTEM.md](APPEND_SYSTEM.md).
Cross-session narrative handoff: [SESSION_HANDOFF.md](SESSION_HANDOFF.md).

## Build, test, run

```bash
make build   # bin/agentic-hooks, with version/build-time ldflags
make test    # go test ./... -v
make vet     # go vet ./...
make check   # vet + test + build — run this before considering any change done
make bench   # Go-native benchmarks, no API cost
make eval    # golden-set eval against the real model — costs API calls, opt-in only
make tidy    # go mod tidy
make clean   # remove bin/
```

```bash
make server                                   # start the MCP server over stdio
make dev TASK="review: func Foo() {...}"      # run the Search+Review loop on a task
make dev FILE=path/to/code.go                 # same, but review a real file
```

`GEMINI_API_KEY` (falls back to `GOOGLE_API_KEY`) is required for `make dev`
and `make eval` only — not for `build`/`test`/`vet`/`serve`.

Full command reference (every flag, every subcommand): [docs/reference/cli.md](docs/reference/cli.md).
Makefile target reference: [docs/reference/makefile-targets.md](docs/reference/makefile-targets.md).

## Repo conventions

- **Work on `main` directly.** No feature-branch requirement is enforced
  for this repo (personal/internal project).
- **Never `git commit` unless explicitly asked in the current message.**
  This is a standing instruction, not a one-time preference — re-confirm
  it hasn't changed if resuming an old session, but default to not
  committing.
- **Naming**: Go idiomatic (`camelCase` unqualified, `PascalCase`
  exported), package-per-concern under `internal/` (`agent`, `mcpserver`,
  `secondbrain`, `feedback`) — don't introduce a new top-level package
  without checking whether an existing one already owns that
  responsibility.
- **No mocks/scaffolding beyond what's needed.** Second Brain fixtures for
  tests are real Markdown files in test-local temp directories, not fake
  in-memory doubles layered on top of the real loader.
- **Verify library claims before writing them down**, especially for
  `google.golang.org/adk/v2` (bleeding-edge, thin public docs). Use
  `go doc` or the `adk-api-verifier` subagent — see
  [.claude/skills/adk-v2-verification/SKILL.md](.claude/skills/adk-v2-verification/SKILL.md).
  This project's own history has real bugs caused by skipping this step
  (see [MEMORY.md](MEMORY.md)).
- **Run `make check` before calling any change done.**

## Where everything lives

| Path | What's there |
|---|---|
| `POLICY.md` | Agent collaboration policy index — read before this file or `knowledge/` |
| `CLAUDE.md` | Auto-loaded gate summarizing the highest-severity policies |
| `policies/` | The 10 individual agent collaboration policy documents |
| `cmd/agentic-hooks/` | CLI entrypoint — `run`, `serve`, `version` subcommands |
| `internal/agent/` | Generator, Review, Search sub-agents + the self-correcting loop |
| `internal/mcpserver/` | MCP server exposing `list_knowledge`/`get_knowledge` |
| `internal/secondbrain/` | Markdown knowledge-base loader (OKF frontmatter) |
| `internal/feedback/` | Append-only HITL decision audit log (`feedback/feedback.jsonl`) |
| `knowledge/` | The Second Brain content itself — see [SKILL.md](SKILL.md) |
| `docs/tutorials/` | Diátaxis: learning-oriented, linear, assumes nothing |
| `docs/how-to/` | Diátaxis: task-oriented "how do I ___" guides |
| `docs/reference/` | Diátaxis: dry, exhaustive, structural facts only |
| `docs/explanation/` | Diátaxis: the "why" behind design choices |
| `docs/adr/` | Architecture Decision Records — immutable once written |
| `docs/diagrams/` | Mermaid + D2 source diagrams |
| `docs/superpowers/` | Original design specs and implementation plans (historical) |
| `.claude/agents/` | Project-scoped subagents |
| `.claude/skills/` | Project-scoped skills |
| `llms.txt` / `llms-full.txt` | Machine-readable doc index, per [llmstxt.org](https://llmstxt.org/) |

## Known limitations (don't rediscover these)

- No semantic/vector search — `secondbrain.Brain.Query` is substring/tag
  match only.
- No network A2A — sub-agents are in-process ADK delegation only.
- No tree-sitter structural analysis — Review is LLM-over-diff-text only
  (`StructuralFacts` seam reserved, unused).
- Single-provider model only — no retry/failover across LLM providers.
- `serve` is stdio-only, no HTTP transport.

Full known-limitations list with rationale: [README.md](README.md#️-known-limitations).
