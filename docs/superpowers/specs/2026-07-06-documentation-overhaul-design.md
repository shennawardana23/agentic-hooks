# Documentation Overhaul — Design Spec v0.1

Status: approved, pending implementation
Date: 2026-07-06

## 1. Purpose

`agentic-hooks` currently has minimal documentation: a one-line `README.md`,
a `TESTING.md`, and a `SESSION_HANDOFF.md` carrying session-to-session
context. This spec covers building out a full documentation system so that
(a) a new human developer can get productive quickly, and (b) any agent
(Claude Code, Cursor, or this project's own Search/Review agents) picking up
the repo cold has everything it needs without re-deriving context from git
history or prior chat transcripts.

## 2. Scope and non-goals

In scope:
- Diátaxis-structured docs (tutorials / how-to / reference / explanation).
- Architecture Decision Records for decisions already made.
- Diagrams (Mermaid inline, D2 rendered to SVG).
- `llms.txt` / `llms-full.txt` per llmstxt.org.
- Nine agent-legible root files (`AGENTS.md`, `SYSTEM.md`, `MEMORY.md`,
  `SKILL.md`, `PLAN.md`, `DESIGN.md`, `PRODUCT.md`, `APPEND_SYSTEM.md`,
  updates to the existing `SESSION_HANDOFF.md`).
- A final HTML summary report (via the Artifact tool, not committed to
  the repo — it's a one-off index for the human, not a project file).

Non-goals (explicitly not doing, and why):
- No Excalidraw or draw.io diagrams — neither MCP server is connected this
  session; would require fabricating a tool connection that doesn't exist.
- No historical ticket/code-review log with real past entries — this repo
  has one commit and no issue tracker. `APPEND_SYSTEM.md` is seeded only
  with real dated entries already documented in `SESSION_HANDOFF.md`; a
  ticket/review section starts empty and fills going forward.
- No restructuring of `docs/superpowers/` (existing specs/plans), no
  restructuring of `knowledge/` (Second Brain content data, not docs about
  the project).
- No `git commit` — standing no-commit instruction for this engagement.

## 3. Structure

```
agentic-hooks/
├── AGENTS.md
├── SYSTEM.md
├── MEMORY.md
├── SKILL.md
├── PLAN.md
├── DESIGN.md
├── PRODUCT.md
├── APPEND_SYSTEM.md
├── SESSION_HANDOFF.md          (update in place)
├── llms.txt
├── llms-full.txt
├── README.md                   (update in place)
├── TESTING.md                  (update in place)
└── docs/
    ├── superpowers/             (existing, untouched)
    ├── tutorials/
    │   └── first-run.md
    ├── how-to/
    │   ├── add-a-concept.md
    │   ├── point-search-at-another-mcp-server.md
    │   ├── run-benchmarks.md
    │   ├── run-the-golden-set-eval.md
    │   └── test-with-mcp-inspector.md
    ├── reference/
    │   ├── cli.md
    │   ├── mcp-tools.md
    │   ├── second-brain-frontmatter.md
    │   └── makefile-targets.md
    ├── explanation/
    │   ├── why-adk-not-genkit.md
    │   ├── self-correcting-loop.md
    │   ├── hitl-design.md
    │   └── architecture-overview.md
    ├── adr/
    │   ├── 0001-adk-go-v2-as-sole-orchestrator.md
    │   ├── 0002-offline-eval-as-plain-go-harness.md
    │   ├── 0003-mcp-sdk-bidirectional-one-library.md
    │   ├── 0004-second-brain-as-markdown-not-database.md
    │   ├── 0005-hitl-as-cli-prompt-not-tool-confirmation.md
    │   ├── 0006-gemini-api-key-canonical.md
    │   ├── 0007-tree-sitter-deferred.md
    │   ├── 0008-feedback-log-unconditional-append.md
    │   └── 0009-self-correcting-loop-via-loopagent.md
    └── diagrams/
        ├── architecture-overview.d2 (+ .svg)
        ├── run-sequence.mmd
        ├── serve-sequence.mmd
        └── loop-state-machine.mmd
```

## 4. Diátaxis content plan

Each doc is written to satisfy exactly one Diátaxis quadrant — no mixing
tutorial steps into reference material, no explanation prose in how-to
guides (per diataxis.fr's core rule: separate by user need, not by topic).

- **Tutorial** (`tutorials/first-run.md`): one linear path — clone, build,
  create a one-file Second Brain, run `make dev` against a sample diff,
  see a verdict, approve it. Learning-oriented; assumes nothing.
- **How-to guides**: task-oriented, assume the reader already knows the
  basics. Each answers one "how do I ___" question a working developer
  will actually have (add a concept, swap the Search MCP server, run
  benchmarks, run the eval harness, drive MCP Inspector manually).
- **Reference**: dry, structural, exhaustive for its scope — CLI flags,
  MCP tool JSON schemas, frontmatter fields, Makefile targets. No prose
  justification, just facts (mirrors `make help`'s own terseness).
- **Explanation**: the "why" — ADK vs. Genkit, why the loop converges,
  why HITL is CLI-level, an architecture walkthrough tying the diagrams
  together. Pulled from the real rationale already in
  `docs/superpowers/specs/2026-07-02-second-brain-orchestration-design.md`
  and `SESSION_HANDOFF.md`, not re-derived from scratch.

## 5. Architecture Decision Records

Nine ADRs, all retroactive records of decisions already made and locked in
`SESSION_HANDOFF.md`'s "Decisions already locked" section — none are new
decisions invented for this spec. Format follows the common
context/decision/consequences template (Fowler / adr.github.io convention),
numbered sequentially, immutable once written (a reversed decision gets a
new ADR that supersedes the old one, not an edit).

| ADR | Decision |
|---|---|
| 0001 | ADK Go v2 is the sole request-path orchestrator |
| 0002 | Offline eval is a plain Go golden-set harness (supersedes the original Genkit-eval plan noted in `knowledge/go-genkit/offline-eval-not-inline.md`) |
| 0003 | One MCP SDK, used as both client and server |
| 0004 | Second Brain is OKF-frontmatter Markdown files, not a database |
| 0005 | HITL is a CLI-level approve/reject prompt, not ADK tool-confirmation |
| 0006 | `GEMINI_API_KEY` is canonical; `GOOGLE_API_KEY` is a fallback only |
| 0007 | Tree-sitter structural analysis is deferred; `structuralFacts` seam reserved |
| 0008 | Feedback log appends unconditionally (every run, approved or rejected) |
| 0009 | Self-correction loop via ADK `loopagent` + `exitlooptool`, bounded by `--max-iterations` |

## 6. Diagrams

- `docs/diagrams/architecture-overview.d2` — static component diagram
  (CLI → ADK runtime / MCP server → Second Brain), rendered to `.svg` via
  the locally installed `d2` CLI (v0.7.1, confirmed installed) and
  embedded in `explanation/architecture-overview.md`.
- `run-sequence.mmd`, `serve-sequence.mmd` — Mermaid sequence diagrams,
  inlined directly as fenced ` ```mermaid ` blocks (GitHub renders these
  natively, no build step).
- `loop-state-machine.mmd` — Mermaid state diagram for the
  Generator↔Review convergence loop, inlined in
  `explanation/self-correcting-loop.md`.

## 7. llms.txt / llms-full.txt

Root-level, following llmstxt.org's spec: `llms.txt` is an H1 title, a
blockquote summary, and Markdown-link sections pointing at every doc in
§3 (grouped by Diátaxis quadrant plus ADRs and root files). `llms-full.txt`
is the same index with every linked document's full content concatenated
inline, so an agent with no filesystem access can still get everything in
one fetch.

## 8. Root agent-legible files

Per the table already confirmed with the user:

| File | Content |
|---|---|
| `AGENTS.md` | Build/test/run commands, conventions, doc map |
| `SYSTEM.md` | Architecture overview, component responsibilities, data flow |
| `MEMORY.md` | Durable project-level facts/decisions (distinct from the user's personal `~/.claude` memory system) |
| `SKILL.md` | How Second Brain concepts are authored/tagged/matched |
| `PLAN.md` | Current/active roadmap (forward-looking; distinct from per-feature plans under `docs/superpowers/plans/`) |
| `DESIGN.md` | Standing design principles (distinct from per-feature specs) |
| `PRODUCT.md` | What this tool is for, two real personas (CLI user, MCP-consuming agent host), what success looks like |
| `APPEND_SYSTEM.md` | Append-only changelog, seeded with real dated entries from `SESSION_HANDOFF.md` (2026-07-02 onward) |
| `SESSION_HANDOFF.md` | Updated in place to reference the new doc tree; history preserved |

## 9. Verification approach

- Every internal Markdown link checked by hand (no link-checker tool
  installed; small enough doc set to verify manually).
- Every code/command snippet in how-to/tutorial docs must be one already
  verified working in this session or in `TESTING.md` — no untested
  commands presented as working.
- ADR content cross-checked line-by-line against `SESSION_HANDOFF.md` and
  the existing design spec — no new claims invented.
- `go build ./...` / `go vet ./...` / `go test ./...` re-run after any doc
  change that touches `docs/superpowers/` conventions, to confirm
  documentation work didn't accidentally touch code.

## 10. Deferred (explicit roadmap, not silently dropped)

- Excalidraw / draw.io diagrams — revisit if those MCP servers get
  connected.
- Real historical ticket/code-review log — revisit if/when a tracker
  (GitHub Issues, Linear, etc.) is actually connected.
- `adr-tools` CLI adoption — ADRs are hand-written Markdown for now; the
  CLI just automates numbering/templating, not required for correctness.

## 11. Open items for user review

- None outstanding — all prior ambiguities (history source, sequencing,
  root-file contracts, doc layout) were resolved via clarifying questions
  before this spec was written.
