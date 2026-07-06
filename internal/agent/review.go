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
		Name:        "review",
		Model:       m,
		Description: "Reviews a code diff against the Second Brain's coding principles.",
		Instruction: "You review the generator agent's latest draft for adherence to " +
			"the provided principles. Always end with an explicit APPROVE or " +
			"CHANGES_REQUESTED verdict. When your verdict is APPROVE, call the " +
			"exit_loop tool to end the review loop; when it is CHANGES_REQUESTED, " +
			"do not call exit_loop.",
		Tools: []tool.Tool{exitTool},
	})
}
