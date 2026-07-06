---
type: pattern
title: Thread context.Context through the whole call chain
description: A context created deep in a call stack cannot carry cancellation or deadlines from the caller.
tags: [go, context, patterns]
timestamp: 2026-07-04
---

`context.Context` must be the first parameter of any function in the call
chain that does I/O, and it must be the one the caller passed in — not a
fresh `context.Background()` created partway down. Creating a new
background context mid-chain silently breaks cancellation propagation: the
caller's timeout or Ctrl-C no longer reaches that branch of work.

Don't store a Context in a struct field for later reuse — pass it explicitly
per call, since a stored context can outlive the request it was scoped to.

```go
// Bad — caller's cancellation never reaches the DB call.
func (s *Service) Get(id string) (Row, error) {
    return s.db.QueryRowContext(context.Background(), query, id)
}

// Good — caller's ctx flows all the way through.
func (s *Service) Get(ctx context.Context, id string) (Row, error) {
    return s.db.QueryRowContext(ctx, query, id)
}
```
