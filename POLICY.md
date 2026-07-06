# Agent Collaboration Policy

> This is the guard entry point for `agentic-hooks`. Read this before the
> Second Brain (`knowledge/`) — this file governs *how* any agent
> collaborates with the maintainer; the Second Brain governs *what* good
> code looks like. They are separate concerns, cross-linked, not merged.

## What this is

~100 policies across 10 categories, covering agent conduct, session and
context handling, memory, caching, per-project persona adaptation,
guardrails, security, environment/secrets protection, anti-cheating and
honesty, communication style, and escalation/human-in-the-loop defaults.
Full detail lives in [`policies/`](policies/); this file is the index.

## How this gets enforced (honestly stated)

Three layers, each real but with different guarantees:

1. **[`CLAUDE.md`](CLAUDE.md) auto-load** — genuinely enforced for
   Claude-Code-family agents. Claude Code loads a repo's `CLAUDE.md` into
   context automatically before any tool call.
2. **`get_agent_policy` MCP tool** (`agentic-hooks serve`) — advisory only.
   MCP has no mechanism to force a client to call this before
   `list_knowledge`/`get_knowledge`. Its tool description asks callers to
   use it first; nothing blocks a client that doesn't.
3. **`.claude/settings.json` hooks** — real, automatic blocking for the
   narrow case they cover (env-dump-shaped Bash commands via
   `PreToolUse`), plus automatic context injection at session start
   (`SessionStart`). Hooks don't cover most of the ~100 policies — most are
   behavioral guidance for an LLM to follow, not machine-checkable.

No layer here claims to force compliance with all ~100 policies. That
honesty is itself policy [10.1](policies/10-escalation-hitl.md).

## Categories

| # | Category | Covers |
|---|---|---|
| 01 | [Agent Conduct](policies/01-agent-conduct.md) | Instruction-following, scope discipline |
| 02 | [Session, Context & Compaction](policies/02-session-context-compaction.md) | What survives compaction, when to checkpoint |
| 03 | [Memory & Persistence](policies/03-memory-persistence.md) | Durable facts vs. session-scoped state |
| 04 | [Caching](policies/04-caching.md) | What's safe to reuse vs. must re-verify |
| 05 | [Persona by Project](policies/05-persona-by-project.md) | Adapting to the project actually at hand |
| 06 | [Guardrails & Security](policies/06-guardrails-security.md) | General security posture |
| 07 | [Env & Secrets Protection](policies/07-env-secrets-protection.md) | Never exposing or injecting environment data |
| 08 | [Anti-Cheating & Honesty](policies/08-anti-cheating-honesty.md) | No gaming tests/metrics, no fabricated verification |
| 09 | [Communication Style](policies/09-communication-style.md) | Matching the user's actual signal |
| 10 | [Escalation & HITL](policies/10-escalation-hitl.md) | When to stop and ask a human |

## Relationship to the rest of this repo

- [`AGENTS.md`](AGENTS.md) — build/test/run commands and this repo's own
  coding conventions. Read after this file, not instead of it.
- [`knowledge/`](knowledge/) — the Second Brain itself: coding principles,
  queried via `list_knowledge`/`get_knowledge`.
- [`docs/superpowers/specs/2026-07-06-agent-policy-layer-design.md`](docs/superpowers/specs/2026-07-06-agent-policy-layer-design.md) —
  the full design rationale behind this file and its enforcement layers.
