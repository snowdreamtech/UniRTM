# IDE Integration

UniRTM is designed to work seamlessly with modern IDEs. Unlike legacy version managers (like `asdf` or `pyenv`) that rely on slow Bash shims which confuse IDEs and language servers, UniRTM operates using **Native Environment Resolution**.

This means your IDE can instantly and correctly detect the right toolchain, linters, and language servers without any hacks or complex shell wrappers.

## Visual Studio Code

VS Code integrates beautifully with UniRTM. Since UniRTM can automatically inject environment variables and update the `PATH`, VS Code picks up your tooling out of the box when launched from the terminal.

### 1. Launching from the Terminal (Recommended)

The most robust way to ensure VS Code inherits your exact project environment is to launch it from a directory managed by UniRTM:

```bash
cd my-project
unirtm use node@20
code .
```

All VS Code extensions (like ESLint, Prettier, or Go) will instantly use the specific tools defined in your `.unirtm.toml`.

### 2. VS Code `settings.json` Integration

If you prefer to open VS Code directly from your GUI launcher, you can explicitly point your extensions to the UniRTM executable or shim paths.

To find the exact binary path of a tool, use:

```bash
$ unirtm which node
/Users/username/.local/share/unirtm/installs/node/20.0.0/bin/node
```

You can then define these explicitly in your `.vscode/settings.json`:

```json
{
  "eslint.nodePath": "/Users/username/.local/share/unirtm/installs/node/20.0.0/bin/node",
  "go.goroot": "/Users/username/.local/share/unirtm/installs/go/1.22.0/go"
}
```

## JetBrains IDEs (IntelliJ, WebStorm, GoLand, PyCharm)

JetBrains IDEs read your environment on startup. Similar to VS Code, the most reliable integration is launching your IDE from the terminal.

### GUI Configuration

If you launch via the JetBrains Toolbox, you can manually point the IDE to your project's SDK.

1. Go to **Settings/Preferences**.
2. Navigate to your language SDK settings (e.g., **Languages & Frameworks > Node.js** or **Go > GOROOT**).
3. Point the executable path to the path resolved by `unirtm which <tool>`.

UniRTM's directory structure is highly predictable, making it easy to select the exact version you need from:
`~/.local/share/unirtm/installs/<tool>/<version>`

## Neovim / Vim

If you use Neovim with plugins like `nvim-lspconfig`, `mason.nvim`, or `null-ls`, they will automatically respect your `PATH`.

Since UniRTM prepends the correct toolchain binaries to your `PATH` the moment you enter a directory, Neovim will instantly pick up the correct language servers (LSP), linters, and formatters when you open a file.

```bash
cd my-project
nvim .
```

*(No additional configuration is required!)*

## Summary

Because UniRTM eschews complicated, slow Bash shims in favor of direct environment manipulation, it eliminates the "IDE compatibility tax" that plagues other version managers. Your IDE sees the raw, native binary, ensuring zero overhead and absolute accuracy.
