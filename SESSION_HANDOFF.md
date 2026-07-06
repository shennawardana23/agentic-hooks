# Session Handoff — agentic-hooks

Last updated: 2026-07-06 (documentation overhaul, v5 session). Read this
before doing anything else if you're picking this project up in a new
session.

## Start here: the doc tree

As of 2026-07-06 this repo has a full documentation system, not just this
file. **[AGENTS.md](AGENTS.md)** is now the recommended first read for any
agent or developer picking this up cold — compact entry point, build/test
commands, doc map. `docs/` holds the Diátaxis-structured content
(`tutorials/`, `how-to/`, `reference/`, `explanation/`), and
`docs/adr/` holds immutable Architecture Decision Records for every locked
decision. This file remains the deep narrative/session-to-session history —
see the "2026-07-06 — Documentation overhaul" section below for the full
new structure and rationale. Nothing below this point has been altered;
only this section and the dated entry near the bottom are new.

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
  **correction, 2026-07-06**: this entry originally claimed the code reads
  `os.Getenv("GOOGLE_API_KEY")` in `newDefaultModel` — that's backwards.
  `cmd/agentic-hooks/run.go`'s `newDefaultModel` reads
  `os.Getenv("GEMINI_API_KEY")` directly; `GOOGLE_API_KEY` is the one that
  works as a fallback, not the other way around. See
  [ADR-0006](docs/adr/0006-gemini-api-key-canonical.md) for the corrected,
  verified account. The mechanism below is still accurate — confirmed by reading
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
- [x] **Answered, 2026-07-06 (v5 session).** The item below was asked twice
      with no answer; a third framing — "Intelligent Communication Through
      Tool Calling" (agents see each other's public manifests, decide
      dynamically who to delegate to via a communication tool, instead of a
      fixed orchestration graph) — arrived this session and the user
      confirmed it **is** the same initiative. This directly reopens the
      locked "no network A2A, in-process ADK delegation only" decision
      above — that reopening is intentional and user-confirmed, not a
      silent scope change. **Sequencing, user-confirmed:** queued after the
      agent-policy-layer work (`docs/superpowers/specs/2026-07-06-agent-policy-layer-design.md`)
      finishes implementation — do not start the agent-communication
      brainstorm until that lands. When it's time, it gets its own fresh
      brainstorming cycle (new architecture, not an add-on to the policy
      layer), and should pull in the two sibling repos referenced below as
      real prior art, not reinvent from scratch.
  - Original ask, preserved for context: a "massive audit" of Provider
    Fallback + a Skills system, pointing at two sibling repos that already
    have relevant code:
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

## 2026-07-06 — Documentation overhaul (v5 session)

Full documentation system built out per
[docs/superpowers/specs/2026-07-06-documentation-overhaul-design.md](docs/superpowers/specs/2026-07-06-documentation-overhaul-design.md)
(status: approved, then implemented same day). New/updated structure:

```
agentic-hooks/
├── AGENTS.md            (new — start-here for any agent)
├── SYSTEM.md            (new — architecture overview)
├── MEMORY.md            (new — durable project facts/decisions)
├── SKILL.md             (new — how the Second Brain works)
├── PLAN.md              (new — forward-looking roadmap)
├── DESIGN.md            (new — standing design principles)
├── PRODUCT.md           (new — personas, what success looks like)
├── APPEND_SYSTEM.md     (new — append-only dated changelog)
├── SESSION_HANDOFF.md   (this file, updated in place — history preserved above)
├── llms.txt / llms-full.txt   (new — per llmstxt.org)
├── README.md            (updated — links into new doc tree)
├── TESTING.md           (updated — links into new doc tree)
└── docs/
    ├── superpowers/      (existing, untouched)
    ├── tutorials/first-run.md
    ├── how-to/{add-a-concept,point-search-at-another-mcp-server,run-benchmarks,run-the-golden-set-eval,test-with-mcp-inspector}.md
    ├── reference/{cli,mcp-tools,second-brain-frontmatter,makefile-targets}.md
    ├── explanation/{why-adk-not-genkit,self-correcting-loop,hitl-design,architecture-overview}.md
    ├── adr/0001–0009 (see MEMORY.md for the decision-to-ADR mapping)
    └── diagrams/{architecture-overview.d2+svg, run-sequence.mmd, serve-sequence.mmd, loop-state-machine.mmd}
```

