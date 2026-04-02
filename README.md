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

```bash
/plugin marketplace add neuradex/blindenv
/plugin install blindenv@blindenv
/reload-plugins
```

That's it. From this moment, the agent reads your `.env` and sees the full structure — but every secret value is replaced with `[BLINDED]`. The magic: Claude Code can still use those values when running commands. The real values are injected into the subprocess invisibly, and any output that contains a secret is automatically redacted too.

---

## Need more control?

A `blindenv.yml` is auto-created in your project root. Open it to configure which files are protected:

```yaml
secret_files:
  - .env
  - .env.local
  - secrets.yaml
```

You can also explicitly mask specific environment variables or add name patterns:

```yaml
secret_files:
  - .env

mask_env:
  - MY_CUSTOM_VAR      # mask a specific variable

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
