#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/audit.sh - Automated Security Auditor
#
# Purpose:
#   Standardizes execution of dependency scans and secret detection modules.
#   Identifies vulnerabilities and leaks across all project stacks.
#
# Usage:
#   sh scripts/audit.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (Network), Rule 05 (Dependencies), Rule 08 (Dev Env).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Multi-stack scanning (Node.js, pip, go, cargo, osv).
#   - Gitleaks integration for secrets detection.
#   - Binary artifact discovery to prevent poisoning.

# Note: We use set -eu but ensure that each audit module handles its own failure
# status via '|| true' or 'if' blocks to allow the full audit to complete.

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"
. "${SCRIPT_DIR:-}/lib/github-api-info.sh"

# Purpose: Displays usage information for the security auditor.
# Examples:
#   show_help
# shellcheck disable=SC2317,SC2329
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Run a full security audit across all detected project stacks.

Options:
  --dry-run        Preview audit steps without real execution.
  -q, --quiet      Suppress verbose orchestration details.
  -v, --verbose    Enable verbose output for all sub-tools.
  -h, --help       Show this help message.

EOF
}

# Purpose: Main entry point for the security auditing engine.
#          Coordinates multi-stack dependency scans and secret detection.
# Params:
#   $@ - Command line arguments
# Examples:
#   main --verbose
main() {
  # Generate a temporary full manifest so unirtm can resolve all locked versions.
  ./scripts/gen-full-manifest.sh >.unirtm.audit.toml
  UNIRTM_CONFIG="$(pwd)/.unirtm.audit.toml"
  export UNIRTM_CONFIG
  trap "rm -f .unirtm.audit.toml" EXIT INT TERM
  guard_project_root

  # 2. Argument Parsing
  parse_common_args "$@"

  local _OVERALL_EXIT_AUDIT=0
  local _START_TIME_AUDIT
  _START_TIME_AUDIT=$(date +%s)

  # Initialize Summary File correctly
  init_summary_table "Security Audit Execution Summary"

  # 3. Secrets Scanning
  local _GITLEAKS_BIN
  _GITLEAKS_BIN=$(resolve_bin "gitleaks") || true
  if [ -n "${_GITLEAKS_BIN:-}" ]; then
    local _T0_GL
    _T0_GL=$(date +%s)
    log_info "── Scanning for Secrets (gitleaks) ──"
    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_success "DRY-RUN: Would run gitleaks detect"
      log_summary "Security" "gitleaks" "⚖️ Previewed" "-" "0"
    else
      # Run without --verbose to avoid git rename-detection warnings on stderr that
      # cause gitleaks to emit error="stderr is not empty" even when no leaks exist.
      # Set diff.renameLimit=5000 to handle large repos (>3490 renames).
      # GIT_CONFIG_PARAMETERS format requires "'key=value'" per entry, comma-separated.
      _GL_GIT_PARAMS="'diff.renameLimit=5000'"
      if [ -n "${GIT_CONFIG_PARAMETERS:-}" ]; then
        _GL_GIT_PARAMS="${GIT_CONFIG_PARAMETERS:-},'diff.renameLimit=5000'"
      fi
      _GL_ARGS="detect --source . --no-banner"
      if [ -n "${GITHUB_BASE_REF:-}" ] && { is_ci_env || [ "${GITLEAKS_FORCE_INCREMENTAL:-0}" -eq 1 ]; }; then
        log_info "Gitleaks: Performing incremental PR scan (origin/${GITHUB_BASE_REF:-}..HEAD)..."
        _GL_ARGS="detect --log-opts=origin/${GITHUB_BASE_REF:-}..HEAD --no-banner"
      fi

      # shellcheck disable=SC2086
      if GIT_CONFIG_PARAMETERS="${_GL_GIT_PARAMS:-}" \
        run_quiet "${_GITLEAKS_BIN:-}" ${_GL_ARGS:-}; then
        log_summary "Security" "gitleaks" "✅ Clean" "$(get_version "${_GITLEAKS_BIN:-}")" "$(($(date +%s) - _T0_GL))"
      else
        log_summary "Security" "gitleaks" "❌ Leaks Found" "$(get_version "${_GITLEAKS_BIN:-}")" "$(($(date +%s) - _T0_GL))"
        _OVERALL_EXIT_AUDIT=1
      fi
    fi
  elif is_ci_env; then
    log_error "gitleaks not found in CI. Failing Secrets Scan."
    log_summary "Security" "gitleaks" "❌ Missing (Fatal)" "-" "0"
    _OVERALL_EXIT_AUDIT=1
  else
    log_warn "gitleaks not found locally. Skipping Secrets Scan."
    log_summary "Security" "gitleaks" "⏭️  Skipped" "-" "0"
  fi

  # 4. GitHub Actions Security
  if [ -d ".github/workflows" ]; then
    local _T0_ZM
    _T0_ZM=$(date +%s)
    local _ZIZMOR_BIN
    _ZIZMOR_BIN=$(resolve_bin "zizmor") || true

    # Logic: Always run zizmor as it is dynamically installed.
    local _ZM_SPEC="${VER_ZIZMOR_PROVIDER:-zizmor}@${VER_ZIZMOR:-latest}"
    if [ -n "${_ZM_SPEC:-}" ]; then
      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        log_success "DRY-RUN: Would run zizmor"
        log_summary "GitHub" "zizmor" "⚖️ Previewed" "-" "0"
      else
        # zizmor logic...
        local _ZM_OK=0
        if [ -n "${GITHUB_TOKEN:-}" ]; then
          log_info "Zizmor: Attempting authenticated scan..."
          export GH_TOKEN="${GITHUB_TOKEN:-}"
          if run_quiet "${_G_UNIRTM_BIN:-unirtm}" exec "${_ZM_SPEC:-}" -- zizmor . --format plain --config .zizmor.yml --gh-token "${GITHUB_TOKEN:-}"; then
            _ZM_OK=1
          fi
        fi

        if [ "${_ZM_OK:-}" -eq 0 ]; then
          log_info "Zizmor: Attempting offline scan (fallback)..."
          if run_quiet "${_G_UNIRTM_BIN:-unirtm}" exec "${_ZM_SPEC:-}" -- zizmor . --format plain --config .zizmor.yml --offline; then
            _ZM_OK=1
          fi
        fi

        if [ "${_ZM_OK:-}" -eq 1 ]; then
          log_summary "GitHub" "zizmor" "✅ Secure" "${VER_ZIZMOR:-dynamic}" "$(($(date +%s) - _T0_ZM))"
        else
          log_summary "GitHub" "zizmor" "❌ Vulnerable" "${VER_ZIZMOR:-dynamic}" "$(($(date +%s) - _T0_ZM))"
          _OVERALL_EXIT_AUDIT=1
        fi
      fi
    elif is_ci_env; then
      log_error "zizmor not found in CI. Failing GitHub Actions Security Scan."
      log_summary "GitHub" "zizmor" "❌ Missing (Fatal)" "-" "0"
      _OVERALL_EXIT_AUDIT=1
    else
      log_warn "zizmor not found locally. Skipping GitHub Actions Security Scan."
      log_summary "GitHub" "zizmor" "⏭️  Skipped" "-" "0"
    fi
  fi

  # 5. Dependency Audits (Node.js)
  if [ -f "${PACKAGE_JSON:-}" ]; then
    local _NPM_BIN
    _NPM_BIN=$(resolve_bin "${NPM:-npm}") || true

    # CI Fallback: Try unirtm exec if resolve_bin fails (Windows hollow shims issue)
    if [ -z "${_NPM_BIN:-}" ] && is_ci_env; then
      log_info "Attempting to run ${NPM:-npm} via unirtm exec..."
      # Determine unirtm tool spec for the package manager
      local _NPM_SPEC=""
      case "${NPM:-npm}" in
      pnpm) _NPM_SPEC="npm:pnpm" ;;
      yarn) _NPM_SPEC="npm:yarn" ;;
      bun) _NPM_SPEC="bun" ;;
      *) _NPM_SPEC="node" ;;
      esac

      # Test if unirtm exec works
      if "${_G_UNIRTM_BIN:-unirtm}" exec "${_NPM_SPEC:-}" -- "${NPM:-npm}" --version >/dev/null 2>&1; then
        _NPM_BIN="unirtm_exec_wrapper"
        log_info "Successfully validated ${NPM:-npm} via unirtm exec"
      fi
    fi

    if [ -n "${_NPM_BIN:-}" ]; then
      local _T0_JS
      _T0_JS=$(date +%s)
      log_info "\n── Auditing Node.js dependencies (${NPM:-npm} audit) ──"
      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        log_success "DRY-RUN: Would run ${NPM:-npm} audit"
        log_summary "Node.js" "${NPM:-npm}-audit" "⚖️ Previewed" "-" "0"
      else
        # Always use the official npm registry for audit — mirror registries (e.g.
        # npmmirror.com) do not implement the audit endpoint and return 404.
        local _AUDIT_REGISTRY="https://registry.npmjs.org"

        # Determine unirtm tool spec for the package manager
        local _NPM_SPEC=""
        case "${NPM:-npm}" in
        pnpm) _NPM_SPEC="npm:pnpm" ;;
        yarn) _NPM_SPEC="npm:yarn" ;;
        bun) _NPM_SPEC="bun" ;;
        *) _NPM_SPEC="node" ;;
        esac

        # Execute audit command (use unirtm exec if resolve_bin failed)
        local _AUDIT_OK=0
        if [ "${_NPM_BIN:-}" = "unirtm_exec_wrapper" ]; then
          # Use unirtm exec for hollow shims on Windows
          if run_quiet "${_G_UNIRTM_BIN:-unirtm}" exec "${_NPM_SPEC:-}" -- "${NPM:-npm}" audit --registry="${_AUDIT_REGISTRY:-}"; then
            _AUDIT_OK=1
          fi
        else
          # Use resolved binary path
          if run_quiet "${_NPM_BIN:-}" audit --registry="${_AUDIT_REGISTRY:-}"; then
            _AUDIT_OK=1
          fi
        fi

        if [ "${_AUDIT_OK:-}" -eq 1 ]; then
          log_summary "Node.js" "${NPM:-npm}-audit" "✅ Secure" "$(get_version "${NPM:-npm}")" "$(($(date +%s) - _T0_JS))"
        else
          log_summary "Node.js" "${NPM:-npm}-audit" "⚠️ Vulnerabilities" "$(get_version "${NPM:-npm}")" "$(($(date +%s) - _T0_JS))"
          # Non-fatal: audit findings are informational for a template project.
          # Downstream projects should run their own audit with their own registry.
        fi
      fi
    elif is_ci_env; then
      log_error "${NPM:-npm} not found in CI. Failing Node.js audit."
      log_summary "Node.js" "${NPM:-npm}-audit" "❌ Missing (Fatal)" "-" "0"
      _OVERALL_EXIT_AUDIT=1
    else
      log_warn "${NPM:-npm} not found locally. Skipping Node.js audit."
      log_summary "Node.js" "${NPM:-npm}-audit" "⏭️  Skipped" "-" "0"
    fi
  fi

  # 6. Dependency Audits (Python)
  local _PA_BIN
  _PA_BIN=$(resolve_bin "pip-audit") || true
  if { [ -f "requirements-dev.txt" ] || [ -f "requirements.txt" ] || [ -f "pyproject.toml" ]; }; then
    if [ -n "${_PA_BIN:-}" ]; then
      local _T0_PY_AUD
      _T0_PY_AUD=$(date +%s)
      log_info "\n── Auditing Python dependencies (pip-audit) ──"

      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        log_success "DRY-RUN: Would run ${_PA_BIN:-}"
        log_summary "Python" "pip-audit" "⚖️ Previewed" "-" "0"
      else
        local _PA_SPEC="${VER_PIP_AUDIT_PROVIDER:-pip-audit}@${VER_PIP_AUDIT:-latest}"
        if run_quiet "${_G_UNIRTM_BIN:-unirtm}" exec "${_PA_SPEC:-}" -- pip-audit; then
          log_summary "Python" "pip-audit" "✅ Secure" "$(get_version pip-audit --version)" "$(($(date +%s) - _T0_PY_AUD))"
        else
          log_summary "Python" "pip-audit" "❌ Vulnerable" "$(get_version pip-audit --version)" "$(($(date +%s) - _T0_PY_AUD))"
          _OVERALL_EXIT_AUDIT=1
        fi
      fi
    elif is_ci_env; then
      log_error "pip-audit not found in CI. Failing Python audit."
      log_summary "Python" "pip-audit" "❌ Missing (Fatal)" "-" "0"
      _OVERALL_EXIT_AUDIT=1
    else
      log_warn "pip-audit not found locally. Skipping Python audit."
      log_summary "Python" "pip-audit" "⏭️  Skipped" "-" "0"
    fi
  fi

  # 7. Multi-Stack Audit (OSV-Scanner)
  # Only run when at least one lockfile is present to avoid pointless scan on bare projects.
  _HAS_LOCKFILE=0
  for _lf in package-lock.json pnpm-lock.yaml yarn.lock go.sum Cargo.lock requirements.txt Pipfile.lock; do
    [ -f "${_lf:-}" ] && _HAS_LOCKFILE=1 && break
  done
  local _OSV_BIN
  _OSV_BIN=$(resolve_bin "osv-scanner") || true

  if [ "${_HAS_LOCKFILE:-}" -eq 1 ]; then
    if [ -n "${_OSV_BIN:-}" ]; then
      local _T0_OSV_AUD
      _T0_OSV_AUD=$(date +%s)
      log_info "\n── Generic Vulnerability Scan (osv-scanner) ──"

      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        log_success "DRY-RUN: Would run osv-scanner"
        log_summary "Security" "osv-scanner" "⚖️ Previewed" "-" "0"
      else
        local _OSV_SPEC="${VER_OSV_SCANNER_PROVIDER:-osv-scanner}@${VER_OSV_SCANNER:-latest}"
        _OSV_OUT=$("${_G_UNIRTM_BIN:-unirtm}" exec "${_OSV_SPEC:-}" -- osv-scanner scan . --config .osv-scanner.toml --call-analysis=all --format table 2>&1) || _OSV_EXIT=$?
        [ -n "${_OSV_EXIT:-}" ] || _OSV_EXIT=0
        if [ "${_OSV_EXIT:-}" -eq 0 ]; then
          log_summary "Security" "osv-scanner" "✅ Secure" "$(get_version osv-scanner)" "$(($(date +%s) - _T0_OSV_AUD))"
        elif echo "${_OSV_OUT:-}" | grep -q "No package sources found"; then
          log_info "osv-scanner: No package sources found. Skipping."
          log_summary "Security" "osv-scanner" "⏭️  Skipped" "$(get_version osv-scanner)" "$(($(date +%s) - _T0_OSV_AUD))"
        else
          echo "${_OSV_OUT:-}"
          log_summary "Security" "osv-scanner" "⚠️ Findings" "$(get_version osv-scanner)" "$(($(date +%s) - _T0_OSV_AUD))"
        fi
      fi
    elif is_ci_env; then
      log_error "osv-scanner not found in CI. Failing generic vulnerability scan."
      log_summary "Security" "osv-scanner" "❌ Missing (Fatal)" "-" "0"
      _OVERALL_EXIT_AUDIT=1
    else
      log_warn "osv-scanner not found locally. Skipping generic vulnerability scan."
      log_summary "Security" "osv-scanner" "⏭️  Skipped" "-" "0"
    fi
  fi

  # 8. Stack Specific (Go)
  local _GOVULN_BIN
  _GOVULN_BIN=$(resolve_bin "govulncheck") || true
  if [ -f "go.mod" ]; then
    if [ -n "${_GOVULN_BIN:-}" ]; then
      local _T0_GO_AUD
      _T0_GO_AUD=$(date +%s)
      log_info "\n── Auditing Go dependencies (govulncheck) ──"
      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        log_success "DRY-RUN: Would run govulncheck"
        log_summary "Go" "govulncheck" "⚖️ Previewed" "-" "0"
      else
        local _GOV_SPEC="${VER_GOVULNCHECK_PROVIDER:-govulncheck}@${VER_GOVULNCHECK:-latest}"
        if run_quiet "${_G_UNIRTM_BIN:-unirtm}" exec "${_GOV_SPEC:-}" -- govulncheck ./...; then
          log_summary "Go" "govulncheck" "✅ Secure" "$(get_version govulncheck)" "$(($(date +%s) - _T0_GO_AUD))"
        else
          log_summary "Go" "govulncheck" "❌ Vulnerable" "$(get_version govulncheck)" "$(($(date +%s) - _T0_GO_AUD))"
          _OVERALL_EXIT_AUDIT=1
        fi
      fi
    elif is_ci_env; then
      log_error "govulncheck not found in CI. Failing Go audit."
      log_summary "Go" "govulncheck" "❌ Missing (Fatal)" "-" "0"
      _OVERALL_EXIT_AUDIT=1
    else
      log_warn "govulncheck not found locally. Skipping Go audit."
      log_summary "Go" "govulncheck" "⏭️  Skipped" "-" "0"
    fi
  fi

  # 8.2 Stack Specific (Rust)
  local _CA_BIN
  _CA_BIN=$(resolve_bin "cargo-audit") || true
  if [ -f "Cargo.toml" ]; then
    if [ -n "${_CA_BIN:-}" ]; then
      local _T0_RS_AUD
      _T0_RS_AUD=$(date +%s)
      log_info "\n── Auditing Rust dependencies (cargo audit) ──"
      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        log_success "DRY-RUN: Would run cargo audit"
        log_summary "Rust" "cargo-audit" "⚖️ Previewed" "-" "0"
      else
        local _RS_SPEC="${VER_CARGO_AUDIT_PROVIDER:-cargo-audit}@${VER_CARGO_AUDIT:-latest}"
        if run_quiet "${_G_UNIRTM_BIN:-unirtm}" exec "${_RS_SPEC:-}" -- cargo audit; then
          log_summary "Rust" "cargo-audit" "✅ Secure" "$(get_version cargo-audit)" "$(($(date +%s) - _T0_RS_AUD))"
        else
          log_summary "Rust" "cargo-audit" "❌ Vulnerable" "$(get_version cargo-audit)" "$(($(date +%s) - _T0_RS_AUD))"
          _OVERALL_EXIT_AUDIT=1
        fi
      fi
    elif is_ci_env; then
      log_error "cargo-audit not found in CI. Failing Rust audit."
      log_summary "Rust" "cargo-audit" "❌ Missing (Fatal)" "-" "0"
      _OVERALL_EXIT_AUDIT=1
    else
      log_warn "cargo-audit not found locally. Skipping Rust audit."
      log_summary "Rust" "cargo-audit" "⏭️  Skipped" "-" "0"
    fi
  fi

  # NOTE: Trivy CLI container audit removed (handled by GHA).
  # 9. SBOM Generation (CycloneDX) - Automatic Activation: run if Trivy is available.
  # 9. SBOM Generation (CycloneDX)
  local _TRIVY_BIN
  _TRIVY_BIN=$(resolve_bin "trivy") || true
  if [ -n "${_TRIVY_BIN:-}" ]; then
    local _T0_SBOM
    _T0_SBOM=$(date +%s)
    local _T_VER
    _T_VER=$(get_version "${_TRIVY_BIN:-}")
    local _R_VER
    _R_VER=$(get_unirtm_tool_version "trivy")

    log_info "\n── Generating SBOM (trivy fs) ──"
    log_info "Trivy: Using version ${_T_VER:-} (Required: ${_R_VER:-})"

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_success "DRY-RUN: Would generate CycloneDX SBOM"
      log_summary "Security" "sbom" "⚖️ Previewed" "-" "0"
    elif [ "${_T_VER:-}" != "${_R_VER:-}" ] && [ "${_R_VER:-}" != "latest" ]; then
      log_error "SECURITY ERROR: Trivy version mismatch! Found: ${_T_VER:-}, Expected: ${_R_VER:-}"
      log_error "This may indicate a compromised binary (Ref: March 2026 Incident)."
      log_summary "Security" "sbom" "⛔ Version Mismatch" "${_T_VER:-}" "0"
      _OVERALL_EXIT_AUDIT=1
    else
      if run_quiet "${_TRIVY_BIN:-}" fs --format cyclonedx --output sbom.json .; then
        log_summary "Security" "sbom" "✅ Generated" "${_T_VER:-}" "$(($(date +%s) - _T0_SBOM))"
        log_success "SBOM generated at sbom.json"

        # 9.1 SBOM Vulnerability Scan
        log_info "── Auditing SBOM for vulnerabilities ──"
        if run_quiet "${_TRIVY_BIN:-}" sbom sbom.json; then
          log_success "SBOM audit passed"
        else
          log_warning "SBOM contains known vulnerabilities (Review sbom.json)"
        fi
      else
        log_summary "Security" "sbom" "⚠️ Failed" "${_T_VER:-}" "$(($(date +%s) - _T0_SBOM))"
      fi
    fi
  else
    # Trivy CLI not installed - this is expected in CI (uses trivy-action instead)
    if is_ci_env; then
      log_info "⏭️  Trivy CLI not installed (CI uses trivy-action). Skipping SBOM generation."
      log_summary "Security" "sbom" "⏭️ Skipped (CI uses trivy-action)" "-" "0"
    else
      log_warn "trivy not found locally. Skipping SBOM generation."
      log_summary "Security" "sbom" "⏭️  Skipped" "-" "0"
    fi
  fi

  # 10. Binary Artifact Audit (Preventing Binary Poisoning)
  local _T0_BIN_AUD
  _T0_BIN_AUD=$(date +%s)
  log_info "\n── Auditing for unexpected binary artifacts ──"
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_success "DRY-RUN: Would scan for binary files"
    log_summary "Security" "binary-audit" "⚖️ Previewed" "-" "0"
  else
    # Look for common executable formats and archives that shouldn't be in source
    # Excluding .git, node_modules, .venv, and dist
    # We use -maxdepth 5 to keep it fast.
    _BIN_PATTERN="*.exe *.so *.dll *.dylib *.bin *.out *.elf *.o *.a *.lib"
    # shellcheck disable=SC2086
    _BIN_FOUND=$(find . -maxdepth 5 -not -path '*/.*' -not -path './node_modules/*' -not -path './vendor/*' -not -path './dist/*' -not -path "./${VENV:-.venv}/*" -type f \( -name "*.exe" -o -name "*.so" -o -name "*.dll" -o -name "*.dylib" -o -name "*.bin" -o -name "*.out" -o -name "*.elf" -o -name "*.o" -o -name "*.a" -o -name "*.lib" \))

    # Advanced: Detect 'Stealth Binaries' (non-text files masquerading as text)
    # This is a bit slow, so we only do a sampled check or check specific extensions
    # if 'file' command is available.
    if command -v file >/dev/null 2>&1; then
      # Scan for files that 'file' identifies as executable but don't have known allowed extensions
      _STEALTH_FOUND=$(find . -maxdepth 4 -not -path '*/.*' -not -path './node_modules/*' -not -path './dist/*' -not -path "./${VENV:-.venv}/*" -type f -exec file --mime {} + | grep -v "; charset=binary" | grep "application/x-executable\|application/x-sharedlib\|application/x-archive" | cut -d: -f1)
      if [ -n "${_STEALTH_FOUND:-}" ]; then
        _BIN_FOUND="${_BIN_FOUND:-}\n${_STEALTH_FOUND:-}"
      fi
    fi

    # Combine results and strip leading/trailing whitespace/newlines
    _FINAL_BIN_LIST=$(printf "%s\n%s" "${_BIN_FOUND:-}" "${_STEALTH_FOUND:-}" | tr -d '\r' | sed '/^[[:space:]]*$/d')

    if [ -z "${_FINAL_BIN_LIST:-}" ]; then
      log_summary "Security" "binary-audit" "✅ Clean" "-" "$(($(date +%s) - _T0_BIN_AUD))"
    else
      log_warning "Unexpected binary artifacts found:"
      echo "${_FINAL_BIN_LIST:-}"
      log_summary "Security" "binary-audit" "⚠️ Findings" "-" "$(($(date +%s) - _T0_BIN_AUD))"
    fi
  fi

  # ── Final Report ─────────────────────────────────────────────────────────────

  if [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    local _TOTAL_DUR_AUD
    _TOTAL_DUR_AUD=$(($(date +%s) - _START_TIME_AUDIT))
    printf "\n**Total Duration: %ss**\n" "${_TOTAL_DUR_AUD:-}" >>"${CI_STEP_SUMMARY:-}"

    # Add GitHub API rate limit info
    append_github_api_info

    printf "\n"
    finalize_summary_table
    log_info "\n✨ Audit complete!"

    if [ "${_OVERALL_EXIT_AUDIT:-}" -eq 0 ]; then
      log_success "\n✨ Security audit finished successfully."
      if [ "${DRY_RUN:-0}" -eq 0 ]; then
        printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
        printf "  - Run %bmake commit%b to record your verified changes.\n" "${GREEN:-}" "${NC:-}"
        printf "  - Run %bmake build%b to generate production artifacts.\n" "${GREEN:-}" "${NC:-}"
      fi
    else
      log_error "\n⚠️ Security audit finished with vulnerabilities or leaks found."
    fi
  fi

  exit "${_OVERALL_EXIT_AUDIT:-}"
}

main "$@"
