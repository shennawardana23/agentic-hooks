# ADR-0002: Offline eval is a plain Go golden-set harness

## Status
Accepted — supersedes the original Genkit-eval plan in the 2026-07-02 design
spec §4.5/§7 (`internal/eval` was originally scoped as a package that would
export ADK session traces to a Genkit raw-eval dataset and score them via
`genkit eval:run` + `DefineEvaluator`). This ADR does not edit that original
plan; it replaces it. If the Genkit-eval approach is ever revisited, that
would be a new ADR superseding this one, not an edit to either.

## Context
The project's own locked decision (`knowledge/go-genkit/offline-eval-not-inline.md`)
establishes that Genkit's only role in this project is offline evaluation —
LLM-as-judge scoring over exported ADK session traces and the human
approve/reject feedback log, run after the fact, never inline. The original
design spec assumed this would be implemented by adopting Genkit's
`eval:run` CLI and dataset format. In practice, no Genkit dependency was
ever added to `go.mod`, and the project's Go-only tooling preference made a
Node/npm-based eval CLI an awkward fit for a single Go binary project.

## Decision
Offline eval is implemented as a plain Go golden-set harness:
`internal/agent/eval_test.go`'s `TestEval_ReviewGoldenSet`, driven by
`make eval`. It runs the real Review agent against a fixed table of
`evalCase` entries (diff, expected verdict, expected keyword) and reports a
pass rate. It requires a real API key, is skipped by default (guarded by
`AGENTIC_HOOKS_EVAL=1`), and never runs as part of `make check` or
`go test ./...`.

## Consequences
- No Genkit dependency needed anywhere in `go.mod` — one fewer toolchain
  (Node/npm) required to run the project's own eval.
- The golden set lives as ordinary Go test data, versioned and reviewed like
  any other test — adding a case is a one-line struct literal, not a
  separate dataset file/schema to keep in sync with the code.
- Loses Genkit's built-in LLM-as-judge tooling and eval-dashboard UI; pass/
  fail is currently substring-matching on verdict and keyword, not a
  judged/scored rubric. If richer judging is needed later, that is a
  candidate for a new decision, not a silent scope creep of this harness.