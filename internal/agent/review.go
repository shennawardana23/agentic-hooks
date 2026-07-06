package agent

import (
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/exitlooptool"

	"agentic-hooks/internal/secondbrain"
)

// StructuralFacts is a reserved seam for a future tree-sitter pre-filter
// stage (deferred — see design spec §7). It carries no fields yet.
type StructuralFacts struct{}

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

	sb.WriteString("\nCode change:\n")
	sb.WriteString(diff)
	sb.WriteString("\n\nProduce a verdict: APPROVE or CHANGES_REQUESTED, with reasoning tied to the principles above.")

	return sb.String()
}

// matchConceptsInDiff finds Second Brain concepts whose tag or title keyword
// appears in the diff — the reverse of secondbrain.Brain.Query, which
// checks whether a short topic string appears inside a concept. A full diff
// is rarely a substring of a concept's short body text, so review matching
// needs the opposite direction: does the concept's keyword show up in the
// (much larger) diff.
func matchConceptsInDiff(diff string, brain *secondbrain.Brain) []secondbrain.Concept {
	lowerDiff := strings.ToLower(diff)
	var matched []secondbrain.Concept
	for _, c := range brain.List("", "") {
		if strings.Contains(lowerDiff, strings.ToLower(c.Title)) {
			matched = append(matched, c)
			continue
		}
		for _, tag := range c.Tags {
			if strings.Contains(lowerDiff, strings.ToLower(tag)) {
				matched = append(matched, c)
				break
			}
		}
	}
	return matched
}

// reviewVerdictInstruction is appended to every prompt BuildReviewPrompt
// produces for the live agent — BuildReviewPrompt itself stays verdict-
// agnostic (it's also used directly by review_test.go/eval_test.go, which
// assert on its Second-Brain-matching content, not on loop-control wording).
const reviewVerdictInstruction = "\n\nAlways end with an explicit APPROVE or " +
	"CHANGES_REQUESTED verdict. When your verdict is APPROVE, call the " +
	"exit_loop tool to end the review loop; when it is CHANGES_REQUESTED, " +
	"do not call exit_loop."

// newReviewInstructionProvider reads the Generator's latest draft from
// session state (published under NewGeneratorAgent's OutputKey, "draft")
// and rebuilds the review prompt from it on every loop iteration via
// BuildReviewPrompt — the same Second-Brain-matching logic already covered
// by review_test.go/review_bench_test.go/eval_test.go, now wired into the
// live path instead of only being exercised by tests. This is deliberately
// an InstructionProvider (deterministic, runs before every model call) and
// not a tool the model could choose not to call — a verdict that skips
// matching against the Second Brain is exactly the gap this fixes.
func newReviewInstructionProvider(brain *secondbrain.Brain) func(agent.ReadonlyContext) (string, error) {
	return func(ctx agent.ReadonlyContext) (string, error) {
		draftVal, err := ctx.ReadonlyState().Get("draft")
		if err != nil || draftVal == nil {
			// First-ever pass before the Generator has produced anything yet,
			// or a state-plumbing gap — fail open with general judgment rather
			// than erroring the whole run, matching the pre-fix behavior for
			// diffs that match no concept.
			return "Review the generator agent's latest draft using general " +
				"SOLID and clean-code judgment (no draft found yet in session " +
				"state)." + reviewVerdictInstruction, nil
		}
		draft, _ := draftVal.(string)
		return BuildReviewPrompt(draft, brain, nil) + reviewVerdictInstruction, nil
	}
}

// NewReviewAgent builds the critic half of the self-correcting loop (see
// NewSelfCorrectingLoop). It carries the exit_loop tool: LoopAgent watches
// for exit_loop's Escalate action to stop iterating, so Review must call it
// itself on APPROVE — the loop has no other way to know convergence.
func NewReviewAgent(m model.LLM, brain *secondbrain.Brain) (agent.Agent, error) {
	exitTool, err := exitlooptool.New()
	if err != nil {
		return nil, err
	}

	return llmagent.New(llmagent.Config{
		Name:                "review",
		Model:               m,
		Description:         "Reviews a code diff against the Second Brain's coding principles.",
		InstructionProvider: newReviewInstructionProvider(brain),
		Tools:               []tool.Tool{exitTool},
	})
}
