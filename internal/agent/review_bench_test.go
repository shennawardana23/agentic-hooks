package agent

import (
	"testing"

	"agentic-hooks/internal/secondbrain"
)

const benchDiff = "func DoManyThings() { /* validates input, writes to disk, sends " +
	"email, all in one func — violates solid, leaks goroutines, swallows errors */ }"

func BenchmarkBuildReviewPrompt(b *testing.B) {
	brain, err := secondbrain.Load("../../knowledge")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildReviewPrompt(benchDiff, brain, nil)
	}
}

func BenchmarkMatchConceptsInDiff(b *testing.B) {
	brain, err := secondbrain.Load("../../knowledge")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchConceptsInDiff(benchDiff, brain)
	}
}
