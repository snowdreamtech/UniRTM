#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/collect-baseline.sh - Baseline data collection script
#
# Purpose:
#   Establishes performance baselines with cold/warm cache measurements
#   for tool installation times. Stores baseline data in benchmarks/baseline.json
#   with timestamp and git commit hash for traceability.
#
# Usage:
#   ./scripts/collect-baseline.sh [OPTIONS]
#
# Options:
#   --cache-mode <cold|warm|both>  Cache state (default: both)
#   --output <path>                Output file (default: benchmarks/baseline.json)
#   --platform <linux|macos|windows>
#   -h, --help                     Show this help message
#
# Requirements: 1.1, 1.2, 1.3, 4.1, 4.2

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
CACHE_MODE="${CACHE_MODE:-both}"
OUTPUT_FILE="${OUTPUT_FILE:-benchmarks/baseline.json}"
PLATFORM="${PLATFORM:-}"
TIMEOUT_SETUP=1800 # 30 minutes for full setup

# ── Helper Functions ─────────────────────────────────────────────────────────

# Purpose: Display usage information
usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Establishes performance baselines with cold/warm cache measurements.

Options:
  --cache-mode <cold|warm|both>  Cache state to measure (default: both)
  --output <path>                Output file (default: benchmarks/baseline.json)
  --platform <linux|macos|windows>  Override platform detection
  -h, --help                     Show this help message

Examples:
  # Collect both cold and warm cache baselines
  ./scripts/collect-baseline.sh

  # Collect only cold cache baseline
  ./scripts/collect-baseline.sh --cache-mode cold

  # Save to custom location
  ./scripts/collect-baseline.sh --output /tmp/baseline.json

Requirements: 1.1, 1.2, 1.3, 4.1, 4.2
EOF
  exit "${1:-0}"
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

# Purpose: Detect platform
detect_platform() {
  if [ -n "${PLATFORM:-}" ]; then
    echo "$PLATFORM"
    return 0
  fi

  case "$(uname -s)" in
  Darwin) echo "macos" ;;
  Linux) echo "linux" ;;
  MINGW* | MSYS* | CYGWIN*) echo "windows" ;;
  *) echo "unknown" ;;
  esac
}

