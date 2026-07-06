---
type: pattern
title: Wrap AI logic in a Genkit Flow, not a bare function
description: Flows get tracing, HTTP exposure, and schema validation for free.
tags: [go, genkit, ai]
timestamp: 2026-07-04
resource: https://genkit.dev
---

Define AI-calling logic with `genkit.DefineFlow(g, name, fn)` rather than a
plain Go function that calls `genkit.Generate` directly. A Flow gets a
stable name for tracing/observability, typed input/output via generics, and
can be exposed as an HTTP endpoint with `genkit.Handler(flow)` with no extra
wiring.

```go
greetingFlow := genkit.DefineFlow(g, "greetingFlow",
    func(ctx context.Context, name string) (string, error) {
        resp, err := genkit.Generate(ctx, g,
            ai.WithPrompt(fmt.Sprintf("Greet %s warmly", name)),
        )
        if err != nil {
            return "", err
        }
        return resp.Text(), nil
    })

mux.HandleFunc("POST /greetingFlow", genkit.Handler(greetingFlow))
```

Keep the Flow function itself thin — request shaping and response
post-processing belong in the Flow body, but business logic unrelated to
the model call belongs in a regular function the Flow calls into.
