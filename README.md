<p align="center">
  <img src="https://img.shields.io/github/v/release/neuradex/blindenv?style=flat-square&color=blue" alt="release" />
  <img src="https://img.shields.io/badge/Claude_Code-plugin-blueviolet?style=flat-square" alt="Claude Code plugin" />
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="license" />
</p>

<h1 align="center">blindenv</h1>

<p align="center">
  <strong>Secret isolation for AI coding agents.</strong>
  <br>
  Agents use your secrets but never see them.
</p>

<p align="center">
  <a href="#the-problem">The Problem</a> ·
  <a href="#how-it-works">How It Works</a> ·
  <a href="#install">Install</a> ·
  <a href="#configuration">Configuration</a> ·
  <a href="#defense-layers">Defense Layers</a> ·
  <a href="#cli-reference">CLI Reference</a>
</p>

---

Your AI agent needs your API key. So you paste it into the chat. Or you put it in `.env` and let the agent read it.

Either way, the secret is now in the agent's context — one prompt injection away from leaking. The agent doesn't need to be malicious. It just needs to be tricked.

**blindenv solves this structurally.** Secrets are injected into an isolated subprocess, never exposed to the agent's context, and scrubbed from all output before the agent sees it.

```
Agent writes:    curl -H "Authorization: $API_KEY" https://api.example.com
                         ↓
blindenv proxy:  Injects real value into subprocess env
                         ↓
Agent receives:  {"result": "ok", "token": "[REDACTED]"}
```

The agent uses `$VAR` references. blindenv resolves the real values behind the scenes. The agent gets working results — but never the keys.

---

## The Problem

AI coding agents (Claude Code, Cursor, etc.) need credentials to do real work — deploy code, call APIs, access databases. Today you either:

1. **Paste keys into chat** — directly in the agent's context
2. **Put them in `.env` files** — the agent reads them with `cat` or `grep`

Both approaches expose secrets to the LLM. The agent now knows your keys, and anything it outputs — logs, summaries, code comments — could contain them. A well-crafted prompt injection in any file the agent reads could exfiltrate them.

This isn't hypothetical. It's the default.

---

## How It Works

```
┌─────────────────────────────────────────────────────┐
│  Agent Context (no secrets)                         │
│                                                     │
│  "curl -H 'Authorization: $API_KEY' example.com"   │
│          │                                          │
│          ▼                                          │
│  ┌─────────────────────────────────────┐            │
│  │  blindenv proxy                     │            │
│  │  ┌──────────────┐                  │            │
│  │  │ Resolve       │ API_KEY=sk-a1b2 │            │
│  │  │ Isolate       │ subprocess only │            │
│  │  │ Execute       │ real curl runs  │            │
│  │  │ Redact        │ sk-a1b2→[REDACTED] │         │
│  │  └──────────────┘                  │            │
│  └─────────────────────────────────────┘            │
│          │                                          │
│          ▼                                          │
│  {"result": "ok", "token": "[REDACTED]"}            │
└─────────────────────────────────────────────────────┘
```

Four layers, zero exposure:

1. **Resolve** — Reads secrets from `.env` files and host environment
2. **Isolate** — Injects values into a subprocess environment only
3. **Execute** — Runs the command with real credentials
4. **Redact** — Scans output and replaces secret values with `[REDACTED]`

The agent never sees, stores, or outputs any actual secret value.

---

## Install

### Claude Code Plugin (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/neuradex/blindenv/main/scripts/install.sh | sh
```

Then in Claude Code:

```
/install-plugin neuradex/blindenv
```

The plugin hooks into Claude Code's PreToolUse system — every Bash command is automatically routed through blindenv, and secret files are blocked from Read, Grep, Edit, and Write.

### Manual install

```bash
# Download latest release
curl -fsSL https://raw.githubusercontent.com/neuradex/blindenv/main/scripts/install.sh | sh

# Or build from source
go install github.com/neuradex-labs/blindenv@latest
```

### Verify

```bash
blindenv has-config   # exit 0 if blindenv.yml found with secrets
blindenv run 'echo $API_KEY'   # test secret injection + redaction
```

---

## Quick Start

Create `blindenv.yml` in your project root:

```yaml
secret_files:
  - .env
