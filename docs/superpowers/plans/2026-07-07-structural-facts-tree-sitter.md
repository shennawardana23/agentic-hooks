# Structural Facts via Tree-sitter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Feed the Review agent's already-reserved `structuralFacts` seam
(`internal/agent/review.go`'s `BuildReviewPrompt(diff, brain, facts)`,
always `nil` today) with real tree-sitter-derived facts about Go functions
in the reviewed draft: function boundaries, length, and max nesting depth
— the exact three gaps [ADR-0007](../../adr/0007-tree-sitter-deferred.md)
named as missing.

**Architecture:** New file `internal/agent/structuralfacts.go` exposes
`AnalyzeStructuralFacts(source []byte) (*StructuralFacts, error)`, built on
`github.com/odvcencio/gotreesitter` (pure Go, no CGo). `review.go`'s
`newReviewInstructionProvider` calls it on every loop iteration and passes
the result into `BuildReviewPrompt`, which gains a new rendering block.
Parse failure fails open to `nil` facts — byte-identical to today's prompt
output, never blocks the Review loop.

**Tech Stack:** Go 1.25, `github.com/odvcencio/gotreesitter` v0.21.0 +
`github.com/odvcencio/gotreesitter/grammars` (both new dependencies, pure
Go).

## Global Constraints

- Pure Go only — no CGo, no new build-toolchain requirement. This is the
  one part of [ADR-0007](../../adr/0007-tree-sitter-deferred.md)'s original
  reasoning this design preserves even while lifting the deferral itself.
- Scope locked to exactly three facts: function boundaries, length, max
  nesting depth. No graph DB, no new MCP tools, no multi-language support,
  no visualization, no whole-repo indexing — see the approved design spec's
  "Explicitly out of scope" section
  (`docs/superpowers/specs/2026-07-07-structural-facts-tree-sitter-design.md`).
- Additive/fail-open only: a parse failure (non-Go draft, malformed
  snippet, empty draft) must never block the Review loop — falls back to
  exactly today's behavior (`nil` facts, LLM reasons from diff text alone).
- Existing call sites passing literal `nil` for `facts`
  (`internal/agent/review_test.go:41,55`, `internal/agent/eval_test.go:132`,
  `internal/agent/review_bench_test.go:20`) must still compile and pass
  unchanged — they pass `nil`, not a struct literal, so this holds
  automatically once `StructuralFacts` gains fields.
- Standing project instruction: do not run `git commit` without an
  explicit user ask. Every task below ends with a build/vet/test
  verification step, not a commit step — this deliberately deviates from
  the writing-plans template's default "Commit" step.
- Node types (confirmed against real `.SExpr()` output during design,
  re-confirmed via `go doc` + a real probe run before this plan was
  written): function/method boundaries are `function_declaration` and
  `method_declaration`; nesting constructs are `if_statement`,
  `for_statement`, `expression_switch_statement`, `type_switch_statement`,
  `select_statement`.

---

### Task 1: Structural facts analyzer

**Files:**
- Modify: `go.mod`, `go.sum` (via `go get` + `go mod tidy`)
- Create: `internal/agent/structuralfacts.go`
- Create: `internal/agent/structuralfacts_test.go`

**Interfaces:**
- Produces: `FunctionFacts{Name string, StartLine, EndLine, LengthLines, MaxNestingDepth int}`,
  `StructuralFacts{Functions []FunctionFacts}`,
  `AnalyzeStructuralFacts(source []byte) (*StructuralFacts, error)` — all
  consumed by Task 2.

- [ ] **Step 1: Add the dependency**

Run: `go get github.com/odvcencio/gotreesitter@v0.21.0 && go mod tidy`

Expected: `go.mod` gains a `require github.com/odvcencio/gotreesitter v0.21.0`
line (and its transitive deps), `go.sum` updates. Nothing imports it yet,
so `go build ./...` must still succeed with zero errors — this only proves
the dependency resolves cleanly, it doesn't wire anything in yet.

