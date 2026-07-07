package agent

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

// echoExecutor is a minimal a2asrv.AgentExecutor: it always replies with a
// single fixed text message, regardless of the incoming request. Enough to
// prove one full round trip through the real A2A wire protocol.
type echoExecutor struct{}

func (echoExecutor) Execute(ctx context.Context, ec *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		yield(a2a.NewMessage(a2a.MessageRoleAgent, a2a.NewTextPart("hello from remote")), nil)
	}
}

func (echoExecutor) Cancel(ctx context.Context, ec *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {}
}

// runnableTool is the unexported surface agenttool's returned tool.Tool
// actually implements (tool.Tool itself only has Name/Description/
// IsLongRunning — no call method). Declared locally: Go interface
// satisfaction is structural, so this works across package boundaries.
type runnableTool interface {
	Run(ctx agent.Context, args any) (map[string]any, error)
}

func TestBuildAgentTools_InvokesRealAgentOverA2AWireProtocol(t *testing.T) {
	mux := http.NewServeMux()

	var cardURL string
	mux.HandleFunc("/.well-known/agent-card.json", func(w http.ResponseWriter, r *http.Request) {
		card := a2a.AgentCard{
			Name: "remote-echo",
			SupportedInterfaces: []*a2a.AgentInterface{
				a2a.NewAgentInterface(cardURL, a2a.TransportProtocolJSONRPC),
			},
			Capabilities: a2a.AgentCapabilities{Streaming: true},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(card); err != nil {
			t.Errorf("encode agent card: %v", err)
		}
	})
	mux.Handle("/", a2asrv.NewJSONRPCHandler(a2asrv.NewHandler(echoExecutor{})))

	server := httptest.NewServer(mux)
	defer server.Close()
	cardURL = server.URL

	tools, warnings := BuildAgentTools([]RegistryEntry{
		{Name: "remote-echo", Description: "A test remote agent.", CardURL: server.URL},
	})
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1", len(tools))
	}
	remoteTool, ok := tools[0].(runnableTool)
	if !ok {
		t.Fatalf("tools[0] (%T) does not implement runnableTool", tools[0])
	}

	var gotResult map[string]any
	var runErr error
	harness, err := agent.New(agent.Config{
		Name: "harness",
		Run: func(ic agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				toolCtx := agent.NewToolContext(ic, "", &session.EventActions{}, nil)
				gotResult, runErr = remoteTool.Run(toolCtx, map[string]any{"request": "hi"})
				yield(&session.Event{}, nil)
			}
		},
	})
	if err != nil {
		t.Fatalf("agent.New(harness) error = %v", err)
	}

	ctx := context.Background()
	sessionService := session.InMemoryService()
	sess, err := sessionService.Create(ctx, &session.CreateRequest{AppName: "test", UserID: "u"})
	if err != nil {
		t.Fatalf("sessionService.Create() error = %v", err)
	}
	r, err := runner.New(runner.Config{AppName: "test", Agent: harness, SessionService: sessionService})
	if err != nil {
		t.Fatalf("runner.New() error = %v", err)
	}

	message := genai.NewContentFromText("go", genai.RoleUser)
	for _, runErr2 := range r.Run(ctx, "u", sess.Session.ID(), message, agent.RunConfig{
		StreamingMode: agent.StreamingModeNone,
	}) {
		if runErr2 != nil {
			t.Fatalf("runner.Run() error = %v", runErr2)
		}
	}

	if runErr != nil {
		t.Fatalf("remoteTool.Run() error = %v", runErr)
	}
	if gotResult == nil {
		t.Fatal("remoteTool.Run() returned nil result")
	}
}

func TestBuildAgentTools_WrapsValidEntriesAsTools(t *testing.T) {
	entries := []RegistryEntry{
		{Name: "example-agent", Description: "An example.", CardURL: "http://localhost:9003"},
	}

	tools, warnings := BuildAgentTools(entries)

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1", len(tools))
	}
	if got := tools[0].Name(); got != "example-agent" {
		t.Errorf("tools[0].Name() = %q, want %q", got, "example-agent")
	}
}

func TestBuildAgentTools_SkipsEmptyNameWithWarning(t *testing.T) {
	entries := []RegistryEntry{
		{Name: "", Description: "no name", CardURL: "http://localhost:9003"},
		{Name: "valid-agent", Description: "fine", CardURL: "http://localhost:9004"},
	}

	tools, warnings := BuildAgentTools(entries)

	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1 (only the valid entry)", len(tools))
	}
	if len(warnings) != 1 {
		t.Fatalf("len(warnings) = %d, want 1", len(warnings))
	}
}

func TestBuildAgentTools_SkipsEmptyCardURLWithWarning(t *testing.T) {
	entries := []RegistryEntry{
		{Name: "no-url-agent", Description: "missing url", CardURL: ""},
	}

	tools, warnings := BuildAgentTools(entries)

	if len(tools) != 0 {
		t.Errorf("len(tools) = %d, want 0", len(tools))
	}
	if len(warnings) != 1 {
		t.Fatalf("len(warnings) = %d, want 1", len(warnings))
	}
}

func TestBuildAgentTools_SkipsUnparseableCardURLWithWarning(t *testing.T) {
	entries := []RegistryEntry{
		{Name: "bad-url-agent", Description: "bad url", CardURL: "http://[::1]:namedport"},
	}

	tools, warnings := BuildAgentTools(entries)

	if len(tools) != 0 {
		t.Errorf("len(tools) = %d, want 0", len(tools))
	}
	if len(warnings) != 1 {
		t.Fatalf("len(warnings) = %d, want 1", len(warnings))
	}
}

func TestBuildAgentTools_EmptyRegistryReturnsEmptyNoWarnings(t *testing.T) {
	tools, warnings := BuildAgentTools(nil)
	if len(tools) != 0 || len(warnings) != 0 {
		t.Errorf("BuildAgentTools(nil) = (%v, %v), want (empty, empty)", tools, warnings)
	}
}
