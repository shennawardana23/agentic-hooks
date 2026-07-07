# MEMORY.md — Durable Project Facts and Decisions

This is `agentic-hooks`'s own durable, project-level memory: decisions
already locked, verified facts about the libraries this project depends
on, and bugs already found and fixed. It is **not** the same thing as the
user's personal `~/.claude` memory system (that stores facts about the
user across all their projects; this stores facts about this codebase for
any agent or developer working in it). Full narrative history and
session-to-session context: [SESSION_HANDOFF.md](SESSION_HANDOFF.md).
Standing principles derived from these decisions: [DESIGN.md](DESIGN.md).
Full decision rationale for each: [docs/adr/](docs/adr/).

## Decisions already locked — do not re-litigate without new information

- **Runtime**: ADK Go v2 (`google.golang.org/adk/v2`) is the sole
  request-path orchestrator. Genkit is never added to the request path.
  ([docs/adr/0001](docs/adr/0001-adk-go-v2-as-sole-orchestrator.md))
- **Network A2A (opt-in, additive)**: the root agent can delegate to
  external agents over the real A2A protocol via `agent/remoteagent/v2` +
  `tool/agenttool`, driven by a static YAML registry
  (`--agents-config`, no default — omitted means zero behavior change).
  This supersedes ADR-0001's original "no network A2A" consequence for
  the tool-calling path only; in-process `SubAgents` delegation is
  unaffected.
  ([docs/adr/0010](docs/adr/0010-network-a2a-via-adk-remoteagent.md))
- **Genkit's only role**: offline eval (LLM-as-judge over exported ADK
  session traces), never inline.
  ([docs/adr/0002](docs/adr/0002-offline-eval-as-plain-go-harness.md))
- **MCP**: `github.com/modelcontextprotocol/go-sdk`, used both as client
  (Search sub-agent, via `mcptoolset`) and server (`serve` subcommand).
  Not Genkit's `GenkitMCPServer` (different underlying library,
  `mark3labs/mcp-go`) anywhere in this project.
  ([docs/adr/0003](docs/adr/0003-mcp-sdk-bidirectional-one-library.md))
- **Second Brain storage**: local Markdown files with real OKF frontmatter
  (`type`/`title`/`description`/`tags`/`timestamp`) plus freeform body
  content — not a database. The file path (minus `.md`) is the
  identifier; there is no separate `id` field.
  ([docs/adr/0004](docs/adr/0004-second-brain-as-markdown-not-database.md))
- **HITL**: a CLI-level approve/reject prompt in `cmd/agentic-hooks/run.go`
  after the Review verdict comes back — not ADK's tool-level
  `ctx.RequestConfirmation()`/`ctx.ToolConfirmation()` (that mechanism is
  real, confirmed via `examples/toolconfirmation` in the ADK Go v2
  source, but wiring it would mean the Review verdict itself is a tool
  call, which it isn't — it's the agent's own reply text). Fail-closed on
  anything but literal `y`/`Y`.
  ([docs/adr/0005](docs/adr/0005-hitl-as-cli-prompt-not-tool-confirmation.md))
- **`GEMINI_API_KEY` is canonical**; `GOOGLE_API_KEY` is a fallback only.
  Confirmed by reading `google.golang.org/genai`'s client source:
  `genai.NewClient` falls back to reading either env var itself whenever
  the passed-in `ClientConfig.APIKey` is empty.
  ([docs/adr/0006](docs/adr/0006-gemini-api-key-canonical.md))
- **Tree-sitter structural analysis is deferred** (explicit user
  decision, not an oversight). `internal/agent.StructuralFacts` is a
  reserved seam on the Review entry point — don't refactor that signature
  away when adding real structural analysis later.
  ([docs/adr/0007](docs/adr/0007-tree-sitter-deferred.md))
- **Feedback log appends unconditionally** — every `run` invocation
  writes a JSONL record to `feedback/feedback.jsonl`, whether the HITL
  decision was approve or reject.
  ([docs/adr/0008](docs/adr/0008-feedback-log-unconditional-append.md))
- **Self-correction loop** uses ADK's `loopagent` +
  `exitlooptool.New()`, bounded by `--max-iterations` (default 4).
  `exitlooptool.New()` sets `ctx.Actions().Escalate = true`; `loopagent`
  checks `event.Actions.Escalate` after each full sub-agent pass to stop —
  a real termination signal, not decorative.
  ([docs/adr/0009](docs/adr/0009-self-correcting-loop-via-loopagent.md))
- **Binary/CLI name**: `agentic-hooks`, subcommands `run <task>` and
  `serve`.
- **No commits without explicit ask**: standing instruction for this
  engagement — do not `git commit` anything unless asked in the current
  message.

## Verified library facts (re-verify if the pinned version changes)

- `google.golang.org/adk/v2@v2.0.0`: sub-agents share the runner's
  session, so on loop iteration 2 the Generator sees the Review agent's
  iteration-1 verdict text in conversation history — the self-correcting
  loop can actually correct, not just repeat. Verified by driving the
  real `NewSelfCorrectingLoop` through a real `runner.Run` with scripted
  model stand-ins (`internal/agent/loop_test.go`,
  `TestSelfCorrectingLoop_ConvergesAfterCorrection`) — no live API key
  needed for this proof.
- `retryandreflect` plugin
  (`google.golang.org/adk/v2/plugin/retryandreflect`) is attached via
  `runner.Config.PluginConfig` for tool-call self-healing. This is a
  resilience layer independent of `--max-iterations`.
- `go.mod` declares `go 1.25.0` (bumped from a stale `1.21.13` to match
  the actual installed toolchain during the 2026-07-02 session) — do not
  assume this is still current without re-checking `go version`.

## Bugs found and fixed (don't reintroduce)

- `tool.Set` doesn't exist as a type; the real type is `tool.Toolset`
  (caught by `go build`, not by docs — the design spec itself had this
  wrong before implementation).
- `internal/agent/review.go`'s `matchConceptsInDiff` originally called
  `secondbrain.Brain.Query(diff)`, checking whether the *entire diff*
  appears inside a concept's text — backwards. Fixed to check whether a
  concept's title/tag appears inside the diff instead. `Brain.Query`
  itself is correct for its actual use case (a short user-typed topic
  matching against concepts); it was the wrong tool for this call site.
- `cmd/agentic-hooks/run.go` importing both `agentic-hooks/internal/agent`
  and `google.golang.org/adk/v2/agent` in the same file collided on the
  identifier `agent` — aliased the internal one as `myagent`.
- `make run`/`make dev` with no `TASK` used to silently send an empty
  string to the model, producing a confusing API error instead of a clear
  local failure. The Makefile now guards and fails fast locally.
- The Search sub-agent's `--search-mcp-server-args` flag was originally
  unwired (`NewSearchAgent` accepted `mcpArgs` but `run.go` hardcoded
  `nil`) — fixed.

## Open, unresolved facts (don't assume answered)

- The Review agent's own verdict text (`APPROVE`/`CHANGES_REQUESTED`)
  was observed not streaming visibly in `[review]` output during one live
  run (2026-07-04) — only `[generator]` text appeared before the HITL
  prompt. Root cause not diagnosed (plausibly the model calling
  `exit_loop` in the same turn without a separate text part). Not
  currently blocking (live LLM output content isn't asserted on in
  tests), but a candidate follow-up if it recurs.
