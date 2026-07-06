# How to run the golden-set eval

The golden-set eval drives the real Review agent against a table of known
cases and reports a pass rate. Unlike the unit test suite, this makes real
model calls and costs real API usage — it is opt-in only and never runs as
part of `make check` or CI.

## Prerequisites

- `GEMINI_API_KEY` (or `GOOGLE_API_KEY` as a fallback) set in your
  environment.
- Willingness to spend real API calls — each case in the golden set is a
  live request to the configured Gemini model.

## Run it

```bash
export GEMINI_API_KEY="your-real-key"
make eval
```

Under the hood this runs:

```bash
AGENTIC_HOOKS_EVAL=1 go test ./internal/agent/... -run TestEval -v
```

The `AGENTIC_HOOKS_EVAL=1` environment variable is the opt-in gate — the
same test file is skipped by default so `go test ./...` and `make check`
never spend API calls unintentionally.

## Adding a new case

The golden set lives in `internal/agent/eval_test.go`. Add a new case
grounded in a real observed failure, not a hypothetical one — this
project's own `agent-quality-evaluator` subagent is built specifically to
extend this file and can do it for you if you describe the failure you
saw.

## Interpreting results

The eval reports a pass rate across the golden set. A failing case means
the Review agent's real-model output diverged from the expected verdict
for that case — investigate the specific case's prompt and Second Brain
concepts before assuming the model or the harness is at fault; live LLM
output has some run-to-run variance by nature.
