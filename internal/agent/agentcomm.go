package agent

import (
	"fmt"
	"net/url"

	remoteagent "google.golang.org/adk/v2/agent/remoteagent/v2"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/agenttool"
)

// BuildAgentTools turns each valid RegistryEntry into a callable tool.Tool
// backed by a real A2A remote agent (agenttool.New wrapping
// remoteagent.NewA2A). A per-entry validation or construction failure is
// skipped with a warning, not fatal — the rest of the registry still
// loads. Card fetching itself is deferred to first invocation (see
// remoteagent.AgentCardProvider) — this function makes no network calls.
func BuildAgentTools(entries []RegistryEntry) (tools []tool.Tool, warnings []string) {
	for _, e := range entries {
		if e.Name == "" {
			warnings = append(warnings, fmt.Sprintf("registry entry skipped: empty name (card_url=%q)", e.CardURL))
			continue
		}
		if e.CardURL == "" {
			warnings = append(warnings, fmt.Sprintf("registry entry %q skipped: empty card_url", e.Name))
			continue
		}
		if _, err := url.Parse(e.CardURL); err != nil {
			warnings = append(warnings, fmt.Sprintf("registry entry %q skipped: invalid card_url %q: %v", e.Name, e.CardURL, err))
			continue
		}

		remote, err := remoteagent.NewA2A(remoteagent.A2AConfig{
			Name:              e.Name,
			Description:       e.Description,
			AgentCardProvider: remoteagent.NewAgentCardProvider(e.CardURL),
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("registry entry %q skipped: %v", e.Name, err))
			continue
		}

		tools = append(tools, agenttool.New(remote, nil))
	}
	return tools, warnings
}