# Purpose: Clear unirtm cache for cold cache measurement
clear_unirtm_cache() {
  log_info "Clearing unirtm cache for cold cache measurement..."

  # Clear unirtm installs directory
  local unirtm_installs_dir
  case "$(uname -s)" in
  Darwin)
    # macOS: Check both standard and XDG locations
    if [ -d "$HOME/Library/Application Support/unirtm/installs" ]; then
      unirtm_installs_dir="$HOME/Library/Application Support/unirtm/installs"
    elif [ -d "$HOME/.local/share/unirtm/installs" ]; then
      unirtm_installs_dir="$HOME/.local/share/unirtm/installs"
    fi
    ;;
  Linux)
    unirtm_installs_dir="$HOME/.local/share/unirtm/installs"
    ;;
  MINGW* | MSYS* | CYGWIN*)
    # Windows: Check both Git Bash and native Windows locations
    if [ -d "$HOME/.local/share/unirtm/installs" ]; then
      unirtm_installs_dir="$HOME/.local/share/unirtm/installs"
    elif [ -n "${LOCALAPPDATA:-}" ]; then
      if command -v cygpath >/dev/null 2>&1; then
        unirtm_installs_dir="$(cygpath -u "${LOCALAPPDATA}")/unirtm/installs"
      fi
    fi
    ;;
  esac

  if [ -n "${unirtm_installs_dir:-}" ] && [ -d "${unirtm_installs_dir:-}" ]; then
    log_info "  Removing: ${unirtm_installs_dir}"
    rm -rf "${unirtm_installs_dir:?}"/*
    log_success "  Cache cleared successfully"
  else
    log_warn "  UniRTM installs directory not found, skipping cache clear"
  fi
}

# Purpose: Run performance measurement
# Params:
#   $1 - Cache state (cold|warm)
# Returns: JSON output from test-performance.sh
run_performance_measurement() {
  local cache_state="${1:-}"
  local start_time end_time elapsed

  log_info "Running ${cache_state} cache measurement..."

  start_time=$(date +%s)

  # Run test-performance.sh with JSON output
  local perf_output
  if perf_output=$(run_with_timeout "$TIMEOUT_SETUP" "$SCRIPT_DIR/test-performance.sh" --output-format json 2>&1); then
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))
    log_success "  ${cache_state} cache measurement completed in ${elapsed}s"
    echo "$perf_output"
  else
    end_time=$(date +%s)
    elapsed=$((end_time - start_time))
    log_error "  ${cache_state} cache measurement failed after ${elapsed}s"
    return 1
  fi
}

# Purpose: Calculate cache effectiveness metrics
# Params:
#   $1 - Cold cache total time (seconds)
#   $2 - Warm cache total time (seconds)
# Returns: JSON object with cache metrics
calculate_cache_effectiveness() {
  local cold_time="${1:-0}"
  local warm_time="${2:-0}"

  # Avoid division by zero
  if [ "$cold_time" -eq 0 ]; then
    echo '{"speedup_percentage": 0, "poorly_cached_tools": []}'
    return 0
  fi

  # Calculate speedup percentage: ((cold - warm) / cold) * 100
  local speedup_percentage
  speedup_percentage=$(awk "BEGIN {printf \"%.1f\", (($cold_time - $warm_time) / $cold_time) * 100}")

  # For now, we don't have per-tool timing data to identify poorly cached tools
  # This would require parsing the detailed output from test-performance.sh
  # TODO: Implement per-tool cache effectiveness analysis

  cat <<-EOF
{
  "speedup_percentage": $speedup_percentage,
  "poorly_cached_tools": []
}
EOF
}

# Purpose: Build baseline JSON output
# Params:
#   $1 - Cold cache JSON (optional)
#   $2 - Warm cache JSON (optional)
build_baseline_json() {
  local cold_json="${1:-}"
  local warm_json="${2:-}"
  local timestamp commit_sha platform

  timestamp=$(get_timestamp)
  commit_sha=$(get_commit_sha)
  platform=$(detect_platform)

  # Start JSON output
  printf '{\n'
  printf '  "baseline_version": "1.0",\n'
  printf '  "collected_at": "%s",\n' "$timestamp"
  printf '  "commit_sha": "%s",\n' "$commit_sha"
  printf '  "platform": "%s"' "$platform"

  # Add cold cache data if available
  if [ -n "$cold_json" ]; then
    printf ',\n  "cold_cache": %s' "$cold_json"
  fi

  # Add warm cache data if available
  if [ -n "$warm_json" ]; then
    printf ',\n  "warm_cache": %s' "$warm_json"
  fi

  # Add cache effectiveness if both measurements available
  if [ -n "$cold_json" ] && [ -n "$warm_json" ]; then
    local cold_time warm_time
    # Extract total_time_seconds from JSON (simple grep/sed approach)
    cold_time=$(echo "$cold_json" | grep -o '"total_time_seconds":[[:space:]]*[0-9]*' | grep -o '[0-9]*$')
    warm_time=$(echo "$warm_json" | grep -o '"total_time_seconds":[[:space:]]*[0-9]*' | grep -o '[0-9]*$')

    if [ -n "$cold_time" ] && [ -n "$warm_time" ]; then
      local cache_metrics
      cache_metrics=$(calculate_cache_effectiveness "$cold_time" "$warm_time")
      printf ',\n  "cache_effectiveness": %s' "$cache_metrics"
    fi
  fi

  printf '\n}\n'
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Starting baseline data collection..." >&2
  log_info "Cache mode: $CACHE_MODE" >&2
  log_info "Output file: $OUTPUT_FILE" >&2

  local cold_json="" warm_json=""

  # Cold cache measurement
  if [ "$CACHE_MODE" = "cold" ] || [ "$CACHE_MODE" = "both" ]; then
    clear_unirtm_cache
    cold_json=$(run_performance_measurement "cold")
  fi

  # Warm cache measurement
  if [ "$CACHE_MODE" = "warm" ] || [ "$CACHE_MODE" = "both" ]; then
    # If we just did cold measurement, cache is now populated
    # Otherwise, run setup once to populate cache
    if [ "$CACHE_MODE" = "warm" ]; then
      log_info "Running setup to populate cache..."
      if ! run_with_timeout "$TIMEOUT_SETUP" "$SCRIPT_DIR/setup.sh" >/dev/null 2>&1; then
        log_warn "  Setup failed, warm cache measurement may be inaccurate"
      fi
    fi

    warm_json=$(run_performance_measurement "warm")
  fi

  # Build and save baseline JSON
  log_info "Building baseline JSON..." >&2

  # Ensure output directory exists
  local output_dir
  output_dir=$(dirname "$OUTPUT_FILE")
  if [ ! -d "$output_dir" ]; then
    mkdir -p "$output_dir"
  fi

  # Write baseline JSON to file
  build_baseline_json "$cold_json" "$warm_json" >"$OUTPUT_FILE"

  log_success "Baseline data saved to: $OUTPUT_FILE" >&2

  # Display summary
  log_info "" >&2
  log_success "═══════════════════════════════════════════════════════════════" >&2
  log_success "Baseline Collection Complete" >&2
  log_success "═══════════════════════════════════════════════════════════════" >&2
  log_info "" >&2
  log_info "Output: $OUTPUT_FILE" >&2
  log_info "Platform: $(detect_platform)" >&2
  log_info "Commit: $(get_commit_sha)" >&2

  if [ -n "$cold_json" ] && [ -n "$warm_json" ]; then
    local cold_time warm_time speedup
    cold_time=$(echo "$cold_json" | grep -o '"total_time_seconds":[[:space:]]*[0-9]*' | grep -o '[0-9]*$')
    warm_time=$(echo "$warm_json" | grep -o '"total_time_seconds":[[:space:]]*[0-9]*' | grep -o '[0-9]*$')

    if [ -n "$cold_time" ] && [ -n "$warm_time" ]; then
      speedup=$(awk "BEGIN {printf \"%.1f\", (($cold_time - $warm_time) / $cold_time) * 100}")
      log_info "Cold cache: ${cold_time}s" >&2
      log_info "Warm cache: ${warm_time}s" >&2
      log_info "Speedup: ${speedup}%" >&2
    fi
  fi

  log_info "" >&2
  log_success "═══════════════════════════════════════════════════════════════" >&2
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --cache-mode)
    CACHE_MODE="${2:-both}"
    shift 2
    ;;
  --output)
    OUTPUT_FILE="${2:-benchmarks/baseline.json}"
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
