# Advanced Configuration

By default, blindenv operates in **blind mode** — the agent can read secret files but all values are replaced with `[BLINDED]`. This is the recommended mode for most use cases.

For higher security requirements, two additional modes are available.

---

## Security Modes

```yaml
# blindenv.yml
mode: block    # blind (default) | block | stash
```

| Mode | Secret file access | Secret values | `ls` reveals files? |
|------|-------------------|--------------|---------------------|
| **`blind`** (default) | Readable (masked) | `[BLINDED]` | Yes |
| **`block`** | Explicit deny | Hidden | Yes |
| **`stash`** | File not found | Hidden | **No** (physically removed) |

### `block` — Explicit Deny

The agent sees "access denied" when it tries to read secret files. It knows the files exist but cannot access them. Use this when you want clear feedback about what's protected.

### `stash` — Complete Invisibility

The strongest mode. At session start, secret files are moved to a secure cache (`~/.cache/blindenv/`) and physically deleted from disk. Even `ls`, `find`, and `tree` in Bash reveal nothing. The files genuinely don't exist during the session.

Secrets remain fully functional — they're served from cache for injection and output blinding. Use `blindenv cache-restore` to bring them back after the session.

---

## Strict Environment Mode

By default, the subprocess inherits the full host environment (permissive mode). For high-security environments, use `passthrough` to restrict which non-secret variables are available:

```yaml
# blindenv.yml
passthrough:         # only these vars are inherited (strict mode)
  - PATH
  - HOME
  - LANG
```

When `passthrough` is set, the subprocess gets ONLY those vars plus injected secrets. Without it, the full host env is inherited.

---

## Additional CLI Commands

These commands are used with `block` and `stash` modes:

```
blindenv stash                        Move secret files to cache, delete originals
blindenv cache-restore                Restore secret files from cache
blindenv cache-refresh                Re-cache secret files (after you edit .env)
blindenv check-file <path>            Check if file is blocked (exit 2 = blocked)
blindenv has-config                   Exit 0 if config with secrets exists, 1 otherwise
```
