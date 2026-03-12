#!/bin/sh

set -eu

REPO="basilysf1709/ship"
INSTALL_DIR="${HOME}/.local/bin"
BIN_PATH="${INSTALL_DIR}/ship"
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "${TMP_DIR}"
}

trap cleanup EXIT INT TERM

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

download() {
  url="$1"
  dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
    return 0
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -qO "$dest" "$url"
    return 0
  fi
  echo "missing required command: curl or wget" >&2
  exit 1
}

uname_s="$(uname -s)"
uname_m="$(uname -m)"

case "$uname_s" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  *)
    echo "unsupported OS: $uname_s" >&2
    exit 1
    ;;
esac

case "$uname_m" in
  arm64|aarch64) arch="arm64" ;;
  x86_64|amd64) arch="amd64" ;;
  *)
    echo "unsupported architecture: $uname_m" >&2
    exit 1
    ;;
esac

mkdir -p "${INSTALL_DIR}"

latest_url="https://github.com/${REPO}/releases/latest/download"
asset_names="
ship-${os}-${arch}
ship
"

for asset in $asset_names; do
  if download "${latest_url}/${asset}" "${TMP_DIR}/ship" 2>/dev/null; then
    chmod +x "${TMP_DIR}/ship"
    mv "${TMP_DIR}/ship" "${BIN_PATH}"
    echo "installed ship to ${BIN_PATH}"
    echo "ensure ${INSTALL_DIR} is in your PATH"
    exit 0
  fi
done

need_cmd git
need_cmd go

SRC_DIR="${TMP_DIR}/src"
git clone --depth 1 "https://github.com/${REPO}.git" "${SRC_DIR}" >/dev/null 2>&1
(cd "${SRC_DIR}" && go build -o "${BIN_PATH}")

echo "installed ship to ${BIN_PATH}"
echo "ensure ${INSTALL_DIR} is in your PATH"
