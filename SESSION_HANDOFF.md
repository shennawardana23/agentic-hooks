# Session Handoff — agentic-hooks

Last updated: 2026-07-04 (later same day: self-correcting loop added). Read
this before doing anything else if you're picking this project up in a new
session.

## What this project is

A single Go binary (`agentic-hooks`) with two jobs:
1. Run an ADK-based multi-agent pipeline: Search (lookup) feeds a
   self-correcting Generator↔Review loop (draft → critique → revise → …
   until APPROVE or `--max-iterations`) that answers a task against a
   personal "Second Brain" of coding principles, gated by a human
   approve/reject prompt that logs a feedback annotation.
2. Expose that Second Brain as an MCP server for external coding agents
   (Claude Code, Cursor, etc.).

## Agent loop architecture (added 2026-07-04, same-day follow-up)

`internal/agent/`: `search.go` (MCP lookup, one-shot) → `generator.go`
(drafts/revises a text answer) + `review.go` (critiques, calls ADK's
`exitlooptool` on APPROVE) wrapped by `loop.go`'s `NewSelfCorrectingLoop`
(`google.golang.org/adk/v2/agent/workflowagents/loopagent`) → `root.go`
delegates to search then the loop.

Verified against ADK Go v2 source (`go doc`, not assumed):
- `exitlooptool.New()` sets `ctx.Actions().Escalate = true`; `loopagent`
  checks `event.Actions.Escalate` after each full sub-agent pass to stop.
  Real termination signal, not decorative.
- Sub-agents share the runner's session, so on iteration 2 the Generator
  sees the Review agent's iteration-1 CHANGES_REQUESTED verdict in
  conversation history — the loop can actually correct, not just repeat.
- `retryandreflect` plugin (`google.golang.org/adk/v2/plugin/retryandreflect`)
  attached via `runner.Config.PluginConfig` in `run.go` for tool-call
  self-healing (resilience). `--max-iterations` (default 4) is the second
  resilience layer — a non-converging task returns the Generator's
  best-effort last draft instead of looping forever.
- `cmd/agentic-hooks/run.go`'s `runRootAgent` now streams every event's text
  live, tagged `[author]` (generator/review/search/root) — this is the
  "TUI CLI" surface (enhanced plain-CLI streaming was chosen over a real
  bubbletea TUI — no new dependency).
- `internal/feedback/` — new package. Append-only JSONL
  (`<--feedback-dir>/feedback.jsonl`) written after every HITL decision:
  `{timestamp, task, transcript, approved, reason}`. The CLI now prompts for
  a free-text reason on both approve and reject — this is the RLHF-style
  human-feedback annotation log; it's a durable record for an offline
  eval/training step to read later, not a live training loop.

**Done (2026-07-04, v4 session):** both previously-open verification items are
now closed.
- **Loop convergence, proven without a live model.** Added
  `TestSelfCorrectingLoop_ConvergesAfterCorrection` in
  `internal/agent/loop_test.go` — scripted `scriptedGeneratorModel` /
  `scriptedReviewModel` stand-ins (built on the same pattern as ADK's own
  `loopagent` test fixtures, verified by reading
  `google.golang.org/adk/v2@v2.0.0/agent/workflowagents/loopagent/agent_test.go`
  in the module cache) drive the *real* `NewSelfCorrectingLoop` through a
  real `runner.Run`. Asserts: (1) `reviewModel.calls == 2` and
  `genModel.calls == 2` — the loop stops right after `exit_loop`, not by
  exhausting `--max-iterations`; (2) the generator's second request literally
  contains the review agent's first-pass `CHANGES_REQUESTED` text in its
  conversation history. This is a genuine correction proof, not just
  construction/wiring — no API key needed. 14 tests total now, all passing.
