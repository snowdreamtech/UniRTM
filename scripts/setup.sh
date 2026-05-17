#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/setup.sh - Modular Project Setup Engine
#
# Purpose:
#   Facilitates local development and CI/CD JIT toolchain installation.
#   Maintains an isolated, reproducible development environment.
#
# Usage:
#   sh scripts/setup.sh [OPTIONS] [MODULES]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (Network), Rule 04 (Security), Rule 05 (Dependencies), Rule 08 (Dev Env).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Modularized toolchain installation.
#   - Multi-language support (Node, Python, Go, Rust, Java, etc.).
#   - JIT security toolchain (OSV-Scanner, Zizmor).

set -eu

# ── 🎒 Library Sourcing ──────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
# shellcheck source=/dev/null
. "${SCRIPT_DIR:-}/lib/common.sh"
# shellcheck source=/dev/null
. "${SCRIPT_DIR:-}/lib/registry.sh"
# shellcheck source=/dev/null
. "${SCRIPT_DIR:-}/lib/versions.sh"
# shellcheck source=/dev/null
. "${SCRIPT_DIR:-}/lib/github-api-info.sh"

# ── Extension Modules Sourcing ───────────────────────────────────────────────
# Dynamically load all language-specific setup modules.
# shellcheck source=/dev/null
for _lang_mod in "${SCRIPT_DIR:-}/lib/langs"/*.sh; do
  if [ -f "${_lang_mod:-}" ]; then
    # shellcheck disable=SC1090
    . "${_lang_mod:-}"
  fi
done
unset _lang_mod

# ── Configuration ────────────────────────────────────────────────────────────
# Global variables (VENV, PYTHON, etc.) are sourced from common.sh

# Purpose: Displays usage information for the setup engine.
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS] [MODULES]

Modularized Project Setup Script for local development and CI/CD environments.

Options:
  -q, --quiet        Suppress informational output.
  -v, --verbose      Enable verbose/debug output.
  --dry-run          Preview what will be installed without making changes.
  -h, --help         Show this help message.

Modules (default: all):
  base               Setup universal tools (pipx, gitleaks, hooks, etc.)
  node, python, go, rust, java, kotlin, php, ruby, dart, swift, lua, cpp, etc.
  docker, sql, markdown, yaml, openapi, protobuf, security, runners, testing, docs, ai

Environment Variables:
  VENV               Virtualenv directory (default: .venv)
  PYTHON             Python executable (default: python3)
  GITHUB_PROXY       Github proxy URL for asset downloads

EOF
}

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Internal helper to display a consistent setup header with version info.
_log_setup() {
  local _TITLE="${1:-}"
  local _LOOKUP="${2:-}"
  local _VER=""
  [ -n "${_LOOKUP:-}" ] && _VER=$(get_unirtm_tool_version "${_LOOKUP:-}")

  if [ -n "${_VER:-}" ]; then
    log_info "── Setting up ${_TITLE:-} (${_VER:-}) ──"
  else
    log_info "── Setting up ${_TITLE:-} ──"
  fi
}

main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  parse_common_args "$@"

  # ── Concurrency Guard (Lockfile) ──
  # Using project-local lock to allow concurrent setup in different clones/test environments
  local _LOCKFILE="${_G_PROJECT_ROOT:-}/.setup.lock"
  if [ -f "${_LOCKFILE:-}" ]; then
    local _PID
    _PID=$(cat "${_LOCKFILE:-}")
    if ps -p "${_PID:-}" >/dev/null 2>&1; then
      log_error "Setup already in progress (PID: ${_PID:-})."
      log_info "If you are sure no other setup is running, you can:"
      log_info "  1. Kill the process: kill -9 ${_PID:-}"
      log_info "  2. Remove the lock: rm -f ${_LOCKFILE:-}"
      exit 1
    else
      log_warn "Stale lockfile detected (PID: ${_PID:-} is dead). Cleaning up..."
      rm -f "${_LOCKFILE:-}"
    fi
  fi
  echo "$$" >"${_LOCKFILE:-}"
  # shellcheck disable=SC2064
  trap "rm -f ${_LOCKFILE:-}" EXIT INT TERM

  # 3. Network Optimization
  optimize_network

  # Re-extract raw args to avoid flags
  local _RAW_ARGS=""
  local _arg
  for _arg in "$@"; do
    case "${_arg:-}" in
    -q | --quiet | -v | --verbose | --dry-run | -h | --help) ;;
    *) _RAW_ARGS="${_RAW_ARGS:-} ${_arg}" ;;
    esac
  done

  # ── Execution Timing & Summary Management ──
  local _START_TIME_MAIN
  _START_TIME_MAIN=$(date +%s)

  # Unified table initialization (handles sentinels internally)
  init_summary_table "Setup Execution Summary"

  # Initialize Status Legend (Only once per session)
  if [ "${_SETUP_SUMMARY_INITIALIZED:-false}" != "true" ] && ! check_ci_summary "Status Legend:"; then
    {
      cat <<EOF
> **Status Legend:**
> ⚖️ **Previewed**: Running in \`--dry-run\` mode.
> ✅ **Active/Detected/Available**: System/Shell active or Runtime detected.
> ✅ **Installed**: Tool was missing and successfully installed.
> ✅ **Exists**: Tool already exists in \`$VENV/bin\`.
> ✅ **Activated**: Git Hooks successfully attached to \`.git/\`.
> ⏭️ **Skipped/Missing**: Module skipped or required runtime not found.
> ⚠️ **Warning**: Tool exists but version verification failed.
> ❌ **Failed**: An error occurred during installation or setup.

EOF
      # Add Global Environment Detections immediately after the legend
      log_summary "Environment" "System" "✅ Active" "$(uname -s)/$(uname -m)" "0"
      log_summary "Environment" "Shell" "✅ Active" "$(basename "${SHELL:-}")" "0"

      # Detect language runtimes only when corresponding project files exist.
      # This prevents spurious "⚠️ Warning" entries for unrelated runtimes
      # caused by unirtm cache key substring mismatches (e.g. "go" matching "goreleaser").
      for _r in node python; do
        local _v
        _v=$(get_version "${_r:-}")
        if [ "${_v:-}" != "-" ]; then
          log_summary "Runtime" "${_r:-}" "✅ Detected" "${_v:-}" "0"
        fi
      done
      # Only detect Go when go.mod is present (avoids unirtm cache key false-positive on "goreleaser/...")
      if has_lang_files "go.mod" "*.go"; then
        local _vgo
        _vgo=$(get_version "go")
        if [ "${_vgo:-}" != "-" ]; then
          log_summary "Runtime" "go" "✅ Detected" "${_vgo:-}" "0"
        fi
      fi
      # Only detect Rust when Cargo.toml is present
      if has_lang_files "Cargo.toml Cargo.lock" "*.rs"; then
        local _vrust
        _vrust=$(get_version "rust")
        if [ "${_vrust:-}" != "-" ]; then
          log_summary "Runtime" "rust" "✅ Detected" "${_vrust:-}" "0"
        fi
      fi
    } >>"${CI_STEP_SUMMARY:-}"

    # Set master sentinel for subsequent steps in CI
    [ -n "${GITHUB_ENV:-}" ] && echo "_SETUP_SUMMARY_INITIALIZED=true" >>"${GITHUB_ENV:-}"
    export _SETUP_SUMMARY_INITIALIZED=true
  fi

  # Provide table header if not already present in the summary
  if [ "${_SUMMARY_TABLE_HEADER_SENTINEL:-false}" != "true" ] && ! check_ci_summary "| Category | Module | Status |"; then
    {
      printf "| Category | Module | Status | Version | Time |\n"
      printf "| :--- | :--- | :--- | :--- | :--- |\n"
    } >>"${CI_STEP_SUMMARY:-}"
    [ -n "${GITHUB_ENV:-}" ] && echo "_SUMMARY_TABLE_HEADER_SENTINEL=true" >>"${GITHUB_ENV:-}"
    export _SUMMARY_TABLE_HEADER_SENTINEL=true
  fi

  # ── Mode & Module Selection ──
  local _IS_ALL_MODULES=false
  if echo " ${_RAW_ARGS:-} " | grep -q " all "; then
    _IS_ALL_MODULES=true
  fi

  local _MODULES_LIST
  if [ -z "$(echo "${_RAW_ARGS:-}" | tr -d ' ')" ] || [ "${_IS_ALL_MODULES:-}" = "true" ]; then
    # Grouped list for "On-demand" (default) or "All" (explicit)
    # NOTE: yaml and toml are ESSENTIAL modules required for local pre-commit hooks
    # (yamllint, taplo). They are listed after markdown in BASE_LIST to ensure
    # they are always installed by default, never skipped in local development.
    local _BASE_LIST="python node go base shell markdown yaml toml rust java kotlin php ruby dart swift lua cpp terraform solidity perl julia r groovy dotnet zig elixir haskell scala ada assemblyscript ballerina bun clojure crystal deno dlang duckdb elm erlang fortran fpc gleam grain haxe jsonnet kcl lean lisp luau mojo moonbit move nim ocaml odin pkl prolog pulumi racket raku rescript starlark tcl tofu typst vala vcpkg vlang wat"
    # Domain-specific tools: Removed sql, openapi, protobuf, runners (rarely used, install on-demand only)
    local _DOMAIN_LIST="docker security testing docs ai helm k8s terraform terragrunt tofu pulumi"
    _MODULES_LIST="${_BASE_LIST:-} ${_DOMAIN_LIST:-}"
  else
    _MODULES_LIST="${_RAW_ARGS:-}"
  fi

  # ── CI/Local Environment Filtering ──
  # Skip heavyweight tools in local dev unless explicitly requested or 'all' is specified.
  if ! is_ci_env && [ "${_IS_ALL_MODULES:-}" != "true" ] && [ "${SETUP_FORCE_ALL:-0}" -ne 1 ] && [ -z "$(echo "${_RAW_ARGS:-}" | tr -d ' ')" ]; then
    # Skip heavyweight and rarely-needed modules in local dev to prevent
    # long network downloads (especially GitHub-released binaries).
    # NOTE: yaml and toml are NOT included here because they are essential
    # for pre-commit hooks (yamllint, taplo) and must run locally for quality checks.
    # These can still be installed on-demand via: sh scripts/setup.sh <module>
    local _HEAVY_MODULES="markdown security docs testing ai helm k8s terragrunt grain moonbit move luau rescript starlark tcl wat pkl kcl ada assemblyscript ballerina fpc lean lisp racket prolog fortran typst"
    local _SMART_LIST=""
    for _m in ${_MODULES_LIST:-}; do
      case " ${_HEAVY_MODULES:-} " in *" ${_m} "*)
        log_debug "Skipping heavyweight module in local dev: $_m"
        ;;
      *) _SMART_LIST="${_SMART_LIST:-} ${_m}" ;;
      esac
    done
    _MODULES_LIST=${_SMART_LIST:-}
  fi

  # ── Module Skipping & Filtering (SKIP_MODULES) ──
  if [ -n "${SKIP_MODULES:-}" ]; then
    local _NEW_LIST=""
    for _m in ${_MODULES_LIST:-}; do
      case " ${SKIP_MODULES:-} " in *" ${_m} "*)
        log_warn "Skipping module per SKIP_MODULES: $_m"
        log_summary "Skipped" "${_m:-}" "⏭️ Stopped" "-" "0"
        ;;
      *) _NEW_LIST="${_NEW_LIST:-} ${_m}" ;;
      esac
    done
    _MODULES_LIST=${_NEW_LIST:-}
  fi

  # 5. Bootstrap Toolchain Manager
  bootstrap_unirtm || log_warn "Warning: unirtm/unirtm bootstrap failed. Falling back to local tool installation."

  # 6. Toolchain Manager Strategy
  if [ "${DRY_RUN:-0}" -eq 0 ]; then
    export GIT_PROTOCOL=version=2
    export UNIRTM_GIT_ALWAYS_USE_GIX=0

    # Performance Opt: Cache unirtm state once per session
    refresh_unirtm_cache

    if [ "${_IS_ALL_MODULES:-}" = "true" ] && [ "$(uname -s)" != "Windows_NT" ]; then
      log_info "Performing full toolchain synchronization via unirtm..."
      unirtm install
    else
      log_info "Performing on-demand module installation..."
    fi
  fi

  # 7. Execution Loop
  local _cur_grp=""
  for _cur_module in ${_MODULES_LIST:-}; do
    # Visual Grouping Headers
    if [ "${_IS_ALL_MODULES:-}" = "true" ]; then
      case " ${_BASE_LIST:-} " in *" ${_cur_module} "*)
        [ "${_cur_grp:-}" != "base" ] && log_info "── Base/Language Toolsets ──" && _cur_grp="base"
        ;;
      esac
      case " ${_DOMAIN_LIST:-} " in *" ${_cur_module} "*)
        [ "${_cur_grp:-}" != "domain" ] && log_info "── Domain Toolsets ──" && _cur_grp="domain"
        ;;
      esac
    fi

    # ── Force Mode Detection (Architectural Enhancement) ──
    # If a specific module is explicitly requested by the user from the CLI
    # (e.g., 'make setup go'), we set FORCE_SETUP=1. This instructs the
    # language setup functions to skip the 'has_lang_files' check, which
    # is essential for pre-provisioning (like DevContainer building) where
    # tools are installed before source code exists.
    export FORCE_SETUP=0
    if [ "${_IS_ALL_MODULES:-}" = "false" ] && [ -n "$(echo "${_RAW_ARGS:-}" | tr -d ' ')" ]; then
      case " ${_RAW_ARGS:-} " in *" ${_cur_module} "*)
        export FORCE_SETUP=1
        log_debug "Force setup enabled for explicitly requested module: $_cur_module"
        ;;
      esac
    fi

    # Dispatch to modular setup functions
    case $_cur_module in
    base) setup_base ;;
    shell) setup_shell ;;
    toml) setup_toml ;;
    yaml) setup_yaml ;;
    markdown) setup_markdown ;;
    node) setup_node ;;
    python) setup_python ;;
    go) setup_go ;;
    rust) setup_rust ;;
    java) setup_java ;;
    kotlin) setup_kotlin ;;
    php) setup_php ;;
    ruby) setup_ruby ;;
    dart) setup_dart ;;
    swift) setup_swift ;;
    lua) setup_lua ;;
    cpp) setup_cpp ;;
    terraform) setup_terraform ;;
    solidity) setup_solidity ;;
    perl) setup_perl ;;
    julia) setup_julia ;;
    r) setup_r ;;
    groovy) setup_groovy ;;
    dotnet) setup_dotnet ;;
    zig) setup_zig ;;
    elixir) setup_elixir ;;
    haskell) setup_haskell ;;
    scala) setup_scala ;;
    ada) setup_ada ;;
    assemblyscript) setup_assemblyscript ;;
    ballerina) setup_ballerina ;;
    bun) setup_bun ;;
    clojure) setup_clojure ;;
    crystal) setup_crystal ;;
    deno) setup_deno ;;
    dlang) setup_dlang ;;
    duckdb) setup_duckdb ;;
    elm) setup_elm ;;
    erlang) setup_erlang ;;
    fortran) setup_fortran ;;
    fpc) setup_fpc ;;
    gleam) setup_gleam ;;
    grain) setup_grain ;;
    haxe) setup_haxe ;;
    jsonnet) setup_jsonnet ;;
    kcl) setup_kcl ;;
    lean) setup_lean ;;
    lisp) setup_lisp ;;
    luau) setup_luau ;;
    mojo) setup_mojo ;;
    moonbit) setup_moonbit ;;
    move) setup_move ;;
    nim) setup_nim ;;
    ocaml) setup_ocaml ;;
    odin) setup_odin ;;
    pkl) setup_pkl ;;
    prolog) setup_prolog ;;
    pulumi) setup_pulumi ;;
    racket) setup_racket ;;
    raku) setup_raku ;;
    rescript) setup_rescript ;;
    starlark) setup_starlark ;;
    tcl) setup_tcl ;;
    tofu) setup_tofu ;;
    typst) setup_typst ;;
    vala) setup_vala ;;
    vcpkg) setup_vcpkg ;;
    vlang) setup_vlang ;;
    wat) setup_wat ;;
    docker) setup_docker ;;
    sql) setup_sql ;;
    openapi) setup_openapi ;;
    protobuf) setup_protobuf ;;
    security) setup_security ;;
    runners) setup_runners ;;
    testing) setup_testing ;;
    docs) setup_docs ;;
    ai) setup_ai ;;
    helm | k8s) setup_helm ;;
    terragrunt) setup_terragrunt ;;
    # Legacy/Mapping aliases
    hadolint | dockerfile-utils) setup_docker ;;
    sqlfluff) setup_sql ;;
    markdownlint) setup_markdown ;;
    yamllint | dotenv-linter) setup_yaml ;;
    osv-scanner | zizmor | cargo-audit) setup_security ;;
    spectral) setup_openapi ;;
    buf) setup_protobuf ;;
    ghc | stack | cabal) setup_haskell ;;
    v | v-lang) setup_vlang ;;
    kt | kts) setup_kotlin ;;
    py) setup_python ;;
    ts | js) setup_node ;;
    rb) setup_ruby ;;
    pl) setup_perl ;;
    pipx) setup_base ;;
    just | task) setup_runners ;;
    playwright | cypress | vitest | bats | bats-libs) setup_testing ;;
    docusaurus | mkdocs | sphinx) setup_docs ;;
    jupyter | dvc) setup_ai ;;
    cue) setup_cue ;;
    *) log_error "Unknown module: $_cur_module" ;;
    esac
  done

  # ── Final Output Management ──
  if [ "${_IS_TOP_LEVEL:-true}" = "true" ]; then
    local _TOTAL_DUR_MAIN=$(($(date +%s) - _START_TIME_MAIN))
    printf "\n**Total Duration: %ss**\n" "${_TOTAL_DUR_MAIN:-}" >>"${CI_STEP_SUMMARY:-}"

    # Add GitHub API rate limit info
    append_github_api_info

    printf "\n"
    finalize_summary_table
    log_info "\n✨ Setup step complete!"

    # CI PATH Persistence: Ensure unirtm paths are persisted after setup completes
    # This is critical for Windows CI where tools are installed during setup
    if is_ci_env && [ "${DRY_RUN:-0}" -eq 0 ]; then
      log_info "[CI-PATH] Persisting unirtm paths to CI..."
      if [ -d "${_G_UNIRTM_BIN_BASE:-}" ] && [ -n "$(ls -A "${_G_UNIRTM_BIN_BASE:-}" 2>/dev/null)" ]; then
        _persist_path_to_ci "${_G_UNIRTM_BIN_BASE:-}"
        # Use echo to avoid printf issues with Windows paths
        echo "  [OK] Persisted unirtm bin: ${_G_UNIRTM_BIN_BASE:-}" >&2
      else
        echo "  [WARN] unirtm bin directory not found or empty: ${_G_UNIRTM_BIN_BASE:-}" >&2
      fi
      if [ -d "${_G_UNIRTM_SHIMS_BASE:-}" ] && [ -n "$(ls -A "${_G_UNIRTM_SHIMS_BASE:-}" 2>/dev/null)" ]; then
        _persist_path_to_ci "${_G_UNIRTM_SHIMS_BASE:-}"
        # Use echo to avoid printf issues with Windows paths
        echo "  [OK] Persisted unirtm shims: ${_G_UNIRTM_SHIMS_BASE:-}" >&2
      else
        echo "  [WARN] unirtm shims directory not found or empty: ${_G_UNIRTM_SHIMS_BASE:-}" >&2
      fi
    fi

    if [ "${DRY_RUN:-0}" -eq 0 ]; then
      if ! command -v unirtm >/dev/null 2>&1 && ! command -v unirtm >/dev/null 2>&1; then
        log_warn "Warning: unirtm/unirtm binary not found on PATH. You may need to restart your shell."
      fi
      printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
      printf "  - Run %bmake install%b to install project dependencies.\n" "${GREEN:-}" "${NC:-}"
      printf "  - Run %bmake verify%b to ensure environment health.\n" "${GREEN:-}" "${NC:-}"
    fi
  fi
}

main "$@"
