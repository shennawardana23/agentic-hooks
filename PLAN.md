# PLAN.md — Current Roadmap

This is the forward-looking, currently-active roadmap for `agentic-hooks`
as a whole. It is distinct from per-feature implementation plans under
[docs/superpowers/plans/](docs/superpowers/plans/), which are historical
records of how a specific already-built feature was planned and executed.
This file is updated as priorities change; those are not.

## Status as of 2026-07-06

Core MVP (both entry points) is implemented and passing its test suite:
`run` (Search + self-correcting Generator↔Review loop + HITL + feedback
log) and `serve` (MCP server exposing the Second Brain). See
[MEMORY.md](MEMORY.md) for the full list of locked decisions and verified
facts, and [SESSION_HANDOFF.md](SESSION_HANDOFF.md) for narrative history.

Nothing has been committed to git yet — that remains the user's explicit
decision to make (standing no-commit instruction, see
[AGENTS.md](AGENTS.md)).

## Open items needing the user's input before proceeding

1. **Which MCP server(s) the Search sub-agent points at by default.**
   Nothing is hardcoded on purpose (`--search-mcp-server` /
   `--search-mcp-server-args` are required flags), but local dev needs at
   least one configured. Not yet decided.
2. **Final sign-off** — the user has not given final sign-off on the
   original design spec or the implementation as a whole.
3. **`go.mod` Go version** — currently pins `go 1.25.0`. Re-check current
   Go stable before assuming this is still correct; don't trust a stale
   number carried forward from an old session.

## Pending decision — explicitly not started, asked twice already

The user has twice raised wanting a "massive audit" of two things:

- **Provider fallback** — a circuit-breaker/retry wrapper across LLM
  providers. A reference implementation exists at
  `archpublicwebsite-agentic/internal/model/failover`, but it's pinned to
  `google.golang.org/adk/model` (the **pre-v2** import path), not
  `google.golang.org/adk/v2/model` that this project uses —
  compatibility with this codebase is unconfirmed.
- **A Skills system** — referencing a sibling repo's `skills/` taxonomy
  (`go-adk-q`) plus external specs (genkit.dev's "Skills middleware",
  adk.dev/skills, agentskills.io) to reconcile against.

**Do not start this unprompted.** It is a large scope addition on top of
an already-complete MVP. Ask the user again, explicitly, before spending
any time on it — this instruction is carried forward verbatim from
[SESSION_HANDOFF.md](SESSION_HANDOFF.md) and should not be silently
dropped by whichever agent or developer picks this up next.

## Deferred roadmap (real, not abandoned)

These were explicitly scoped out of the MVP, each a candidate for its own
future design spec — see
[docs/superpowers/specs/2026-07-02-second-brain-orchestration-design.md §7](docs/superpowers/specs/2026-07-02-second-brain-orchestration-design.md)
for the original list and rationale:

- Frontend visual canvas (Vue + VueFlow + Pinia) to observe/edit runs.
- True network A2A (agent cards, remote agent servers).
- Additional agents (Fetch, Image-gen) beyond Search/Review.
- Semantic/vector search over the Second Brain (`Query` is
  substring/tag match only today).
- A wired-up offline eval loop (`internal/eval` is a package boundary
  only right now, not implemented) — see
  [docs/adr/0002](docs/adr/0002-offline-eval-as-plain-go-harness.md).
- Token-bucket/rate limiting, cost tracking, full observability/tracing
  dashboard.
- Tree-sitter-based structural/AST pre-filtering ahead of the Review
  agent — see [docs/adr/0007](docs/adr/0007-tree-sitter-deferred.md).
- Excalidraw / draw.io diagrams for this documentation set — revisit only
  if those MCP servers become connected in a future session (see
  [docs/superpowers/specs/2026-07-06-documentation-overhaul-design.md §10](docs/superpowers/specs/2026-07-06-documentation-overhaul-design.md)).
- A real historical ticket/code-review log in
  [APPEND_SYSTEM.md](APPEND_SYSTEM.md) — this repo has no issue tracker
  connected yet; the changelog section is seeded with real dated entries
  and fills in going forward, not backfilled with fabricated history.

## Non-goals (still true, re-confirm before reversing)

- No network-based A2A.
- No database, no RAG pipeline.
- No frontend.
- No live/inline Genkit involvement in the request path.
