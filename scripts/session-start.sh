#!/bin/bash
# blindenv SessionStart hook.
# 1. Ensure binary is installed
# 2. Add plugin bin to shell PATH (zshrc/bashrc)
# 3. Auto-create blindenv.yml if not found

PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
BIN_DIR="$PLUGIN_ROOT/bin"
BIN="$BIN_DIR/blindenv"

# ── 1. Install binary if missing ─────────────────────────────────
if [ ! -x "$BIN" ]; then
  REPO="neuradex/blindenv"
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$ARCH" in
    x86_64)        ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
  esac

  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
  if [ -z "$VERSION" ]; then
    exit 0  # network error — skip silently, retry next session
  fi

  VERSION_NUM="${VERSION#v}"
  TARBALL="blindenv_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

  TMP="$(mktemp -d)"
  trap 'rm -rf "$TMP"' EXIT

  curl -fsSL "$URL" -o "${TMP}/${TARBALL}" 2>/dev/null || exit 0
  tar -xzf "${TMP}/${TARBALL}" -C "$TMP" 2>/dev/null || exit 0

  mkdir -p "$BIN_DIR"
  mv "${TMP}/blindenv" "$BIN"
  chmod +x "$BIN"
fi

# ── 2. Ensure plugin bin dir is in shell PATH ────────────────────
MARKER="# [blindenv] plugin bin"
EXPORT_LINE="export PATH=\"$BIN_DIR:\$PATH\" $MARKER"

for RC in "$HOME/.zshrc" "$HOME/.bashrc"; do
  [ ! -f "$RC" ] && continue
  if grep -q "$MARKER" "$RC" 2>/dev/null; then
    # Update path if plugin location changed
    sed -i '' "/$MARKER/c\\
$EXPORT_LINE" "$RC" 2>/dev/null
  else
    echo "$EXPORT_LINE" >> "$RC"
  fi
done

# ── 3. Auto-create blindenv.yml if not found ─────────────────────
"$BIN" init

exit 0
