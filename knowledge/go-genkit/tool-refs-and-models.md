---
type: pattern
title: Pass tools and models as typed refs, not ad hoc strings
description: ai.ToolRef and ai.ModelRef give compile-time checking Genkit's string-based APIs lack.
tags: [go, genkit, tools]
timestamp: 2026-07-04
resource: https://genkit.dev
---

When a Flow needs tools or a specific model, thread them in as
`[]ai.ToolRef` / `ai.ModelRef` parameters at construction time rather than
looking them up by string name inside the Flow body — this keeps the
dependency visible in the function signature and catches a missing
tool/model registration at wiring time, not at first invocation.

```go
func NewOperatingSystemFlow(g *genkit.Genkit, tools []ai.ToolRef) *core.Flow[string, string, struct{}] {
    return genkit.DefineFlow(g, "operatingSystemFlow", func(ctx context.Context, userRequest string) (string, error) {
        resp, err := genkit.Generate(ctx, g,
            ai.WithPrompt("The user wants to: %s", userRequest),
            ai.WithTools(tools...),
        )
        if err != nil {
            return "", fmt.Errorf("failed to generate response: %w", err)
        }
        return resp.Text(), nil
    })
}
```
