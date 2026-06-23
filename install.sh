#!/bin/sh
# Install NoiseCheck CLI from GitHub releases.
#   curl -fsSL https://raw.githubusercontent.com/github-clb520/noisecheck/main/install.sh | sh
# Env: NC_INSTALL_DIR (default /usr/local/bin or ~/.local/bin), NC_VERSION (default latest).
set -eu

main() {
  REPO="github-clb520/noisecheck"
  BIN="nc"
  ASSET_PREFIX="noisecheck"
  INSTALL_DIR="${NC_INSTALL_DIR:-}"
  VERSION="${NC_VERSION:-}"

  command -v curl >/dev/null 2>&1 || err "curl is required"

  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    darwin|linux) ;;
    mingw*|msys*|cygwin*) os="windows" ;;
    *) err "unsupported OS: $os (download from GitHub releases)" ;;
  esac

  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) err "unsupported architecture: $arch" ;;
  esac

  # Default install dir: try /usr/local/bin, fallback to ~/.local/bin
  if [ -z "$INSTALL_DIR" ]; then
    if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
      INSTALL_DIR="/usr/local/bin"
    else
      INSTALL_DIR="${HOME:-}/.local/bin"
    fi
  fi

  if [ -z "$VERSION" ]; then
    release_json="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest")" ||
      err "failed to fetch latest release info from GitHub API"
    VERSION="$(printf '%s' "$release_json" |
      sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
    [ -n "$VERSION" ] || err "could not resolve latest release tag"
  fi

  asset="${ASSET_PREFIX}-${os}-${arch}"
  if [ "$os" = "windows" ]; then
    asset="${asset}.exe"
  fi
  base="https://github.com/$REPO/releases/download/$VERSION"
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' INT TERM EXIT

  printf '📥 Downloading NoiseCheck %s (%s/%s)...\n' "$VERSION" "$os" "$arch"
  curl -fsSL -o "$tmp/$asset" "$base/$asset" || err "download failed: $base/$asset"
  curl -fsSL -o "$tmp/sha256sum.txt" "$base/sha256sum.txt" || err "sha256sum.txt download failed"

  want="$(awk -v a="$asset" '$2 == a {print tolower($1)}' "$tmp/sha256sum.txt")"
  [ -n "$want" ] || err "no checksum entry for $asset in sha256sum.txt"
  got="$(sha256 "$tmp/$asset" | awk '{print tolower($1)}')"
  [ "$got" = "$want" ] || err "checksum mismatch for $asset (got $got, want $want)"

  install_binary "$tmp/$asset" "$INSTALL_DIR" "$BIN"

  printf '\n✅ NoiseCheck %s installed → %s/%s\n\n' "$VERSION" "$INSTALL_DIR" "$BIN"
  printf '快速开始:\n'
  printf '  1. nc init         — 交互式配置 LLM\n'
  printf '  2. nc review        — 审查当前工作区变更\n'
  printf '  3. nc review -c HASH — 审查指定 commit\n'
  printf '\n文档: https://github.com/github-clb520/noisecheck\n'
  post_install_path_notice "$BIN" "$INSTALL_DIR"
}

install_binary() {
  src="$1"
  dir="$2"
  bin="$3"
  if mkdir -p "$dir" 2>/dev/null && [ -w "$dir" ]; then
    install -m 0755 "$src" "$dir/$bin"
  elif command -v sudo >/dev/null 2>&1; then
    printf 'note: %s is not writable; escalating with sudo\n' "$dir"
    sudo mkdir -p "$dir"
    sudo install -m 0755 "$src" "$dir/$bin"
  else
    err "$dir is not writable and sudo is unavailable; set NC_INSTALL_DIR to a writable path"
  fi
}

post_install_path_notice() {
  bin="$1"
  install_dir="$2"
  case ":$PATH:" in
    *":$install_dir:"*) ;;
    *) printf 'note: %s is not on your PATH; add it:\n  export PATH="\$PATH:%s"\n' "$install_dir" "$install_dir"; return ;;
  esac
  command -v "$bin" >/dev/null 2>&1 || printf 'note: open a new shell so %s resolves on PATH\n' "$bin"
}

sha256() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1"
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1"
  else
    err "shasum or sha256sum is required for checksum verification"
  fi
}

err() { printf '❌ error: %s\n' "$1" >&2; exit 1; }

main "$@"
