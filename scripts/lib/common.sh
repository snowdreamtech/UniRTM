#!/usr/bin/env sh
set -eu
# shellcheck disable=SC2034
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/lib/common.sh - Shared utility library for automation scripts.
#
# Purpose:
#   Provides centralized configuration, logging, and helper functions
#   used across the project's orchestration layer.
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General, Network), Rule 03 (Architecture), Rule 09 (Interaction).

# ── 🛤️ Path Sentinel (Early Tool Initialization) ────────────────────────────────
# Ensure common binary directories are in PATH early to avoid redundant
# bootstrapping if tools already exist but aren't on the caller's PATH.
for _p in "$HOME/.local/bin" "/usr/local/bin" "/opt/homebrew/bin"; do
  if [ -d "${_p:-}" ]; then
    case ":${PATH:-}:" in
    *":${_p:-}:"*) ;;
    *) export PATH="${_p:-}:${PATH:-}" ;;
    esac
  fi
done
unset _p

# shellcheck disable=SC2034
export PAGER="cat"
export UNIRTM_LOCKFILE=0
export UNIRTM_LOCKED=0
export NO_UPDATE_NOTIFIER=1

# ── 🎨 Visual Assets ─────────────────────────────────────────────────────────

# Colors (using printf to generate literal ESC characters for maximum compatibility)
# Respect NO_COLOR standard (https://no-color.org/) and detect non-TTY for cleaner logs/tests.
if [ -z "${NO_COLOR:-}" ] && [ -t 1 ]; then
  BLUE=$(printf '\033[0;34m')
  GREEN=$(printf '\033[0;32m')
  YELLOW=$(printf '\033[1;33m')
  RED=$(printf '\033[0;31m')
  NC=$(printf '\033[0m')
else
  BLUE="" GREEN="" YELLOW="" RED="" NC=""
fi

# ── 🛠️ POSIX realpath Fallback ───────────────────────────────────────────────
# Ensures compatibility for tools/plugins that expect 'realpath' (missing on macOS/minimal systems).
if ! command -v realpath >/dev/null 2>&1; then
  if command -v grealpath >/dev/null 2>&1; then
    realpath() { grealpath "$@"; }
  elif command -v python3 >/dev/null 2>&1; then
    realpath() { python3 -c "import os, sys; print(os.path.realpath(sys.argv[1]))" "${1:-}"; }
  elif command -v perl >/dev/null 2>&1; then
    realpath() { perl -MCwd -e 'print Cwd::abs_path($ARGV[0])' "${1:-}"; }
  fi
  # shellcheck disable=SC3045
  export -f realpath 2>/dev/null || true
fi

# ── 🎭 Global Environment Detection ──────────────────────────────────────────

# Detect OS and set pathing conventions dynamically to ensure absolute parity
# between Linux, macOS, and Windows (via POSIX shells like Git Bash).
_G_UNAME=$(uname -s)
case "${_G_UNAME:-}" in
Darwin)
  _G_OS="macos"
  _G_VENV_BIN="bin"
  # macOS: unirtm can install in multiple locations
  # 1. Standard: ~/Library/Application Support/unirtm (default)
  # 2. XDG-style: ~/.local/share/unirtm (if XDG_DATA_HOME is set)
  if [ -d "$HOME/Library/Application Support/unirtm/shims" ]; then
    _G_UNIRTM_BIN_BASE="$HOME/.local/bin"
    _G_UNIRTM_SHIMS_BASE="$HOME/Library/Application Support/unirtm/shims"
  elif [ -d "$HOME/.local/share/unirtm/shims" ]; then
    _G_UNIRTM_BIN_BASE="$HOME/.local/bin"
    _G_UNIRTM_SHIMS_BASE="$HOME/.local/share/unirtm/shims"
  else
    # Fallback: assume standard macOS location
    _G_UNIRTM_BIN_BASE="$HOME/.local/bin"
    _G_UNIRTM_SHIMS_BASE="$HOME/Library/Application Support/unirtm/shims"
  fi
  ;;
Linux)
  _G_OS="linux"
  _G_VENV_BIN="bin"
  # Linux: standard XDG Base Directory Specification
  _G_UNIRTM_BIN_BASE="$HOME/.local/bin"
  _G_UNIRTM_SHIMS_BASE="$HOME/.local/share/unirtm/shims"
  ;;
MINGW* | MSYS* | CYGWIN*)
  _G_OS="windows"
  _G_VENV_BIN="Scripts"
  # In Windows-based POSIX shells, AppData/Local is the standard base for unirtm data
  if command -v cygpath >/dev/null 2>&1; then
    _G_APP_DATA_LOCAL=$(cygpath -u "${LOCALAPPDATA:-}")
  else
    # Fallback to manual path translation if cygpath is missing
    _G_APP_DATA_LOCAL=$(echo "${LOCALAPPDATA:-}" | sed 's/\\/\//g; s/:\(.*\)/\/\1/; s/^\([A-Za-z]\)\//\/\L\1\//')
  fi

  # UniRTM on Windows can install in multiple locations - check in order of preference
  # 1. Git Bash style: $HOME/.local/share/unirtm (most common in CI)
  # 2. Windows style: %LOCALAPPDATA%\unirtm (native Windows installs)
  if [ -d "$HOME/.local/share/unirtm/shims" ]; then
    _G_UNIRTM_BIN_BASE="$HOME/.local/bin"
    _G_UNIRTM_SHIMS_BASE="$HOME/.local/share/unirtm/shims"
  elif [ -d "${_G_APP_DATA_LOCAL:-}/unirtm/shims" ]; then
    _G_UNIRTM_BIN_BASE="${_G_APP_DATA_LOCAL:-}/unirtm/bin"
    _G_UNIRTM_SHIMS_BASE="${_G_APP_DATA_LOCAL:-}/unirtm/shims"
  else
    # Fallback: assume Git Bash style (will be created during setup)
    _G_UNIRTM_BIN_BASE="$HOME/.local/bin"
    _G_UNIRTM_SHIMS_BASE="$HOME/.local/share/unirtm/shims"
  fi
  ;;
*)
  _G_OS="linux"
  _G_VENV_BIN="bin"
  _G_UNIRTM_BIN_BASE="$HOME/.local/bin"
  _G_UNIRTM_SHIMS_BASE="$HOME/.local/share/unirtm/shims"
  ;;
esac

# Debug: Log detected paths (only in verbose mode)
if [ "${VERBOSE:-1}" -ge 2 ]; then
  printf "[DEBUG] OS: %s\n" "${_G_OS:-}" >&2
  printf "[DEBUG] UNIRTM_BIN_BASE: %s\n" "${_G_UNIRTM_BIN_BASE:-}" >&2
  printf "[DEBUG] UNIRTM_SHIMS_BASE: %s\n" "${_G_UNIRTM_SHIMS_BASE:-}" >&2
fi

# Purpose: Get unirtm npm installs base path in Unix style (for NODE_PATH in CI)
# Returns: Unix-style path to unirtm npm installs directory
get_unirtm_npm_base() {
  case "$(uname -s)" in
  MINGW* | MSYS* | CYGWIN*)
    # Windows: Convert LOCALAPPDATA to Unix style for Git Bash
    if [ -n "${LOCALAPPDATA:-}" ]; then
      if command -v cygpath >/dev/null 2>&1; then
        local _cyg_appdata
        _cyg_appdata=$(cygpath -u "${LOCALAPPDATA}")
        echo "${_cyg_appdata}/unirtm/installs"
      else
        # Fallback: manual conversion if cygpath is missing
        local _win_base
        _win_base=$(echo "${LOCALAPPDATA}" | sed 's/\\/\//g; s/:\(.*\)/\/\1/; s/^\([A-Za-z]\)\//\/\L\1\//')
        echo "${_win_base}/unirtm/installs"
      fi
    else
      # Fallback to Git Bash style if LOCALAPPDATA is not set
      echo "${HOME}/.local/share/unirtm/installs"
    fi
    ;;
  *)
    # Linux/macOS: Standard XDG path
    echo "${HOME}/.local/share/unirtm/installs"
    ;;
  esac
}

