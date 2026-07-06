# Agent Policy Layer — Design Spec v0.1

Status: approved, pending implementation
Date: 2026-07-06

## 1. Purpose

`agentic-hooks` today is a Second Brain (coding-principle knowledge base,
`knowledge/*.md`) exposed via CLI and MCP server. This spec adds a second,
distinct layer on top: a set of ~100 policies governing **how any agent
should collaborate with the user**, independent of any specific project's
coding principles. The goal, in the user's own framing: an agent connecting
to this project (or reading its root files) should encounter the
collaboration policy *before* it encounters the Second Brain — a guard
entry point, not just a knowledge base.

This is explicitly broader than "how to write Go for this repo." It covers
agent conduct, session/context/compaction behavior, memory/persistence,
caching, per-project persona adaptation, guardrails, security, environment-
variable protection, anti-cheating, communication style, and
escalation/HITL — collaboration policy, not coding policy.

## 2. Scope

**In scope:**
- `POLICY.md` (root) — index and rationale for the whole layer.
- `policies/` directory, 10 category files, ~10 numbered policies each
  (~100 total).
- `CLAUDE.md` (root, new file) — the auto-loaded gate for Claude-Code-family
  agents, pointing to `POLICY.md` as a mandatory read.
- New MCP tool `get_agent_policy` in `internal/mcpserver`, alongside the
  existing `list_knowledge`/`get_knowledge`.
- `.claude/settings.json` hooks: `SessionStart` (inject policy summary),
  `PreToolUse` (flag/block env-dump-shaped Bash commands).