- [ ] **Step 2: Write the failing test**

Create `internal/agent/structuralfacts_test.go`:

```go
package agent

import "testing"

func TestAnalyzeStructuralFacts(t *testing.T) {
	t.Run("single function no nesting", func(t *testing.T) {
		source := []byte(`package main

func add(a, b int) int {
	return a + b
}
`)
		facts, err := AnalyzeStructuralFacts(source)
		if err != nil {
			t.Fatalf("AnalyzeStructuralFacts() error = %v", err)
		}
		if len(facts.Functions) != 1 {
			t.Fatalf("len(Functions) = %d, want 1", len(facts.Functions))
		}
		fn := facts.Functions[0]
		if fn.Name != "add" {
			t.Errorf("Name = %q, want %q", fn.Name, "add")
		}
		if fn.MaxNestingDepth != 0 {
			t.Errorf("MaxNestingDepth = %d, want 0", fn.MaxNestingDepth)
		}
		if fn.StartLine != 3 || fn.EndLine != 5 {
			t.Errorf("StartLine/EndLine = %d/%d, want 3/5", fn.StartLine, fn.EndLine)
		}
		if fn.LengthLines != 3 {
			t.Errorf("LengthLines = %d, want 3", fn.LengthLines)
		}
	})

	t.Run("nested if inside for inside if reaches depth 3", func(t *testing.T) {
		source := []byte(`package main

func deep(items []int) int {
	total := 0
	if len(items) > 0 {
		for _, x := range items {
			if x > 0 {
				total += x
			}
		}
	}
	return total
}
`)
		facts, err := AnalyzeStructuralFacts(source)
		if err != nil {
			t.Fatalf("AnalyzeStructuralFacts() error = %v", err)
		}
		if len(facts.Functions) != 1 {
			t.Fatalf("len(Functions) = %d, want 1", len(facts.Functions))
		}
		if got := facts.Functions[0].MaxNestingDepth; got != 3 {
			t.Errorf("MaxNestingDepth = %d, want 3", got)
		}
	})

	t.Run("multiple functions and one method all captured in order", func(t *testing.T) {
		source := []byte(`package main

func first() int {
	return 1
}

func second() int {
	return 2
}

type T struct{}

func (t T) Method() int {
	return 3
}
`)
		facts, err := AnalyzeStructuralFacts(source)
		if err != nil {
			t.Fatalf("AnalyzeStructuralFacts() error = %v", err)
		}
		wantNames := []string{"first", "second", "Method"}
		if len(facts.Functions) != len(wantNames) {
			t.Fatalf("len(Functions) = %d, want %d", len(facts.Functions), len(wantNames))
		}
		for i, want := range wantNames {
			if facts.Functions[i].Name != want {
				t.Errorf("Functions[%d].Name = %q, want %q", i, facts.Functions[i].Name, want)
			}
		}
	})

	t.Run("syntactically invalid Go returns an error", func(t *testing.T) {
		source := []byte(`package main

