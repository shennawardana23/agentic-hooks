# APPEND_SYSTEM.md — Append-Only Changelog

**Ordering: chronological, oldest first. New entries are appended at the
bottom of this file, never inserted above or used to rewrite an existing
entry.** If a past entry turns out to be wrong, add a new entry that
corrects it — don't edit history.

This is a dated log of what actually happened to this project, session by
session. It is reconstructed here from [SESSION_HANDOFF.md](SESSION_HANDOFF.md)'s
existing narrative for entries before 2026-07-06; every entry below is
backed by that file (or, from 2026-07-06 onward, by this same session).
There is no connected issue tracker yet, so this is not a ticket log —
see [PLAN.md](PLAN.md) for that caveat.

---

## 2026-07-02 — Initial design and implementation

- Brainstormed and scoped `agentic-hooks` down from an initially much
  larger ask (full multi-agent platform, VueFlow frontend, RAG, tracing
  dashboard) to two pieces: the Search+Review pipeline and the MCP
  server. See [PLAN.md](PLAN.md)'s "Deferred roadmap" for the full
  original list.
- Wrote and self-reviewed the design spec
  ([docs/superpowers/specs/2026-07-02-second-brain-orchestration-design.md](docs/superpowers/specs/2026-07-02-second-brain-orchestration-design.md)),
  correcting one API assumption mid-session (`RequestInputEvent` doesn't
  exist; real HITL API is `ctx.RequestConfirmation()`/`ctx.ToolConfirmation()`;
  `Tools` field is actually `Toolsets []tool.Toolset`).
- Implemented all 7 plan tasks
  ([docs/superpowers/plans/2026-07-02-second-brain-orchestration.md](docs/superpowers/plans/2026-07-02-second-brain-orchestration.md)).
  `go build`, `go vet`, and `go test` (12 tests) all clean.
- Added `TESTING.md` (3-tier guide) and a real MCP stdio integration test
  that builds the actual binary and drives it with a real MCP client.
- Added the Makefile (`build`, `test`, `vet`, `tidy`, `check`, `run`,
  `server`, `clean`).
- Found and fixed implementation-phase bugs during actual compilation
  (not caught by research alone): `tool.Set` doesn't exist (real type
  `tool.Toolset`); `matchConceptsInDiff` had its match direction
  backwards; an `agent` package-name collision in `run.go`; an unwired
  `--search-mcp-server-args` flag. Full list: [MEMORY.md](MEMORY.md).
- Nothing committed to git — standing no-commit instruction for this
  engagement.

## 2026-07-04 — Self-correcting loop, feedback log, live-run verification

- Added the full agent-loop architecture: `internal/agent/search.go`,
  `generator.go`, `review.go`, `loop.go` (`NewSelfCorrectingLoop` via
  ADK's `loopagent`), `root.go`. Verified against actual ADK Go v2 source
  (not assumed) that `exitlooptool.New()`'s `Escalate` action is a real
  termination signal that `loopagent` checks after each pass.
- Attached the `retryandreflect` plugin for tool-call self-healing.
- Added `internal/feedback` — append-only JSONL log of every HITL
  decision, with a free-text reason prompt added to the CLI on both
  approve and reject.
- Proved loop convergence without a live model:
  `TestSelfCorrectingLoop_ConvergesAfterCorrection` in
  `internal/agent/loop_test.go`, using scripted model stand-ins driving
  the real loop through a real `runner.Run`. 14 tests total, all passing.
- Ran a full live end-to-end `run` against a real Gemini API key: the
  Generator produced a correct SOLID/DIP refactor, the loop reached the
  HITL prompt, the reject path worked, and a correct feedback JSONL
  record was written. One transient `503` from the model, resolved on
  retry — not a bug. One unresolved, non-blocking gap: the Review agent's
  own verdict text didn't stream visibly in `[review]` output (see
  [MEMORY.md](MEMORY.md)'s "Open, unresolved facts").
- Security note: a `GEMINI_API_KEY` was pasted in plaintext into the chat
  transcript by the user during this session; recommended rotation.
  Not persisted to any file by the agent.
- Found and fixed a Makefile bug: `make run`/`make dev` with no `TASK`
  silently sent an empty string to the model instead of failing fast
  locally. Added a guard.

## 2026-07-06 — Documentation overhaul

- Wrote the documentation overhaul design spec
  ([docs/superpowers/specs/2026-07-06-documentation-overhaul-design.md](docs/superpowers/specs/2026-07-06-documentation-overhaul-design.md)),
  resolving all open questions (doc layout, root-file contracts,
  sequencing) before implementation.
- Built out the full Diátaxis-structured documentation tree
  (`docs/tutorials/`, `docs/how-to/`, `docs/reference/`,
  `docs/explanation/`), 9 Architecture Decision Records
  (`docs/adr/0001`–`0009`), Mermaid + D2 diagrams (`docs/diagrams/`),
  `llms.txt`/`llms-full.txt` per [llmstxt.org](https://llmstxt.org/), and
  nine agent-legible root files: this file, `AGENTS.md`, `SYSTEM.md`,
  `MEMORY.md`, `SKILL.md`, `PLAN.md`, `DESIGN.md`, `PRODUCT.md`, and an
  update to `SESSION_HANDOFF.md` in place.
- Explicitly deferred Excalidraw/draw.io diagrams — neither MCP server
  was connected this session; revisit if/when they are (see
  [PLAN.md](PLAN.md)).
- Also, earlier the same day: added the `agentic-hooks serve` MCP server
  entry to the user's `claude_desktop_config.json`, built the binary,
  and verified all `list_knowledge`/`get_knowledge` behavior (including
  the bogus-id error path) via MCP Inspector CLI — all checks passed.
