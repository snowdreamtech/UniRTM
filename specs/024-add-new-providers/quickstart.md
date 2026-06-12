# Quickstart Validation Guide

To validate these changes end-to-end after implementation:

1. Create or modify your `.unirtm.toml` to include tools from these ecosystems:

```toml
[tools]
"composer:phpstan" = "latest"
"luarocks:luacheck" = "latest"
"pub:fvm" = "latest"
"cabal:shellcheck" = "latest"
```

1. Run the UniRTM install command:

```bash
unirtm install
```

1. Verify that the binaries were installed and are executable:

```bash
unirtm run phpstan --version
unirtm run luacheck --version
unirtm run fvm --version
unirtm run shellcheck --version
```

1. Verify isolation by checking your global system directories (e.g. `~/.composer/vendor/bin` or `~/.pub-cache/bin`) to ensure no binaries were unexpectedly leaked globally.
