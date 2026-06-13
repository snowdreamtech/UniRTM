# CLI Overview

UniRTM provides a rich, exhaustive CLI interface. Below is a complete catalog of all available commands. You can always run `unirtm help [command]` in your terminal to see specific flag usage.

<details>
<summary><code>unirtm activate</code></summary>

**Description**: Activate a tool version in the current shell

```text
Activate a tool version in the current shell.

The activate command generates a shell activation script that sets up
the environment for using the specified tool version. You can source
this script directly or add it to your shell configuration.

Examples:
  # Activate all tools (generate activation script for current shell)
  eval "$(unirtm activate)"

  # Activate for a specific shell
  unirtm activate --shell bash

  # Activate with project scope
  unirtm activate --scope project --project-dir /path/to/project

  # Activate a specific tool version
  unirtm activate node 20.0.0

Usage:
  unirtm activate [tool] [version] [flags]

Flags:
      --project-dir string   project directory for project-scoped activation (default: current directory)
      --scope string         activation scope (global or project) (default "global")
  -s, --shell string         shell type (bash, zsh, fish, powershell) — auto-detected if not specified
      --shims                use shims instead of dynamic PATH mode

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm alias</code></summary>

**Description**: Manage version aliases

```text
Manage version aliases for tools.

Aliases allow you to refer to a specific version by a name (e.g. "lts", "work").
They can be managed globally or at the project level using the --global flag.

If no subcommand is provided, it lists all aliases.

Usage:
  unirtm alias [flags]
  unirtm alias [command]

Aliases:
  alias, tool-alias

Available Commands:
  ls          List aliases
  resolve     Resolve an alias to its actual version
  set         Set an alias
  unset       Delete an alias

