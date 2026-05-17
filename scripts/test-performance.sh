#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/test-performance.sh - Performance testing script for tool installation
#
# Purpose:
#   Measures setup time for all tools, categorizes by type, and outputs structured data
#   for performance regression detection and analysis.
#
# Usage:
#   ./scripts/test-performance.sh [OPTIONS]
#
# Options:
#   --output-format <json|text>  Output format (default: text)
#   --categories <all|security|linters|formatters|runtimes>
#   --verbose                    Enable detailed timing logs
#   --dry-run                    Preview without actual installation
#   -h, --help                   Show this help message
#
# Requirements: 1.1, 1.2, 1.3, 1.4, 1.5

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
OUTPUT_FORMAT="${OUTPUT_FORMAT:-text}"
CATEGORIES="${CATEGORIES:-all}"
INCLUDE_BINARY_RESOLUTION="${INCLUDE_BINARY_RESOLUTION:-0}"
TIMEOUT_PER_TOOL=300 # 5 minutes per tool
# shellcheck disable=SC2034
TIMEOUT_TOTAL=1800 # 30 minutes total (reserved for future use)
# shellcheck disable=SC2034
TIMEOUT_BINARY_VERIFY=5 # 5 seconds per binary verification (Requirement 3.1)
# shellcheck disable=SC2034
TIMEOUT_PLATFORM_RESOLVE=3 # 3 seconds per platform-specific resolution (Requirement 3.3)

# ── Global State ─────────────────────────────────────────────────────────────
TOTAL_START_TIME=0
TOTAL_END_TIME=0
CATEGORY_TIMES=""
TOOL_TIMINGS=""
# shellcheck disable=SC2034
BINARY_RESOLUTION_RESULTS=""

# ── Helper Functions ─────────────────────────────────────────────────────────

# Purpose: Display usage information
usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Measures setup time for all tools and outputs structured performance data.

Options:
  --output-format <json|text>  Output format (default: text)
  --categories <all|security|linters|formatters|runtimes>
                                Tool categories to test (default: all)
  --include-binary-resolution  Include binary resolution benchmarks
  --verbose                    Enable detailed timing logs
  --dry-run                    Preview without actual installation
  -h, --help                   Show this help message

Examples:
  # Run full performance test with JSON output
  ./scripts/test-performance.sh --output-format json

  # Test only security tools
  ./scripts/test-performance.sh --categories security

  # Include binary resolution benchmarks
  ./scripts/test-performance.sh --include-binary-resolution

  # Dry run to preview what would be tested
  ./scripts/test-performance.sh --dry-run

Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 3.1, 3.3
EOF
  exit "${1:-0}"
}

# Purpose: Collect system metadata
# Returns: JSON object with system information
collect_system_metadata() {
  local os arch cpu_cores memory_gb network_type

  # Detect OS
  case "$(uname -s)" in
  Darwin) os="macos" ;;
  Linux) os="linux" ;;
  MINGW* | MSYS* | CYGWIN*) os="windows" ;;
  *) os="unknown" ;;
  esac

  # Detect architecture
  arch=$(uname -m)

  # Detect CPU cores
  if command -v nproc >/dev/null 2>&1; then
    cpu_cores=$(nproc)
  elif command -v sysctl >/dev/null 2>&1; then
    cpu_cores=$(sysctl -n hw.ncpu 2>/dev/null || echo "1")
  else
    cpu_cores="1"
  fi

  # Detect memory (GB)
  if [ -f /proc/meminfo ]; then
    memory_gb=$(awk '/MemTotal/ {printf "%.1f", $2/1024/1024}' /proc/meminfo)
  elif command -v sysctl >/dev/null 2>&1; then
    memory_gb=$(sysctl -n hw.memsize 2>/dev/null | awk '{printf "%.1f", $1/1024/1024/1024}')
  else
    memory_gb="0"
  fi

  # Detect network type
  if [ "${CI:-}" = "true" ] || [ "${GITHUB_ACTIONS:-}" = "true" ]; then
    network_type="github_actions"
  else
    network_type="local"
  fi

  # Output JSON
  cat <<-EOF
{
  "os": "$os",
  "arch": "$arch",
  "cpu_cores": $cpu_cores,
  "memory_gb": $memory_gb,
  "network_type": "$network_type"
}
EOF
}

