#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/update-tools.sh - Intelligent UniRTM Tool Version Upgrader
#
# Purpose:
#   Automatically detects the latest upstream versions for all UniRTM-managed tools
#   using 'unirtm ls-remote' and updates the project's version registry (versions.sh).
#
# Usage:
#   sh scripts/update-tools.sh [OPTIONS]
#
# Options:
#   --dry-run        Preview version upgrades without modifying versions.sh.
#   -q, --quiet      Suppress informational output.
#   -v, --verbose    Enable verbose/debug output.
#   -h, --help       Show this help message.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Displays usage information for the tool upgrader.
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Intelligent tool version upgrader for Mise.

Options:
  --dry-run        Preview upgrades without applying them.
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

EOF
}

# Purpose: Detects the latest stable version for a given tool/provider.
# Params:
#   $1 - Tool name/Provider (e.g., "github:cli/cli", "npm:pnpm")
# Examples:
#   _get_latest_version "github:cli/cli"
_get_latest_version() {
  local _PROVIDER="${1:-}"
  local _LATEST=""

  # Use unirtm ls-remote to fetch versions.
  # We filter out typical pre-release keywords and common single-letter suffixes (a/b/rc).
  # We use sort -V to ensure semver-compliant ordering.
  _LATEST=$(mise ls-remote "${_PROVIDER:-}" 2>/dev/null |
    grep -Ev "(rc|beta|alpha|dev|nightly|test|pre|-build|[0-9][ab][0-9])" |
    grep -E "^[v]?[0-9]" |
    sort -V |
    tail -n 1 || true)

  echo "${_LATEST:-}"
}

