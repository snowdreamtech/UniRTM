#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/generate-tool-docs.sh - Automatic tool documentation generator
#
# Purpose:
#   Extracts tool metadata from scripts/lib/langs/*.sh and generates
#   documentation tables automatically to keep docs in sync with code.
#
# Usage:
#   ./scripts/generate-tool-docs.sh [OPTIONS]
#
# Options:
#   --output <path>              Output file (default: docs/tools.md)
#   --format <markdown|json>     Output format (default: markdown)
#   --update-in-place            Update existing documentation
#   -h, --help                   Show this help message
#
# Requirements: 10.2, 10.3, 10.5

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
OUTPUT_FILE="${OUTPUT_FILE:-docs/tools.md}"
FORMAT="${FORMAT:-markdown}"
UPDATE_IN_PLACE="${UPDATE_IN_PLACE:-0}"

# ── Helper Functions ─────────────────────────────────────────────────────────

usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Generates tool documentation from source code metadata.

Options:
  --output <path>              Output file (default: docs/tools.md)
  --format <markdown|json>     Output format (default: markdown)
  --update-in-place            Update existing documentation
  -h, --help                   Show this help message

Examples:
  # Generate tool documentation
  ./scripts/generate-tool-docs.sh

  # JSON output
  ./scripts/generate-tool-docs.sh --format json

  # Update existing docs
  ./scripts/generate-tool-docs.sh --update-in-place

Requirements: 10.2, 10.3, 10.5
EOF
  exit "${1:-0}"
}

# Purpose: Extract tool list from lang files
extract_tools() {
  if [ ! -d "scripts/lib/langs" ]; then
    log_error "scripts/lib/langs directory not found"
    return 1
  fi

  # Extract tools using install_tool_safe
  grep -r "install_tool_safe" scripts/lib/langs/*.sh 2>/dev/null |
    sed -n 's/.*install_tool_safe[[:space:]]*"\([^"]*\)".*/\1/p' |
    sort -u
}

# Purpose: Generate documentation
generate_docs() {
  log_info "Extracting tool metadata..."

  local tools
  tools=$(extract_tools)

  if [ -z "$tools" ]; then
    log_warn "No tools found"
    return 1
  fi

  local tool_count
  tool_count=$(echo "$tools" | wc -l | tr -d ' ')

  if [ "$FORMAT" = "json" ]; then
    printf '{\n'
    printf '  "generated_at": "%s",\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    printf '  "tool_count": %s,\n' "$tool_count"
    printf '  "tools": [\n'

    local first=true
    for tool in $tools; do
      [ "$first" = false ] && printf ',\n'
      first=false
      printf '    {\n'
      printf '      "name": "%s",\n' "$tool"
      printf '      "status": "installed",\n'
      printf '      "method": "install_tool_safe"\n'
      printf '    }'
    done

    printf '\n  ]\n'
    printf '}\n'
  else
    # Markdown format
    cat <<-EOF
# Tool Documentation

> Auto-generated on $(date -u +%Y-%m-%d)

## Overview

This project uses ${tool_count} tools managed via the \`install_tool_safe()\` pattern.

## Tool List

| Tool | Status | Installation Method |
|------|--------|---------------------|
EOF

    for tool in $tools; do
      printf "| %s | ✅ Installed | install_tool_safe |\n" "$tool"
    done

    cat <<-EOF

## Installation Architecture

All tools are installed using the \`install_tool_safe()\` function which provides:

- Binary-first detection strategy
- Platform-specific binary name handling
- Atomic installation with verification
- Comprehensive error handling

For more details, see [Tool Installation Architecture](reference/tool-installation.md).

---

*This documentation is automatically generated. Do not edit manually.*
*Run \`./scripts/generate-tool-docs.sh\` to regenerate.*
EOF
  fi
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Starting tool documentation generation..." >&2

  if [ "$UPDATE_IN_PLACE" = "1" ] && [ -f "$OUTPUT_FILE" ]; then
    log_info "Updating existing documentation: $OUTPUT_FILE" >&2
  fi

  if [ "$FORMAT" = "markdown" ]; then
    # Ensure output directory exists
    mkdir -p "$(dirname "$OUTPUT_FILE")"

    if generate_docs >"$OUTPUT_FILE"; then
      log_success "Documentation generated: $OUTPUT_FILE" >&2
      return 0
    else
      log_error "Documentation generation failed" >&2
      return 1
    fi
  else
    # JSON output to stdout
    generate_docs
    log_success "Documentation generation complete" >&2
    return 0
  fi
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --output)
    OUTPUT_FILE="${2:-docs/tools.md}"
    shift 2
    ;;
  --format)
    FORMAT="${2:-markdown}"
    shift 2
    ;;
  --update-in-place)
    UPDATE_IN_PLACE=1
    shift
    ;;
  -h | --help)
    usage 0
    ;;
  *)
    log_error "Unknown option: $1"
    usage 1
    ;;
  esac
done

# ── Entry Point ──────────────────────────────────────────────────────────────

main "$@"
