#!/bin/bash
# NexusGate firmware build script using OpenWrt ImageBuilder
# Usage: ./build.sh [profile]
# Example: ./build.sh x86-64

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FIRMWARE_DIR="$(dirname "$SCRIPT_DIR")"
PROFILE="${1:-x86-64}"
OPENWRT_VERSION="23.05.5"
BUILD_DIR="/tmp/nexusgate-build"
OUTPUT_DIR="$FIRMWARE_DIR/output"

# Load profile
PROFILE_FILE="$FIRMWARE_DIR/profiles/${PROFILE}.conf"
if [ ! -f "$PROFILE_FILE" ]; then
    echo "Error: profile '$PROFILE' not found at $PROFILE_FILE"
    exit 1
fi
source "$PROFILE_FILE"

echo "=== NexusGate Firmware Builder ==="
echo "Profile: $PROFILE"
echo "OpenWrt: $OPENWRT_VERSION"
echo "Target:  $TARGET/$SUBTARGET"
echo ""

# Download ImageBuilder if not exists
IB_NAME="openwrt-imagebuilder-${OPENWRT_VERSION}-${TARGET}-${SUBTARGET}.Linux-x86_64"
IB_URL="https://downloads.openwrt.org/releases/${OPENWRT_VERSION}/targets/${TARGET}/${SUBTARGET}/${IB_NAME}.tar.xz"
IB_DIR="$BUILD_DIR/$IB_NAME"

mkdir -p "$BUILD_DIR" "$OUTPUT_DIR"

if [ ! -d "$IB_DIR" ]; then
    echo "Downloading ImageBuilder..."
    wget -q -O "$BUILD_DIR/${IB_NAME}.tar.xz" "$IB_URL"
    tar -xf "$BUILD_DIR/${IB_NAME}.tar.xz" -C "$BUILD_DIR"
fi

# Read package list
PACKAGES=$(cat "$FIRMWARE_DIR/packages/enterprise.txt" | grep -v '^#' | grep -v '^$' | tr '\n' ' ')

# Copy custom files
FILES_DIR="$FIRMWARE_DIR/files"

# Build firmware
echo "Building firmware with packages: $PACKAGES"
cd "$IB_DIR"

make image \
    PROFILE="$DEVICE_PROFILE" \
    PACKAGES="$PACKAGES" \
    FILES="$FILES_DIR" \
    EXTRA_IMAGE_NAME="nexusgate"

# Copy output
cp -r "$IB_DIR/bin/targets/$TARGET/$SUBTARGET/"* "$OUTPUT_DIR/"
echo ""
echo "=== Build complete ==="
echo "Output: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR/"*.img* 2>/dev/null || ls -lh "$OUTPUT_DIR/"*.bin* 2>/dev/null || true
