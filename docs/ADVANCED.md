# Advanced Configuration

By default, blindenv operates in **blind mode** — the agent can read secret files but all values are replaced with `[BLINDED]`. This is the recommended mode for most use cases.

---

## Security Modes

```yaml
# blindenv.yml
mode: block    # blind (default) | block
```

| Mode | Secret file access | Secret values |
|------|-------------------|--------------|
| **`blind`** (default) | Readable (masked) | `[BLINDED]` |
| **`block`** | Explicit deny | Hidden |

### `block` — Explicit Deny

The agent sees "access denied" when it tries to read secret files. It knows the files exist but cannot access them. Use this when you want clear feedback about what's protected.

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

```
blindenv check-file <path>            Check if file is blocked (exit 2 = blocked)
blindenv has-config                   Exit 0 if config with secrets exists, 1 otherwise
```