**Non-goals (explicit, and why):**
- No generic cross-project distribution mechanism (e.g. auto-syncing
  `POLICY.md` into the user's other repos). This spec covers this repo
  holding both the general standard and its own local enforcement point,
  per the user's own confirmed answer ("Both — general standard + local
  enforcement point"); syncing elsewhere is a separate, unrequested feature.
- No attempt at hard MCP-protocol-level call-ordering enforcement. MCP has
  no session/auth state between stdio calls in this project's design, and
  a server cannot force a client to call `get_agent_policy` before
  `list_knowledge`/`get_knowledge`. Confirmed with the user directly during
  brainstorming — the three layers below (auto-load, MCP tool description,
  hooks) are the honest achievable set, not a promise of hard enforcement
  MCP can't back.
- No rewriting all ~100 policies into enforced Go runtime code. Most are
  behavioral rules for an LLM agent to follow, not machine-checkable
  invariants. Only the subset covered by hooks (env-dump pattern matching)
  gets real enforcement; the rest are auto-loaded/advisory.
- Second Brain content (`knowledge/*.md`) is untouched — policies and
  knowledge stay separate concerns, cross-linked from `POLICY.md` and
  `AGENTS.md`, not merged.

## 3. Structure

```
agentic-hooks/
├── POLICY.md                              # index: what this is, 3-layer enforcement, links to all 10
├── CLAUDE.md                              # NEW — auto-loaded gate for Claude-Code-family agents
├── policies/
│   ├── 01-agent-conduct.md                # instruction-following, scope discipline, no silent scope creep
│   ├── 02-session-context-compaction.md   # what to preserve across compaction, when to checkpoint
│   ├── 03-memory-persistence.md           # what's durable vs session-scoped, where facts live
│   ├── 04-caching.md                      # what's safe to cache/reuse vs must re-verify
│   ├── 05-persona-by-project.md           # detect project type/stack, adapt tone/depth accordingly
│   ├── 06-guardrails-security.md          # general security posture, OWASP-adjacent awareness
│   ├── 07-env-secrets-protection.md       # grounded in the real GEMINI_API_KEY-pasted-in-chat incident
│   ├── 08-anti-cheating-honesty.md        # no gaming tests/metrics, no fabricated verification claims
│   ├── 09-communication-style.md          # terse vs verbose, when to ask vs act, tone matching
│   └── 10-escalation-hitl.md              # when to stop and ask a human, fail-closed defaults
└── .claude/
    └── settings.json                      # NEW (or extended) — SessionStart + PreToolUse hooks
```

`internal/mcpserver/server.go` gains one new tool registration
(`get_agent_policy`), following the existing `list_knowledge`/`get_knowledge`
pattern exactly (per `.claude/skills/mcp-server-development/SKILL.md`).

## 4. Policy file format

Each of the 10 category files follows one consistent format so the set
reads as a system, not 10 unrelated documents:

```markdown
# Policy Category NN: <Name>

<1-2 sentence framing of what this category governs and why it's separate
from the others.>

## NN.1 — <Short imperative title>
**Rule:** <the policy, stated as a plain, checkable instruction>
**Why:** <rationale — a real incident/reason where one exists (e.g. the
GEMINI_API_KEY plaintext-paste incident for env protection), otherwise a
plainly-stated principle. No invented incidents.>
**Enforcement:** auto-loaded | mcp-advisory | hook-enforced | advisory-only
```

Numbering is `<category>.<item>` (e.g. `07.3`) so an individual policy can
be referenced unambiguously from anywhere (a hook message, a PR comment,
another doc) without re-stating its full text.

## 5. The three enforcement layers

1. **`CLAUDE.md` auto-load** — real, not advisory, for Claude-Code-family
   agents: this file is loaded into context automatically before any tool
   call, the same mechanism already visible working every turn of this very
   session via the user's own `~/.claude/CLAUDE.md`. Content: a short
   (~30-line) summary of the highest-severity rules (env/secrets,
   anti-cheating, escalation defaults) plus a mandatory pointer to
   `POLICY.md` for the full set, and a pointer to `AGENTS.md`/Second Brain
   for project-specific conventions. Kept deliberately short — the full
   detail lives in `POLICY.md`/`policies/`, this file's job is to guarantee
   the agent has *seen* that the policy layer exists.

2. **`get_agent_policy` MCP tool** — for any MCP client (Claude Code via
   its own MCP connection, Cursor, or any other host), not just ones that
   auto-load `CLAUDE.md`. Returns `POLICY.md`'s content. Its `Description`
   field explicitly instructs callers to invoke it before
   `list_knowledge`/`get_knowledge` — this is advisory only (§2's
   non-goal), stated as such in the tool description itself, not oversold.

3. **`.claude/settings.json` hooks** — the only layer with real, automatic
   *blocking* power, scoped narrowly:
   - `SessionStart`: emits a short policy summary into context
     automatically at the start of any Claude Code session in this repo
     (belt-and-suspenders with `CLAUDE.md` auto-load — redundant on purpose,
     since `SessionStart` fires even in contexts where `CLAUDE.md` handling
     might vary by harness version).
   - `PreToolUse` (Bash matcher): flags/blocks commands shaped like env
     dumps — `env`, `printenv`, `cat .env`, `echo $VAR`-style patterns —
     mirroring the GateGuard-style hook already active this session, scoped
     to this repo's own `.claude/settings.json` rather than relying on an
     external plugin. Warns (does not silently allow) rather than crashing
     the session on a false positive — a legitimate `echo $PATH` for
     debugging shouldn't be permanently blocked, just requires
     acknowledgment.

## 6. Data flow

**Claude Code agent opens this repo:**
1. `~/.claude/CLAUDE.md` (user's global) + this repo's `CLAUDE.md` both
   auto-load into context.
2. `SessionStart` hook fires, injects the policy summary.
3. Agent proceeds with both the collaboration policy and (via `CLAUDE.md`'s
   pointer) awareness that `AGENTS.md`/Second Brain exist for project
   specifics.
4. Any Bash call shaped like an env dump triggers the `PreToolUse` hook.

**External MCP client (Cursor, etc.) connects to `agentic-hooks serve`:**
1. `tools/list` shows `get_agent_policy`, `list_knowledge`, `get_knowledge`.
2. Tool descriptions nudge policy-first, but nothing blocks calling
   `list_knowledge` first — documented limitation, not silently glossed
   over anywhere in this system (README, `POLICY.md`, and this spec all
   say so identically).

## 7. Testing approach

- `internal/mcpserver`: new handler-level unit test for
  `get_agent_policy`, following the exact pattern of
  `TestGetKnowledge_ReturnsBodyForKnownID` — call the handler directly,
  assert it returns `POLICY.md`'s real content (read from disk in the
  test, not hardcoded, so the test doesn't drift from the actual file).
- Extend `TestServeStdio_RealBinaryOverStdio` (or add a sibling test) to
  confirm `get_agent_policy` is reachable over real stdio, not just at the
  handler level — matching this repo's existing "handler test +
  wire-protocol test" pairing for every MCP tool.
- Hooks: manually verified via the `mcp-inspector-tester`-style approach
  used earlier this session — no automated test harness exists for
  `.claude/settings.json` hooks in this repo, and building one is out of
  scope for this spec (hooks are configuration, not application code).
- Link check: extend the same relative-link-checker script used for the
  documentation overhaul (§9 of that spec) to cover `POLICY.md` and
  `policies/*.md` before calling this done.

## 8. Content authoring approach

Writing ~100 policies with real, non-invented rationale is a content task
similar in shape to the documentation overhaul completed earlier this
session (10 parallel category writers, each grounded in real project
history — `SESSION_HANDOFF.md`'s incidents, this session's audit findings,
and general industry-standard practice where no project-specific incident
exists — never fabricated incidents). Execution detail (how many
parallel writers, what each is handed) belongs in the implementation plan
(`writing-plans` skill, next step after this spec is approved), not this
design doc.

## 9. Deferred (explicit roadmap, not silently dropped)

- Cross-project distribution/sync of `POLICY.md` to the user's other repos.
- Hard MCP-level call-ordering enforcement, if MCP or this project's own
  transport ever gains session state that would make it feasible.
- Automated test harness for hook behavior (currently manual-verification
  only, consistent with how this repo already treats `.claude/settings.json`
  as configuration rather than application code).

## 10. Open items for user review

- None outstanding — scope, enforcement mechanism, and category breakdown
  were all resolved via clarifying questions before this spec was written.
