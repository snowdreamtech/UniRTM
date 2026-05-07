#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/compare-performance.sh - Performance comparison script
#
# Purpose:
#   Compares current performance measurements against baseline and detects
#   regressions. Generates human-readable reports with category breakdowns,
#   slowest tools, and regression status.
#
# Usage:
#   ./scripts/compare-performance.sh [OPTIONS]
#
# Options:
#   --baseline <path>              Baseline file (default: benchmarks/baseline.json)
#   --current <path>               Current measurements (default: stdin or latest run)
#   --threshold-warning <percent>  Warning threshold (default: 20)
#   --threshold-error <percent>    Error threshold (default: 50)
#   --output-format <text|markdown|json>
#   -h, --help                     Show this help message
#
# Requirements: 2.1, 2.2, 2.3, 2.4

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
BASELINE_FILE="${BASELINE_FILE:-benchmarks/baseline.json}"
CURRENT_FILE="${CURRENT_FILE:-}"
THRESHOLD_WARNING="${THRESHOLD_WARNING:-20}"
THRESHOLD_ERROR="${THRESHOLD_ERROR:-50}"
OUTPUT_FORMAT="${OUTPUT_FORMAT:-text}"

# ── Global State ─────────────────────────────────────────────────────────────
EXIT_CODE=0
REGRESSION_COUNT=0
WARNING_COUNT=0

# ── Helper Functions ─────────────────────────────────────────────────────────

# Purpose: Display usage information
usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Compares current performance measurements against baseline and detects regressions.

Options:
  --baseline <path>              Baseline file (default: benchmarks/baseline.json)
  --current <path>               Current measurements (default: stdin or latest run)
  --threshold-warning <percent>  Warning threshold (default: 20)
  --threshold-error <percent>    Error threshold (default: 50)
  --output-format <text|markdown|json>  Output format (default: text)
  -h, --help                     Show this help message

Examples:
  # Compare against baseline using stdin
  ./scripts/test-performance.sh --output-format json | \\
    ./scripts/compare-performance.sh

  # Compare using saved measurement file
  ./scripts/compare-performance.sh --current measurements.json

  # Use custom thresholds
  ./scripts/compare-performance.sh --threshold-warning 15 --threshold-error 40

Exit Codes:
  0 - No regressions or warnings only
  1 - Error threshold exceeded (regression > 50%)

Requirements: 2.1, 2.2, 2.3, 2.4
EOF
  exit "${1:-0}"
}

# Purpose: Extract value from JSON using grep/sed (POSIX-compatible)
# Params:
#   $1 - JSON string
#   $2 - Key to extract
# Returns: Value (stdout)
json_extract() {
  local json="${1:-}"
  local key="${2:-}"

  echo "$json" | grep -o "\"$key\":[[:space:]]*[0-9.]*" | grep -o '[0-9.]*$' | head -n 1
}

# Purpose: Extract string value from JSON
# Params:
#   $1 - JSON string
#   $2 - Key to extract
# Returns: String value (stdout)
json_extract_string() {
  local json="${1:-}"
  local key="${2:-}"

  echo "$json" | grep -o "\"$key\":[[:space:]]*\"[^\"]*\"" | sed 's/.*"\([^"]*\)"/\1/' | head -n 1
}

# Purpose: Load JSON file
# Params:
#   $1 - File path
# Returns: JSON content (stdout)
load_json() {
  local file="${1:-}"

  if [ ! -f "$file" ]; then
    log_error "File not found: $file"
    return 1
  fi

  cat "$file"
}

# Purpose: Calculate percentage difference
# Params:
#   $1 - Baseline value
#   $2 - Current value
# Returns: Percentage difference (stdout)
calculate_diff_percent() {
  local baseline="${1:-0}"
  local current="${2:-0}"

  # Avoid division by zero
  if [ "$baseline" = "0" ] || [ -z "$baseline" ]; then
    echo "0"
    return 0
  fi

  # Calculate: ((current - baseline) / baseline) * 100
  awk "BEGIN {printf \"%.1f\", (($current - $baseline) / $baseline) * 100}"
}

