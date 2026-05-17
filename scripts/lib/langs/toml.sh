#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# TOML Logic Module

# Purpose: Installs Taplo.
# Delegate: Managed by unirtm (.unirtm.toml)
install_taplo() {
  install_tool_safe "taplo" "${VER_TAPLO_PROVIDER:-}" "Taplo" "--version" 0 "*.toml" ""
}

# Purpose: Sets up TOML environment.
setup_toml() {
  install_taplo
}
