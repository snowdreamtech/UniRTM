#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

#
# smoke-test-gen-dependabot.sh
# Purpose: Verifies that gen-dependabot.sh correctly detects ecosystems and generates valid YAML.

set -e

SCRIPT_DIR=$(cd "$(dirname "$0")/.." && pwd)
TEMP_DIR=$(mktemp -d)

echo "🧪 Starting Dependabot Generator Smoke Test..."
echo "📂 Temp directory: $TEMP_DIR"

cleanup() {
  rm -rf "$TEMP_DIR"
  echo "🧹 Cleanup complete."
}
trap cleanup EXIT

# 1. Setup Mock Workspace
mkdir -p "$TEMP_DIR/root"
mkdir -p "$TEMP_DIR/root/docs"
mkdir -p "$TEMP_DIR/root/.devcontainer"
mkdir -p "$TEMP_DIR/root/docker"
mkdir -p "$TEMP_DIR/root/.github/workflows"

cd "$TEMP_DIR/root"
git init -q
git config user.email "test@example.com"
git config user.name "Test User"

# Create dummy files
touch "package.json"
touch "docs/package.json"
# Create devcontainer.json to trigger devcontainers ecosystem
echo '{"name": "test"}' >".devcontainer/devcontainer.json"
touch ".devcontainer/docker-compose.yml" # Should be ignored by docker ecosystem
touch "docker/Dockerfile"                # Should be detected by docker ecosystem
touch ".pre-commit-config.yaml"
touch ".github/workflows/ci.yml"

# MUST track files for git ls-files to work
git add .
git commit -m "initial" -q

# Run the script and capture output (using --dry-run to get STDOUT)
echo "🏃 Running generator..."
OUTPUT_FILE="$TEMP_DIR/dependabot.yml"
unirtm generate dependabot --dry-run >"$OUTPUT_FILE"

# Debug: Show generated file if test fails
on_failure() {
  echo "DEBUG: Current Directory: $(pwd)"
  echo "DEBUG: Git status:"
  git status
  echo "DEBUG: git ls-files outcome:"
  git ls-files
  echo "DEBUG: Generated dependabot.yml contents:"
  cat "$OUTPUT_FILE"
}

# 3. Assertions
assert_contains() {
  _pattern="$1"
  if grep -q "$_pattern" "$OUTPUT_FILE"; then
    echo "✅ Found: $_pattern"
  else
    echo "❌ Missing: $_pattern"
    on_failure
    exit 1
  fi
}

assert_not_contains() {
  _pattern="$1"
  if grep -q "$_pattern" "$OUTPUT_FILE"; then
    echo "❌ Should NOT contain: $_pattern"
    on_failure
    exit 1
  else
    echo "✅ Correctly excluded: $_pattern"
  fi
}

echo "🔍 Verifying generated content..."
assert_contains 'package-ecosystem: "npm"'
assert_contains 'directory: "/"'
assert_contains 'directory: "/docs"'
# Docker ecosystem should detect docker/Dockerfile but not .devcontainer
assert_contains 'package-ecosystem: "docker"'
assert_contains 'directory: "/docker"'
# Devcontainers ecosystem should be detected with directory pointing to root
assert_contains 'package-ecosystem: "devcontainers"'
assert_contains 'package-ecosystem: "pre-commit"'
# New unified grouping strategy
assert_contains "all-dependencies"
assert_contains 'patterns: \["\*"\]'
# Verify all update types are included
assert_contains '"patch"'
assert_contains '"minor"'
assert_contains '"major"'
assert_contains 'rebase-strategy: "auto"'
assert_contains 'open-pull-requests-limit: 5'
assert_contains 'update-types:'
assert_contains 'cooldown:'
assert_contains 'default-days: 7'

echo "🔍 Verifying exclusions..."
# Verify .devcontainer is NOT detected by docker ecosystem
# (it should only be handled by devcontainers ecosystem)
assert_not_contains 'directory: "/.devcontainer"'

echo "✨ Smoke test PASSED successfully!"
