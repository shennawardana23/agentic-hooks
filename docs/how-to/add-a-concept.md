# How to add a Second Brain concept

## Choose a location

Concepts live under `knowledge/<topic>/<slug>.md`. The file path, minus
the `.md` extension, is the concept's identifier — there is no separate
`id` field. Group by topic the same way the existing directories do
(`go/`, `go-adk/`, `go-genkit/`, `go-patterns/`, `php-codeigniter/`,
`vanilla-js/`); create a new topic directory if none of the existing ones
fit.

```bash
mkdir -p knowledge/<topic>
touch knowledge/<topic>/<slug>.md
```

## Write the frontmatter

Every concept file needs OKF-style YAML frontmatter followed by a
Markdown body:

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

Frontmatter fields:

| Field | Required | Notes |
|---|---|---|
| `type` | yes | Free-text category, e.g. `principle`, `pattern`. Used as a filter in `list_knowledge` and `Brain.List`. |
| `title` | yes | Human-readable name of the concept. |
| `description` | no | One-sentence summary, shown in `list_knowledge` output. |
| `tags` | no | YAML list, used as a filter in `list_knowledge` and `Brain.List`, and matched against diff text by the Review agent. |
| `timestamp` | no | Date the concept was written or last revised. |

See [the frontmatter reference](../reference/second-brain-frontmatter.md)
for the exhaustive field list and parsing behavior.

## Write the body

The body is freeform Markdown — the actual coding-principle content a
human or the Review agent reads. Code examples are welcome; keep them
short and focused on the one concept the file is about.

## Verify it loads

Malformed frontmatter in one file is skipped at load time (logged, not
fatal), so a typo will not crash the whole knowledge base — but it also
means a broken new file silently disappears from results unless you
check. Confirm your new concept loads:

```bash
make build
bin/agentic-hooks serve --knowledge-dir knowledge &
```

Then query it with MCP Inspector (see
[testing with MCP Inspector](test-with-mcp-inspector.md)):

```bash
npx -y @modelcontextprotocol/inspector --cli bin/agentic-hooks serve --knowledge-dir knowledge --method tools/call --tool-name get_knowledge --tool-arg id=<topic>/<slug>
```

Expected: your concept's `title` and `body` come back. If instead you get
a "no concept with id" error, double check the path matches the file
exactly (minus `.md`) and that the frontmatter is valid YAML.

## Run the automated tests

```bash
go test ./internal/secondbrain/... -v
```

This is the fastest way to catch a frontmatter parsing problem before it
reaches the MCP server or the Review agent.
