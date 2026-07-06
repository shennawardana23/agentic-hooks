# ADR-0007: Tree-sitter structural analysis is deferred

## Status
Accepted

## Context
The Review agent's MVP is LLM-over-diff-text only — it has no AST or
structural parsing of the code it reviews (no function-length, nesting-depth,
or SRP-by-structure checks derived from real syntax trees). Tree-sitter was
considered as a pre-filter stage ahead of the Review agent, but Go
tree-sitter bindings are CGo-based (e.g. `tree-sitter/go-tree-sitter` or
`smacker/go-tree-sitter`), adding a non-trivial build/toolchain dependency
for a capability not required by the current MVP scope. This was an
explicit user decision to defer, not an oversight.

## Decision
Tree-sitter-based structural analysis is deferred. To avoid a future
refactor when it does land, the Review agent's entry point already takes an
optional `structuralFacts` parameter (`Review(ctx, diff, structuralFacts)`),
`nil` in the current MVP. A future tree-sitter pre-filter stage would
populate this parameter rather than changing the function signature.

## Consequences
- No CGo build dependency in the current codebase, keeping the build
  simple (`go build` with no external toolchain requirements).
- Review quality is bounded by what an LLM can infer from diff text alone —
  no guaranteed detection of structural issues that require a real parse
  (e.g. exact nesting depth, precise function boundaries).
- The `structuralFacts` seam must be kept in the `Review` signature and not
  refactored away casually — removing it would reintroduce the exact
  refactor this decision was designed to avoid when tree-sitter is
  eventually implemented. Before implementing, re-verify the state of Go
  tree-sitter bindings rather than assuming today's landscape (CGo-based,
  no confirmed pure-Go option) still holds.