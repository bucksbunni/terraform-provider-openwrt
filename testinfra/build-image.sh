#!/bin/bash
# Builds the OpenWrt VM image used by the local acceptance-test infrastructure
# (see testinfra/). Uses the official OpenWrt ImageBuilder container image to
# bake in the packages and files/ overlay needed for the JSON-RPC acceptance
# tests, then converts the result to a qcow2 disk image.
#
# Requires podman (used to run the ImageBuilder; the host's own perl/build
# toolchain is not required) and qemu-img.
#
# Usage: ./testinfra/build-image.sh [--force]
#   --force  rebuild even if testinfra/build/openwrt-acceptance.qcow2 exists

set -euo pipefail

OPENWRT_VERSION="24.10.2"
IMAGEBUILDER_IMAGE="docker.io/openwrt/imagebuilder:x86-64-${OPENWRT_VERSION}"

# Packages installed on top of the "generic" profile defaults:
# - luci-mod-rpc, luci-lib-ipkg, luci-compat: provide the /cgi-bin/luci/rpc/*
#   JSON-RPC backend used by the provider's client (matches README.md's
#   "Requirements" section).
# - uhttpd: web server. Without it nothing listens on port 80 and the
#   JSON-RPC endpoint above is unreachable.
# - luci-theme-bootstrap: OpenWrt 24.10's ucode-based LuCI dispatcher
#   references the configured theme even for "notemplate" API routes.
#   Without a theme installed, /cgi-bin/luci/rpc/* returns a 500
#   ("Failed to load template 'themes/bootstrap/header'") instead of JSON,
#   even though luci-mod-rpc itself works fine.
# - kmod-wireguard, wireguard-tools: enable RequireWireguard-gated
#   acceptance tests (see image/files/etc/modules.d/30-wireguard).
# - kmod-mac80211-hwsim, wpad-mbedtls, wireless-tools: simulated wireless
#   radios for RequireWireless-gated acceptance tests (see
#   image/files/etc/modules.d/30-mac80211-hwsim and the `wifi config` step
#   in image/files/etc/uci-defaults/99-acceptance-setup).
PACKAGES="luci-mod-rpc luci-lib-ipkg luci-compat uhttpd luci-theme-bootstrap kmod-wireguard wireguard-tools kmod-mac80211-hwsim wpad-mbedtls wireless-tools"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="${SCRIPT_DIR}/build"
DL_CACHE_DIR="${BUILD_DIR}/dl-cache"
OUTPUT_DIR="${BUILD_DIR}/output"
FILES_DIR="${SCRIPT_DIR}/image/files"
OUT_QCOW2="${BUILD_DIR}/openwrt-acceptance.qcow2"

FORCE=0
if [ "${1:-}" = "--force" ]; then
  FORCE=1
fi

if [ "$FORCE" = "0" ] && [ -f "$OUT_QCOW2" ]; then
  echo "Reusing existing image: $OUT_QCOW2 (use --force to rebuild)"
  ls -lh "$OUT_QCOW2"
  exit 0
fi

mkdir -p "$DL_CACHE_DIR" "$OUTPUT_DIR"
rm -rf "${OUTPUT_DIR:?}"/*

echo "Building OpenWrt image (profile=generic) via ${IMAGEBUILDER_IMAGE}..."
podman run --rm \
  --userns=keep-id \
  -v "${FILES_DIR}:/files:ro,Z" \
  -v "${DL_CACHE_DIR}:/builder/dl:Z" \
  -v "${OUTPUT_DIR}:/output:Z" \
  "$IMAGEBUILDER_IMAGE" \
  make image \
    PROFILE=generic \
    PACKAGES="$PACKAGES" \
    FILES=/files \
    BIN_DIR=/output \
    ROOTFS_PARTSIZE=200

RAW_GZ=$(find "$OUTPUT_DIR" -name "*-generic-ext4-combined.img.gz" | head -n1)
if [ -z "$RAW_GZ" ]; then
  echo "ERROR: could not find a built *-generic-ext4-combined.img.gz image" >&2
  exit 1
fi

echo "Converting $(basename "$RAW_GZ") to qcow2..."
RAW_IMG="${BUILD_DIR}/openwrt-acceptance.img"
# OpenWrt's combined image is zero-padded after the gzip stream, which makes
# gunzip exit 2 ("trailing garbage ignored") even though decompression
# succeeded; only treat other exit codes as fatal.
set +e
gunzip -c "$RAW_GZ" > "$RAW_IMG"
gz_rc=$?
set -e
if [ "$gz_rc" -ne 0 ] && [ "$gz_rc" -ne 2 ]; then
  echo "ERROR: gunzip failed with exit code $gz_rc" >&2
  exit "$gz_rc"
fi
qemu-img convert -O qcow2 -c "$RAW_IMG" "$OUT_QCOW2"
rm -f "$RAW_IMG"

echo ""
echo "Image built: $OUT_QCOW2"
ls -lh "$OUT_QCOW2"
