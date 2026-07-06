package agent

import (
	"context"
	"iter"

	"google.golang.org/adk/v2/model"
)

// stubModel is the minimum model.LLM implementation needed to exercise
// agent *construction* (Name/SubAgents wiring) without a live API key.
// GenerateContent is never invoked by these tests — only llmagent.New's own
// setup runs.
type stubModel struct{}

func (stubModel) Name() string { return "stub-model" }

func (stubModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {}
}
