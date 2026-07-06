# ADR-0001: ADK Go v2 is the sole request-path orchestrator

## Status
Accepted

## Context
`agentic-hooks` needs a runtime to orchestrate the Search → Generator/Review
loop → HITL pipeline behind the `run` subcommand. Two Google Go frameworks
were candidates: Genkit and ADK Go v2. Both can make model calls, but they
serve different purposes: Genkit is single-flow/prototyping oriented, while
ADK Go v2 has a graph-based workflow engine, built-in sub-agent delegation,
native MCP client support (`adk/tool/mcptoolset`, built on
`modelcontextprotocol/go-sdk`), and first-class HITL confirmation
(`tool.Context.RequestConfirmation()` / `.ToolConfirmation()`, plus a
`RequireConfirmation` flag directly on `mcptoolset.Config`). Using both
frameworks for the same responsibility (model calls, orchestration) would
mean two tracing pipelines and two config surfaces for no capability
gained — confirmed against ADK Go's `model.LLM` interface, which technically
could wrap Genkit's `ai.Generate()`, but there was no reason to.

## Decision
ADK Go v2 (`google.golang.org/adk/v2`) is the sole orchestrator for the
request path (Search sub-agent, Generator/Review self-correcting loop, root
agent delegation). Genkit is not added to the request path.

## Consequences
- One orchestration framework, one tracing/config surface, no duplicated
  agent/session/tool model.
- Sub-agent delegation (Search, Generator, Review) is in-process ADK
  delegation only — no network-based A2A this iteration.
- Genkit is still useful elsewhere in the project (see ADR-0002) but never
  runs inline during a `run` invocation.