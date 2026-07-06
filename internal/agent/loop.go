package agent

import (
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagents/loopagent"
)

// NewSelfCorrectingLoop wraps generator and review in a LoopAgent: each pass
// runs generator (draft/revise) then review (critique) in the same session,
// so generator sees review's prior verdict in conversation history on the
// next pass. review calls exit_loop on APPROVE, which sets Actions.Escalate
// and stops the loop; otherwise it keeps going until maxIterations is hit
// (resilience bound — a non-converging task still returns generator's last
// draft instead of looping forever).
func NewSelfCorrectingLoop(generator, review agent.Agent, maxIterations uint) (agent.Agent, error) {
	return loopagent.New(loopagent.Config{
		AgentConfig: agent.Config{
			Name:        "self-correcting-loop",
			Description: "Iteratively drafts and critiques an answer until the review agent approves or the iteration bound is reached.",
			SubAgents:   []agent.Agent{generator, review},
		},
		MaxIterations: maxIterations,
	})
}