# Purpose: Get git commit SHA
get_commit_sha() {
  if [ -d .git ]; then
    git rev-parse HEAD 2>/dev/null || echo "unknown"
  else
    echo "unknown"
  fi
}

# Purpose: Get current timestamp in ISO 8601 format
get_timestamp() {
  date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%SZ"
}

# Purpose: Measure time for a tool installation
# Params:
#   $1 - Tool name
#   $2 - Category
# Returns: Time in seconds (stdout)
measure_tool_time() {
  local tool_name="${1:-}"
  local category="${2:-}"
  local start_time end_time elapsed

  start_time=$(date +%s)

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    [ "${VERBOSE:-1}" -ge 1 ] && log_info "  [DRY-RUN] Would measure: $tool_name" >&2
    elapsed=0
  else
    # Run tool installation with timeout
    if run_with_timeout "$TIMEOUT_PER_TOOL" unirtm install "$tool_name" >/dev/null 2>&1; then
      end_time=$(date +%s)
      elapsed=$((end_time - start_time))
      log_info "  ✓ $tool_name: ${elapsed}s" >&2
    else
      end_time=$(date +%s)
      elapsed=$((end_time - start_time))
      log_warn "  ⚠ $tool_name: TIMEOUT or FAILED (${elapsed}s)" >&2
    fi
  fi

  echo "$elapsed"
}

# Purpose: Get tools by category from .unirtm.toml
# Params:
#   $1 - Category name
# Returns: Space-separated list of tools
get_tools_by_category() {
  local category="${1:-}"
  local tools=""

  case "$category" in
  security)
    # Security tools from versions.sh
    tools="gitleaks osv-scanner zizmor"
    ;;
  linters)
    # Linters from versions.sh
    tools="shellcheck hadolint tflint actionlint yamllint"
    ;;
  formatters)
    # Formatters from versions.sh
    tools="shfmt prettier"
    ;;
  runtimes)
    # Runtimes from .unirtm.toml
    if [ -f .unirtm.toml ]; then
      tools=$(grep -E '^\s*[a-z]' .unirtm.toml | grep -v '^\[' | cut -d'=' -f1 | tr -d '"' | tr -d ' ' | head -10)
    fi
    ;;
  all)
    # Combine all categories
    tools="gitleaks osv-scanner zizmor shellcheck hadolint tflint actionlint yamllint shfmt prettier"
    if [ -f .unirtm.toml ]; then
      tools="$tools $(grep -E '^\s*[a-z]' .unirtm.toml | grep -v '^\[' | cut -d'=' -f1 | tr -d '"' | tr -d ' ' | head -10)"
    fi
    ;;
  esac

  echo "$tools"
}

# Purpose: Measure category performance
# Params:
#   $1 - Category name
measure_category() {
  local category="${1:-}"
  local tools category_start category_end category_time
  local tool_list=""

  log_info "Measuring category: $category" >&2

  tools=$(get_tools_by_category "$category")
  category_start=$(date +%s)

  for tool in $tools; do
    [ -z "$tool" ] && continue
    tool_time=$(measure_tool_time "$tool" "$category")
    tool_list="$tool_list $tool"

    # Store timing for slowest tools tracking
    TOOL_TIMINGS="$TOOL_TIMINGS
$tool:$tool_time"
  done

  category_end=$(date +%s)
  category_time=$((category_end - category_start))

  # Store category timing
  CATEGORY_TIMES="$CATEGORY_TIMES
$category:$category_time:$tool_list"

  log_success "Category $category completed in ${category_time}s" >&2
}

# Purpose: Get top N slowest tools
# Params:
#   $1 - Number of tools to return (default: 5)
get_slowest_tools() {
  local count="${1:-5}"

  echo "$TOOL_TIMINGS" | grep -v '^$' | sort -t: -k2 -rn | head -n "$count"
}