# Purpose: Build NODE_PATH from unirtm npm installations
# Returns: Colon-separated NODE_PATH string with all npm package node_modules
# Examples:
#   export NODE_PATH=$(build_unirtm_node_path)
build_unirtm_node_path() {
  local _npm_base
  _npm_base=$(get_unirtm_npm_base)

  local _extra_node_path=""
  if [ -d "${_npm_base:-}" ]; then
    for _pkg_dir in "${_npm_base}"/npm-*/*/lib/node_modules; do
      [ -d "${_pkg_dir:-}" ] || continue
      _extra_node_path="${_extra_node_path}${_extra_node_path:+:}${_pkg_dir}"
    done
  fi

  # Merge with existing NODE_PATH
  if [ -n "${_extra_node_path:-}" ]; then
    echo "${_extra_node_path}${NODE_PATH:+:}${NODE_PATH:-}"
  else
    echo "${NODE_PATH:-}"
  fi
}

# ── ⚙️ Global Configuration ──────────────────────────────────────────────────

# Default verbosity
# shellcheck disable=SC2034
VERBOSE=${VERBOSE:-1} # 0: quiet, 1: normal, 2: verbose
DRY_RUN=${DRY_RUN:-0}

export UNIRTM_YES=true
export UNIRTM_NON_INTERACTIVE=true
export UNIRTM_QUIET=true
# Suppress unirtm's built-in update checker to avoid GitHub API calls on every invocation.
export UNIRTM_CHECK_FOR_UPDATES=0
# Force unirtm to use system git for better proxy/config compatibility
export UNIRTM_GIT_ALWAYS_USE_GIX=0
export UNIRTM_GIX=0
export UNIRTM_USE_GIX=0
# In CI, ensure GitHub tokens are correctly normalized for GitHub CLI (gh).
# This MUST happen at bootstrap, before any direct tool invocation.
if [ -n "${GITHUB_TOKEN:-}" ]; then
  # GitHub CLI (gh) preferred variable
  [ -z "${GH_TOKEN:-}" ] && export GH_TOKEN="${GITHUB_TOKEN:-}"
fi

# ──  Project Context Detection ──────────────────────────────────────────────
# robustly identify the project root directory relative to the script location.
if [ -z "${_G_PROJECT_ROOT:-}" ]; then
  # Unified Context Detection: Prioritize physical location of the CALLING script ($0).
  # This avoids dependency on caller-defined variables like SCRIPT_DIR.
  _G_CALLER_DIR=$(cd "$(dirname "${0:-}")" && pwd)
  if [ -f "${_G_CALLER_DIR:-}/lib/common.sh" ]; then
    # Caller is in 'scripts/' folder (Standard Orchestration pattern).
    _G_LIB_DIR="${_G_CALLER_DIR:-}/lib"
    _G_PROJECT_ROOT=$(cd "${_G_CALLER_DIR:-}/.." && pwd)
  elif [ -f "${_G_CALLER_DIR:-}/scripts/lib/common.sh" ]; then
    # Caller is in project root (Mock tests or direct root execution).
    _G_LIB_DIR="${_G_CALLER_DIR:-}/scripts/lib"
    _G_PROJECT_ROOT="${_G_CALLER_DIR:-}"
  elif [ -f "${_G_CALLER_DIR:-}/common.sh" ]; then
    # Caller is inside 'scripts/lib/' folder itself.
    _G_LIB_DIR="${_G_CALLER_DIR:-}"
    _G_PROJECT_ROOT=$(cd "${_G_CALLER_DIR:-}/../.." && pwd)
  fi

  # Fallback: Multi-Marker Sentinel (If $0 doesn't lead to library or for direct sourcing)
  if [ -z "${_G_PROJECT_ROOT:-}" ]; then
    if [ -n "${SCRIPT_DIR:-}" ]; then
      if [ -f "${SCRIPT_DIR:-}/lib/common.sh" ]; then
        _G_LIB_DIR="${SCRIPT_DIR:-}/lib"
        _G_PROJECT_ROOT=$(cd "${SCRIPT_DIR:-}/.." && pwd)
      elif [ -f "${SCRIPT_DIR:-}/scripts/lib/common.sh" ]; then
        _G_LIB_DIR="${SCRIPT_DIR:-}/scripts/lib"
        _G_PROJECT_ROOT="${SCRIPT_DIR:-}"
      elif [ -f "${SCRIPT_DIR:-}/common.sh" ]; then
        _G_LIB_DIR="${SCRIPT_DIR:-}"
        _G_PROJECT_ROOT=$(cd "${SCRIPT_DIR:-}/../.." && pwd)
      fi
    fi
  fi

  # Final Fallback: Traverse upwards from PWD (for direct library sourcing or missing markers)
  if [ -z "${_G_PROJECT_ROOT:-}" ]; then
    # Try markers starting from PWD traversal
    _G_PROJECT_ROOT=$(pwd)
    while [ "${_G_PROJECT_ROOT:-}" != "/" ] && [ "${_G_PROJECT_ROOT:-}" != "." ]; do
      if [ -f "${_G_PROJECT_ROOT:-}/package.json" ] || [ -f "${_G_PROJECT_ROOT:-}/Makefile" ] || [ -d "${_G_PROJECT_ROOT:-}/.git" ]; then
        break
      fi
      _G_PROJECT_ROOT=$(dirname "${_G_PROJECT_ROOT:-}")
    done
  fi
  export _G_PROJECT_ROOT
fi
# Resolve global UniRTM binary absolute path early for reliable invocation in subshells and timeouts
if [ -z "${_G_UNIRTM_BIN:-}" ]; then
  if [ -f "${_G_PROJECT_ROOT:-}/unirtm" ] && ! head -n 1 "${_G_PROJECT_ROOT:-}/unirtm" 2>/dev/null | grep -q "^#!"; then
    _G_UNIRTM_BIN="${_G_PROJECT_ROOT:-}/unirtm"
  elif [ -f "${_G_PROJECT_ROOT:-}/unirtm.exe" ]; then
    _G_UNIRTM_BIN="${_G_PROJECT_ROOT:-}/unirtm.exe"
  elif command -v unirtm >/dev/null 2>&1; then
    _G_UNIRTM_BIN=$(command -v unirtm)
  else
    _G_UNIRTM_BIN="${_G_UNIRTM_BIN_BASE:-$HOME/.local/bin}/unirtm"
    [ "${_G_OS:-}" = "windows" ] && _G_UNIRTM_BIN="${_G_UNIRTM_BIN:-}.exe"
  fi
  export _G_UNIRTM_BIN
fi

# ── 📄 SSoT Version Registry ────────────────────────────────────────────────
# Load the centralized version registry to provide a Single Source of Truth
# for all Tier 2 and optional tools across the entire toolchain.
if [ -f "${_G_LIB_DIR:-}/versions.sh" ]; then
  # shellcheck disable=SC1091
  . "${_G_LIB_DIR:-}/versions.sh"
fi

# ── ⏱️ Timeout Configuration ─────────────────────────────────────────────────
# Default timeout values for various operations (seconds)
# These can be overridden via environment variables
TIMEOUT_RESOLVE_BIN="${TIMEOUT_RESOLVE_BIN:-5}"   # Binary resolution
TIMEOUT_JSON_PARSE="${TIMEOUT_JSON_PARSE:-3}"     # JSON parsing
TIMEOUT_UNIRTM_WHICH="${TIMEOUT_UNIRTM_WHICH:-5}" # unirtm which command
TIMEOUT_FIND_BINARY="${TIMEOUT_FIND_BINARY:-10}"  # Filesystem search
TIMEOUT_NETWORK="${TIMEOUT_NETWORK:-30}"          # Network operations

# ── 🐛 Debug Mode Switches ───────────────────────────────────────────────────
# Enable detailed debug logging for specific subsystems
# Set to 1 to enable, 0 to disable
DEBUG_RESOLVE_BIN="${DEBUG_RESOLVE_BIN:-0}" # Binary resolution debug
DEBUG_TIMEOUT="${DEBUG_TIMEOUT:-0}"         # Timeout mechanism debug
DEBUG_JSON_PARSE="${DEBUG_JSON_PARSE:-0}"   # JSON parsing debug

# ── 🔧 Modular Components ────────────────────────────────────────────────────
# Source new modular components for timeout, JSON parsing, process management,
# and binary resolution. These modules provide zero-hang guarantees and robust
# error handling.

# Source timeout mechanism module
if [ -f "${_G_LIB_DIR:-}/timeout.sh" ]; then
  # shellcheck disable=SC1091
  . "${_G_LIB_DIR:-}/timeout.sh"
fi

# Source JSON parser module
if [ -f "${_G_LIB_DIR:-}/json-parser.sh" ]; then
  # shellcheck disable=SC1091
  . "${_G_LIB_DIR:-}/json-parser.sh"
fi

# Source process manager module
if [ -f "${_G_LIB_DIR:-}/process-manager.sh" ]; then
  # shellcheck disable=SC1091
  . "${_G_LIB_DIR:-}/process-manager.sh"
fi

# Source binary resolver module
if [ -f "${_G_LIB_DIR:-}/bin-resolver.sh" ]; then
  # shellcheck disable=SC1091
  . "${_G_LIB_DIR:-}/bin-resolver.sh"
fi

# ── 🔍 Tooling Metadata Cache (UniRTM LS) ──────────────────────────────────────
# Caching 'unirtm ls --json' results provides a massive performance boost for
# scripts that perform multiple version checks (like setup and check-env).
# Initialized LAZILY at first tool resolution, or manually refreshed after installs.
# shellcheck disable=SC2120
refresh_unirtm_cache() {
  # Re-enabled with timeout protection to prevent indefinite hangs
  # Uses 5-second timeout and UNIRTM_OFFLINE=1 to prevent network calls
  if command -v unirtm >/dev/null 2>&1; then
    local _bin="unirtm"
    # Use run_with_timeout_robust if available, otherwise fallback to run_with_timeout
    if command -v run_with_timeout_robust >/dev/null 2>&1; then
      _G_UNIRTM_LS_JSON_CACHE=$(run_with_timeout_robust 5 "${_bin}" ls --json 2>/dev/null || echo "{}")
    else
      _G_UNIRTM_LS_JSON_CACHE=$(UNIRTM_OFFLINE=1 run_with_timeout 5 "${_bin}" ls --json 2>/dev/null || echo "{}")
    fi
  else
    _G_UNIRTM_LS_JSON_CACHE="{}"
  fi
  export _G_UNIRTM_LS_JSON_CACHE
  return 0
}

# Initial state: Empty (triggers lazy load on first resolution)
_G_UNIRTM_LS_JSON_CACHE=""
# ── 📊 CI Step Summary Abstraction (Cross-Platform) ──────────────────────────
# Detect and unify CI summary reporting paths (GitHub, GitLab, Gitea, Local).
# Ref: Rule 09 (Interaction/Summary Integration)
if [ -n "${GITHUB_STEP_SUMMARY:-}" ] && [ -z "${GITEA_ACTIONS:-}" ] && [ -z "${FORGEJO_ACTIONS:-}" ]; then
  # GitHub Actions: Native summary file
  CI_STEP_SUMMARY="${GITHUB_STEP_SUMMARY:-}"
elif [ -n "${GITEA_ACTIONS:-}" ] || [ -n "${FORGEJO_ACTIONS:-}" ]; then
  # Gitea/Forgejo: Often follows GitHub conventions but may need fallback
  CI_STEP_SUMMARY="${GITHUB_STEP_SUMMARY:-${_G_PROJECT_ROOT:-}/.ci_summary.log}"
elif [ -n "${GITLAB_CI:-}" ]; then
  # GitLab: Use a standard log file that can be rendered as an artifact
  CI_STEP_SUMMARY="${_G_PROJECT_ROOT:-}/ci_summary.md"
else
  # Local Development / Other environments: Default to a local log file
  CI_STEP_SUMMARY="${_G_PROJECT_ROOT:-}/.ci_summary.log"
fi
export CI_STEP_SUMMARY

# Mandatory Fallback: Ensure CI_STEP_SUMMARY is NEVER empty (fixes "cannot create : Directory nonexistent")
if [ -z "${CI_STEP_SUMMARY:-}" ]; then
  CI_STEP_SUMMARY="${_G_PROJECT_ROOT:-.}/.ci_summary.log"
fi
export CI_STEP_SUMMARY

# In CI, prevent unirtm from fetching remote version lists from GitHub.
# All versions are pinned in .unirtm.toml / versions.sh, so remote lookups are unnecessary
# and are the biggest hidden source of GitHub API calls during `unirtm install`.
if [ "${CI:-}" = "true" ] || [ "${GITHUB_ACTIONS:-}" = "true" ]; then
  export UNIRTM_FETCH_REMOTE_VERSIONS_TIMEOUT=30s
fi

# Orchestration tracking (detect if we are running as a sub-script)
if [ -z "${_SNOWDREAM_TOP_LEVEL_SCRIPT:-}" ]; then
  _SCRIPT_NAME=$(basename "${0:-}")
  export _SNOWDREAM_TOP_LEVEL_SCRIPT="${_SCRIPT_NAME:-}"
  _IS_TOP_LEVEL=true
else
  _IS_TOP_LEVEL=false
fi

# Truncate the persistent local summary file only at the start of a top-level session (Local Dev only).
if [ "${_IS_TOP_LEVEL:-}" = "true" ] && [ "${CI:-}" != "true" ] && [ "${GITHUB_ACTIONS:-}" != "true" ]; then
  # Ensure the directory exists before truncation
  _CS_DIR=$(dirname "${CI_STEP_SUMMARY:-}")
  [ ! -d "${_CS_DIR:-}" ] && mkdir -p "${_CS_DIR:-}"
  : >"${CI_STEP_SUMMARY:-}"
fi

# ── 📄 SSoT Constants (Paths and Files) ──────────────────────────────────────

VENV="${VENV:-.venv}"
PYTHON="${PYTHON:-python3}"

# ── 🛣️ PATH Augmentation ──────────────────────────────────────────────────────

# Automatically add local bin directories to PATH to ensure orchestrated tools
# are prioritized over system globals without requiring manual activation.
_LOCAL_BIN_VENV=$(pwd)/${VENV:-}/${_G_VENV_BIN:-}
_LOCAL_BIN_NODE=$(pwd)/node_modules/.bin
_LOCAL_UNIRTM_BIN="${_G_UNIRTM_BIN_BASE:-}"
_LOCAL_UNIRTM_SHIMS="${_G_UNIRTM_SHIMS_BASE:-}"

# Resilience: Always attempt to add these paths to ensure toolchain availability
# even if directories are created later (like during setup JIT).
case ":$PATH:" in
*":${_LOCAL_UNIRTM_BIN:-}:"*) ;;
*) export PATH="${_LOCAL_UNIRTM_BIN:-}:$PATH" ;;
esac

case ":$PATH:" in
*":${_LOCAL_UNIRTM_SHIMS:-}:"*) ;;
*) export PATH="${_LOCAL_UNIRTM_SHIMS:-}:$PATH" ;;
esac

case ":$PATH:" in
*":${_LOCAL_BIN_VENV:-}:"*) ;;
*) export PATH="${_LOCAL_BIN_VENV:-}:$PATH" ;;
esac

case ":$PATH:" in
*":${_LOCAL_BIN_NODE:-}:"*) ;;
*) export PATH="${_LOCAL_BIN_NODE:-}:$PATH" ;;
esac

# Purpose: Dynamically detects the Node.js package manager based on lockfiles.
# Returns: "pnpm", "yarn", "bun", or "npm".
_detect_node_manager() {
  if [ -f "pnpm-lock.yaml" ]; then
    echo "pnpm"
  elif [ -f "yarn.lock" ]; then
    echo "yarn"
  elif [ -f "bun.lockb" ]; then
    echo "bun"
  elif [ -f "package-lock.json" ]; then
    echo "npm"
  elif command -v pnpm >/dev/null 2>&1; then
    echo "pnpm"
  elif command -v yarn >/dev/null 2>&1; then
    echo "yarn"
  elif command -v bun >/dev/null 2>&1; then
    echo "bun"
  else
    echo "npm"
  fi
}

# Dynamically# Purpose: Runs a Node.js package manager script safely.
if [ -z "${NPM:-}" ]; then
  NPM=$(_detect_node_manager)
fi
export NPM
# shellcheck disable=SC2034
DOCS_DIR="docs"
PACKAGE_JSON="${PACKAGE_JSON:-package.json}"
REQUIREMENTS_TXT="${REQUIREMENTS_TXT:-requirements.txt}"
PYPROJECT_TOML="${PYPROJECT_TOML:-pyproject.toml}"
CARGO_TOML="${CARGO_TOML:-Cargo.toml}"
VERSION_FILE="${VERSION_FILE:-VERSION}"
CHANGELOG="${CHANGELOG:-CHANGELOG.md}"
LOCK_DIR="${LOCK_DIR:-.archival_lock}"
ARCHIVE_DIR="${ARCHIVE_DIR:-.}"

# Network Optimization & Mirror Configuration
# NOTE: GITHUB_PROXY is optimized for Release/Archive/File downloads.
# It does NOT support project folder clones (git clone).
ENABLE_GITHUB_PROXY="${ENABLE_GITHUB_PROXY:-0}"
GITHUB_PROXY="${GITHUB_PROXY:-https://gh-proxy.sn0wdr1am.com/}"

# ── 🔨 SSoT Tool Versions ────────────────────────────────────────────────────

# Source version registry
# shellcheck source=/dev/null
. "${_G_LIB_DIR:-${_G_PROJECT_ROOT:-}/scripts/lib}/versions.sh"

# Runtime versions (Managed via .unirtm.toml, but some logic might still reference these for bootstrap purposes)
# Only unirtm is hardcoded here to facilitate the zero-dependency bootstrap phase.
UNIRTM_VERSION="${UNIRTM_VERSION:-${VER_UNIRTM:-}}"

# Note: All other tools (Gitleaks, Shellcheck, Shfmt, Java Format, etc.) are purely managed
# by the project's .unirtm.toml file. Do not add hardcoded version variables here.
# Any tool added below MUST have a corresponding entry in .unirtm.toml Tools section.

# Standardized library directory reference (calculated during early detection)
_G_LIB_DIR="${_G_LIB_DIR:-${_G_PROJECT_ROOT:-}/scripts/lib}"
export _G_LIB_DIR

# ── 🛣️ CI Persistence (GitHub Actions) ───────────────────────────────────────
# Note: Moved to regulated block at the end of the file to satisfy ShellCheck SC2218

# ── 🪄 UniRTM Bootstrap ────────────────────────────────────────────────────────
# Logic extracted to ./lib/bootstrap.sh
# shellcheck source=/dev/null
. "${_G_LIB_DIR:-}/bootstrap.sh"

# Purpose: Runs a command with a timeout, handling gtimeout (macOS), timeout (Linux), or Bash fallback.
# Params:
#   $1 - Timeout in seconds
#   $@ - Command and arguments
run_with_timeout() {
  local _SEC="${1:-}"
  shift

  # Delegate to robust implementation if available
  if command -v run_with_timeout_robust >/dev/null 2>&1; then
    run_with_timeout_robust "${_SEC:-}" "$@"
    return $?
  fi

  # Legacy fallback
  if command -v gtimeout >/dev/null 2>&1; then
    gtimeout "${_SEC:-}" "$@"
  elif command -v timeout >/dev/null 2>&1; then
    timeout "${_SEC:-}" "$@"
  else
    # Fallback: Lightweight Bash-native timeout mechanism
    # Works by spawning the command in the background and a sleep watcher.
    # Note: Using setsid or similar would be better but is not POSIX.
    # We attempt to kill the process group by using -P if supported.
    ("$@") &
    local _PID=$!
    (sleep "${_SEC:-}" && kill "${_PID:-}" 2>/dev/null) &
    local _WATCH_PID=$!
    wait "${_PID:-}" 2>/dev/null
    local _RET=$?
    # Cleanup: kill the watcher if it's still running.
    kill "${_WATCH_PID:-}" 2>/dev/null
    return "${_RET:-}"
  fi
}

# ── 🌐 Network Optimization ──────────────────────────────────────────────────

# Purpose: Dynamically detects network connectivity and applies mirrors/proxies.
#          Tests access to GitHub and handles broken global git/proxy settings.
# Examples:
#   optimize_network
optimize_network() {
  if [ "${_NETWORK_OPTIMIZED:-}" = "true" ] || [ "${DRY_RUN:-0}" -eq 1 ]; then return 0; fi

  local _TEMP_GIT_CONFIG
  _TEMP_GIT_CONFIG="${TMPDIR:-/tmp}/.git_config_$(id -u)"

  log_debug "Detecting network connectivity and global proxy health... (DRY_RUN=${DRY_RUN:-0})"

  # 1. Handle Git Protocols & Proxies
  # Guard: If GITHUB_TOKEN is set, verify it's not broken (avoid 401 errors).
  # Test via `/rate_limit` endpoint because GitHub Bot tokens lack `/user` access.
  # Cache: Skip verification if already validated within the last hour (3600s)
  # to avoid hitting the GitHub API on every script invocation.
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    local _TOKEN_CACHE
    _TOKEN_CACHE="${TMPDIR:-/tmp}/.unirtm_token_verified_$(id -u)"
    local _SKIP_VERIFY=false
    if [ -f "${_TOKEN_CACHE:-}" ]; then
      local _CACHE_AGE=0
      # Try BSD stat first (macOS default), fallback to GNU stat (Linux/coreutils)
      if _MTIME=$(stat -f "%m" "${_TOKEN_CACHE:-}" 2>/dev/null); then
        _CACHE_AGE=$(($(date +%s) - _MTIME))
      elif _MTIME=$(stat -c "%Y" "${_TOKEN_CACHE:-}" 2>/dev/null); then
        _CACHE_AGE=$(($(date +%s) - _MTIME))
      else
        _CACHE_AGE=9999 # Force verification if stat fails
      fi
      [ "${_CACHE_AGE:-}" -lt 3600 ] && _SKIP_VERIFY=true
    fi

    if [ "${_SKIP_VERIFY:-}" = "true" ] || [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_debug "GITHUB_TOKEN check skipped (cache/dry-run). Keeping token: ${_SKIPPED_TOKEN:-}"
    else
      local _HTTP_CODE
      # Use --max-time to prevent indefinite hangs in broken network environments
      _HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}" -H "Authorization: Bearer $GITHUB_TOKEN" https://api.github.com/rate_limit --connect-timeout 2 --max-time 10 2>/dev/null || echo "000")
      if [ "${_HTTP_CODE:-}" = "401" ]; then
        log_warn "Current GITHUB_TOKEN appears invalid or unauthorized (${_HTTP_CODE:-}). Unsetting for this session..."
        unset GITHUB_TOKEN
        rm -f "${_TOKEN_CACHE:-}"
      elif [ -z "${_HTTP_CODE:-}" ] || [ "${_HTTP_CODE:-}" = "000" ] || [ "${_HTTP_CODE:-}" = "200" ]; then
        # 200 is success, anything else is treated as transient or invalid (handled above)
        [ "${_HTTP_CODE:-}" = "200" ] && touch "${_TOKEN_CACHE:-}"
        log_debug "Network verification for GITHUB_TOKEN completed: ${_HTTP_CODE:-}"
      else
        log_debug "Unexpected response verifying GITHUB_TOKEN (${_HTTP_CODE:-}). Keeping token."
      fi
    fi
  fi

  # Apply Git optimization and GitHub Proxy if ENABLE_GITHUB_PROXY is active.
  # Registry mirrors (npm, pip, etc.) are now always active via .unirtm.toml [env].
  if [ "${ENABLE_GITHUB_PROXY:-}" = "1" ] || [ "${ENABLE_GITHUB_PROXY:-}" = "true" ]; then
    log_info "Bypassing broken global git proxies and applying network optimization..."

    mkdir -p "$(dirname "${_TEMP_GIT_CONFIG:-}")"
    cat >"${_TEMP_GIT_CONFIG:-}" <<EOF
[http]
  postBuffer = 524288000
  lowSpeedLimit = 0
  lowSpeedTime = 999999
[protocol]
  version = 2
EOF
    export GIT_CONFIG_GLOBAL="${_TEMP_GIT_CONFIG:-}"
    export GIT_CONFIG_SYSTEM="/dev/null"
  fi

  # Ensure unirtm uses a long timeout for HTTP downloads regardless of proxy settings
  export UNIRTM_HTTP_TIMEOUT="${UNIRTM_HTTP_TIMEOUT:-300s}"

  export _NETWORK_OPTIMIZED=true
}

# Purpose: Extracts the configured version of a tool from .unirtm.toml, .unirtm.toml or VER_* env vars.
# Lookup order:
#   1. Exact key match in .unirtm.toml / .unirtm.toml
#   2. Basename match in config files (e.g. "github:foo/bar" -> "bar")
#   3. VER_<UPPER_NAME> env variable set by scripts/lib/versions.sh (Tier 2 tools)
#   4. Fallback: "latest"
# Params:
#   $1 - Tool name (e.g., "rust", "helm", "github:goreleaser/goreleaser")
# Returns:
#   The pinned version string, or "latest" if not found anywhere.
# Examples:
#   VER=$(get_unirtm_tool_version "rust")      # -> VER_RUST from versions.sh
#   VER=$(get_unirtm_tool_version "node")      # -> from .unirtm.toml
get_unirtm_tool_version() {
  local _TOOL_NAME_UNIRTM="${1:-}"
  local _UNIRTM_TOM_PATH
  _UNIRTM_TOM_PATH=$(get_project_root)/.unirtm.toml

  local _VER=""

  if [ -f "${_UNIRTM_TOM_PATH:-}" ]; then
    # 1. Try exact match (including quotes and provider prefix if provider string given)
    _VER=$(grep -E "^\"?${_TOOL_NAME_UNIRTM:-}\"?[[:space:]]*=" "${_UNIRTM_TOM_PATH:-}" 2>/dev/null |
      sed -E 's/^[^=]*=[[:space:]]*"([^"]*)".*/\1/' | head -n 1 || true)

    # 2. Try matching the "basename" of the tool (e.g. github:foo/bar -> bar)
    if [ -z "${_VER:-}" ]; then
      local _SHORT_NAME
      _SHORT_NAME=$(echo "${_TOOL_NAME_UNIRTM:-}" | sed -E 's/^[^:]+://; s/.*\///')
      _VER=$(grep -E "^\"?([^:]+:)?${_SHORT_NAME:-}\"?[[:space:]]*=" "${_UNIRTM_TOM_PATH:-}" 2>/dev/null |
        sed -E 's/^[^=]*=[[:space:]]*"([^"]*)".*/\1/' | head -n 1 || true)
    fi
  fi

  # 3. Check VER_<UPPER> env variable from versions.sh (Tier 2 tools not in .unirtm.toml)
  if [ -z "${_VER:-}" ]; then
    # Normalize: strip provider prefix, take basename, uppercase, replace non-alnum with _
    local _VAR_KEY
    _VAR_KEY=$(echo "${_TOOL_NAME_UNIRTM:-}" |
      sed -E 's/^[^:]+://; s/.*\///' |
      tr '[:lower:]' '[:upper:]' |
      tr -c 'A-Z0-9\n' '_' |
      sed 's/_*$//')
    # Safety: Only eval if key is a valid shell variable name (A-Z, 0-9, _)
    case "${_VAR_KEY:-}" in
    *[!A-Z0-9_]*) ;;
    *) eval "_VER=\${VER_${_VAR_KEY:-}:-}" ;;
    esac
  fi

  # 4. Fallback to 'latest' if no version is explicitly defined anywhere
  echo "${_VER:-latest}"
}

