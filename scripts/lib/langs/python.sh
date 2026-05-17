#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Python Logic Module

# Purpose: Installs the Python runtime, creates a venv, and installs dependencies.
# Delegate: Managed via unirtm (.unirtm.toml) and pip.
install_runtime_python() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Python runtime and virtual environment."
    return 0
  fi

  # Alpine/musl detection and build dependencies installation
  # Python needs to be compiled from source on Alpine (no prebuilt musl binaries)
  if [ -f /etc/alpine-release ] || [ "${ALPINE_VERSION:-}" != "" ]; then
    log_info "Alpine/musl environment detected. Installing Python build dependencies..."

    # Check if apk is available
    if command -v apk >/dev/null 2>&1; then
      # Install essential build dependencies for Python compilation
      # These are required for building Python from source on Alpine
      apk add --no-cache \
        build-base \
        libffi-dev \
        openssl-dev \
        bzip2-dev \
        zlib-dev \
        xz-dev \
        readline-dev \
        sqlite-dev \
        tk-dev \
        ncurses-dev \
        gdbm-dev \
        libnsl-dev \
        libtirpc-dev \
        linux-headers ||
        log_warn "Failed to install some Python build dependencies. Compilation may fail."
    fi
  fi

  # 1. Runtime initialization
  local _VERSION="${VER_PYTHON:-}"
  unirtm install "python@${_VERSION:-}"

  # 2. Virtualenv management
  if [ ! -d "${VENV:-}" ]; then
    run_quiet "${PYTHON:-python3}" -m venv "${VENV:-}"
  fi

  # 3. Dependency resolution
  if [ -d "${VENV:-}" ]; then
    # Standard requirements
    if [ -f "${REQUIREMENTS_TXT:-}" ]; then
      run_quiet "$VENV/${_G_VENV_BIN:-}/pip" install -r "${REQUIREMENTS_TXT:-}"
    fi
    # Dev requirements (setup.sh specific but safe here)
    if [ -f "requirements-dev.txt" ]; then
      run_quiet "$VENV/${_G_VENV_BIN:-}/pip" install -r "requirements-dev.txt"
    fi
  fi
}

# Purpose: Installs Ruff.
# Delegate: Managed by unirtm (.unirtm.toml)
install_ruff() {
  if ! has_lang_files "requirements.txt pyproject.toml" "*.py"; then
    return 0
  fi

  install_tool_safe "ruff" "${VER_RUFF_PROVIDER:-}" "Ruff" "--version" 0 "*.py" ""
}

# Purpose: Installs pip-audit for Python dependency vulnerability scanning.
# NOTE: CI-only tool — security audit. Skipped on local environments.
install_pip_audit() {
  # CI-only: Optional security tool, local dev skips to avoid pip-audit installation overhead.
  if ! is_ci_env && [ "${PA_FORCE_INSTALL:-0}" -ne 1 ]; then
    return 0
  fi

  if ! has_lang_files "requirements.txt pyproject.toml" "*.py"; then
    return 0
  fi

  install_tool_safe "pip-audit" "${VER_PIP_AUDIT_PROVIDER:-}" "pip-audit" "--version" 1
}

# Purpose: Sets up Python runtime for project.
# Delegate: Managed by unirtm (.unirtm.toml)
setup_python() {
  # Python is a first-class citizen: setup is always performed.

  local _T0_PY_RT
  _T0_PY_RT=$(date +%s)
  local _TITLE="Python Virtual Environment"
  local _PROVIDER="python"

  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version python)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "${_PROVIDER:-}")

  # Always log setup start for consistency and test assertions
  _log_setup "${_TITLE:-}" "${_PROVIDER:-}"

  if [ "${_CUR_VER:-}" != "-" ] && [ "${_CUR_VER:-}" = "${_REQ_VER:-}" ] && [ -d "${VENV:-}" ]; then
    log_summary "Runtime" "Python" "✅ Detected" "${_CUR_VER:-}" "0"
  else

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_summary "Runtime" "Python" "⚖️ Previewed" "-" "0"
    else
      local _STAT_PY_RT="✅ Installed"
      install_runtime_python || _STAT_PY_RT="❌ Failed"

      local _DUR_PY_RT
      _DUR_PY_RT=$(($(date +%s) - _T0_PY_RT))
      log_summary "Runtime" "Python" "${_STAT_PY_RT:-}" "$(get_version python)" "${_DUR_PY_RT:-}"
    fi
  fi

  # Setup related tools (Conditional on project files)
  if has_lang_files "requirements.txt pyproject.toml" "*.py"; then
    install_ruff
    install_pip_audit
  fi
}
# Purpose: Checks if Python runtime is available.
# Examples:
#   check_runtime_python "Linter"
check_runtime_python() {
  local _TOOL_DESC_PY="${1:-Python}"
  if ! command -v "${PYTHON:-python3}" >/dev/null 2>&1; then
    log_warn "Required runtime 'python' for ${_TOOL_DESC_PY:-} is missing. Skipping."
    return 1
  fi
  return 0
}
