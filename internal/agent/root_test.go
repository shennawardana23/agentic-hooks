package agent

import (
	"testing"

	"google.golang.org/adk/v2/tool"
)

func TestNewRootAgent_WrapsSearchAndLoop(t *testing.T) {
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

	// A no-op search agent is enough here — root construction never invokes it.
	search, err := NewGeneratorAgent(stubModel{})
	if err != nil {
		t.Fatalf("NewGeneratorAgent (search stand-in) error = %v", err)
	}

	root, err := NewRootAgent(search, loop, stubModel{}, nil)
	if err != nil {
		t.Fatalf("NewRootAgent error = %v", err)
	}

	sub := root.SubAgents()
	if len(sub) != 2 {
		t.Fatalf("SubAgents() len = %d, want 2", len(sub))
	}
	if got, want := sub[1].Name(), "self-correcting-loop"; got != want {
		t.Errorf("SubAgents()[1].Name() = %q, want %q", got, want)
	}
}

// fakeTool is the minimal tool.Tool implementation needed to prove
// NewRootAgent merges a supplied []tool.Tool into the agent's Tools
// without needing a real agenttool/remoteagent round trip (that's
// agentcomm_test.go's job) — this test is only about root.go's wiring.
type fakeTool struct{ name string }

func (f fakeTool) Name() string        { return f.name }
func (f fakeTool) Description() string { return "a fake tool for testing" }
func (f fakeTool) IsLongRunning() bool { return false }

func TestNewRootAgent_AcceptsNonNilAgentToolsWithoutError(t *testing.T) {
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
	search, err := NewGeneratorAgent(stubModel{})
	if err != nil {
		t.Fatalf("NewGeneratorAgent (search stand-in) error = %v", err)
	}

	root, err := NewRootAgent(search, loop, stubModel{}, []tool.Tool{fakeTool{name: "remote-echo"}})
	if err != nil {
		t.Fatalf("NewRootAgent() error = %v", err)
	}

	// Construction succeeding with a non-nil agentTools slice, and
	// SubAgents staying at exactly 2 (search + loop, untouched by the new
	// param), is what this task's change is responsible for — the tool
	// actually being invokable end-to-end is proven in agentcomm_test.go.
	sub := root.SubAgents()
	if len(sub) != 2 {
		t.Fatalf("SubAgents() len = %d, want 2 (agentTools must not affect SubAgents)", len(sub))
	}
}
