package agent

import (
	"context"
	"os"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"agentic-hooks/internal/secondbrain"
)

// evalCase is one entry in the offline golden set: a diff paired with the
// verdict and principle keyword a correct Review pass must produce. This is
// the harness SESSION_HANDOFF.md's "Genkit only for offline eval" decision
// pointed at — kept plain Go instead of pulling in Genkit's npm-based eval
// CLI, since no Genkit dependency exists in go.mod and this project prefers
// Go-only tooling.
type evalCase struct {
	name        string
	diff        string
	wantVerdict string
	wantKeyword string // lowercase substring expected in the transcript; empty = skip
}

var evalCases = []evalCase{
	{
		name:        "swallowed_error",
		diff:        `data, _ := os.ReadFile(path); process(data)`,
		wantVerdict: "CHANGES_REQUESTED",
		wantKeyword: "error",
	},
	{
		name:        "goroutine_leak",
		diff:        `go func() { for { doWork() } }()`,
		wantVerdict: "CHANGES_REQUESTED",
		wantKeyword: "goroutine",
	},
	{
		name:        "mixed_receivers",
		diff:        `func (t Thing) Read() string { return t.name }` + "\n" + `func (t *Thing) Write(s string) { t.name = s }`,
		wantVerdict: "CHANGES_REQUESTED",
		wantKeyword: "receiver",
	},
	{
		name:        "concrete_dependency",
		diff:        `func NewService(s *MySQLStore) *Service { return &Service{store: s} }`,
		wantVerdict: "CHANGES_REQUESTED",
		wantKeyword: "interface",
	},
	{
		name:        "clean_add",
		diff:        `func Add(a, b int) int { return a + b }`,
		wantVerdict: "APPROVE",
	},
}

// TestEval_ReviewGoldenSet drives the real Review agent against every
// evalCase and reports a pass rate. Skipped by default — every case is a
// live model call, so this must never run inside `make check`/`go test
// ./...`; opt in explicitly with AGENTIC_HOOKS_EVAL=1 (see `make eval`).
func TestEval_ReviewGoldenSet(t *testing.T) {
	if os.Getenv("AGENTIC_HOOKS_EVAL") == "" {
		t.Skip("set AGENTIC_HOOKS_EVAL=1 to run the real-model eval harness (costs API calls) — see `make eval`")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Fatal("GEMINI_API_KEY or GOOGLE_API_KEY is required for the eval harness")
	}

	ctx := context.Background()
	m, err := gemini.NewModel(ctx, "gemini-flash-latest", &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		t.Fatalf("gemini.NewModel error = %v", err)
	}

	brain, err := secondbrain.Load("../../knowledge")
	if err != nil {
		t.Fatalf("secondbrain.Load error = %v", err)
	}

	passed := 0
	for _, tc := range evalCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			transcript := runReviewOnce(t, ctx, m, brain, tc.diff)

			ok := strings.Contains(transcript, tc.wantVerdict)
			if ok && tc.wantKeyword != "" {
				ok = strings.Contains(strings.ToLower(transcript), tc.wantKeyword)
			}
			if !ok {
				t.Logf("transcript:\n%s", transcript)
				t.Errorf("case %q: want verdict %q (keyword %q) in transcript, not found", tc.name, tc.wantVerdict, tc.wantKeyword)
				return
			}
			passed++
		})
	}
	t.Logf("eval golden set: %d/%d passed", passed, len(evalCases))
}

func runReviewOnce(t *testing.T, ctx context.Context, m model.LLM, brain *secondbrain.Brain, diff string) string {
	t.Helper()

	review, err := NewReviewAgent(m, brain)
	if err != nil {
		t.Fatalf("NewReviewAgent error = %v", err)
	}

	sessionService := session.InMemoryService()
	sess, err := sessionService.Create(ctx, &session.CreateRequest{AppName: "eval_app", UserID: "eval_user"})
	if err != nil {
		t.Fatalf("session create error = %v", err)
	}

	r, err := runner.New(runner.Config{AppName: "eval_app", Agent: review, SessionService: sessionService})
	if err != nil {
		t.Fatalf("runner.New error = %v", err)
	}

	message := genai.NewContentFromText(BuildReviewPrompt(diff, brain, nil), genai.RoleUser)

	var transcript strings.Builder
	for event, err := range r.Run(ctx, "eval_user", sess.Session.ID(), message, agent.RunConfig{StreamingMode: agent.StreamingModeNone}) {
		if err != nil {
			t.Fatalf("agent run: %v", err)
		}
		if event.LLMResponse.Content == nil {
			continue
		}
		for _, part := range event.LLMResponse.Content.Parts {
			transcript.WriteString(part.Text)
		}
	}
	return transcript.String()
}
