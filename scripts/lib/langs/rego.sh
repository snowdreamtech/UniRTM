#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Rego Logic Module

# Purpose: Installs OPA/Rego.
# Delegate: Managed by unirtm (.unirtm.toml)
install_rego() {
  install_tool_safe "opa" "${VER_OPA_PROVIDER:-}" "Rego (OPA)" "version" 0 "REGO" ""
}

# Purpose: Sets up Rego environment for project.
setup_rego() {
  install_rego
}
