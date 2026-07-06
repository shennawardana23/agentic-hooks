package agent

import "testing"

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

	root, err := NewRootAgent(search, loop, stubModel{})
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
