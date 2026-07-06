# How to test with MCP Inspector

MCP Inspector CLI lets you drive the `serve` subcommand's MCP server
directly, without a full agent host like Claude Desktop or Claude Code
attached. The commands below were run against this project's real binary
and knowledge base and confirmed working.

## Prerequisites

```bash
make build
```

Run all commands below from the repo root, so the relative `bin/agentic-hooks`
and `knowledge` paths resolve.

## List available tools

```bash
npx -y @modelcontextprotocol/inspector --cli bin/agentic-hooks serve --knowledge-dir knowledge --policy-file POLICY.md --method tools/list
```

Expected: `get_agent_policy`, `list_knowledge`, and `get_knowledge` all
appear, each with its JSON input schema (`get_agent_policy` takes no
arguments; `list_knowledge` takes optional `type`/`tag`; `get_knowledge`
takes a required `id`).

## Call `get_agent_policy`

```bash
npx -y @modelcontextprotocol/inspector --cli bin/agentic-hooks serve --knowledge-dir knowledge --policy-file POLICY.md --method tools/call --tool-name get_agent_policy
```

Expected: the full text of `POLICY.md` comes back in `content`. Re-verified
live this session (2026-07-06, v7) against the real binary and real
`POLICY.md`.

## Call `list_knowledge` with no filter

```bash
npx -y @modelcontextprotocol/inspector --cli bin/agentic-hooks serve --knowledge-dir knowledge --method tools/call --tool-name list_knowledge
```

Expected: every concept under `knowledge/` comes back as a JSON array of
`{id, type, title, description, tags}` objects.

## Call `list_knowledge` with a tag filter

```bash
npx -y @modelcontextprotocol/inspector --cli bin/agentic-hooks serve --knowledge-dir knowledge --method tools/call --tool-name list_knowledge --tool-arg tag=go
```

Expected: only concepts tagged `go` come back — a strict subset of the
unfiltered list.

## Call `get_knowledge` with a valid id

```bash
npx -y @modelcontextprotocol/inspector --cli bin/agentic-hooks serve --knowledge-dir knowledge --method tools/call --tool-name get_knowledge --tool-arg id=go/error-handling
```

Expected: the full `id`, `title`, and `body` of that one concept file.

## Call `get_knowledge` with a bogus id

```bash
npx -y @modelcontextprotocol/inspector --cli bin/agentic-hooks serve --knowledge-dir knowledge --method tools/call --tool-name get_knowledge --tool-arg id=nonexistent/does-not-exist
```

Expected: a clean structured error (`isError: true`, message
`mcpserver: secondbrain: no concept with id "nonexistent/does-not-exist"`),
not a crash or a hang.

## Alternative: the project's own subagent

If you're working inside Claude Code on this repo, the `mcp-inspector-tester`
subagent runs the same kind of checks for you and reports back a summary,
rather than you running each Inspector command by hand.