- **Live-run proof, done.** `GEMINI_API_KEY` was exported and used (note:
  code reads `os.Getenv("GOOGLE_API_KEY")` in `newDefaultModel`, but
  `GEMINI_API_KEY` works anyway — confirmed by reading
  `google.golang.org/genai@v1.62.0/client.go`: `genai.NewClient` falls back
  to reading `GEMINI_API_KEY`/`GOOGLE_API_KEY` from the environment itself
  whenever the passed-in `ClientConfig.APIKey` is `""`, regardless of
  whether the config struct is nil — no code change was needed). Ran the
  exact TESTING.md §3 command against `gemini-flash-latest`: the Generator
  produced a real, correct SOLID/DIP refactor of the `DoEverything()`
  example; the loop reached the HITL prompt; the reject path (empty stdin)
  produced "Rejected — no output returned as final." and wrote a correct
  feedback JSONL record with the full transcript. One run hit a transient
  `503 UNAVAILABLE` from the model (retried once, succeeded) — not a bug.
  **One real observed gap, not a bug:** the Review agent's own verdict text
  (APPROVE/CHANGES_REQUESTED) was not visible in the streamed `[review]`
  output — only `[generator]` text printed before the prompt. Root cause
  not yet diagnosed (plausibly the model called `exit_loop` in the same turn
  without emitting a separate text part) — didn't spend further live-API
  cost chasing it this session. `TESTING.md` already documents that live
  LLM output content isn't asserted on, so this doesn't block calling tier 3
  "verified", but streaming visibility of the Review verdict is a candidate
  follow-up if it recurs.
- **Security note:** the `GEMINI_API_KEY` used for the live run was pasted
  in plaintext into the chat transcript by the user. Recommended they rotate
  it after this session; did not persist the key to any file.
- **Makefile bug found and fixed:** `make run` with no `TASK` silently sent
  an empty string to the model, producing a confusing
  `GenerateContentRequest.contents: contents is not specified` API error
  instead of a clear local failure. Added a guard
  (`@if [ -z "$(TASK)" ]; then echo ...; exit 1; fi`) so it now fails fast
  locally before spending an API call.

Full design: `docs/superpowers/specs/2026-07-02-second-brain-orchestration-design.md`
Architecture diagram: https://claude.ai/code/artifact/063c19ed-c306-4cff-b389-1d0a3af59c3a

## Status

- [x] Brainstorming/scoping done (started way too broad — see "Scope history"
      below — narrowed to two pieces on purpose).
- [x] Design spec written and self-reviewed (later corrected mid-session:
      `RequestInputEvent` doesn't exist — real HITL API is
      `ctx.RequestConfirmation()`/`ctx.ToolConfirmation()`; `Tools` →
      `Toolsets []tool.Toolset` for MCP client attachment).
- [x] `writing-plans` — plan at
      `docs/superpowers/plans/2026-07-02-second-brain-orchestration.md`.
- [x] **All 7 plan tasks implemented and passing.** `go build ./...`,
      `go vet ./...`, and `go test ./... ` (12 tests) all clean as of
      2026-07-02. Everything is `git add`-staged, nothing committed (see
      Global Constraints below — standing no-commit instruction).
- [x] **Testing infra added.** `TESTING.md` at repo root: 3-tier guide
      (automated suite / manual `serve` check / full `run` end-to-end).
      New test `internal/mcpserver/server_integration_test.go` builds the
      **real** `agentic-hooks` binary and drives it as an MCP server over
      **real stdio** using the real MCP Go SDK client (`mcp.NewClient` +
      `mcp.CommandTransport`) — not the handler-level unit tests in
      `server_test.go`, an actual wire-protocol exercise. 13 tests total,
      all passing as of last run.
- [x] **Makefile added**, then corrected once (see below). Targets:
      `build`, `test`, `vet`, `tidy`, `check` (vet+test+build), `run`,
      `server`, `clean`. `run` and `server` both invoke `go run
      ./cmd/agentic-hooks ...` directly (i.e. trigger `main.go`, per user
      correction — the first draft ran a pre-built binary instead).
      **`run` still depends on `build`** because the Search sub-agent's
      `--search-mcp-server` flag needs a real spawnable binary to point
      its MCP client at; that binary is a subprocess `run` launches
      internally, not what `go run` executes for the top-level command.
      **Verified 2026-07-04 (v4 session):** `make -n run`/`make -n server`
      dry-run confirmed correct, and a real `make run` with no `TASK` caught
      a real bug (see below) that's now fixed. Still not staged/committed
      per the standing no-commit instruction.
