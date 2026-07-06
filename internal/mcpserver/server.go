package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"agentic-hooks/internal/secondbrain"
)

type ListKnowledgeInput struct {
	Type string `json:"type,omitempty" jsonschema:"filter by concept type, e.g. principle"`
	Tag  string `json:"tag,omitempty" jsonschema:"filter by tag, e.g. solid"`
}

type ConceptSummary struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type ListKnowledgeOutput struct {
	Concepts []ConceptSummary `json:"concepts"`
}

type GetKnowledgeInput struct {
	ID string `json:"id" jsonschema:"the concept id, i.e. its file path without .md"`
}

type GetKnowledgeOutput struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func listKnowledgeHandler(brain *secondbrain.Brain) func(context.Context, *mcp.CallToolRequest, ListKnowledgeInput) (*mcp.CallToolResult, ListKnowledgeOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListKnowledgeInput) (*mcp.CallToolResult, ListKnowledgeOutput, error) {
		concepts := brain.List(input.Type, input.Tag)
		out := ListKnowledgeOutput{Concepts: make([]ConceptSummary, 0, len(concepts))}
		for _, c := range concepts {
			out.Concepts = append(out.Concepts, ConceptSummary{
				ID: c.ID, Type: c.Type, Title: c.Title, Description: c.Description, Tags: c.Tags,
			})
		}
		return nil, out, nil
	}
}

func getKnowledgeHandler(brain *secondbrain.Brain) func(context.Context, *mcp.CallToolRequest, GetKnowledgeInput) (*mcp.CallToolResult, GetKnowledgeOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetKnowledgeInput) (*mcp.CallToolResult, GetKnowledgeOutput, error) {
		concept, err := brain.Get(input.ID)
		if err != nil {
			return nil, GetKnowledgeOutput{}, fmt.Errorf("mcpserver: %w", err)
		}
		return nil, GetKnowledgeOutput{ID: concept.ID, Title: concept.Title, Body: concept.Body}, nil
	}
}

func NewServer(brain *secondbrain.Brain) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "agentic-hooks-secondbrain", Version: "v0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_knowledge",
		Description: "List Second Brain concepts, optionally filtered by type or tag",
	}, listKnowledgeHandler(brain))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_knowledge",
		Description: "Get the full content of a Second Brain concept by id",
	}, getKnowledgeHandler(brain))

	return server
}
