#!/bin/bash
set -e

echo "🚀 Setting up vendor library management..."
echo ""

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install Node.js and npm first."
    echo "   Visit: https://nodejs.org/"
    exit 1
fi

echo "npm found: $(npm --version)"
echo ""

# Install dependencies
echo "Installing npm dependencies..."
npm install
echo ""

# Sync vendor files
echo "Syncing vendor files to internal/web/static/vendor/..."
npm run vendor:sync
echo ""

echo "Setup complete!"
echo ""
echo "Quick reference:"
echo "   make vendor-check    - Check for updates"
echo "   make vendor-update   - Update libraries (safe)"
echo "   make vendor-sync     - Re-sync vendor files"
echo ""
echo "Full documentation: docs/vendor-management.md"
