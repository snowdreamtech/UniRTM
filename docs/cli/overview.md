# CLI Overview

UniRTM provides a rich, exhaustive CLI interface. Below is a complete catalog of all available commands. You can always run `unirtm help [command]` in your terminal to see specific flag usage.

## Tool Management Commands

<details open>
<summary><code>unirtm install &lt;tool@version&gt;</code></summary>

Downloads and installs a specific version of a tool.

- **Example**: `unirtm install node@20.14.0`
- **Behavior**: If no version is specified, UniRTM reads the `.unirtm.toml` in the current directory and installs all missing tools defined there.

</details>

<details open>
<summary><code>unirtm uninstall &lt;tool@version&gt;</code></summary>

Removes an installed tool version from the local cache.

- **Example**: `unirtm uninstall go@1.22.0`

</details>

<details open>
<summary><code>unirtm use &lt;tool@version&gt;</code></summary>

Pins a specific tool version in the current directory's `.unirtm.toml` file.

- **Example**: `unirtm use python@3.12`
- **Behavior**: Creates or updates `.unirtm.toml` and automatically triggers an installation if the tool is not locally cached.

</details>

<details open>
<summary><code>unirtm current</code></summary>

Displays the currently active tool versions for the current directory context.

- **Example**: `unirtm current`
- **Output**: Lists the tools, the resolved active versions, and the file path of the configuration that enforces that version.

</details>

<details open>
<summary><code>unirtm ls</code></summary>

Lists all installed tools and their versions on your local machine.

- **Example**: `unirtm ls`

</details>

<details open>
<summary><code>unirtm ls-remote &lt;tool&gt;</code></summary>

Queries the remote registry to list all installable versions of a specific tool.

- **Example**: `unirtm ls-remote java`

</details>

<details open>
<summary><code>unirtm outdated</code></summary>

Checks your `.unirtm.toml` and reports if newer versions of your pinned tools are available in the remote registry.
</details>

<details open>
<summary><code>unirtm upgrade &lt;tool&gt;</code></summary>

Upgrades the specified tool in your `.unirtm.toml` to the latest compatible version based on your SemVer constraints.
</details>

<details open>
<summary><code>unirtm bin-paths</code></summary>

Outputs the absolute paths to the `bin/` directories of all currently active tools. This is primarily used internally for `PATH` injection.
</details>

---

## Execution Commands

<details open>
<summary><code>unirtm run &lt;task&gt;</code></summary>

Executes a named task defined in your `.unirtm.toml` file.

- **Example**: `unirtm run build`

</details>

<details open>
<summary><code>unirtm exec &lt;tool&gt; -- &lt;command&gt;</code></summary>

Executes a specific command under the context of a tool without modifying the global `PATH`.

- **Example**: `unirtm exec node@18 -- npm run build`
- **Behavior**: This is extremely useful for running one-off scripts in a specific runtime version without affecting the current shell.

</details>

---

## Environment Commands

<details open>
<summary><code>unirtm env</code></summary>

Outputs the evaluated environment variables and `PATH` modifications for the current directory context.

- **Example**: `eval "$(unirtm env)"` (This is how shell integrations apply the environment).

</details>

---

## Core & Diagnostic Commands

<details open>
<summary><code>unirtm doctor</code></summary>

Runs a comprehensive suite of diagnostic checks on your environment, validating file permissions, network connectivity to remote registries, and ensuring `.unirtm.toml` syntax is valid.
</details>

<details open>
<summary><code>unirtm completion &lt;shell&gt;</code></summary>

Generates shell completion scripts for `bash`, `zsh`, `fish`, or `powershell`.
</details>

<details open>
<summary><code>unirtm config</code></summary>

Displays the global UniRTM configuration (e.g., cache directories, parallel download limits).
</details>

<details open>
<summary><code>unirtm plugin</code></summary>

Manages custom backend plugins. Supports subcommands like `unirtm plugin add`, `unirtm plugin list`, and `unirtm plugin remove`.
</details>
