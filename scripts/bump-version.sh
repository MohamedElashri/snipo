#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo "Usage: ./scripts/bump-version.sh <version>"
    echo "Example: ./scripts/bump-version.sh 2.0.0"
    exit 1
fi

NEW_VERSION=$1
echo "Bumping versions to $NEW_VERSION..."

# Update Go backend
sed -i "s/const Current = \".*\"/const Current = \"$NEW_VERSION\"/" internal/version/version.go

# Update Snippy TUI
sed -i "s/Version = \".*\"/Version = \"$NEW_VERSION\"/" tui/cmd/snippy/main.go

# Update Browser Ext
sed -i "s/\"version\": \".*\"/\"version\": \"$NEW_VERSION\"/" extension/manifest.json
sed -i "s/\"version\": \".*\"/\"version\": \"$NEW_VERSION\"/" extension/manifest-chrome.json

# Update VS Code Ext
sed -i "s/\"version\": \".*\"/\"version\": \"$NEW_VERSION\"/" vscode-extension/package.json

echo "Done! Run 'git diff' to verify."
