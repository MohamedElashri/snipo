#!/bin/bash

# Build script for Snipo Browser Extension
# Creates distribution packages for Chrome and Firefox

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/dist"
VERSION=$(grep '"version"' "$SCRIPT_DIR/manifest.json" | head -1 | sed 's/.*"version": "\(.*\)".*/\1/')
MODE="${1:-all}"

case "$MODE" in
    chrome|firefox|all) ;;
    *)
        echo "Usage: $0 [chrome|firefox|all]" >&2
        exit 1
        ;;
esac

echo "Building Snipo Extension v$VERSION"
echo "=================================="

clean_previous_builds() {
    case "$MODE" in
        chrome)
            rm -rf "$BUILD_DIR/chrome" "$BUILD_DIR/snipo-chrome-v$VERSION.zip"
            ;;
        firefox)
            rm -rf "$BUILD_DIR/firefox" "$BUILD_DIR/snipo-firefox-v$VERSION.zip" "$BUILD_DIR/snipo-source-v$VERSION.zip"
            ;;
        all)
            rm -rf "$BUILD_DIR"
            ;;
    esac
    mkdir -p "$BUILD_DIR"
}

# Clean previous builds
echo "Cleaning previous builds..."
clean_previous_builds

# Files to include in both packages
FILES=(
    "background.js"
    "content.js"
    "styles.css"
    "icons"
    "options"
    "PRIVACY.md"
)

copy_common_files() {
    local target_dir="$1"

    for file in "${FILES[@]}"; do
        if [ -d "$SCRIPT_DIR/$file" ]; then
            cp -r "$SCRIPT_DIR/$file" "$target_dir/"
        else
            cp "$SCRIPT_DIR/$file" "$target_dir/"
        fi
    done
}

build_chrome() {
    echo ""
    echo "Building Chrome package..."
    CHROME_DIR="$BUILD_DIR/chrome"
    mkdir -p "$CHROME_DIR"

    copy_common_files "$CHROME_DIR"
    cp "$SCRIPT_DIR/manifest-chrome.json" "$CHROME_DIR/manifest.json"

    cd "$CHROME_DIR"
    zip -r "../snipo-chrome-v$VERSION.zip" . -x "*.DS_Store" "*/.DS_Store"
    cd "$BUILD_DIR"
    echo "✓ Chrome package created: dist/snipo-chrome-v$VERSION.zip"
}

build_firefox() {
    echo ""
    echo "Building Firefox package..."
    FIREFOX_DIR="$BUILD_DIR/firefox"
    mkdir -p "$FIREFOX_DIR"

    copy_common_files "$FIREFOX_DIR"
    cp "$SCRIPT_DIR/manifest.json" "$FIREFOX_DIR/manifest.json"

    cd "$FIREFOX_DIR"
    zip -r "../snipo-firefox-v$VERSION.zip" . -x "*.DS_Store" "*/.DS_Store"
    cd "$BUILD_DIR"
    echo "✓ Firefox package created: dist/snipo-firefox-v$VERSION.zip"

    echo ""
    echo "Creating source code archive for Firefox review..."
    cd "$SCRIPT_DIR/.."
    zip -r "$BUILD_DIR/snipo-source-v$VERSION.zip" extension/ \
        -x "extension/dist/*" \
        -x "extension/.DS_Store" \
        -x "extension/node_modules/*" \
        -x "*.git*"
    echo "✓ Source archive created: dist/snipo-source-v$VERSION.zip"
}

case "$MODE" in
    chrome)
        build_chrome
        ;;
    firefox)
        build_firefox
        ;;
    all)
        build_chrome
        build_firefox
        ;;
esac

# Summary
echo ""
echo "Build complete!"
echo "=================================="
if [ "$MODE" = "chrome" ] || [ "$MODE" = "all" ]; then
    echo "Chrome package:  dist/snipo-chrome-v$VERSION.zip"
fi
if [ "$MODE" = "firefox" ] || [ "$MODE" = "all" ]; then
    echo "Firefox package: dist/snipo-firefox-v$VERSION.zip"
    echo "Source archive:  dist/snipo-source-v$VERSION.zip"
fi
echo ""
echo "Next steps:"
echo "1. Test the packages locally"
echo "2. Upload to Chrome Web Store and Firefox Add-ons"
