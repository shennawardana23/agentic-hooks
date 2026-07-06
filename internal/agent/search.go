package agent

import (
	"os/exec"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/mcptoolset"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewSearchAgent(mcpCommand string, mcpArgs []string, m model.LLM) (agent.Agent, error) {
	toolset, err := mcptoolset.New(mcptoolset.Config{
		Transport: &mcp.CommandTransport{Command: exec.Command(mcpCommand, mcpArgs...)},
	})
	if err != nil {
		return nil, err
	}

	return llmagent.New(llmagent.Config{
		Name:        "search",
		Model:       m,
		Description: "Looks up external information via a configured MCP tool server.",
		Instruction: "Use the available tools to find information relevant to the task. Summarize findings concisely.",
		Toolsets:    []tool.Toolset{toolset},
	})
}
