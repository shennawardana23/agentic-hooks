package agent

import (
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
)

// NewGeneratorAgent builds the draft/revise half of the self-correcting
// loop. It shares a session with the Review agent (see NewSelfCorrectingLoop):
// on iteration N>1 its own prior draft and Review's CHANGES_REQUESTED
// critique are already in conversation history, so the instruction below
// only needs to tell it to look there rather than being handed the critique
// explicitly.
func NewGeneratorAgent(m model.LLM) (agent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "generator",
		Model:       m,
		Description: "Drafts and revises a text answer to the task.",
		Instruction: "Draft a concise answer to the task. If the conversation " +
			"history contains a CHANGES_REQUESTED verdict from the review agent, " +
			"revise your previous draft to address every point raised instead of " +
			"starting over. Output only the current best draft, nothing else.",
	})
}
