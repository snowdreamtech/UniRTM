# Shims & PATH Manipulation

In multi-version environment managers (like asdf, pyenv, etc.), intercepting system commands and routing them to a specific version usually involves two mainstream approaches: **Shims** and **Dynamic PATH Manipulation**.

UniRTM supports both modes, but we highly recommend and default to **Dynamic PATH Manipulation** for maximum performance.

## 1. Dynamic PATH Manipulation (Recommended)

This is the core magic of UniRTM. By hooking into your `.zshrc` or `.bashrc`:

```bash
eval "$(unirtm activate zsh)"
```

**How it works:**

- When you use `cd` to enter a project containing a `.unirtm.toml`, UniRTM's shell hook is instantly triggered.
- It reads the configuration, resolves the required tool versions, and safely prepends the actual binary paths (e.g., `~/.local/share/unirtm/installs/node/20.10.0/bin`) to your `$PATH` environment variable.
- When you leave the directory, it seamlessly removes those paths from `$PATH`.

**Why is it better?**

- **Zero Overhead:** When you execute `node -v`, the OS directly executes the real Node.js binary. There are no intermediary scripts, resulting in absolute 0 performance loss.
- **High Transparency:** Running `which node` directly returns the real installation path, making debugging incredibly straightforward.

## 2. Shims Mode (IDE & Fallback)

While PATH manipulation is perfect for interactive terminals, some environments (like certain GUI IDEs, legacy Makefiles, or system-wide shortcut runners) do not execute shell hooks and therefore remain blind to the dynamically modified `$PATH`.

This is where **Shims** come into play.

If you need Shims, you can manually add `~/.local/share/unirtm/shims` to your system's global `$PATH`, or run:

```bash
unirtm reshim
```

**How it works:**

- Unlike asdf which uses extremely slow Bash scripts for shims, UniRTM uses ultra-lightweight **native symlinks**.
- Commands (like `node`, `python`) are symlinked back to the core UniRTM Go executable.
- When your IDE invokes `node`, it's actually invoking `unirtm`. UniRTM computes the required version for the current working directory in milliseconds and forwards the arguments to the real binary transparently.

**Performance comparison:**
Even when using Shims, because UniRTM is a statically compiled Go binary, its routing overhead is around 1-2 milliseconds, compared to traditional Bash shims which often take 50-100 milliseconds.

## Summary & Best Practices

- **Daily Terminal Usage:** Make sure to configure `unirtm activate` to enjoy zero-latency PATH manipulation.
- **IDEs (VSCode / JetBrains / Cursor):** If your IDE struggles to find the correct tool version, ensure the IDE is configured to inherit the terminal environment, or add UniRTM's shim directory to the IDE's global path settings.