func broken( {
`)
		_, err := AnalyzeStructuralFacts(source)
		if err == nil {
			t.Fatal("AnalyzeStructuralFacts() error = nil, want non-nil for invalid syntax")
		}
	})

	t.Run("empty source yields zero functions and no error", func(t *testing.T) {
		facts, err := AnalyzeStructuralFacts([]byte{})
		if err != nil {
			t.Fatalf("AnalyzeStructuralFacts() error = %v, want nil", err)
		}
		if len(facts.Functions) != 0 {
			t.Errorf("len(Functions) = %d, want 0", len(facts.Functions))
		}
	})
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/agent/... -run TestAnalyzeStructuralFacts -v`
Expected: FAIL — `undefined: AnalyzeStructuralFacts` (compile error, the
type/function don't exist yet).

- [ ] **Step 4: Write the implementation**

Create `internal/agent/structuralfacts.go`:

```go
package agent

import (
	"fmt"

	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// FunctionFacts describes one Go function or method extracted from a
// tree-sitter parse of a reviewed draft.
type FunctionFacts struct {
	Name            string
	StartLine       int // 1-indexed
	EndLine         int
	LengthLines     int
	MaxNestingDepth int
}

// StructuralFacts is populated by AnalyzeStructuralFacts and consumed by
// BuildReviewPrompt to give the Review agent real, parser-verified facts
// about the reviewed draft instead of inferring structure from diff text
// alone (see docs/adr/0011-go-structural-facts-via-gotreesitter.md).
type StructuralFacts struct {
	Functions []FunctionFacts
}

// nestingNodeTypes are the tree-sitter Go grammar node types that count
// toward a function's nesting depth, confirmed against real .SExpr() parse
// output (docs/superpowers/specs/2026-07-07-structural-facts-tree-sitter-design.md).
var nestingNodeTypes = map[string]bool{
	"if_statement":                true,
	"for_statement":               true,
	"expression_switch_statement": true,
	"type_switch_statement":       true,
	"select_statement":            true,
}

// AnalyzeStructuralFacts parses source as Go and extracts per-function
// structural facts. Empty source yields a zero-value StructuralFacts and a
// nil error (nothing to analyze yet, not a failure). Source with syntax
// errors returns a non-nil error — callers must fail open (pass nil facts
// to BuildReviewPrompt) rather than block the Review loop on a parse
// failure.
func AnalyzeStructuralFacts(source []byte) (*StructuralFacts, error) {
	lang := grammars.GoLanguage()
	tree, err := gotreesitter.NewParser(lang).Parse(source)
	if err != nil {
		return nil, fmt.Errorf("agent: parse structural facts: %w", err)
	}

	root := tree.RootNode()
	if root == nil {
		return &StructuralFacts{}, nil
	}
	if root.HasError() {
		return nil, fmt.Errorf("agent: source has syntax errors, cannot extract structural facts")
	}

	facts := &StructuralFacts{}
	gotreesitter.Walk(root, func(n *gotreesitter.Node, depth int) gotreesitter.WalkAction {
		nodeType := n.Type(lang)
		if nodeType != "function_declaration" && nodeType != "method_declaration" {
			return gotreesitter.WalkContinue
		}

		nameNode := n.ChildByFieldName("name", lang)
		if nameNode == nil {
			return gotreesitter.WalkContinue
		}

		startLine := int(n.StartPoint().Row) + 1
		endLine := int(n.EndPoint().Row) + 1
		facts.Functions = append(facts.Functions, FunctionFacts{
			Name:            nameNode.Text(source),
			StartLine:       startLine,
			EndLine:         endLine,
			LengthLines:     endLine - startLine + 1,
			MaxNestingDepth: maxNestingDepth(n, lang),
		})
		return gotreesitter.WalkContinue
	})

	return facts, nil
}

// maxNestingDepth returns the deepest count of nested nesting-construct
// ancestors (if/for/switch/select) reached anywhere inside fn's subtree,
// relative to fn itself — not raw AST depth, which would also count
// non-nesting nodes like block or expression statements.
func maxNestingDepth(fn *gotreesitter.Node, lang *gotreesitter.Language) int {
	max := 0
	gotreesitter.Walk(fn, func(n *gotreesitter.Node, depth int) gotreesitter.WalkAction {
		if n == fn || !nestingNodeTypes[n.Type(lang)] {
			return gotreesitter.WalkContinue
		}
		d := 0
		for anc := n; anc != nil && anc != fn; anc = anc.Parent() {
			if nestingNodeTypes[anc.Type(lang)] {
				d++
			}
		}
		if d > max {
			max = d
		}
		return gotreesitter.WalkContinue
	})
	return max
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/agent/... -run TestAnalyzeStructuralFacts -v`
Expected: PASS, all 5 subtests.

- [ ] **Step 6: Verify no regressions**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: all clean, all existing tests still pass. Do not commit — see
Global Constraints.

---

### Task 2: Wire structural facts into the Review prompt

**Files:**
- Modify: `internal/agent/review.go:16-18` (remove the old empty
  `StructuralFacts` placeholder — it now lives in `structuralfacts.go`),
  `review.go:20-37` (`BuildReviewPrompt`), `review.go:93-94`
  (`newReviewInstructionProvider`)
- Modify: `internal/agent/review_test.go` (add two new test functions)

**Interfaces:**
- Consumes: `FunctionFacts`, `StructuralFacts`, `AnalyzeStructuralFacts`
  from Task 1.
- Produces: `BuildReviewPrompt` renders a "Structural facts" section when
  `facts != nil && len(facts.Functions) > 0`; output is byte-identical to
  today when `facts` is `nil` or has zero functions. This is the final
  observable behavior — no later task depends on anything new here.

- [ ] **Step 1: Write the failing tests**

Add to `internal/agent/review_test.go` (after the existing test functions):

```go
func TestBuildReviewPrompt_IncludesStructuralFactsWhenPresent(t *testing.T) {
	brain := testBrainForReview(t)
	diff := "+func add(a, b int) int { return a + b }"
	facts := &StructuralFacts{
		Functions: []FunctionFacts{
			{Name: "add", StartLine: 1, EndLine: 1, LengthLines: 1, MaxNestingDepth: 0},
		},
	}

	prompt := BuildReviewPrompt(diff, brain, facts)

	if !strings.Contains(prompt, "Structural facts") {
		t.Error("prompt missing structural facts section when facts is non-nil with functions")
	}
	if !strings.Contains(prompt, "add") {
		t.Error("prompt missing function name from structural facts")
	}
}

func TestBuildReviewPrompt_OmitsStructuralFactsWhenNilOrEmpty(t *testing.T) {
	brain := testBrainForReview(t)
	diff := "+func add(a, b int) int { return a + b }"

	nilPrompt := BuildReviewPrompt(diff, brain, nil)
	emptyPrompt := BuildReviewPrompt(diff, brain, &StructuralFacts{})

	if strings.Contains(nilPrompt, "Structural facts") {
		t.Error("prompt should not include structural facts section when facts is nil")
	}
	if nilPrompt != emptyPrompt {
		t.Error("nil facts and empty-Functions facts should produce byte-identical prompts")
	}
}
```

`review_test.go` already imports `strings` and defines `testBrainForReview`
— no new imports needed.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/agent/... -run TestBuildReviewPrompt -v`
Expected: FAIL — `TestBuildReviewPrompt_IncludesStructuralFactsWhenPresent`
fails because today's `BuildReviewPrompt` never renders a "Structural
facts" section for any input.

- [ ] **Step 3: Remove the old placeholder**

In `internal/agent/review.go`, delete lines 16-18:

```go
// StructuralFacts is a reserved seam for a future tree-sitter pre-filter
// stage (deferred — see design spec §7). It carries no fields yet.
type StructuralFacts struct{}
```

(The real `StructuralFacts` type now lives in `structuralfacts.go` from
Task 1 — this file no longer needs its own definition.)

- [ ] **Step 4: Update `BuildReviewPrompt`**

Replace the function (currently `review.go:20-37`, after Step 3's deletion
the line numbers shift up by 3 — locate by function signature, not by
exact line number):

```go
func BuildReviewPrompt(diff string, brain *secondbrain.Brain, facts *StructuralFacts) string {
	var sb strings.Builder
	sb.WriteString("Review the following code change against these Second Brain principles:\n\n")

	matched := matchConceptsInDiff(diff, brain)
	if len(matched) == 0 {
		sb.WriteString("(no specific Second Brain concept matched this diff by keyword — review using general SOLID and clean-code judgment)\n\n")
	}
	for _, c := range matched {
		fmt.Fprintf(&sb, "- %s: %s\n", c.Title, c.Body)
	}

	if facts != nil && len(facts.Functions) > 0 {
		sb.WriteString("\nStructural facts (from tree-sitter parse):\n")
		for _, fn := range facts.Functions {
			fmt.Fprintf(&sb, "- %s (lines %d-%d, %d lines, max nesting depth %d)\n",
				fn.Name, fn.StartLine, fn.EndLine, fn.LengthLines, fn.MaxNestingDepth)
		}
	}

	sb.WriteString("\nCode change:\n")
	sb.WriteString(diff)
	sb.WriteString("\n\nProduce a verdict: APPROVE or CHANGES_REQUESTED, with reasoning tied to the principles above.")

	return sb.String()
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/agent/... -run TestBuildReviewPrompt -v`
Expected: PASS, including the two new tests and every pre-existing
`TestBuildReviewPrompt_*` test (they all pass `nil` for `facts`, which
still produces the same output as before — the new rendering block is
skipped).

- [ ] **Step 6: Wire the live path**

In `newReviewInstructionProvider` (`internal/agent/review.go`), replace the
line `return BuildReviewPrompt(draft, brain, nil) + reviewVerdictInstruction, nil`
with:

```go
			draft, _ := draftVal.(string)
			facts, err := AnalyzeStructuralFacts([]byte(draft))
			if err != nil {
				facts = nil // fail open — never block the Review loop on a parse failure
			}
			return BuildReviewPrompt(draft, brain, facts) + reviewVerdictInstruction, nil
```

- [ ] **Step 7: Verify no regressions**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: all clean, including `internal/agent/eval_test.go` and
`internal/agent/review_bench_test.go`, which both still pass `nil` for
`facts` and must be unaffected. Do not commit — see Global Constraints.

---

### Task 3: ADR-0011 and cross-references

**Files:**
- Create: `docs/adr/0011-go-structural-facts-via-gotreesitter.md`
- Modify: `docs/adr/0007-tree-sitter-deferred.md` (append a cross-reference,
  same pattern ADR-0010 used on ADR-0001)
- Modify: `MEMORY.md` (add a bullet cross-referencing ADR-0011, same
  pattern as the existing A2A bullet)

**Interfaces:** none — this is a documentation-only task, no code
dependencies in either direction.

- [ ] **Step 1: Write ADR-0011**

Create `docs/adr/0011-go-structural-facts-via-gotreesitter.md`:

```markdown
# ADR-0011: Go structural facts via gotreesitter

## Status
Accepted

## Context
ADR-0007 deferred tree-sitter-based structural analysis because the only
known Go tree-sitter bindings at the time
(`tree-sitter/go-tree-sitter`, `smacker/go-tree-sitter`) were CGo-based,
adding a build-toolchain dependency not justified by the MVP's scope. It
explicitly instructed re-verifying that landscape before implementing,
rather than assuming it still held.

Re-verified 2026-07-07: `tree-sitter/go-tree-sitter@v0.25.0` and
`smacker/go-tree-sitter` are both still CGo-based — ADR-0007's original
blocker for those two libraries still holds. `github.com/odvcencio/gotreesitter`
(v0.21.0), which did not exist when ADR-0007 was written, is a pure-Go
tree-sitter runtime with no CGo and no C toolchain requirement, and ships a
Go grammar. This changes ADR-0007's calculus for the specific, narrow use
case of Go-only structural analysis.

## Decision
Lift ADR-0007's deferral specifically for Go-only structural facts
(function boundaries, length, max nesting depth) feeding the Review
agent's already-reserved `structuralFacts` seam
(`internal/agent/review.go`'s `BuildReviewPrompt`), implemented via
`github.com/odvcencio/gotreesitter` in `internal/agent/structuralfacts.go`.

This is not a reversal of ADR-0007's CGo-avoidance reasoning — that
reasoning is preserved by using a pure-Go library instead of the two
originally-considered CGo-based ones. Scope stays locked to exactly the
three facts ADR-0007 named as missing; no graph database, no new MCP
tools, no multi-language support, no visualization, and no whole-repo
indexing are introduced by this decision.

## Consequences
- New direct dependency: `github.com/odvcencio/gotreesitter` +
  `github.com/odvcencio/gotreesitter/grammars`. Pure Go, no new
  build-toolchain requirement — ADR-0007's "keep the build simple"
  consequence still holds.
- The Review agent now receives real, parser-verified structural facts
  instead of inferring function length/nesting/boundaries from diff text
  alone — closing the exact gap ADR-0007's Consequences section named.
- A parse failure (non-Go draft, malformed snippet) fails open to `nil`
  facts, reproducing pre-ADR-0011 behavior exactly — this decision adds no
  new way for the Review loop to block or error out.
- If a future need arises for graph storage, multi-language support, or a
  GitNexus/Graphify-style code-knowledge-graph product, that is explicitly
  a separate, much larger initiative needing its own ADR and design spec —
  not an extension of this one.
```

- [ ] **Step 2: Cross-reference from ADR-0007**

In `docs/adr/0007-tree-sitter-deferred.md`, in the `## Consequences`
section, append a sentence to the existing bullet about re-verifying the
tree-sitter landscape before implementing (same pattern ADR-0010 used to
annotate ADR-0001 — see `docs/adr/0001-adk-go-v2-as-sole-orchestrator.md:30-32`):

```markdown
  eventually implemented. Before implementing, re-verify the state of Go
  tree-sitter bindings rather than assuming today's landscape (CGo-based,
  no confirmed pure-Go option) still holds. **Lifted for Go-only structural
  facts via `gotreesitter` — see [ADR-0011](0011-go-structural-facts-via-gotreesitter.md).**
```

- [ ] **Step 3: Cross-reference from MEMORY.md**

In `MEMORY.md`, add a bullet immediately after the existing Network A2A
bullet (same list, same pattern — see the existing entry citing ADR-0010):

```markdown
- **Go structural facts via tree-sitter (additive)**: the Review agent's
  reserved `structuralFacts` seam is now populated by
  `internal/agent/structuralfacts.go`'s `AnalyzeStructuralFacts`, built on
  the pure-Go `github.com/odvcencio/gotreesitter`. This lifts ADR-0007's
  deferral for this narrow use only — ADR-0007's CGo-avoidance reasoning is
  preserved, not reversed. Fail-open on parse error, byte-identical prompt
  output when no functions are found.
  ([docs/adr/0011](docs/adr/0011-go-structural-facts-via-gotreesitter.md))
```

- [ ] **Step 4: Verify links resolve**

Confirm both new relative links (`0011-go-structural-facts-via-gotreesitter.md`
from ADR-0007, and the `docs/adr/0011-...` link from `MEMORY.md`) point to
the file created in Step 1 with the exact filename used. Do not commit —
see Global Constraints.

---

### Task 4: Final full-repo verification

**Files:** none created or modified — this task only runs commands.

**Interfaces:** none.

- [ ] **Step 1: Full build/vet/test**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: zero errors, zero vet warnings, all tests pass — including every
pre-existing package (`internal/feedback`, `internal/mcpserver`,
`internal/secondbrain`, `cmd/agentic-hooks`) that this plan never touches.

- [ ] **Step 2: Confirm the byte-identical claim empirically**

Run: `go test ./internal/agent/... -run TestBuildReviewPrompt_OmitsStructuralFactsWhenNilOrEmpty -v`
Expected: PASS — this is the test that actually proves `nil` and
zero-functions `facts` produce identical prompt strings, not just an
assertion in the ADR.

- [ ] **Step 3: Report, do not commit**

Report the final `go build`/`go vet`/`go test` output to the user. Per
Global Constraints, do not run `git add`/`git commit` — this project's
standing instruction requires an explicit user ask before any git command.
