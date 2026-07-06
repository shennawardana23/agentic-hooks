# Policy Category 06: Guardrails & Security

General security posture — distinct from Category 07, which covers
environment variables and secrets specifically.

## 06.1 — Untrusted External Input
**Rule:** Treat every external input as untrusted until validated — MCP tool arguments, CLI flags, file contents.
**Why:** Any data that crosses a trust boundary (a tool call from an agent host, a flag passed on the command line, or bytes read from a file) can be malformed, adversarial, or simply unexpected, and code that assumes otherwise is the root cause of most injection and traversal bugs.
**Enforcement:** advisory-only

## 06.2 — Guard Against Injection
**Rule:** Check for injection risks (command injection, path traversal, SQL injection) before shipping code that handles user-controlled strings.
**Why:** This project's own MCP server was verified this session to correctly resist a live path-traversal attempt against `get_knowledge`'s `id` parameter — that successful defense is the bar every input-handling path should meet, not a one-off result to be complacent about.
**Enforcement:** advisory-only

## 06.3 — Never Disable Checks to Pass
**Rule:** Never disable a security check, test, or validation to make something pass — fix the underlying issue.
**Why:** Disabling a check only hides the failure signal; the vulnerability or defect it was catching remains in the system and resurfaces later, usually in production where it is more expensive to fix.
**Enforcement:** advisory-only

## 06.4 — Flag Posture-Weakening Actions
**Rule:** Flag when a requested action would weaken an existing security posture before executing it.
**Why:** A requester may not realize that a convenience change (loosening a permission, widening an allowlist, removing a validation step) reduces the system's overall security; surfacing the trade-off before acting preserves informed consent.
**Enforcement:** advisory-only

## 06.5 — Prefer Least Privilege
**Rule:** Prefer the least-privilege option when a task has multiple ways to accomplish it — read-only investigation before a destructive fix.
**Why:** Starting with the option that has the smallest blast radius (inspecting before mutating, reading before deleting) limits the damage of a mistaken assumption and keeps the action reversible for longer.
**Enforcement:** advisory-only

## 06.6 — Vet New Dependencies
**Rule:** Check a new dependency for known vulnerabilities or abandonment before pulling it in.
**Why:** An unmaintained or vulnerable dependency becomes a permanent liability once it is woven into the codebase, so the cost of checking upfront is far lower than the cost of discovering the problem after adoption.
**Enforcement:** advisory-only

## 06.7 — Report Findings Even If Inconvenient
**Rule:** Report security findings even when they're inconvenient or contradict a prior claim of "done."
**Why:** Suppressing an uncomfortable finding to preserve the appearance of completion trades a known, fixable problem now for an unknown, possibly larger one later.
**Enforcement:** advisory-only

## 06.8 — Don't Roll Custom Crypto
**Rule:** Don't roll a custom crypto/auth/security primitive when a maintained standard library or well-known package exists.
**Why:** Security primitives are notoriously easy to get subtly wrong, and standard libraries have had far more scrutiny, testing, and real-world hardening than a bespoke implementation could receive in the time available.
**Enforcement:** advisory-only

## 06.9 — Verify Security Fixes
**Rule:** Treat a security-relevant fix as needing verification (a test, a live check), not just a code change.
**Why:** A code change that looks correct can still fail to close the actual vulnerability; only an executed test or a live reproduction confirms the fix behaves as intended under the conditions that matter.
**Enforcement:** advisory-only

## 06.10 — When in Doubt, Ask
**Rule:** When in doubt about whether an action is a security risk, treat it as one and ask before proceeding.
**Why:** The cost of a brief pause to confirm is small compared to the cost of an unreviewed action that turns out to have weakened the system's security.
**Enforcement:** advisory-only
