# blindenv

Secret isolation for AI coding agents. Agents use secrets but never see them.

## The problem

AI coding agents (Claude Code, OpenClaw, etc.) need API keys to do real work. Today you either paste keys into chat or put them in `.env` files the agent can read. Both expose secrets to the agent's context - one prompt injection away from leaking.

## How blindenv works

```
Agent writes:    curl -H "Authorization: $API_KEY" https://api.example.com
                         ↓
blindenv proxy:  Injects real value into subprocess env
                         ↓
Agent receives:  {"result": "ok", "token": "[REDACTED]"}
```

The agent uses `$VAR` names. blindenv resolves values in an isolated subprocess. Output is scanned and secret values replaced with `[REDACTED]`. The agent never sees the actual secret.

## Install

```bash
# macOS / Linux
curl -fsSL https://blindenv.ai/install.sh | sh

# Or build from source
go install github.com/neuradex-labs/blindenv@latest
```

## Quick start

1. Create `blindenv.yml` in your project root:

```yaml
secret_files:
  - .env
```

That's it. All key-value pairs in `.env` are now injected into commands and redacted from output.

2. Run a command:

```bash
blindenv run 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
# Output: {"result": "ok", "key": "[REDACTED]"}
```

## Configuration

```yaml
# blindenv.yml

inject:              # env vars from host process - injected + redacted
  - CI_TOKEN
  - DEPLOY_KEY

passthrough:         # non-secret vars - explicit allowlist (strict mode)
  - PATH
  - HOME
  - LANG

secret_files:        # .env files - auto-parsed, paths blocked from agent
  - .env
  - .env.local
  - ~/.aws/credentials
```

**`secret_files` alone is usually enough.** `inject` is for env vars that exist in the host process but not in any file (e.g. CI secrets).

When `passthrough` is set, the subprocess gets ONLY those vars plus injected secrets. Without it, the full host env is inherited.

## Defense layers

| # | Layer | What it does |
|---|-------|-------------|
| 1 | Subprocess env isolation | Subprocess gets only permitted env vars |
| 2 | Output redaction | stdout/stderr scanned for secret values, replaced with `[REDACTED]` |
| 3 | Secret file blocking | Agent cannot read files listed in `secret_files` |
| 4 | Config protection | Agent cannot modify `blindenv.yml` |

## Agent integration

### Claude Code

Add hooks to your project's `.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{ "type": "command", "command": "blindenv hook cc bash" }]
      },
      {
        "matcher": "Read",
        "hooks": [{ "type": "command", "command": "blindenv hook cc read" }]
      },
      {
        "matcher": "Grep",
        "hooks": [{ "type": "command", "command": "blindenv hook cc grep" }]
      },
      {
        "matcher": "Edit",
        "hooks": [{ "type": "command", "command": "blindenv hook cc guard-config" }]
      },
      {
        "matcher": "Write",
        "hooks": [{ "type": "command", "command": "blindenv hook cc guard-config" }]
      }
    ]
  }
}
```

With hooks active, Claude Code's bare `Bash` calls are blocked. The agent is guided to use `blindenv run '...'` instead. Secret files are blocked from `Read` and `Grep`. The agent cannot edit `blindenv.yml`.

### OpenClaw

Coming soon.

## CLI reference

```
blindenv run '<command>'              Execute with secret isolation + output redaction
blindenv check-file <path>            Check if a file contains or exposes secrets (exit 2 = blocked)
blindenv has-config                   Exit 0 if config with secrets exists, 1 otherwise
blindenv hook cc <bash|read|grep|guard-config>   Claude Code PreToolUse hooks
blindenv version                      Show version
```

## License

MIT
