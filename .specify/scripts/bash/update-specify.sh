#!/usr/bin/env bash
SCRIPT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(CDPATH="" cd "$SCRIPT_DIR/../../.." && pwd)"

cd "$REPO_ROOT" || exit 1
specify init . --integration generic --integration-options="--commands-dir .specify/commands/"
