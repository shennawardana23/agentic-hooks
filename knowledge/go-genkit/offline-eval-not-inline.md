---
type: decision
title: Genkit's role here is offline eval, never inline generation
description: Locked project decision for agentic-hooks — don't re-litigate without new information.
tags: [go, genkit, adk, eval, architecture-decision]
timestamp: 2026-07-04
---

In this repository specifically: ADK Go v2 is the sole runtime for the
request path (Search → Generator/Review loop → HITL). Genkit is not in that
path at all — its only role is offline evaluation, i.e. LLM-as-judge scoring
over exported ADK session traces (and, per the RLHF feedback annotator,
over the human approve/reject JSONL log) run after the fact, not inline
during a user's `run` invocation.

Don't add Genkit to the live request path to get "a flow" for something ADK
already does (ADK has its own agent/tool/session model) — that would mean
running two AI orchestration frameworks for one request. If a future need
requires inline Genkit (e.g. a Genkit-specific plugin with no ADK
equivalent), that's a new decision to make explicitly, not a default.
