---
type: principle
title: Wrap and propagate errors, never discard them
description: Go has no exceptions — a discarded error is a silently lost failure.
tags: [go, error-handling]
timestamp: 2026-07-04
---

Always check the error return of a call that can fail. Never assign it to
`_` unless the call genuinely cannot fail for reasons documented at the call
site (e.g. `Buffer.Write` on an in-memory buffer).

When propagating an error up the stack, wrap it with `fmt.Errorf("doing X: %w", err)`
so callers can still `errors.Is`/`errors.As` against the original, and so
the resulting error message reads as a call stack, not a bare code.

Don't wrap when there is nothing useful to add — a thin pass-through wrapper
around a single call to another internal function that already wraps
correctly just duplicates context.

```go
// Bad — error silently dropped.
data, _ := os.ReadFile(path)

// Good — checked, wrapped with call-site context.
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("read config %s: %w", path, err)
}
```
