# How to point Search at a different MCP server

The Search sub-agent's MCP client is fully config-driven — nothing is
hardcoded, so it can point at any MCP server that speaks stdio, not just
`agentic-hooks serve` itself.

## Flags

`agentic-hooks run` takes two flags for this:

| Flag | Required | Notes |
|---|---|---|
| `--search-mcp-server` | yes | The command to launch the target MCP server. |
| `--search-mcp-server-args` | no | Comma-separated arguments passed to that command. |

## Using agentic-hooks' own server (the default pattern)

```bash
bin/agentic-hooks run "<task>" \
  --knowledge-dir knowledge \
  --search-mcp-server bin/agentic-hooks \
  --search-mcp-server-args "serve,--knowledge-dir,knowledge"
```

## Pointing at a different MCP server binary

Swap `--search-mcp-server` for any other executable that implements an
MCP stdio server, and pass whatever arguments it needs via
`--search-mcp-server-args`:

```bash
bin/agentic-hooks run "<task>" \
  --knowledge-dir knowledge \
  --search-mcp-server /path/to/other-mcp-server \
  --search-mcp-server-args "arg1,arg2,arg3"
```

The Search sub-agent will attach to whatever tools that server advertises
via `tools/list` — it is not hardcoded to `list_knowledge`/`get_knowledge`.

## Notes

- Both flags are passed straight through to the Search sub-agent's MCP
  client (`mcptoolset`) — there is no validation that the target process
  is actually an MCP server until the client tries to connect.
- If the target server is unreachable, the Search sub-agent returns a
  typed error to the root agent rather than failing the whole run; the
  Review agent can still proceed using only the Second Brain.
- There is no default value for `--search-mcp-server` — it is a required
  flag, by design, so nothing is silently pointed at a server you didn't
  choose.
