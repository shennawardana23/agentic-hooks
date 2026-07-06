---
type: pattern
title: Table-driven tests for multiple input/output cases
description: One test function, many cases, via t.Run subtests.
tags: [go, testing, patterns]
timestamp: 2026-07-04
---

When a function has several input/output cases worth covering, express them
as a slice of structs and drive one loop with `t.Run(tc.name, ...)`, rather
than writing a separate `Test...` function per case or repeating the same
assertions inline for each case.

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        want     int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
        {"zero", 0, 0, 0},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            if got := Add(tc.a, tc.b); got != tc.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
            }
        })
    }
}
```

Named subtests also mean `go test -run TestAdd/negative` can target one case.