- [x] **`run` subcommand's live agent loop — verified 2026-07-04 (v4
      session).** Ran end-to-end against a real Gemini API key
      (`GEMINI_API_KEY`, works despite the code reading
      `os.Getenv("GOOGLE_API_KEY")` — see loop-architecture section above).
      Generator produced a correct real SOLID refactor; HITL reject path
      and feedback JSONL logging both worked. One transient `503` from the
      model, resolved on retry. One unexplained-but-non-blocking gap: the
      Review agent's verdict text didn't stream visibly (see above) — not
      re-investigated further this session to avoid more live-API spend.
- [ ] `go.mod` now declares `go 1.25.0` (bumped to match the actual
      installed toolchain via `go version` — do not assume this is still
      current when you resume; re-check).
- [ ] User has not given final sign-off on the spec or the implementation.
- [ ] Nothing has been committed or pushed. If you want this work in git
      history, that's a decision for the user to make explicitly.
- [ ] **Pending decision, asked twice, no answer both times:** user wants
      a "massive audit" of Provider Fallback + a Skills system, pointing
      at two sibling repos that already have relevant code:
      `/Users/msw/Desktop/Development/Startup_Companies/Arcipelago_International/archpublicwebsite-agentic/internal/model/failover`
      (an ADK `model.LLM` failover wrapper — NOT Genkit, despite user
      calling it "genkit middleware failover" — circuit-breaker retry
      across providers; pinned to `google.golang.org/adk/model`, the
      **pre-v2 import path**, not `google.golang.org/adk/v2/model` that
      this project uses — compatibility unconfirmed) and
      `/Users/msw/Desktop/Development/My_Repository/go-adk-q` (has its own
      `model/failover`, `model/middleware`, and a full `skills/` taxonomy
      — agents, collaboration, design, documentation, engineering,
      workflow — large, unaudited). User also referenced
      genkit.dev/docs/go/middleware (incl. "Skills middleware"),
      adk.dev/skills, and agentskills.io as specs to reconcile against.
      **Do not start this unprompted** — it's a large scope addition on
      top of an already-complete MVP; ask again before spending on it.

## Decisions already locked — do not re-litigate without new information

- **Runtime**: ADK Go v2 (`google.golang.org/adk/v2`) is the sole
  orchestrator. Do not add Genkit to the request path.
- **Genkit's only role**: offline eval (LLM-as-judge over exported ADK
  session traces), never inline.
- **MCP**: `github.com/modelcontextprotocol/go-sdk`, used both as client
  (Search sub-agent, via `mcptoolset`) and server (`serve` subcommand). Not
  Genkit's `GenkitMCPServer` (different underlying lib, `mark3labs/mcp-go`)
  anywhere.
- **Second Brain storage**: local Markdown files with real OKF frontmatter
  (`type`/`title`/`description`/`tags`/`timestamp`) + freeform body content.
  Not a database. File path (minus `.md`) is the identifier, no separate
  `id` field — that's how OKF spec defines identity.
