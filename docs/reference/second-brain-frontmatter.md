# Second Brain Frontmatter Reference

Diátaxis quadrant: **Reference**. Exact schema for `knowledge/**/*.md`
files, as parsed by `internal/secondbrain.parseConcept`. For the "why
Markdown files, not a database" rationale, see
[ADR 0004](../adr/0004-second-brain-as-markdown-not-database.md).

## File format

Each concept file is one Markdown file with a YAML frontmatter block
delimited by `---\n` lines, followed by a freeform Markdown body:

```markdown
---
type: principle
title: Wrap and propagate errors, never discard them
description: Go has no exceptions — a discarded error is a silently lost failure.
tags: [go, error-handling]
timestamp: 2026-07-04
---

Always check the error return of a call that can fail. ...
```

This is a real file: `knowledge/go/error-handling.md`.

A file with fewer than 3 parts after splitting on `---\n` (i.e. missing the
opening or closing delimiter) fails to parse and is **skipped, not fatal**
— logged via `log.Printf("secondbrain: skipping %s: %v", ...)` at load
time. One malformed file does not take down the rest of the knowledge base.

## Frontmatter fields

| YAML key | Go field (`secondbrain.Concept`) | Type | Required | Notes |
|---|---|---|---|---|
| `type` | `Type` | string | **yes** | Free-form string, e.g. `principle`, `pattern`. Load fails this one file (skip, not fatal) if empty or missing — the only field enforced at parse time. |
| `title` | `Title` | string | no | Human-readable concept name. Returned by both `list_knowledge` and `get_knowledge`. |
| `description` | `Description` | string | no | One-line summary. Returned by `list_knowledge` only (not `get_knowledge` — see [`mcp-tools.md`](mcp-tools.md)). Also used, together with `Title`/`Body`/`Tags`, by `Brain.Query`'s substring match. |
| `resource` | `Resource` | string | no | A single reference URL (e.g. `resource: https://genkit.dev`, `resource: https://github.com/google/adk-go`). Used in 7 of the 24 files under `knowledge/` as of this writing, concentrated in `go-adk/` and `go-genkit/` — not yet a convention followed everywhere. **Not exposed by either MCP tool** (`list_knowledge`/`get_knowledge` both omit it) — a documentation gap worth knowing about if you rely on it; the only way to surface it today is to also put the URL in the Markdown body. |
| `tags` | `Tags` | list of strings | no | YAML flow-sequence, e.g. `[go, error-handling]`. Matched exactly (case-sensitive, no partial match) by `list_knowledge`'s `tag` filter and `Brain.List`'s `containsTag`. |
| `timestamp` | `Timestamp` | string | no | Free-form string in practice written as `YYYY-MM-DD` (e.g. `2026-07-04`) across every file in this repo's `knowledge/` directory — not validated or parsed as a date by `secondbrain.go`, just stored and passed through as a string. Not exposed by either MCP tool. |

## Identity

**No separate `id` field.** The concept's identifier is its file path,
relative to `--knowledge-dir`, with the `.md` extension stripped and
converted to forward slashes (`filepath.ToSlash`). Example:
`knowledge/go-patterns/context-propagation.md` → id
`go-patterns/context-propagation`. This is deliberate — see
[ADR 0004](../adr/0004-second-brain-as-markdown-not-database.md) for why
this follows the Open Knowledge Format (OKF) convention rather than
introducing a redundant explicit id field.

## Body

Everything after the closing `---\n` delimiter, trimmed of leading/trailing
whitespace (`strings.TrimSpace`). No further structure is enforced —
`secondbrain.go` does not parse or validate the body at all. Most files in
this repo's `knowledge/` use prose plus a fenced code block contrasting a
"Bad" and "Good" example, but that's an authoring convention (see
[`docs/how-to/add-a-concept.md`](../how-to/add-a-concept.md)), not
something the loader checks.

## How fields are used

| Consumer | Fields read |
|---|---|
| `Brain.List(typeFilter, tagFilter)` | `Type`, `Tags` |
| `Brain.Get(id)` | `ID` (path-derived) — returns the whole `Concept` |
| `Brain.Query(topic)` | `Title`, `Description`, `Body`, `Tags` — lower-cased substring match across the concatenation of all four; no semantic/vector search |
| `list_knowledge` MCP tool | `ID`, `Type`, `Title`, `Description`, `Tags` (not `Resource`/`Timestamp`) |
| `get_knowledge` MCP tool | `ID`, `Title`, `Body` (not `Description`/`Tags`/`Resource`/`Timestamp`) |
| `internal/agent` Review sub-agent | Matches concept `Title`/tag substrings against the diff text directly (see [`docs/explanation/self-correcting-loop.md`](../explanation/self-correcting-loop.md)) |

## Adding a new concept

See [`docs/how-to/add-a-concept.md`](../how-to/add-a-concept.md) for the
task-oriented steps and this project's own `second-brain-authoring` skill
(`.claude/skills/second-brain-authoring/`).