Explicitly deferred: Excalidraw/draw.io diagrams (neither MCP server
connected this session — see [PLAN.md](PLAN.md)); a real historical
ticket/code-review log (no issue tracker connected — `APPEND_SYSTEM.md`
seeds real dated entries from this file only, fills in going forward).

Earlier the same day (before the doc work): the `agentic-hooks serve` MCP
server was wired into the user's `claude_desktop_config.json` (backed up
the original config first), the binary was rebuilt, and all
`list_knowledge`/`get_knowledge` behavior — including the bogus-id error
path — was verified via MCP Inspector CLI. All checks passed; no repo
files were touched by that verification pass.

**For whoever resumes next**: `AGENTS.md` is now the recommended first
read (it's the compact "start here" entry point); this file remains the
deep narrative/decision history and is still the right place to look for
*why* something is the way it is, or what was tried and abandoned.
`MEMORY.md` now duplicates the "Decisions already locked" list from this
file in a more durable/linkable form (with ADR cross-references) — keep
both in sync if a decision changes; don't let them drift.

## 2026-07-06 — Full-repo audit + fixes (v5 session, same day)

A requested "massive audit, double-check" pass (acting as senior
backend/AI-infra reviewer) found the documentation overhaul above had a
real gap: the flagship "Review verdict grounded in a matched Second Brain
concept" claim (asserted in `PRODUCT.md`, a reference doc, an explanation
doc, and the run-sequence diagram) was **not actually wired into the live
`run` pipeline** — `NewReviewAgent` took `brain` as a parameter and never
used it; `BuildReviewPrompt`/`matchConceptsInDiff` were only ever called
from tests/eval. **Fixed**, verified via `adk-api-verifier`: `generator.go`
now sets `OutputKey: "draft"`; `review.go` uses a new `InstructionProvider`
(`newReviewInstructionProvider`) that reads the draft from session state
and rebuilds the prompt via `BuildReviewPrompt` on every loop iteration —
deterministic, not an LLM-optional tool call. New test in `loop_test.go`,
`TestSelfCorrectingLoop_ReviewGroundedInMatchedSecondBrainConcept`, proves
it end-to-end by inspecting the real `LLMRequest.Config.SystemInstruction`
sent to the model.

Also fixed: no `.gitignore` existed — `bin/agentic-hooks` (27MB binary) and
`feedback/feedback.jsonl` (real HITL transcripts) were both tracked in
git, contradicting ADR-0008's claim that the feedback log is "not checked
into git." Added `.gitignore`, ran `git rm --cached` on both (staged, not
committed — historical commits still contain the old content; a real purge
would need an explicit history rewrite, not done). Also fixed: a reference
doc claimed 21 knowledge files, real count is 24; this file's own
`GOOGLE_API_KEY`/`GEMINI_API_KEY` claim (line ~78 above) was backwards,
corrected inline with a pointer to ADR-0006.

**Not fixed, tracked as follow-ups, not done this session** (Medium/Low
from the audit — user chose to fix Criticals/quick-High only):
- No automated test coverage for `run.go`'s HITL approve/reject parsing or
  the "feedback write failure just logs a warning" path.
- `internal/feedback.Append` has no file-locking note/test for concurrent
  writers (relies on OS-level `O_APPEND` atomicity, undocumented).
- `google.golang.org/genai` is pinned to v1.57.0; v1.62.0 is available.
  `gopkg.in/yaml.v3` is on the legacy import path (`go.yaml.in/yaml/v3`
  now supersedes it).
