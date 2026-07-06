# ADR-0003: One MCP SDK, used as both client and server

## Status
Accepted

## Context
`agentic-hooks` needs MCP in two directions: as a client (the Search
sub-agent looking up external context via a configured MCP server) and as a
server (`serve` subcommand exposing the Second Brain to external agent
hosts). Two separate Go MCP libraries exist in the wider ecosystem:
`github.com/modelcontextprotocol/go-sdk` (the official SDK) and
`mark3labs/mcp-go` (used internally by Genkit's `GenkitMCPServer`). Using a
different library for each direction would mean maintaining two MCP wire
implementations for one project, with no protocol-level benefit.

## Decision
`github.com/modelcontextprotocol/go-sdk` is used for both roles: as a client
inside the ADK Search sub-agent (`mcptoolset.New()`, consuming external MCP
servers — configurable, not hardcoded to one vendor) and as a server in
`agentic-hooks serve` (`mcp.NewServer` + `mcp.AddTool`, exposing Second
Brain queries). Genkit's own MCP plugin (`GenkitMCPServer`, built on
`mark3labs/mcp-go`) is not used anywhere in this project.

## Consequences
- One MCP dependency in `go.mod`, one wire-protocol implementation to trust
  and test (including the real stdio integration test that builds the
  actual binary and drives it with this same SDK's client).
- The Search sub-agent's external MCP server is swappable via
  `--search-mcp-server`/`--search-mcp-server-args` without touching code,
  since nothing about a specific server is hardcoded.
- Ties the project to `modelcontextprotocol/go-sdk`'s API stability and
  release cadence for both roles — a breaking change there affects both the
  client and server code paths simultaneously.