---
type: principle
title: Every goroutine needs a clear owner and exit path
description: A goroutine with no way to stop is a leak, not a feature.
tags: [go, concurrency, goroutines, context]
timestamp: 2026-07-04
---

Before starting a goroutine, know who will wait for it to finish
(`sync.WaitGroup`, a result channel, or an errgroup) and what makes it stop
(a `context.Context` cancellation or a closed channel). "Fire and forget"
goroutines that outlive the request that spawned them accumulate as leaks
under load.

Pass `context.Context` as the first parameter to any function that does I/O
or can block, and select on `ctx.Done()` in loops so cancellation actually
propagates instead of leaving the goroutine to finish on its own schedule.

```go
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            process(job)
        }
    }
}
```
