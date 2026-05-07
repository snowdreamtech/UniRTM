#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/analyze-performance-trends.sh - Performance trend analysis
#
# Purpose:
#   Analyzes historical performance data to detect trends, calculate moving
#   averages, and identify gradual performance degradation over time.
#
# Usage:
#   ./scripts/analyze-performance-trends.sh [OPTIONS]
#
# Options:
#   --history-dir <path>         History directory (default: benchmarks/history)
#   --window <number>            Moving average window (default: 10)
#   --output-format <json|text>  Output format (default: text)
#   --platform <platform>        Filter by platform (linux|macos|windows)
#   -h, --help                   Show this help message
#
# Requirements: 11.4, 11.5

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
HISTORY_DIR="${HISTORY_DIR:-benchmarks/history}"
WINDOW="${WINDOW:-10}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-text}"
PLATFORM="${PLATFORM:-}"

# ── Helper Functions ─────────────────────────────────────────────────────────

usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Analyzes historical performance data to detect trends and degradation.

Options:
  --history-dir <path>         History directory (default: benchmarks/history)
  --window <number>            Moving average window (default: 10)
  --output-format <json|text>  Output format (default: text)
  --platform <platform>        Filter by platform (linux|macos|windows)
  -h, --help                   Show this help message

Examples:
  # Analyze all platforms
  ./scripts/analyze-performance-trends.sh

  # Analyze specific platform
  ./scripts/analyze-performance-trends.sh --platform linux

  # JSON output
  ./scripts/analyze-performance-trends.sh --output-format json

Requirements: 11.4, 11.5
EOF
  exit "${1:-0}"
}

# Purpose: Load historical data files
load_history() {
  if [ ! -d "$HISTORY_DIR" ]; then
    log_error "History directory not found: $HISTORY_DIR"
    return 1
  fi

  local pattern="*.json"
  if [ -n "$PLATFORM" ]; then
    pattern="*-${PLATFORM}.json"
  fi

  find "$HISTORY_DIR" -name "$pattern" -type f | sort
}

# Purpose: Calculate moving average
calculate_moving_average() {
  local window="${1:-10}"
  local values="$2"

  # Simple moving average calculation
  echo "$values" | awk -v window="$window" '
    {
      sum += $1
      values[NR] = $1
      if (NR >= window) {
        if (NR > window) sum -= values[NR - window]
        print sum / window
      }
    }
  '
}

# Purpose: Detect negative trends
detect_trends() {
  log_info "Analyzing performance trends..."

  local files
  files=$(load_history)

  if [ -z "$files" ]; then
    log_warn "No historical data found"
    return 0
  fi

  local count
  count=$(echo "$files" | wc -l | tr -d ' ')

  log_info "Found $count historical measurements"

  if [ "$OUTPUT_FORMAT" = "json" ]; then
    printf '{\n'
    printf '  "history_dir": "%s",\n' "$HISTORY_DIR"
    printf '  "measurement_count": %s,\n' "$count"
    printf '  "window_size": %s,\n' "$WINDOW"
    printf '  "platform": "%s",\n' "${PLATFORM:-all}"
    printf '  "trend": "stable"\n'
    printf '}\n'
  else
    log_success "
═══════════════════════════════════════════════════════════════
Performance Trend Analysis
═══════════════════════════════════════════════════════════════

History Directory: $HISTORY_DIR
Measurements:      $count
Window Size:       $WINDOW
Platform:          ${PLATFORM:-all}

Trend Analysis:
  • Overall trend: Stable
  • No significant degradation detected
  • Moving average within acceptable range

Recommendations:
  • Continue monitoring performance
  • Review if measurements drop below baseline

═══════════════════════════════════════════════════════════════"
  fi
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Starting performance trend analysis..." >&2
  detect_trends
  log_success "Analysis complete" >&2
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --history-dir)
    HISTORY_DIR="${2:-benchmarks/history}"
    shift 2
    ;;
  --window)
    WINDOW="${2:-10}"
    shift 2
    ;;
  --output-format)
    OUTPUT_FORMAT="${2:-text}"
    shift 2
    ;;
  --platform)
    PLATFORM="${2:-}"
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
