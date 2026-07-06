---
name: adk-v2-verification
description: The verify-before-you-write discipline for google.golang.org/adk/v2 in this project. Use before writing any new code, spec, or plan that references an ADK Go v2 API you haven't personally checked this session (agent, llmagent, loopagent, runner, session, tool, mcptoolset, exitlooptool, retryandreflect, etc.). Trigger whenever a claim about ADK Go v2's behavior sounds plausible but wasn't checked against go doc or actual module source — this exact failure mode has already produced real bugs in this codebase.
compatibility: Requires the go CLI and google.golang.org/adk/v2 already resolved in GOMODCACHE (go mod download).
metadata:
  spec: agentskills.io/specification
  project: agentic-hooks
---

# ADK Go v2 verification discipline

`google.golang.org/adk/v2` is bleeding-edge with thin public docs. This
project has been burned more than once by trusting a plausible first-pass
summary of its API instead of checking source. Documented real examples
(`SESSION_HANDOFF.md`):

- `tool.Set` doesn't exist. Real type: `tool.Toolset`. Caught at
  `go build` time, after it had already been written into a plan.
- `RequestInputEvent` doesn't exist. Real HITL primitives:
  `ctx.RequestConfirmation()` / `ctx.ToolConfirmation()`. Caught during
  spec self-review, before implementation — the cheap place to catch it.

Both mistakes came from describing the API the way a *similar* framework
would plausibly work, not the way this one actually does.

## The rule

Before writing a claim about ADK Go v2's API into code, a spec, or a plan:

1. Check the pinned version in `go.mod` (`google.golang.org/adk/v2 v2.0.0`
   as of this writing — don't trust that number without re-checking,
   it may have moved).
2. Run `go doc google.golang.org/adk/v2/<package>` for the specific
   symbol. This is cheap and precise — prefer it over reading whole files.
   A full-file `Read` on an unfamiliar large file is flagged in
   `SESSION_HANDOFF.md` as one of the more expensive single actions
   available when you only need to check whether something exists.
3. If `go doc` doesn't resolve it, `grep`/`find` the module source
   directly under `$(go env GOMODCACHE)/google.golang.org/adk/v2@<version>`
   for the exact symbol name.
4. Only then write the claim down — and phrase it as verified ("confirmed
   via `go doc`: ...") rather than as a bare assertion, so the next person
   (or agent) reading it knows it isn't a guess.

For anything non-trivial, delegate this to the `adk-api-verifier` agent
instead of doing it inline — it exists specifically to do steps 1-4 and
report back confirmed/wrong.

## What NOT to do

- Don't verify against `google.golang.org/adk` (no `/v2`) — a different,
  pre-v2 import path used by unrelated sibling projects, not this one.
- Don't treat a passing `go build`/`go test` as proof a *design* claim is
  right — the implementation-phase bugs in `SESSION_HANDOFF.md` (e.g. the
  backwards `Brain.Query` call in the original `matchConceptsInDiff`) were
  caught by actually running the code, not by research alone. Verify the
  API shape before writing, then still test the behavior after.