# Purpose: Determine status based on diff percentage
# Params:
#   $1 - Diff percentage
# Returns: Status string (stdout)
# Side effects: Updates REGRESSION_COUNT, WARNING_COUNT, EXIT_CODE
get_status() {
  local diff_percent="${1:-0}"

  # Convert to absolute value for comparison
  local diff_abs
  diff_abs=$(echo "$diff_percent" | awk '{if ($1 < 0) print -$1; else print $1}')

  # Convert to integer for comparison (remove decimal)
  local diff_int
  diff_int=$(echo "$diff_abs" | awk '{printf "%d", $1}')

  # Check if it's a positive regression (slower)
  local is_regression
  is_regression=$(echo "$diff_percent" | awk '{if ($1 > 0) print 1; else print 0}')

  if [ "$is_regression" = "1" ] && [ "$diff_int" -gt "$THRESHOLD_ERROR" ]; then
    echo "REGRESSION (FAIL)"
    REGRESSION_COUNT=$((REGRESSION_COUNT + 1))
    EXIT_CODE=1
  elif [ "$is_regression" = "1" ] && [ "$diff_int" -gt "$THRESHOLD_WARNING" ]; then
    echo "REGRESSION (WARNING)"
    WARNING_COUNT=$((WARNING_COUNT + 1))
  elif [ "$is_regression" = "0" ] && [ "$diff_int" -gt 10 ]; then
    echo "IMPROVED"
  else
    echo "OK"
  fi
}

# Purpose: Get status emoji
# Params:
#   $1 - Status string
# Returns: Emoji (stdout)
get_status_emoji() {
  local status="${1:-}"

  case "$status" in
  "REGRESSION (FAIL)") echo "❌" ;;
  "REGRESSION (WARNING)") echo "⚠️" ;;
  "IMPROVED") echo "✅" ;;
  "OK") echo "✅" ;;
  *) echo "❓" ;;
  esac
}

# Purpose: Format diff percentage with sign
# Params:
#   $1 - Diff percentage
# Returns: Formatted string (stdout)
format_diff() {
  local diff="${1:-0}"
  local diff_int
  diff_int=$(echo "$diff" | awk '{printf "%d", $1}')

  if [ "$diff_int" -gt 0 ]; then
    echo "+${diff}%"
  else
    echo "${diff}%"
  fi
}

# Purpose: Compare category performance
# Params:
#   $1 - Baseline JSON
#   $2 - Current JSON
#   $3 - Category name
# Side effects: Updates REGRESSION_COUNT, WARNING_COUNT, EXIT_CODE
compare_category() {
  local baseline_json="${1:-}"
  local current_json="${2:-}"
  local category="${3:-}"

  # Extract category data
  local baseline_cat current_cat
  baseline_cat=$(echo "$baseline_json" | grep -A 10 "\"$category\":" | head -n 10)
  current_cat=$(echo "$current_json" | grep -A 10 "\"$category\":" | head -n 10)

  # Extract time values
  local baseline_time current_time
  baseline_time=$(json_extract "$baseline_cat" "time_seconds")
  current_time=$(json_extract "$current_cat" "time_seconds")

  # Handle missing data
  if [ -z "$baseline_time" ] || [ -z "$current_time" ]; then
    return 0
  fi

  # Calculate diff
  local diff_percent
  diff_percent=$(calculate_diff_percent "$baseline_time" "$current_time")

  # Determine status and update counters
  local status
  local diff_abs diff_int is_regression

  diff_abs=$(echo "$diff_percent" | awk '{if ($1 < 0) print -$1; else print $1}')
  diff_int=$(echo "$diff_abs" | awk '{printf "%d", $1}')
  is_regression=$(echo "$diff_percent" | awk '{if ($1 > 0) print 1; else print 0}')

  if [ "$is_regression" = "1" ] && [ "$diff_int" -gt "$THRESHOLD_ERROR" ]; then
    status="REGRESSION (FAIL)"
    REGRESSION_COUNT=$((REGRESSION_COUNT + 1))
    EXIT_CODE=1
  elif [ "$is_regression" = "1" ] && [ "$diff_int" -gt "$THRESHOLD_WARNING" ]; then
    status="REGRESSION (WARNING)"
    WARNING_COUNT=$((WARNING_COUNT + 1))
  elif [ "$is_regression" = "0" ] && [ "$diff_int" -gt 10 ]; then
    status="IMPROVED"
  else
    status="OK"
  fi

  # Store result for output
  echo "$category|$baseline_time|$current_time|$diff_percent|$status"
}

