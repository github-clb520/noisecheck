#!/usr/bin/env bash
# NoiseCheck — 跨平台构建脚本
set -euo pipefail

BIN_NAME="noisecheck"
VERSION="${NC_VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
OUTPUT_DIR="${1:-./bin}"

PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

echo "🔨 NoiseCheck v${VERSION} — 跨平台构建"
echo "═══════════════════════════════════════"
echo ""

mkdir -p "$OUTPUT_DIR"

for platform in "${PLATFORMS[@]}"; do
  GOOS="${platform%%/*}"
  GOARCH="${platform##*/}"

  output_name="${BIN_NAME}"
  ext=""
  if [ "$GOOS" = "windows" ]; then
    ext=".exe"
  fi

  # Linux builds use a suffix
  full_name="${output_name}-${GOOS}-${GOARCH}${ext}"

  echo "  🚀  $GOOS/$GOARCH → $OUTPUT_DIR/$full_name"

  GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "$OUTPUT_DIR/$full_name" \
    ./cmd/noisecheck/

  # Calculate hash
  if command -v sha256sum &>/dev/null; then
    sha256sum "$OUTPUT_DIR/$full_name" | awk '{print $1}' > "$OUTPUT_DIR/$full_name.sha256"
  fi

  # Show size
  size=$(stat --format=%s "$OUTPUT_DIR/$full_name" 2>/dev/null || stat -f%z "$OUTPUT_DIR/$full_name" 2>/dev/null)
  echo "         📦 $(numfmt --to=iec "$size" 2>/dev/null || echo "$size bytes")"
done

echo ""
echo "✅ 构建完成。文件位于: $OUTPUT_DIR/"
ls -1 "$OUTPUT_DIR"/"${BIN_NAME}"-* 2>/dev/null || true
