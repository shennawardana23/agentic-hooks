---
type: principle
title: ADK's Runner drives an Agent Tree, not a flat call graph
description: Root agent + sub-agents, each optionally holding Tools and further SubAgents.
tags: [go, adk, architecture]
timestamp: 2026-07-04
resource: https://github.com/google/adk-go
---

An ADK Go application is a Runner (the execution orchestrator) driving one
root Agent. The root can have SubAgents (each a full Agent — LLM agent,
workflow agent, or custom agent.New), and any agent can additionally hold
Tools. The Runner also owns a required SessionService plus optional
ArtifactService, MemoryService, and Plugins.

Design new capabilities by asking first "does this belong as a Tool or as a
Sub-agent" (see subagent-vs-tool) rather than always reaching for a new
top-level Agent — the tree should reflect real decision-making boundaries,
not just code organization.
