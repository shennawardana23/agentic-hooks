# PRODUCT.md — What This Tool Is For

`agentic-hooks` is a personal/internal tool, not a product being built for
external distribution (see [README.md License](README.md#-license)). This
file exists so anyone extending it — human or agent — makes changes that
serve its actual two users, rather than generic "best practice" additions
that don't serve either.

## Persona 1 — the CLI user running `run`

**Who**: a developer (currently, the project's sole user) who wants a
second opinion on a code change, grounded in a personal, curated set of
coding principles, before treating that opinion as final.

**What they do**: `make dev TASK="review: ..."` or `make dev
FILE=path/to/code.go`. They watch the Search/Generator/Review pipeline
stream its reasoning live (tagged `[author]`), read the final verdict, and
explicitly type `y` to approve or anything else to reject.

**What success looks like**:
- The verdict is grounded in a real matched Second Brain concept (title or
  tag literally present in the diff), not generic LLM opinion — see
  [SKILL.md](SKILL.md) for why title/tags drive matching.
- The loop converges (`APPROVE`) in fewer than `--max-iterations` passes
  for a diff that has an obvious principle violation and an obvious fix.
- Every run — approved or rejected — leaves a durable trace in
  `feedback/feedback.jsonl`, so the user (or a future offline eval step)
  can look back at what was decided and why.
- Nothing is ever presented as a final answer without the explicit `y`
  gate — the user can always trust that unapproved output was actually
  discarded, not silently used.

**What failure looks like**: a verdict that ignores the Second Brain
entirely (falls back to "no specific concept matched" too often because
concepts are titled too generically — an authoring problem, not a bug);
a loop that never converges and always exhausts `--max-iterations`; a
feedback log entry that's missing or malformed.

## Persona 2 — an MCP-consuming agent host

**Who**: an external coding agent host — Claude Code, Cursor, or any other
MCP-compatible client — that wants read access to the same curated
knowledge base, independent of this project's own review pipeline.

**What they do**: connect to `agentic-hooks serve` over stdio (config
example in [README.md](README.md#-add-to-your-mcp-client)), call
`list_knowledge` (optionally filtered by `type`/`tag`) to browse, then
`get_knowledge` by id to pull a concept's full body into their own
context.

**What success looks like**:
- `list_knowledge` returns accurate summaries (title, description, tags)
  fast enough to browse interactively, with no partial/stale results if
  one concept file is malformed (that file is skipped, not fatal).
- `get_knowledge` returns the exact same content a human would see
  opening the Markdown file directly — no lossy transformation.
- The connecting agent's own coding suggestions start reflecting this
  project's actual conventions (e.g. a suggestion respecting
  `internal/secondbrain`'s existing frontmatter schema instead of
  inventing a new one) because it queried the Second Brain rather than
  guessing.
- Zero data leaves the local machine on this path — `serve` only answers
  local read queries (see [README.md Privacy](README.md#-privacy)).

**What failure looks like**: the external host gets a tool-call error
instead of a clean "no such concept" response for a bad id (see
[docs/reference/mcp-tools.md](docs/reference/mcp-tools.md) for the actual
error contract); `list_knowledge` silently omitting concepts due to a
parse failure with no way for the caller to know that happened.

## What this project deliberately does not try to be

- Not a multi-user SaaS product — no auth, no multi-tenant knowledge
  bases, no hosted deployment story.
- Not a general-purpose code review bot — the Second Brain is this
  project's own curated, personal set of principles, not a generic
  best-practices database.
- Not trying to compete with semantic-search-backed review tools — the
  substring/tag matching in [SKILL.md](SKILL.md) is a deliberate,
  currently-sufficient simplicity trade-off, not an oversight.
