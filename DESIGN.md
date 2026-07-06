# DESIGN.md — Standing Design Principles

These are the principles that generalize from `agentic-hooks`'s locked
decisions ([MEMORY.md](MEMORY.md)) — apply them when making a new design
choice in this codebase, rather than re-deriving from first principles
each time. Per-feature rationale lives in [docs/adr/](docs/adr/) and
[docs/explanation/](docs/explanation/); this file is the durable "how we
think about this codebase," not a decision log.

## Fail-closed on human checkpoints

Anything gating on a human decision (HITL approve/reject) treats "no
clear approval" as reject, not approve. The `run` HITL prompt requires a
literal `y`/`Y`; anything else — empty input, any other character, no
response — is a rejection. No agent output that could influence a code
decision reaches the user as "final" without passing through this gate.
See [docs/explanation/hitl-design.md](docs/explanation/hitl-design.md).

## One library per cross-cutting concern, not one per feature

Where two frameworks could technically do the same job (ADK vs. Genkit
for orchestration; `modelcontextprotocol/go-sdk` vs. a second MCP library
for client vs. server roles), this project picks exactly one and uses it
for every instance of that concern. Two tracing pipelines or two config
surfaces for the same responsibility is treated as an avoidable cost, not
a flexibility win, unless there's a concrete capability gap forcing the
split. See [docs/explanation/why-adk-not-genkit.md](docs/explanation/why-adk-not-genkit.md).

## Text-first, database-never for durable content this project owns

The Second Brain is Markdown files with real OKF frontmatter, not rows in
a database. Durable content that's small in volume, edited by hand or by
an agent, and benefits from being human-readable and git-diffable is kept
as plain files unless there's a concrete query pattern a database
actually solves better. This is a default, not a blanket rule — revisit
if scale or query complexity changes.

## Prefer the framework's built-in primitive over hand-rolled machinery, when it's close to free

HITL uses ADK's own `ctx.RequestConfirmation()`-adjacent CLI-level
pattern rather than building a separate confirmation subsystem, because
the primitive already existed and cost little to wire in. The same logic
applies to `exitlooptool`/`loopagent` for the self-correcting loop rather
than a hand-rolled retry counter. Conversely, don't force a framework
primitive to fit a shape it wasn't designed for (see
[docs/adr/0005](docs/adr/0005-hitl-as-cli-prompt-not-tool-confirmation.md)
for a case where the "more native" option was rejected because it didn't
actually fit).

## Fail small, not silent, on partial data

A malformed Second Brain concept file is logged and skipped, not fatal to
the whole directory load — one bad file doesn't take down the knowledge
base, but it also doesn't fail invisibly; it's logged so it can be found.
The same shape applies elsewhere: prefer "skip this one unit of work and
say so" over either "crash everything" or "silently drop with no trace."

## Reserve seams for deferred work; don't refactor twice

When a capability is explicitly deferred (tree-sitter structural
analysis, offline eval), the call site that will eventually need it
carries a reserved, currently-unused parameter or package boundary
(`StructuralFacts`, the `internal/eval` package boundary) rather than
being designed as if the capability will never exist. This avoids a
breaking signature change later, at the cost of one unused parameter now.
Don't remove these seams as "dead code" — they're intentional.

## Verify library/API claims against source, not memory

For any library where behavior is being asserted as fact (especially
bleeding-edge ones like `google.golang.org/adk/v2`), check `go doc` or
actual source before writing the claim into a spec, ADR, or comment. This
project's own history has real bugs and wasted research spend caused by
skipping this (see [MEMORY.md](MEMORY.md)'s "Bugs found and fixed" and
[.claude/skills/adk-v2-verification/SKILL.md](.claude/skills/adk-v2-verification/SKILL.md)).
A first-pass summary of a framework's capabilities is not verification.

## Build only what's asked; defer the rest explicitly

Scope decisions favor a small, complete MVP over a large, partially-built
platform. Deferred work is tracked in [PLAN.md](PLAN.md) with the reason
it was deferred, not silently dropped and not silently re-expanded
without the user asking for a specific deferred item back.
