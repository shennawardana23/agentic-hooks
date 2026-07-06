---
name: second-brain-authoring
description: How to add or edit a knowledge concept in agentic-hooks's Second Brain (knowledge/*.md). Use whenever the user wants to add a new coding principle, convention, or lesson-learned to the knowledge base, or asks why a concept isn't being matched by the Review agent or the MCP server's list_knowledge/get_knowledge tools. Make sure to use this skill whenever the user mentions "knowledge base", "second brain", "add a principle", or edits files under knowledge/, even if they don't say "skill" or name this file explicitly.
metadata:
  spec: agentskills.io/specification
  project: agentic-hooks
---

# Second Brain authoring

Every file under `knowledge/` is one concept, in OKF (Open Knowledge Format)
terms. The file's path minus `.md` *is* its identifier — there is no
separate `id` field anywhere. `internal/secondbrain.Load` walks this
directory at startup; both the Review agent and the `serve` MCP server read
from the same loaded set.

## Required shape

```markdown
---
type: principle
title: <short, human-readable name>
description: <one sentence — this shows up in list_knowledge summaries>
tags: [tag-one, tag-two]
timestamp: YYYY-MM-DD
---

<freeform Markdown body — the actual guidance>
```

- `type`, `title`, `tags`, `timestamp` are real OKF fields — don't invent
  new frontmatter keys without checking `internal/secondbrain/secondbrain.go`
  first, the parser only reads what it's told to read.
- Directory layout mirrors topic, e.g. `knowledge/go/error-handling.md`,
  `knowledge/go-patterns/table-driven-tests.md`. Follow the existing
  per-topic subdirectory convention (`go/`, `go-adk/`, `go-patterns/`,
  `go-genkit/`, `php-codeigniter/`, `vanilla-js/`) rather than inventing a
  new grouping style — check what's already there before adding a new
  subdirectory.

## Why title/tags matter more than the body

The Review agent's matching (`matchConceptsInDiff` in
`internal/agent/review.go`) works by checking whether a concept's `title`
or any `tags` entry appears as a **substring** in the code being reviewed —
not semantic search, not full-text search of the body. A concept with a
vague title like "Best Practices" will almost never match anything real; a
concept whose title or tags actually contain the identifier/keyword a
reviewer would look for (`error-handling`, `goroutine`, `interface`) will.
When authoring a new concept, pick `title`/`tags` values that a real diff
touching this topic would plausibly contain literally.

## After adding a concept

- `go build ./... && go test ./...` — a malformed frontmatter file is
  skipped at load time (logged, not fatal), so a typo won't crash
  anything, but it also won't be usable until fixed. Confirm the new file
  actually loads.
- If you want to confirm it's queryable via MCP, use the
  `mcp-inspector-tester` agent against `list_knowledge`/`get_knowledge`
  rather than guessing from the file alone.
- The `second-brain-reviewer` agent can be pointed at a real diff right
  after adding a concept, to confirm the new title/tags actually get
  matched the way you intended.