# Purpose: Output comparison in text format
# Params:
#   $1 - Baseline JSON
#   $2 - Current JSON
output_text() {
  local baseline_json="${1:-}"
  local current_json="${2:-}"

  # Extract metadata
  local baseline_time current_time timestamp commit_sha
  baseline_time=$(json_extract "$baseline_json" "total_time_seconds")
  current_time=$(json_extract "$current_json" "total_time_seconds")
  timestamp=$(json_extract_string "$current_json" "timestamp")
  commit_sha=$(json_extract_string "$current_json" "commit_sha")

  # Calculate total diff
  local total_diff
  total_diff=$(calculate_diff_percent "$baseline_time" "$current_time")

  # Determine total status (inline to avoid subshell)
  local total_status total_diff_int
  total_diff_int=$(echo "$total_diff" | awk '{if ($1 < 0) printf "%d", -$1; else printf "%d", $1}')
  if echo "$total_diff" | awk '{exit ($1 > 0) ? 0 : 1}' && [ "$total_diff_int" -gt "$THRESHOLD_ERROR" ]; then
    total_status="REGRESSION (FAIL)"
  elif echo "$total_diff" | awk '{exit ($1 > 0) ? 0 : 1}' && [ "$total_diff_int" -gt "$THRESHOLD_WARNING" ]; then
    total_status="REGRESSION (WARNING)"
  elif echo "$total_diff" | awk '{exit ($1 < 0) ? 0 : 1}' && [ "$total_diff_int" -gt 10 ]; then
    total_status="IMPROVED"
  else
    total_status="OK"
  fi

  log_success "
═══════════════════════════════════════════════════════════════
Performance Comparison Report
═══════════════════════════════════════════════════════════════

Summary:
  Timestamp:   $timestamp
  Commit:      $commit_sha
  Baseline:    ${baseline_time}s
  Current:     ${current_time}s
  Difference:  $(format_diff "$total_diff")
  Status:      $total_status $(get_status_emoji "$total_status")

Category Breakdown:
"

  # Compare each category (inline to preserve global counters)
  for category in security linters formatters runtimes; do
    # Extract category data
    local baseline_cat current_cat
    baseline_cat=$(echo "$baseline_json" | grep -A 10 "\"$category\":" | head -n 10)
    current_cat=$(echo "$current_json" | grep -A 10 "\"$category\":" | head -n 10)

    # Extract time values
    local baseline_val current_val
    baseline_val=$(json_extract "$baseline_cat" "time_seconds")
    current_val=$(json_extract "$current_cat" "time_seconds")

    # Skip if missing data
    if [ -z "$baseline_val" ] || [ -z "$current_val" ]; then
      continue
    fi

    # Calculate diff
    local diff_val
    diff_val=$(calculate_diff_percent "$baseline_val" "$current_val")

    # Determine status and update counters (inline)
    local status_val diff_abs diff_int
    diff_abs=$(echo "$diff_val" | awk '{if ($1 < 0) print -$1; else print $1}')
    diff_int=$(echo "$diff_abs" | awk '{printf "%d", $1}')

    if echo "$diff_val" | awk '{exit ($1 > 0) ? 0 : 1}' && [ "$diff_int" -gt "$THRESHOLD_ERROR" ]; then
      status_val="REGRESSION (FAIL)"
      REGRESSION_COUNT=$((REGRESSION_COUNT + 1))
      EXIT_CODE=1
    elif echo "$diff_val" | awk '{exit ($1 > 0) ? 0 : 1}' && [ "$diff_int" -gt "$THRESHOLD_WARNING" ]; then
      status_val="REGRESSION (WARNING)"
      WARNING_COUNT=$((WARNING_COUNT + 1))
    elif echo "$diff_val" | awk '{exit ($1 < 0) ? 0 : 1}' && [ "$diff_int" -gt 10 ]; then
      status_val="IMPROVED"
    else
      status_val="OK"
    fi

    local emoji
    emoji=$(get_status_emoji "$status_val")

    printf "  %-15s %6ss → %6ss  %8s  %s %s\n" \
      "$category:" "$baseline_val" "$current_val" "$(format_diff "$diff_val")" "$emoji" "$status_val"
  done

  log_success "
═══════════════════════════════════════════════════════════════

Regression Summary:
  Errors:   $REGRESSION_COUNT
  Warnings: $WARNING_COUNT

Thresholds:
  Warning:  > ${THRESHOLD_WARNING}%
  Error:    > ${THRESHOLD_ERROR}%
"

  if [ "$REGRESSION_COUNT" -gt 0 ]; then
    log_error "Performance regression detected! ($REGRESSION_COUNT categories exceeded error threshold)"
  elif [ "$WARNING_COUNT" -gt 0 ]; then
    log_warn "Performance warnings detected ($WARNING_COUNT categories exceeded warning threshold)"
  else
    log_success "All performance checks passed ✅"
  fi

  echo "
═══════════════════════════════════════════════════════════════"
}

