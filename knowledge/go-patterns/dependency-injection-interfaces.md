---
type: pattern
title: Inject dependencies as small interfaces, not concrete singletons
description: Makes call sites testable without a DI framework.
tags: [go, patterns, testing, di]
timestamp: 2026-07-04
---

Go doesn't need a DI container — pass dependencies explicitly through
constructors, typed as the small consumer-defined interface the caller
actually needs (see interface-design principle). This is enough to swap a
real Postgres-backed implementation for an in-memory fake in tests, with no
framework.

```go
type UserStore interface {
    Get(ctx context.Context, id string) (User, error)
}

type Handler struct {
    store UserStore
}

func NewHandler(store UserStore) *Handler { return &Handler{store: store} }
```

Reach for a real DI framework only once wiring a large object graph by hand
becomes the actual bottleneck — most Go services never reach that point.
