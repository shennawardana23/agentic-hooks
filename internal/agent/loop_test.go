package agent

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

func TestNewSelfCorrectingLoop_WrapsGeneratorAndReview(t *testing.T) {
	brain := testBrainForReview(t)

	gen, err := NewGeneratorAgent(stubModel{})
	if err != nil {
		t.Fatalf("NewGeneratorAgent error = %v", err)
	}
	review, err := NewReviewAgent(stubModel{}, brain)
	if err != nil {
		t.Fatalf("NewReviewAgent error = %v", err)
	}

	loop, err := NewSelfCorrectingLoop(gen, review, 4)
	if err != nil {
		t.Fatalf("NewSelfCorrectingLoop error = %v", err)
	}

	if got, want := loop.Name(), "self-correcting-loop"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}

	sub := loop.SubAgents()
	if len(sub) != 2 {
		t.Fatalf("SubAgents() len = %d, want 2", len(sub))
	}
	if got, want := sub[0].Name(), "generator"; got != want {
		t.Errorf("SubAgents()[0].Name() = %q, want %q", got, want)
	}
	if got, want := sub[1].Name(), "review"; got != want {
		t.Errorf("SubAgents()[1].Name() = %q, want %q", got, want)
	}
}

// scriptedGeneratorModel records every request it receives and always
// answers with plain draft text (it never calls a tool).
type scriptedGeneratorModel struct {
	calls    int
	requests []*model.LLMRequest
}

func (m *scriptedGeneratorModel) Name() string { return "scripted-generator" }

func (m *scriptedGeneratorModel) GenerateContent(_ context.Context, req *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		m.calls++
		m.requests = append(m.requests, req)
		yield(&model.LLMResponse{
			Content: genai.NewContentFromText(fmt.Sprintf("draft %d", m.calls), genai.RoleModel),
		}, nil)
	}
}

// scriptedReviewModel answers CHANGES_REQUESTED (no tool call) on its first
// invocation, then APPROVE plus an exit_loop function call on its second —
// mirroring exactly how a real model is instructed to behave in review.go.
type scriptedReviewModel struct {
	calls int
}

func (m *scriptedReviewModel) Name() string { return "scripted-review" }

func (m *scriptedReviewModel) GenerateContent(_ context.Context, _ *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		m.calls++
		if m.calls == 1 {
			yield(&model.LLMResponse{
				Content: genai.NewContentFromText("CHANGES_REQUESTED: tighten single responsibility", genai.RoleModel),
			}, nil)
			return
		}
		yield(&model.LLMResponse{
			Content: &genai.Content{
				Role: genai.RoleModel,
				Parts: []*genai.Part{
					genai.NewPartFromText("APPROVE"),
					genai.NewPartFromFunctionCall("exit_loop", map[string]any{}),
				},
			},
		}, nil)
	}
}

// TestSelfCorrectingLoop_ConvergesAfterCorrection is the multi-iteration
// proof the handoff doc flagged as missing: TestNewSelfCorrectingLoop_Wraps*
// above only asserts construction/wiring, never a real generate->critique
// ->revise pass. This drives the actual loop with scripted models (no API
// key) and asserts: (1) the loop stops after exactly 2 iterations instead of
// exhausting maxIterations, proving exit_loop's Escalate action is honored,
// and (2) the generator's second call really receives the review agent's
// first-iteration CHANGES_REQUESTED verdict in its conversation history,
// proving the correction loop can actually correct, not just repeat.
func TestSelfCorrectingLoop_ConvergesAfterCorrection(t *testing.T) {
	brain := testBrainForReview(t)

	genModel := &scriptedGeneratorModel{}
	reviewModel := &scriptedReviewModel{}

	gen, err := NewGeneratorAgent(genModel)
	if err != nil {
		t.Fatalf("NewGeneratorAgent error = %v", err)
	}
	review, err := NewReviewAgent(reviewModel, brain)
	if err != nil {
		t.Fatalf("NewReviewAgent error = %v", err)
	}

	loop, err := NewSelfCorrectingLoop(gen, review, 4)
	if err != nil {
		t.Fatalf("NewSelfCorrectingLoop error = %v", err)
	}

	ctx := context.Background()
	sessionService := session.InMemoryService()
	sess, err := sessionService.Create(ctx, &session.CreateRequest{AppName: "test_app", UserID: "test_user"})
	if err != nil {
		t.Fatalf("session create error = %v", err)
	}

	r, err := runner.New(runner.Config{
		AppName:        "test_app",
		Agent:          loop,
		SessionService: sessionService,
	})
	if err != nil {
		t.Fatalf("runner.New error = %v", err)
	}

	message := genai.NewContentFromText("review: func DoEverything() { ... }", genai.RoleUser)

	var escalated bool
	for event, err := range r.Run(ctx, "test_user", sess.Session.ID(), message, agent.RunConfig{StreamingMode: agent.StreamingModeNone}) {
		if err != nil {
			t.Fatalf("agent run: %v", err)
		}
		if event.Actions.Escalate {
			escalated = true
		}
	}

	if !escalated {
		t.Error("loop never escalated — exit_loop's Escalate action was not observed in any event")
	}
	if reviewModel.calls != 2 {
		t.Errorf("reviewModel.calls = %d, want 2 (loop must stop right after APPROVE+exit_loop, not exhaust maxIterations)", reviewModel.calls)
	}
	if genModel.calls != 2 {
		t.Errorf("genModel.calls = %d, want 2 (one draft, one revision)", genModel.calls)
	}

	if len(genModel.requests) < 2 {
		t.Fatalf("genModel.requests len = %d, want >= 2", len(genModel.requests))
	}
	secondRequest := genModel.requests[1]
	var sawReviewFeedback bool
	for _, c := range secondRequest.Contents {
		for _, p := range c.Parts {
			if strings.Contains(p.Text, "CHANGES_REQUESTED") {
				sawReviewFeedback = true
			}
		}
	}
	if !sawReviewFeedback {
		t.Error("generator's second call never saw the review agent's CHANGES_REQUESTED verdict in conversation history — the loop is not actually self-correcting")
	}
}