```

That's it. All key-value pairs in `.env` are now:
- **Injected** into commands via `blindenv run`
- **Redacted** from all output
- **Blocked** from agent file access

```bash
blindenv run 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
# → {"result": "ok", "key": "[REDACTED]"}
```

---

## Configuration

```yaml
# blindenv.yml

secret_files:        # .env files — auto-parsed, paths blocked from agent
  - .env
  - .env.local
  - ~/.aws/credentials

inject:              # env vars from host process — injected + redacted
  - CI_TOKEN
  - DEPLOY_KEY

passthrough:         # non-secret vars — explicit allowlist (strict mode)
  - PATH
  - HOME
  - LANG
```

### What each field does

| Field | Purpose | When to use |
|-------|---------|-------------|
| `secret_files` | Parse `.env` files, inject values, block file access | **Always** — this is the primary mechanism |
| `inject` | Pull env vars from the host process | CI/CD secrets, vars not in any file |
| `passthrough` | Strict allowlist for non-secret vars | High-security environments |

**`secret_files` alone is usually enough.** `inject` is for env vars that exist in the host process but not in any file (e.g. CI secrets).

When `passthrough` is set, the subprocess gets ONLY those vars plus injected secrets (strict mode). Without it, the full host env is inherited (permissive mode).

Config is discovered by walking up from `cwd` to `/`, then checking `~/.blindenv.yml`. The nearest one wins — just like `.gitignore`.

---

## Defense Layers

| # | Layer | What it does |
|---|-------|-------------|
| 1 | **Subprocess isolation** | Secrets exist only in the subprocess environment — never in the agent's context |
| 2 | **Output redaction** | stdout/stderr scanned for secret values, replaced with `[REDACTED]` |
| 3 | **File blocking** | Agent cannot Read, Grep, Edit, or Write files listed in `secret_files` |
| 4 | **Config protection** | Agent cannot modify `blindenv.yml` — the rules are tamper-proof |

```
┌─ blindenv.yml ──────────────────────────────────────┐
│                                                      │
│  Bash hook          Read/Grep hook    Edit/Write hook│
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │ Rewrite cmd   │  │ Block secret │  │ Block      │ │
│  │ → blindenv    │  │ file access  │  │ secret     │ │
│  │   run '...'   │  │              │  │ files +    │ │
│  │ Inject secrets│  │              │  │ config     │ │
│  │ Redact output │  │              │  │            │ │
│  └──────────────┘  └──────────────┘  └────────────┘ │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

## Claude Code Integration

With the plugin installed, five PreToolUse hooks guard every agent action:

| Tool | Hook | Behavior |
|------|------|----------|
| **Bash** | `blindenv hook cc bash` | Rewrites command to `blindenv run '...'` — secrets injected, output redacted |
| **Read** | `blindenv hook cc read` | Blocks read access to secret files |
| **Grep** | `blindenv hook cc grep` | Blocks search in secret files |
| **Edit** | `blindenv hook cc guard-file` | Blocks edits to secret files and `blindenv.yml` |
| **Write** | `blindenv hook cc guard-file` | Blocks writes to secret files and `blindenv.yml` |

The hooks use **exit 2** (blocking error), which works in all Claude Code permission modes. This isn't a suggestion — it's a structural gate.

---

## CLI Reference

```
blindenv run '<command>'              Execute with secret isolation + output redaction
blindenv check-file <path>            Check if file is blocked (exit 2 = blocked)
blindenv has-config                   Exit 0 if config with secrets exists, 1 otherwise
blindenv hook cc <hook>               Claude Code PreToolUse hooks
                                       bash | read | grep | guard-file
```

---

## Platform Support

| Platform | Architecture | Status |
|----------|-------------|--------|
| macOS    | Apple Silicon (arm64) | Supported |
| macOS    | Intel (amd64) | Supported |
| Linux    | x86_64 (amd64) | Supported |
| Linux    | ARM (arm64) | Supported |

---

## License

MIT

---

<p align="center">
  <strong>Your agent doesn't need your keys. It needs what your keys unlock.</strong>
  <br>
  blindenv gives it access without exposure.
</p>