# ── 🔄 GITHUB_PATH Synchronization ──────────────────────────────────────────

# Purpose: Executes a unirtm/unirtm command with retry logic and intelligent fallback.
# Params:
#   $@ - Command and arguments for unirtm/unirtm
# Examples:
#   run_unirtm install node
run_unirtm() {
  local _CMD="${1:-}"
  if [ $# -gt 0 ]; then
    shift
  fi

  # Save the first argument (tool spec) for later use in PATH management
  local _TOOL_ARG="${1:-}"

  # Guard: Only unset GITHUB_TOKEN if we are NOT in CI.
  # In CI (GitHub Actions, etc.), we MUST keep the token to avoid 403 Rate Limit errors.
  # optimize_network() has already verified the token's validity during bootstrap.
  local _OLD_GITHUB_TOKEN="${GITHUB_TOKEN:-}"
  if ! is_ci_env && [ "${GITHUB_TOKEN_FORCE_KEEP:-0}" -ne 1 ]; then
    unset GITHUB_TOKEN
  else
    # Ensure GITHUB_API_TOKEN is set for unirtm's internal GitHub API calls.
    # Workflows set this at env level, but ensure it survives subshell/export boundaries.
    if [ -n "${GITHUB_TOKEN:-}" ] && [ -z "${GITHUB_API_TOKEN:-}" ]; then
      export GITHUB_API_TOKEN="${GITHUB_TOKEN:-}"
      log_debug "Forwarded GITHUB_TOKEN -> GITHUB_API_TOKEN for unirtm."
    fi

    # Ensure UNIRTM_GITHUB_ENTERPRISE_TOKEN is set for unirtm's internal GitHub API calls.
    # Workflows set this at env level, but ensure it survives subshell/export boundaries.
    if [ -n "${GITHUB_TOKEN:-}" ]; then
      if [ -z "${UNIRTM_GITHUB_ENTERPRISE_TOKEN:-}" ]; then
        export UNIRTM_GITHUB_ENTERPRISE_TOKEN="${GITHUB_TOKEN:-}"
        log_debug "Forwarded GITHUB_TOKEN -> UNIRTM_GITHUB_ENTERPRISE_TOKEN for unirtm."
      fi
    fi
  fi

  # Adaptive Lock Forgiveness (ALF)
  # unirtm cannot reliably lock source-compiled tools (go: prefix). To prevent CI
  # failures in --locked mode, we automatically drop the strict requirement
  # for these tools while preserving it for the rest of the orchestration.
  local _EFFECTIVE_LOCKED="${UNIRTM_LOCKED:-}"
  if [ "${_CMD:-}" = "install" ]; then
    if [ $# -eq 0 ]; then
      # Full install (no args): check if config contains any go: tools
      if grep -q '^"go:' "${_G_PROJECT_ROOT:-}/.unirtm.toml" 2>/dev/null; then
        _EFFECTIVE_LOCKED="0"
      fi
    else
      for _arg in "$@"; do
        case "${_arg:-}" in
        go:*)
          _EFFECTIVE_LOCKED="0"
          break
          ;;
        esac
      done
    fi
  fi

  # Smart Bootstrapping: Prioritize local development version, fall back to global
  local _M_BIN=""
  if [ -f "${_G_PROJECT_ROOT:-}/unirtm" ]; then
    _M_BIN="${_G_PROJECT_ROOT:-}/unirtm"
  elif [ -f "${_G_PROJECT_ROOT:-}/unirtm.exe" ]; then
    _M_BIN="${_G_PROJECT_ROOT:-}/unirtm.exe"
  elif command -v unirtm >/dev/null 2>&1; then
    _M_BIN=$(command -v unirtm)
  else
    _M_BIN=$(command -v unirtm || echo "${_G_UNIRTM_BIN_BASE:-$HOME/.local/bin}/unirtm")
    [ "${_G_OS:-}" = "windows" ] && [ ! -x "${_M_BIN:-}" ] && _M_BIN="${_M_BIN:-}.exe"
  fi

  # Performance Opt: Skip installation if version already matches SSoT
  # BUT still ensure PATH synchronization for CI environments
  # CRITICAL: In CI, we MUST verify tool executability, not just version match
  # because unirtm may have hollow shims (especially on Windows)
  local _SKIP_INSTALL=0
  if [ "${_CMD:-}" = "install" ] && [ -n "${1:-}" ]; then
    local _T_CHECK="${1:-}"
    local _R_VER
    _R_VER=$(get_unirtm_tool_version "${_T_CHECK:-}")
    local _T_BASE
    _T_BASE=$(echo "${_T_CHECK:-}" | sed -E 's/^([^:]+:)?(@[^/]+\/)?//; s/.*\///') # Fast-path: Check version-aware existence
    local _C_VER
    _C_VER=$(get_version "${_T_BASE:-}" | tr -d '\r')

    if [ "${_C_VER:-}" != "-" ] && [ -n "${_R_VER:-}" ]; then
      # Use prefix matching: e.g. 3.12.0.2 (required) matches 3.12.0 (current)
      case "${_R_VER:-}" in "${_C_VER:-}"*)
        # In CI, don't skip - the caller (install_* functions) has already verified executability
        # If we reach here in CI, it means the tool needs reinstallation
        if ! is_ci_env; then
          _SKIP_INSTALL=1
        fi
        ;;
      esac
    fi

    # Native/Backend Manager Awareness
    # Ref: https://unirtm.jdx.dev/dev-tools/backends/
    # Ensure required backend package managers are available before attempting installation
    case "${_T_CHECK:-}" in
    cargo:*)
      if ! command -v cargo >/dev/null 2>&1; then
        log_error "Cannot install '${_T_CHECK:-}': 'cargo' (Rust) is missing. Install with: unirtm use -g rust" && return 1
      fi
      ;;
    go:*)
      if ! command -v go >/dev/null 2>&1; then
        log_error "Cannot install '${_T_CHECK:-}': 'go' (Golang) is missing. Install with: unirtm use -g go" && return 1
      fi
      ;;
    npm:*)
      # npm backend supports npm, bun, or pnpm as package managers
      if ! command -v npm >/dev/null 2>&1 && ! command -v bun >/dev/null 2>&1 && ! command -v pnpm >/dev/null 2>&1; then
        log_error "Cannot install '${_T_CHECK:-}': No Node.js package manager found (npm/bun/pnpm). Install with: unirtm use -g node" && return 1
      fi
      ;;
    pipx:*)
      # pipx backend prefers uvx (from uv) but falls back to pipx
      if ! command -v uv >/dev/null 2>&1 && ! command -v pipx >/dev/null 2>&1; then
        log_error "Cannot install '${_T_CHECK:-}': Neither 'uv' nor 'pipx' found. Install with: unirtm use -g uv (or: unirtm use -g python && pip install pipx)" && return 1
      fi
      ;;
    gem:*)
      if ! command -v gem >/dev/null 2>&1; then
        log_error "Cannot install '${_T_CHECK:-}': 'gem' (Ruby) is missing. Install with: unirtm use -g ruby" && return 1
      fi
      ;;
    conda:*)
      # conda backend (experimental) requires conda or mamba
      if ! command -v conda >/dev/null 2>&1 && ! command -v mamba >/dev/null 2>&1; then
        log_error "Cannot install '${_T_CHECK:-}': Neither 'conda' nor 'mamba' found. Install conda/mamba first." && return 1
      fi
      ;;
    dotnet:*)
      # dotnet backend (experimental) requires .NET SDK
      if ! command -v dotnet >/dev/null 2>&1; then
        log_error "Cannot install '${_T_CHECK:-}': 'dotnet' (.NET SDK) is missing. Install from: https://dotnet.microsoft.com/" && return 1
      fi
      ;;
    spm:*)
      # spm backend (experimental) requires Swift Package Manager
      if ! command -v swift >/dev/null 2>&1; then
        log_error "Cannot install '${_T_CHECK:-}': 'swift' (Swift Package Manager) is missing. Install Swift first." && return 1
      fi
      ;;
    esac
  fi

  # ── Execution with Retry & Timeout ──
  local _MAX_RETRIES=3
  local _RETRY_COUNT=0
  local _STATUS=1
  # Use TIMEOUT_NETWORK for network operations (default 30s, can be overridden)
  # For install operations, use longer timeout to handle large GitHub releases
  local _T_OUT="${TIMEOUT_NETWORK:-30}"
  if [ "${_CMD:-}" = "install" ] || [ "${_CMD:-}" = "i" ]; then
    _T_OUT=300 # 300s for install operations
  fi

  local _UNIRTM_OPTS=""
  if [ "${VERBOSE:-1}" -ge 2 ]; then _UNIRTM_OPTS="--verbose"; fi

  # Skip actual installation if tool is already at correct version
  if [ "${_SKIP_INSTALL:-0}" -eq 1 ]; then
    log_debug "Skipping installation - tool already at correct version"
    _STATUS=0
  else
    while [ ${_RETRY_COUNT:-} -lt ${_MAX_RETRIES:-} ]; do
      # Ensure UNIRTM_HTTP_TIMEOUT is synchronized with the execution timeout
      # to prevent internal unirtm/unirtm network calls from hanging the wrapper.
      export UNIRTM_HTTP_TIMEOUT="${_T_OUT:-300}s"

      # Wrap in timeout utility (Standardized via run_with_timeout_robust)
      # shellcheck disable=SC2086
      UNIRTM_LOCKED="${_EFFECTIVE_LOCKED:-}" run_with_timeout_robust "${_T_OUT:-}" "${_M_BIN:-}" ${_UNIRTM_OPTS:-} "${_CMD:-}" "$@"
      _STATUS=$?
      # Exit code 124 = timeout expiry; treat as retryable network failure.
      # Exit codes > 128 = signal (SIGTERM/SIGKILL); abort immediately.
      if [ ${_STATUS:-} -gt 128 ] && [ ${_STATUS:-} -ne 124 ]; then break; fi

      _RETRY_COUNT=$((_RETRY_COUNT + 1))
      if [ ${_RETRY_COUNT:-} -lt ${_MAX_RETRIES:-} ]; then
        # Exponential backoff: 1s, 2s, 4s... to recover from transient rate limits.
        local _BACKOFF=$((1 << (_RETRY_COUNT - 1)))
        log_warn "unirtm/unirtm ${_CMD:-} failed (attempt ${_RETRY_COUNT:-}/${_MAX_RETRIES:-}). Retrying in ${_BACKOFF:-}s..."
        sleep "${_BACKOFF:-}"
      fi
    done
  fi

  # Restore GITHUB_TOKEN
  if [ -n "${_OLD_GITHUB_TOKEN:-}" ]; then
    export GITHUB_TOKEN="${_OLD_GITHUB_TOKEN:-}"
  else
    unset GITHUB_TOKEN
  fi
  # Centralized Metadata Cache Refresh:
  # If we just performed an installation, refresh the global unirtm/unirtm metadata cache
  # to ensure subsequent version checks (get_version) or resolution (resolve_bin)
  # see the newly available tools/binaries immediately.
  if [ ${_STATUS:-} -eq 0 ] &&
    { [ "${_CMD:-}" = "install" ] || [ "${_CMD:-}" = "i" ]; }; then
    refresh_unirtm_cache

    # Unified PATH Management (Task 3.1):
    # Automatically add unirtm/unirtm shims to PATH after successful installation
    # if not already present. This ensures resolve_bin can immediately
    # locate newly installed tools without manual PATH manipulation.
    if [ -n "${_G_UNIRTM_SHIMS_BASE:-}" ]; then
      case ":$PATH:" in
      *":${_G_UNIRTM_SHIMS_BASE:-}:"*) ;;
      *)
        export PATH="${_G_UNIRTM_SHIMS_BASE:-}:$PATH"
        log_debug "Added unirtm shims to PATH: ${_G_UNIRTM_SHIMS_BASE:-}"
        ;;
      esac
    fi

    # Enhanced PATH Management for Dynamically Installed Tools:
    # For tools installed but not activated (not in .unirtm.toml), unirtm/unirtm won't
    # create shims. We need to add the tool's actual bin directory to PATH.
    # This supports the "dynamic install without .unirtm.toml pollution" pattern.
    if [ -n "${_TOOL_ARG:-}" ]; then
      # Extract tool spec (remove version if present)
      local _TOOL_SPEC
      _TOOL_SPEC=$(echo "${_TOOL_ARG:-}" | sed 's/@.*//')

      # Try to get the tool's bin directory from unirtm/unirtm
      if command -v unirtm >/dev/null 2>&1; then
        local _TOOL_BIN_DIR
        # Use unirtm/unirtm where to get the installation path
        _TOOL_BIN_DIR=$("${_M_BIN:-}" where "${_TOOL_SPEC:-}" 2>/dev/null || true)

        if [ -n "${_TOOL_BIN_DIR:-}" ] && [ -d "${_TOOL_BIN_DIR:-}/bin" ]; then
          # Add tool's bin directory to PATH
          case ":$PATH:" in
          *":${_TOOL_BIN_DIR:-}/bin:"*) ;;
          *)
            export PATH="${_TOOL_BIN_DIR:-}/bin:$PATH"
            log_debug "Added tool bin to PATH: ${_TOOL_BIN_DIR:-}/bin"
            ;;
          esac

          # CI PATH Persistence for tool bin directory
          if is_ci_env; then
            _persist_path_to_ci "${_TOOL_BIN_DIR:-}/bin"
          fi
        fi
      fi
    fi

    # CI PATH Persistence (Task 3.2):
    # In CI environments, persist unirtm/unirtm shims to ensure
    # subsequent workflow steps can resolve tools installed in this step.
    if is_ci_env && [ -n "${_G_UNIRTM_SHIMS_BASE:-}" ]; then
      _persist_path_to_ci "${_G_UNIRTM_SHIMS_BASE:-}"
    fi
  fi

  return ${_STATUS:-}
}

