# Policy Category 07: Env & Secrets Protection

Never exposing or injecting environment data — the most concretely
incident-grounded category in this policy set.

## 07.1 — Never Print Full Secret Values
**Rule:** Never print, log, or echo the full contents of an environment variable known to hold a secret.
**Why:** Secrets that appear in terminal output, logs, or transcripts can be captured, cached, or shared long after the original command runs, turning a routine debugging step into a durable exposure.
**Enforcement:** advisory-only

## 07.2 — Never Commit Real Secret Values
**Rule:** Never write a real secret value into a committed file, including docs, examples, or configs — use placeholders instead.
**Why:** Committed files are replicated across clones, history, and backups indefinitely, so a real secret checked in even once must be treated as permanently exposed and rotated.
**Enforcement:** advisory-only

## 07.3 — Rotate Immediately on Exposure
**Rule:** If a secret is accidentally exposed, recommend rotation immediately and never persist the exposed value anywhere.
**Why:** This project's own history includes a real Gemini API key pasted into a chat transcript in plaintext — the concrete lesson is that once a secret has left its intended storage, the only safe response is treating it as compromised and rotating it, not trying to scrub or hide the exposed copy.
**Enforcement:** advisory-only

## 07.4 — No Full Env Dumps for Debugging
**Rule:** Don't dump full env/printenv output as a debugging shortcut — target the single variable actually needed.
**Why:** A full environment dump exposes every secret present in the session for the sake of inspecting one value, needlessly widening the blast radius of any accidental leak.
**Enforcement:** advisory-only

## 07.5 — Treat .env Files as Sensitive by Default
**Rule:** Treat .env files as sensitive by default — never cat or display their contents unless explicitly asked and necessary.
**Why:** .env files exist specifically to hold credentials outside of version control, so displaying them casually defeats their purpose and risks surfacing secrets in logs or shared output.
**Enforcement:** advisory-only

## 07.6 — Back Up and Diff-Review Credential-Bearing Configs
**Rule:** Back up config files containing real credentials before editing them, and review diffs for accidental credential exposure before sharing them further.
**Why:** This project did exactly this earlier in the session when editing claude_desktop_config.json — a real file containing multiple live API keys and tokens — backing it up first meant the original values could be restored if the edit went wrong, and diff review before sharing caught what would otherwise have been exposed.
**Enforcement:** advisory-only

## 07.7 — Document Canonical vs. Fallback Env Vars
**Rule:** Document which environment variable name is canonical and which is a fallback when more than one could satisfy a requirement — don't leave it ambiguous.
**Why:** This project's own GEMINI_API_KEY-canonical-vs-GOOGLE_API_KEY-fallback decision led to a real documentation contradiction that had to be corrected, showing that undocumented precedence between env vars creates confusion about which one actually governs behavior.
**Enforcement:** advisory-only

## 07.8 — Avoid Secrets as CLI Arguments
**Rule:** Don't pass secrets as CLI arguments when an environment variable or file-based option is available.
**Why:** CLI arguments are visible in process listings (e.g. `ps`) to any other user or process on the same machine, making them a far less contained channel than environment variables or credential files.
**Enforcement:** advisory-only

## 07.9 — Prefer Indirection Over Raw Secret Arguments
**Rule:** Don't accept a raw secret as a plain tool/MCP argument if a safer indirection exists.
**Why:** Passing a secret directly as a plain argument places it in call logs, tool-call histories, and traces; routing through a reference (such as an env var name or a secret store lookup) avoids that duplication.
**Enforcement:** advisory-only

## 07.10 — Flag, Don't Silently Route Around, Secret Leak Paths
**Rule:** Flag, don't silently work around, any code path that would cause a secret to flow somewhere it doesn't need to reach — a log file, a committed artifact, a third-party API call.
**Why:** Silently rerouting around a leak path hides the underlying design flaw instead of fixing it, leaving the same risk in place for the next person or the next run; surfacing it lets the actual exposure be addressed.
**Enforcement:** advisory-only
