#!/bin/sh
set -e
REPO_ROOT="$(git rev-parse --show-toplevel)"
cp "$REPO_ROOT/scripts/commit-msg" "$REPO_ROOT/.git/hooks/commit-msg"
chmod +x "$REPO_ROOT/.git/hooks/commit-msg"
echo "Git hooks installed."
