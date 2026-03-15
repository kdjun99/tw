#!/bin/sh
set -e

REPO="kdjun99/tw"
INSTALL_DIR="${TW_INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release. Falling back to go install..."
  if command -v go >/dev/null 2>&1; then
    go install "github.com/dongjunkim/tw@latest"
    echo "Installed tw via go install"
    exit 0
  else
    echo "Error: No release found and Go is not installed."
    exit 1
  fi
fi

# Download binary
FILENAME="tw-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

echo "Downloading tw ${LATEST} for ${OS}/${ARCH}..."
mkdir -p "$INSTALL_DIR"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "${INSTALL_DIR}/tw" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "${INSTALL_DIR}/tw" "$URL"
else
  echo "Error: curl or wget required"
  exit 1
fi

chmod +x "${INSTALL_DIR}/tw"

# Verify
if [ -x "${INSTALL_DIR}/tw" ]; then
  echo "Successfully installed tw to ${INSTALL_DIR}/tw"
  echo ""
  if ! echo "$PATH" | tr ':' '\n' | grep -q "^${INSTALL_DIR}$"; then
    echo "Add ${INSTALL_DIR} to your PATH:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
  fi
  echo "Run 'tw --help' to get started."
else
  echo "Installation failed."
  exit 1
fi
