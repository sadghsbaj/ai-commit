#!/usr/bin/env bash
set -euo pipefail

# Determine target base directory (default to ~/dotfiles/bin)
TARGET_BASE="${DOTFILES_BIN_DIR:-$HOME/dotfiles/bin}"

echo "==> Running tests..."
go test ./...

echo "==> Building x86_64 (amd64) binary..."
mkdir -p "$TARGET_BASE/x86_64"
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$TARGET_BASE/x86_64/ai-commit" .
chmod +x "$TARGET_BASE/x86_64/ai-commit"
echo "    -> $TARGET_BASE/x86_64/ai-commit"

echo "==> Building aarch64 (arm64) binary..."
mkdir -p "$TARGET_BASE/aarch64"
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "$TARGET_BASE/aarch64/ai-commit" .
chmod +x "$TARGET_BASE/aarch64/ai-commit"
echo "    -> $TARGET_BASE/aarch64/ai-commit"

echo "==> Build complete!"
