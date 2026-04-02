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
  <a href="#cli-reference">CLI Reference</a> ·
  <a href="#beyond-secret-managers">Comparison</a>
</p>

<p align="center">
  <strong>English</strong> ·
  <a href="./docs/README.ko.md">한국어</a> ·
  <a href="./docs/README.ja.md">日本語</a>
</p>

---

## What It Does

blindenv lets AI agents **use** your API keys, database credentials, and tokens — without ever **seeing** them.

The agent reads your `.env` file and sees structure, variable names, and comments — but every secret value is replaced with `[BLINDED]`. Commands run with real credentials injected behind the scenes, and any output containing a secret is automatically redacted.

```
Agent reads .env:    API_KEY=[BLINDED]
                     DB_PASSWORD=[BLINDED]
                     DEBUG=true              ← non-secret values pass through

Agent runs command:  curl -H "Authorization: $API_KEY" https://api.example.com
                         ↓
blindenv proxy:      Injects real value into subprocess env
                         ↓
Agent receives:      {"result": "ok", "token": "[BLINDED]"}
```

### What the agent sees

| Agent action | Result |
|---|---|
| Read `.env` | Variable names visible, all values → `[BLINDED]` |
| `echo $API_KEY` in Bash | `[BLINDED]` |
| `curl` with `$API_KEY` | Works — real value injected into subprocess |
| Any output containing a secret | Automatically replaced with `[BLINDED]` |

The agent gets full context about your project's configuration structure. It just can't see the values that matter. Meanwhile, API calls work, deploys succeed, and services respond.

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

### Build from source

```bash
git clone https://github.com/neuradex/blindenv.git
cd blindenv
make build      # → ./blindenv
```

Or install globally: `go install github.com/neuradex/blindenv@latest`

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and project structure.

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

With `blindenv.yml` in place, everything works automatically:

| | What happens |
|---|---|
| **Masked** | Agent reads `.env` files with all values replaced by `[BLINDED]` |
| **Injected** | Real secret values available as `$VAR` in commands via `blindenv run` |
| **Redacted** | Any output containing a secret value → `[BLINDED]` |

```bash
# Your .env contains: API_KEY=sk-a1b2c3d4
blindenv run 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
# → {"result": "ok", "key": "[BLINDED]"}
```

When used as a Claude Code plugin, you don't even need `blindenv run` — the hook rewrites Bash commands automatically.

---

## How It Works

### Defense layers

| # | Layer | What it does |
|---|-------|-------------|
| 1 | **File masking** | Secret files are readable, but all values replaced with `[BLINDED]` |
| 2 | **Subprocess isolation** | Real secrets exist only in the subprocess environment — never in the agent's context |
| 3 | **Output redaction** | stdout/stderr scanned for secret values, replaced with `[BLINDED]` |
| 4 | **Auto-detection** | Environment variables matching patterns like `KEY`, `SECRET`, `TOKEN` are automatically masked |
| 5 | **Config protection** | Agent cannot modify `blindenv.yml` — the rules are tamper-proof |

### Claude Code hooks

With the plugin installed, six PreToolUse hooks guard every agent action:

| Tool | Hook | Behavior |
|------|------|----------|
| **Bash** | `blindenv hook cc bash` | Rewrites command to `blindenv run '...'` — secrets injected, output redacted |
| **Read** | `blindenv hook cc read` | Secret files → masked copy with all values as `[BLINDED]` |
| **Edit** | `blindenv hook cc guard-file` | Secret files and `blindenv.yml` → protected from modification |
| **Write** | `blindenv hook cc guard-file` | Secret files and `blindenv.yml` → protected from modification |
| **Grep** | `blindenv hook cc grep` | Secret files omitted from search results |
| **Glob** | `blindenv hook cc glob` | Secret files omitted from file listings |

---

## Configuration

```yaml
# blindenv.yml

secret_files:            # .env files — auto-parsed, values masked
  - .env
  - .env.local
  - ~/.aws/credentials

# mask_patterns:         # env var name patterns for auto-detection
#   - KEY               # (defaults apply when omitted — KEY, SECRET, TOKEN, etc.)
#   - SECRET

# mask_env:              # explicit env vars to mask (for names not matching patterns)
#   - MY_CUSTOM_VAR

# inject:                # env vars from host process — injected + redacted
#   - CI_TOKEN
#   - DEPLOY_KEY
```

| Field | Purpose | When to use |
|-------|---------|-------------|
| `secret_files` | Parse `.env` files, mask values, block file access | **Always** — this is the primary mechanism |
| `mask_patterns` | Auto-detect env vars by name pattern (e.g. `KEY`, `TOKEN`) | Defaults apply when omitted; customize to narrow or widen |
| `mask_env` | Explicit env var names to mask | Vars that don't match any pattern |
| `inject` | Pull env vars from the host process into subprocess | CI/CD secrets, vars not in any file |

**`secret_files` alone is usually enough.** The default `mask_patterns` automatically catch common secret variable names (`KEY`, `SECRET`, `TOKEN`, `PASSWORD`, etc.) from the process environment.

Config is discovered by walking up from `cwd` to `/`, then checking `~/.blindenv.yml`. The nearest one wins — just like `.gitignore`.

> For additional security modes (`block`) and advanced options (`passthrough`), see [Advanced Configuration](./docs/ADVANCED.md).

---

## CLI Reference

```
blindenv run '<command>'              Execute with secret isolation + output redaction
blindenv init                         Create blindenv.yml in current directory
blindenv hook cc <hook>               Claude Code PreToolUse hooks
                                       bash | read | grep | glob | guard-file
```

> For additional commands (`check-file`, `has-config`), see [Advanced Configuration](./docs/ADVANCED.md).

---

## Beyond Secret Managers

Traditional secret managers solve **storage and delivery** — where secrets live and how they reach your process. blindenv solves a different problem: **what happens after delivery**, when an AI agent is the one running the process.

| Capability | Secret managers | blindenv |
|---|---|---|
| Centralized secret storage | Yes | — (uses your existing `.env`) |
| Runtime injection into processes | Yes | Yes |
| Output redaction | — | Yes |
| File masking (`[BLINDED]` values) | — | Yes |
| Auto-detection by variable name | — | Yes |
| Config tamper-proofing | — | Yes |
| AI agent tool hooks | — | Yes |

Secret managers and blindenv are complementary. A secret manager gets the right values into your `.env` or CI pipeline. blindenv makes sure the agent that runs your commands can **use** those values without **seeing** them.

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
