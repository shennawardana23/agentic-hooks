# ADR-0004: Second Brain is OKF-frontmatter Markdown, not a database

## Status
Accepted

## Context
The Review agent and the MCP server both need to read a personal knowledge
base of coding principles. Options considered included a database (SQL or
embedded) with a schema for concepts, versus plain files. The knowledge base
is small, human-authored, and edited directly by the user rather than
through an application — a natural fit for the Open Knowledge Format (OKF),
a documented plaintext Markdown-plus-frontmatter convention
(`type`/`title`/`description`/`tags`/`timestamp` fields), rather than a
storage engine.

## Decision
The Second Brain is a directory of Markdown files (`knowledge/<topic>/<slug>.md`),
each with OKF frontmatter and a freeform Markdown body. The file's path,
minus the `.md` extension, is the concept's identifier — there is no
separate `id` field, per the OKF spec's own definition of identity. No
database, no ORM, no query language beyond `internal/secondbrain`'s
`List`/`Get`/`Query` functions (substring/tag match).

## Consequences
- Adding, editing, or removing a concept is a plain file operation — no
  migration, no schema versioning, editable in any text editor.
- One bad file (malformed frontmatter) is logged and skipped at load time,
  not fatal to the whole directory scan.
- No semantic/vector search, no relational queries across concepts — the
  current `Query` is substring/tag match only, an explicit and accepted
  limitation for this iteration, not an oversight.
- Both consumers (Review, direct function call; MCP server, wrapped as
  tools) share the exact same loader and data model, so there is no
  drift between how a human-editable file becomes visible to either path.