#!/bin/bash
# CanYouHack Steg — Installer
# Works on Linux & macOS. Copies the pre-built binary to /usr/local/bin.

set -e

BINARY="steg"
INSTALL_DIR="/usr/local/bin"

echo ""
echo "  CanYouHack Steg — Installer"
echo "  ─────────────────────────────"

if [ "$EUID" -ne 0 ]; then
  echo "  ✗ Please run as root: sudo ./install.sh"
  exit 1
fi

if [ ! -f "$BINARY" ]; then
  echo "  ✗ Binary '$BINARY' not found in current directory."
  echo "    Build it first with: go build -o steg ./"
  exit 1
fi

install -Dm755 "$BINARY" "$INSTALL_DIR/$BINARY"
echo "  ✓ Installed to $INSTALL_DIR/$BINARY"
echo "  Run 'steg --help' to get started."
echo ""
