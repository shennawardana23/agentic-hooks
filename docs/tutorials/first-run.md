# Your first run

This is a single, linear path from a fresh clone to seeing the
Search+Review loop produce a verdict you approve. It assumes nothing about
prior familiarity with the project. If you already know the basics and
just need to look something up, see the [reference](../reference/) docs
instead — this tutorial does not explain *why* things work the way they
do (see [explanation](../explanation/) for that), only what to do and what
you should see at each step.

## Prerequisites

- Go 1.25 or later (`go version`).
- A `GEMINI_API_KEY` (or `GOOGLE_API_KEY` as a fallback) environment
  variable set, for the last step only. The build and MCP server steps
  below need no API key at all.

## Step 1 — clone and build

```bash
git clone <this-repo>
cd agentic-hooks
make build
```

You should see a single `go build` invocation complete with no output
other than the command itself, and a new binary at `bin/agentic-hooks`.
Confirm it:

```bash
bin/agentic-hooks version
```

Expected output: `agentic-hooks <version> (built <timestamp>, <go version>)`.

## Step 2 — create a one-file Second Brain

The project ships its own `knowledge/` directory with real content, but
for this first run it's clearer to build a minimal one by hand so you can
see exactly which concept gets matched and why.

```bash
mkdir -p /tmp/agentic-hooks-knowledge/solid
cat > /tmp/agentic-hooks-knowledge/solid/single-responsibility.md <<'EOF'
---
type: principle
title: Single Responsibility Principle
tags: [solid]
---

A component should have one reason to change.
EOF
```

This one file is now a complete, valid Second Brain — one concept, one
tag. See [how to add a concept](../how-to/add-a-concept.md) for the full
field reference when you're ready to add more.

## Step 3 — confirm the MCP server sees it

Before running the full agent loop, confirm the Second Brain loads
correctly by starting the MCP server against it:

```bash
bin/agentic-hooks serve --knowledge-dir /tmp/agentic-hooks-knowledge
```

Nothing will print, and the command will not exit — this is correct. The
server is now blocked on stdin, waiting for MCP JSON-RPC requests. Press
`Ctrl-C` to stop it before continuing. If you want to see it actually
answer a request rather than just start and stop, see
[testing with MCP Inspector](../how-to/test-with-mcp-inspector.md).

## Step 4 — run the Search+Review loop

Now run a task through the full pipeline. This uses the project's own
`serve` command as the Search sub-agent's MCP server — a valid stand-in
for a first run, since it proves the MCP-client wiring end-to-end even
though it searches the Second Brain rather than the open web.

```bash
export GEMINI_API_KEY="your-real-key"

bin/agentic-hooks run \
  "review: func DoEverything() { validates input, writes to disk, sends email, all violates solid }" \
  --knowledge-dir /tmp/agentic-hooks-knowledge \
  --search-mcp-server bin/agentic-hooks \
  --search-mcp-server-args "serve,--knowledge-dir,/tmp/agentic-hooks-knowledge"
```

You should see streamed lines tagged by author, e.g. `[generator] ...` and
`[review] ...`, as the Generator drafts an answer and the Review agent
critiques it against the Single Responsibility Principle concept you just
created. This can take one or more draft/critique passes (bounded by
`--max-iterations`, default 4).

## Step 5 — approve or reject at the HITL gate

Once the loop converges (or hits its iteration bound), the CLI prints the
final transcript and prompts:

```
Approve? [y/N]:
```

Type `y` and press Enter. Expected: the verdict prints again as the final
output, and a new record is appended to `feedback/feedback.jsonl`
containing the task, the transcript, and your decision. Typing anything
else (or nothing) is treated as a reject — nothing is returned as final,
but the rejection is still logged.

You'll then be prompted for an optional free-text reason, which is also
written to the feedback log:

```
Reason (optional, for the feedback log):
```

Press Enter to skip it, or type a short reason and press Enter.

## What you just did

You built the binary, created a minimal knowledge base, confirmed the MCP
server serves it, ran the full Search+Review pipeline against a real
model, and exercised the human-in-the-loop approval gate — the same path
any real usage of this project follows, just with a one-file Second Brain
instead of the full `knowledge/` directory.

## Next steps

- [Add a real concept](../how-to/add-a-concept.md) to the project's actual
  `knowledge/` directory.
- [Point Search at a different MCP server](../how-to/point-search-at-another-mcp-server.md)
  instead of `agentic-hooks serve` itself.
- Read [why the loop converges](../explanation/self-correcting-loop.md) to
  understand what you just watched happen in Step 4.