- `secondbrain.Load` reports skipped malformed files only via
  `log.Printf` to stderr — an MCP client calling `list_knowledge` has no
  way to learn concepts were silently dropped (this is `PRODUCT.md`'s own
  documented failure mode, currently unmitigated).

## 2026-07-06 — Agent Policy Layer: spec + plan written, NOT yet implemented (v5 session, same day)

User asked for a ~100-policy "collaboration policy" layer — how any agent
should work with the user (conduct, session/compaction, memory, caching,
persona-by-project, guardrails, security, env/secrets, anti-cheating,
communication style, escalation/HITL), read *before* the Second Brain, not
instead of it. Went through full `superpowers:brainstorming` →
`superpowers:writing-plans`:

- **Spec (approved):** [`docs/superpowers/specs/2026-07-06-agent-policy-layer-design.md`](docs/superpowers/specs/2026-07-06-agent-policy-layer-design.md).
  Scope confirmed by user: "Both" (a general personal standard AND this
  repo's own local enforcement point). Enforcement confirmed: 3 layers —
  auto-loaded `CLAUDE.md` (real enforcement for Claude-Code-family agents),
  advisory `get_agent_policy` MCP tool (MCP has no call-ordering
  enforcement — stated honestly, not oversold), and `.claude/settings.json`
  hooks (`SessionStart` + `PreToolUse` env-dump blocking). Structure
  confirmed: `POLICY.md` index + `policies/` dir, 10 category files, ~10
  policies each.
- **Plan (written, self-reviewed, NOT yet executed):**
  [`docs/superpowers/plans/2026-07-06-agent-policy-layer.md`](docs/superpowers/plans/2026-07-06-agent-policy-layer.md).
  15 tasks: `POLICY.md`, `CLAUDE.md`, the `get_agent_policy` MCP tool
  (`internal/mcpserver/server.go` signature change —
  `NewServer(brain, policyFilePath)` — touches 4 call sites), a Claude
  Desktop config update for the new `--policy-file` flag, all 10
  `policies/*.md` category files (full policy titles/one-liners already
  drafted in the plan itself, real-incident-grounded, not generic filler),
  the hooks (via the `update-config` skill, not hand-written JSON), and a
  final cross-link + link-check + build/vet/test verification task.
- **STOPPED HERE, mid-handoff:** the plan was presented with an execution
  choice — Subagent-Driven (`superpowers:subagent-driven-development`,
  recommended) vs. Inline (`superpowers:executing-plans`) — **user had not
  yet answered this when the session ended.** Whoever resumes: ask this
  question again before executing, don't assume an answer.
- Session cost was explicitly flagged mid-work ($16 → $27 → $37 → $39 across
  the audit-fix + brainstorm sequence); user said "proceed" at full scale
  each time asked — don't re-litigate scale unprompted, but do keep
  flagging cost growth per policy [10.5]/[10.9] (once `policies/` exists).

## 2026-07-06 — Agent-to-agent communication: queued, not started (v5 session, same day)

See the "Answered, 2026-07-06" entry in the Status section above — this is
the same initiative as the twice-previously-unanswered Provider
Fallback + Skills audit. **Explicitly queued after the Agent Policy Layer
plan above finishes implementation.** Gets its own fresh
`superpowers:brainstorming` cycle when it's time (new architecture: dynamic
manifest-based agent discovery/delegation via a communication tool,
directly reopening the "no network A2A, in-process ADK only" locked
decision — user-confirmed reopening, not silent). Pull in the two sibling
repos referenced in the Status-section entry as real prior art.

## 2026-07-06 — Agent Policy Layer: implemented and reviewed clean (v6 session, same day)

