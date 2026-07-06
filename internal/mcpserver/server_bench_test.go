package mcpserver

import (
	"context"
	"testing"

	"agentic-hooks/internal/secondbrain"
)

func BenchmarkListKnowledgeHandler(b *testing.B) {
	brain, err := secondbrain.Load("../../knowledge")
	if err != nil {
		b.Fatal(err)
	}
	handler := listKnowledgeHandler(brain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := handler(context.Background(), nil, ListKnowledgeInput{}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetKnowledgeHandler(b *testing.B) {
	brain, err := secondbrain.Load("../../knowledge")
	if err != nil {
		b.Fatal(err)
	}
	handler := getKnowledgeHandler(brain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := handler(context.Background(), nil, GetKnowledgeInput{ID: "go/error-handling"}); err != nil {
			b.Fatal(err)
		}
	}
}
