<!-- AGENT NOTE: If you modify this file, also update docs/README.ko.md (Korean) and docs/README.ja.md (Japanese) to keep translations in sync. -->

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
  <a href="#what-it-does">What It Does</a> ·
  <a href="#install">Install</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#how-it-works">How It Works</a> ·
  <a href="#configuration">Configuration</a> ·
  <a href="#cli-reference">CLI Reference</a>
</p>

<p align="center">
  <strong>English</strong> ·
  <a href="./docs/README.ko.md">한국어</a> ·
  <a href="./docs/README.ja.md">日本語</a>
</p>

---

## What It Does

blindenv lets AI agents **use** your API keys, database credentials, and tokens — without ever **seeing** them.

- **Secret injection** — Resolves `$VAR` references in an isolated subprocess. The agent writes `$API_KEY`, the real value is injected behind the scenes.
- **Output redaction** — Scans all stdout/stderr and replaces secret values with `[REDACTED]` before the agent sees it.
- **File blocking** — Prevents agents from reading, searching, or editing your `.env` files and credentials.
- **Config protection** — Agents cannot modify `blindenv.yml`. The rules are tamper-proof.
- **Content-aware blocking** — Even if a secret file is copied or renamed, any file containing a secret value is blocked. Path evasion doesn't work.

```
Agent writes:    curl -H "Authorization: $API_KEY" https://api.example.com
                         ↓
blindenv proxy:  Injects real value into subprocess env
                         ↓
Agent receives:  {"result": "ok", "token": "[REDACTED]"}
```

---

## Install

### Claude Code Plugin (recommended)

```bash
/plugin marketplace add neuradex/blindenv
/plugin install blindenv@blindenv
```

That's it. On the next session start, the binary is automatically downloaded from [GitHub Releases](https://github.com/neuradex/blindenv/releases) for your platform and `blindenv.yml` is auto-generated in your project root.

Open `blindenv.yml` and configure which files contain secrets:

```yaml
secret_files:
  - .env
  - .env.local
  # - ~/.aws/credentials
```

Once configured, the agent cannot even know these files exist — all access is structurally blocked.

### Build from source

```bash
go install github.com/neuradex/blindenv@latest
```

### Platform support

| Platform | Architecture | |
|----------|-------------|-|
| macOS | Apple Silicon (arm64) | Supported |
| macOS | Intel (amd64) | Supported |
| Linux | x86_64 (amd64) | Supported |
| Linux | ARM (arm64) | Supported |
| Windows | x86_64 (amd64) | Supported |
| Windows | ARM64 | Supported |

---

## Quick Start

With `blindenv.yml` in place, all key-value pairs in your `.env` files are automatically:

| | What happens |
|---|---|
| **Injected** | Secret values available as `$VAR` in commands via `blindenv run` |
| **Redacted** | Any output containing a secret value → `[REDACTED]` |
| **Blocked** | Agent cannot Read, Grep, Edit, or Write secret files |

```bash
# Your .env contains: API_KEY=sk-a1b2c3d4
blindenv run 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
# → {"result": "ok", "key": "[REDACTED]"}
```

When used as a Claude Code plugin, you don't even need `blindenv run` — the hook rewrites Bash commands automatically.

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

### Defense layers

| # | Layer | What it does |
|---|-------|-------------|
| 1 | **Subprocess isolation** | Secrets exist only in the subprocess environment — never in the agent's context |
| 2 | **Output redaction** | stdout/stderr scanned for secret values, replaced with `[REDACTED]` |
| 3 | **File blocking** | Agent cannot Read, Grep, Edit, or Write files listed in `secret_files` |
| 4 | **Config protection** | Agent cannot modify `blindenv.yml` — the rules are tamper-proof |
| 5 | **Content-aware blocking** | Files containing secret values are blocked regardless of path — copying or renaming won't help |

### Claude Code hooks

With the plugin installed, five PreToolUse hooks guard every agent action:

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

| Tool | Hook | Behavior |
|------|------|----------|
| **Bash** | `blindenv hook cc bash` | Rewrites command to `blindenv run '...'` — secrets injected, output redacted |
| **Read** | `blindenv hook cc read` | Blocks read access to secret files |
| **Grep** | `blindenv hook cc grep` | Blocks search in secret files |
| **Edit** | `blindenv hook cc guard-file` | Blocks edits to secret files and `blindenv.yml` |
| **Write** | `blindenv hook cc guard-file` | Blocks writes to secret files and `blindenv.yml` |

The hooks use **exit 2** (blocking error), which works in all Claude Code permission modes. This isn't a suggestion — it's a structural gate.

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

| Field | Purpose | When to use |
|-------|---------|-------------|
| `secret_files` | Parse `.env` files, inject values, block file access | **Always** — this is the primary mechanism |
| `inject` | Pull env vars from the host process | CI/CD secrets, vars not in any file |
| `passthrough` | Strict allowlist for non-secret vars | High-security environments |

**`secret_files` alone is usually enough.** `inject` is for env vars that exist in the host process but not in any file (e.g. CI secrets).

When `passthrough` is set, the subprocess gets ONLY those vars plus injected secrets (strict mode). Without it, the full host env is inherited (permissive mode).

Config is discovered by walking up from `cwd` to `/`, then checking `~/.blindenv.yml`. The nearest one wins — just like `.gitignore`.

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

## Why not just use .env?

Your AI agent needs your API key. So you paste it into the chat. Or you put it in `.env` and let the agent read it.

Either way, the secret is now in the agent's context — one prompt injection away from leaking. The agent doesn't need to be malicious. It just needs to be tricked.

**"But I opted out of training."** — Maybe. Your provider may not use your data for training, and may delete it after the retention period as promised. But between the moment your secret hits their server and the moment it's deleted, anything can happen. A breach, a subpoena, an internal incident, a misconfigured backup. You don't control that timeline.

And unlike other services, what's stored here isn't just a password in isolation — it's your credential **with full context**. The conversation contains what the key is for, which service it accesses, how the API is called, what infrastructure it connects to. If that record is ever compromised, it's not just a leaked key. It's a complete playbook for exploitation.

The only real defense is to never send the secret in the first place.

**"Just don't give agents API keys."** — Sure. And don't give your car keys to the valet. Park three blocks away, carry your luggage in the rain, and congratulate yourself on your security posture. The point of an AI agent is to do real work — deploy code, call APIs, access services. An agent without credentials is a very expensive autocomplete. The goal isn't to stop using keys. It's to use them safely.

blindenv solves this structurally. Not through prompts. Not through trust. Through isolation.

---

## License

MIT

---

<p align="center">
  <strong>Your agent doesn't need your keys. It needs what your keys unlock.</strong>
  <br>
  blindenv gives it access without exposure.
</p>
