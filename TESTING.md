# Testing Guide — agentic-hooks

How to verify this project actually works, from fastest/cheapest checks to
a full end-to-end run.

New to the project? [docs/tutorials/first-run.md](docs/tutorials/first-run.md)
walks the same ground as tier 2/3 below as a guided first run. For driving
the MCP server manually with MCP Inspector step by step, see
[docs/how-to/test-with-mcp-inspector.md](docs/how-to/test-with-mcp-inspector.md).

## 1. Automated test suite (no API key needed)

```bash
go build ./...
go vet ./...
go test ./... -v
```

Expected: all packages build clean, `go vet` reports nothing, and 13 tests
pass across `cmd/agentic-hooks`, `internal/secondbrain`, `internal/mcpserver`,
`internal/agent`. One of these — `TestServeStdio_RealBinaryOverStdio` in
`internal/mcpserver` — builds the actual `agentic-hooks` binary and drives
it as a real MCP server over stdio using the real MCP Go SDK client. This
is not a mock: it's the same wire protocol an external agent host (Claude
Code, Cursor, etc.) would use, exercised end-to-end.

## 2. Manual `serve` check with a real Go binary (no API key needed)

This confirms the MCP server works standalone, independent of the ADK/model
half of the project.

```bash
# Build the real binary
go build -o /tmp/agentic-hooks-bin ./cmd/agentic-hooks

# Create a minimal Second Brain
mkdir -p /tmp/agentic-hooks-knowledge/solid
cat > /tmp/agentic-hooks-knowledge/solid/single-responsibility.md <<'EOF'
---
type: principle
title: Single Responsibility Principle
tags: [solid]
---

A component should have one reason to change.
EOF

# Run it — it will sit waiting for MCP JSON-RPC messages on stdin
/tmp/agentic-hooks-bin serve --knowledge-dir /tmp/agentic-hooks-knowledge
```

It won't print anything and won't exit — that's correct, it's an MCP
server waiting on stdio, not an interactive CLI. Press Ctrl-C to stop it.
To actually see it respond, use the automated integration test above (it
drives this same binary with a real client), or connect an MCP-capable
tool (Claude Code, MCP Inspector, etc.) to it via stdio.

## 3. Full end-to-end `run` (needs GEMINI_API_KEY)

This exercises the real ADK agent pipeline: root → Search (MCP client) +
Review (Second Brain) → HITL approval prompt.

```bash
export GEMINI_API_KEY="your-real-key"

go build -o /tmp/agentic-hooks-bin ./cmd/agentic-hooks

# Reuse the same knowledge dir from step 2. The Search agent's MCP server
# doesn't have to be a different tool — pointing it at our own `serve`
# subcommand is a valid stand-in for this smoke test (it calls
# list_knowledge/get_knowledge instead of real web search, but proves the
# MCP-client wiring end-to-end).
/tmp/agentic-hooks-bin run \
  "review: func DoEverything() { validates input, writes to disk, sends email, all violates solid }" \
  --knowledge-dir /tmp/agentic-hooks-knowledge \
  --search-mcp-server /tmp/agentic-hooks-bin \
  --search-mcp-server-args "serve,--knowledge-dir,/tmp/agentic-hooks-knowledge"
```

Expected: the CLI prints a verdict (APPROVE or CHANGES_REQUESTED, citing
the Single Responsibility Principle), then prompts `Approve? [y/N]:`.
Answering `y` prints the verdict again as final output; anything else
prints "Rejected — no output returned as final." and nothing else.

**If it fails**, the error will point at one of two places:
- **Model/API errors** (auth, quota, network) — from `newDefaultModel` in
  `cmd/agentic-hooks/run.go`. Confirm `GEMINI_API_KEY` is set and valid.
- **Agent run errors** — from `runRootAgent` in the same file. This is the
  part of the plan that was flagged as needing real-world verification
  (see `docs/superpowers/plans/2026-07-02-second-brain-orchestration.md`
  Task 7) — if the ADK Go API has moved since this was written, this is
  where it will surface. Run `go doc google.golang.org/adk/v2/runner` to
  check the `Runner`/`Run` signature still matches.

## What's not tested here

- No live-model regression testing (nothing asserts on actual LLM output
  content beyond structural checks like "did the CLI print a prompt").
- No test for the HITL reject path against a real model run — the unit
  tests cover the CLI's approve/reject string logic, not a full live
  rejected verdict.
- Provider fallback (single Gemini key only, no retry/failover chain) is
  explicitly out of scope for this iteration — see design spec §7.