# Purpose: Output results in JSON format
output_json() {
  local timestamp commit_sha system_json
  local total_time

  timestamp=$(get_timestamp)
  commit_sha=$(get_commit_sha)
  system_json=$(collect_system_metadata)
  total_time=$((TOTAL_END_TIME - TOTAL_START_TIME))

  # Build categories JSON
  printf '{\n'
  printf '  "timestamp": "%s",\n' "$timestamp"
  printf '  "commit_sha": "%s",\n' "$commit_sha"
  printf '  "system": %s,\n' "$system_json"
  printf '  "total_time_seconds": %d,\n' "$total_time"
  printf '  "categories": {\n'

  # Output categories
  first_cat=true
  echo "$CATEGORY_TIMES" | grep -v '^$' | while IFS=: read -r cat_name cat_time cat_tools; do
    if [ "$first_cat" = false ]; then
      printf ',\n'
    fi
    first_cat=false

    # Build tools array
    printf '    "%s": {\n' "$cat_name"
    printf '      "time_seconds": %s,\n' "$cat_time"
    printf '      "tools": ['

    first_tool=true
    for tool in $cat_tools; do
      if [ "$first_tool" = false ]; then
        printf ','
      fi
      first_tool=false
      printf '"%s"' "$tool"
    done
    printf ']\n'
    printf '    }'
  done

  printf '\n  },\n'
  printf '  "slowest_tools": [\n'

  # Output slowest tools
  first_slow=true
  get_slowest_tools 5 | while IFS=: read -r tool_name tool_time; do
    [ -z "$tool_name" ] && continue
    if [ "$first_slow" = false ]; then
      printf ',\n'
    fi
    first_slow=false
    printf '    {"name": "%s", "time_seconds": %s}' "$tool_name" "$tool_time"
  done

  printf '\n  ]\n'
  printf '}\n'
}

# Purpose: Output results in text format
output_text() {
  local timestamp commit_sha total_time

  timestamp=$(get_timestamp)
  commit_sha=$(get_commit_sha)
  total_time=$((TOTAL_END_TIME - TOTAL_START_TIME))

  log_success "
═══════════════════════════════════════════════════════════════
Performance Test Results
═══════════════════════════════════════════════════════════════

Timestamp:   $timestamp
Commit:      $commit_sha
Total Time:  ${total_time}s

System Information:
$(collect_system_metadata | sed 's/^/  /')

Category Breakdown:
"

  echo "$CATEGORY_TIMES" | grep -v '^$' | while IFS=: read -r cat_name cat_time cat_tools; do
    printf "  %-15s %6ss\n" "$cat_name:" "$cat_time"
  done

  log_success "
Top 5 Slowest Tools:
"

  rank=1
  get_slowest_tools 5 | while IFS=: read -r tool_name tool_time; do
    [ -z "$tool_name" ] && continue
    printf "  %d. %-20s %6ss\n" "$rank" "$tool_name" "$tool_time"
    rank=$((rank + 1))
  done

  echo "
═══════════════════════════════════════════════════════════════"
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Starting performance test..." >&2
  log_info "Output format: $OUTPUT_FORMAT" >&2
  log_info "Categories: $CATEGORIES" >&2

  # Record total start time
  TOTAL_START_TIME=$(date +%s)

  # Measure categories
  if [ "$CATEGORIES" = "all" ]; then
    measure_category "security"
    measure_category "linters"
    measure_category "formatters"
    measure_category "runtimes"
  else
    measure_category "$CATEGORIES"
  fi

  # Record total end time
  TOTAL_END_TIME=$(date +%s)

  # Output results
  if [ "$OUTPUT_FORMAT" = "json" ]; then
    output_json
  else
    output_text
  fi

  log_success "Performance test completed successfully" >&2
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --output-format)
    OUTPUT_FORMAT="${2:-text}"
    shift 2
    ;;
  --categories)
    CATEGORIES="${2:-all}"
    shift 2
    ;;
  --verbose)
    export VERBOSE=2
    shift
    ;;
  --dry-run)
    DRY_RUN=1
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
