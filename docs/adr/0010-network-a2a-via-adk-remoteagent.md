# ADR-0010: Network A2A via ADK's `remoteagent/v2` primitives

## Status
Accepted

## Context
ADR-0001 recorded "no network-based A2A this iteration" as a consequence
of choosing ADK Go v2 as sole orchestrator — at the time, in-process
delegation covered every sub-agent this project had (Search, Generator,
Review). The user has since asked for "Intelligent Communication Through
Tool Calling": the root agent should be able to see other agents' public
manifests and decide, at runtime, whether to delegate to them — instead of
a fixed orchestration graph. This is a deliberate, user-confirmed
reopening of that consequence, not a silent scope change (see
`SESSION_HANDOFF.md`'s 2026-07-06 "Answered" entry and
`docs/superpowers/specs/2026-07-06-agent-to-agent-communication-design.md`
for the full design).

ADK Go v2 (this project's exact pinned dependency, `v2.0.0`) already ships
non-deprecated network A2A primitives: `agent/remoteagent/v2.NewA2A` wraps
a real A2A-protocol remote agent (via `github.com/a2aproject/a2a-go/v2`) as
a normal `agent.Agent`; `tool/agenttool.New` wraps any `agent.Agent` as a
callable `tool.Tool`. No hand-rolled A2A client is needed.

## Decision
Add network A2A support as an **additive, opt-in** capability: a static
YAML registry (`--agents-config`, no default) is loaded into
`agenttool`-wrapped `remoteagent.NewA2A` tools and merged into the root
agent's `Tools`. In-process delegation (Search, Generator/Review loop via
`SubAgents`) is completely untouched — this is a second, independent
delegation mechanism (tool-calling to remote agents), not a replacement.

## Consequences
- ADR-0001's "no network-based A2A this iteration" consequence is
  superseded by this ADR for the tool-calling delegation path specifically.
  In-process `SubAgents` delegation (Search, loop) is unaffected and remains
  the primary orchestration mechanism.
- A new dependency, `github.com/a2aproject/a2a-go/v2`, enters `go.mod`
  (already an indirect dependency of `google.golang.org/adk/v2`; this pins
  it as direct once `internal/agent/agentcomm.go` imports `remoteagent/v2`).
- No discovery-of-unknown-agents mechanism exists or is added — the
  registry is a static, explicitly-configured YAML file. Wiring a specific
  real remote agent (e.g. a sibling repo's A2A server) in as a live
  registry entry is out of scope for this decision; see the design spec's
  "Explicitly out of scope" section.
