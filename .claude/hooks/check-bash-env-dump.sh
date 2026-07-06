#!/usr/bin/env bash
# PreToolUse hook for the Bash tool.
#
# Blocks commands that look like environment/secret dumps:
#   - bare `env` (optionally piped/chained, e.g. `env | grep API_KEY`)
#   - bare `printenv` (no args = full dump)
#   - `cat` of a file literally named .env (or .env.local, .env.production,
#     etc.), any path ending in one of those, or a quoted form of the same
#   - `echo $ALLCAPS_VAR` (including "$VAR", ${VAR}, "${VAR}" forms)
#
# See policies/07-env-secrets-protection.md for the rationale (07.4, 07.5).
#
# Override paths (so this isn't a permanent dead end):
#   - A short allowlist of common, non-secret-shaped vars (PATH, HOME, ...)
#     is always permitted for the `echo $VAR` check.
#   - Set AGENTIC_HOOKS_ALLOW_ENV_ECHO=1 to allow any `echo $VAR` for this
#     command. This does NOT relax the env/printenv/cat-.env checks, since
#     those always disclose an entire environment or file rather than one
#     named value.

set -uo pipefail

input="$(cat)"
cmd="$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)"

# No command field (not a Bash tool_use, or malformed input) -> allow.
if [ -z "$cmd" ]; then
  exit 0
fi

POLICY_MSG="Blocked: this command looks like it dumps environment variables or a secrets file (see policies/07-env-secrets-protection.md). Target the single variable/file you actually need instead. For a narrow, legitimate echo of a known-benign var, set AGENTIC_HOOKS_ALLOW_ENV_ECHO=1 for this command."

# 1. Bare `env` (optionally with flags), alone or piped/chained.
if printf '%s' "$cmd" | grep -Eq '(^|[;&|]\s*)env(\s+-[A-Za-z0-9]+)*\s*($|[|;&])'; then
  echo "$POLICY_MSG" >&2
  exit 2
fi

# 2. Bare `printenv` (no args = full dump), alone or piped/chained.
if printf '%s' "$cmd" | grep -Eq '(^|[;&|]\s*)printenv\s*($|[|;&])'; then
  echo "$POLICY_MSG" >&2
  exit 2
fi

# 3. `cat` of a file literally named .env (or the .env.<suffix> family,
#    e.g. .env.local, .env.production) or any path ending in one of those,
#    optionally wrapped in single or double quotes (`cat ".env"`, `cat '.env'`).
#    Known, accepted limitation: this is a regex heuristic, not a shell
#    parser, so flags before the filename (`cat -A .env`) or indirect
#    invocation (`bash -c 'cat .env'`, `eval "cat .env"`) can still evade it.
if printf '%s' "$cmd" | grep -Eq '(^|[;&|[:space:]])cat[[:space:]]+["'"'"']?[^[:space:]]*\.env(\.[A-Za-z0-9_.-]+)?["'"'"']?([[:space:]]|$)'; then
  echo "$POLICY_MSG" >&2
  exit 2
fi

# 4. `echo $ALLCAPS_VAR` (also `echo "$ALLCAPS_VAR"`, `echo ${ALLCAPS_VAR}`,
#    `echo "${ALLCAPS_VAR}"`) — allowlist common benign vars, else require
#    explicit override. Only ALL-CAPS names are treated as "looks like a
#    secret"; this is a heuristic, not full env-var syntax support (e.g.
#    `${VAR:-default}`, `echo $(env)`, or `bash -c 'echo $SECRET'` are known,
#    accepted gaps in a regex-based check).
if [[ "$cmd" =~ echo[[:space:]]+\"?\$\{?([A-Z][A-Z0-9_]*)\}?\"? ]]; then
  var="${BASH_REMATCH[1]}"
  case "$var" in
    PATH|HOME|USER|SHELL|PWD|OLDPWD|LANG|LC_ALL|TERM|LOGNAME|HOSTNAME|TMPDIR|EDITOR|VISUAL|TZ)
      exit 0
      ;;
    *)
      if [ "${AGENTIC_HOOKS_ALLOW_ENV_ECHO:-0}" = "1" ]; then
        exit 0
      fi
      echo "$POLICY_MSG" >&2
      exit 2
      ;;
  esac
fi

exit 0
