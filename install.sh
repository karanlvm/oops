#!/usr/bin/env sh
set -e

REPO="karan/oops"
BINARY="oops"
INSTALL_DIR="/usr/local/bin"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin|linux) ;;
  *) echo "oops install: unsupported OS: $OS" && exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)       ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "oops install: unsupported architecture: $ARCH" && exit 1 ;;
esac

# Resolve latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "oops install: could not determine latest release (check https://github.com/$REPO/releases)"
  exit 1
fi

ASSET="${BINARY}-${OS}-${ARCH}"
URL="https://github.com/$REPO/releases/download/$LATEST/$ASSET"
TMP=$(mktemp)

echo "Downloading oops $LATEST ($OS/$ARCH)..."
curl -fsSL --progress-bar -o "$TMP" "$URL"
chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "$INSTALL_DIR/$BINARY"
else
  echo "Installing to $INSTALL_DIR (sudo required)..."
  sudo mv "$TMP" "$INSTALL_DIR/$BINARY"
fi

echo ""
echo "oops $LATEST installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Next steps:"
echo "  1. Set an LLM key in your shell profile:"
echo "       export ANTHROPIC_API_KEY=sk-ant-..."
echo "       # or: export OPENAI_API_KEY=sk-..."
echo ""
echo "  2. Add shell hooks:"
echo "       oops --install"
echo ""
echo "  3. Restart your shell, then break something."
