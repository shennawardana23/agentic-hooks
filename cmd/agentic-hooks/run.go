package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/plugin"
	"google.golang.org/adk/v2/plugin/retryandreflect"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"

	myagent "agentic-hooks/internal/agent"
	"agentic-hooks/internal/feedback"
	"agentic-hooks/internal/secondbrain"
)

const (
	runAppName           = "agentic-hooks"
	runUserID            = "cli-user"
	defaultMaxIterations = 4
)

func newDefaultModel(ctx context.Context) (model.LLM, error) {
	return gemini.NewModel(ctx, "gemini-flash-latest", &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
}

// runRootAgent drives the runner and streams every sub-agent's output to w
// as it's produced, tagged by event.Author (generator/review/search/root) —
// this is the "TUI CLI" surface: with a self-correcting loop in play, the
// user needs to see each draft/critique pass live, not just a final blob.
// It also returns the full transcript for the HITL prompt and the feedback
// annotator.
func runRootAgent(ctx context.Context, w io.Writer, root agent.Agent, task string) (string, error) {
	sessionService := session.InMemoryService()
	sess, err := sessionService.Create(ctx, &session.CreateRequest{AppName: runAppName, UserID: runUserID})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	resilientPlugin, err := retryandreflect.New()
	if err != nil {
		return "", fmt.Errorf("create retry/reflect plugin: %w", err)
	}

	r, err := runner.New(runner.Config{
		AppName:        runAppName,
		Agent:          root,
		SessionService: sessionService,
		PluginConfig:   runner.PluginConfig{Plugins: []*plugin.Plugin{resilientPlugin}},
	})
	if err != nil {
		return "", fmt.Errorf("create runner: %w", err)
	}

	message := genai.NewContentFromText(task, genai.RoleUser)

	var result string
	for event, err := range r.Run(ctx, runUserID, sess.Session.ID(), message, agent.RunConfig{
		StreamingMode: agent.StreamingModeNone,
	}) {
		if err != nil {
			return "", fmt.Errorf("agent run: %w", err)
		}
		if event.LLMResponse.Content == nil {
			continue
		}
		for _, part := range event.LLMResponse.Content.Parts {
			if part.Text == "" {
				continue
			}
			fmt.Fprintf(w, "[%s] %s\n", event.Author, part.Text)
			result += part.Text
		}
	}

	return result, nil
}

// promptForApprovalAndRecordFeedback prompts on out/in for a HITL
// approve/reject decision plus an optional reason, appends the decision to
// the feedback log, and returns whether it was approved. Approval is
// fail-closed: anything but a literal "y"/"Y" line counts as reject. A
// feedback-write failure only warns to out and does not change the
// returned approval — the human's decision still stands even if the
// annotation couldn't be persisted.
func promptForApprovalAndRecordFeedback(out io.Writer, in *bufio.Reader, feedbackDir, task, result string) bool {
	fmt.Fprintf(out, "\nFinal transcript above. Approve? [y/N]: ")
	approveLine, _ := in.ReadString('\n')
	approved := approveLine == "y\n" || approveLine == "Y\n"

	fmt.Fprint(out, "Reason (optional, for the feedback log): ")
	reasonLine, _ := in.ReadString('\n')
	reason := strings.TrimSuffix(reasonLine, "\n")

	if err := feedback.Append(feedbackDir, feedback.Record{
		Timestamp:  time.Now(),
		Task:       task,
		Transcript: result,
		Approved:   approved,
		Reason:     reason,
	}); err != nil {
		fmt.Fprintf(out, "warning: failed to write feedback record: %v\n", err)
	}

	return approved
}

// loadAgentTools loads the optional --agents-config registry and builds
// the resulting tool.Tool set, printing any per-entry warnings to out. An
// empty path is a no-op (nil tools, no error, no output) — this is the
// default, --agents-config-absent behavior, and must reproduce today's
// exact root-agent construction.
func loadAgentTools(path string, out io.Writer) ([]tool.Tool, error) {
	if path == "" {
		return nil, nil
	}

	entries, err := myagent.LoadRegistry(path)
	if err != nil {
		return nil, err
	}

	tools, warnings := myagent.BuildAgentTools(entries)
	for _, w := range warnings {
		fmt.Fprintf(out, "warning: %s\n", w)
	}
	return tools, nil
}

func newRunCmd() *cobra.Command {
	var knowledgeDir string
	var mcpCommand string
	var mcpArgs []string
	var maxIterations uint
	var feedbackDir string
	var agentsConfigPath string

	cmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run the Search + self-correcting generate/review loop on a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task := args[0]
			ctx := context.Background()

			brain, err := secondbrain.Load(knowledgeDir)
			if err != nil {
				return err
			}

			m, err := newDefaultModel(ctx)
			if err != nil {
				return err
			}

			search, err := myagent.NewSearchAgent(mcpCommand, mcpArgs, m)
			if err != nil {
				return err
			}
			generator, err := myagent.NewGeneratorAgent(m)
			if err != nil {
				return err
			}
			review, err := myagent.NewReviewAgent(m, brain)
			if err != nil {
				return err
			}
			loop, err := myagent.NewSelfCorrectingLoop(generator, review, maxIterations)
			if err != nil {
				return err
			}
			agentTools, err := loadAgentTools(agentsConfigPath, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			root, err := myagent.NewRootAgent(search, loop, m, agentTools)
			if err != nil {
				return err
			}

			result, err := runRootAgent(ctx, cmd.OutOrStdout(), root, task)
			if err != nil {
				return err
			}

			reader := bufio.NewReader(os.Stdin)
			approved := promptForApprovalAndRecordFeedback(cmd.OutOrStdout(), reader, feedbackDir, task, result)

			if !approved {
				fmt.Fprintln(cmd.OutOrStdout(), "Rejected — no output returned as final.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), result)
			return nil
		},
	}

	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", "", "path to the Second Brain knowledge directory (required)")
	cmd.Flags().StringVar(&mcpCommand, "search-mcp-server", "", "command to launch the Search agent's MCP server (required)")
	cmd.Flags().StringSliceVar(&mcpArgs, "search-mcp-server-args", nil, "comma-separated arguments passed to --search-mcp-server")
	cmd.Flags().UintVar(&maxIterations, "max-iterations", defaultMaxIterations, "max generate/review passes before the loop returns its best-effort draft")
	cmd.Flags().StringVar(&feedbackDir, "feedback-dir", "feedback", "directory for the append-only human-feedback JSONL log")
	cmd.Flags().StringVar(&agentsConfigPath, "agents-config", "", "optional path to a YAML registry of remote agents to expose as tools (see internal/agent.RegistryEntry)")
	cmd.MarkFlagRequired("knowledge-dir")
	cmd.MarkFlagRequired("search-mcp-server")

	return cmd
}
