#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-AnimeshRajwar/loki}"
CHANNEL="${1:-stable}" # stable | latest
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

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
ARCHIVE_PATH="${TMP_DIR}/${ASSET}"
EXTRACTED_DIR="${TMP_DIR}/loki-${OS}-${ARCH}"
RC_FILE="$HOME/.bashrc"

if [[ "${SHELL:-}" == *"zsh"* ]]; then
  RC_FILE="$HOME/.zshrc"
fi

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "Downloading ${URL}"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$ARCHIVE_PATH"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$ARCHIVE_PATH" "$URL"
else
  echo "Neither curl nor wget is installed. Please install one of them and retry."
  exit 1
fi

tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

mkdir -p "$INSTALL_DIR"
cp "${EXTRACTED_DIR}/loki" "${INSTALL_DIR}/loki"
chmod +x "${INSTALL_DIR}/loki"

PATH_EXPORT="export PATH=\"${INSTALL_DIR}:\$PATH\""
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
  if [[ -f "$RC_FILE" ]]; then
    if ! grep -Fq "$PATH_EXPORT" "$RC_FILE"; then
      echo "$PATH_EXPORT" >> "$RC_FILE"
      echo "Added PATH entry to ${RC_FILE}"
    fi
  else
    echo "$PATH_EXPORT" > "$RC_FILE"
    echo "Created ${RC_FILE} and added PATH entry"
  fi
  export PATH="${INSTALL_DIR}:$PATH"
fi

echo "Installed Loki to ${INSTALL_DIR}/loki"
echo "Run: loki --help"
echo "If command is not found, open a new terminal or run: source ${RC_FILE}"
