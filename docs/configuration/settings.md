# Global Settings

While `.unirtm.toml` manages project-level configuration, UniRTM also supports global settings that apply to all projects. These settings are configured in `~/.config/unirtm/config.toml`.

## Common Settings

```toml
[settings]
# Enable support for legacy version files like .nvmrc, .node-version, .python-version
legacy_version_file = true

# Always keep downloaded archives after extraction (useful for debugging)
always_keep_download = false

# How often to check for UniRTM self-updates (in hours, 0 to disable)
plugin_autoupdate_last_check_duration = "7d"

# Configure the default number of jobs for parallel downloads and compilation
jobs = 4

# Output verbosity (trace, debug, info, warn, error)
log_level = "info"
```

## Environment Variables
You can also override settings using environment variables by prefixing the setting name with `UNIRTM_`.

For example, to temporarily enable trace logging:
```bash
UNIRTM_LOG_LEVEL=trace unirtm install
```

To temporarily disable legacy version file parsing:
```bash
UNIRTM_LEGACY_VERSION_FILE=false unirtm run build
```
