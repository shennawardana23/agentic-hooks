---
type: pattern
title: Use Genkit's three middleware hooks for cross-cutting concerns
description: WrapGenerate, WrapModel, and WrapTool each wrap a different layer of the request.
tags: [go, genkit, middleware]
timestamp: 2026-07-04
resource: https://genkit.dev
---

Genkit Go's Middleware V2 exposes three distinct wrap points — pick the one
that matches what you're actually intercepting, don't reach for the
outermost one by default:

- `WrapGenerate` — wraps the whole tool-call loop (multiple model calls plus
  tool executions). Use for end-to-end concerns like overall retry budgets.
- `WrapModel` — wraps a single model API call. Use for logging, metrics, or
  auth on each request to the LLM provider.
- `WrapTool` — wraps a single tool execution. Use for tool-specific retries
  or input/output transformation.

A single middleware can combine pre- and post-processing, logging, metrics,
retries, and authentication — but attach it at the narrowest hook that
covers the concern, so an unrelated layer isn't re-wrapped unnecessarily.