// scriptedDraftGeneratorModel always returns the same fixed draft text
// (rather than scriptedGeneratorModel's "draft %d" placeholder) so a test
// can control exactly what Review's InstructionProvider sees as the
// Generator's OutputKey-published draft.
type scriptedDraftGeneratorModel struct {
	draft string
}

func (m *scriptedDraftGeneratorModel) Name() string { return "scripted-draft-generator" }

func (m *scriptedDraftGeneratorModel) GenerateContent(_ context.Context, _ *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(&model.LLMResponse{
			Content: genai.NewContentFromText(m.draft, genai.RoleModel),
		}, nil)
	}
}

// recordingApproveReviewModel always approves+exit_loops on its first call
// (so the loop runs exactly one iteration) but records the full request it
// received, so the test can inspect what InstructionProvider actually built.
type recordingApproveReviewModel struct {
	requests []*model.LLMRequest
}

func (m *recordingApproveReviewModel) Name() string { return "recording-approve-review" }

func (m *recordingApproveReviewModel) GenerateContent(_ context.Context, req *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		m.requests = append(m.requests, req)
		yield(&model.LLMResponse{
			Content: &genai.Content{
				Role: genai.RoleModel,
				Parts: []*genai.Part{
					genai.NewPartFromText("APPROVE"),
					genai.NewPartFromFunctionCall("exit_loop", map[string]any{}),
				},
			},
		}, nil)
	}
}

// TestSelfCorrectingLoop_ReviewGroundedInMatchedSecondBrainConcept is the
// live-wiring proof for the fix to the gap a full-repo audit found: Review's
// verdict must be grounded in a real matched Second Brain concept, not
// free-text critique alone. It drives the actual loop (no API key) with a
// generator draft engineered to match testBrainForReview's fixture concept
// ("Single Responsibility Principle", tag "solid") via OutputKey ("draft") +
// Review's InstructionProvider (see newReviewInstructionProvider in
// review.go), then asserts the review model's system instruction — not just
// BuildReviewPrompt in isolation — actually contains the matched concept's
// title and body text.
func TestSelfCorrectingLoop_ReviewGroundedInMatchedSecondBrainConcept(t *testing.T) {
	brain := testBrainForReview(t)

	genModel := &scriptedDraftGeneratorModel{
		draft: "func DoManyThings() { /* violates solid: validates input, writes to disk, sends email */ }",
	}
	reviewModel := &recordingApproveReviewModel{}

	gen, err := NewGeneratorAgent(genModel)
	if err != nil {
		t.Fatalf("NewGeneratorAgent error = %v", err)
	}
	review, err := NewReviewAgent(reviewModel, brain)
	if err != nil {
		t.Fatalf("NewReviewAgent error = %v", err)
	}

	loop, err := NewSelfCorrectingLoop(gen, review, 4)
	if err != nil {
		t.Fatalf("NewSelfCorrectingLoop error = %v", err)
	}

	ctx := context.Background()
	sessionService := session.InMemoryService()
	sess, err := sessionService.Create(ctx, &session.CreateRequest{AppName: "test_app", UserID: "test_user"})
	if err != nil {
		t.Fatalf("session create error = %v", err)
	}

	r, err := runner.New(runner.Config{
		AppName:        "test_app",
		Agent:          loop,
		SessionService: sessionService,
	})
	if err != nil {
		t.Fatalf("runner.New error = %v", err)
	}

	message := genai.NewContentFromText("review: func DoManyThings() { ... }", genai.RoleUser)

	for _, err := range r.Run(ctx, "test_user", sess.Session.ID(), message, agent.RunConfig{StreamingMode: agent.StreamingModeNone}) {
		if err != nil {
			t.Fatalf("agent run: %v", err)
		}
	}

	if len(reviewModel.requests) != 1 {
		t.Fatalf("reviewModel.requests len = %d, want 1", len(reviewModel.requests))
	}
	sysInstr := reviewModel.requests[0].Config.SystemInstruction
	if sysInstr == nil {
		t.Fatal("review request's Config.SystemInstruction is nil — InstructionProvider was not applied")
	}
	var instrText string
	for _, p := range sysInstr.Parts {
		instrText += p.Text
	}

	if !strings.Contains(instrText, "Single Responsibility Principle") {
		t.Errorf("review instruction = %q, want it to include the matched Second Brain concept's title (proves live wiring, not just BuildReviewPrompt in isolation)", instrText)
	}
	if !strings.Contains(instrText, "A component should have one reason to change.") {
		t.Errorf("review instruction = %q, want it to include the matched concept's body text", instrText)
	}
	if !strings.Contains(instrText, genModel.draft) {
		t.Error("review instruction does not include the generator's actual draft text — OutputKey/state plumbing is not reaching InstructionProvider")
	}
}
