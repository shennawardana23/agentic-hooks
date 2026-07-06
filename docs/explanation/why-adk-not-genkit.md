# Why ADK Go v2, not Genkit, as the runtime

`agentic-hooks` is built on **ADK Go v2** (`google.golang.org/adk/v2`) as its
sole in-request orchestrator. Genkit — also a Google Go framework — is
deliberately not used anywhere in the request path. This document explains
why, since both frameworks could plausibly do the job and the choice isn't
obvious from the code alone.

## The decision

Both frameworks serve different purposes. Genkit is single-flow and
prototyping oriented. ADK Go v2 has a graph-based workflow engine, built-in
sub-agent delegation, native MCP client support
(`adk/tool/mcptoolset`, built on `modelcontextprotocol/go-sdk`), and
first-class human-in-the-loop confirmation
(`tool.Context.RequestConfirmation()` / `.ToolConfirmation()`, plus a
`RequireConfirmation` flag directly on `mcptoolset.Config`).

Using both frameworks for the same responsibility — model calls,
orchestration — would mean two tracing pipelines and two config surfaces for
no capability gained. This was confirmed against ADK Go's `model.LLM`
interface, which technically *could* wrap Genkit's `ai.Generate()`, but
there's no reason to: ADK already covers everything this project's request
path needs.

## Genkit's actual role: offline evaluation only

Genkit is not excluded from the project — it has one job, and it's
deliberately kept out of the request path. ADK session traces are exported
as a dataset (`{testCaseId, input, output, context, traceIds}`, Genkit's
documented "raw evaluation" format for externally-produced output) and
scored via `genkit eval:run` + `DefineEvaluator` (LLM-as-judge). This runs
**after** a session completes, never inside it.

In practice, this project's actual golden-set eval (`make eval`,
`internal/agent/eval_test.go`) is a plain Go harness rather than a wired-up
Genkit `eval:run` pipeline — see [ADR 0002](../adr/0002-offline-eval-as-plain-go-harness.md)
for why that superseded the original plan.

## One MCP SDK, both directions

A related decision, made for the same "don't run two pipelines for one
job" reason: `github.com/modelcontextprotocol/go-sdk` is used both as a
*client* (inside the Search sub-agent, via `mcptoolset`, consuming
externally configured MCP servers) and as a *server* (`agentic-hooks
serve`, exposing the Second Brain). Genkit ships its own MCP plugin
(`GenkitMCPServer`), but that sits on a different underlying library
(`mark3labs/mcp-go`) — introducing it anywhere in this project would mean a
second MCP implementation for no capability gained, the same anti-pattern
this whole decision is trying to avoid. See
[ADR 0003](../adr/0003-mcp-sdk-bidirectional-one-library.md).

## What this means in practice

- If you're extending the request path (Search, Generator, Review, Root),
  stay inside `google.golang.org/adk/v2`. Don't reach for a Genkit
  primitive because it looks more familiar or convenient for one call site.
- If you're extending the eval/scoring side, Genkit-flavored patterns (or
  the existing plain-Go harness) are the right place — see
  [`llm-agent-quality`](../../.claude/skills/llm-agent-quality/SKILL.md) for
  how the golden-set eval works today.
- Related decisions this reasoning locks in: [ADR 0001](../adr/0001-adk-go-v2-as-sole-orchestrator.md),
  [ADR 0002](../adr/0002-offline-eval-as-plain-go-harness.md),
  [ADR 0003](../adr/0003-mcp-sdk-bidirectional-one-library.md).
