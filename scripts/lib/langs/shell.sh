#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Shell Logic Module

# Purpose: Installs Shfmt.
# Delegate: Managed by unirtm (.unirtm.toml)
install_shfmt() {
  install_tool_safe "shfmt" "${VER_SHFMT_PROVIDER:-}" "Shfmt" "--version" 0 "*.sh *.bash *.bats" ""
}

# Purpose: Installs Shellcheck.
# Delegate: Managed by unirtm (.unirtm.toml)
install_shellcheck() {
  install_tool_safe "shellcheck" "${VER_SHELLCHECK_PROVIDER:-}" "Shellcheck" "--version" 0 "*.sh *.bash *.bats" ""
}

# Purpose: Installs Actionlint.
# Delegate: Managed by unirtm (.unirtm.toml)
install_actionlint() {
  install_tool_safe "actionlint" "${VER_ACTIONLINT_PROVIDER:-}" "Actionlint" "--version" 0 "*.yml *.yaml" ".github/workflows"
}

# Purpose: Sets up Shell environment.
setup_shell() {
  install_shfmt
  install_shellcheck
  install_actionlint
}
