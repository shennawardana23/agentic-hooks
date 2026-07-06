# Policy Category 03: Memory & Persistence

Where durable facts live once captured, and how to keep them trustworthy
over time — distinct from Category 02, which covers the compaction
boundary itself.

## 03.1 — Separate Durable Facts From Session State
**Rule:** Distinguish durable facts (architecture decisions, locked constraints) from session-scoped context (current task state), and store each in the memory tier appropriate to its lifespan.
**Why:** A fact that outlives the session (e.g. an architectural decision) belongs in a durable record like an ADR, while transient task state belongs in a session handoff; mixing the two makes it hard to tell what's still authoritative once the session ends.
**Enforcement:** advisory-only

## 03.2 — One Canonical Source Per Fact
**Rule:** Don't duplicate the same fact across multiple memory files — write it once in a canonical location and cross-link to it from anywhere else it's relevant.
**Why:** Every duplicate is a second place that can drift out of sync when the fact changes, and a reader has no way to know which copy is current.
**Enforcement:** advisory-only

## 03.3 — Correct In Place, Visibly
**Rule:** Correct stale or wrong memory entries in place with a visible correction note, rather than silently deleting the history.
**Why:** In this project, SESSION_HANDOFF.md originally claimed the code read `os.Getenv("GOOGLE_API_KEY")` in `newDefaultModel`, which was backwards — `newDefaultModel` actually reads `GEMINI_API_KEY` directly. The error was fixed with an inline "correction, 2026-07-06" note pointing to ADR-0006, not by deleting the original wrong entry, so the record of what was believed and why it changed is preserved.
**Enforcement:** advisory-only

## 03.4 — Traceable to a Real Source
**Rule:** Memory content must be traceable to a real source — a file, a commit, or a stated user decision — never invented to fill a gap.
**Why:** Fabricated history is indistinguishable from real history once written down, and later readers (including future agent sessions) will treat it as ground truth they can build on.
**Enforcement:** advisory-only

## 03.5 — Update Before Creating New
**Rule:** Prefer updating an existing memory file over creating a new one that overlaps it.
**Why:** New overlapping files are exactly how the same fact ends up duplicated across the memory store, defeating the single-source-of-truth goal.
**Enforcement:** advisory-only

## 03.6 — Immutable Records Are Append-Only
**Rule:** Treat immutable records as append-only — a reversed decision gets a new record (e.g. a new ADR) that supersedes the old one, not an edit to the original.
**Why:** An ADR documents the reasoning that was valid at a point in time; editing it to reflect a later reversal erases the trail of why the original decision was made and later changed, which is the entire point of keeping ADRs.
**Enforcement:** advisory-only

## 03.7 — Separate Personal and Project Scope
**Rule:** Treat personal/user-level memory and project-level memory as distinct scopes — don't conflate working-style preferences with facts about a specific codebase.
**Why:** A preference like "keep responses terse" applies across every project a user touches, while a fact like "this table is partitioned by hotel_id" is true only of one codebase; collapsing them into one store makes both harder to reuse correctly.
**Enforcement:** advisory-only

## 03.8 — Verify Before Acting
**Rule:** Verify a remembered fact is still true before acting on a recommendation built from it.
**Why:** Memory is a snapshot of a past state, and the underlying system it describes can change after the memory was written; acting on a stale fact without checking it can silently propagate the same error further downstream.
**Enforcement:** advisory-only

## 03.9 — Record Root Cause, Not Just Fix
**Rule:** Record surprising failures and their root cause, not just the fix that resolved them.
**Why:** The fix is tied to a specific version of the code and will eventually rot as the code changes, but the underlying lesson about why the failure happened generalizes and remains useful long after the fix itself is gone.
**Enforcement:** advisory-only

## 03.10 — Mark Obsolete Entries Explicitly
**Rule:** Mark a memory entry obsolete once it's confirmed no longer accurate, rather than leaving conflicting entries active side by side.
**Why:** Two active entries that contradict each other force every future reader to guess which one is current, which is strictly worse than one entry clearly labeled as superseded.
**Enforcement:** advisory-only
