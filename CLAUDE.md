# agentic-hooks — Read This First

This repo has a collaboration policy. Read [`POLICY.md`](POLICY.md) before
doing anything else, including before touching the Second Brain
(`knowledge/`). The highest-severity rules, so you have them even before
opening that file:

- **Never** print, log, or echo an environment variable known to hold a
  secret, and never write a real secret into a committed file.
  ([07.1](policies/07-env-secrets-protection.md), [07.2](policies/07-env-secrets-protection.md))
- **Never** mark a task complete, a test passing, or a build green without
  having actually run it this session. ([08.1](policies/08-anti-cheating-honesty.md))
- **Never** disable, skip, or delete a failing test to make a suite pass —
  fix the code, or say explicitly the test is wrong and fix the test.
  ([08.3](policies/08-anti-cheating-honesty.md))
- **Stop and ask** before any destructive or hard-to-reverse action
  (force-push, history rewrite, dropping data) ([10.11](policies/10-escalation-hitl.md)),
  and before reopening a locked architectural decision (an ADR, a design
  spec) ([10.3](policies/10-escalation-hitl.md)).
- **Build only what's asked** — no unrequested features or "while I'm
  here" scope creep. ([01.2](policies/01-agent-conduct.md))

Full policy set (~100 items, 10 categories): [`POLICY.md`](POLICY.md).

For this repo's own build/test/run commands and coding conventions:
[`AGENTS.md`](AGENTS.md). For the Second Brain (coding principles this
project's own agents review against): [`knowledge/`](knowledge/).
