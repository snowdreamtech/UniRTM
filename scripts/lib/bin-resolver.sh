#!/usr/bin/env sh
# shellcheck disable=SC2034
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/lib/bin-resolver.sh - Binary resolution module with layered lookup
#
# Purpose:
#   Provides robust binary resolution with timeout protection and caching.
#   Implements a 4-layer lookup strategy to find executables across different
#   environments (venv, node_modules, system PATH, unirtm, filesystem).
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Requirements: 2.1.1, 2.1.2, 3.1, 3.2

# ── 🗂️ Global Cache ──────────────────────────────────────────────────────────

# Global cache for resolved binary paths
# Format: "binary_name:path\n" for each cached entry
# Note: Cache persists within a single script execution context. When functions
# are called via command substitution $(), the cache modifications occur in a
# subshell and don't persist to the parent. This is expected POSIX behavior.
# For optimal caching, source this module and call functions directly.
_G_BIN_CACHE=""

# ── 🔍 Layer 1: Local Cache Lookup ───────────────────────────────────────────

# Purpose: Searches for binaries in local project caches (venv, node_modules)
# Params:
#   $1 - Binary name
# Returns:
#   0 - Binary found (path echoed to stdout)
#   1 - Binary not found
# Examples:
#   resolve_bin_layer1 "eslint"
resolve_bin_layer1() {
  local _BIN="${1:-}"
  [ -z "${_BIN:-}" ] && return 1

  # Python venv lookup
  local _VP="${VENV:-.venv}/${_G_VENV_BIN:-}/${_BIN:-}"
  if [ -x "${_VP:-}" ]; then
    echo "${_VP:-}"
    return 0
  fi

  # Windows: venv scripts use .exe suffix
  if [ "${_G_OS:-}" = "windows" ] && [ -x "${_VP:-}.exe" ]; then
    echo "${_VP:-}.exe"
    return 0
  fi

  # Node modules lookup
  local _NP="node_modules/.bin/${_BIN:-}"
  if [ -x "${_NP:-}" ]; then
    echo "${_NP:-}"
    return 0
  fi

  # Windows: npm generates .cmd wrappers
  if [ "${_G_OS:-}" = "windows" ] && [ -f "${_NP:-}.cmd" ]; then
    echo "${_NP:-}.cmd"
    return 0
  fi

  return 1
}

# ── 🔍 Layer 2: System PATH Lookup ───────────────────────────────────────────

# Purpose: Searches for binaries in system PATH with unirtm/unirtm shim validation
# Params:
#   $1 - Binary name
# Returns:
#   0 - Binary found (path echoed to stdout)
#   1 - Binary not found or invalid shim
# Timeout: 1 second
# Examples:
#   resolve_bin_layer2 "node"
resolve_bin_layer2() {
  local _BIN="${1:-}"
  [ -z "${_BIN:-}" ] && return 1

  local _SP
  _SP=$(command -v "${_BIN:-}" 2>/dev/null) || true

  # Windows: command -v might miss extensions or return sh wrappers
  if [ -z "${_SP:-}" ] && [ "${_G_OS:-}" = "windows" ]; then
    _SP=$(command -v "${_BIN:-}.exe" 2>/dev/null) ||
      _SP=$(command -v "${_BIN:-}.cmd" 2>/dev/null) || true
  fi

  [ -z "${_SP:-}" ] && return 1

  # Check if this is a unirtm or unirtm shim
  case "${_SP:-}" in
  *"${_G_MISE_SHIMS_BASE:-}"*)
    # Validate shim with timeout protection
    local _MW
    if command -v run_with_timeout_robust >/dev/null 2>&1; then
      _MW=$(run_with_timeout_robust 1 unirtm which "${_BIN:-}" 2>/dev/null) ||
        _MW=$(run_with_timeout_robust 1 mise which "${_BIN:-}" 2>/dev/null) || true
    else
      _MW=$(unirtm which "${_BIN:-}" 2>/dev/null) ||
        _MW=$(mise which "${_BIN:-}" 2>/dev/null) || true
    fi

    if [ -n "${_MW:-}" ] && [ -x "${_MW:-}" ]; then
      echo "${_MW:-}"
      return 0
    fi

    # Shim is hollow - try to find non-shim alternative in PATH
    local _OLD_IFS="$IFS"
    IFS=":"
    # shellcheck disable=SC2086
    for _p in $PATH; do
      if [ "${_p:-}" != "${_G_MISE_SHIMS_BASE:-}" ] && [ -x "${_p:-}/${_BIN:-}" ]; then
        IFS="$_OLD_IFS"
        echo "${_p:-}/${_BIN:-}"
        return 0
      fi
    done
    IFS="$_OLD_IFS"

    return 1
    ;;
  *)
    # Not a shim - it's a real system binary
    echo "${_SP:-}"
    return 0
    ;;
  esac
}

# ── 🔍 Layer 3: UniRTM Metadata Query ───────────────────────────────────

