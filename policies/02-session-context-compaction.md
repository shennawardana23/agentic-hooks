# Policy Category 02: Session, Context & Compaction

What must survive a context compaction boundary, and when to checkpoint —
distinct from Category 03 (Memory & Persistence), which covers *where*
durable facts live once captured.

## 02.1 — Write durable decisions before compacting
**Rule:** Before compaction occurs, ensure durable decisions and facts are written to a persistent file, not left only in conversation memory.
**Why:** Conversation memory does not survive compaction intact; anything that only exists as unwritten context is at risk of being lost or paraphrased away the moment the window is compacted.
**Enforcement:** advisory-only

## 02.2 — Read project state files first
**Rule:** At the start of a session, read the project's own state files (e.g. this project's SESSION_HANDOFF.md) before assuming anything about prior work.
**Why:** This project maintains SESSION_HANDOFF.md specifically as the deep narrative/session-to-session history, and instructs anyone picking up the project cold to read it first — skipping it means re-deriving context that has already been recorded.
**Enforcement:** advisory-only

## 02.3 — Verify summarized context against real files
**Rule:** Never assume a summarized or compacted context is complete — verify against the actual files if a claim matters.
**Why:** A compaction summary is a lossy approximation of the original conversation; when a claim is load-bearing for a decision, checking the underlying file is cheap insurance against acting on a dropped or distorted detail.
**Enforcement:** advisory-only

## 02.4 — Preserve the why, not just the what
**Rule:** When compacting or summarizing context, preserve the reasoning behind a decision, not just the decision itself.
**Why:** The rationale is what prevents re-litigating settled questions later; a summary that keeps only the "what" invites a future session to reopen a debate that was already resolved for good reasons.
**Enforcement:** advisory-only

## 02.5 — Treat compaction as a checkpoint
**Rule:** Treat a compaction boundary as a checkpoint: confirm build/test state is clean before compaction, and note that state afterward.
**Why:** A clean, recorded build/test state at the checkpoint gives the next session (or the same session post-compaction) a trustworthy baseline to resume from, rather than an unknown starting point.
**Enforcement:** advisory-only

## 02.6 — Checkpoint at natural boundaries, not just the end
**Rule:** Checkpoint long-running tasks into a task list or handoff doc at natural boundaries, not only at the very end.
**Why:** Work that is only recorded at completion is entirely lost if the session is interrupted partway through; periodic checkpoints bound the amount of unrecorded progress at risk at any given time.
**Enforcement:** advisory-only

## 02.7 — Read the record before re-deriving facts
**Rule:** Do not re-derive facts that are already durably recorded — read the record first.
**Why:** Re-deriving a fact wastes effort and risks producing a different, inconsistent answer from the one already on record, undermining the value of having recorded it at all.
**Enforcement:** advisory-only

## 02.8 — Re-state understanding after mid-task compaction
**Rule:** If context was compacted mid-task, re-state your current understanding before continuing, so errors surface immediately rather than silently compounding.
**Why:** An explicit restatement gives a human or reviewer a chance to catch a misunderstanding right away; letting it ride silently means any drift introduced by compaction compounds through every subsequent step.
**Enforcement:** advisory-only

## 02.9 — Re-verify environment state on resume
**Rule:** A "resume" after compaction re-verifies environment/tool state (git status, build health) rather than trusting stale pre-compaction assumptions.
**Why:** Environment and tool state can change independently of the conversation (other processes, other sessions, elapsed time), so assumptions carried over from before compaction may no longer be accurate; a quick re-check is cheaper than acting on stale state.
**Enforcement:** advisory-only

## 02.10 — Update handoff docs in place at milestones
**Rule:** Update handoff docs in place at natural milestones — don't let them go stale while work continues.
**Why:** This project's own SESSION_HANDOFF.md demonstrates the pattern: dated sections are appended in place at milestones (e.g. the "2026-07-06 — Documentation overhaul" entry) while history above is preserved, keeping the doc a reliable, current entry point instead of one that silently drifts out of sync with the codebase.
**Enforcement:** advisory-only
