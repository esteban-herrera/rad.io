#!/bin/sh
set -e

REPO="github.com/esteban-herrera/rad.io"
INSTALL_DIR="$HOME/.local/bin"
BINARY="rad.io"

# ── colours ────────────────────────────────────────────────────────────────
if [ -t 1 ]; then
  RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
  BOLD='\033[1m'; RESET='\033[0m'
else
  RED=''; GREEN=''; YELLOW=''; BOLD=''; RESET=''
fi

info()  { printf "  ${GREEN}✓${RESET}  %s\n" "$1"; }
warn()  { printf "  ${YELLOW}!${RESET}  %s\n" "$1"; }
die()   { printf "  ${RED}✗${RESET}  %s\n" "$1" >&2; exit 1; }
title() { printf "\n${BOLD}%s${RESET}\n" "$1"; }

# ── check mpv ──────────────────────────────────────────────────────────────
title "Checking dependencies"

if command -v mpv >/dev/null 2>&1; then
  info "mpv found: $(mpv --version | head -1)"
else
  warn "mpv not found — attempting to install"

  if command -v brew >/dev/null 2>&1; then
    brew install mpv
  elif command -v apt-get >/dev/null 2>&1; then
    sudo apt-get update -qq && sudo apt-get install -y mpv
  elif command -v dnf >/dev/null 2>&1; then
    sudo dnf install -y mpv
  elif command -v pacman >/dev/null 2>&1; then
    sudo pacman -Sy --noconfirm mpv
  else
    die "Could not install mpv automatically. Please install it manually: https://mpv.io/installation/"
  fi

  command -v mpv >/dev/null 2>&1 || die "mpv installation failed"
  info "mpv installed"
fi

# ── check Go ───────────────────────────────────────────────────────────────
if ! command -v go >/dev/null 2>&1; then
  die "Go is required but not installed. Install it from https://go.dev/dl/ then re-run this script."
fi

GO_VERSION=$(go version | sed 's/go version go\([0-9]*\.[0-9]*\).*/\1/')
GO_MAJOR=$(printf '%s' "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(printf '%s' "$GO_VERSION" | cut -d. -f2)

if [ "$GO_MAJOR" -lt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 22 ]; }; then
  die "Go 1.22 or newer is required (found $GO_VERSION). Update at https://go.dev/dl/"
fi

info "Go $GO_VERSION found"

# ── build and install ──────────────────────────────────────────────────────
title "Installing rad.io"

mkdir -p "$INSTALL_DIR"

# Try go install first (clean, versioned)
if GOBIN="$INSTALL_DIR" go install "${REPO}@latest" 2>/dev/null; then
  info "Installed via go install"
else
  # Fall back to building from a local clone (useful during development)
  warn "go install failed — building from current directory"
  SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  go build -o "$INSTALL_DIR/$BINARY" "$SCRIPT_DIR"
  info "Built and installed from source"
fi

# ── PATH check ─────────────────────────────────────────────────────────────
title "Done"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    warn "$INSTALL_DIR is not in your PATH"
    printf "\n  Add this line to your shell config (e.g. ~/.zshrc or ~/.bashrc):\n\n"
    printf "    ${BOLD}export PATH=\"\$HOME/.local/bin:\$PATH\"${RESET}\n\n"
    printf "  Then reload your shell:\n\n"
    printf "    ${BOLD}source ~/.zshrc${RESET}   # or ~/.bashrc\n\n"
    ;;
esac

info "rad.io is ready — run: ${BOLD}rad.io${RESET}"
