#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Protobuf Logic Module

# Purpose: Installs buf for Protobuf linting/management.
# Delegate: Managed by unirtm (.unirtm.toml)
install_buf() {
  setup_registry_buf
  install_tool_safe "buf" "${VER_BUF_PROVIDER:-}" "Buf" "--version" 0 "PROTOC" ""
}

# Purpose: Sets up Protobuf environment.
setup_protobuf() {
  # Skip if no Protobuf files exist (unless forced)
  if [ "${FORCE_SETUP:-0}" -eq 0 ]; then
    if ! has_lang_files "buf.yaml buf.gen.yaml" "*.proto"; then
      log_info "⏭️  Skipping Protobuf: No .proto files or buf config detected"
      log_summary "Protobuf" "buf" "⏭️ Skipped" "-" "0"
      return 0
    fi
  fi

  install_buf
}
