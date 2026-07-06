package agent

import "testing"

func TestNewGeneratorAgent_SetsName(t *testing.T) {
	gen, err := NewGeneratorAgent(stubModel{})
	if err != nil {
		t.Fatalf("NewGeneratorAgent error = %v", err)
	}
	if got, want := gen.Name(), "generator"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}