# POSIX Shell Function wrapper for direct 'unirtm' calls to use the local binary safely
unirtm() {
  run_unirtm "$@"
}

# ── 📢 Standardized Logging ──────────────────────────────────────────────────

# Standardized logging functions for consistent colored output.
#
# Purpose: Log an informational message in blue.
# Params:
#   $1 - Message to log
# Examples:
#   log_info "Starting build process..."
log_info() {
  local _msg_info="${1:-}"
  if [ "${VERBOSE:-1}" -ge 1 ]; then printf "%s%s%s\n" "${BLUE:-}" "${_msg_info:-}" "${NC:-}"; fi
}

# Purpose: Log a success message in green.
# Params:
#   $1 - Message to log
# Examples:
#   log_success "Build completed successfully."
log_success() {
  local _msg_suc="${1:-}"
  if [ "${VERBOSE:-1}" -ge 1 ]; then printf "%s%s%s\n" "${GREEN:-}" "${_msg_suc:-}" "${NC:-}"; fi
}

# Purpose: Log a warning message in yellow.
# Params:
#   $1 - Message to log
# Examples:
#   log_warn "Dependency 'jq' not found. Some features may be limited."
log_warn() {
  local _msg_warn="${1:-}"
  if [ "${VERBOSE:-1}" -ge 1 ]; then printf "%s%s%s\n" "${YELLOW:-}" "${_msg_warn:-}" "${NC:-}"; fi
}

# Purpose: Log an error message in red to stderr.
# Params:
#   $1 - Message to log
# Examples:
#   log_error "Critical error: Database connection failed."
log_error() {
  local _msg_err="${1:-}"
  printf "%s%s%s\n" "${RED:-}" "${_msg_err:-}" "${NC:-}" >&2
}

# Purpose: Verifies that a required toolchain manager (e.g., cargo, npm, go) is available.
# Params:
#   $1 - Manager command name
# Examples:
#   ensure_manager cargo
ensure_manager() {
  local _MGR="${1:-}"
  if ! command -v "${_MGR:-}" >/dev/null 2>&1; then
    log_error "Error: Toolchain manager '${_MGR:-}' is missing but required for this installation."
    exit 1
  fi
}

# Purpose: Log a debug message if verbose level is 2 or higher.
# Params:
#   $1 - Message to log
# Examples:
#   log_debug "Temporary path: /tmp/build-123"
log_debug() {
  local _msg_dbg="${1:-}"
  if [ "${VERBOSE:-1}" -ge 2 ]; then printf "[DEBUG] %b\n" "${_msg_dbg:-}"; fi
}

# Purpose: Attempts to install a tool using native package managers (brew, apt, choco, etc.)
# Params:
#   $1 - Tool/Package name
# Returns:
#   0 - Success
#   1 - Failure or no manager found
install_native_tool() {
  local _PKG="${1:-}"
  [ -z "${_PKG:-}" ] && return 1

  case "${_G_OS:-}" in
  macos)
    if command -v brew >/dev/null 2>&1; then
      log_info "Installing ${_PKG:-} via Homebrew..."
      brew install "${_PKG:-}" && return 0
    elif command -v port >/dev/null 2>&1; then
      log_info "Installing ${_PKG:-} via MacPorts..."
      sudo port install "${_PKG:-}" && return 0
    fi
    ;;
  linux)
    if command -v apt-get >/dev/null 2>&1; then
      log_info "Installing ${_PKG:-} via apt..."
      sudo apt-get update -y && sudo apt-get install -y "${_PKG:-}" && return 0
    elif command -v dnf >/dev/null 2>&1; then
      log_info "Installing ${_PKG:-} via dnf..."
      sudo dnf install -y "${_PKG:-}" && return 0
    elif command -v pacman >/dev/null 2>&1; then
      log_info "Installing ${_PKG:-} via pacman..."
      sudo pacman -S --noconfirm "${_PKG:-}" && return 0
    fi
    ;;
  windows)
    if command -v choco >/dev/null 2>&1; then
      log_info "Installing ${_PKG:-} via Chocolatey..."
      choco install -y "${_PKG:-}" && return 0
    elif command -v scoop >/dev/null 2>&1; then
      log_info "Installing ${_PKG:-} via Scoop..."
      scoop install "${_PKG:-}" && return 0
    elif command -v winget >/dev/null 2>&1; then
      log_info "Installing ${_PKG:-} via Winget..."
      winget install "${_PKG:-}" && return 0
    fi
    ;;
  esac

  return 1
}

# Purpose: Ensures a tool is available (Check system -> Try Native -> Try UniRTM).
# Params:
#   $1 - Tool name
#   $2 - UniRTM provider name (optional, defaults to tool name)
ensure_tool() {
  local _TOOL="${1:-}"
  local _PRV="${2:-${_TOOL:-}}"

  if command -v "${_TOOL:-}" >/dev/null 2>&1; then
    return 0
  fi

  install_native_tool "${_TOOL:-}" && return 0

  if command -v unirtm >/dev/null 2>&1; then
    "${_G_UNIRTM_BIN:-unirtm}" install "${_PRV:-}" && return 0
  fi

  return 1
}

log_debug "common.sh (v2026.03.14.01) loaded"

# Purpose: Returns the absolute path to the project root directory.
# Returns:
#   Absolute path string.
# Examples:
#   ROOT=$(get_project_root)
get_project_root() {
  local _DIR
  _DIR=$(pwd)
  while [ "${_DIR:-}" != "/" ]; do
    if [ -f "${_DIR:-}/Makefile" ] || [ -f "${_DIR:-}/package.json" ] || [ -d "${_DIR:-}/.git" ]; then
      echo "${_DIR:-}"
      return 0
    fi
    _DIR=$(dirname "${_DIR:-}")
  done
  pwd
}

# Purpose: Verifies that the current working directory is the project root.
#          Exit 1 if critical root files (Makefile or package.json) are missing.
# Examples:
#   guard_project_root
guard_project_root() {
  if [ ! -f "Makefile" ] && [ ! -f "package.json" ]; then
    log_error "Error: This script must be run from the project root."
    exit 1
  fi
}

# Purpose: Checks if a specific string exists in the GITHUB_STEP_SUMMARY.
#          Used as part of the Dual-Sentinel (双重哨兵) pattern.
# Params:
#   $1 - String/Pattern to search for
# Returns:
#   0 - Pattern found
#   1 - Pattern missing
# Examples:
#   if ! check_ci_summary "### Summary"; then ...; fi
check_ci_summary() {
  [ -n "${CI_STEP_SUMMARY:-}" ] && [ -f "${CI_STEP_SUMMARY:-}" ] && grep -qF "${1:-}" "${CI_STEP_SUMMARY:-}"
}

# Purpose: Detects the current CI platform by inspecting well-known environment variables.
# Returns: Platform name string, or "local" if not in a CI environment.
# Examples:
#   PLATFORM=$(detect_ci_platform)
detect_ci_platform() {
  if [ "${FORGEJO_ACTIONS:-}" = "true" ]; then
    echo "forgejo-actions"
  elif [ "${GITEA_ACTIONS:-}" = "true" ]; then
    echo "gitea-actions"
  elif [ "${GITHUB_ACTIONS:-}" = "true" ]; then
    echo "github-actions"
  elif [ "${GITLAB_CI:-}" = "true" ]; then
    echo "gitlab-ci"
  elif [ "${DRONE:-}" = "true" ]; then
    echo "drone"
  elif [ "${WOODPECKER_CI:-}" = "true" ] || [ "${CI:-}" = "woodpecker" ]; then
    echo "woodpecker"
  elif [ "${CIRCLECI:-}" = "true" ]; then
    echo "circleci"
  elif [ "${TRAVIS:-}" = "true" ]; then
    echo "travis"
  elif [ "${TF_BUILD:-}" = "true" ]; then
    echo "azure-pipelines"
  elif [ "${JENKINS_URL:-}" != "" ]; then
    echo "jenkins"
  elif [ "${CI:-}" = "true" ]; then
    echo "ci-unknown"
  else
    echo "local"
  fi
}

# Purpose: Returns true (0) if running in any CI environment, false (1) if local.
# Examples:
#   if is_ci_env; then echo "CI"; fi
is_ci_env() {
  # Use cached global flag if available for performance and consistency
  if [ -n "${_G_IS_CI:-}" ]; then
    [ "${_G_IS_CI:-}" = "1" ]
    return $?
  fi
  [ "$(detect_ci_platform)" != "local" ]
}

# --- GLOBAL CI STATE (Atomic & Inherited) ---

# Calculate once at bootstrap and export for sub-shells (make -> sh -> registry)
if [ -z "${_G_IS_CI:-}" ]; then
  if is_ci_env; then _G_IS_CI=1; else _G_IS_CI=0; fi
  export _G_IS_CI
fi

# Purpose: Detects project language affiliation based on manifests or extensions.
# Params:
#   $1 - Manifest files (space-separated, e.g., "go.mod package.json")
#   $2 - File globs/extensions (space-separated, e.g., "*.go *.ts")
# Returns:
#   0 - Detected (or Force Setup active)
#   1 - Not detected
#
# NOTE: If the environment variable FORCE_SETUP is set to 1, this function
#       will always return 0 (success). This is useful for pre-provisioning
#       environments (like DevContainer image building) where language-specific
#       tools need to be installed before the actual source code is present.
#
# Examples:
#   if has_lang_files "package.json" "*.ts *.js"; then echo "Node project"; fi
has_lang_files() {
  # Support for pre-provisioning: Skip file detection if force mode is active.
  # This allows 'make setup <lang>' to work in headless/empty environments.
  if [ "${FORCE_SETUP:-0}" -eq 1 ]; then
    return 0
  fi

  local _FILES_LANG="${1:-}"
  local _EXTS_LANG="${2:-}"

  # 1. Check for specific config files in root
  local _f_lang
  for _f_lang in ${_FILES_LANG:-}; do
    [ -f "${_f_lang:-}" ] && return 0
  done

  # 2. Check for file extensions (recursive, maxdepth 5 for performance)
  # Exclude common build/dependency/cache directories to avoid false positives and improve speed

  local _ext_lang
  for _ext_lang in ${_EXTS_LANG:-}; do
    # Specialty cases for common multi-file structures
    case "${_ext_lang:-}" in
    CHARTS)
      [ -f "Chart.yaml" ] && return 0
      [ -d "charts" ] && return 0
      ;;
    K8S)
      # Check for common Kubernetes folder or specific schemas
      [ -d "kubernetes" ] && return 0
      [ -d "k8s" ] && return 0
      [ -d "manifests" ] && return 0
      ;;
    HCL)
      [ -f ".terraform.lock.hcl" ] && return 0
      [ -f "terragrunt.hcl" ] && return 0
      # HCL is generic, so we also check for .tf files via the find loop later
      ;;
    PROTOC)
      [ -f "buf.yaml" ] && return 0
      # Also check for .proto files via find loop
      _ext_lang="*.proto"
      ;;
    LUA)
      _ext_lang="*.lua"
      ;;
    JUST)
      [ -f "Justfile" ] && return 0
      [ -f ".justfile" ] && return 0
      ;;
    TASK)
      [ -f "Taskfile.yml" ] && return 0
      [ -f "Taskfile.yaml" ] && return 0
      ;;
    ZIG)
      [ -f "build.zig" ] && return 0
      _ext_lang="*.zig"
      ;;
    CUES)
      _ext_lang="*.cue *.jsonnet"
      ;;
    REGO)
      _ext_lang="*.rego"
      ;;
    EDGE)
      # Modern edge/frontend deployment configs
      [ -f "vercel.json" ] && return 0
      [ -f "netlify.toml" ] && return 0
      ;;
    FLUTTER)
      [ -f "pubspec.yaml" ] && return 0
      _ext_lang="*.dart"
      ;;
    RN)
      [ -f "metro.config.js" ] && return 0
      [ -f "metro.config.ts" ] && return 0
      ;;
    PULUMI)
      [ -f "Pulumi.yaml" ] && return 0
      [ -f "Pulumi.stack.yaml" ] && return 0
      ;;
    CROSSPLANE)
      _ext_lang="*.crossplane.yaml"
      ;;
    PLAYWRIGHT)
      [ -f "playwright.config.ts" ] && return 0
      [ -f "playwright.config.js" ] && return 0
      ;;
    CYPRESS)
      [ -f "cypress.config.ts" ] && return 0
      [ -f "cypress.config.js" ] && return 0
      [ -f "cypress.json" ] && return 0
      ;;
    VITEST)
      [ -f "vitest.config.ts" ] && return 0
      [ -f "vitest.config.js" ] && return 0
      ;;
    DOCUSAURUS)
      [ -f "docusaurus.config.js" ] && return 0
      [ -f "docusaurus.config.ts" ] && return 0
      ;;
    MKDOCS)
      [ -f "mkdocs.yml" ] && return 0
      [ -f "mkdocs.yaml" ] && return 0
      ;;
    SPHINX)
      [ -f "conf.py" ] && return 0
      [ -d "docs/source" ] && return 0
      ;;
    JUPYTER)
      _ext_lang="*.ipynb"
      ;;
    DVC)
      [ -d ".dvc" ] && return 0
      [ -f "dvc.yaml" ] && return 0
      ;;
    ELIXIR)
      [ -f "mix.exs" ] && return 0
      _ext_lang="*.ex *.exs"
      ;;
    HASKELL)
      [ -f "stack.yaml" ] && return 0
      [ -f "package.yaml" ] && return 0
      _ext_lang="*.hs *.cabal"
      ;;
    SCALA)
      [ -f "build.sbt" ] && return 0
      _ext_lang="*.scala"
      ;;
    esac

    # Use find for POSIX compatibility and performance
    # Prune common build/dependency/cache/AI/IDE directories to ensure speed
    # CRITICAL: -maxdepth must come BEFORE -prune for correct behavior
    if [ "$(find . -maxdepth 5 \( -name .git -o -name node_modules -o -name .venv -o -name venv -o -name env -o -name vendor -o -name dist -o -name build -o -name out -o -name target -o -name .next -o -name .nuxt -o -name .output -o -name __pycache__ -o -name .specify -o -name .tmp -o -name tmp -o -name .agent -o -name .agents -o -name .gemini -o -name .trae -o -name .windsurf -o -name .cursor -o -name .cline -o -name .roo -o -name .aide -o -name .bob -o -name .pi -o -name .adal -o -name .zencoder -o -name .supermaven -o -name .neovate -o -name .qoder -o -name .kode -o -name .mux -o -name .shai -o -name .vibe -o -name .void -o -name .factory -o -name .crush -o -name .pochi -o -name .opencode -o -name .iflow -o -name .kilocode -o -name .bito -o -name .amazonq -o -name .codeium -o -name .tabnine -o -name .codegeex -o -name .blackbox -o -name .cody -o -name .continue -o -name .codebuddy -o -name .codex -o -name .cortex -o -name .openhands -o -name .melty -o -name .pearai -o -name .mcpjam -o -name .aider.conf.yml -o -name .commandcode -o -name .goose \) -prune -o -type f -name "${_ext_lang:-}" -print -quit 2>/dev/null)" ]; then
      return 0
    fi
  done

  return 1
}

# Purpose: Executes a command while suppressing its output in quiet mode.
# Params:
#   $@ - Command and arguments to execute
# Examples:
#   run_quiet git rev-parse --is-inside-work-tree
run_quiet() {
  if [ "${VERBOSE:-1}" -eq 0 ]; then
    "$@" >/dev/null 2>&1
  else
    "$@"
  fi
}

