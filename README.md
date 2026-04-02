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

From this moment, the agent reads your `.env` and sees the full structure — but every secret value is replaced with `[BLINDED]`. The magic: Claude Code can still use those values when running commands. The real values are injected into the subprocess invisibly, and any output containing a secret is automatically redacted.

---

## blindenv.yml

A `blindenv.yml` is auto-created in your project root on first run. If it wasn't created, make one yourself:

```yaml
# blindenv.yml
secret_files:
  - .env
```

That's the minimum. Open it to add more files or options:

```yaml
# blindenv.yml
secret_files:
  - .env
  - .env.local
  - secrets.yaml

mask_keys:
  - MY_CUSTOM_VAR      # mask a specific env var by name

mask_patterns:
  - KEY                # mask any env var whose name contains "KEY"
  - TOKEN
```

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