# Purpose: Main execution logic for upgrading versions.sh and .unirtm.toml.
run_upgrade() {
  local _VERSIONS_FILE="scripts/lib/versions.sh"
  local _MISE_FILE=".mise.toml"
  local _UPDATED_COUNT=0
  local _CHECK_COUNT=0
  local _SUMMARY_DATA=""

  log_info "🔍 Checking for tool version upgrades..."

  # --- Phase 1: Update versions.sh (SSoT for Tier 2) ---
  if [ -f "${_VERSIONS_FILE:-}" ]; then
    log_debug "Scanning ${_VERSIONS_FILE:-}..."
    # shellcheck source=scripts/lib/versions.sh
    . "${_VERSIONS_FILE:-}"

    _TMP_VARS=".mise.versions.tmp"
    grep -E "^VER_[A-Z0-9_]+=" "${_VERSIONS_FILE:-}" >"${_TMP_VARS:-}"

    while IFS= read -r line; do
      [ -z "${line:-}" ] && continue
      _VAR_NAME=$(echo "${line:-}" | cut -d= -f1)
      _CUR_VER=$(echo "${line:-}" | cut -d= -f2- | tr -d '"')

      case "${_VAR_NAME:-}" in
      *_PROVIDER | *_REF | *_URL | *_SHA* | *_PROVIDER_REF_*) continue ;;
      esac

      _PROV_VAR="${_VAR_NAME:-}_PROVIDER"
      _PROV_VAL=$(eval "echo \${${_PROV_VAR:-}:-}")

      if [ -z "${_PROV_VAL:-}" ]; then
        _TOOL_NAME=$(echo "${_VAR_NAME:-}" | sed 's/^VER_//' | tr '[:upper:]' '[:lower:]' | tr '_' '-')
        case "${_TOOL_NAME:-}" in
        go | node | python | ruby | java | rust | kotlin | dotnet | bun | deno | zig) _PROV_VAL="${_TOOL_NAME:-}" ;;
        *) continue ;;
        esac
      fi

      _LATEST_VER=$(_get_latest_version "${_PROV_VAL:-}")
      if [ -n "${_LATEST_VER:-}" ] && [ "${_CUR_VER:-}" != "${_LATEST_VER:-}" ]; then
        log_success "✨ Upgrade [${_VERSIONS_FILE:-}]: ${_VAR_NAME:-} (${_PROV_VAL:-}) ${_CUR_VER:-} -> ${_LATEST_VER:-}"
        _SUMMARY_DATA="${_SUMMARY_DATA:-}| \`${_PROV_VAL:-}\` | \`${_CUR_VER:-}\` | \`${_LATEST_VER:-}\` | \`versions.sh\` |\n"
        if [ "${DRY_RUN:-0}" -eq 0 ]; then
          if [ "$(uname -s)" = "Darwin" ]; then
            sed -i '' "s/${_VAR_NAME:-}=\"${_CUR_VER:-}\"/${_VAR_NAME:-}=\"${_LATEST_VER:-}\"/" "${_VERSIONS_FILE:-}"
          else
            sed -i "s/${_VAR_NAME:-}=\"${_CUR_VER:-}\"/${_VAR_NAME:-}=\"${_LATEST_VER:-}\"/" "${_VERSIONS_FILE:-}"
          fi
        fi
        _UPDATED_COUNT=$((_UPDATED_COUNT + 1))
      fi
      _CHECK_COUNT=$((_CHECK_COUNT + 1))
    done <"${_TMP_VARS:-}"
    rm -f "${_TMP_VARS:-}"
  fi

  # --- Phase 2: Update .unirtm.toml (Tier 1 & Static Global Tools) ---
  if [ -f "${_MISE_FILE:-}" ]; then
    log_debug "Scanning ${_MISE_FILE:-}..."
    # We look for lines in the [tools] section: tool = "version"
    # Note: We only process lines that look like Assignments until we hit the next section
    _IN_TOOLS=0
    _TMP_MISE_SCAN=".mise.toml.scan.tmp"
    cp "${_MISE_FILE:-}" "${_TMP_MISE_SCAN:-}"

    # Process line by line from the scan file to avoid SC2094
    while IFS= read -r line; do
      case "${line:-}" in
      "[tools]") _IN_TOOLS=1 ;;
      "["*"]") _IN_TOOLS=0 ;;
      esac

      if [ "${_IN_TOOLS:-}" -eq 1 ] && echo "${line:-}" | grep -qE '^[a-zA-Z0-9:"/@._-]+ = "[0-9].*"'; then
        _TOOL_IDENT=$(echo "${line:-}" | cut -d= -f1 | sed 's/[[:space:]]*$//' | tr -d '"')
        _CUR_VER=$(echo "${line:-}" | cut -d= -f2- | sed 's/^[[:space:]]*//' | tr -d '"')

        # Determine provider
        _PROV_VAL="${_TOOL_IDENT:-}"

        # [CI ISOLATION]: In CI environments, strictly forbid updating Tier 3 dynamic tools
        # to ensure they are never persisted back to the repository via automated syncs.
        if is_ci_env; then
          case "${_TOOL_IDENT:-}" in
          *osv-scanner* | *zizmor* | *pip-audit* | *govulncheck* | *cargo-audit*)
            log_warn "⏭️  CI Filter: Bypassing upgrade for Tier 3 tool [${_TOOL_IDENT:-}] in ${_MISE_FILE:-}"
            continue
            ;;
          esac
        fi

        _LATEST_VER=$(_get_latest_version "${_PROV_VAL:-}")
        if [ -n "${_LATEST_VER:-}" ] && [ "${_CUR_VER:-}" != "${_LATEST_VER:-}" ]; then
          log_success "✨ Upgrade [${_MISE_FILE:-}]: ${_TOOL_IDENT:-} ${_CUR_VER:-} -> ${_LATEST_VER:-}"
          _SUMMARY_DATA="${_SUMMARY_DATA:-}| \`${_TOOL_IDENT:-}\` | \`${_CUR_VER:-}\` | \`${_LATEST_VER:-}\` | \`.mise.toml\` |\n"
          if [ "${DRY_RUN:-0}" -eq 0 ]; then
            if [ "$(uname -s)" = "Darwin" ]; then
              sed -i '' "s#${_TOOL_IDENT:-} = \"${_CUR_VER:-}\"#${_TOOL_IDENT:-} = \"${_LATEST_VER:-}\"#" "${_MISE_FILE:-}"
              # Also handle quoted variant if any
              sed -i '' "s#\"${_TOOL_IDENT:-}\" = \"${_CUR_VER:-}\"#\"${_TOOL_IDENT:-}\" = \"${_LATEST_VER:-}\"#" "${_MISE_FILE:-}"
            else
              sed -i "s#${_TOOL_IDENT:-} = \"${_CUR_VER:-}\"#${_TOOL_IDENT:-} = \"${_LATEST_VER:-}\"#" "${_MISE_FILE:-}"
              sed -i "s#\"${_TOOL_IDENT:-}\" = \"${_CUR_VER:-}\"#\"${_TOOL_IDENT:-}\" = \"${_LATEST_VER:-}\"#" "${_MISE_FILE:-}"
            fi
          fi
          _UPDATED_COUNT=$((_UPDATED_COUNT + 1))
        fi
        _CHECK_COUNT=$((_CHECK_COUNT + 1))
      fi
    done <"${_TMP_MISE_SCAN:-}"
    rm -f "${_TMP_MISE_SCAN:-}"
  fi

  # --- Write Summary to GitHub Actions ---
  # --- Write Summary to CI Actions ---
  if [ -n "${CI_STEP_SUMMARY:-}" ] && [ -n "${_SUMMARY_DATA:-}" ]; then
    {
      echo "### 🛠️ Toolchain Upgrade Summary"
      echo ""
      echo "| Tool | Old Version | New Version | Source |"
      echo "| :--- | :--- | :--- | :--- |"
      printf "%b" "${_SUMMARY_DATA:-}"
      echo ""
      echo "> Generated by \`scripts/update-tools.sh\` at $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
    } >>"${CI_STEP_SUMMARY:-}"
  fi

  if [ "${_UPDATED_COUNT:-}" -gt 0 ]; then
    if [ "${DRY_RUN:-0}" -eq 0 ]; then
      log_success "\n✅ Successfully updated ${_UPDATED_COUNT:-} tool(s) across registry and config."
      log_info "💡 Next: Run 'make sync-lock' to update cryptographic hashes."
    else
      log_info "\n[DRY-RUN] Process completed. ${_UPDATED_COUNT:-} tool(s) have pending updates."
    fi
  else
    log_success "\n🎉 All tools are up to date! (${_CHECK_COUNT:-} checks performed)"
  fi
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
  # 0. Governance Check: Skip if auto-update is disabled
  if [ "${CONFIG_AUTO_UPDATE:-1}" = "0" ]; then
    echo "⏭️  Auto-update is disabled (CONFIG_AUTO_UPDATE=0). Skipping tool upgrades." >&2
    exit 0
  fi

  guard_project_root
  parse_common_args "$@"

  # Load versions.sh to have the providers available
  . "scripts/lib/versions.sh"

  run_upgrade "$@"
}

main "$@"
