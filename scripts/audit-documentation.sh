#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/audit-documentation.sh - Documentation completeness audit
#
# Purpose:
#   Audits documentation completeness by checking that all tools using
#   install_tool_safe() have corresponding documentation.
#
# Usage:
#   ./scripts/audit-documentation.sh [OPTIONS]
#
# Options:
#   --threshold <percent>        Minimum coverage threshold (default: 80)
#   --output-format <json|text>  Output format (default: text)
#   -h, --help                   Show this help message
#
# Requirements: 6.1, 6.2, 6.3, 6.4

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
THRESHOLD="${THRESHOLD:-80}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-text}"

# ── Helper Functions ─────────────────────────────────────────────────────────

usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Audits documentation completeness for tools using install_tool_safe().

Options:
  --threshold <percent>        Minimum coverage (default: 80)
  --output-format <json|text>  Output format (default: text)
  -h, --help                   Show this help message

Examples:
  # Run documentation audit
  ./scripts/audit-documentation.sh

  # Custom threshold
  ./scripts/audit-documentation.sh --threshold 90

  # JSON output
  ./scripts/audit-documentation.sh --output-format json

Requirements: 6.1, 6.2, 6.3, 6.4
EOF
  exit "${1:-0}"
}

# Purpose: Extract tools using install_tool_safe
extract_tools() {
  if [ ! -d "scripts/lib/langs" ]; then
    log_error "scripts/lib/langs directory not found"
    return 1
  fi

  grep -r "install_tool_safe" scripts/lib/langs/*.sh 2>/dev/null |
    sed -n 's/.*install_tool_safe[[:space:]]*"\([^"]*\)".*/\1/p' |
    sort -u
}

# Purpose: Check if tool is documented
is_documented() {
  local tool="${1:-}"

  if [ ! -d "docs" ]; then
    return 1
  fi

  grep -r "$tool" docs/*.md docs/**/*.md 2>/dev/null >/dev/null
}

# Purpose: Audit documentation
audit_docs() {
  log_info "Extracting tools using install_tool_safe()..."

  local tools
  tools=$(extract_tools)

  if [ -z "$tools" ]; then
    log_warn "No tools found using install_tool_safe()"
    return 0
  fi

  local total=0 documented=0 missing=""

  for tool in $tools; do
    total=$((total + 1))
    if is_documented "$tool"; then
      documented=$((documented + 1))
    else
      missing="$missing $tool"
    fi
  done

  local coverage=0
  if [ "$total" -gt 0 ]; then
    coverage=$((100 * documented / total))
  fi

  if [ "$OUTPUT_FORMAT" = "json" ]; then
    printf '{\n'
    printf '  "total_tools": %s,\n' "$total"
    printf '  "documented_tools": %s,\n' "$documented"
    printf '  "coverage_percent": %s,\n' "$coverage"
    printf '  "threshold_percent": %s,\n' "$THRESHOLD"
    printf '  "passed": %s,\n' "$([ "$coverage" -ge "$THRESHOLD" ] && echo "true" || echo "false")"
    printf '  "missing_tools": ['
    first=true
    for tool in $missing; do
      [ "$first" = false ] && printf ','
      first=false
      printf '"%s"' "$tool"
    done
    printf ']\n'
    printf '}\n'
  else
    log_success "
═══════════════════════════════════════════════════════════════
Documentation Audit Report
═══════════════════════════════════════════════════════════════

Total Tools:       $total
Documented:        $documented
Coverage:          ${coverage}%
Threshold:         ${THRESHOLD}%

Status: $([ "$coverage" -ge "$THRESHOLD" ] && echo "✅ PASSED" || echo "❌ FAILED")
"

    if [ -n "$missing" ]; then
      echo "Missing Documentation:"
      for tool in $missing; do
        echo "  • $tool"
      done
    fi

    echo "
═══════════════════════════════════════════════════════════════"
  fi

  if [ "$coverage" -lt "$THRESHOLD" ]; then
    log_error "Documentation coverage (${coverage}%) below threshold (${THRESHOLD}%)"
    return 1
  fi

  return 0
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Starting documentation audit..." >&2
  if audit_docs; then
    log_success "Documentation audit passed" >&2
    return 0
  else
    log_error "Documentation audit failed" >&2
    return 1
  fi
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --threshold)
    THRESHOLD="${2:-80}"
    shift 2
    ;;
  --output-format)
    OUTPUT_FORMAT="${2:-text}"
    shift 2
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
