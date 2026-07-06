---
name: agentic-hooks-dev-loop
description: How to build, run, test, benchmark, and eval agentic-hooks locally using its Makefile. Use whenever the user wants to run make dev/build/test/bench/eval in this repo, is confused about TASK= vs FILE= vs a bare make dev, hits a "GEMINI_API_KEY or GOOGLE_API_KEY is required" error, or asks why make dev prompted for something instead of erroring. Trigger on any mention of "make dev", "make bench", "make eval", or running this project's CLI locally.
compatibility: Requires Go 1.25+ and make; GEMINI_API_KEY (or GOOGLE_API_KEY) needed for make dev/make eval, not for make build/test/bench.
metadata:
  spec: agentskills.io/specification
  project: agentic-hooks
---

# agentic-hooks dev loop

Single Makefile at repo root, `make help` lists everything with
descriptions. The two prompts people most often mistake for errors:

- **`Task (e.g. review: ...):`** — `make dev` with neither `FILE=` nor
  `TASK=` set prompts interactively instead of failing. This is intended
  behavior, not a bug report.
- **`Approve? [y/N]:`** — the HITL gate in `cmd/agentic-hooks/run.go`,
  unrelated to the Makefile prompt above. Every run (approved or rejected)
  gets appended to `feedback/feedback.jsonl` regardless of the answer —
  that's an intentional unconditional audit log, not a bug either.

## Core targets

- `make build` — `bin/agentic-hooks`, with version/build-time ldflags.
- `make server` — starts the MCP server over stdio (`serve` subcommand).
  It prints nothing and doesn't exit — that's correct, it's waiting on
  stdio for MCP JSON-RPC, not an interactive CLI. Ctrl-C to stop.
- `make dev` — runs the `run` subcommand (Search+Review loop + HITL).
  Three ways to give it a task:
  - `make dev FILE=path/to/code.go` — reviews a real file's contents.
  - `make dev TASK="review: ..."` — reviews inline text.
  - `make dev` bare — prompts for the task interactively.
  Requires `GEMINI_API_KEY` (or `GOOGLE_API_KEY` as a fallback — the
  Makefile re-exports whichever one is set under the name `run.go` actually
  reads, `GEMINI_API_KEY`).
- `make bench` — Go-native benchmarks (`go test -bench=. -benchmem`), no
  API cost. Use the `go-bench-runner` agent for comparing runs over time.
- `make eval` — golden-set eval against the *real* Gemini model
  (`internal/agent/eval_test.go`). Costs real API calls. Opt-in via
  `AGENTIC_HOOKS_EVAL=1`, which `make eval` sets for you — never runs as
  part of `make check`/`go test ./...`.
- `make check` — `vet` + `test` + `build`, the fast no-API-cost gate to
  run before considering any change done.

## If `make dev`'s Search sub-agent needs another MCP server

`--search-mcp-server`/`--search-mcp-server-args` are the flags; nothing is
hardcoded by design. `make dev` defaults to pointing Search at this
project's own `serve` subcommand (a valid stand-in — it calls
`list_knowledge`/`get_knowledge` instead of a real external lookup, but
proves the MCP-client wiring end-to-end). To point at a different MCP
server, edit the `Makefile`'s `dev` target's `--search-mcp-server*` flags,
or call `go run ./cmd/agentic-hooks run ...` directly with your own flags.

## Full manual testing guide

`TESTING.md` at repo root has the complete 3-tier walkthrough (automated
suite, manual `serve` check, full end-to-end `run`) plus MCP Inspector CLI
usage — read that for anything this summary doesn't cover.
