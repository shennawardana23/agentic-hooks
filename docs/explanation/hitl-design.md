# Why human-in-the-loop is a CLI prompt, not ADK tool-confirmation

Every `agentic-hooks run` ends with a plain approve/reject prompt in the
terminal, not ADK's built-in tool-level confirmation primitive. ADK Go v2
does have a real mechanism for this — `ctx.RequestConfirmation()` /
`ctx.ToolConfirmation()`, confirmed via the `examples/toolconfirmation`
sample in the ADK Go v2 source — so this wasn't a case of the "real" HITL
feature being unavailable. It was a deliberate choice not to use it here.

## The reasoning

ADK's tool-confirmation mechanism gates a *tool call*: the framework pauses
before a specific tool executes and waits for a human decision on whether
that call should proceed. But in this project, the Review verdict isn't a
tool call — it's the Review agent's own reply text, produced by an LLM
generating a response, not a discrete tool invocation with side effects to
gate.

Wiring `ctx.RequestConfirmation()` in would mean forcing the verdict itself
to be represented as a tool call just to get a confirmation checkpoint on
it — solving the problem by reshaping the architecture around the
mechanism, rather than using the mechanism where it actually fits (gating
a tool with side effects, not gating "should this text be shown to the
user").

Instead, the CLI itself is the gate: `cmd/agentic-hooks/run.go` prints the
Review agent's final transcript and prompts `Approve? [y/N]:` after the
self-correcting loop exits (either via `APPROVE`/`exit_loop` or via
`--max-iterations`). Only on an explicit `y`/`Y` does the transcript get
printed as final output and logged; anything else — including no input, a
typo, or an empty line — is treated as a reject. This is a fail-closed
design: the default outcome of ambiguity is "don't trust the output,"
never "let it through."

## Why this still satisfies the actual requirement

The requirement driving HITL in this project is: **no agent output that
could influence a code decision should reach the user as final without a
human checkpoint.** A CLI-level gate satisfies that requirement completely
— nothing is treated as final until a human explicitly approves it — with
less machinery than tool-level confirmation would require, and without
distorting the Review agent's role into something it isn't (a tool with
side effects).

## What gets logged either way

Whether the human approves or rejects, `internal/feedback` appends a
record to `feedback/feedback.jsonl`: `{timestamp, task, transcript,
approved, reason}`. The CLI prompts for an optional free-text reason on
both paths. This is the durable human-feedback annotation log — a record
for a future offline eval/training step to read, not a live training loop.
Unconditional logging (approved *and* rejected runs both get recorded) is
itself a locked decision — see
[ADR 0008](../adr/0008-feedback-log-unconditional-append.md).

## Related decisions

- [ADR 0005](../adr/0005-hitl-as-cli-prompt-not-tool-confirmation.md) — this
  decision, formally recorded.
- [Self-correcting loop](self-correcting-loop.md) — what produces the
  verdict this gate sits after.
