package agent

import (
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
)

// NewRootAgent coordinates a one-shot Search lookup with the self-correcting
// generator/review loop (see NewSelfCorrectingLoop). loop is itself an
// agent.Agent — LoopAgent implements the same interface as any other
// sub-agent — so root can delegate to it exactly like search.
//
// agentTools is an optional set of additional tools (typically
// agenttool-wrapped remote agents built by BuildAgentTools) the root LLM
// can choose to call. nil/empty reproduces the exact behavior of the
// pre-agentTools NewRootAgent — this parameter is purely additive.
func NewRootAgent(search, loop agent.Agent, m model.LLM, agentTools []tool.Tool) (agent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "root",
		Model:       m,
		Description: "Coordinates Search and the self-correcting generate/review loop to handle a task.",
		Instruction: "Delegate lookups to the search agent first if the task needs " +
			"external information, then delegate to the loop agent to draft and " +
			"refine a final answer. Always surface the loop's final approved " +
			"answer and the review agent's verdict to the user.",
		SubAgents: []agent.Agent{search, loop},
		Tools:     agentTools,
	})
}
