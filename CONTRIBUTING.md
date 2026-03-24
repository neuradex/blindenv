<!-- AGENT NOTE: If you modify this file, also update docs/CONTRIBUTING.ko.md (Korean) and docs/CONTRIBUTING.ja.md (Japanese) to keep translations in sync. -->

# Contributing to blindenv

<p align="center">
  <strong>English</strong> ·
  <a href="./docs/CONTRIBUTING.ko.md">한국어</a> ·
  <a href="./docs/CONTRIBUTING.ja.md">日本語</a>
</p>

## Prerequisites

- Go 1.22+
- Make

## Development Setup

```bash
git clone https://github.com/neuradex/blindenv.git
cd blindenv
make build
```

This produces a `./blindenv` binary in the project root.

## Common Commands

```bash
make build    # Build local binary
make test     # Run all tests
make vet      # Run go vet
make clean    # Remove built binary
```

## Project Structure

```
blindenv/
├── main.go                  # Entrypoint
├── cmd/
│   ├── root.go              # CLI dispatcher (run, init, check-file, ...)
│   └── hook.go              # Hook handlers (bash, read, grep, glob, guard-file)
├── config/
│   └── config.go            # YAML config loading and discovery
├── engine/
│   ├── exec.go              # Subprocess execution with secret isolation
│   ├── secrets.go           # Secret resolution, caching, redaction
│   └── file_guard.go        # File access checks (path match, content scan)
├── provider/
│   ├── provider.go          # Platform-agnostic hook interface
│   └── cc/
│       └── cc.go            # Claude Code provider implementation
├── .claude-plugin/
│   ├── plugin.json          # Plugin metadata
│   └── hooks.json           # Claude Code hook configuration
└── scripts/
    └── session-start.sh     # Auto-install + init on session start
```

## Architecture

blindenv has two execution modes:

1. **`blindenv run '<cmd>'`** — Runs a command in an isolated subprocess with secrets injected and output redacted.
2. **`blindenv hook cc <hook>`** — PreToolUse hook handlers that intercept Claude Code tool calls before they execute.

Hooks read JSON from stdin (Claude Code hook protocol), apply security logic, and respond with allow/block/modify actions via stdout/stderr + exit code.

The `provider` package abstracts the hook protocol so adding support for other AI coding agents (e.g. Cursor, Windsurf) only requires implementing the `Provider` interface.

## Adding a New Provider

1. Create `provider/<name>/<name>.go` implementing `provider.Provider`
2. Register it in `cmd/hook.go` → `resolveProvider()`
3. Add hook configuration for the new platform

## Testing

```bash
make test
```

Tests live alongside source files (`engine/*_test.go`). When adding new features, include tests for the engine layer at minimum.

## Submitting Changes

1. Fork the repo and create a feature branch
2. Make your changes
3. Run `make test && make vet`
4. Open a PR with a clear description of what and why
