---
type: principle
title: Accept interfaces, return concrete types
description: Small, consumer-defined interfaces keep packages loosely coupled.
tags: [go, interfaces, api-design]
timestamp: 2026-07-04
---

Define interfaces at the point of use (the consuming package), sized to
exactly what that consumer needs — often one or two methods. Don't define a
broad interface in the producing package and force every implementation to
satisfy all of it "just in case."

Constructors and functions that produce a value should return the concrete
type, not an interface — that keeps the zero value usable, keeps godoc on
the concrete type discoverable, and doesn't hide fields callers may need
directly. The exception is when a package deliberately hides its
implementation (e.g. returning an `io.Reader` from a decompressor).

```go
// Good — the consumer only needs to read, so it only asks for that.
type Fetcher interface {
    Fetch(ctx context.Context, id string) ([]byte, error)
}

func Summarize(ctx context.Context, f Fetcher, id string) (string, error) { ... }
```
