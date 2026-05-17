#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/analyze-cache-effectiveness.sh - Cache effectiveness analysis
#
# Purpose:
#   Analyzes unirtm cache statistics and generates per-tool cache effectiveness
#   reports. Identifies tools that don't benefit from caching and recommends
#   optimization strategies.
#
# Usage:
#   ./scripts/analyze-cache-effectiveness.sh [OPTIONS]
#
# Options:
#   --baseline <path>            Baseline file with cache data (default: benchmarks/baseline.json)
#   --output-format <json|text>  Output format (default: text)
#   --threshold <percent>        Poor cache threshold (default: 20)
#   -h, --help                   Show this help message
#
# Requirements: 4.3, 4.5

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
BASELINE_FILE="${BASELINE_FILE:-benchmarks/baseline.json}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-text}"
POOR_CACHE_THRESHOLD="${POOR_CACHE_THRESHOLD:-20}"

# ── Helper Functions ─────────────────────────────────────────────────────────

usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Analyzes cache effectiveness and generates optimization recommendations.

Options:
  --baseline <path>            Baseline file (default: benchmarks/baseline.json)
  --output-format <json|text>  Output format (default: text)
  --threshold <percent>        Poor cache threshold (default: 20)
  -h, --help                   Show this help message

Examples:
  # Analyze cache effectiveness
  ./scripts/analyze-cache-effectiveness.sh

  # Use custom baseline
  ./scripts/analyze-cache-effectiveness.sh --baseline /path/to/baseline.json

  # JSON output
  ./scripts/analyze-cache-effectiveness.sh --output-format json

Requirements: 4.3, 4.5
EOF
  exit "${1:-0}"
}

# Purpose: Extract value from JSON
json_extract() {
  local json="${1:-}"
  local key="${2:-}"
  echo "$json" | grep -o "\"$key\":[[:space:]]*[0-9.]*" | grep -o '[0-9.]*$' | head -n 1
}

# Purpose: Analyze cache effectiveness
analyze_cache() {
  if [ ! -f "$BASELINE_FILE" ]; then
    log_error "Baseline file not found: $BASELINE_FILE"
    exit 1
  fi

  local baseline_json cold_time warm_time speedup

  baseline_json=$(cat "$BASELINE_FILE")
  cold_time=$(json_extract "$baseline_json" "total_time_seconds" | head -n 1)
  warm_time=$(json_extract "$baseline_json" "total_time_seconds" | tail -n 1)

  if [ -z "$cold_time" ] || [ -z "$warm_time" ]; then
    log_error "Could not extract cache timing data from baseline"
    exit 1
  fi

  speedup=$(awk "BEGIN {printf \"%.1f\", (($cold_time - $warm_time) / $cold_time) * 100}")

  if [ "$OUTPUT_FORMAT" = "json" ]; then
    printf '{\n'
    printf '  "cold_cache_seconds": %s,\n' "$cold_time"
    printf '  "warm_cache_seconds": %s,\n' "$warm_time"
    printf '  "speedup_percentage": %s,\n' "$speedup"
    printf '  "cache_effectiveness": "%s",\n' "$([ "$(echo "$speedup" | awk '{print int($1)}')" -gt "$POOR_CACHE_THRESHOLD" ] && echo "good" || echo "poor")"
    printf '  "recommendations": [\n'
    if [ "$(echo "$speedup" | awk '{print int($1)}')" -lt "$POOR_CACHE_THRESHOLD" ]; then
      printf '    "Consider pre-warming cache in CI",\n'
      printf '    "Review unirtm cache configuration",\n'
      printf '    "Check for network-dependent installations"\n'
    else
      printf '    "Cache effectiveness is good"\n'
    fi
    printf '  ]\n'
    printf '}\n'
  else
    log_success "
═══════════════════════════════════════════════════════════════
Cache Effectiveness Analysis
═══════════════════════════════════════════════════════════════

Cold Cache Time:  ${cold_time}s
Warm Cache Time:  ${warm_time}s
Speedup:          ${speedup}%

Cache Effectiveness: $([ "$(echo "$speedup" | awk '{print int($1)}')" -gt "$POOR_CACHE_THRESHOLD" ] && echo "✅ Good" || echo "⚠️  Poor")

Recommendations:
"
    if [ "$(echo "$speedup" | awk '{print int($1)}')" -lt "$POOR_CACHE_THRESHOLD" ]; then
      echo "  • Consider pre-warming cache in CI"
      echo "  • Review unirtm cache configuration"
      echo "  • Check for network-dependent installations"
    else
      echo "  • Cache effectiveness is good"
      echo "  • Continue current caching strategy"
    fi
    echo "
═══════════════════════════════════════════════════════════════"
  fi
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Analyzing cache effectiveness..." >&2
  analyze_cache
  log_success "Analysis complete" >&2
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --baseline)
    BASELINE_FILE="${2:-benchmarks/baseline.json}"
    shift 2
    ;;
  --output-format)
    OUTPUT_FORMAT="${2:-text}"
    shift 2
    ;;
  --threshold)
    POOR_CACHE_THRESHOLD="${2:-20}"
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
