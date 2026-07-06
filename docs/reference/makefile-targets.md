# Makefile Targets Reference

Diátaxis quadrant: **Reference**. Exhaustive list of every target in the
repo-root `Makefile`. Run `make help` for the live, always-current version
of this table (it's generated from the same `.PHONY` comments this
document transcribes).

## Variables

| Variable | Default | Purpose |
|---|---|---|
| `BINARY` | `agentic-hooks` | Output binary name. |
| `CMD` | `./cmd/agentic-hooks` | Go package path built. |
| `BIN_DIR` | `bin` | Output directory for the built binary. |
| `KNOWLEDGE_DIR` | `knowledge` | Overridable via `make <target> KNOWLEDGE_DIR=...`. |
| `VERSION` | `git describe --tags --always --dirty`, or `dev` if that fails | Baked into the binary via `-ldflags`. |
| `BUILD_TIME` | `date +%Y-%m-%dT%H:%M:%S%z` | Baked into the binary via `-ldflags`. |
| `GO_VERSION` | `go version` output, awk'd to the version token | Baked into the binary via `-ldflags`. |
| `apiKey` | `$GEMINI_API_KEY`, falling back to `$GOOGLE_API_KEY` | Used by `dev` and `eval` targets. |

## Targets

| Target | What it runs | Reads |
|---|---|---|
| `help` (default goal) | Prints the banner plus every `##`-commented target, and a "Quick start" hint. | — |
| `banner` | Prints project name/version/Go-version/build-time. | `VERSION`, `GO_VERSION`, `BUILD_TIME` |
| `build` | `go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)` | — |
| `test` | `go test ./... -v` | — |
| `vet` | `go vet ./...` | — |
| `tidy` | `go mod tidy` | — |
| `check` | Runs `vet`, `test`, `build` in sequence. | — |
| `dev` | Depends on `build`. Runs the Search+Review loop via `go run $(CMD) run "$$task" --knowledge-dir $(KNOWLEDGE_DIR) --search-mcp-server $(BIN_DIR)/$(BINARY) --search-mcp-server-args "serve,--knowledge-dir,$(KNOWLEDGE_DIR)"`. | `FILE` (reads and wraps as `review: $(cat FILE)"`) or `TASK` (used verbatim); prompts interactively if neither is set; fails fast with a clear message if the resolved task is still empty. Requires `apiKey` to be non-empty or fails fast. |
| `server` | `go run $(CMD) serve --knowledge-dir $(KNOWLEDGE_DIR)` | `KNOWLEDGE_DIR` |
| `clean` | `rm -rf $(BIN_DIR)` | — |
| `bench` | `go test -bench=. -benchmem ./...` | — (no API key, no network) |
| `eval` | `GEMINI_API_KEY=$(apiKey) AGENTIC_HOOKS_EVAL=1 go test ./internal/agent/... -run TestEval -v` | Requires `apiKey` to be non-empty or fails fast. Real model calls — costs money. Opt-in only, never part of `check`. |

## Usage examples

```bash
make build                                    # bin/agentic-hooks
make server                                   # MCP server over stdio, knowledge/ dir
make server KNOWLEDGE_DIR=/tmp/other-brain    # MCP server over a different dir
make dev TASK="review: func Foo() {}"         # one-off task
make dev FILE=path/to/code.go                 # review a real file's contents
make dev                                      # prompts interactively for a task
make check                                    # vet + test + build
make bench                                    # no API cost
make eval                                     # real API calls, costs money
```

## Related

- [`docs/reference/cli.md`](cli.md) — the underlying `agentic-hooks` binary flags these targets wrap.
- [`docs/how-to/run-benchmarks.md`](../how-to/run-benchmarks.md)
- [`docs/how-to/run-the-golden-set-eval.md`](../how-to/run-the-golden-set-eval.md)
