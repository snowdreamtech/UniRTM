#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# OpenAPI Logic Module

# Purpose: Installs spectral for OpenAPI/AsyncAPI linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_spectral() {
  setup_registry_spectral
  install_tool_safe "spectral" "${VER_SPECTRAL_PROVIDER:-}" "Spectral" "--version" 0 "openapi.yaml openapi.json asyncapi.yaml asyncapi.json" ""
}

# Purpose: Sets up OpenAPI environment.
setup_openapi() {
  # Skip if no OpenAPI files exist (unless forced)
  if [ "${FORCE_SETUP:-0}" -eq 0 ]; then
    if ! has_lang_files "" "openapi.yaml openapi.json openapi.yml swagger.yaml swagger.json swagger.yml asyncapi.yaml asyncapi.json asyncapi.yml"; then
      log_info "⏭️  Skipping OpenAPI: No OpenAPI/Swagger/AsyncAPI files detected"
      log_summary "OpenAPI" "spectral" "⏭️ Skipped" "-" "0"
      return 0
    fi
  fi

  install_spectral
}
