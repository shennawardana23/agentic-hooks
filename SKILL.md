# SKILL.md — How the Second Brain Works

This file explains how Second Brain concepts are authored, stored, and
matched, for anyone (human or agent) who needs to add to or reason about
`knowledge/`. For the step-by-step authoring workflow and a Claude Code
skill that auto-triggers on relevant requests, see
[.claude/skills/second-brain-authoring/SKILL.md](.claude/skills/second-brain-authoring/SKILL.md) —
this file summarizes the same rules; that one is the operational skill
definition. They must not contradict each other.

## Storage format

Every file under `knowledge/` is one concept, in OKF (Open Knowledge
Format) terms. The file's path minus `.md` **is** its identifier — there
is no separate `id` field anywhere in the schema.
([docs/adr/0004](docs/adr/0004-second-brain-as-markdown-not-database.md)
records why this is Markdown and not a database.)

```markdown
---
type: principle
title: <short, human-readable name>
description: <one sentence — shown in list_knowledge summaries>
tags: [tag-one, tag-two]
timestamp: YYYY-MM-DD
---

<freeform Markdown body — the actual guidance>
```

`type`, `title`, `description`, `tags`, `timestamp` are the only
frontmatter fields the parser reads (`internal/secondbrain`). Directory
layout mirrors topic — `go/`, `go-adk/`, `go-genkit/`, `go-patterns/`,
`php-codeigniter/`, `vanilla-js/` are the existing subdirectories; follow
that convention rather than inventing a new grouping.

Full field-by-field schema reference:
[docs/reference/second-brain-frontmatter.md](docs/reference/second-brain-frontmatter.md).

## How loading and querying work

`internal/secondbrain.Load(dir)` walks the directory once at startup and
returns a `*Brain` with three read operations:

- `List(type, tag)` — optional filters, used by `list_knowledge` and by
  Review's own concept lookup.
- `Get(id)` — exact lookup by path-derived id, used by `get_knowledge`.
- `Query(topic)` — substring/tag match where a short topic string is
  checked against each concept (used for topic-driven lookup, not
  diff-matching — see below). No semantic/vector search this iteration.

Malformed frontmatter in one file is logged and skipped at load time, not
fatal to the whole directory scan — a typo breaks that one concept, not
the knowledge base.

## How the Review agent matches concepts — title/tags matter more than the body

`internal/agent/review.go`'s `matchConceptsInDiff` runs the **opposite**
direction from `Brain.Query`: instead of checking whether a topic string
appears inside a concept, it checks whether a concept's `title` or any
`tags` entry appears as a **substring inside the diff being reviewed**
(case-insensitive). This is not semantic matching — a concept titled
"Best Practices" will almost never match anything real, because that
phrase rarely appears literally in code. A concept titled
`error-handling` or tagged `goroutine` will match any diff that contains
that literal string.

**Practical consequence for authoring**: pick `title`/`tags` values that a
real diff touching this topic would plausibly contain literally, not
generic category names. This is the single most important authoring rule
— it determines whether a concept is ever actually surfaced during
review, regardless of how good the body content is.

## Two consumers, one loader

Both the Review sub-agent (direct Go function call, no protocol overhead)
and the MCP server's `list_knowledge`/`get_knowledge` tools
(`internal/mcpserver`) read from the same `*Brain` instance shape loaded
by `internal/secondbrain.Load`. There is exactly one loading/parsing
implementation; neither consumer has its own copy of this logic.

## After adding or editing a concept

1. `go build ./... && go test ./...` — confirms the new/edited file
   actually loads (a parse failure is silent-skip, not a crash, so this
   is the only way to know a typo didn't quietly disable the concept).
2. To confirm it's queryable via MCP without guessing: use the
   `mcp-inspector-tester` subagent against `list_knowledge`/`get_knowledge`,
   or follow [docs/how-to/test-with-mcp-inspector.md](docs/how-to/test-with-mcp-inspector.md).
3. To confirm the Review agent actually matches it on a real diff: point
   the `second-brain-reviewer` subagent at a real diff touching the new
   concept's topic.

See also: [docs/how-to/add-a-concept.md](docs/how-to/add-a-concept.md) for
the task-oriented walkthrough.