# Purpose: Performs an atomic file replacement (Build-then-Swap).
# Params:
#   $1 - Source (temporary) file path
#   $2 - Target destination path
# Returns:
#   0 - Success
#   1 - Source missing
# Examples:
#   atomic_swap "new_config.json.tmp" "config.json"
atomic_swap() {
  local _SRC_ATOMIC="${1:-}"
  local _DST_ATOMIC="${2:-}"
  if [ ! -f "${_SRC_ATOMIC:-}" ]; then
    log_warn "atomic_swap: Source file ${_SRC_ATOMIC:-} does not exist."
    return 1
  fi
  # Use mv for atomic operation on the same filesystem
  mv "${_SRC_ATOMIC:-}" "${_DST_ATOMIC:-}"
}

# Purpose: Orchestrates global argument parsing for all project scripts.
# Params:
#   $@ - Command-line arguments to parse
# Examples:
#   parse_common_args "$@"
parse_common_args() {
  local _arg_common
  for _arg_common in "$@"; do
    case "${_arg_common:-}" in
    --dry-run)
      # shellcheck disable=SC2034
      DRY_RUN=1
      log_warn "Running in DRY-RUN mode. No changes will be applied."
      ;;
    -q | --quiet) # shellcheck disable=SC2034
      VERBOSE=0
      ;;
    -v | --verbose) # shellcheck disable=SC2034
      export VERBOSE=2
      ;;
    -h | --help)
      # shellcheck disable=SC2317
      if command -v show_help >/dev/null 2>&1; then
        show_help
      else
        printf "Usage: %s [OPTIONS]\n\nOptions:\n  --dry-run      Show what would be done\n  -q, --quiet    Suppress output\n  -v, --verbose  Enable debug logging\n  -h, --help     Show this help\n" "${0:-}"
      fi
      exit 0
      ;;
    esac
  done
}

# Purpose: Appends a status record to the centralized execution summary table.
# Params:
#   $1 - Category (e.g., Runtime, Tool, Audit)
#   $2 - Module name (e.g., Node.js, Gitleaks)
#   $3 - Status indicator (e.g., ✅ Success, ❌ Failed)
#   $4 - Version identifier string (or "-" if unavailable)
#   $5 - Duration in seconds (elapsed time)
#   $6 - Summary file path (optional, default: $CI_STEP_SUMMARY)
# Examples:
#   log_summary "Security" "Gitleaks" "✅ Clean" "v8.1.0" "5"
log_summary() {
  local _CAT_SUM="${1:-Other}"
  local _MOD_SUM="${2:-Unknown}"
  local _STAT_SUM="${3:-⏭️ Skipped}"
  local _VER_SUM="${4:--}"
  local _DUR_SUM="${5:--}"
  local _FILE_SUM="${6:-${CI_STEP_SUMMARY:-}}"

  if [ -z "${_FILE_SUM:-}" ] || [ ! -f "${_FILE_SUM:-}" ]; then
    return 0
  fi

  # Automatically demote to Warning if status is supposedly Active/Installed but version detection failed
  case "${_STAT_SUM:-}" in
  ✅*)
    if [ "${_VER_SUM:-}" = "-" ] || [ -z "${_VER_SUM:-}" ]; then
      case "${_MOD_SUM:-}" in
      System | Shell | React | Vue | Tailwind | VitePress | Vite | pnpm-deps | Python-Venv | Homebrew | Hooks | Go-Mod | Cargo-Deps | Ruby-Gems | Go | Rust | Pipx) ;; # These are complex or bootstrap components
      *) _STAT_SUM="⚠️ Warning" ;;
      esac
    fi
    ;;
  esac

  printf "| %-12s | %-15s | %-20s | %-15s | %-6s |\n" "${_CAT_SUM:-}" "${_MOD_SUM:-}" "${_STAT_SUM:-}" "${_VER_SUM:-}" "${_DUR_SUM:-}s" >>"${_FILE_SUM:-}"
}

# Purpose: Safely extracts version strings from various command outputs.
# Params:
#   $1 - Binary or Command name to execute
#   $2 - Argument to fetch version (default: --version)
#   $3 - Optional: Exact UniRTM plugin/provider name for cache lookup
# Returns:
#   Detected version string (stripped) or "-" if command fails/missing.
# Examples:
#   V=$(get_version "node")
#   V=$(get_version "shfmt" "" "shfmt-py")
get_version() {
  local _CMD_VER="${1:-}"
  local _ARG_VER="${2:---version}"
  local _M_PLUGIN="${3:-${_CMD_VER:-}}"
  [ -z "${_CMD_VER:-}" ] && {
    echo "-"
    return 0
  }

  local _BIN_PATH
  _BIN_PATH=$(command -v "${_CMD_VER:-}" 2>/dev/null || true)

  # 1. Try UniRTM First (Fast & Reliable for JIT tools)
  # Query unirtm current directly to check version
  local _UNIRTM_VER_OUT
  if _UNIRTM_VER_OUT=$(UNIRTM_OFFLINE=1 "${_G_UNIRTM_BIN:-unirtm}" current "${_M_PLUGIN:-${_CMD_VER:-}}" 2>/dev/null) && \
     [ -n "${_UNIRTM_VER_OUT:-}" ] && \
     [ "${_UNIRTM_VER_OUT:-}" != "null" ] && \
     ! echo "${_UNIRTM_VER_OUT:-}" | grep -q "✗"; then
    echo "${_UNIRTM_VER_OUT:-}" | tr -d '\r' && return 0
  fi

  # Fallback to system command or unirtm direct binary
  local _LV_RESOLVED
  _LV_RESOLVED=$(resolve_bin "${_CMD_VER:-}") || true

  if [ -n "${_LV_RESOLVED:-}" ]; then
    # Guard: If binary is a unirtm shim but version wasn't found in cache,
    # it likely means the tool is installed globally (pipx, etc.) but not in .unirtm.toml.
    # To prevent 'unirtm ERROR No version is set', we use 'unirtm exec' if it's a shim.
    case "${_LV_RESOLVED:-}" in
    *"${_G_UNIRTM_SHIMS_BASE:-}"*)
      "${_G_UNIRTM_BIN:-unirtm}" exec "${_M_PLUGIN:-latest}" -- "${_CMD_VER:-}" "${_ARG_VER:---version}" 2>/dev/null | grep -E -o "[0-9]+\.[0-9]+\.[0-9]+" | head -n 1
      return 0
      ;;
    esac

    # Special cases for tools with unusual version output or slow shims
    case "${_CMD_VER:-}" in
    python*)
      "${_LV_RESOLVED:-}" --version 2>/dev/null | cut -d' ' -f2 && return 0
      ;;
    node)
      "${_LV_RESOLVED:-}" --version 2>/dev/null | sed 's/^v//'
      ;;
    go)
      "${_LV_RESOLVED:-}" version 2>/dev/null | awk '{print $3}' | sed 's/^go//'
      ;;
    java)
      # java -version outputs to stderr and puts version in quotes
      "${_LV_RESOLVED:-}" "${_ARG_VER:-}" 2>&1 | sed -n 's/.*version "\([0-9][0-9.]*\).*/\1/p' | head -n 1
      ;;
    vitepress | docusaurus)
      # Avoid running binaries that might start a dev server on --version/--help
      echo "-"
      ;;
    *)
      # For other binaries, try to get version from the output
      # We strip 'v' or 'V' prefix and focus on the version number
      # Use UNIRTM_OFFLINE=1 to prevent shims from trying to resolve versions over network
      # On Windows, ensures we don't hang on interactive prompts or external lookups.
      # shellcheck disable=SC2155
      local _VERSION_RAW
      _VERSION_RAW="$(UNIRTM_OFFLINE=1 "${_LV_RESOLVED:-}" "${_ARG_VER:-}" 2>/dev/null | tr -d '\r' | sed 's/^[vV]//' | grep -o '[0-9][0-9.]*' | head -n 1 | cut -c1-15 2>/dev/null)"
      echo "${_VERSION_RAW:--}" && return 0
      ;;

    esac
  else
    echo "-"
  fi
}

