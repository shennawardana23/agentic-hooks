---
type: principle
title: Sub-agent for delegated reasoning, Tool for a deterministic capability
description: Both attach to an agent.Config, but they solve different problems.
tags: [go, adk, agent-design]
timestamp: 2026-07-04
resource: https://github.com/google/adk-go
---

Give an agent a Tool when the capability is a deterministic, bounded
operation with a fixed input/output contract — a lookup, a calculation, an
API call — and the calling LLM just needs to invoke it and read the result.

Give an agent a SubAgent when the capability itself requires further LLM
reasoning, its own instructions, or its own tools — e.g. a "review" step
that has to weigh multiple principles and produce a verdict, not just
return a fixed-shape answer.

Routing a task that's really "look this up" through a full sub-agent adds
an unnecessary LLM round-trip; routing a task that's really "reason about
this" through a Tool loses the ability for that step to think.
