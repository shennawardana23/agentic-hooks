# ADR-0005: HITL is a CLI-level approve/reject prompt, not ADK tool-confirmation

## Status
Accepted

## Context
No agent output should reach the user as final without a human checkpoint.
ADK Go v2 has a real, built-in tool-level confirmation mechanism
(`tool.Context.RequestConfirmation()` / `.ToolConfirmation()`, demonstrated
in the ADK Go v2 source's `examples/toolconfirmation`), which was
considered for this gate. However, that mechanism confirms a *tool call*
before it executes — and the Review agent's verdict is not a tool call, it
is the agent's own reply text (its final answer to the root agent). Wiring
tool-confirmation onto the Review step would require rearchitecting the
verdict into a fake tool call solely to get a confirmation hook, adding
machinery that doesn't match what is actually happening.

## Decision
Human-in-the-loop approval is implemented as a CLI-level approve/reject
prompt in `cmd/agentic-hooks/run.go`, shown after the Review verdict
streams back from the agent loop. The prompt fails closed: any input other
than a literal `y` or `Y` is treated as a reject, and a rejected verdict is
never returned as final output. Every decision (approved or rejected) is
logged with a free-text reason to `feedback/feedback.jsonl` (see ADR-0008).

## Consequences
- Satisfies the "no output without a human checkpoint" requirement with
  less machinery than tool-confirmation would require, since the gate sits
  where the real decision point already is (the CLI, after the agent loop
  finishes) rather than being forced onto a tool-call boundary that doesn't
  correspond to the actual verdict-producing step.
- The gate is CLI-specific — an MCP-based or other non-CLI caller of the
  `run` pipeline would need its own equivalent gate; this decision does not
  generalize automatically to a future non-interactive caller.
- ADK's tool-confirmation primitive remains unused for this path; if a
  future tool call genuinely needs per-call confirmation (as opposed to a
  final-verdict gate), that is a separate decision, not a reversal of this
  one.