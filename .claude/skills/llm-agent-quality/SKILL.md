---
name: llm-agent-quality
description: Measure or improve the output quality of agentic-hooks's Generator/Review agents using the golden-set eval harness (internal/agent/eval_test.go), separate from Go performance benchmarks. Use this skill whenever the user asks whether the agent's actual answers/verdicts are good, wants to add a golden-set eval case, asks to run make eval, or asks about LLM/agent quality, accuracy, or regression — even if they don't say "eval" and just ask "is the review agent actually any good."
compatibility: Requires GEMINI_API_KEY (or GOOGLE_API_KEY) to run make eval; costs real API calls per run.
metadata:
  spec: agentskills.io/specification
  project: agentic-hooks
---

# LLM agent quality (golden-set eval)

`internal/agent/eval_test.go`'s `TestEval_ReviewGoldenSet` drives the real
Review agent against a table of `evalCase`s (a diff, an expected verdict,
an optional keyword expected in the transcript) and reports a pass rate.
This is quality measurement, not performance measurement — for speed/
allocations, use `go-bench-runner`/`make bench` instead.

Skipped by default (`t.Skip` unless `AGENTIC_HOOKS_EVAL=1`) so it never
runs inside `go test ./...`/`make check` and never costs an API call by
accident.

## Running it

```
make eval
```

Equivalent to `AGENTIC_HOOKS_EVAL=1 go test ./internal/agent/... -run TestEval -v`,
with the Makefile handling the `GEMINI_API_KEY`/`GOOGLE_API_KEY` fallback.
Requires a real key — there is no mocked/offline mode for this specific
test, that would defeat its purpose.

## Adding a golden-set case

Add an `evalCase` to the `evalCases` table:

```go
{
    name:        "descriptive_case_name",
    diff:        `<the code snippet to review>`,
    wantVerdict: "CHANGES_REQUESTED", // or "APPROVE"
    wantKeyword: "goroutine",         // lowercase substring expected in the transcript; "" to skip
}
```

Ground every new case in a real failure mode you actually observed (a bad
review the agent gave, a principle it missed), not a hypothetical one —
the existing 5 cases (swallowed error, goroutine leak, mixed receivers,
concrete dependency, clean add) came from real Second Brain concepts this
project already has. A case testing a principle that isn't in `knowledge/`
yet will just fail for the wrong reason.

## Gotchas

- `wantKeyword` checks a lowercase substring of the model's raw text
  transcript, not a structured field — the model's phrasing has to
  actually contain that word. If a case is flaky, check whether the
  keyword is too specific to one phrasing (e.g. "goroutine" is safer than
  "goroutine leak" if the model sometimes says "leaking goroutine").
- A `CHANGES_REQUESTED` case with no matched Second Brain concept will
  likely fail non-deterministically — the Review agent's grounding comes
  from `knowledge/`, not general training knowledge. Check
  `matchConceptsInDiff` would actually match something before adding the
  case (see the `second-brain-authoring` skill).
- Live model calls are non-deterministic. A single failing run isn't
  automatically evidence of a regression — rerun before concluding
  something broke, the same way `go-bench-runner` treats a single noisy
  benchmark run with caution.

## What NOT to do

- Don't loosen `AGENTIC_HOOKS_EVAL`'s skip guard so this runs by default —
  it must stay opt-in; a surprise API charge on a routine `go test ./...`
  is exactly what the guard exists to prevent.
- Don't grade success by "the model said something plausible" — use the
  verdict + keyword check consistently, that's what makes the pass rate
  mean something across runs and sessions.
