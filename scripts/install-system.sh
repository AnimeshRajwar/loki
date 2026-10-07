#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-AnimeshRajwar/loki}"
CHANNEL="${1:-stable}" # stable | latest
INSTALL_DIR="${2:-/usr/local/bin}"

case "$(uname -s)" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $(uname -s)"
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $(uname -m)"
    exit 1
    ;;
esac

ASSET="loki-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${CHANNEL}/${ASSET}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${URL}"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$TMP_DIR/$ASSET"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP_DIR/$ASSET" "$URL"
else
  echo "Neither curl nor wget is installed. Please install one of them and retry."
  exit 1
fi

echo "Extracting"
mkdir -p "$TMP_DIR/extract"
tar -xzf "$TMP_DIR/$ASSET" -C "$TMP_DIR/extract"

BINPATH=$(find "$TMP_DIR/extract" -type f -name loki -print -quit)
if [[ -z "$BINPATH" ]]; then
  echo "Binary 'loki' not found in archive"
  exit 1
fi

if [[ ! -d "$INSTALL_DIR" ]]; then
  echo "Creating $INSTALL_DIR (requires sudo)"
  sudo mkdir -p "$INSTALL_DIR"
fi

echo "Installing to $INSTALL_DIR (sudo may be required)"
sudo cp "$BINPATH" "$INSTALL_DIR/loki"
sudo chmod 0755 "$INSTALL_DIR/loki"

echo "Installed: $INSTALL_DIR/loki"
echo "Run 'loki --help'. If command not found, ensure $INSTALL_DIR is in PATH or add it system-wide."
