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
//
// OutputKey publishes the draft into session state under "draft" as soon as
// this agent's event is appended — before Review runs in the same loop
// iteration (see llmagent's OutputKey handling and session.AppendEvent's
// state-delta merge). Review's InstructionProvider (review.go) reads it from
// there to ground its verdict in a real Second Brain concept match instead
// of free-text critique alone.
func NewGeneratorAgent(m model.LLM) (agent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "generator",
		Model:       m,
		Description: "Drafts and revises a text answer to the task.",
		Instruction: "Draft a concise answer to the task. If the conversation " +
			"history contains a CHANGES_REQUESTED verdict from the review agent, " +
			"revise your previous draft to address every point raised instead of " +
			"starting over. Output only the current best draft, nothing else.",
		OutputKey: "draft",
	})
}
