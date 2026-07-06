# How to run benchmarks

Go-native benchmarks, no API cost — they measure `secondbrain.Load`,
prompt construction, and MCP handler performance.

## Run once

```bash
make bench
```

This runs `go test -bench=. -benchmem ./...` across the whole module.

## Compare before/after a change

Capture a baseline before making a change:

```bash
make bench > old.txt
```

Make your change, then capture a new run:

```bash
make bench > new.txt
```

Compare the two with `benchstat`:

```bash
benchstat old.txt new.txt
```

If `benchstat` is not installed:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## What to look for

`benchstat` reports the delta between the two runs per benchmark, with a
significance indicator. A regression worth investigating is a
consistent, statistically significant increase in time/op or bytes/op —
not a single noisy run. If you suspect a regression, use this project's
own `go-bench-runner` subagent, which runs and compares benchmarks for
you and flags real regressions versus noise.
