# MCP Tools Reference

Diátaxis quadrant: **Reference**. Exhaustive schema documentation for the
three tools `agentic-hooks serve` exposes over stdio, defined in
`internal/mcpserver/server.go`. All three are read-only — no HITL gate, by
design (see [`docs/explanation/hitl-design.md`](../explanation/hitl-design.md)).

Server implementation identity (from `mcp.NewServer(&mcp.Implementation{...})`):

| Field | Value |
|---|---|
| Name | `agentic-hooks-secondbrain` |
| Version | `v0.1.0` |

## `list_knowledge`

Lists Second Brain concepts, optionally filtered by `type` and/or `tag`.
Both filters are ANDed when both are given (`internal/secondbrain.Brain.List`).

### Input schema (`ListKnowledgeInput`)

| Field | JSON key | Type | Required | Description |
|---|---|---|---|---|
| `Type` | `type` | string | no | Filter by exact concept type match, e.g. `principle`. |
| `Tag` | `tag` | string | no | Filter by exact tag match (one of the concept's `tags` list), e.g. `solid`. |

### Output schema (`ListKnowledgeOutput`)

| Field | JSON key | Type | Description |
|---|---|---|---|
| `Concepts` | `concepts` | array of `ConceptSummary` | Matching concepts. Empty array (not null/error) if nothing matches. |
| `Warnings` | `warnings` | array of string, omitted if empty | One `"path: reason"` entry per knowledge file `secondbrain.Load` skipped for malformed frontmatter — previously only visible in server-side `log.Printf` output, now surfaced to the caller. |

`ConceptSummary` fields: `id`, `type`, `title`, `description`, `tags`
(array of string). Note this is a **subset** of the underlying
`secondbrain.Concept` struct — `Resource` and `Timestamp` (both present on
`Concept` and in the frontmatter, see
[`second-brain-frontmatter.md`](second-brain-frontmatter.md)) are not
exposed through this tool. `get_knowledge` also does not expose them; the
full body is the only way to see resource/timestamp content today (they'd
appear only if included in the Markdown body itself).

### Example call

Request (tool arguments):
```json
{ "tag": "go" }
```

Response (`concepts`, truncated to one entry — real filtered call against
this repo's `knowledge/` on 2026-07-06 returned 16 matches for `tag=go`):
```json
{
  "concepts": [
    {
      "id": "go/error-handling",
      "type": "principle",
      "title": "Wrap and propagate errors, never discard them",
      "description": "Go has no exceptions — a discarded error is a silently lost failure.",
      "tags": ["go", "error-handling"]
    }
  ]
}
```

## `get_knowledge`

Returns one concept's full content by id.

### Input schema (`GetKnowledgeInput`)

| Field | JSON key | Type | Required | Description |
|---|---|---|---|---|
| `ID` | `id` | string | yes | The concept id — its file path under the knowledge directory, minus `.md`. |

### Output schema (`GetKnowledgeOutput`)

| Field | JSON key | Type | Description |
|---|---|---|---|
| `ID` | `id` | string | Echoes the requested id. |
| `Title` | `title` | string | Concept title from frontmatter. |
| `Body` | `body` | string | Full Markdown body (frontmatter stripped, trimmed). |

### Example call

Request:
```json
{ "id": "go/error-handling" }
```

Response (real content from `knowledge/go/error-handling.md`):
```json
{
  "id": "go/error-handling",
  "title": "Wrap and propagate errors, never discard them",
  "body": "Always check the error return of a call that can fail. Never assign it to `_` unless the call genuinely cannot fail for reasons documented at the call site (e.g. `Buffer.Write` on an in-memory buffer).\n\n..."
}
```

### Error case

Unknown id returns a structured MCP tool error, not a crash or hang.
Verified this session via MCP Inspector CLI against id
`nonexistent/does-not-exist`:
```json
{ "isError": true, "text": "mcpserver: secondbrain: no concept with id \"nonexistent/does-not-exist\"" }
```

## `get_agent_policy`

Returns the full content of the agent collaboration policy file
(`POLICY.md` by default; overridable via `serve --policy-file`). Advisory
only — see [`POLICY.md`](../../POLICY.md) itself for the honest statement
of what this tool can and can't enforce.

### Input schema (`GetAgentPolicyInput`)

Takes no arguments (`{}`).

### Output schema (`GetAgentPolicyOutput`)

| Field | JSON key | Type | Description |
|---|---|---|---|
| `Content` | `content` | string | The raw file contents of the configured policy file. |

### Example call

Request: `{}`

Response (live call against this repo's real `POLICY.md`, verified via MCP
Inspector CLI):
```json
{ "content": "# Agent Collaboration Policy\n\n> This is the guard entry point for `agentic-hooks`...\n" }
```

### Error case

A missing or unreadable policy file returns a structured MCP tool error
(`mcpserver: read policy file: ...`), not a crash.

## Connecting a client

See the root [`README.md`](../../README.md#-add-to-your-mcp-client) for the
`claude_desktop_config.json` / `.mcp.json` snippet, and
[`docs/how-to/test-with-mcp-inspector.md`](../how-to/test-with-mcp-inspector.md)
to drive these tools manually without a full client.
