#!/bin/bash
set -e

echo "Setting up vendor library management..."
echo ""

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install Node.js and npm first."
    echo "   Visit: https://nodejs.org/"
    exit 1
fi

echo "  npm found: $(npm --version)"
echo ""

# Install dependencies
echo "Installing npm dependencies..."
npm install --no-audit --no-fund
echo ""

# Sync vendor files
echo "Syncing vendor files..."
node "$(dirname "$0")/sync-vendor.js"
echo ""

# Verify
echo "Verifying vendor files..."
node "$(dirname "$0")/verify-vendor.js"
echo ""

echo "Setup complete!"
echo ""
echo "  Quick reference:"
echo "    make vendor            - Full setup (install + sync + verify)"
echo "    make vendor-check      - Check for outdated packages"
echo "    make vendor-status     - Show current versions"
echo "    make vendor-update     - Update (minor/patch)"
echo "    make vendor-update-major - Update (incl. major)"
echo "    make vendor-cleanup    - Remove orphaned files"
echo ""
echo "  Full documentation: docs/Development.md"
