# ADR-0009: Self-correcting loop via ADK loopagent + exitlooptool

## Status
Accepted

## Context
The Generator/Review pipeline needs to iterate — draft, critique, revise —
until the Review agent approves or a bound is reached, rather than
producing one unreviewed draft. ADK Go v2 provides a purpose-built
workflow primitive for this:
`google.golang.org/adk/v2/agent/workflowagents/loopagent`, paired with
`exitlooptool` as the termination signal. This was verified against actual
ADK Go v2 source (`go doc`, and `loopagent`'s own test fixtures in the
module cache), not assumed from documentation summaries alone.

## Decision
The self-correcting loop is implemented with ADK's `loopagent` wrapping the
Generator and Review sub-agents, bounded by `--max-iterations` (default 4).
`exitlooptool.New()` sets `ctx.Actions().Escalate = true` when the Review
agent approves; `loopagent` checks `event.Actions.Escalate` after each full
sub-agent pass and stops when set — a real termination signal, not a
decorative one. Sub-agents share the runner's session, so on a second
iteration the Generator sees the Review agent's prior
`CHANGES_REQUESTED` verdict in conversation history, enabling genuine
correction rather than repetition. This was proven without a live model via
`TestSelfCorrectingLoop_ConvergesAfterCorrection`, which asserts both that
the loop stops immediately after `exit_loop` (not by exhausting
`--max-iterations`) and that the generator's second request literally
contains the review agent's first-pass verdict text.
`retryandreflect` (ADK's tool-call self-healing plugin) is attached via
`runner.Config.PluginConfig` for resilience against transient tool-call
failures, as a second, independent resilience layer alongside the
iteration bound.

## Consequences
- A non-converging task returns the Generator's best-effort last draft
  after `--max-iterations` passes, instead of looping forever — bounded
  cost and latency at the expense of potentially returning an
  unapproved draft in pathological cases.
- Loop convergence is verified by a real scripted test against the actual
  `NewSelfCorrectingLoop`/`runner.Run`, not just by construction/wiring —
  a genuine regression test exists if this behavior ever breaks.
- One known, non-blocking gap remains: in a live run, the Review agent's
  own verdict text (APPROVE/CHANGES_REQUESTED) was not observed in the
  streamed `[review]` output — only `[generator]` text printed before the
  HITL prompt. Root cause not yet diagnosed (plausibly the model calling
  `exit_loop` in the same turn without emitting a separate text part).
  This does not change the decision recorded here, but is a candidate
  follow-up if it recurs.