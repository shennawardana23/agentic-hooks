# ADR-0006: GEMINI_API_KEY is canonical; GOOGLE_API_KEY is a fallback only

## Status
Accepted

## Context
`cmd/agentic-hooks/run.go`'s `newDefaultModel` reads `GEMINI_API_KEY` from
the environment when constructing the Gemini model client. During live
verification, a session was run with only `GEMINI_API_KEY` exported and it
worked correctly even though, at one point, the code's own comments assumed
`GOOGLE_API_KEY` was being read. Reading `google.golang.org/genai`'s
`client.go` directly (`genai.NewClient`) confirmed the SDK itself falls
back to reading `GEMINI_API_KEY`/`GOOGLE_API_KEY` from the environment
whenever the passed-in `ClientConfig.APIKey` is empty, regardless of
whether the config struct is nil — this fallback is a genai SDK behavior,
not something `agentic-hooks` needs to special-case in its own code.

## Decision
`GEMINI_API_KEY` is the canonical environment variable name for this
project's documentation, `Makefile`, and error messages. `GOOGLE_API_KEY`
remains a working fallback, handled transparently by the `genai` SDK
itself — no additional fallback logic is written into `agentic-hooks`'
own code beyond what `run.go` already does (`os.Getenv("GEMINI_API_KEY")`)
and what the Makefile's `apiKey` variable does (`${GEMINI_API_KEY:-$GOOGLE_API_KEY}`).

## Consequences
- Documentation and error messages consistently point at `GEMINI_API_KEY`
  first, avoiding confusion about which variable is "the real one."
  `GOOGLE_API_KEY` still works because the underlying SDK handles it, not
  because `agentic-hooks` duplicates that fallback logic.
- If the `genai` SDK ever changes its own fallback behavior, this project's
  documented guidance would need to be re-verified against the new SDK
  version — this decision is coupled to `google.golang.org/genai`'s
  current implementation, not an independent guarantee `agentic-hooks`
  enforces itself.