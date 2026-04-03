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
  <strong>English</strong> ·
  <a href="./docs/README.ko.md">한국어</a> ·
  <a href="./docs/README.ja.md">日本語</a>
</p>

---

## Install

First, add the plugin from the marketplace:

```
/plugin marketplace add neuradex/blindenv
```

Then install it — either in the folder you want to protect, or at the user level to cover all your projects:

```bash
# Project scope — protects this project only
cd /your/project
/plugin install blindenv@blindenv

# User scope — protects all projects
/plugin install blindenv@blindenv --user
```

Restart Claude Code. On the next session start, blindenv is active.

---

## How it works

When the agent reads your `.env`, it sees the full structure — variable names, comments — but every secret value is replaced with `[BLINDED]`. When the agent runs a command, blindenv injects the real values into the subprocess invisibly. Any output that contains a secret is automatically redacted too.

```
Agent reads .env:     API_KEY=[BLINDED]
                      DB_URL=[BLINDED]
                      DEBUG=true        ← non-secret values pass through

Agent runs command:   curl -H "Authorization: $API_KEY" https://api.example.com
                          ↓
blindenv injects:     real API_KEY value into subprocess env
                          ↓
Agent receives:       {"result": "ok", "token": "[BLINDED]"}
```

The values in `secret_files` take priority over any shell environment variables with the same name.

---

## blindenv.yml

A `blindenv.yml` is auto-created in your project root on first run. If it wasn't created, make one yourself:

```yaml
# blindenv.yml
secret_files:
  - .env
```

That's the minimum. Open it to add more:

```yaml
# blindenv.yml
secret_files:         # files to parse, mask, and inject into subprocesses
  - .env
  - .env.local

mask_keys:            # mask specific env vars by name (already in your shell env)
  - MY_CUSTOM_VAR

mask_patterns:        # mask any env var whose name contains these substrings
  - INTERNAL          # (defaults cover KEY, SECRET, TOKEN, PASSWORD, etc.)
```

**`secret_files` vs `mask_keys`:** use `secret_files` when the secret lives in a file. Use `mask_keys` when the secret is already exported in your shell or injected by CI/CD — no file involved.

> For advanced options (`block` mode, `passthrough`, etc.), see [Advanced Configuration](./docs/ADVANCED.md).

---

## Config discovery

blindenv walks up from your current directory to find the nearest `blindenv.yml` — just like `.gitignore`. A config in a parent directory applies to all projects underneath it.

To protect a specific folder separately, place a `blindenv.yml` there.

---

## License

MIT

---

<p align="center">
  <strong>Your agent doesn't need your keys. It needs what your keys unlock.</strong>
  <br>
  blindenv gives it access without exposure.
</p>