The plan from the section above (`docs/superpowers/plans/2026-07-06-agent-policy-layer.md`,
15 tasks) is now **fully executed**, via `superpowers:subagent-driven-development`
(user's answer to the previously-unanswered question — Subagent-Driven,
confirmed this session). User also explicitly instructed **no git commands
of any kind** for this execution (not just "no commit" — no `git add`,
`status`, `diff`, worktree, nothing) — so the skill's usual worktree
isolation and per-task-commit steps were both skipped by design; every
task ran directly in the working tree, and task reviewers verified diffs
via direct file reads instead of commit ranges.

**What exists now, all as uncommitted working-tree changes:**
- `POLICY.md` (root index), `CLAUDE.md` (auto-loaded gate), `policies/01-agent-conduct.md`
  through `policies/10-escalation-hitl.md` — 101 policies total (10 files ×
  10 items, plus one added during review, see below).
- `internal/mcpserver/server.go` gained a third tool, `get_agent_policy`;
  `NewServer(brain, policyFilePath string)` signature change, all real call
  sites updated; `cmd/agentic-hooks/serve.go` gained `--policy-file`
  (default `POLICY.md`). New tests `TestGetAgentPolicy_ReturnsRealPolicyFileContent`/
  `TestGetAgentPolicy_ErrorsWhenFileMissing`, both passing. Full suite
  (`go build`/`vet`/`test`) clean throughout.
- `/Users/msw/Library/Application Support/Claude/claude_desktop_config.json`
  updated with `--policy-file <absolute path>` for the existing
  `agentic-hooks-secondbrain` entry — backed up first (two timestamped
  `.bak.*` files now exist there), edited via targeted JSON mutation, not
  hand-editing, to avoid touching other servers' credentials in that file.
  This one required a second explicit user authorization mid-session —
  auto-mode correctly blocked the first attempt as a sensitive
  external-file change the generic plan approval didn't cover.
- `.claude/settings.json` (new) — `SessionStart` hook (prints a policy
  pointer) + `PreToolUse` hook on `Bash` (`.claude/hooks/check-bash-env-dump.sh`,
  blocks env-dump-shaped commands, points to `policies/07-env-secrets-protection.md`,
  has an `AGENTIC_HOOKS_ALLOW_ENV_ECHO=1` override). **This hook is now
  live and will fire on every future Bash tool call in this repo** —
  whoever resumes next should expect it.
- `AGENTS.md`, `README.md`, `llms.txt`, `llms-full.txt` cross-linked to the
  new material; `llms-full.txt` regenerated (43 `## SOURCE:` sections).

**Review found and fixed two real issues, not rubber-stamped:**
- Task 14's first pass on the env-dump hook had real false negatives on
  *idiomatic, non-adversarial* shell usage — `cat ".env"` (quoted),
  `echo "$SECRET_KEY"` / `echo ${SECRET_KEY}` (quoted/braced), and the
  `.env.local`/`.env.production` family all bypassed detection. Fixed by
  tightening the regexes; re-verified with a 24-case battery (all correct);
  subshell/indirect-invocation evasion (`$(env)`, `bash -c env`) was
  explicitly left as a documented, accepted limitation of a regex-based
  heuristic — not a hard security boundary, consistent with this layer's
  own "advisory-only" honesty stance.
