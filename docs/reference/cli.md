# CLI Reference

Diátaxis quadrant: **Reference**. Exhaustive, structural, no prose
justification — see [`docs/explanation/architecture-overview.md`](../explanation/architecture-overview.md)
for the "why" behind these commands.

Binary name: `agentic-hooks`. Verified against `./bin/agentic-hooks --help`
and each subcommand's `--help` output.

## `agentic-hooks`

```
Second Brain orchestration CLI

Usage:
  agentic-hooks [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         Run the Search + self-correcting generate/review loop on a task
  serve       Run the Second Brain as an MCP server over stdio
  version     Print the agentic-hooks version

Flags:
  -h, --help   help for agentic-hooks
```

## `agentic-hooks version`

Prints `agentic-hooks <version> (built <build-time>, <go-version>)`. `Version`,
`BuildTime`, `GoVersion` are set via `-ldflags` at build time (see
[`makefile-targets.md`](makefile-targets.md#build)); under `go run` they stay
at their zero values (`dev`, `unknown`, `unknown`).

No flags beyond global `-h`/`--help`.

## `agentic-hooks run [task]`

```
Run the Search + self-correcting generate/review loop on a task

Usage:
  agentic-hooks run [task] [flags]

Flags:
      --feedback-dir string              directory for the append-only human-feedback JSONL log (default "feedback")
  -h, --help                             help for run
      --knowledge-dir string             path to the Second Brain knowledge directory (required)
      --max-iterations uint              max generate/review passes before the loop returns its best-effort draft (default 4)
      --search-mcp-server string         command to launch the Search agent's MCP server (required)
      --search-mcp-server-args strings   comma-separated arguments passed to --search-mcp-server
```

| Flag | Type | Default | Required | Notes |
|---|---|---|---|---|
| `[task]` (positional) | string | — | yes (`cobra.ExactArgs(1)`) | The task text to run through Search + Generator/Review. |
| `--knowledge-dir` | string | — | yes | Path to the Second Brain directory (`internal/secondbrain.Load`). |
| `--search-mcp-server` | string | — | yes | Command used to spawn the Search sub-agent's MCP client subprocess (e.g. the same `agentic-hooks` binary, invoked with `serve`). |
| `--search-mcp-server-args` | comma-separated string list | none | no | Args passed to `--search-mcp-server`. Example: `"serve,--knowledge-dir,knowledge"`. |
| `--max-iterations` | uint | `4` | no | Bound on Generator↔Review passes (`internal/agent.NewSelfCorrectingLoop`). |
| `--feedback-dir` | string | `"feedback"` | no | Directory for the append-only `feedback.jsonl` HITL log. |

Reads `GEMINI_API_KEY` env var (falls back to `GOOGLE_API_KEY` via the
underlying `genai` client, even though `cmd/agentic-hooks/run.go` only
references `GEMINI_API_KEY` directly — see
[`docs/adr/0006-gemini-api-key-canonical.md`](../adr/0006-gemini-api-key-canonical.md)).

Exit behavior: streams `[author] text` lines to stdout as the pipeline runs,
then prompts `Approve? [y/N]:` and `Reason (optional, for the feedback log):`
on stdin. Only a literal `y` or `Y` answer is treated as approved; anything
else (including empty input) is a reject. A feedback record is appended to
`<feedback-dir>/feedback.jsonl` regardless of the decision.

## `agentic-hooks serve`

```
Run the Second Brain as an MCP server over stdio

Usage:
  agentic-hooks serve [flags]

Flags:
  -h, --help                   help for serve
      --knowledge-dir string   path to the Second Brain knowledge directory (required)
```

| Flag | Type | Default | Required | Notes |
|---|---|---|---|---|
| `--knowledge-dir` | string | — | yes | Path to the Second Brain directory. |

Blocks on stdio waiting for MCP JSON-RPC messages — prints nothing, does not
exit on its own. This is correct behavior for an MCP server meant to be
spawned by a host process, not run as an interactive command. Ctrl-C to
stop it manually. See [`mcp-tools.md`](mcp-tools.md) for the tools it
exposes and [`docs/how-to/test-with-mcp-inspector.md`](../how-to/test-with-mcp-inspector.md)
to drive it manually.

## `agentic-hooks completion`

Standard Cobra-generated shell completion script command (`bash`, `zsh`,
`fish`, `powershell`). Not project-specific; see `agentic-hooks completion --help`
for the current Cobra version's exact subcommand list.

## `agentic-hooks help [command]`

Standard Cobra help command — equivalent to `agentic-hooks [command] --help`.
