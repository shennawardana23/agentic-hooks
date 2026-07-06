---
name: mcp-server-development
description: Add or modify a tool on agentic-hooks's MCP server (internal/mcpserver), which exposes the Second Brain over stdio via the modelcontextprotocol/go-sdk. Use this skill whenever the user wants to expose new data or a new capability to external agent hosts (Claude Code, Cursor) through this project's serve subcommand, add a new MCP tool, change list_knowledge/get_knowledge's schema, or debug why an MCP client isn't seeing what it expects — even if they just say "expose X to Claude Code" without mentioning MCP by name.
compatibility: Requires Go 1.25+ and github.com/modelcontextprotocol/go-sdk (already in go.mod).
metadata:
  spec: agentskills.io/specification
  project: agentic-hooks
---

# MCP server development

`internal/mcpserver/server.go` builds an `mcp.Server`, registers tools with
`mcp.AddTool`, and is driven over stdio by the `serve` subcommand. Two
tools exist today: `list_knowledge` (optional tag/type filter) and
`get_knowledge` (by id). Both are thin wrappers over
`internal/secondbrain`, direct Go function calls, no extra protocol layer
in between.

## Adding a new tool

1. Define the input struct as a plain Go struct with JSON tags — the SDK
   derives the JSON schema from it, don't hand-write a schema.
2. Write the handler with the same shape as the existing two:
   `func fooHandler(brain *secondbrain.Brain) mcp.ToolHandlerFor[FooInput, FooOutput]`.
   Keep it a closure over `*secondbrain.Brain` (or whatever backing data it
   needs) rather than a package-level global — this is the pattern both
   existing handlers use and it's what makes them independently testable.
3. Register it in the server setup alongside the existing `mcp.AddTool`
   calls.
4. Write a handler-level unit test in `server_test.go` calling the handler
   directly (see the existing pattern:
   `fooHandler(brain)(context.Background(), nil, FooInput{...})`) — this
   is cheap and doesn't need a real client/server pair.
5. If the change is significant, also exercise it through the real
   `TestServeStdio_RealBinaryOverStdio`-style integration test (builds the
   actual binary, drives it with a real MCP Go SDK client over real
   stdio) — or, for a quick manual check without writing a test, use the
   `mcp-inspector-tester` agent.

## Gotchas

- `serve` prints nothing and doesn't exit on its own — it's blocking on
  stdio waiting for JSON-RPC messages. That's correct behavior, not a
  hang. Don't "fix" this by adding output.
- No HITL gate on this path, by design — the MCP server only answers read
  queries, it never produces a verdict or takes an action. If a new tool
  would *write* something or make a decision, that's a design question to
  raise before building it, not something to add silently to this
  read-only server.
- Malformed frontmatter in one `knowledge/*.md` file is skipped at load
  time (logged, not fatal) — a new tool that assumes every concept loaded
  successfully will be wrong on a partially-broken knowledge directory.
- This project deliberately uses one MCP SDK
  (`modelcontextprotocol/go-sdk`) for both server and client roles (Search
  sub-agent). Don't introduce a second MCP library for a new tool even if
  it looks more convenient for one specific case.

## What NOT to do

- Don't add semantic/vector search to `list_knowledge`/`get_knowledge` —
  `internal/secondbrain.Query` is substring/tag match by design this
  iteration (see the design spec's deferred items).
- Don't skip the handler-level unit test because "it's just a thin
  wrapper" — both existing tools have one; a new tool without one is an
  inconsistency the next person has to explain.
