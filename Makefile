BINARY        := agentic-hooks
CMD           := ./cmd/agentic-hooks
BIN_DIR       := bin
KNOWLEDGE_DIR ?= knowledge

VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS    := -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# Reads GEMINI_API_KEY from the shell env; falls back to GOOGLE_API_KEY if
# that's what you have set. cmd/agentic-hooks/run.go only reads
# GEMINI_API_KEY, so `dev` re-exports whichever one you set under that name.
apiKey ?= $(shell echo $${GEMINI_API_KEY:-$$GOOGLE_API_KEY})

.PHONY: help banner build test vet tidy check dev server clean bench eval

.DEFAULT_GOAL := help

banner: ## Show project banner
	@echo ""
	@echo "  agentic-hooks — Second Brain orchestration CLI"
	@echo "  v$(VERSION) | $(GO_VERSION)"
	@echo "  $(BUILD_TIME)"
	@echo ""

help: banner ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Quick start:"
	@echo "  make server                            # start the MCP server over stdio"
	@echo "  make dev TASK=\"review: ...\"             # run the Search+Review loop on a task"

build: ## Build the agentic-hooks binary -> bin/agentic-hooks
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)

test: ## Run all tests verbosely
	go test ./... -v

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go.mod/go.sum
	go mod tidy

check: vet test build ## Run vet + test + build

# dev triggers cmd/agentic-hooks/main.go's `run` subcommand via `go run` —
# the ADK pipeline (root -> Search + Review -> HITL). Still depends on
# `build` because the Search sub-agent needs a real spawnable binary to
# point its MCP client at (--search-mcp-server); that binary is not what
# `go run` executes here, it's a subprocess this target launches internally.
dev: build ## Run the generate/review loop. make dev FILE=path/to/code.go | TASK="..." | (prompts if neither given)
	@if [ -z "$(apiKey)" ]; then echo "GEMINI_API_KEY or GOOGLE_API_KEY is required" >&2; exit 1; fi
	@if [ -n "$(FILE)" ]; then \
		if [ ! -f "$(FILE)" ]; then echo "FILE not found: $(FILE)" >&2; exit 1; fi; \
		task="review: $$(cat "$(FILE)")"; \
	else \
		task="$(TASK)"; \
		if [ -z "$$task" ]; then printf "Task (e.g. review: func DoEverything() {...}): "; read task; fi; \
	fi; \
	if [ -z "$$task" ]; then echo "Task is required" >&2; exit 1; fi; \
	GEMINI_API_KEY=$(apiKey) go run $(CMD) run "$$task" \
		--knowledge-dir $(KNOWLEDGE_DIR) \
		--search-mcp-server $(BIN_DIR)/$(BINARY) \
		--search-mcp-server-args "serve,--knowledge-dir,$(KNOWLEDGE_DIR)"

# server triggers cmd/agentic-hooks/main.go's `serve` subcommand directly
# via `go run` — the MCP server exposing the Second Brain over stdio.
server: ## Start the Second Brain MCP server over stdio
	go run $(CMD) serve --knowledge-dir $(KNOWLEDGE_DIR)

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)

bench: ## Run Go benchmarks (no API cost). Compare runs with benchstat old.txt new.txt
	go test -bench=. -benchmem ./...

# eval drives the real Review agent against internal/agent/eval_test.go's
# golden set — every case is a live model call, so this is opt-in only
# (never part of `check`/`test`) and always requires a real key.
eval: ## Run the real-model golden-set eval (costs API calls)
	@if [ -z "$(apiKey)" ]; then echo "GEMINI_API_KEY or GOOGLE_API_KEY is required" >&2; exit 1; fi
	GEMINI_API_KEY=$(apiKey) AGENTIC_HOOKS_EVAL=1 go test ./internal/agent/... -run TestEval -v
