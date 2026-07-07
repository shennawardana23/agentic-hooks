# Design: Structural Facts via Tree-sitter (Go-only)

| | |
|---|---|
| **Date** | 2026-07-07 |
| **Status** | Approved |
| **Session** | continuation of the agentic-hooks engagement |

## Context

The user asked for a feature "exact like"
[GitNexus](https://github.com/abhigyanpatwari/GitNexus) and
[Graphify](https://github.com/Graphify-Labs/graphify) — both verified
directly (not guessed): GitNexus (Node/TS) parses a repo via tree-sitter
into an embedded graph DB (LadybugDB) and exposes 17 MCP tools
(`query`/`impact`/`trace`/`explain`/`cypher`/...); Graphify (Python) parses
a folder via tree-sitter into a `graph.json`/`graph.html`, with
`explain`/`path`/`query` CLI verbs and `EXTRACTED`/`INFERRED` edge tags.
Both are full standalone code-knowledge-graph products.

Building "exactly" either is out of scope for this project as a single
feature — it's a multi-subsystem product (AST parsing, graph storage,
query/explain/path/impact logic, an MCP or file-based exposure layer).
Scoped down through direct clarifying questions with the user to: **feed
the existing Review agent's already-reserved `structuralFacts` seam**
(`internal/agent/review.go`'s `BuildReviewPrompt(diff, brain, facts)`,
`facts *StructuralFacts` always `nil` today) with real tree-sitter-derived
facts about Go functions in the reviewed draft — no graph DB, no new MCP
tools, no visualization.

This directly reopens **ADR-0007** ("Tree-sitter-based structural analysis
is deferred... explicit user decision, not an oversight"). Reopening is
user-confirmed this session, not a silent scope change — same pattern as
ADR-0010's A2A reopening.

## Verified since ADR-0007 was written (2026, session prior)

ADR-0007 recorded that Go tree-sitter bindings were CGo-based
(`tree-sitter/go-tree-sitter`, `smacker/go-tree-sitter`) and explicitly
instructed: "re-verify the state of Go tree-sitter bindings rather than
assuming today's landscape... still holds" before implementing. Re-checked
directly this session:

- `tree-sitter/go-tree-sitter@v0.25.0` and `smacker/go-tree-sitter` are
  both **still CGo-based** (confirmed via `grep "import \"C\""` on their
  fetched source) — ADR-0007's original blocker for those two libraries
  still holds.
- **New fact, didn't exist when ADR-0007 was written**:
  `github.com/odvcencio/gotreesitter` — a pure-Go tree-sitter runtime, no
  CGo, no C toolchain, cross-compiles to any `GOOS`/`GOARCH` including
  `wasip1`. Verified real and active via `gh api`: MIT license, 527 stars,
  created 2026-02-20, last pushed 2026-07-06 (yesterday relative to this
  session), not archived. Ships 206 grammars including Go
  (`grammars.GoLanguage()`), confirmed via its own README quick-start
  example. API confirmed via `go doc` against the fetched module
  (`v0.21.0`): `NewParser(lang).Parse(source []byte) (*Tree, error)`,
  `Node.Type(lang)`, `Node.StartPoint()/EndPoint()` (`Point{Row, Column}`),
  `Node.ChildByFieldName(name, lang)`, `Node.HasError()`, and a built-in
  `gotreesitter.Walk(node, fn func(*Node, depth int) WalkAction)` helper.

This directly changes ADR-0007's calculus: the CGo-build-complexity
concern that justified deferring no longer applies if `gotreesitter` is
used instead of the two originally-considered libraries.

## Decision

Build Go-only structural facts extraction using `github.com/odvcencio/gotreesitter`
+ `github.com/odvcencio/gotreesitter/grammars`, populating the
already-reserved `StructuralFacts` seam. Scope locked to exactly the three
gaps ADR-0007 itself named as missing: function length, nesting depth,
function boundaries. No graph storage, no new MCP tools, no
multi-language support, no visualization — those are explicitly the parts
of GitNexus/Graphify's scope this decision does **not** replicate.

**Additive/fail-open only.** A parse failure (non-Go draft, malformed
snippet, empty draft) must never block the Review loop — it falls back to
exactly today's behavior (`nil` facts, LLM reasons from diff text alone).

## Architecture & components

New file `internal/agent/structuralfacts.go` (one-concept-per-file,
matching the package's existing convention):

```go
package agent

type FunctionFacts struct {
    Name            string
    StartLine       int // 1-indexed
    EndLine         int
    LengthLines     int
    MaxNestingDepth int
}

type StructuralFacts struct {
    Functions []FunctionFacts
}

func AnalyzeStructuralFacts(source []byte) (*StructuralFacts, error)
```

This **replaces** the current empty `type StructuralFacts struct{}` in
`internal/agent/review.go` (moved out to its own file, now populated) —
`BuildReviewPrompt`'s signature (`facts *StructuralFacts`) is unchanged,
only the type's shape and what populates it change.

`AnalyzeStructuralFacts`:
1. `gotreesitter.NewParser(grammars.GoLanguage()).Parse(source)`.
2. If `tree.RootNode() == nil` (empty input, per `Parse`'s own documented
   contract) → return `&StructuralFacts{}, nil` (zero functions, not an
   error — matches the existing "no draft yet" fail-open convention).
3. If the root node reports `HasError()` (invalid/incomplete Go syntax) →
   return `nil, error` — caller fails open.
4. Otherwise, use `gotreesitter.Walk` over the root node; for each
   `function_declaration`/`method_declaration` node (exact grammar
   node-type strings to be confirmed against a real parse's `.SExpr(lang)`
   output during implementation, not assumed from memory of the generic
   tree-sitter-go grammar): read `Name` via `ChildByFieldName("name",
   lang).Text(source)`, `StartLine`/`EndLine` via
   `StartPoint().Row+1`/`EndPoint().Row+1`, `LengthLines = EndLine -
   StartLine + 1`. Compute `MaxNestingDepth` via a bounded sub-walk of the
   function node's subtree, incrementing a depth counter on
   `if_statement`/`for_statement`/`expression_switch_statement`/
   `type_switch_statement`/`select_statement` node types and tracking the
   maximum reached (again, node-type strings confirmed via `.SExpr`, not
   assumed).

## Data flow

`newReviewInstructionProvider` (`review.go`), after reading `draft` from
session state: calls `AnalyzeStructuralFacts([]byte(draft))`.
- Success (even zero functions) → pass the resulting `*StructuralFacts`
  into `BuildReviewPrompt`.
- Error → pass `nil` into `BuildReviewPrompt`, exactly today's call.

`BuildReviewPrompt` gains a new rendering block: when `facts != nil &&
len(facts.Functions) > 0`, render a "Structural facts (from tree-sitter
parse)" section (function name, line range, length, max nesting depth)
before the existing "Code change" section — giving the Review agent real,
verified facts instead of inferring structure from diff text alone (the
exact gap ADR-0007 named). When `facts` is `nil` or has zero functions,
prompt output is byte-identical to today.

## Error handling

- Empty source → zero functions, no error (see Architecture step 2).
- Syntax-error tree → explicit error, caller fails open to `nil` facts —
  never blocks the Review loop on a parse failure.
- No panics: every `ChildByFieldName` result is nil-checked before use (a
  malformed/partial node can have a missing field).

## Testing

`internal/agent/structuralfacts_test.go`, table-driven against real Go
source fixtures (no mocking of the parser):
- Single function, no nesting → `MaxNestingDepth == 0`.
- Single function with nested `if`/`for` (e.g. depth 3) → correct depth.
- Multiple functions in one source → all captured, correct order.
- Syntactically invalid Go source → `AnalyzeStructuralFacts` returns a
  non-nil error.
- Empty source (`[]byte{}`) → `&StructuralFacts{}`, no error.

`internal/agent/review_test.go` gains one case: `BuildReviewPrompt` called
with a non-nil `*StructuralFacts{Functions: [...]}` → prompt output
contains the rendered structural-facts section. Existing
`review_test.go`/`eval_test.go` call sites passing `nil` (or the
zero-value struct literal, if any exist — to be checked during planning)
must be verified to still compile and pass unchanged.

## Dependency

`github.com/odvcencio/gotreesitter` (+ `.../grammars`) added via `go get`
+ `go mod tidy`. Pure Go, no CGo, no new build-toolchain requirement — the
one part of ADR-0007's original reasoning ("keeping the build simple")
this design deliberately preserves, even while lifting the deferral
itself.

## Documentation

New `docs/adr/0011-go-structural-facts-via-gotreesitter.md` records that
ADR-0007's deferral is lifted specifically for Go-only structural facts
via `gotreesitter` — not a reversal of ADR-0007's original CGo-avoidance
reasoning, which this design still honors via a different library.
Cross-referenced from ADR-0007 and `MEMORY.md`, same pattern as ADR-0010.

## Explicitly out of scope (this is not GitNexus/Graphify)

- No graph data model, no graph storage (in-memory or persistent).
- No new MCP tools (`query`/`impact`/`trace`/`explain`/`cypher`-equivalents).
- No multi-language support — Go only.
- No visualization (`graph.html`-equivalent) or report file
  (`GRAPH_REPORT.md`-equivalent).
- No whole-repo indexing — this analyzes only the single draft string
  passed to Review at loop time, not a persistent cross-file graph.

If a full GitNexus/Graphify-style subsystem is wanted later, it is a
separate, much larger project needing its own decomposition and spec —
not an extension of this one.