# Purpose: Resolves the executable path for a tool across venv, node_modules,
#          system PATH, and unirtm-managed environments (shim + direct install).
# Params:
#   $1 - Binary name (e.g., "eslint", "pytest", "bats")
# Returns:
#   Echoes the resolved path. Exit 0 on success, 1 if not found.
# Environments:
#   - Local dev (with/without unirtm cache)
#   - CI runners (clean or pre-cached)
#   - Windows (Git Bash/MSYS2), macOS, Linux
# Examples:
#   BIN=$(resolve_bin "eslint") || true
# Feature Flag:
#   USE_NEW_RESOLVE_BIN=1 - Use new layered implementation with timeout protection
#   USE_NEW_RESOLVE_BIN=0 - Use legacy implementation (default for gradual rollout)
# Params:
#   $1 - Binary name to resolve
#   $2 - Optional: Additional search paths (colon-separated, e.g., "docs/node_modules/.bin:other/path")
# Examples:
#   resolve_bin "vitepress"
#   resolve_bin "vitepress" "docs/node_modules/.bin"
#   resolve_bin "eslint" "packages/app/node_modules/.bin:packages/lib/node_modules/.bin"
resolve_bin() {
  local _BIN="${1:-}"
  local _EXTRA_PATHS="${2:-}"
  [ -z "${_BIN:-}" ] && return 1

  # Feature flag: Use new implementation if enabled
  if [ "${USE_NEW_RESOLVE_BIN:-0}" = "1" ]; then
    # Delegate to new layered implementation with timeout protection
    if command -v resolve_bin_cached >/dev/null 2>&1; then
      resolve_bin_cached "${_BIN:-}" "${_EXTRA_PATHS:-}"
      return $?
    else
      # Fallback to legacy if new implementation not available
      log_debug "resolve_bin_cached not found, falling back to legacy implementation"
    fi
  fi

  # ── 0. Extra Search Paths (highest priority) ──
  if [ -n "${_EXTRA_PATHS:-}" ]; then
    local _OLD_IFS="$IFS"
    IFS=":"
    # shellcheck disable=SC2086
    for _extra_path in ${_EXTRA_PATHS:-}; do
      if [ -x "${_extra_path:-}/${_BIN:-}" ]; then
        IFS="$_OLD_IFS"
        echo "${_extra_path:-}/${_BIN:-}"
        return 0
      fi
      # Windows: check .exe and .cmd extensions
      if [ "${_G_OS:-}" = "windows" ]; then
        if [ -x "${_extra_path:-}/${_BIN:-}.exe" ]; then
          IFS="$_OLD_IFS"
          echo "${_extra_path:-}/${_BIN:-}.exe"
          return 0
        fi
        if [ -f "${_extra_path:-}/${_BIN:-}.cmd" ]; then
          IFS="$_OLD_IFS"
          echo "${_extra_path:-}/${_BIN:-}.cmd"
          return 0
        fi
      fi
    done
    IFS="$_OLD_IFS"
  fi

  # ── 1. Python Venv ──
  local _VP="${VENV:-.venv}/${_G_VENV_BIN:-}/${_BIN:-}"
  if [ -x "${_VP:-}" ]; then echo "${_VP:-}" && return 0; fi
  # Windows: venv scripts use .exe suffix
  if [ "${_G_OS:-}" = "windows" ] && [ -x "${_VP:-}.exe" ]; then echo "${_VP:-}.exe" && return 0; fi

  # ── 2. Node Modules ──
  local _NP="node_modules/.bin/${_BIN:-}"
  if [ -x "${_NP:-}" ]; then echo "${_NP:-}" && return 0; fi
  # Windows: npm generates .cmd wrappers
  if [ "${_G_OS:-}" = "windows" ] && [ -f "${_NP:-}.cmd" ]; then echo "${_NP:-}.cmd" && return 0; fi

  # ── 3. System PATH ──
  local _SP
  _SP=$(command -v "${_BIN:-}" 2>/dev/null) || true

  # Windows: command -v might miss extensions or return sh wrappers
  if [ -z "${_SP:-}" ] && [ "${_G_OS:-}" = "windows" ]; then
    _SP=$(command -v "${_BIN:-}.exe" 2>/dev/null) || _SP=$(command -v "${_BIN:-}.cmd" 2>/dev/null) || true
  fi

  if [ -n "${_SP:-}" ]; then
    # Guard: If it resolves to a unirtm shim, verify the tool is actually installed.
    # Shims exist for ALL tools in .unirtm.toml, even uninstalled ones ("hollow shims").
    case "${_SP:-}" in
    *"${_G_UNIRTM_SHIMS_BASE:-}"*)
      # Use 'unirtm which' — the lightweight, jq-free way to validate a shim.
      # Returns the real binary path if installed, non-zero if not.
      local _MW
      _MW=$(UNIRTM_OFFLINE=1 run_with_timeout_robust 3 "${_G_UNIRTM_BIN:-unirtm}" which "${_BIN:-}" 2>/dev/null) || true
      if [ -n "${_MW:-}" ] && [ -x "${_MW:-}" ]; then
        echo "${_MW:-}" && return 0
      fi

      # Shim is hollow — try to find another match in the path that is NOT a shim.
      # This enables BATS mocks and system fallback when unirtm is inactive.
      local _OLD_IFS="$IFS"
      IFS=":"
      # shellcheck disable=SC2086
      for _p in $PATH; do
        if [ "${_p:-}" != "${_G_UNIRTM_SHIMS_BASE:-}" ] && [ -x "${_p:-}/${_BIN:-}" ]; then
          IFS="$_OLD_IFS" && echo "${_p:-}/${_BIN:-}" && return 0
        fi
      done
      IFS="$_OLD_IFS"
      # Fall through to Layer 4.
      ;;
    *)
      # Not a shim — it's a real system binary.
      echo "${_SP:-}" && return 0
      ;;
    esac
  fi

  # ── 4. UniRTM direct lookup (no shim in PATH, e.g., fresh CI) ──
  # Covers: unirtm installed the tool but shims/PATH not yet activated.
  local _MW
  _MW=$(UNIRTM_OFFLINE=1 run_with_timeout_robust 3 "${_G_UNIRTM_BIN:-unirtm}" which "${_BIN:-}" 2>/dev/null) || true
  if [ -n "${_MW:-}" ] && [ -x "${_MW:-}" ]; then
    echo "${_MW:-}" && return 0
  fi

  # ── 5. UniRTM Cache Fallback (Metadata-aware) ──
  # Handles tools installed JIT (e.g., Tier 2) but not in active .unirtm.toml
  # which causes 'unirtm which' to fail even if the tool exists on disk.
  if [ -z "${_G_UNIRTM_LS_JSON_CACHE:-}" ]; then refresh_unirtm_cache; fi
  _MC_PATH=$(echo "${_G_UNIRTM_LS_JSON_CACHE:-}" | awk -v bin="${_BIN:-}" '
    $0 ~ "\"tool\":" {
      match($0, /"tool":[[:space:]]*"[^"]+"/);
      if (RSTART > 0) {
        curr_tool = substr($0, RSTART, RLENGTH);
        sub(/.*"tool":[[:space:]]*"/, "", curr_tool);
        sub(/"$/, "", curr_tool);
      }
    }
    $0 ~ "\"install_path\":" && (curr_tool == bin || curr_tool ~ ("/" bin "$") || curr_tool ~ (":" bin "$")) {
      match($0, /"install_path":[[:space:]]*"[^"]+"/);
      if (RSTART > 0) {
        res = substr($0, RSTART, RLENGTH);
        sub(/.*"install_path":[[:space:]]*"/, "", res);
        sub(/"$/, "", res);
        print res;
      }
    }
  ' 2>/dev/null | sort -V | tail -n 1 || true)

  if [ -n "${_MC_PATH:-}" ] && [ "${_MC_PATH:-}" != "null" ]; then
    # Robustly find the binary within the install_path (maxdepth 3 for performance)
    _FOUND_BIN=$(find "${_MC_PATH:-}" -maxdepth 3 -name "${_BIN:-}" -type f -perm +111 2>/dev/null | head -n 1) || true
    if [ -z "${_FOUND_BIN:-}" ] && [ "${_G_OS:-}" = "windows" ]; then
      _FOUND_BIN=$(find "${_MC_PATH:-}" -maxdepth 3 -name "${_BIN:-}.exe" -type f 2>/dev/null | head -n 1) || true
    fi
    if [ -n "${_FOUND_BIN:-}" ]; then
      echo "${_FOUND_BIN:-}" && return 0
    fi
  fi

  return 1
}

# Purpose: Dynamically extracts the project version from manifest files.
# Returns: Version string (detected) or "0.0.0" (fallback).
# Examples:
#   VER=$(get_project_version)
get_project_version() {
  if [ -f "${PACKAGE_JSON:-}" ]; then
    grep '"version":' "${PACKAGE_JSON:-}" | head -n 1 | sed 's/.*"version":[[:space:]]*"//;s/".*//'
  elif [ -f "${CARGO_TOML:-}" ]; then
    grep '^version =' "${CARGO_TOML:-}" | head -n 1 | sed -e 's/.*"\(.*\)"/\1/' -e "s/.*'\(.*\)'/\1/"
  elif [ -f "${PYPROJECT_TOML:-}" ]; then
    grep '^version =' "${PYPROJECT_TOML:-}" | head -n 1 | sed 's/.*"//;s/".*//'
  elif [ -f "${VERSION_FILE:-}" ]; then
    awk 'NR==1' "${VERSION_FILE:-}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
  else
    echo "0.0.0"
  fi
}

# Purpose: Verifies that a required runtime is available.
#          Gracefully skips (exit 0) if missing, for linter wrappers.
# Params:
#   $1 - Runtime name (e.g., "go", "python")
#   $2 - Tool description (e.g., "Golang Builder")
# Examples:
#   check_runtime "go" "Golang Builder"
check_runtime() {
  local _RT_NAME="${1:-}"
  local _TOOL_DESC="${2:-Tool}"

  # Priority 1: Modular Check Delegation
  # If check_runtime_<name> exists in the environment, delegate to it.
  if command -v "check_runtime_${_RT_NAME:-}" >/dev/null 2>&1; then
    if ! "check_runtime_${_RT_NAME:-}" "${_TOOL_DESC:-}"; then
      if [ "${_G_AUDIT_MODE:-0}" -eq 1 ]; then
        return 0
      fi
      exit 0 # Graceful skip for pre-commit
    fi
    return 0
  fi

  # Priority 2: In audit mode, skip resolve_bin check entirely
  # The tool version was already checked by check_tool_version above.
  # This prevents hanging on unirtm operations during environment health checks.
  if [ "${_G_AUDIT_MODE:-0}" -eq 1 ]; then
    log_debug "Audit mode: Skipping runtime check for '${_RT_NAME:-}' (already verified by check_tool_version)"
    return 0
  fi

  # Priority 3: Standard Command Check (Fallback using resolve_bin)
  # Only used in pre-commit hooks, not in audit mode.
  if ! resolve_bin "${_RT_NAME:-}" >/dev/null 2>&1; then
    log_warn "Required runtime '${_RT_NAME:-}' for ${_TOOL_DESC:-} is missing. Skipping."
    exit 0 # Graceful skip for pre-commit
  fi
}

# Purpose: Verify tool binary exists and is executable (lightweight check).
# This is a fast pre-check before version detection to avoid false positives
# from unirtm metadata that may be stale or incorrect.
#
# Parameters:
#   $1 - Binary name (e.g., "shfmt")
#   $2 - Version flag (e.g., "--version") [optional, defaults to "--version"]
#
# Returns:
#   0 - Binary exists and is executable
#   1 - Binary missing or not executable
#
# Examples:
#   verify_binary_exists "shfmt" "--version"
#   verify_binary_exists "ruff"
verify_binary_exists() {
  local _BIN="${1:-}"
  local _VER_FLAG="${2:---version}"

  [ -z "${_BIN:-}" ] && return 1

  # Method 1: Check if unirtm can find it (most reliable in CI)
  if "${_G_UNIRTM_BIN:-unirtm}" which "${_BIN:-}" >/dev/null 2>&1; then
    return 0
  fi

  # Method 2: Check if in PATH
  if ! command -v "${_BIN:-}" >/dev/null 2>&1; then
    return 1
  fi

  # Method 3: Try to execute it with timeout protection
  # CRITICAL: For unirtm shims, direct execution may fail with "not currently active"
  # We need to handle this gracefully by checking if it's a shim
  local _BIN_PATH
  _BIN_PATH=$(command -v "${_BIN:-}" 2>/dev/null)

  # If it's a unirtm shim, it will be in the shims directory
  if echo "${_BIN_PATH:-}" | grep -q "/unirtm/shims/"; then
    # For unirtm shims, we can't reliably test execution without unirtm exec
    # Just verify the shim file exists
    # On Windows, skip executable check (files are executable by extension)
    if [ -f "${_BIN_PATH:-}" ]; then
      if [ "${_G_OS:-}" = "windows" ] || [ -x "${_BIN_PATH:-}" ]; then
        return 0
      fi
    fi
  else
    # For non-shim binaries, try to execute with timeout
    if run_with_timeout_robust 3 "${_BIN:-}" "${_VER_FLAG:-}" >/dev/null 2>&1; then
      return 0
    fi
  fi

  return 1
}

# Purpose: Atomically verifies tool installation with comprehensive checks.
#          Ensures: installed, exists, executable, resolvable, usable.
# Params:
#   $1 - Tool binary name (e.g., "shfmt", "ec")
#   $2 - Tool provider spec (e.g., "github:mvdan/sh")
#   $3 - Tool display name (e.g., "Shfmt")
# Returns:
#   0 - Tool is fully verified and usable
#   1 - Tool verification failed
# Purpose: Atomic verification for unirtm-installed tools with robust binary resolution.
# This function ensures a tool is fully functional through 5 verification steps,
# using unirtm which as the primary resolution method for cross-platform compatibility.
#
# Parameters:
#   $1 - Binary name (e.g., "shfmt", "editorconfig-checker")
#   $2 - UniRTM provider (e.g., "github:mvdan/sh")
#   $3 - Display name (e.g., "Shfmt") [optional]
#   $4 - Version flag (e.g., "--version") [optional, defaults to "--version"]
#
# Returns:
#   0 - Tool is fully verified and usable
#   1 - Tool verification failed
#
# Examples:
#   verify_tool_atomic "shfmt" "github:mvdan/sh" "Shfmt"
#   verify_tool_atomic "editorconfig-checker" "github:editorconfig-checker/editorconfig-checker" "Editorconfig-Checker"
#   verify_tool_atomic "ruff" "github:astral-sh/ruff" "Ruff" "--version"
verify_tool_atomic() {
  local _BIN_NAME="${1:-}"
  local _PROVIDER="${2:-}"
  local _DISPLAY_NAME="${3:-Tool}"
  local _VERSION_FLAG="${4:---version}"

  [ -z "${_BIN_NAME:-}" ] && return 1

  log_debug "=== Atomic Verification: ${_DISPLAY_NAME:-} ==="

  # Step 1: Check if tool is registered in unirtm
  log_debug "Step 1/5: Checking unirtm registration..."
  if ! "${_G_UNIRTM_BIN:-unirtm}" list 2>/dev/null | grep -q "${_PROVIDER:-}"; then
    log_error "✗ ${_DISPLAY_NAME:-} not registered in unirtm"
    return 1
  fi
  log_debug "✓ Registered in unirtm"

  # Step 2: Resolve binary path using unirtm which (primary method)
  # This handles platform-specific binaries, shims, and cross-platform compatibility
  log_debug "Step 2/5: Resolving binary path via unirtm which..."
  local _RESOLVED_PATH
  _RESOLVED_PATH=$(UNIRTM_OFFLINE=1 run_with_timeout_robust 3 "${_G_UNIRTM_BIN:-unirtm}" which "${_BIN_NAME:-}" 2>/dev/null) || _RESOLVED_PATH=""

  if [ -z "${_RESOLVED_PATH:-}" ]; then
    log_debug "✗ unirtm which failed, trying fallback methods..."

    # Fallback 1: Try command -v (for tools already in PATH)
    _RESOLVED_PATH=$(command -v "${_BIN_NAME:-}" 2>/dev/null) || _RESOLVED_PATH=""

    if [ -z "${_RESOLVED_PATH:-}" ]; then
      # Fallback 2: Search in unirtm install directory
      local _INSTALL_DIR
      _INSTALL_DIR=$("${_G_UNIRTM_BIN:-unirtm}" where "${_PROVIDER:-}" 2>/dev/null) || _INSTALL_DIR=""
      log_debug "unirtm where ${_PROVIDER:-} returned: ${_INSTALL_DIR:-<empty>}"

      if [ -n "${_INSTALL_DIR:-}" ]; then
        # Try bin/ directory first
        if [ -d "${_INSTALL_DIR:-}/bin" ]; then
          # Try exact match first
          if [ -f "${_INSTALL_DIR:-}/bin/${_BIN_NAME:-}" ]; then
            _RESOLVED_PATH="${_INSTALL_DIR:-}/bin/${_BIN_NAME:-}"
          else
            # Try pattern match (e.g., ec-* for editorconfig-checker)
            _RESOLVED_PATH=$(find "${_INSTALL_DIR:-}/bin" -name "${_BIN_NAME:-}*" -type f 2>/dev/null | head -n 1)
          fi
        fi

        # If not found in bin/, try root directory
        if [ -z "${_RESOLVED_PATH:-}" ]; then
          if [ -f "${_INSTALL_DIR:-}/${_BIN_NAME:-}" ]; then
            _RESOLVED_PATH="${_INSTALL_DIR:-}/${_BIN_NAME:-}"
          else
            # Try pattern match in root directory
            _RESOLVED_PATH=$(find "${_INSTALL_DIR:-}" -maxdepth 1 -name "${_BIN_NAME:-}*" -type f 2>/dev/null | head -n 1)
          fi
        fi
      fi

      if [ -z "${_RESOLVED_PATH:-}" ]; then
        # Fallback 3: Try to find in unirtm installs directory by searching for provider pattern
        log_debug "Searching in unirtm installs directory..."
        local _UNIRTM_INSTALLS
        _UNIRTM_INSTALLS=$(get_unirtm_npm_base)
        if [ -d "${_UNIRTM_INSTALLS:-}" ]; then
          # For GitHub providers, the directory name pattern is: github-owner-repo/version
          # e.g., github-mvdan-sh/3.13.1
          local _PROVIDER_DIR
          _PROVIDER_DIR=$(echo "${_PROVIDER:-}" | sed 's|github:||; s|/|-|g')
          log_debug "Looking for provider directory pattern: ${_PROVIDER_DIR:-}"

          # Find the most recent version directory
          local _TOOL_DIR
          _TOOL_DIR=$(find "${_UNIRTM_INSTALLS:-}" -maxdepth 2 -type d -name "${_PROVIDER_DIR:-}*" 2>/dev/null | sort -V | tail -n 1)

          if [ -n "${_TOOL_DIR:-}" ] && [ -d "${_TOOL_DIR:-}/bin" ]; then
            log_debug "Found tool directory: ${_TOOL_DIR:-}"
            if [ -f "${_TOOL_DIR:-}/bin/${_BIN_NAME:-}" ]; then
              _RESOLVED_PATH="${_TOOL_DIR:-}/bin/${_BIN_NAME:-}"
            else
              _RESOLVED_PATH=$(find "${_TOOL_DIR:-}/bin" -name "${_BIN_NAME:-}*" -type f 2>/dev/null | head -n 1)
            fi
          fi
        fi
      fi
    fi

    if [ -z "${_RESOLVED_PATH:-}" ]; then
      log_error "✗ ${_BIN_NAME:-} not found via any resolution method"
      log_error "   Tried: unirtm which, command -v, unirtm where, directory search"
      return 1
    fi
    log_debug "✓ Resolved via fallback: ${_RESOLVED_PATH:-}"
  else
    log_debug "✓ Resolved via unirtm which: ${_RESOLVED_PATH:-}"
  fi

  # Step 3: Check if binary exists
  log_debug "Step 3/5: Checking binary existence..."
  if [ ! -f "${_RESOLVED_PATH:-}" ]; then
    log_error "✗ Binary does not exist at: ${_RESOLVED_PATH:-}"
    return 1
  fi
  log_debug "✓ Binary exists"

  # Step 4: Check if binary is executable (skip on Windows)
  log_debug "Step 4/5: Checking executability..."
  log_debug "OS detection: _G_OS=${_G_OS:-<unset>}, _G_UNAME=${_G_UNAME:-<unset>}"

  # Windows doesn't use Unix executable permissions
  # .exe, .cmd, .bat files are executable by extension
  if [ "${_G_OS:-}" != "windows" ]; then
    if [ ! -x "${_RESOLVED_PATH:-}" ]; then
      log_error "✗ ${_RESOLVED_PATH:-} is not executable"
      return 1
    fi
  fi

  # On Windows, check if it's a known executable extension
  if [ "${_G_OS:-}" = "windows" ]; then
    case "${_RESOLVED_PATH:-}" in
    *.exe | *.cmd | *.bat | *.ps1)
      log_debug "✓ Windows executable (by extension)"
      ;;
    *)
      # On Windows, files without extension might still be executable
      # Let the smoke test determine if it works
      log_debug "⚠ Unknown extension on Windows, will verify via smoke test"
      ;;
    esac
  else
    log_debug "✓ Executable"
  fi

  # Step 5: Run smoke test
  log_debug "Step 5/5: Running smoke test..."

  # For unirtm shims, we need to use unirtm exec instead of direct execution
  local _SMOKE_CMD
  local _SMOKE_OUTPUT
  local _SMOKE_EXIT=0

  if echo "${_RESOLVED_PATH:-}" | grep -q "/unirtm/shims/"; then
    # It's a unirtm shim, use unirtm exec
    _SMOKE_CMD="unirtm exec ${_PROVIDER:-} -- ${_BIN_NAME:-} ${_VERSION_FLAG:-}"
    log_debug "Using unirtm exec for shim: ${_SMOKE_CMD:-}"
    _SMOKE_OUTPUT=$(run_with_timeout_robust 5 sh -c "${_SMOKE_CMD:-}" 2>&1) || _SMOKE_EXIT=$?
  else
    # Direct binary execution
    _SMOKE_OUTPUT=$(run_with_timeout_robust 5 "${_RESOLVED_PATH:-}" "${_VERSION_FLAG:-}" 2>&1) || _SMOKE_EXIT=$?
  fi

  # Check if smoke test passed
  # Success criteria: exit code 0 OR output contains version-like pattern
  if [ "${_SMOKE_EXIT:-0}" -ne 0 ]; then
    # Exit code non-zero, check if output looks like a version
    if echo "${_SMOKE_OUTPUT:-}" | grep -qE '[0-9]+\.[0-9]+'; then
      log_debug "✓ Smoke test passed (version detected in output despite non-zero exit)"
    else
      log_error "✗ ${_BIN_NAME:-} failed smoke test (${_VERSION_FLAG:-})"
      log_error "   Exit code: ${_SMOKE_EXIT:-}"
      log_error "   Smoke test output: ${_SMOKE_OUTPUT:-<empty>}"
      return 1
    fi
  else
    log_debug "✓ Smoke test passed"
  fi

  log_debug "=== ✓ ${_DISPLAY_NAME:-} fully verified ==="
  return 0
}

# Purpose: Safe tool installation with binary-first verification.
# This function implements the correct detection logic:
#   1. Check if binary exists and is executable (FIRST)
#   2. Check version ONLY if binary exists
#   3. Install if needed
#   4. Verify installation with aggressive cache refresh
#
# This prevents false positives from stale unirtm cache (GitHub Actions cache).
#
# Parameters:
#   $1 - Binary name (e.g., "shfmt")
#   $2 - Provider (e.g., "github:mvdan/sh")
#   $3 - Display name (e.g., "Shfmt")
#   $4 - Version flag (e.g., "--version") [optional, defaults to "--version"]
#   $5 - Skip file check (0=check files, 1=skip) [optional, defaults to 0]
#   $6 - File patterns (e.g., "*.sh *.bash") [optional, only used if skip_file_check=0]
#   $7 - File directory (e.g., ".github/workflows") [optional, defaults to ""]
#
# Returns:
#   0 - Tool installed and verified successfully
#   1 - Installation or verification failed
#
# Examples:
#   install_tool_safe "shfmt" "github:mvdan/sh" "Shfmt" "--version" 0 "*.sh *.bash"
#   install_tool_safe "hadolint" "github:hadolint/hadolint" "Hadolint" "--version" 0 "Dockerfile*"
#   install_tool_safe "gitleaks" "github:gitleaks/gitleaks" "Gitleaks" "version" 1
install_tool_safe() {
  local _BIN_NAME="${1:-}"
  local _PROVIDER="${2:-}"
  local _DISPLAY_NAME="${3:-Tool}"
  local _VERSION_FLAG="${4:---version}"
  local _SKIP_FILE_CHECK="${5:-0}"
  local _FILE_PATTERNS="${6:-}"
  local _FILE_DIR="${7:-}"

  [ -z "${_BIN_NAME:-}" ] || [ -z "${_PROVIDER:-}" ] && return 1

  local _T0
  _T0=$(date +%s)

  # Get version from provider
  # CRITICAL: For tools with table syntax in .unirtm.toml like:
  #   "github:foo/bar" = { version = "1.0.0", bin = "foo" }
  # We need to extract just the version number, not the whole table
  local _VERSION
  _VERSION=$(get_unirtm_tool_version "${_PROVIDER:-}")

  # If version looks like a TOML table (contains "version ="), extract just the version
  # Use [^"]* instead of [^"]+ for better compatibility (handles empty strings)
  # Use [[:space:]]* for POSIX compatibility across all platforms
  if echo "${_VERSION:-}" | grep -q "version[[:space:]]*="; then
    _VERSION=$(echo "${_VERSION:-}" | sed -E 's/.*version[[:space:]]*=[[:space:]]*"([^"]*)".*/\1/')
  fi

  log_info "=== install_tool_safe: ${_DISPLAY_NAME:-} ==="
  log_info "Binary: ${_BIN_NAME:-}, Provider: ${_PROVIDER:-}, Version: ${_VERSION:-}"
  log_info "CI: $(is_ci_env && echo YES || echo NO)"

  # File detection (skip in CI or if skip_file_check=1)
  if [ "${_SKIP_FILE_CHECK:-0}" -eq 0 ] && ! is_ci_env; then
    if ! has_lang_files "${_FILE_DIR:-}" "${_FILE_PATTERNS:-}"; then
      log_info "⏭️  Skipping ${_DISPLAY_NAME:-}: No matching files detected"
      log_summary "Base" "${_DISPLAY_NAME:-}" "⏭️ Skipped" "-" "0"
      return 0
    fi
  fi

  # CRITICAL: In CI, refresh unirtm cache to avoid stale data from GitHub Actions cache
  if is_ci_env; then
    log_info "CI detected: Refreshing unirtm cache to avoid stale data"
    refresh_unirtm_cache
  fi

  # CRITICAL: Resolve actual binary name from unirtm installation
  # For tools with platform-specific binaries (e.g., ec-linux-amd64), we need to find the real name
  local _ACTUAL_BIN="${_BIN_NAME:-}"
  local _INSTALL_DIR
  _INSTALL_DIR=$(unirtm where "${_PROVIDER:-}" 2>/dev/null) || _INSTALL_DIR=""

  if [ -n "${_INSTALL_DIR:-}" ]; then
    # Try exact match first in bin/ directory
    if [ -d "${_INSTALL_DIR:-}/bin" ] && [ -f "${_INSTALL_DIR:-}/bin/${_BIN_NAME:-}" ]; then
      _ACTUAL_BIN="${_BIN_NAME:-}"
    # Try pattern match in bin/ directory (e.g., ec-* for editorconfig-checker)
    elif [ -d "${_INSTALL_DIR:-}/bin" ]; then
      local _FOUND_BIN
      _FOUND_BIN=$(find "${_INSTALL_DIR:-}/bin" -maxdepth 1 -name "${_BIN_NAME:-}*" -type f -executable 2>/dev/null | head -n 1)
      if [ -n "${_FOUND_BIN:-}" ]; then
        _ACTUAL_BIN=$(basename "${_FOUND_BIN:-}")
        log_info "Resolved actual binary name: ${_ACTUAL_BIN:-} (from ${_BIN_NAME:-})"
      fi
    # Try pattern match in root directory (some tools install directly to root, e.g., shfmt_v3.13.1)
    else
      local _FOUND_BIN
      _FOUND_BIN=$(find "${_INSTALL_DIR:-}" -maxdepth 1 -name "${_BIN_NAME:-}*" -type f -executable 2>/dev/null | head -n 1)
      if [ -n "${_FOUND_BIN:-}" ]; then
        _ACTUAL_BIN=$(basename "${_FOUND_BIN:-}")
        log_info "Resolved actual binary name: ${_ACTUAL_BIN:-} (from ${_BIN_NAME:-}, in root dir)"
      fi
    fi
  fi

  # Step 1: Check if binary exists and works (FIRST, before version check)
  log_info "Step 1: Checking if ${_ACTUAL_BIN:-} binary exists and is executable"
  local _BINARY_EXISTS=0
  if verify_binary_exists "${_ACTUAL_BIN:-}" "${_VERSION_FLAG:-}"; then
    log_info "Step 1: ✓ Binary exists and is executable"
    _BINARY_EXISTS=1
  else
    log_info "Step 1: ✗ Binary not found or not executable"
    _BINARY_EXISTS=0
  fi

  # Step 2: Check version ONLY if binary exists
  local _CUR_VER="-"
  local _REQ_VER="${_VERSION:-}"
  log_info "Step 2: Required version: ${_REQ_VER:-<none>}"

  if [ "${_BINARY_EXISTS:-0}" -eq 1 ]; then
    _CUR_VER=$(get_version "${_ACTUAL_BIN:-}")
    log_info "Step 2: Current version (binary exists): ${_CUR_VER:-<none>}"
  else
    log_info "Step 2: Skipping version check (binary doesn't exist)"
  fi

  # Step 3: Determine if installation is needed
  log_info "Step 3: Determining if installation is needed"
  local _NEEDS_INSTALL=0

  if [ "${_BINARY_EXISTS:-0}" -eq 0 ]; then
    log_info "Step 3: Binary doesn't exist → INSTALL NEEDED"
    _NEEDS_INSTALL=1
  elif [ "${_CUR_VER:-}" = "-" ] || [ -z "${_CUR_VER:-}" ]; then
    log_info "Step 3: No version detected (despite binary existing) → INSTALL NEEDED"
    _NEEDS_INSTALL=1
  elif ! is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_info "Step 3: Version mismatch (${_CUR_VER:-} != ${_REQ_VER:-}) → INSTALL NEEDED"
    _NEEDS_INSTALL=1
  else
    log_info "Step 3: Binary exists + version matches → NO INSTALL NEEDED"
    log_summary "Base" "${_DISPLAY_NAME:-}" "✅ Exists" "${_CUR_VER:-}" "0"
    return 0
  fi

  # Step 4: Clean up if binary exists but needs reinstall
  if [ "${_BINARY_EXISTS:-0}" -eq 1 ] && [ "${_NEEDS_INSTALL:-0}" -eq 1 ]; then
    log_warn "Step 4: Binary exists but needs reinstall - cleaning up"
    unirtm uninstall "${_PROVIDER:-}" 2>/dev/null || true

    # CRITICAL: Aggressive cache refresh after uninstall
    # This is essential for version changes (e.g., taplo 0.7.0 -> 0.10.0)
    refresh_unirtm_cache
    unirtm reshim 2>/dev/null || true

    # Wait for filesystem sync
    sleep 1
  fi

  # Step 5: Install
  log_info "Step 5: Installing ${_PROVIDER:-}@${_VERSION:-}"
  _log_setup "${_DISPLAY_NAME:-}" "${_PROVIDER:-}"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Base" "${_DISPLAY_NAME:-}" '⚖️ Previewed' "-" '0'
    return 0
  fi

  local _STAT="✅ unirtm"
  if ! "${_G_UNIRTM_BIN:-unirtm}" install "${_PROVIDER:-}@${_VERSION:-}"; then
    _STAT="❌ Failed"
    log_error "Step 5: unirtm install FAILED"
    log_summary "Base" "${_DISPLAY_NAME:-}" "${_STAT:-}" "-" "$(($(date +%s) - _T0))"
    if is_ci_env; then
      return 1
    else
      return 0
    fi
  fi

  log_info "Step 5: unirtm install succeeded"

  # Step 6: Post-install verification with aggressive cache refresh
  log_info "Step 6: Post-install verification"
  unirtm reshim 2>/dev/null || true

  # CRITICAL: Refresh unirtm cache after installation to ensure get_version sees new binary
  refresh_unirtm_cache

  # Wait for filesystem sync (especially important in CI with network filesystems)
  sleep 2

  # CRITICAL: Re-resolve actual binary name after installation
  log_info "Step 6: Re-resolving binary name after installation"
  _INSTALL_DIR=$(unirtm where "${_PROVIDER:-}" 2>/dev/null) || _INSTALL_DIR=""
  log_info "Step 6: Install dir = ${_INSTALL_DIR:-<empty>}"

  if [ -n "${_INSTALL_DIR:-}" ]; then
    log_info "Step 6: Resolving binary name in ${_INSTALL_DIR:-}"
    local _FOUND_BIN=""

    # Strategy 1: Try exact match in bin/ directory
    if [ -d "${_INSTALL_DIR:-}/bin" ] && [ -f "${_INSTALL_DIR:-}/bin/${_BIN_NAME:-}" ]; then
      _ACTUAL_BIN="${_BIN_NAME:-}"
      log_info "Step 6: Found exact match in bin/"
    else
      # Strategy 2: Try pattern match in bin/ directory
      if [ -d "${_INSTALL_DIR:-}/bin" ]; then
        log_info "Step 6: Searching for ${_BIN_NAME:-}* in bin/"
        _FOUND_BIN=$(find "${_INSTALL_DIR:-}/bin" -maxdepth 1 -name "${_BIN_NAME:-}*" -type f 2>/dev/null | head -n 1)
        log_info "Step 6: find result: ${_FOUND_BIN:-<empty>}"
        if [ -n "${_FOUND_BIN:-}" ] && [ -x "${_FOUND_BIN:-}" ]; then
          _ACTUAL_BIN=$(basename "${_FOUND_BIN:-}")
          log_info "Step 6: Resolved actual binary name: ${_ACTUAL_BIN:-} (pattern match in bin/)"
        fi
      fi

      # Strategy 3: If still not found, try root directory
      if [ -z "${_FOUND_BIN:-}" ]; then
        log_info "Step 6: Searching for ${_BIN_NAME:-}* in root dir"
        _FOUND_BIN=$(find "${_INSTALL_DIR:-}" -maxdepth 1 -name "${_BIN_NAME:-}*" -type f 2>/dev/null | head -n 1)
        log_info "Step 6: find result: ${_FOUND_BIN:-<empty>}"
        if [ -n "${_FOUND_BIN:-}" ] && [ -x "${_FOUND_BIN:-}" ]; then
          _ACTUAL_BIN=$(basename "${_FOUND_BIN:-}")
          log_info "Step 6: Resolved actual binary name: ${_ACTUAL_BIN:-} (in root dir)"
        fi
      fi
    fi

    log_info "Step 6: Final _ACTUAL_BIN = ${_ACTUAL_BIN:-}"
  else
    log_warn "Step 6: Install dir not found, cannot resolve binary name"
  fi

  # Step 6a: Verify binary now exists
  log_info "Step 6a: Verifying binary existence"

  # On Windows, if the binary file exists in install dir but doesn't have .exe extension,
  # skip command -v check (unirtm shims may not work for non-.exe files)
  local _SKIP_CMD_CHECK=0
  if [ "${_G_OS:-}" = "windows" ] && [ -n "${_INSTALL_DIR:-}" ]; then
    if [ -f "${_INSTALL_DIR:-}/${_ACTUAL_BIN:-}" ] || [ -f "${_INSTALL_DIR:-}/bin/${_ACTUAL_BIN:-}" ]; then
      case "${_ACTUAL_BIN:-}" in
      *.exe | *.cmd | *.bat | *.ps1)
        # Has Windows executable extension, use normal check
        _SKIP_CMD_CHECK=0
        ;;
      *)
        # No Windows extension, skip command -v check
        log_info "Step 6a: Windows binary without .exe extension, skipping command -v check"
        _SKIP_CMD_CHECK=1
        ;;
      esac
    fi
  fi

  # CRITICAL: For unirtm bins (tools installed via unirtm but not through backends),
  # unirtm which may fail even though the binary exists. Check install dir directly.
  if [ -n "${_INSTALL_DIR:-}" ]; then
    local _BIN_FILE=""
    if [ -f "${_INSTALL_DIR:-}/bin/${_ACTUAL_BIN:-}" ]; then
      _BIN_FILE="${_INSTALL_DIR:-}/bin/${_ACTUAL_BIN:-}"
    elif [ -f "${_INSTALL_DIR:-}/${_ACTUAL_BIN:-}" ]; then
      _BIN_FILE="${_INSTALL_DIR:-}/${_ACTUAL_BIN:-}"
    fi

    if [ -n "${_BIN_FILE:-}" ] && [ -x "${_BIN_FILE:-}" ]; then
      log_info "Step 6a: ✓ Binary exists in install dir and is executable"
      _SKIP_CMD_CHECK=1
    fi
  fi

  if [ "${_SKIP_CMD_CHECK:-0}" -eq 0 ]; then
    if ! verify_binary_exists "${_ACTUAL_BIN:-}" "${_VERSION_FLAG:-}"; then
      log_error "Step 6a: Binary still not found after installation!"
      log_error "Debugging info:"
      log_error "  - Logical name: ${_BIN_NAME:-}"
      log_error "  - Actual name: ${_ACTUAL_BIN:-}"
      log_error "  - Install dir: ${_INSTALL_DIR:-}"
      log_error "  - PATH: ${PATH:-}"
      log_error "  - command -v ${_ACTUAL_BIN:-}: $(command -v "${_ACTUAL_BIN:-}" 2>&1 || echo 'NOT FOUND')"
      log_error "  - unirtm which ${_ACTUAL_BIN:-}: $(${_G_UNIRTM_BIN:-unirtm} which "${_ACTUAL_BIN:-}" 2>&1 || echo 'NOT FOUND')"
      log_error "  - unirtm where ${_PROVIDER:-}: $(${_G_UNIRTM_BIN:-unirtm} where "${_PROVIDER:-}" 2>&1 || echo 'NOT FOUND')"
      if [ -n "${_INSTALL_DIR:-}" ] && [ -d "${_INSTALL_DIR:-}/bin" ]; then
        log_error "  - Binaries in install dir: $(ls -la "${_INSTALL_DIR:-}/bin" 2>&1 || echo 'FAILED')"
      fi
      if [ -n "${_INSTALL_DIR:-}" ] && [ -d "${_INSTALL_DIR:-}" ]; then
        log_error "  - Install dir contents: $(ls -la "${_INSTALL_DIR:-}" 2>&1 || echo 'FAILED')"
      fi
      log_summary "Base" "${_DISPLAY_NAME:-}" "❌ Not Found" "-" "$(($(date +%s) - _T0))"
      return 1
    fi
  fi
  log_info "Step 6a: ✓ Binary exists after installation"

  # Step 6b: Atomic verification (comprehensive check in CI)
  if is_ci_env; then
    log_info "Step 6b: Running atomic verification"
    if ! verify_tool_atomic "${_ACTUAL_BIN:-}" "${_PROVIDER:-}" "${_DISPLAY_NAME:-}" "${_VERSION_FLAG:-}"; then
      log_error "Step 6b: Atomic verification FAILED"
      log_summary "Base" "${_DISPLAY_NAME:-}" "❌ Not Usable" "-" "$(($(date +%s) - _T0))"
      return 1
    fi
    log_info "Step 6b: ✓ Atomic verification succeeded"
  fi

  log_summary "Base" "${_DISPLAY_NAME:-}" "${_STAT:-}" "$(get_version "${_ACTUAL_BIN:-}")" "$(($(date +%s) - _T0))"
  log_info "=== install_tool_safe: ${_DISPLAY_NAME:-} completed successfully ==="
  return 0
}


