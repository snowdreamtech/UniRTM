# Quickstart Validation: Native Lua & LuaRocks Provider

This guide outlines how to manually validate the `lua` provider natively downloading from LuaBinaries and bootstrapping LuaRocks, strictly without using ASDF.

## Prerequisites

- UniRTM binary built from the feature branch.
- Network access to `sourceforge.net` and `luarocks.org`.
- On Unix (macOS/Linux), standard `make` must be available in PATH for LuaRocks bootstrap.

## Validation Scenarios

### Scenario 1: Native Lua & LuaRocks Installation

1. Create an isolated directory and navigate to it:

   ```bash
   mkdir test-lua && cd test-lua
   ```

2. Create a `.unirtm.toml` with the `lua` configuration:

   ```bash
   cat <<EOF > .unirtm.toml
   [tools]
   lua = "5.4.2"
   "luarocks:busted" = "2.0.0"
   EOF
   ```

3. Run UniRTM to enforce strict ASDF bypassing:

   ```bash
   export UNIRTM_DISABLE_ASDF=true
   unirtm install
   ```

4. Verify Lua engine installation:

   ```bash
   unirtm exec lua -- lua -v
   # Expected Output: Lua 5.4.2  Copyright (C) 1994-2020 Lua.org, PUC-Rio
   ```

5. Verify LuaRocks bootstrap:

   ```bash
   unirtm exec lua -- luarocks --version
   # Expected Output: luarocks 3.11.x
   ```

### Scenario 2: Package Execution (Validates Integration)

1. Run the installed package:

   ```bash
   unirtm exec luarocks:busted -- busted --version
   # Expected Output: 2.0.0
   ```

*(This confirms LuaRocks correctly resolved the package, bound it to our native Lua environment, and generated the executable shim.)*