- The final whole-branch review (dispatched on the most capable model)
  caught a real fabricated-citation defect in `CLAUDE.md` itself: it cited
  `[10.1]` for a "stop before destructive actions" rule, but the real
  `10.1` is about something unrelated (enforcement-limits honesty), and no
  policy in the set actually stated that rule anywhere — a citation to
  nothing, in the single most-read file of the whole layer, violating the
  layer's own policy 08.8 ("never fabricate a citation"). Fixed properly
  per the reviewer's own recommendation (adding the missing rule rather
  than just repointing the citation to something adjacent-but-wrong): added
  `policies/10-escalation-hitl.md`'s `10.11 — Pause Before Destructive or
  Hard-to-Reverse Actions` as a pure append (10.1–10.10 untouched), then
  corrected `CLAUDE.md` to cite `[10.11]`/`[10.3]` correctly. This is why
  the total is 101, not a round 100 — `POLICY.md`'s wording was adjusted
  to say "~100" consistently rather than claiming an exact count.
  `llms-full.txt` was regenerated again afterward so the embedded copy
  isn't stale.
- Minor, left as-is per the reviewer's own judgment (non-blocking): policy
  `04.6`'s grounding incident traces only to the plan document itself, not
  independently to `SESSION_HANDOFF.md`/an ADR — faithful to its stated
  source, just not as deeply cross-verified as its siblings.

**Known non-issue, checked and confirmed:** `bin/agentic-hooks` and
`feedback/feedback.jsonl` show as staged deletions (`D`) in `git status` —
this predates this session (from the `.gitignore`/`git rm --cached` fix in
the 2026-07-06 audit-fixes entry above) and is unrelated to the policy
layer; the actual 27MB binary is still present on disk, so the Claude
Desktop config fix above will still resolve correctly.

**Verified clean at every level:** each of the 15 tasks got its own
task-scoped review (spec compliance + code quality, both required
verdicts); the whole-set final review checked cross-task consistency (are
`CLAUDE.md`'s other citations to `07.1`/`07.2`/`08.1`/`08.3`/`01.2`/`10.3`
all correct — yes; does `POLICY.md`'s "honesty is itself policy 10.1" claim
actually hold — yes; do all `SESSION_HANDOFF.md`/ADR-grounded incident
claims across the 10 files actually check out against the real source —
yes, spot-checked ~10 of them). `go build`/`vet`/`test` clean at every
checkpoint. No git command mutated repo state — confirmed by comparing
`git log --oneline` before and after: still the same 3 pre-existing
commits.

**Not done, still open:**
- The Claude Desktop config change hasn't been exercised live (no fresh
  Claude Desktop session/MCP Inspector run against the new `--policy-file`
  flag this session) — the JSON is correct and the flag is wired
  end-to-end in Go, but an actual live `get_agent_policy` call through
  Claude Desktop itself wasn't performed.
- The `.claude/settings.json` hooks were verified by direct script
  invocation with constructed JSON stdin, not by an actual fresh Claude
  Code session start in this repo (which would require ending this session
  to observe).
- Agent-to-agent communication (see the two sections above) is now
  unblocked (the policy layer it was queued behind has shipped) but was
  **not started** this session; still needs its own fresh
  `superpowers:brainstorming` cycle per the "Answered, 2026-07-06" entry
  above.

## Outstanding tasks, all in one place (read this first if resuming cold)

1. Start a fresh `superpowers:brainstorming` cycle for agent-to-agent
   communication (see the two sections above) — now unblocked, not started.
2. Everything from this entire multi-session engagement (~60+ files: the
   documentation overhaul, the audit fixes, the policy layer spec/plan/
   implementation) is **staged/modified but not committed** — standing
   no-commit-unless-asked instruction still applies. If asked to commit,
   this is a lot of unrelated-looking changes accumulated across multiple
   long sessions; consider whether the user wants one big commit or
   several logical ones (docs overhaul / audit fixes / policy layer)
   before running `git add -A`.
3. ~~Live-exercise the two "Not done, still open" items directly above~~ —
   **done, 2026-07-06 v7 session, see entry below.**
4. ~~Medium/Low audit follow-ups from the 2026-07-06 audit-fixes entry
   above~~ — **done, 2026-07-06 v7 session, see entry below.**

## 2026-07-06 — Items 3 and 4 closed out (v7 session, same day)

Grepped the whole codebase (`TODO`/`FIXME`/`deferred`/`not yet` etc. across
`.go` and `.md`) to confirm this file's outstanding list was actually
complete before acting — it was; no code `TODO`/`FIXME` exist anywhere in
this repo. `go build`/`vet`/`test` were all re-verified clean before and
after every change below. Items 1 and 2 above were **not** touched — both
are explicit user decisions (item 1 reopens a locked ADR and needs its own
brainstorm; item 2 is gated by the standing no-commit instruction) and
weren't asked for this session.

**Item 3, both sub-parts, actually exercised live (not just re-claimed):**
- The `.claude/settings.json` `SessionStart` hook fired for real at the
  start of this session (visible in this session's own startup context) —
  the "real fresh-session hook fire" gap is closed.
- `get_agent_policy` was called for real via MCP Inspector CLI
  (`npx @modelcontextprotocol/inspector --cli bin/agentic-hooks serve
  --knowledge-dir knowledge --policy-file POLICY.md --method tools/call
  --tool-name get_agent_policy`) against a freshly `make build`-rebuilt
  binary — the real `POLICY.md` content came back correctly.
  `tools/list` and a `list_knowledge --tool-arg tag=go` call were also
  re-run live to confirm nothing regressed. Documented in
  `docs/how-to/test-with-mcp-inspector.md` and `docs/reference/mcp-tools.md`
  (the latter was also found to be stale — still said "two tools", missing
  `get_agent_policy` entirely and the new `warnings` field below; fixed).

**Item 4, all four Medium/Low audit follow-ups implemented, tested, and
verified clean (`go build`/`vet`/`test`, including `-race` on the new
concurrency test):**
- **HITL parse + feedback-write-failure coverage.** `cmd/agentic-hooks/run.go`'s
  inline approve/reject-and-record logic was extracted into
  `promptForApprovalAndRecordFeedback` (same fail-closed behavior: anything
  but literal `y`/`Y` rejects). New `run_test.go` covers 6 approval-parsing
  cases plus a dedicated test proving a feedback-write failure only warns
  to output and does not flip the human's approval decision.
- **`internal/feedback.Append` concurrency.** Docstring now states the real
  guarantee (`O_APPEND` + single `f.Write`, POSIX-only, no in-process lock,
  `PIPE_BUF` caveat for very long transcripts) instead of leaving it
  undocumented. New `TestAppend_ConcurrentWritersProduceNoCorruptedLines`
  runs 50 concurrent `Append` calls and asserts every line survives intact
  and parses — passes under `-race`.
- **Dependency bumps**, one at a time with a build+test gate after each:
  `google.golang.org/genai` v1.57.0 → v1.62.0 (clean). `gopkg.in/yaml.v3`
  → `go.yaml.in/yaml/v3` v3.0.4 — this is a **module-path migration**, not
  a version bump (same yaml.v3 maintainers moved the canonical import path;
  `gopkg.in/yaml.v3` itself has had no release past v3.0.1). Only
  `internal/secondbrain/secondbrain.go` imported it — one-line import swap,
  `go mod tidy`, rebuilt clean. `google.golang.org/adk/v2` checked via
  `go list -m -versions` — still pinned at v2.0.0, nothing newer exists to
  bump to.
- **Silent-skip visibility.** `secondbrain.Brain` gained `SkippedFiles()
  []string` (one `"path: reason"` per file `Load` couldn't parse — additive,
  no signature change to `Load`/`Brain`). `ListKnowledgeOutput` gained
  `Warnings []string` (`omitempty`, additive to the wire schema) populated
  from `brain.SkippedFiles()` in `listKnowledgeHandler`. `log.Printf` is
  still there too (kept for local-dev visibility) — this adds the
  previously-missing caller-visible channel, doesn't replace the log. New
  tests in both `internal/secondbrain` and `internal/mcpserver`; verified
  live via MCP Inspector (`tools/list` shows `warnings` in the real output
  schema).

**Not re-litigated, correctly left alone:** items 1 and 2 above remain
open user decisions — see those two entries; nothing in this session's
work started the agent-to-agent brainstorm or committed anything.