# Purpose: Installs git hooks using pre-commit.
# Delegate: Managed via pipx (pre-commit).
# Examples:
#   install_runtime_hooks
install_runtime_hooks() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would active pre-commit hooks."
    return 0
  fi

  if [ ! -d ".git" ]; then
    log_debug "Not a git repository. Skipping hook installation."
    return 0
  fi

  local _PRE_COMMIT_BIN
  _PRE_COMMIT_BIN=$(resolve_bin "pre-commit") || true
  if [ -n "${_PRE_COMMIT_BIN:-}" ]; then
    log_info "Running pre-commit install..."
    run_quiet "${_PRE_COMMIT_BIN:-}" install
  else
    log_warn "pre-commit binary not found. Skipping hook installation."
  fi
}

# Purpose: Synchronizes Node.js lockfile safely.
sync_node_lockfile() {
  local _SYNC_DIR="${1:-.}"
  local _OLDPWD_SYNC
  _OLDPWD_SYNC="$(pwd)"

  cd "${_SYNC_DIR:-}" || return 1

  if [ -f "package.json" ]; then
    case "${NPM:-}" in
    pnpm)
      log_info "Syncing Node.js lockfile (pnpm)..."
      run_quiet pnpm install --no-frozen-lockfile
      ;;
    yarn)
      log_info "Syncing Node.js lockfile (yarn)..."
      run_quiet yarn install --no-immutable
      ;;
    bun)
      log_info "Syncing Node.js lockfile (bun)..."
      run_quiet bun install
      ;;
    *)
      log_info "Syncing Node.js lockfile (npm)..."
      run_quiet npm install --package-lock-only
      ;;
    esac
  fi

  cd "${_OLDPWD_SYNC:-}" || exit 1
}

