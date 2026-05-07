# uninstall Command

## Overview

The `uninstall` command removes a specific version of a development tool from your system. This is a destructive operation that removes:

- The tool installation directory
- All tool binaries and files
- Shim scripts
- Database records

## Usage

```bash
unirtm uninstall <tool> <version> [flags]
```

## Arguments

- `<tool>` - The name of the tool to uninstall (e.g., `node`, `python`, `go`)
- `<version>` - The specific version to uninstall (e.g., `20.0.0`, `3.11.5`)

## Flags

- `-f, --force` - Skip confirmation prompt and proceed with uninstallation
- `-h, --help` - Display help information for the uninstall command

## Global Flags

- `-c, --config string` - Path to configuration file (default: `.unirtm.toml` or `unirtm.toml`)
- `-j, --json` - Enable JSON output format for scripting
- `-q, --quiet` - Enable quiet mode with minimal output
- `-v, --verbose` - Enable verbose output with debug logging

## Examples

### Basic Uninstall (with confirmation)

```bash
unirtm uninstall node 20.0.0
```

This will prompt for confirmation before uninstalling:

```
Tool to uninstall: node@20.0.0
Are you sure you want to uninstall node@20.0.0? This action cannot be undone. [y/N]:
```

### Force Uninstall (skip confirmation)

```bash
unirtm uninstall python 3.11.5 --force
```

This will uninstall without prompting for confirmation.

### Uninstall with JSON Output

```bash
unirtm uninstall go 1.21.0 --json --force
```

Output:

```json
{
  "level": "info",
  "message": "Successfully uninstalled go@1.21.0",
  "tool": "go",
  "version": "1.21.0",
  "duration": "125ms"
}
```

### Uninstall with Verbose Output

```bash
unirtm uninstall ruby 3.2.0 --verbose --force
```

This will display detailed debug information during the uninstallation process.

## Behavior

### Confirmation Prompt

By default, the uninstall command requires explicit user confirmation before proceeding. This is a safety measure to prevent accidental deletion of tools.

The confirmation prompt accepts:

- `y` or `yes` (case-insensitive) - Proceed with uninstallation
- `n`, `no`, or any other input - Cancel the operation

To skip the confirmation prompt, use the `--force` flag.

### Quiet Mode

When running in quiet mode (`--quiet` or `-q`), the confirmation prompt is automatically skipped, similar to using the `--force` flag.

### Uninstallation Process

The uninstall command performs the following steps:

1. **Validation** - Verifies that the tool and version are specified
2. **Database Check** - Confirms the tool version is installed
3. **Confirmation** - Prompts for user confirmation (unless `--force` is used)
4. **Provider Cleanup** - Runs tool-specific cleanup (e.g., removing virtual environments for Python)
5. **Directory Removal** - Removes the entire installation directory
6. **Database Cleanup** - Removes the installation record from the database
7. **Shim Cleanup** - Removes associated shim scripts

### Error Handling

The command will fail and display an error message if:

- The tool name or version is not specified
- The tool version is not installed
- The database cannot be accessed
- The installation directory cannot be removed
- The database record cannot be deleted

## Exit Codes

- `0` - Successful uninstallation or user cancelled
- `1` - Error occurred during uninstallation

## Requirements

This command implements the following requirements:

- **Requirement 8.2** - Explicit confirmation for destructive operations
- **Requirement 23.2** - CLI command implementation

## Related Commands

- [`install`](./install.md) - Install a specific version of a tool
- [`list`](./list.md) - List all installed tools
- [`doctor`](./doctor.md) - Verify system health and installed tools

## Notes

- Uninstallation is permanent and cannot be undone
- If you need the tool again, you must reinstall it using the `install` command
- The uninstall command does not affect other versions of the same tool
- Provider-specific cleanup ensures tool-specific resources are properly removed (e.g., npm global packages for Node.js, virtual environments for Python)

## Troubleshooting

### "Tool not found" Error

If you receive an error that the tool is not installed:

```bash
# List all installed tools to verify
unirtm list

# Check the exact version string
unirtm list | grep <tool>
```

### Permission Denied

If you encounter permission errors:

```bash
# Check the installation directory permissions
ls -la /opt/unirtm/tools/<tool>/<version>

# You may need to run with appropriate permissions
# or check the ownership of the installation directory
```

### Database Locked

If the database is locked:

```bash
# Ensure no other unirtm processes are running
ps aux | grep unirtm

# Wait a moment and try again
```

## See Also

- [Installation Guide](../installation.md)
- [Configuration Reference](../configuration.md)
- [CLI Reference](../cli-reference.md)
