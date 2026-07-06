---
type: principle
title: Sub-agents under one Runner.Run call share session history
description: This is what makes a generate/critique loop capable of actually correcting, not just repeating.
tags: [go, adk, session, state]
timestamp: 2026-07-04
resource: https://github.com/google/adk-go
---

Every sub-agent invoked within the same `runner.Run` call sees the same
session's conversation history — a critic's verdict from an earlier pass is
already visible to a generator agent on a later pass with no explicit
"pass this feedback along" plumbing required.

This has a corollary: if you don't want one sub-agent to see another's
output (e.g. two independent proposals that shouldn't anchor on each
other), you need to isolate them — ADK's workflow package exposes branch
isolation for exactly this case. Don't assume isolation by default; the
default is shared visibility.
