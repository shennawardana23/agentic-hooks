# ADR-0008: Feedback log appends unconditionally, every run

## Status
Accepted

## Context
Human approve/reject decisions on Review verdicts are a valuable signal for
future offline evaluation and RLHF-style training work (see ADR-0002 for
how offline eval currently consumes this kind of data). A design question
was whether to log only approved runs (a "clean" record of accepted
verdicts) or every run regardless of outcome.

## Decision
`internal/feedback` appends a record to `<--feedback-dir>/feedback.jsonl`
after every HITL decision, whether approved or rejected:
`{timestamp, task, transcript, approved, reason}`. The CLI prompts for a
free-text reason on both approve and reject. This is a durable,
append-only audit trail, not a live training loop — nothing reads this
file back into the running pipeline.

## Consequences
- Rejected verdicts are preserved with their reason, which is exactly the
  negative-signal data a future offline eval or preference-learning step
  would need — logging approvals only would have discarded this.
- The log grows unconditionally with every `run` invocation; no pruning or
  rotation exists yet. If volume becomes a concern, that is a new decision,
  not an implicit change to this one.
- The log is local-only (`feedback/feedback.jsonl`, not checked into git,
  not sent anywhere) — consistent with the project's privacy stance that
  `serve` and the feedback mechanism don't transmit data beyond the
  configured Gemini model call itself.