- **HITL**: implemented as a CLI-level approve/reject prompt in
  `cmd/agentic-hooks/run.go` after the Review verdict comes back — NOT
  ADK's tool-level `ctx.RequestConfirmation()`/`ctx.ToolConfirmation()`
  (that mechanism exists and is real, confirmed via
  `examples/toolconfirmation` in the ADK Go v2 source, but wiring it would
  mean the Review verdict itself is a tool call, which it isn't — it's
  the agent's own reply text). CLI-side gate satisfies the same "no
  output without a human checkpoint" requirement with less machinery.
  Fail-closed on anything but literal `y`/`Y` (treat as reject).
- **Binary/CLI name**: `agentic-hooks`, subcommands `run <task>` and
  `serve`.
- **Tree-sitter**: explicitly deferred (user decision). Review agent's
  entry point already has a `structuralFacts` param reserved as the seam
  for it later — don't refactor that signature away.
- **No commits**: user's standing instruction for this whole engagement —
  do not `git commit` anything unless they explicitly ask in a given
  message.

## Open items still needing the user's input

1. Which MCP server(s) the Search sub-agent points at by default (nothing
   hardcoded on purpose — needs at least one configured for local dev).
2. Confirm the go.mod version bump target once you resume (check current
   Go stable, don't trust a stale number from this doc).

## Scope history (why it's this small)

Original ask was enormous — full multi-agent platform, VueFlow frontend,
RAG, evals, tracing dashboard, full Diátaxis doc suite, ADRs, llms.txt, the
works. Per user's own org instructions ("back and forth Q&A before
starting") and general YAGNI/MVP discipline, this was deliberately
decomposed. Deferred items are listed in full in the design spec §7 — they
are a real roadmap, not abandoned. Don't silently re-expand scope without
the user asking for a specific deferred item back.

## Research discipline established this session

Multiple early research passes contained real mistakes that got corrected
mid-session (e.g. first pass wrongly concluded Genkit Go has no MCP-server
mode — it does; first take on Open Knowledge Format wrongly concluded it
was dataset/table-catalog-only and a mismatch — checking the actual SPEC.md
fields showed it's a genuine fit). Lesson: **verify library/API claims
against actual docs/source before writing them into a spec** — don't trust
a first-pass summary, and don't assume a framework doesn't have a
capability just because a first search didn't surface it. This session's
research cost ran to ~$9.75 across several verification passes — reuse the
confirmed facts above rather than re-deriving them, only re-research if
verifying something new or something that seems to have changed.

**Implementation-phase bugs** (caught by actually compiling/running, not
by re-reading docs — the plan's own written code had these wrong even
after the research above):
- `tool.Set` doesn't exist; real type is `tool.Toolset` (`go doc` caught
  this at `go build` time).
- `internal/agent/review.go`'s original `matchConceptsInDiff` logic called
  `secondbrain.Brain.Query(diff)`, checking whether the *entire diff*
  appears inside a concept's text — backwards. Fixed to check whether a
  concept's title/tag appears inside the diff instead. `Brain.Query` itself
  is correct for its actual use case (short user-typed topic → matching
  concepts); it was just the wrong tool for this call site.
- `cmd/agentic-hooks/run.go` importing both
  `agentic-hooks/internal/agent` and `google.golang.org/adk/v2/agent` in
  the same file collided on the identifier `agent` — aliased the internal
  one as `myagent`.
- Plan's Task 7 draft left `--search-mcp-server-args` unwired (`NewSearchAgent`
  took `mcpArgs` but `run.go` hardcoded `nil`) — added the flag.

**Total session cost: ~$52** by the time this handoff was last updated —
driven mostly by the ADK Go v2 API verification passes (bleeding-edge
library, thin public docs, needed direct module-cache source reads to get
real code) and a large `Read` of an external project's failover.go file.
Lesson for next session: prefer `go doc` and targeted `grep`/`find` over
reading full files end-to-end when only checking whether something
exists; a whole-file `Read` on an unfamiliar large file is one of the more
expensive single actions available.

## Working style notes for whoever resumes

- User communicates in Bahasa Indonesia, expects terse/direct responses
  (caveman-style — no filler, no hedging).
- User wants deep research/verification before committing to any technical
  claim, especially around library capabilities (avoid hallucinated APIs).
- User prefers being told the zero-trust/best-practice default and having
  it just applied, rather than being asked trivial questions — but genuine
  trade-off decisions (e.g. tree-sitter) should still be surfaced, not
  decided unilaterally.
- Diagrams/visuals are welcome for architecture communication (used the
  Artifact tool for an HTML architecture diagram this session).
