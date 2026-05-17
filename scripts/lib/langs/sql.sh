#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# SQL Logic Module

# Purpose: Installs sqlfluff for SQL linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_sqlfluff() {
  install_tool_safe "sqlfluff" "${VER_SQLFLUFF_PROVIDER:-}" "Sqlfluff" "--version" 0 "*.sql" ""
}

# Purpose: Sets up SQL environment.
setup_sql() {
  install_sqlfluff
}