# Purpose: Queries unirtm metadata for tool installation path
# Params:
#   $1 - Binary name
# Returns:
#   0 - Binary found (path echoed to stdout)
#   1 - Binary not found
# Timeout: 5 seconds
# Examples:
#   resolve_bin_layer3 "python"
resolve_bin_layer3() {
  local _BIN="${1:-}"
  [ -z "${_BIN:-}" ] && return 1

  # Try unirtm which or unirtm which with timeout protection
  local _MW
  if command -v run_with_timeout_robust >/dev/null 2>&1; then
    _MW=$(run_with_timeout_robust 5 unirtm which "${_BIN:-}" 2>/dev/null) ||
      _MW=$(run_with_timeout_robust 5 mise which "${_BIN:-}" 2>/dev/null) || true
  else
    _MW=$(unirtm which "${_BIN:-}" 2>/dev/null) ||
      _MW=$(mise which "${_BIN:-}" 2>/dev/null) || true
  fi

  if [ -n "${_MW:-}" ] && [ -x "${_MW:-}" ]; then
    echo "${_MW:-}"
    return 0
  fi

  return 1
}

# ── 🔍 Layer 4: Filesystem Search ────────────────────────────────────────────

# Purpose: Searches filesystem for binary using unirtm cache metadata
# Params:
#   $1 - Binary name
# Returns:
#   0 - Binary found (path echoed to stdout)
#   1 - Binary not found
# Timeout: 10 seconds
# Examples:
#   resolve_bin_layer4 "rustc"
resolve_bin_layer4() {
  local _BIN="${1:-}"
  [ -z "${_BIN:-}" ] && return 1

  # Ensure unirtm cache is populated
  if [ -z "${_G_MISE_LS_JSON_CACHE:-}" ]; then
    if command -v refresh_mise_cache >/dev/null 2>&1; then
      refresh_mise_cache
    fi
  fi

  [ -z "${_G_MISE_LS_JSON_CACHE:-}" ] && return 1

  # Extract install path from unirtm cache using awk
  local _MC_PATH
  _MC_PATH=$(echo "${_G_MISE_LS_JSON_CACHE:-}" | awk -v bin="${_BIN:-}" '
    BEGIN { found_bin = 0; }
    # Portable matching of tool key: matches "bin", "prefix:bin", or "prefix:owner/bin"
    # Matches strings ending in "bin" preceded by " , : or /
    $0 ~ "(\"|:|/)" bin "\"" && $0 ~ ":" && $0 ~ "\\[" {
      found_bin = 1;
      next;
    }
    found_bin {
      if ($0 ~ "\"install_path\":") {
        match($0, /"install_path":[[:space:]]*"[^"]+"/);
        if (RSTART > 0) {
    res = substr($0, RSTART, RLENGTH);
          # Extract between quotes: "install_path": "PATH"
          sub(/.*"install_path":[[:space:]]*"/, "", res);
          sub(/"$/, "", res);
          print res;
        }
      }
      # Stop if we hit a new tool key or end of array
      if ($0 ~ /^[[:space:]]*\],?/ || $0 ~ /^[[:space:]]*\}/ || ($0 ~ /^  "[^"]+": \[/ && !($0 ~ bin))) {
        found_bin = 0;
      }
    }
  ' 2>/dev/null | sort -V | tail -n 1) || true

  [ -z "${_MC_PATH:-}" ] || [ "${_MC_PATH:-}" = "null" ] && return 1

  # Search for binary in install path with timeout and depth limit
  local _FOUND_BIN
  if command -v run_with_timeout_robust >/dev/null 2>&1; then
    _FOUND_BIN=$(run_with_timeout_robust 10 find "${_MC_PATH:-}" -maxdepth 3 -name "${_BIN:-}" -type f -perm /111 2>/dev/null | head -n 1) || true
  else
    _FOUND_BIN=$(find "${_MC_PATH:-}" -maxdepth 3 -name "${_BIN:-}" -type f -perm /111 2>/dev/null | head -n 1) || true
  fi

  # Windows fallback: try .exe extension
  if [ -z "${_FOUND_BIN:-}" ] && [ "${_G_OS:-}" = "windows" ]; then
    if command -v run_with_timeout_robust >/dev/null 2>&1; then
      _FOUND_BIN=$(run_with_timeout_robust 10 find "${_MC_PATH:-}" -maxdepth 3 -name "${_BIN:-}.exe" -type f 2>/dev/null | head -n 1) || true
    else
      _FOUND_BIN=$(find "${_MC_PATH:-}" -maxdepth 3 -name "${_BIN:-}.exe" -type f 2>/dev/null | head -n 1) || true
    fi
  fi

  if [ -n "${_FOUND_BIN:-}" ]; then
    echo "${_FOUND_BIN:-}"
    return 0
  fi

  return 1
}

# ── 🎯 Main Resolution Function ──────────────────────────────────────────────

# Purpose: Resolves binary path using layered lookup with caching
# Params:
#   $1 - Binary name
# Returns:
#   0 - Binary found (path echoed to stdout)
#   1 - Binary not found
# Environment:
#   DEBUG_RESOLVE_BIN - Enable detailed debug logging (0/1)
#   VERBOSE - Control verbosity level (0: quiet, 1: normal, 2+: debug)
# Examples:
#   BIN_PATH=$(resolve_bin_cached "eslint") || echo "Not found"
#   DEBUG_RESOLVE_BIN=1 resolve_bin_cached "node"
resolve_bin_cached() {
  local _BIN="${1:-}"
  [ -z "${_BIN:-}" ] && return 1

  # Debug logging helper
  _log_resolve_debug() {
    if [ "${DEBUG_RESOLVE_BIN:-0}" = "1" ] || [ "${VERBOSE:-1}" -ge 2 ]; then
      if command -v log_debug >/dev/null 2>&1; then
        log_debug "$@"
      else
        printf "[DEBUG] %s\n" "$*" >&2
      fi
    fi
  }

  _log_resolve_debug "resolve_bin_cached: Starting lookup for '${_BIN:-}'"

  # Check cache first
  local _CACHED
  _CACHED=$(echo "${_G_BIN_CACHE:-}" | grep "^${_BIN:-}:" | cut -d: -f2-) || true
  if [ -n "${_CACHED:-}" ]; then
    _log_resolve_debug "resolve_bin_cached: Cache hit for '${_BIN:-}' -> '${_CACHED:-}'"
    echo "${_CACHED:-}"
    return 0
  fi

  _log_resolve_debug "resolve_bin_cached: Cache miss for '${_BIN:-}', starting layered lookup"

  # Layer 1: Local cache (venv, node_modules) - no timeout
  local _RESULT
  _log_resolve_debug "resolve_bin_cached: Layer 1 (local cache) - checking venv/node_modules"

  _RESULT=$(resolve_bin_layer1 "${_BIN:-}") && {
    _log_resolve_debug "resolve_bin_cached: Layer 1 SUCCESS - found '${_BIN:-}' at '${_RESULT:-}'"
    _G_BIN_CACHE="${_G_BIN_CACHE:-}${_BIN:-}:${_RESULT:-}
"
    echo "${_RESULT:-}"
    return 0
  }

  _log_resolve_debug "resolve_bin_cached: Layer 1 FAILED - not found in local cache"

  # Layer 2: System PATH with shim validation - 1 second timeout
  _log_resolve_debug "resolve_bin_cached: Layer 2 (system PATH) - checking with shim validation (timeout: 1s)"

  _RESULT=$(resolve_bin_layer2 "${_BIN:-}") && {
    _log_resolve_debug "resolve_bin_cached: Layer 2 SUCCESS - found '${_BIN:-}' at '${_RESULT:-}'"
    _G_BIN_CACHE="${_G_BIN_CACHE:-}${_BIN:-}:${_RESULT:-}
"
    echo "${_RESULT:-}"
    return 0
  }

  _log_resolve_debug "resolve_bin_cached: Layer 2 FAILED - not found in system PATH or invalid shim"

  # Layer 3: UniRTM metadata query - 5 seconds timeout
  _log_resolve_debug "resolve_bin_cached: Layer 3 (unirtm metadata) - querying unirtm which (timeout: 5s)"

  _RESULT=$(resolve_bin_layer3 "${_BIN:-}") && {
    _log_resolve_debug "resolve_bin_cached: Layer 3 SUCCESS - found '${_BIN:-}' at '${_RESULT:-}'"
    _G_BIN_CACHE="${_G_BIN_CACHE:-}${_BIN:-}:${_RESULT:-}
"
    echo "${_RESULT:-}"
    return 0
  }

  _log_resolve_debug "resolve_bin_cached: Layer 3 FAILED - not found via unirtm metadata"

  # Layer 4: Filesystem search - 10 seconds timeout
  _log_resolve_debug "resolve_bin_cached: Layer 4 (filesystem search) - searching with depth limit (timeout: 10s)"

  _RESULT=$(resolve_bin_layer4 "${_BIN:-}") && {
    _log_resolve_debug "resolve_bin_cached: Layer 4 SUCCESS - found '${_BIN:-}' at '${_RESULT:-}'"
    _G_BIN_CACHE="${_G_BIN_CACHE:-}${_BIN:-}:${_RESULT:-}
"
    echo "${_RESULT:-}"
    return 0
  }

  _log_resolve_debug "resolve_bin_cached: Layer 4 FAILED - not found via filesystem search"

  # Not found in any layer
  _log_resolve_debug "resolve_bin_cached: FAILED - '${_BIN:-}' not found in any layer"
  return 1
}

# ── 🧹 Cache Management ──────────────────────────────────────────────────────

# Purpose: Clears the binary resolution cache
# Examples:
#   clear_bin_cache
clear_bin_cache() {
  _G_BIN_CACHE=""
}

# Purpose: Displays cache contents for debugging
# Examples:
#   show_bin_cache
show_bin_cache() {
  if [ -z "${_G_BIN_CACHE:-}" ]; then
    echo "Binary cache is empty"
  else
    echo "Binary cache contents:"
    printf "%s" "${_G_BIN_CACHE:-}" | grep -v '^$' || true
  fi
}

# Export functions for use in other scripts
# Note: export -f is Bash-specific, but we provide the functions for sourcing
