package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"agentic-hooks/internal/mcpserver"
	"agentic-hooks/internal/secondbrain"
)

func newServeCmd() *cobra.Command {
	var knowledgeDir string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the Second Brain as an MCP server over stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			brain, err := secondbrain.Load(knowledgeDir)
			if err != nil {
				return err
			}
			server := mcpserver.NewServer(brain)
			return server.Run(context.Background(), &mcp.StdioTransport{})
		},
	}

	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", "", "path to the Second Brain knowledge directory (required)")
	cmd.MarkFlagRequired("knowledge-dir")

	return cmd
}