# Purpose: Output comparison in markdown format
# Params:
#   $1 - Baseline JSON
#   $2 - Current JSON
output_markdown() {
  local baseline_json="${1:-}"
  local current_json="${2:-}"

  # Extract metadata
  local baseline_time current_time timestamp commit_sha platform
  baseline_time=$(json_extract "$baseline_json" "total_time_seconds")
  current_time=$(json_extract "$current_json" "total_time_seconds")
  timestamp=$(json_extract_string "$current_json" "timestamp")
  commit_sha=$(json_extract_string "$current_json" "commit_sha")
  platform=$(json_extract_string "$current_json" "os")

  # Calculate total diff
  local total_diff
  total_diff=$(calculate_diff_percent "$baseline_time" "$current_time")
  local total_status
  total_status=$(get_status "$total_diff")
  local total_emoji
  total_emoji=$(get_status_emoji "$total_status")

  # Output markdown
  cat <<-EOF
## Performance Comparison Report

### Summary
- **Total Time**: ${current_time}s (baseline: ${baseline_time}s) $total_emoji **$(format_diff "$total_diff")**
- **Platform**: $platform
- **Commit**: \`$commit_sha\`
- **Timestamp**: $timestamp

### Category Breakdown

| Category | Baseline | Current | Diff | Status |
|----------|----------|---------|------|--------|
EOF

  # Compare each category
  for category in security linters formatters runtimes; do
    result=$(compare_category "$baseline_json" "$current_json" "$category")

    if [ -n "$result" ]; then
      cat_name=$(echo "$result" | cut -d'|' -f1)
      baseline_val=$(echo "$result" | cut -d'|' -f2)
      current_val=$(echo "$result" | cut -d'|' -f3)
      diff_val=$(echo "$result" | cut -d'|' -f4)
      status_val=$(echo "$result" | cut -d'|' -f5)
      emoji=$(get_status_emoji "$status_val")

      printf "| %s | %ss | %ss | %s | %s %s |\n" \
        "$cat_name" "$baseline_val" "$current_val" "$(format_diff "$diff_val")" "$emoji" "$status_val"
    fi
  done

  cat <<-EOF

### Regression Summary
- **Errors**: $REGRESSION_COUNT (threshold: > ${THRESHOLD_ERROR}%)
- **Warnings**: $WARNING_COUNT (threshold: > ${THRESHOLD_WARNING}%)

EOF

  if [ "$REGRESSION_COUNT" -gt 0 ]; then
    echo "⚠️ **Performance regression detected!** $REGRESSION_COUNT categories exceeded error threshold."
  elif [ "$WARNING_COUNT" -gt 0 ]; then
    echo "⚠️ **Performance warnings detected.** $WARNING_COUNT categories exceeded warning threshold."
  else
    echo "✅ **All performance checks passed.**"
  fi
}

# Purpose: Output comparison in JSON format
# Params:
#   $1 - Baseline JSON
#   $2 - Current JSON
output_json() {
  local baseline_json="${1:-}"
  local current_json="${2:-}"

  # Extract metadata
  local baseline_time current_time
  baseline_time=$(json_extract "$baseline_json" "total_time_seconds")
  current_time=$(json_extract "$current_json" "total_time_seconds")

  # Calculate total diff
  local total_diff
  total_diff=$(calculate_diff_percent "$baseline_time" "$current_time")

  printf '{\n'
  printf '  "baseline_time_seconds": %s,\n' "$baseline_time"
  printf '  "current_time_seconds": %s,\n' "$current_time"
  printf '  "total_diff_percent": %s,\n' "$total_diff"
  printf '  "thresholds": {\n'
  printf '    "warning": %s,\n' "$THRESHOLD_WARNING"
  printf '    "error": %s\n' "$THRESHOLD_ERROR"
  printf '  },\n'
  printf '  "categories": {\n'

  # Compare each category
  first=true
  for category in security linters formatters runtimes; do
    result=$(compare_category "$baseline_json" "$current_json" "$category")

    if [ -n "$result" ]; then
      if [ "$first" = false ]; then
        printf ',\n'
      fi
      first=false

      baseline_val=$(echo "$result" | cut -d'|' -f2)
      current_val=$(echo "$result" | cut -d'|' -f3)
      diff_val=$(echo "$result" | cut -d'|' -f4)
      status_val=$(echo "$result" | cut -d'|' -f5)

      printf '    "%s": {\n' "$category"
      printf '      "baseline_seconds": %s,\n' "$baseline_val"
      printf '      "current_seconds": %s,\n' "$current_val"
      printf '      "diff_percent": %s,\n' "$diff_val"
      printf '      "status": "%s"\n' "$status_val"
      printf '    }'
    fi
  done

  printf '\n  },\n'
  printf '  "summary": {\n'
  printf '    "regression_count": %s,\n' "$REGRESSION_COUNT"
  printf '    "warning_count": %s,\n' "$WARNING_COUNT"
  printf '    "exit_code": %s\n' "$EXIT_CODE"
  printf '  }\n'
  printf '}\n'
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Starting performance comparison..." >&2

  # Load baseline
  if [ ! -f "$BASELINE_FILE" ]; then
    log_error "Baseline file not found: $BASELINE_FILE"
    log_info "Run scripts/collect-baseline.sh to create a baseline"
    exit 1
  fi

  local baseline_json
  baseline_json=$(load_json "$BASELINE_FILE")
  log_info "Loaded baseline: $BASELINE_FILE" >&2

  # Load current measurements
  local current_json
  if [ -n "$CURRENT_FILE" ]; then
    if [ ! -f "$CURRENT_FILE" ]; then
      log_error "Current file not found: $CURRENT_FILE"
      exit 1
    fi
    current_json=$(load_json "$CURRENT_FILE")
    log_info "Loaded current: $CURRENT_FILE" >&2
  else
    # Read from stdin
    log_info "Reading current measurements from stdin..." >&2
    current_json=$(cat)
  fi

  # Output comparison
  case "$OUTPUT_FORMAT" in
  text)
    output_text "$baseline_json" "$current_json"
    ;;
  markdown)
    output_markdown "$baseline_json" "$current_json"
    ;;
  json)
    output_json "$baseline_json" "$current_json"
    ;;
  *)
    log_error "Unknown output format: $OUTPUT_FORMAT"
    usage 1
    ;;
  esac

  exit "$EXIT_CODE"
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --baseline)
    BASELINE_FILE="${2:-benchmarks/baseline.json}"
    shift 2
    ;;
  --current)
    CURRENT_FILE="${2:-}"
    shift 2
    ;;
  --threshold-warning)
    THRESHOLD_WARNING="${2:-20}"
    shift 2
    ;;
  --threshold-error)
    THRESHOLD_ERROR="${2:-50}"
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
