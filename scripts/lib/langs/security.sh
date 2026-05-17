#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Security Logic Module

# Purpose: Installs osv-scanner for vulnerability scanning.
# CI-only: Compiled via `go install` from Go proxy. Local dev skips to avoid build time.
install_osv_scanner() {
  # Tier 3 Tool: Optional for local development.
  if ! is_ci_env && [ "${OSV_FORCE_INSTALL:-0}" -ne 1 ]; then
    log_summary "Security" "OSV-Scanner" "⏭️ Optional (CI-only by default)" "-" "0"
    log_info "⏭️  OSV-Scanner: Optional for local development. Set OSV_FORCE_INSTALL=1 to force installation."
    return 0
  fi

  # Skip if manually disabled
  if [ "${SKIP_OSV:-0}" -eq 1 ]; then
    return 0
  fi

  setup_registry_osv_scanner
  install_tool_safe "osv-scanner" "${VER_OSV_SCANNER_PROVIDER:-}" "OSV-Scanner" "--version" 1
}

# NOTE: Trivy CLI installation removed. Vulnerability scanning is handled by
# aquasecurity/trivy-action in CI workflows (ci.yml), which bundles its own binary.
# This eliminates ~80MB GitHub Release downloads and prevents GitHub API 403 rate limits.

# Purpose: Installs zizmor for GitHub Actions security linting.
# Delegate: Managed by unirtm (.unirtm.toml)
# NOTE: Zizmor is a Tier 1 tool in .unirtm.toml and should be installed in CI environments.
install_zizmor() {
  # CI-only: GH Actions security linter is rarely needed for local app code.
  # Tier 1 Tool in CI: Critical for GitHub Actions security scanning.
  # Tier 3 Tool locally: Optional for local development.
  if ! is_ci_env && [ "${ZIZMOR_FORCE_INSTALL:-0}" -ne 1 ]; then
    log_summary "Security" "Zizmor" "⏭️ Optional (CI-only by default)" "-" "0"
    log_info "⏭️  Zizmor: Optional for local development. Set ZIZMOR_FORCE_INSTALL=1 to force installation."
    return 0
  fi

  # In CI, install regardless of workflow file presence (Tier 1 tool in .unirtm.toml)
  # Locally, only install if workflow files exist
  if ! is_ci_env && ! has_lang_files ".github/workflows" "*.yaml *.yml"; then
    log_summary "Security" "Zizmor" "⏭️ Skipped (no workflows)" "-" "0"
    return 0
  fi

  setup_registry_zizmor
  install_tool_safe "zizmor" "${VER_ZIZMOR_PROVIDER:-}" "Zizmor" "--version" 1
}

# Purpose: Installs cargo-audit for Rust vulnerability scanning.
# CI-only: Requires downloading the Rust Advisory DB (network-heavy). Local dev skips.
# Delegate: Managed by unirtm (.unirtm.toml)
install_cargo_audit() {
  if ! has_lang_files "Cargo.toml Cargo.lock" ""; then
    return 0
  fi

  # CI-only guard: Advisory DB download is network-heavy, skip locally.
  # Tier 3 Tool: Optional for local development.
  if ! is_ci_env && [ "${CA_FORCE_INSTALL:-0}" -ne 1 ]; then
    log_summary "Security" "Cargo-audit" "⏭️ Optional (CI-only by default)" "-" "0"
    log_info "⏭️  Cargo-audit: Optional for local development. Set CA_FORCE_INSTALL=1 to force installation."
    return 0
  fi

  setup_registry_cargo_audit
  install_tool_safe "cargo-audit" "${VER_CARGO_AUDIT_PROVIDER:-}" "Cargo-Audit" "--version" 1
}

# Purpose: Sets up Security audit environment.
setup_security() {
  install_osv_scanner
  install_zizmor
  setup_rego # From rego.sh
  # Optional: logic for govulncheck/pip-audit can stay in go.sh/python.sh or be here
  install_cargo_audit
}
