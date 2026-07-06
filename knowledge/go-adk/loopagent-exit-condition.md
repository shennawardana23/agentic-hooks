---
type: principle
title: A LoopAgent's exit condition must be a real Escalate signal, not implicit
description: Verified against ADK Go v2 source (workflowagents/loopagent, tool/exitlooptool) — not assumed.
tags: [go, adk, loopagent, self-correction]
timestamp: 2026-07-04
resource: https://github.com/google/adk-go
---

`loopagent.New` re-runs its SubAgents in sequence until `MaxIterations` is
hit or any sub-agent's emitted event sets `Actions.Escalate = true` — that
flag, not any heuristic on the output text, is what actually stops the loop.
`tool/exitlooptool.New()` is the standard way to set it: give the critic
agent in a generate/critique loop that tool and instruct it to call it only
on approval.

Two failure modes to design against explicitly:
- If no sub-agent ever calls the exit tool, the loop runs to
  `MaxIterations` and returns whatever the last pass produced — treat that
  as a "best effort, did not converge" result, not a success.
- If `MaxIterations == 0`, the loop runs until Escalate or forever — always
  set an explicit bound in production code.

Sub-agents in the loop share the same session, so a generator sub-agent on
iteration N can see the critic's iteration N-1 feedback in conversation
history automatically — no explicit data passing between iterations is
needed for that to work, but the generator's instruction must actually tell
it to look for and act on that feedback.
