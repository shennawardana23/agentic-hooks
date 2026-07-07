---
name: a2a-remote-agents
description: Wire a new remote agent into agentic-hooks's root agent as a callable tool over A2A. Use whenever the user wants to add/register a remote agent, edit a --agents-config registry YAML file, debug why an entry from that file didn't show up as a tool, or asks how the root agent delegates to another agent over the network (as opposed to the in-process Search/Review SubAgents). Trigger on "add an agent", "register agent", "agents-config", "registry.yaml", "A2A", "remote agent", "agent card", "well-known agent-card.json".
compatibility: Requires google.golang.org/adk/v2/agent/remoteagent/v2 and github.com/a2aproject/a2a-go/v2 (already in go.mod). The remote agent must serve a real A2A agent card.
metadata:
  spec: agentskills.io/specification
  project: agentic-hooks
---

# A2A remote agents

Additive delegation path, separate from the in-process Search/Review
`SubAgents` loop (that stays primary; see ADR-0010, which supersedes
ADR-0001's "no network A2A" consequence for this tool-calling path only).
A registry YAML file lists remote agents; each becomes a `tool.Tool` the
root agent can call like any other tool — the root agent decides
dynamically whether to invoke one, nothing is hardcoded into the
orchestration graph.

## Adding a remote agent

1. Add an entry to your `--agents-config` YAML file (a plain list, see
   `internal/agent.RegistryEntry`):
   ```yaml
   - name: pwdev
     description: what this agent does, shown to the root agent's model
     card_url: https://example.com
   ```
   `card_url` is the agent's **base URL**, not the well-known path itself —
   `remoteagent.NewAgentCardProvider` appends
   `/.well-known/agent-card.json` internally (verified against
   `a2a-go/v2@v2.3.1`'s own `resolver_test.go` — **not**
   `/.well-known/agent.json`, an earlier draft assumption that was wrong).
2. Run with `--agents-config /path/to/registry.yaml`. `internal/agent.LoadRegistry`
   parses it; `internal/agent.BuildAgentTools` turns each entry into
   `agenttool.New(remoteagent.NewA2A(...), nil)` and the resulting tools
   are merged into `NewRootAgent`'s 4th param (`agentTools []tool.Tool`).
3. No card is fetched at load time — `BuildAgentTools` makes zero network
   calls; the card fetch happens lazily on first tool invocation. A bad
   URL or unreachable server only surfaces when the model actually tries
   to call that tool, not at startup.

## Failure modes (all warnings, never fatal)

`BuildAgentTools` skips a bad entry and returns it as a string in
`warnings`, printed by `cmd/agentic-hooks/run.go`'s `loadAgentTools` as
`warning: ...` — the rest of the registry still loads:
- empty `name`
- empty `card_url`
- `card_url` that fails `url.Parse`
- `remoteagent.NewA2A` construction error

A missing or unparseable **registry file itself** (the `--agents-config`
path) is the one fatal case — that's a config-file problem, not a
per-entry one.

## What NOT to do

- Don't hand-roll the well-known-path suffix onto `card_url` in the YAML —
  it's appended by the library; a URL that already includes
  `/.well-known/agent-card.json` will double it and fail to resolve.
- Don't add a synchronous card-fetch/health-check at load time to "fail
  fast" — the lazy-fetch behavior is deliberate (an unreachable agent
  shouldn't block `agentic-hooks run` from starting if that tool is never
  actually invoked this run).
- Don't confuse this with the Search/Review `SubAgents` loop — those are
  fixed, in-process, and unaffected by `--agents-config`; omitting the
  flag entirely reproduces today's exact root-agent construction (`nil`
  agentTools).