# Purpose: Executes an npm/pnpm script with infinite-recursion detection.
# Params:
#   $1 - Name of the npm script (e.g., "test")
# Examples:
#   run_npm_script "test"
run_npm_script() {
  local _SCRIPT_NAME_NPM="${1:-}"
  shift # Capture remaining arguments
  local _EXTRA_ARGS_NPM="$*"
  local _CURRENT_BASENAME_NPM
  _CURRENT_BASENAME_NPM="$(basename "${0:-}")"

  if [ -f "package.json" ]; then
    # 1. Manager Detection & Guard
    _NODE_MGR="${NPM:-}"

    if ! command -v "${_NODE_MGR:-}" >/dev/null 2>&1; then
      log_warn "Warning: ${_NODE_MGR:-} command not found and no fallback available. Skipping Node.js task: ${_SCRIPT_NAME_NPM:-}."
      return 0
    fi

    # 2. Package.json Integrity check (Avoid empty/invalid JSON crashes)
    if [ ! -s "package.json" ] || ! grep -q "{" "package.json"; then
      log_debug "package.json is empty or invalid. Skipping Node.js task: ${_SCRIPT_NAME_NPM:-}."
      return 0
    fi

    # 3. Check for script in package.json
    local _CMD_NPM
    _CMD_NPM=$(grep "\"${_SCRIPT_NAME_NPM:-}\":" "package.json" | sed "s/.*\"${_SCRIPT_NAME_NPM:-}\":[[:space:]]*\"//;s/\".*//" || true)

    if [ -n "${_CMD_NPM:-}" ]; then
      # Avoid infinite loop if the command points back to this script
      if echo "${_CMD_NPM:-}" | grep -q "${_CURRENT_BASENAME_NPM:-}"; then
        log_debug "Node script '${_SCRIPT_NAME_NPM:-}' is a self-reference to '${_CURRENT_BASENAME_NPM:-}'. Skipping."
        return 0
      fi
      log_info "── Running Node.js script: ${_NODE_MGR:-} ${_SCRIPT_NAME_NPM:-} ${_EXTRA_ARGS_NPM:-} ──"
      # shellcheck disable=SC2086
      "${_NODE_MGR:-}" run "${_SCRIPT_NAME_NPM:-}" ${_EXTRA_ARGS_NPM:-}
    elif [ "${_SCRIPT_NAME_NPM:-}" = "install" ] || [ "${_SCRIPT_NAME_NPM:-}" = "update" ]; then
      # 4. Special Fallback for native commands if not defined in package.json scripts
      log_info "── Node.js standard command: ${_NODE_MGR:-} ${_SCRIPT_NAME_NPM:-} ${_EXTRA_ARGS_NPM:-} ──"
      # shellcheck disable=SC2086
      run_quiet "${_NODE_MGR:-}" "${_SCRIPT_NAME_NPM:-}" ${_EXTRA_ARGS_NPM:-}
    fi
  fi
  return 0
}

# Purpose: Initializes the execution summary table file.
# Params:
#   $1 - Header title (e.g., "Execution Summary")
# Examples:
#   init_summary_table "Setup Execution Summary"
init_summary_table() {
  local _TITLE_TABLE="${1:-Execution Summary}"

  # Sentinel to prevent duplicate headers in the same summary stream
  local _SENTINEL_TABLE
  _SENTINEL_TABLE="_SUMMARY_TABLE_INITIALIZED_$(echo "${_TITLE_TABLE:-}" | tr ' ' '_')"
  if [ "$(eval echo "\${${_SENTINEL_TABLE:-}:-}")" = "true" ]; then
    return 0
  fi

  # Ensure the summary file exists
  touch "${CI_STEP_SUMMARY:-}"

  {
    printf "### %s\n\n" "${_TITLE_TABLE:-}"
    printf "| Category | Module | Status | Version | Time |\n"
    printf "| :--- | :--- | :--- | :--- | :--- |\n"
  } >>"${CI_STEP_SUMMARY:-}"

  eval "export ${_SENTINEL_TABLE:-}=true"
}

# Purpose: Finalizes the summary table. (Deprecated: Writes are now direct).
finalize_summary_table() {
  if [ "${_IS_TOP_LEVEL:-}" = "true" ] && [ -f "${CI_STEP_SUMMARY:-}" ]; then
    # In CI, the platform handles the file. In local dev, we print it to the console.
    if [ "${CI:-}" != "true" ] && [ "${GITHUB_ACTIONS:-}" != "true" ] && [ "${VERBOSE:-1}" -ge 1 ]; then
      cat "${CI_STEP_SUMMARY:-}"
    fi
  fi
  log_debug "Summary table finalized in $CI_STEP_SUMMARY"
}

# Purpose: Checks if the current tool version matches the required version (prefix match).
# Params:
#   $1 - Current version (detected)
#   $2 - Required version (from registry/.unirtm.toml)
# Returns:
#   0 - Match
#   1 - Mismatch
is_version_match() {
  local _CUR_V_M="${1:-}"
  local _REQ_V_M="${2:-}"
  [ "${_CUR_V_M:- -}" = "-" ] && return 1
  [ -z "${_REQ_V_M:-}" ] && return 1
  [ "${_REQ_V_M:-}" = "latest" ] && return 0

  # Normalize versions: strip leading 'v' or 'V' prefix (Robustness Principle - input tolerance)
  _CUR_V_M=$(echo "${_CUR_V_M:-}" | sed 's/^[vV]//')
  _REQ_V_M=$(echo "${_REQ_V_M:-}" | sed 's/^[vV]//')

  # Use prefix match to handle diffs like 3.12.0.2 (pkg) vs 3.12.0 (binary)
  case "${_REQ_V_M:-}" in "${_CUR_V_M:-}"*) return 0 ;; esac
  return 1
}

# ── Extension Modules Sourcing ──
# Note: Dynamic loading of language-specific setup modules has been moved to setup.sh
# to optimize performance for non-setup scripts.

# ── 🛣️ CI Persistence (Cross-Platform) ──────────────────────────────────────

# Purpose: Get the CI-specific PATH persistence file location
# Returns: Path to the file where PATH additions should be written, or empty if not supported
# Note: Returns Unix-style paths even on Windows (Git Bash compatibility)
_get_ci_path_file() {
  local _ci_file=""

  case "$(detect_ci_platform)" in
  github-actions | forgejo-actions | gitea-actions)
    # GitHub Actions and compatible platforms use GITHUB_PATH
    # Already in correct format for the platform
    _ci_file="${GITHUB_PATH:-}"
    ;;
  gitlab-ci)
    # GitLab CI: No direct PATH persistence mechanism
    # PATH modifications only last within the current job
    # We can use a custom file and source it in subsequent scripts
    _ci_file="${CI_PROJECT_DIR:-.}/.ci_path_cache"
    ;;
  drone | woodpecker)
    # Drone/Woodpecker: Similar to GitLab, use custom file
    _ci_file="${DRONE_WORKSPACE:-${CI_WORKSPACE:-.}}/.ci_path_cache"
    ;;
  circleci)
    # CircleCI: Use custom file in workspace
    _ci_file="${CIRCLE_WORKING_DIRECTORY:-.}/.ci_path_cache"
    ;;
  azure-pipelines)
    # Azure Pipelines: Use custom file
    _ci_file="${BUILD_SOURCESDIRECTORY:-.}/.ci_path_cache"
    ;;
  jenkins)
    # Jenkins: Use custom file in workspace
    _ci_file="${WORKSPACE:-.}/.ci_path_cache"
    ;;
  travis)
    # Travis CI: Use custom file
    _ci_file="${TRAVIS_BUILD_DIR:-.}/.ci_path_cache"
    ;;
  *)
    # Unknown or local: no persistence
    echo ""
    return 0
    ;;
  esac

  # Normalize path for Windows Git Bash if needed
  # GITHUB_PATH is already in correct format, skip conversion
  if [ "${_G_OS:-}" = "windows" ] && [ -n "${_ci_file:-}" ]; then
    case "$(detect_ci_platform)" in
    github-actions | forgejo-actions | gitea-actions)
      # GITHUB_PATH is already correct, don't convert
      ;;
    *)
      # For other CI platforms on Windows, ensure Unix-style path
      # Convert backslashes to forward slashes if present
      _ci_file=$(echo "${_ci_file:-}" | sed 's/\\/\//g')
      ;;
    esac
  fi

  echo "${_ci_file:-}"
}

# Purpose: Add a path to CI persistence file for future steps/jobs
# Params: $1 - Path to add
# Security: Only stores directory paths (no credentials/secrets)
_persist_path_to_ci() {
  local _path_to_add="${1:-}"
  [ -z "${_path_to_add:-}" ] && return 0

  local _ci_path_file
  _ci_path_file=$(_get_ci_path_file)
  [ -z "${_ci_path_file:-}" ] && return 0

  # Security: Validate input is a path (no special characters that could inject commands)
  case "${_path_to_add:-}" in
  *\$* | *\`* | *\;* | *\|* | *\&*)
    log_warn "Security: Rejected suspicious path containing special characters: ${_path_to_add:-}"
    return 1
    ;;
  esac

  # Windows PATH Handling: Write both Unix and Windows formats for maximum compatibility
  # Strategy: Redundancy over missing paths - write both formats to ensure tools are found
  if [ "${_G_OS:-}" = "windows" ] && [ "$(detect_ci_platform)" = "github-actions" ]; then
    # Idempotent check: Skip if Unix-style path already present
    if [ -f "${_ci_path_file:-}" ] && grep -qxF "${_path_to_add:-}" "${_ci_path_file:-}" 2>/dev/null; then
      echo "    [INFO] Unix path already in CI cache: ${_path_to_add:-}" >&2
      return 0
    fi

    # Security: Set restrictive permissions on first write
    if [ ! -f "${_ci_path_file:-}" ]; then
      touch "${_ci_path_file:-}"
      chmod 600 "${_ci_path_file:-}" 2>/dev/null || true
      echo "    [NEW] Created CI path cache: ${_ci_path_file:-}" >&2
    fi

    # Write Unix-style path first (Git Bash native format)
    echo "${_path_to_add:-}" >>"${_ci_path_file:-}"
    echo "    [OK] Wrote Unix path: ${_path_to_add:-}" >&2

    # Also write Windows-style path for compatibility with other shells
    if command -v cygpath >/dev/null 2>&1; then
      local _windows_path
      _windows_path=$(cygpath -w "${_path_to_add:-}" 2>/dev/null) || _windows_path=""
      if [ -n "${_windows_path:-}" ] && [ "${_windows_path:-}" != "${_path_to_add:-}" ]; then
        # Check if Windows path already exists
        if ! grep -qxF "${_windows_path:-}" "${_ci_path_file:-}" 2>/dev/null; then
          echo "${_windows_path:-}" >>"${_ci_path_file:-}"
          echo "    [OK] Wrote Windows path: ${_windows_path:-}" >&2
        fi
      fi
    fi

    # CRITICAL: Also update current shell's PATH immediately
    # GitHub Actions only applies GITHUB_PATH changes to NEXT step, not current step
    case ":${PATH:-}:" in
    *":${_path_to_add:-}:"*) ;;
    *)
      export PATH="${_path_to_add:-}:${PATH:-}"
      echo "    [OK] Added to current shell PATH: ${_path_to_add:-}" >&2
      ;;
    esac
  else
    # Non-Windows or non-GitHub Actions: Standard behavior
    # Idempotent: Don't add if already present
    if [ -f "${_ci_path_file:-}" ] && grep -qxF "${_path_to_add:-}" "${_ci_path_file:-}" 2>/dev/null; then
      echo "    [INFO] Path already in CI cache: ${_path_to_add:-}" >&2
      return 0
    fi

    # Security: Set restrictive permissions on first write
    if [ ! -f "${_ci_path_file:-}" ]; then
      touch "${_ci_path_file:-}"
      chmod 600 "${_ci_path_file:-}" 2>/dev/null || true
      echo "    [NEW] Created CI path cache: ${_ci_path_file:-}" >&2
    fi

    echo "${_path_to_add:-}" >>"${_ci_path_file:-}"
    echo "    [OK] Wrote to CI path cache: ${_path_to_add:-}" >&2

    # CRITICAL: Also update current shell's PATH immediately
    # GitHub Actions only applies GITHUB_PATH changes to NEXT step, not current step
    case ":${PATH:-}:" in
    *":${_path_to_add:-}:"*) ;;
    *)
      export PATH="${_path_to_add:-}:${PATH:-}"
      echo "    [OK] Added to current shell PATH: ${_path_to_add:-}" >&2
      ;;
    esac
  fi
}

# Purpose: Read CI persistence file and sync paths to current shell
_sync_ci_paths_to_shell() {
  local _ci_path_file
  _ci_path_file=$(_get_ci_path_file)

  [ -z "${_ci_path_file:-}" ] && return 0
  [ ! -f "${_ci_path_file:-}" ] && return 0

  log_debug "Reading CI path cache and syncing to current shell: ${_ci_path_file:-}"

  while IFS= read -r _ci_path || [ -n "${_ci_path:-}" ]; do
    [ -z "${_ci_path:-}" ] && continue
    _ci_path=$(echo "${_ci_path:-}" | tr -d '\r\n' | sed 's/[[:space:]]*$//')
    [ -z "${_ci_path:-}" ] && continue

    case ":${PATH:-}:" in
    *":${_ci_path:-}:"*) ;;
    *)
      export PATH="${_ci_path:-}:${PATH:-}"
      log_debug "Added to PATH from CI cache: ${_ci_path:-}"
      ;;
    esac
  done <"${_ci_path_file:-}"
}

# Step 1: Read existing CI path cache and sync to current shell
# This ensures tools installed by previous steps are available in current shell
if is_ci_env && [ -z "${_G_CI_PATH_READ:-}" ]; then
  _sync_ci_paths_to_shell
  export _G_CI_PATH_READ=true
fi

# Step 2: Write unirtm paths to CI cache for future steps
if is_ci_env && [ -z "${_G_CI_PATH_SYNCED:-}" ]; then
  # Proactively add unirtm paths to CI persistence
  # Note: Use Unix-style paths even on Windows (GitHub Actions runs in Git Bash)
  _M_BIN_CI="${_G_UNIRTM_BIN_BASE:-}"
  _M_SHIMS_CI="${_G_UNIRTM_SHIMS_BASE:-}"

  # Only persist paths that actually exist and contain files
  if [ -d "$_G_UNIRTM_BIN_BASE" ] && [ -n "$(ls -A "$_G_UNIRTM_BIN_BASE" 2>/dev/null)" ]; then
    _persist_path_to_ci "${_M_BIN_CI:-}"
    log_debug "Persisted unirtm bin to CI: $_M_BIN_CI"
  fi
  if [ -d "$_G_UNIRTM_SHIMS_BASE" ] && [ -n "$(ls -A "$_G_UNIRTM_SHIMS_BASE" 2>/dev/null)" ]; then
    _persist_path_to_ci "${_M_SHIMS_CI:-}"
    log_debug "Persisted unirtm shims to CI: $_M_SHIMS_CI"
  fi

  export _G_CI_PATH_SYNCED=true
fi
