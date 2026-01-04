#!/bin/bash

# Build script for Snipo Browser Extension
# Creates distribution packages for Chrome and Firefox

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/dist"
VERSION=$(grep '"version"' "$SCRIPT_DIR/manifest.json" | head -1 | sed 's/.*"version": "\(.*\)".*/\1/')

echo "Building Snipo Extension v$VERSION"
echo "=================================="

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Files to include in both packages
FILES=(
    "background.js"
    "content.js"
    "styles.css"
    "icons"
    "options"
    "PRIVACY.md"
)

# Build Chrome package
echo ""
echo "Building Chrome package..."
CHROME_DIR="$BUILD_DIR/chrome"
mkdir -p "$CHROME_DIR"

# Copy files
for file in "${FILES[@]}"; do
    if [ -d "$SCRIPT_DIR/$file" ]; then
        cp -r "$SCRIPT_DIR/$file" "$CHROME_DIR/"
    else
        cp "$SCRIPT_DIR/$file" "$CHROME_DIR/"
    fi
done

# Copy Chrome manifest
cp "$SCRIPT_DIR/manifest-chrome.json" "$CHROME_DIR/manifest.json"

# Create Chrome zip 
cd "$CHROME_DIR"
zip -r "../snipo-chrome-v$VERSION.zip" . -x "*.DS_Store" "*/.DS_Store"
cd "$BUILD_DIR"
echo "✓ Chrome package created: dist/snipo-chrome-v$VERSION.zip"

# Build Firefox package
echo ""
echo "Building Firefox package..."
FIREFOX_DIR="$BUILD_DIR/firefox"
mkdir -p "$FIREFOX_DIR"

# Copy files
for file in "${FILES[@]}"; do
    if [ -d "$SCRIPT_DIR/$file" ]; then
        cp -r "$SCRIPT_DIR/$file" "$FIREFOX_DIR/"
    else
        cp "$SCRIPT_DIR/$file" "$FIREFOX_DIR/"
    fi
done

# Copy Firefox manifest
cp "$SCRIPT_DIR/manifest.json" "$FIREFOX_DIR/manifest.json"

# Create Firefox zip 
cd "$FIREFOX_DIR"
zip -r "../snipo-firefox-v$VERSION.zip" . -x "*.DS_Store" "*/.DS_Store"
cd "$BUILD_DIR"
echo "✓ Firefox package created: dist/snipo-firefox-v$VERSION.zip"

# Create source code archive (required by Firefox)
echo ""
echo "Creating source code archive for Firefox review..."
cd "$SCRIPT_DIR/.."
zip -r "$BUILD_DIR/snipo-source-v$VERSION.zip" extension/ \
    -x "extension/dist/*" \
    -x "extension/.DS_Store" \
    -x "extension/node_modules/*" \
    -x "*.git*"
echo "✓ Source archive created: dist/snipo-source-v$VERSION.zip"

# Summary
echo ""
echo "Build complete!"
echo "=================================="
echo "Chrome package:  dist/snipo-chrome-v$VERSION.zip"
echo "Firefox package: dist/snipo-firefox-v$VERSION.zip"
echo "Source archive:  dist/snipo-source-v$VERSION.zip"
echo ""
echo "Next steps:"
echo "1. Test the packages locally"
echo "2. Upload to Chrome Web Store and Firefox Add-ons"
