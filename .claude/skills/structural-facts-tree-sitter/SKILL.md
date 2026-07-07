---
name: structural-facts-tree-sitter
description: How this project plans to feed the Review agent's reserved structuralFacts seam with real tree-sitter-derived Go structural data (function boundaries, length, nesting depth). Use whenever the user asks about structuralFacts, tree-sitter, gotreesitter, function-length/nesting-depth analysis in the Review prompt, or references GitNexus/Graphify-style code-graph features for this project. NOT YET IMPLEMENTED — this skill documents an approved design, not shipped code; do not assume AnalyzeStructuralFacts exists until this skill is updated to say so.
compatibility: Planned dependency is github.com/odvcencio/gotreesitter + its grammars subpackage — pure Go, no CGo. Not yet added to go.mod.
metadata:
  spec: agentskills.io/specification
  project: agentic-hooks
---

# Structural facts via tree-sitter (design only — not implemented)

**Status check first:** run `ls internal/agent/structuralfacts.go` before
doing anything else with this skill. If it doesn't exist (true as of this
writing), the feature described below is design-only — read
`docs/superpowers/specs/2026-07-07-structural-facts-tree-sitter-design.md`
in full before writing any code, don't work from this summary alone.

## What this is

Feeds `internal/agent/review.go`'s already-reserved
`BuildReviewPrompt(diff, brain, facts *StructuralFacts)` seam — `facts` is
`nil` at every call site today — with three specific facts about Go
functions in the reviewed draft that ADR-0007 named as the actual gap:
function boundaries, length, and nesting depth. This reopens ADR-0007
("tree-sitter deferred") for exactly this narrow use — not a reversal of
its CGo-avoidance reasoning, since `gotreesitter` is pure Go.

**Explicitly out of scope**, unlike GitNexus/Graphify (the two products
this was scoped down from): no graph DB, no new MCP tools, no
multi-language support, no visualization, no whole-repo indexing — this
only ever analyzes the single draft string passed to Review at loop time.

## Planned shape (from the approved spec, not yet built)

```go
// internal/agent/structuralfacts.go
type FunctionFacts struct {
    Name            string
    StartLine       int // 1-indexed
    EndLine         int
    LengthLines     int
    MaxNestingDepth int
}
type StructuralFacts struct{ Functions []FunctionFacts }
func AnalyzeStructuralFacts(source []byte) (*StructuralFacts, error)
```

- Empty source → `&StructuralFacts{}, nil` (fail-open, not an error).
- Syntactically invalid Go (`root.HasError() == true`) → `nil, error`;
  caller falls back to today's `nil`-facts behavior, never blocks Review.
- Node types for function/method boundaries: `function_declaration`,
  `method_declaration`. Nesting-depth node types: `if_statement`,
  `for_statement`, `expression_switch_statement`,
  `type_switch_statement`, `select_statement`. Both lists were confirmed
  against real `.SExpr()` parse output in a throwaway probe this session
  — not assumed from generic tree-sitter-go grammar memory.

## Before implementing — one gap not yet closed

The spec's own probe never exercised `HasError() == true` on malformed
input — only the happy path was run. **Verify that first** (a quick probe
with intentionally broken Go source) before trusting the fail-open error
path in real code; this is called out explicitly in
`SESSION_HANDOFF.md`'s resume point for this feature.

## What NOT to do

- Don't build the graph-DB/MCP-tools/visualization parts of
  GitNexus/Graphify under this skill's name — that's a separate,
  much-larger project needing its own spec, per the design doc's own
  "explicitly out of scope" section.
- Don't add multi-language support — this is Go-only by design.
- Don't skip the `HasError()`-on-bad-input verification step above and
  assume the fail-open path works because the happy path did.
