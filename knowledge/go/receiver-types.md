---
type: principle
title: Be consistent about pointer vs value receivers
description: Mixing receiver kinds on the same type is a common source of subtle bugs.
tags: [go, methods, receivers]
timestamp: 2026-07-04
---

Pick pointer receivers if any method needs to mutate the receiver, if the
type is large enough that copying it is wasteful, or if the type contains a
sync.Mutex or other field that must not be copied. Otherwise value receivers
are fine and slightly simpler to reason about.

Once one method on a type uses a pointer receiver, use pointer receivers for
all methods on that type — a mix means some call sites silently operate on a
copy while others mutate the original, and that asymmetry is invisible at
the call site.

```go
// Bad — Inc mutates a copy, does nothing useful.
type Counter struct{ n int }
func (c Counter) Inc()      { c.n++ }
func (c *Counter) Value() int { return c.n }

// Good — consistent pointer receivers.
func (c *Counter) Inc()      { c.n++ }
func (c *Counter) Value() int { return c.n }
```
