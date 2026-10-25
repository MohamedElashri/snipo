#!/usr/bin/env bash
set -euo pipefail
version=0.76.1
directory=$(mktemp -d)
trap 'rm -rf "$directory"' EXIT
base="https://github.com/aquasecurity/trivy/releases/download/v$version"
curl --fail --location --retry 3 "$base/trivy_${version}_Linux-64bit.tar.gz" -o "$directory/trivy.tar.gz"
curl --fail --location --retry 3 "$base/trivy_${version}_checksums.txt" -o "$directory/checksums.txt"
expected=$(awk -v name="trivy_${version}_Linux-64bit.tar.gz" '$2 == name { print $1 }' "$directory/checksums.txt")
test -n "$expected"
echo "$expected  $directory/trivy.tar.gz" | sha256sum --check
tar -xzf "$directory/trivy.tar.gz" -C "$directory" trivy
mkdir -p "$HOME/.local/bin"
install -m 755 "$directory/trivy" "$HOME/.local/bin/trivy"
echo "$HOME/.local/bin" >> "$GITHUB_PATH"