Flags:
      --global   manage global aliases (~/.config/unirtm/unirtm.toml)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm alias [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm backends</code></summary>

**Description**: Manage and list UniRTM backends

```text
Manage and list UniRTM backends.

Backends define where tools are downloaded from (GitHub Releases, Aqua
registry, HTTP URLs, etc.).

Sub-commands:
  ls    List all registered backends
  info  Show details about a specific backend

Examples:
  unirtm backends
  unirtm backends ls
  unirtm backends info github
  unirtm backends --json

Usage:
  unirtm backends [flags]
  unirtm backends [command]

Aliases:
  backends, b

Available Commands:
  info        Show details about a specific backend
  ls          List all registered backends

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm backends [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm bin-paths</code></summary>

**Description**: List all active runtime bin directories

```text
List all active runtime bin directories.

Outputs one directory per line — shims dir first, then each installed
tool's bin directory. Useful for shell hook scripts that need to prepend
the correct directories to PATH.

Examples:
  # Print all bin paths (one per line)
  unirtm bin-paths

  # JSON output
  unirtm bin-paths --json

  # Use in a shell script
  export PATH="$(unirtm bin-paths | tr '\n' ':')$PATH"

Usage:
  unirtm bin-paths [flags]

Aliases:
  bin-paths, bin

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm cache</code></summary>

**Description**: Manage UniRTM cache

```text
Manage the UniRTM download cache.

If no subcommand is provided, it displays the path to the cache directory.

Usage:
  unirtm cache [flags]
  unirtm cache [command]

Available Commands:
  clear       Clear all cache or a specific tool's cache
  list        List all cached artifacts
  path        Display the cache directory path
  prune       Remove expired cache entries
  stats       Display cache statistics

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm cache [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm completion</code></summary>

**Description**: Generate or install shell completion script

```text
Generate or install shell completion script for UniRTM.

By default, it auto-detects your current shell and prints the completion script.
Use the --install (-i) flag to automatically save the script and enable it in your shell configuration.
Use the --dir (-d) flag to export all four completion scripts to a specified directory.
  NOTE: --dir only writes files; it never modifies any shell configuration.
        --dir and --install/--uninstall are mutually exclusive.

Examples:
  # Auto-detect and print to stdout
  unirtm completion

  # Auto-detect and install persistently
  unirtm completion -i

  # Generate for a specific shell and print
  unirtm completion zsh

  # Generate all scripts, install only for shells present on the system
  unirtm completion -i --all

  # Export all four completion scripts to a directory (no shell config changes)
  unirtm completion -d ./completions

Usage:
  unirtm completion [bash|zsh|fish|powershell]

Flags:
  -a, --all          Generate/install/uninstall for all supported shells (zsh, bash, fish, powershell)
  -d, --dir string   Export all completion scripts to the specified directory (implies --all)
  -i, --install      Intelligently install completion script to your shell configuration
  -u, --uninstall    Intelligently uninstall completion script from your shell configuration

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm config</code></summary>

**Description**: Manage UniRTM configuration

```text
Manage UniRTM configuration settings.

This command manages settings that control UniRTM's behavior (e.g. cache TTL, data directory).
It is NOT for managing tools or environment variables (use 'set' or 'alias' for those).

If no subcommand is provided, it displays the current merged configuration.

Usage:
  unirtm config [flags]
  unirtm config [command]

Aliases:
  config, cfg

Available Commands:
  generate    Generate a default configuration file
  get         Get a configuration value
  set         Set a configuration value
  show        Display merged configuration
  validate    Validate configuration files

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm config [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm current</code></summary>

**Description**: Display the active version for each tool

```text
Display the active version for each tool.

This command aligns with 'mise current'. It looks at your configuration files
(unirtm.toml, .tool-versions) to determine which tools and versions are
currently requested in this directory.

If no tool is specified, it shows the active versions for all tools defined
in the current hierarchy.

Usage:
  unirtm current [tool] [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm deactivate</code></summary>

**Description**: Deactivate UniRTM from the current shell environment

```text
Deactivate UniRTM from the current shell environment.

The deactivate command generates a shell script that removes UniRTM shims
from the PATH and unsets environment variables set by the activate command.

Examples:
  # Deactivate from current shell
  eval "$(unirtm deactivate)"

  # Deactivate for a specific shell
  unirtm deactivate --shell bash

Usage:
  unirtm deactivate [flags]

Aliases:
  deactivate, deactive

Flags:
  -s, --shell string   shell type (bash, zsh, fish, powershell) — auto-detected if not specified

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm df</code></summary>

**Description**: Display the disk usage of unirtm data directories

```text
Display the disk usage of various folders within the unirtm data directory.

Usage:
  unirtm df [flags]

Flags:
  -h, --human-readable   print sizes in powers of 1024 (e.g., 1023M)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm disable</code></summary>

**Description**: Remove UniRTM automatic startup from your shell

```text
Safely removes the UniRTM activation command from your shell configuration file (like .zshrc or .bashrc).

This stops UniRTM from automatically managing your environment in future sessions without deleting your tools or data.
By default, it disables UniRTM, but you can specify 'mise' as an argument to disable mise instead.

Usage:
  unirtm disable [unirtm|mise] [flags]

Aliases:
  disable, dis

Flags:
  -a, --all   Disable for all supported shells (zsh, bash, fish, powershell)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm doctor</code></summary>

**Description**: Check system health and diagnose issues

```text
Check UniRTM system health and diagnose potential issues.

This command aligns with and exceeds 'mise doctor' to ensure your environment is
perfectly configured, providing deep insights into tools, paths, and connectivity.

Usage:
  unirtm doctor [flags]

Aliases:
  doctor, dr

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm edit</code></summary>

**Description**: Open the config file in $EDITOR

```text
Open a UniRTM config file in your preferred editor.

If no file is specified, it intelligently discovers all relevant config files
in the current hierarchy and lets you choose interactively.

Priority for finding an editor:
1.  --global / --file flag or argument
2.  UNIRTM_EDITOR environment variable
3.  VISUAL environment variable
4.  EDITOR environment variable
5.  UniRTM settings.editor configuration
6.  Standard system defaults (vim, nano, notepad, etc.)

Examples:
  # Interactively choose which config to edit
  unirtm edit

  # Edit the global config (~/.config/unirtm/unirtm.toml)
  unirtm edit --global

  # Edit a specific file
  unirtm edit .unirtm.toml

Usage:
  unirtm edit [FILE] [flags]

Flags:
  -f, --file string   Edit a specific config file
  -g, --global        Edit the global config file

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm enable</code></summary>

**Description**: Setup UniRTM to start automatically in your shell

```text
Setup UniRTM to start automatically by adding it to your shell's configuration file (like .zshrc or .bashrc).

This command auto-detects your shell and injects the activation command so you don't have to do it manually.
Once enabled, UniRTM will automatically manage your project environments whenever you open a new terminal window.

By default, it enables UniRTM, but you can specify 'mise' as an argument to enable mise instead.

Usage:
  unirtm enable [unirtm|mise] [flags]

Aliases:
  enable, en

Flags:
  -a, --all     Enable for all supported shells (zsh, bash, fish, powershell)
      --shims   Use shims mode for activation (default is path mode)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm env</code></summary>

**Description**: Export shell environment variables for activated tools

```text
Display or export the environment variables for the current UniRTM context.

When run in an interactive terminal, it provides a beautiful, data-rich dashboard
of your current environment, including PATH additions and variable sources.

When redirected or used with 'eval', it outputs shell-specific export statements
suitable for shell integration.

Examples:
  # Display interactive environment dashboard
  unirtm env

  # Export variables to current shell
  eval "$(unirtm env)"

  # Get environment in JSON format
  unirtm env --json

Usage:
  unirtm env [flags]

Aliases:
  env, e

Flags:
      --info           print build/version info instead of tool environment
      --shell string   shell format (bash, zsh, fish, nu, powershell). Default: auto-detect

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm exec</code></summary>

**Description**: Execute a command with tool environment variables set

```text
Execute a command with one or more tools set.

Use this to run ad-hoc commands with specific tool versions without modifying
the current shell session. Tools are loaded from the nearest unirtm.toml and
can be overridden by passing tool@version specifiers before "--".

The "--" separator distinguishes tool specifiers from the command to execute.
If "--" is omitted, all positional arguments are treated as the command itself
(loose mode, no tool overrides — compatible with mise).

Examples:
  # Run node v20 with an explicit version override
  unirtm exec node@20 -- node --version

  # Combine multiple tool overrides
  unirtm exec node@20 python@3.11 -- bash -c "node -v && python -V"

  # Shell-string mode (analogous to mise exec --command "...")
  unirtm exec node@20 -x "node --version && npm -v"

  # Use tools already declared in unirtm.toml
  unirtm exec -- make build

  # Alias: 'x'
  unirtm x node@20 -- node server.js

  # Dry-run: see what would be executed
  unirtm exec --dry-run node@20 -- node app.js

Usage:
  unirtm exec [tool@version...] [-- <command> [args...]] [flags]

Aliases:
  exec, x

Flags:
  -x, --exec-command string   execute this shell command string (analogous to mise exec --command)
      --fresh-env             bypass environment cache and recompute the environment
      --no-deps               skip automatic dependency preparation
      --raw                   pipe stdin/stdout/stderr directly (sets --jobs=1)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm fmt</code></summary>

**Description**: Format configuration files with canonical key ordering

```text
Format configuration files with canonical key ordering and indentation.

Reads UniRTM configuration files, normalizes section ordering, and writes them
back in-place. Use --check in CI to verify formatting without modifying files.

Canonical Section Order:
  [env] -> [tools] -> [tasks] -> [settings] -> [plugins] -> [alias]

Examples:
  # Format the project config file
  unirtm fmt

  # Format all configs in current and subdirectories
  unirtm fmt -r

  # CI mode: exit 1 if files are not formatted
  unirtm fmt --check

Usage:
  unirtm fmt [flags]

Flags:
  -a, --all         format all supported config files (unirtm.toml, .tool-versions)
      --check       check if config is formatted without modifying it (exit 1 if not)
  -r, --recursive   format all unirtm.toml files in subdirectories

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm generate</code></summary>

**Description**: Generate integration files (GitHub Actions, pre-commit hooks, etc.)

```text
Generate integration files for common tooling.

Sub-commands:
  github-action   Generate a GitHub Actions workflow step
  gitlab-ci       Generate a GitLab CI script snippet
  dockerfile      Generate a Dockerfile snippet for UniRTM
  pre-commit      Generate a .pre-commit-hooks.yaml snippet
  shell-alias     Print shell alias definitions
  manpage         Generate manpages

Examples:
  unirtm generate github-action
  unirtm generate gitlab-ci
  unirtm generate dockerfile
  unirtm generate pre-commit --output .pre-commit-hooks.yaml
  unirtm generate shell-alias >> ~/.zshrc
  unirtm generate manpage -d ./manpages

Usage:
  unirtm generate [flags]
  unirtm generate [command]

Aliases:
  generate, gen

Available Commands:
  dependabot    Auto-generate dependabot.yml from detected ecosystems
  dockerfile    Generate a Dockerfile snippet for UniRTM
  github-action Generate a GitHub Actions workflow snippet for UniRTM
  gitlab-ci     Generate a GitLab CI script snippet for UniRTM
  manpage       Generate manpages for UniRTM
  pre-commit    Generate a pre-commit hook snippet for UniRTM
  shell-alias   Print shell alias definitions for common UniRTM commands

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm generate [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm help</code></summary>

**Description**: Help about any command

```text
Help provides help for any command in the application.
Simply type unirtm help [path to command] for full details.

Usage:
  unirtm help [command] [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm hook</code></summary>

**Description**: Manage and execute git hooks

```text
Manage git hooks using UniRTM's multi-engine runner.
This allows you to seamlessly integrate pre-commit, husky, lefthook, or native UniRTM hooks.

Usage:
  unirtm hook [command]

Available Commands:
  install     Install a git hook bridge script
  run         Run a specific git hook

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm hook [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm hook-exec</code></summary>

**Description**: Execution wrapper specifically designed for git hooks

```text
[31m✗[0m [31mcommand not found: --help (checked UniRTM tools and PATH)[0m
```

</details>

<details>
<summary><code>unirtm implode</code></summary>

**Description**: Completely remove all UniRTM data and tool installations

```text
Completely remove all UniRTM data and tool installations.

This command will internal-combust and erase:
  • All tool installations (binaries)
  • All shims and wrapper scripts
  • All download caches and temporary files
  • The central SQLite database
  • All external plugins
  • (Optional) Your configuration directory (~/.config/unirtm)

WARNING: This action is permanent and IRREVERSIBLE.

Usage:
  unirtm implode [flags]

Flags:
      --config   also remove configuration directory (~/.config/unirtm)
  -y, --yes      skip confirmation prompt

Global Flags:
  -C, --cd string    change directory before running command
      --dry-run      show what would happen without making changes
  -E, --env string   set the environment for loading configuration
      --help         help for this command
      --jobs int     how many jobs to run in parallel (default 8)
  -j, --json         enable JSON output format
      --locked       require lockfile URLs to be present during installation
      --no-config    do not load any config files
      --no-env       do not load environment variables from config files
      --no-hooks     do not execute hooks from config files
  -q, --quiet        enable quiet mode (minimal output)
      --silent       suppress all task output and non-error messages
  -V, --verbose      enable verbose output (debug logging)
```

</details>

<details>
<summary><code>unirtm index</code></summary>

**Description**: Manage the tool index

```text
Manage the local tool index used for searching available tools.

Usage:
  unirtm index [command]

Available Commands:
  clear       Clear the local tool index cache
  status      Show status of the local tool index
  update      Refresh the tool index from remote backends

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm index [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm install</code></summary>

**Description**: Install development tools and package specifications

```text
Install development tools and package specifications.

The install command downloads and installs one or more specified versions of tools
using the appropriate backends. It validates the installation, records it in
the database, and generates shim scripts.

Examples:
  # Install Node.js version 20.0.0
  unirtm install node 20.0.0

  # Install Python version 3.11.5 using a specific backend
  unirtm install python 3.11.5 --backend github

  # Install multiple tools concurrently with package syntax
  unirtm install node@20.11.0 go@1.22.1 python@3.12.0

  # Force reinstall a tool even if already installed
  unirtm install node@20.11.0 --force

  # Install with JSON output
  unirtm install go 1.21.0 --json

Usage:
  unirtm install [tool[@version]...] [flags]

Aliases:
  install, i

Flags:
  -b, --backend string   backend to use for installation (default: auto-detect)
  -f, --force            force installation even if already installed

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm install-into</code></summary>

**Description**: Install a tool into a custom directory

```text
Install a tool into a custom directory.

Installs the specified tool into a custom directory instead of the default
UniRTM installs directory. Useful for placing tools in shared locations
(e.g. /usr/local) or project-local directories.

Examples:
  # Install Go 1.22.0 into /usr/local/go
  unirtm install-into golang/go@1.22.0 /usr/local/go

  # Install with a specific backend
  unirtm install-into node@22.14.0 ./local/node --backend native

Usage:
  unirtm install-into <tool@version> <directory> [flags]

Flags:
  -b, --backend string   backend to use (default: auto-detect)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm latest</code></summary>

**Description**: Get the latest available version for a tool

```text
Get the latest available version for a tool from its backend.

When a version prefix is provided, returns the latest patch release within
that prefix (e.g. "1.22" returns the latest 1.22.x release).

Examples:
  # Latest release of the GitHub CLI
  unirtm latest cli/cli

  # Latest 1.22.x release of Go
  unirtm latest golang/go 1.22

  # JSON output
  unirtm latest cli/cli --json

  # Use a specific backend
  unirtm latest typescript --backend npm

Usage:
  unirtm latest <tool> [version-prefix] [flags]

Flags:
  -b, --backend string   backend to use (default: auto-detect)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm license</code></summary>

**Description**: Manage copyright license headers in source files

```text
license ensures source code files contain copyright license headers.

Subcommands:
  add    Add missing license headers to source files
  check  Report source files that are missing a license header

Examples:
  unirtm license add   -f .github/license-header.txt cmd/ internal/ scripts/
  unirtm license check -f .github/license-header.txt cmd/ internal/ scripts/
  unirtm license add   -l MIT -c "SnowdreamTech" -y 2026 cmd/

Usage:
  unirtm license [command]

Available Commands:
  add         Add missing license headers to source files (Safe Mode)
  check       Check for missing license headers (exits non-zero if any found)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm license [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm link</code></summary>

**Description**: Register an externally installed tool into UniRTM

```text
Register an externally installed tool into UniRTM management.

Links an existing binary or directory into UniRTM's database so it
appears in 'unirtm list' and can be used with 'unirtm use'.
No files are moved or copied.

Examples:
  # Register a locally compiled Go binary
  unirtm link golang/go 1.22.0 /usr/local/go

  # Register with a specific backend label
  unirtm link node 22.14.0 ~/.nvm/versions/node/v22.14.0 --backend nvm

Usage:
  unirtm link <tool> <version> <path> [flags]

Aliases:
  link, ln

Flags:
  -b, --backend string   backend name to record in the database (default "custom")

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm list</code></summary>

**Description**: List all installed development tools

```text
List all installed development tools.

Shows all tools installed with UniRTM, their version, backend, activation
status, and installation path. The STATUS column shows whether a version
is currently active (its shim points to it).

Examples:
  # List all installed tools
  unirtm list

  # Filter by tool name
  unirtm list --tool node

  # JSON output
  unirtm list --json

Usage:
  unirtm list [flags]

Aliases:
  list, ls

Flags:
      --current       only show currently active versions
  -t, --tool string   filter by tool name

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm lock</code></summary>

**Description**: Generate or update the unirtm.lock lockfile

```text
Generate or update the unirtm.lock lockfile.

unirtm.lock pins exact tool versions, download URLs, and checksums for
reproducible installs across environments. It is UniRTM's equivalent of
package-lock.json (npm) or Cargo.lock (Rust).

Key benefits:
  • Reproducible builds: everyone uses exactly the same version
  • Avoids API rate limits: URLs are cached, so GitHub API is not called on install
  • Security: checksums are verified against the lockfile on each install
  • Offline installs: combine with UNIRTM_LOCKED=1 for fully offline CI

Once generated, unirtm install automatically uses the lockfile and keeps it
up to date after each successful installation.

Examples:
  # Generate / refresh the lockfile for the current platform
  unirtm lock

  # Generate entries for all standard platforms (for CI reproducibility)
  unirtm lock --all-platforms

  # Generate only for specific platforms
  unirtm lock --platform linux-amd64,macos-arm64

  # Refresh only selected tools
  unirtm lock cli/cli astral-sh/ruff

  # Validate the lockfile without regenerating (CI gate)
  unirtm lock --check

Environment variables:
  UNIRTM_LOCK_FILE   Override lockfile path (default: ./unirtm.lock)
  UNIRTM_LOCKED=1    Strict mode: installs fail unless URL is in lockfile

Usage:
  unirtm lock [tool...] [flags]

Flags:
      --all-platforms     generate lockfile entries for all standard platforms
      --check             validate the lockfile without regenerating (exits non-zero on problems)
      --platform string   comma-separated list of platform keys (e.g. linux-amd64,macos-arm64)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm ls-remote</code></summary>

**Description**: List runtime versions available for install

```text
List runtime versions available for install from the backend.

The results are fetched from the remote backend and may be cached locally.
Installed versions are highlighted with a green ✓ checkmark.

Examples:
  # List all available versions of Node.js
  unirtm ls-remote node

  # List the latest 20 versions
  unirtm ls-remote node --limit 20

  # List versions matching a prefix
  unirtm ls-remote node 20

  # Use a specific backend
  unirtm ls-remote typescript --backend npm

  # JSON output
  unirtm ls-remote node --json

Usage:
  unirtm ls-remote <tool> [version-prefix] [flags]

Aliases:
  ls-remote, lsr

Flags:
  -b, --backend string   backend to use for listing versions (default: auto-detect)
  -l, --limit int        limit the number of versions displayed (0 = all)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm mcp</code></summary>

**Description**: Run an MCP stdio server exposing UniRTM tools to AI agents (experimental)

```text
Run an MCP (Model Context Protocol) stdio server.

Starts a JSON-RPC 2.0 stdio server that AI agents (Claude, Gemini, etc.)
can call to install, list, and manage tools through the MCP protocol.

Exposed tools:
  list_tools     — list all installed tools
  install_tool   — install a tool by name and version
  outdated_tools — list tools with newer versions available
  tool_info      — get info about a specific tool

This command is EXPERIMENTAL. The protocol and tool signatures may change.

Examples:
  # Start the MCP server (typically called by the AI host)
  unirtm mcp

Usage:
  unirtm mcp [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm migrate</code></summary>

**Description**: Migrate from mise or asdf configuration with visual side-by-side diff

```text
Convert mise or asdf configuration to UniRTM format.

Supported source formats:
  • .mise.toml     — mise configuration file
  • mise.toml      — mise configuration file
  • .tool-versions — asdf/mise version pin file

If no source file is specified, the current directory is scanned
automatically. The migration report shows a visual side-by-side
comparison of what was found in the source and what will be written.

Examples:
  unirtm migrate
  unirtm migrate .mise.toml
  unirtm migrate .tool-versions --output .unirtm.toml
  unirtm migrate --dry-run

Usage:
  unirtm migrate [source] [flags]

Flags:
      --dry-run         preview migration without writing files
  -o, --output string   output file path (default ".unirtm.toml")

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm outdated</code></summary>

**Description**: Show installed tools that have newer versions available

```text
Show installed tools that have newer versions available.

Queries each backend for the latest available version and compares it to
what is currently installed. Useful before running 'unirtm update'.

Use --interactive / -i to select which outdated tools to upgrade immediately.

Examples:
  # Check all installed tools
  unirtm outdated

  # Check specific tool(s)
  unirtm outdated cli/cli

  # Interactively select tools to upgrade
  unirtm outdated --interactive

  # JSON output
  unirtm outdated --json

Usage:
  unirtm outdated [tool...] [flags]

Flags:
  -i, --interactive   interactively select tools to upgrade

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm plugin</code></summary>

**Description**: Manage UniRTM plugins

```text
Manage UniRTM backend and provider plugins.

UniRTM supports Go-native plugins that extend its functionality with
custom backends (tool sources) and providers (tool-specific install logic).
Plugins are standalone executable binaries prefixed with 'unirtm-plugin-'
placed in the plugins data directory.

Use 'unirtm plugin list' to see loaded plugins.

Usage:
  unirtm plugin [flags]
  unirtm plugin [command]

Aliases:
  plugin, p, plugins

Available Commands:
  install     Install a plugin from a local executable file
  list        List loaded plugins
  remove      Remove an installed plugin

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm plugin [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm prepare</code></summary>

**Description**: Ensure project dependencies are ready by running applicable prepare steps

```text
Ensure project dependencies are ready by running applicable prepare steps.

This command:
1. Loads the project configuration (e.g., unirtm.toml).
2. Parallel downloads and installs any missing tools configured for this project.
3. Detects other package managers (like npm/pnpm, Go, Python) and prints preparation checks.

Examples:
  # Prepare everything for the current project
  unirtm prepare

  # Prepare only a specific tool
  unirtm prepare node

Usage:
  unirtm prepare [tool] [flags]

Aliases:
  prepare, prep

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm prune</code></summary>

**Description**: Remove unused (non-latest) tool installations and show freed disk space

```text
Remove unused tool installations to free up disk space.

The prune command identifies tool versions that are installed but are not
the latest version of their tool, then removes them. The latest installed
version of each tool is always kept.

After pruning, the total amount of freed disk space is reported.

Examples:
  # Preview what would be pruned (dry-run)
  unirtm prune --dry-run

  # Prune all unused versions (with confirmation)
  unirtm prune

  # Prune without confirmation
  unirtm prune --yes

  # Prune only node versions
  unirtm prune --tool node

Usage:
  unirtm prune [flags]

Flags:
  -t, --tool string   limit prune to a specific tool
  -y, --yes           skip confirmation prompt

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
```

</details>

<details>
<summary><code>unirtm registry</code></summary>

**Description**: List available tools in the UniRTM registry with descriptions and homepages

```text
List available tools in the UniRTM registry.

Shows all tools that can be installed via UniRTM, including the backend source,
native provider, human-readable description, and homepage URL for well-known tools.

Examples:
  # List all available tools
  unirtm registry

  # Filter by name or description
  unirtm registry --search go

  # JSON output
  unirtm registry --json

Usage:
  unirtm registry [flags]

Aliases:
  registry, ls

Flags:
  -s, --search string   filter tools by name or description

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm reshim</code></summary>

**Description**: Regenerate shim scripts for installed tools (concurrent + dead shim cleanup)

```text
Regenerate shim scripts for all installed tools using concurrent workers.

This command recreates the shim scripts in the shims directory in parallel,
making it significantly faster than sequential regeneration for large tool sets.
Use --clean to also remove dead shims for tools that are no longer installed.

Examples:
  # Regenerate all shims (parallel)
  unirtm reshim

  # Regenerate shims for a specific tool
  unirtm reshim --tool node

  # Regenerate and clean up dead shims
  unirtm reshim --clean

  # Preview what would happen
  unirtm reshim --clean --dry-run

Usage:
  unirtm reshim [flags]

Flags:
      --clean         also remove dead shims for uninstalled tools
  -t, --tool string   limit reshim to a specific tool

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm run</code></summary>

**Description**: Run a task using the multi-modal routing engine

```text
Run a task using the multi-modal routing engine.

UniRTM delegates tasks to professional tools (go-task, make, just)
if their configuration files are detected, or falls back to executing
tasks defined in unirtm.toml.

If no task is provided, it lists all available tasks.

Examples:
  # List all available tasks
  unirtm run

  # Run a build task
  unirtm run build

  # Run a task with arguments
  unirtm run test -- -v

  # Run with custom output mode
  unirtm run test --output interleaved

Usage:
  unirtm run [task] [args...] [flags]

Aliases:
  run, r

Flags:
      --fix             Apply automatic fixes if supported by the task
      --output string   Task output mode: spinner, prefix, or interleaved (default: auto-detect)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm search</code></summary>

**Description**: Search for development tools and optionally install on match

```text
Search for available development tools in the tool index.

The search command queries the local tool index by name or description.
Run 'unirtm index update' to refresh the index from remote backends.

Use --install (-i) to get an interactive prompt to install a tool right
after seeing the search results — no need to run a separate install command.

Examples:
  # Search for Node.js
  unirtm search node

  # Search and prompt to install immediately
  unirtm search go --install

  # Filter by backend
  unirtm search python --backend github

  # JSON output
  unirtm search go --json

  # Limit results
  unirtm search tool --limit 10

Usage:
  unirtm search <query> [flags]

Flags:
  -b, --backend string   filter by backend type (github, aqua, http)
  -i, --install          prompt to install a matching tool after search
  -l, --limit int        maximum number of results to display (default 50)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm self-update</code></summary>

**Description**: Update UniRTM to the latest (or specified) version

```text
Update UniRTM to the latest (or specified) version.

Checks the latest release on GitHub, displays release notes, and
prompts before installing. On Linux/macOS uses the install script;
on Windows uses PowerShell.

Examples:
  # Update to latest
  unirtm self-update

  # Update without prompting
  unirtm self-update --yes

  # Update to a specific version
  unirtm self-update --version v1.2.3

Usage:
  unirtm self-update [flags]

Aliases:
  self-update, self-upgrade, self-up

Flags:
      --version string   target version to update to (default: latest)
  -y, --yes              skip confirmation prompt

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
```

</details>

<details>
<summary><code>unirtm set</code></summary>

**Description**: Set environment variables in the config file

```text
Set environment variables in the config file.

Variables are written to the [env] section of unirtm.toml in the current
directory (or --global for the global config).

Examples:
  # Set a single variable
  unirtm set NODE_ENV=production

  # Set multiple variables at once
  unirtm set FOO=bar BAZ=qux

  # Write to global config
  unirtm set --global GITHUB_TOKEN=ghp_xxxx

Usage:
  unirtm set <KEY=value>... [flags]

Flags:
  -e, --env string    environment-specific config file (e.g. unirtm.<env>.toml)
  -g, --global        write to global config (~/.config/unirtm/unirtm.toml)
  -p, --path string   directory to write config file into (default: current directory)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm settings</code></summary>

**Description**: Manage UniRTM settings (compatibility alias for config)

```text
Manage UniRTM configuration settings.
This command is provided for compatibility with 'mise settings' and acts as a smart wrapper around 'unirtm config'.

Behaviors:
  0 args: unirtm settings               -> unirtm config show
  1 arg:  unirtm settings <key>         -> unirtm config get <key>
  2 args: unirtm settings <key> <value> -> unirtm config set <key> <value>

Examples:
  unirtm settings
  unirtm settings settings.cache_ttl
  unirtm settings settings.cache_ttl 48h

Usage:
  unirtm settings [key] [value] [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm shell</code></summary>

**Description**: Set tool version for the current shell session

```text
Set tool version(s) for the current shell session only.

The shell command outputs shell-specific export statements that you can
eval to override the active tool version for the current shell session.
Unlike 'activate', these settings are not persisted and only affect the
current shell.

Examples:
  # Set node version for current shell session
  eval "$(unirtm shell node@20.0.0)"

  # Set multiple tools
  eval "$(unirtm shell node@20.0.0 python@3.11.0)"

  # Show what would be exported (without eval)
  unirtm shell node@20.0.0

Usage:
  unirtm shell <tool>@<version> [tool@version ...] [flags]

Aliases:
  shell, sh

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm sync</code></summary>

**Description**: Synchronize tools from other version managers with UniRTM

```text
Synchronize tools from other version managers with UniRTM.

Autodetects existing versions installed via nvm, pyenv, rbenv, or asdf,
symlinks them into UniRTM's installs directory, and automatically registers
them in UniRTM's database with generated shims.

Examples:
  # Scan and sync all detected tools (node, python, ruby)
  unirtm sync

  # Scan and sync node versions only
  unirtm sync node

  # Scan and sync python versions only
  unirtm sync python

  # Scan and sync ruby versions only
  unirtm sync ruby

Usage:
  unirtm sync [tool] [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm tasks</code></summary>

**Description**: Manage and inspect tasks defined in the config file

```text
Manage and inspect tasks defined in the config file.

Tasks are defined in the [tasks] section of unirtm.toml.
Use 'unirtm run <task>' to execute them.

Sub-commands:
  list   List all defined tasks
  info   Show details about a specific task
  deps   Show task dependency graph
  edit   Open a task in $EDITOR

Examples:
  unirtm tasks
  unirtm tasks list
  unirtm tasks info build
  unirtm tasks deps
  unirtm tasks edit test

Usage:
  unirtm tasks [flags]
  unirtm tasks [command]

Aliases:
  tasks, t

Available Commands:
  deps        Show task dependency graph
  edit        Open task definition in $EDITOR
  info        Show details about a specific task
  list        List all tasks defined in the config file

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts

Use "unirtm tasks [command] --help" for more information about a command.
```

</details>

<details>
<summary><code>unirtm test-tool</code></summary>

**Description**: Test a tool installs and executes

```text
Test a tool installs and executes by downloading it and running a version/help check.

This command installs the specified tools (or all tools in the configuration if no arguments
are provided) and then attempts to execute their provided binaries to verify that they run
correctly without dynamic linking or architecture issues.

Examples:
  # Test all tools in the current unirtm.toml
  unirtm test-tool

  # Test GitHub CLI
  unirtm test-tool cli/cli

  # Test a specific version
  unirtm test-tool node@20.0.0
  unirtm test-tool node 20.0.0

Usage:
  unirtm test-tool [tool[@version]...] [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm token</code></summary>

**Description**: Show tokens from environment variables

```text
Show tokens from environment variables

Usage:
  unirtm token [provider] [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm tool</code></summary>

**Description**: Show information about a specific tool

```text
Show detailed information about a specific tool.

Displays backend, installed versions, active version, shim location,
and the config file that sets the active version.

Examples:
  # Show info for GitHub CLI
  unirtm tool cli/cli

  # JSON output
  unirtm tool cli/cli --json

  # Specify backend explicitly
  unirtm tool github:cli/cli

Usage:
  unirtm tool <tool> [flags]

Flags:
      --active          Only show active versions
      --backend         Only show backend field
      --config-source   Only show config source
      --description     Only show description field
      --installed       Only show installed versions
      --requested       Only show requested versions
      --tool-options    Only show tool options

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm tool-stub</code></summary>

**Description**: Execute a tool stub

```text
[31m✗[0m [31mfailed to read stub file: open --help: no such file or directory[0m
```

</details>

<details>
<summary><code>unirtm trust</code></summary>

**Description**: Mark a configuration file as trusted

```text
Marks all project-related configuration files as trusted.
Trusted files are allowed to be automatically loaded and their environment variables applied.
If no path is provided, it automatically trusts all configuration files in the current directory.

Use --all (-a) to also trust the global configuration file (~/.config/unirtm/unirtm.toml).
Use --list (-l) to view the current project's trusted configuration files.
Use --list --all (-la) to view all globally trusted configuration files.

Usage:
  unirtm trust [path] [flags]

Flags:
  -a, --all    also trust the global config file (~/.config/unirtm/unirtm.toml)
  -l, --list   list trusted configuration files (current project by default)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm uninstall</code></summary>

**Description**: Uninstall development tools and package specifications

```text
Uninstall development tools and package specifications.

The uninstall command removes specified versions of tools, including:
- The tool installation directory
- Shim scripts
- Database records

This is a destructive operation and requires explicit confirmation unless
the --force flag is used.

Examples:
  # Uninstall Node.js version 20.0.0 (with confirmation)
  unirtm uninstall node 20.0.0

  # Uninstall Python version 3.11.5 without confirmation
  unirtm uninstall python 3.11.5 --force

  # Uninstall multiple tools concurrently
  unirtm uninstall node@20.11.0 go@1.22.1 python@3.12.0 --force

  # Uninstall with JSON output
  unirtm uninstall go 1.21.0 --json --force

Usage:
  unirtm uninstall [tool[@version]...] [flags]

Aliases:
  uninstall, un

Flags:
  -f, --force   skip confirmation prompt

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm unset</code></summary>

**Description**: Remove environment variables from the config file

```text
Remove environment variables from the config file.

Deletes the specified key(s) from the [env] section of unirtm.toml.

Examples:
  # Remove a variable
  unirtm unset NODE_ENV

  # Remove from global config
  unirtm unset --global GITHUB_TOKEN

Usage:
  unirtm unset <KEY>... [flags]

Flags:
  -e, --env string    environment-specific config file (e.g. unirtm.<env>.toml)
  -g, --global        write to global config (~/.config/unirtm/unirtm.toml)
  -p, --path string   directory to write config file into (default: current directory)

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm untrust</code></summary>

**Description**: Revoke trust from a configuration file

```text
Revokes trust from a previously trusted configuration file.
Once untrusted, the file's environment variables and configuration will no longer be automatically loaded.

Usage:
  unirtm untrust [path] [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm unuse</code></summary>

**Description**: Remove tools from the config file without uninstalling

```text
Remove tools from the config file without uninstalling them.

Deletes the specified tool(s) from the [tools] section of unirtm.toml
and removes the corresponding database record. The installed files are
kept on disk — use 'unirtm uninstall' to also delete the binaries.

Examples:
  # Remove a specific version
  unirtm unuse cli/cli@2.70.0

  # Remove a tool (removes its DB record regardless of version)
  unirtm unuse node

  # Aliases
  unirtm rm cli/cli
  unirtm remove node

Usage:
  unirtm unuse <tool[@version]>... [flags]

Aliases:
  unuse, rm, remove

Flags:
      --all           remove all versions of the tool from the config
  -e, --env string    environment-specific config file
  -g, --global        remove from global config (~/.config/unirtm/unirtm.toml)
  -p, --path string   directory of config file to edit

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm update</code></summary>

**Description**: Update installed development tools

```text
Update installed development tools to newer versions.

The update command checks for available updates and applies them.
It respects version constraints defined in configuration files.

Examples:
  # Preview all available updates (no changes applied)
  unirtm update --preview

  # Update a specific tool to its latest version
  unirtm update node

  # Update a specific tool to a specific version
  unirtm update node 22.0.0

  # Update all installed tools
  unirtm update --all

  # Preview what would be updated for all tools (explicit)
  unirtm update --all --preview

  # Update with JSON output
  unirtm update node --json

Usage:
  unirtm update [tool] [version] [flags]

Aliases:
  update, up

Flags:
  -a, --all       update all installed tools
  -f, --force     skip confirmation prompt
  -p, --preview   show what would be updated without applying changes

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm upgrade</code></summary>

**Description**: Upgrade installed development tools (alias for update)

```text
Upgrade installed development tools to newer versions.

This command is exactly the same as 'unirtm update' and is provided for compatibility.

Usage:
  unirtm upgrade [tool] [version] [flags]

Aliases:
  upgrade, upg

Flags:
  -a, --all       update all installed tools
  -f, --force     skip confirmation prompt
  -p, --preview   show what would be updated without applying changes

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm use</code></summary>

**Description**: Set a tool version in the config file

```text
Set a tool version in the configuration file.

The use command writes a tool version declaration into the nearest
unirtm.toml or .unirtm.toml configuration file. If neither file exists
it creates unirtm.toml in the target directory.

Examples:
  # Set node version in current directory's config
  unirtm use node@20.0.0

  # Set globally (writes to ~/.config/unirtm/config.toml)
  unirtm use node@20.0.0 --global

  # Set in a specific directory
  unirtm use node@20.0.0 --path /path/to/project

  # Set multiple tools
  unirtm use node@20.0.0 python@3.11.0

Usage:
  unirtm use <tool>@<version> [flags]

Aliases:
  use, u

Flags:
  -e, --env string    environment-specific config file (e.g. unirtm.<env>.toml)
  -f, --force         force reinstall even if the tool is already installed
  -g, --global        write to global config (~/.config/unirtm/config.toml)
  -p, --path string   directory to write config file into (default: current directory)
  -P, --pin           resolve prefix/alias/range to precise concrete version

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm version</code></summary>

**Description**: Print the version number of unirtm

```text
Display version information including build details, commit hash, and build time.

Usage:
  unirtm version [flags]

Aliases:
  version, v

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm watch</code></summary>

**Description**: Watch files and run task on changes

```text
Watch files in the current directory and automatically run the specified task when changes occur.

Usage:
  unirtm watch <task> [flags]

Aliases:
  watch, w

Flags:
      --clear            Clear the screen before running the task
  -d, --debounce int     Debounce delay in milliseconds (default 500)
  -g, --glob strings     Watch only files matching these glob patterns (e.g. *.go)
  -i, --ignore strings   Ignore files matching these patterns
  -s, --shell            Run task inside a shell

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm where</code></summary>

**Description**: Display the installation path of a tool

```text
Display the installation path of a specific tool version.

The where command prints the directory where a tool version is installed.
If no version is specified, it returns the path of the latest installed version.

Examples:
  # Show installation path of node (latest installed)
  unirtm where node

  # Show installation path of a specific version
  unirtm where node 20.0.0

Usage:
  unirtm where <tool> [version] [flags]

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>

<details>
<summary><code>unirtm which</code></summary>

**Description**: Display the path to the tool binary

```text
Display the full path to the binary of a tool.

The which command searches the install_path of the tool in the database,
then looks for the binary in common bin/ sub-directories.

Examples:
  # Show binary path of node (latest installed)
  unirtm which node

  # Show binary path of a specific version
  unirtm which node 20.0.0

Usage:
  unirtm which <tool> [version] [flags]

Flags:
  -a, --all           Show all matches
      --plugin        Alias for --provider
  -p, --provider      Show provider instead of path
  -t, --tool string   Filter by specific tool
  -v, --version       Show version instead of path

Global Flags:
  -C, --cd string       change directory before running command
  -c, --config string   config file path (default: .unirtm.toml or unirtm.toml)
      --dry-run         show what would happen without making changes
  -E, --env string      set the environment for loading configuration
      --help            help for this command
      --jobs int        how many jobs to run in parallel (default 8)
  -j, --json            enable JSON output format
      --locked          require lockfile URLs to be present during installation
      --no-config       do not load any config files
      --no-env          do not load environment variables from config files
      --no-hooks        do not execute hooks from config files
  -q, --quiet           enable quiet mode (minimal output)
      --silent          suppress all task output and non-error messages
  -V, --verbose         enable verbose output (debug logging)
  -y, --yes             answer yes to all confirmation prompts
```

</details>
