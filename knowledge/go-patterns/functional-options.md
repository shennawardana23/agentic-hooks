---
type: pattern
title: Functional options for optional constructor parameters
description: Avoids constructor-signature churn as optional parameters grow.
tags: [go, patterns, api-design]
timestamp: 2026-07-04
---

When a constructor has several optional parameters, prefer a variadic
`...Option` parameter over either a long positional signature or a config
struct with exported zero-value-means-default fields — options make
call-site intent explicit and let new options be added without breaking
existing callers.

```go
type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func New(addr string, opts ...Option) *Server {
    s := &Server{addr: addr, timeout: 30 * time.Second}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

Don't reach for this on a constructor with one or two obviously-required
parameters — plain positional arguments are clearer there.
