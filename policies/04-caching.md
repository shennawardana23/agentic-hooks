# Policy Category 04: Caching

What's safe to reuse without re-checking, and what must be re-verified —
covers both literal caches (prompt cache, subagent results) and informal
ones (a fact you "already know" from earlier in a session).

## 04.1 — Never Cache Security-Sensitive Values
**Rule:** Never cache a security-sensitive value (API key, token, credential) beyond the single operation that needs it.
**Why:** A credential held in memory or a cache longer than the operation that consumes it is an extra place it can leak from — through a log, a crash dump, or a later session that reuses stale state — with no corresponding benefit once that operation completes.
**Enforcement:** advisory-only

## 04.2 — Invalidate Cached Verification On Code Change
**Rule:** Treat any cached research/verification result as stale once the underlying code it describes has changed.
**Why:** A verification result is only true of the code as it existed at the moment it was checked; once that code changes, the result no longer describes reality and continuing to rely on it silently reintroduces the exact risk the verification was meant to eliminate.
**Enforcement:** advisory-only

## 04.3 — Re-Verify Cheap Facts Over Stale Expensive Ones
**Rule:** Re-verify a cheap fact rather than trusting an expensive-but-old cached answer when correctness matters more than cost.
**Why:** When the cost of being wrong exceeds the cost of re-checking, the cache's speed advantage stops mattering — correctness should drive the tradeoff, not sunk cost in the original expensive lookup.
**Enforcement:** advisory-only

## 04.4 — Cache Expensive Lookups With Source And Date
**Rule:** Cache expensive, stable lookups explicitly, with the source and date noted, so staleness is checkable later.
**Why:** This project's own `adk-api-verifier` agent exists specifically to produce checkable, sourced verification of Go ADK API behavior instead of relying on memory of a bleeding-edge library — a cached lookup without a source and date can't be checked for staleness and degrades into an unverifiable claim.
**Enforcement:** advisory-only

## 04.5 — Don't Leak Context Across Unrelated Tasks
**Rule:** Don't let cached context from a prior unrelated task leak into a new task's assumptions.
**Why:** A fact, constraint, or decision that was true for one task's scope may not hold for a different task, and carrying it over unexamined can silently narrow or misdirect the new work based on premises no one actually re-checked.
**Enforcement:** advisory-only

## 04.6 — Verify Subagent Output, Not Just Its Report
**Rule:** Verify a subagent's or workflow's cached/earlier output before building on it, not just its claimed success.
**Why:** In this project, an ADR-writing fork's first attempt silently produced zero files despite reporting success — the failure was caught only by checking disk directly, not by trusting the subagent's own report of what it did.
**Enforcement:** advisory-only

## 04.7 — Prompt Caching Is Performance-Only
**Rule:** Session-level prompt caching is a performance concern only — never let it substitute for re-reading a file whose freshness matters.
**Why:** Prompt caching exists to reduce latency and cost for unchanged context, not to assert that the underlying file is still current; treating a cache hit as proof of freshness conflates two unrelated properties.
**Enforcement:** advisory-only

## 04.8 — "It Works" Claims Need Independent Confirmation
**Rule:** A cached "it works" claim from a subagent is not verified until independently confirmed (build/test/manual check).
**Why:** A subagent's self-report reflects what it believes happened, not an outside check on what actually happened; treating the claim itself as verification removes the independent check that would catch a false positive.
**Enforcement:** advisory-only

## 04.9 — State When A Cached Fact Was Confirmed
**Rule:** State when a cached fact was last confirmed true — cache invalidation should be explicit, not implicit.
**Why:** Without an explicit confirmation date, a reader has no way to judge whether a cached fact is still trustworthy or has quietly gone stale, which pushes staleness detection onto guesswork instead of a checkable timestamp.
**Enforcement:** advisory-only

## 04.10 — Confirm Project State Before Reusing A Cached Plan
**Rule:** Don't reuse a cached plan or design across sessions without confirming project state still matches its assumptions.
**Why:** A plan is only valid relative to the state of the codebase and requirements it was written against; if that state has moved on, executing the old plan unmodified can apply decisions that no longer fit the current reality.
**Enforcement:** advisory-only
