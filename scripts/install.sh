#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="${REPO_OWNER:-${GITHUB_OWNER:-yindia}}"
REPO_NAME="${REPO_NAME:-${GITHUB_REPO:-pltf}}"
VERSION="${VERSION:-latest}"
DEST="${DEST:-/usr/local/bin}"

uname_os() {
  case "$(uname -s)" in
    Linux*) echo "linux" ;;
    Darwin*) echo "darwin" ;;
    CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
    *) echo "unsupported" ;;
  esac
}

uname_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported" ;;
  esac
}

OS="$(uname_os)"
ARCH="$(uname_arch)"

if [[ "$OS" == "unsupported" || "$ARCH" == "unsupported" ]]; then
  echo "Unsupported platform: $(uname -s)/$(uname -m)" >&2
  exit 1
fi

if [[ "$VERSION" == "latest" ]]; then
  VERSION="$(curl -s https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest | grep -Eo '\"tag_name\": \"v?[0-9\\.]+\"' | head -1 | cut -d'\"' -f4)"
fi

if [[ -z "${VERSION}" ]]; then
  echo "Unable to determine version" >&2
  exit 1
fi

TARBALL="pltf_${VERSION}_${OS}_${ARCH}.tar.gz"
if [[ "$OS" == "windows" ]]; then
  TARBALL="pltf_${VERSION}_${OS}_${ARCH}.zip"
fi

URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/${TARBALL}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading ${URL}..."
curl -sSLf "$URL" -o "${TMPDIR}/${TARBALL}"

if [[ "$OS" == "windows" ]]; then
  unzip -q "${TMPDIR}/${TARBALL}" -d "${TMPDIR}"
  BIN="${TMPDIR}/pltf.exe"
else
  tar -xzf "${TMPDIR}/${TARBALL}" -C "${TMPDIR}"
  BIN="${TMPDIR}/pltf"
fi

chmod +x "$BIN"
mkdir -p "$DEST"
cp "$BIN" "$DEST/"

echo "Installed pltf ${VERSION} to ${DEST}"
