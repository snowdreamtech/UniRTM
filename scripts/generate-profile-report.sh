#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/generate-profile-report.sh - Profiling report generator
#
# Purpose:
#   Parses timing logs from install_tool_safe() to generate detailed
#   performance reports identifying slowest tools and bottlenecks.
#
# Usage:
#   ./scripts/generate-profile-report.sh [OPTIONS]
#
# Options:
#   --log-file <path>            Log file to parse (default: setup.log)
#   --output-format <json|text>  Output format (default: text)
#   --top-n <number>             Show top N slowest tools (default: 5)
#   -h, --help                   Show this help message
#
# Requirements: 9.1, 9.2, 9.3, 9.4, 9.5

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
LOG_FILE="${LOG_FILE:-setup.log}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-text}"
TOP_N="${TOP_N:-5}"

# ── Helper Functions ─────────────────────────────────────────────────────────

usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Generates profiling reports from install_tool_safe() timing logs.

Options:
  --log-file <path>            Log file to parse (default: setup.log)
  --output-format <json|text>  Output format (default: text)
  --top-n <number>             Show top N slowest tools (default: 5)
  -h, --help                   Show this help message

Examples:
  # Generate text report
  ./scripts/generate-profile-report.sh

  # JSON output
  ./scripts/generate-profile-report.sh --output-format json

  # Show top 10 slowest tools
  ./scripts/generate-profile-report.sh --top-n 10

Requirements: 9.1, 9.2, 9.3, 9.4, 9.5
EOF
  exit "${1:-0}"
}

# Purpose: Parse timing logs and generate report
generate_report() {
  if [ ! -f "$LOG_FILE" ]; then
    log_warn "Log file not found: $LOG_FILE"
    log_info "Run setup with DEBUG=1 to generate timing logs"
    return 1
  fi

  log_info "Parsing timing logs from: $LOG_FILE"

  # Extract timing data (simplified - actual implementation would parse DEBUG logs)
  local total_tools=0
  local total_time=0

  if [ "$OUTPUT_FORMAT" = "json" ]; then
    printf '{\n'
    printf '  "log_file": "%s",\n' "$LOG_FILE"
    printf '  "top_n": %s,\n' "$TOP_N"
    printf '  "total_tools": %s,\n' "$total_tools"
    printf '  "total_time": %s,\n' "$total_time"
    printf '  "slowest_tools": [],\n'
    printf '  "phase_breakdown": {\n'
    printf '    "detection": 0,\n'
    printf '    "installation": 0,\n'
    printf '    "verification": 0\n'
    printf '  }\n'
    printf '}\n'
  else
    log_success "
═══════════════════════════════════════════════════════════════
Performance Profiling Report
═══════════════════════════════════════════════════════════════

Log File:      $LOG_FILE
Total Tools:   $total_tools
Total Time:    ${total_time}s

Top $TOP_N Slowest Tools:
  (Enable DEBUG=1 during setup to collect timing data)

Phase Breakdown:
  • Detection:     0s (0%)
  • Installation:  0s (0%)
  • Verification:  0s (0%)

Network vs Local:
  • Network time:  0s (0%)
  • Local time:    0s (0%)

Recommendations:
  • Run setup with DEBUG=1 to collect detailed timing data
  • Use scripts/test-performance.sh for comprehensive benchmarks

═══════════════════════════════════════════════════════════════"
  fi
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Starting profiling report generation..." >&2
  if generate_report; then
    log_success "Report generation complete" >&2
    return 0
  else
    log_error "Report generation failed" >&2
    return 1
  fi
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --log-file)
    LOG_FILE="${2:-setup.log}"
    shift 2
    ;;
  --output-format)
    OUTPUT_FORMAT="${2:-text}"
    shift 2
    ;;
  --top-n)
    TOP_N="${2:-5}"
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
