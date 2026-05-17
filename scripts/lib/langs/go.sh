#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Go Logic Module

# Purpose: Installs Go runtime via unirtm (version pinned in scripts/lib/versions.sh).
install_runtime_go() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Go runtime."
    return 0
  fi
  "${_G_UNIRTM_BIN:-unirtm}" install "go@${VER_GO:-}"
}

# Purpose: Installs golangci-lint for Go projects (version pinned in versions.sh).
install_go_lint() {
  if ! has_lang_files "go.mod" "*.go"; then
    return 0
  fi

  # Support verbose logging for Tier 2 tools to track large binary downloads (~50MB)
  log_info "Downloading golangci-lint (this may take a few minutes on slow networks)..."
  UNIRTM_HTTP_TIMEOUT=600s VERBOSE=2 ENABLE_GITHUB_PROXY=1 install_tool_safe "golangci-lint" "golangci-lint" "Go Lint" "--version" 1
}

# Purpose: Installs govulncheck for Go project vulnerability scanning.
# NOTE: CI-only tool — vulnerability scanner. Skipped on local environments.
install_govulncheck() {
  # CI-only: Go vulnerability checks are typically heavier and reserved for CI workflows.
  if ! is_ci_env && [ "${GOVULN_FORCE_INSTALL:-0}" -ne 1 ]; then
    return 0
  fi

  if ! has_lang_files "go.mod" "*.go"; then
    return 0
  fi

  install_tool_safe "govulncheck" "${VER_GOVULNCHECK_PROVIDER:-}" "Govulncheck" "--version" 1
}

# Purpose: Sets up Go runtime for project.
# Delegate: Managed by unirtm (.unirtm.toml)
setup_go() {
  # Go is a first-class citizen: setup is always performed.

  # Dynamically register Go in .unirtm.toml if not already present.
  # This is essential for pre-provisioning (e.g., DevContainer builds)
  # where Go is explicitly requested before source files exist.
  setup_registry_go

  local _T0_GO_RT
  _T0_GO_RT=$(date +%s)
  local _TITLE="Go Runtime"
  local _PROVIDER="go"

  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version go)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "${_PROVIDER:-}")

  # Always log setup start for consistency and test assertions
  _log_setup "${_TITLE:-}" "${_PROVIDER:-}"

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Go" "✅ Detected" "${_CUR_VER:-}" "0"
  else

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_summary "Runtime" "Go" "⚖️ Previewed" "-" "0"
    else
      local _STAT_GO_RT="✅ Installed"
      install_runtime_go || _STAT_GO_RT="❌ Failed"

      local _DUR_GO_RT
      _DUR_GO_RT=$(($(date +%s) - _T0_GO_RT))
      log_summary "Runtime" "Go" "${_STAT_GO_RT:-}" "$(get_version go)" "${_DUR_GO_RT:-}"
    fi
  fi

  # Setup related tools (Conditional on project files)
  if has_lang_files "go.mod" "*.go"; then
    install_go_lint
    install_govulncheck
  fi
}
# Purpose: Checks if Go runtime is available.
# Examples:
#   check_runtime_go "Linter"
check_runtime_go() {
  local _TOOL_DESC_GO="${1:-Go}"
  if ! resolve_bin "go" >/dev/null 2>&1; then
    log_warn "Required runtime 'go' for ${_TOOL_DESC_GO:-} is missing. Skipping."
    return 1
  fi
  return 0
}
