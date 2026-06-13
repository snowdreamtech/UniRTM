# Shims 与 PATH 劫持

在多版本环境管理工具（如 asdf、pyenv 等）中，拦截系统命令并将其路由到特定版本通常有两种主流方式：**Shims（垫片）**和**动态 PATH 劫持**。

UniRTM 同时支持这两种模式，但我们强烈推荐且默认使用**动态 PATH 劫持**，以获得极致的性能。

## 1. 动态 PATH 劫持 (首选推荐)

这是 UniRTM 的核心魔法。通过在你的 `.zshrc` 或 `.bashrc` 中挂载：

```bash
eval "$(unirtm activate zsh)"
```

**工作原理：**

- 当你使用 `cd` 命令进入一个包含 `.unirtm.toml` 的项目时，UniRTM 的 Shell 钩子会立即被触发。
- 它会读取配置，解析出当前目录需要的工具版本，并直接将这些特定版本所在的真实二进制路径（如 `~/.local/share/unirtm/installs/node/20.10.0/bin`）前置插入到你的 `$PATH` 环境变量中。
- 当你离开该目录时，它又会安全地从 `$PATH` 中移除这些路径。

**为什么它更好？**

- **零额外开销（Zero Overhead）：** 当你执行 `node -v` 时，操作系统直接执行了真实的 Node.js 二进制文件，中间没有任何脚本转发，性能损耗为 0。
- **透明度高：** 运行 `which node` 会直接返回真实安装路径，这对于排查问题非常直观。

## 2. Shims 垫片模式 (IDE 与后备模式)

虽然 PATH 劫持在终端里表现完美，但在一些不支持自动执行 Shell 钩子的环境（比如某些 GUI IDE、古老的 Makefile 构建脚本或系统全局快捷键）中，它们可能无法感知到动态修改的 `$PATH`。

这时候就需要 **Shims** 出场了。

如果你需要使用 Shims，只需将 `~/.local/share/unirtm/shims` 添加到你的系统环境变量中，或者运行：

```bash
unirtm reshim
```

**工作原理：**

- 不同于 asdf 使用缓慢的 Bash 脚本作为 Shim，UniRTM 使用的是极其轻量的**原生符号链接 (Symlinks)**。
- 所有的命令（如 `node`、`python`）都会软链接回 UniRTM 的 Go 原生可执行文件。
- 当你在 IDE 中调用 `node` 时，实际上是调用了 `unirtm`，它会瞬间判断当前工作目录所需的 Node.js 版本，然后将参数透明地转发给真实的二进制文件。

**性能对比：**
即使是使用 Shims 模式，由于 UniRTM 是纯 Go 编译的静态二进制文件，它的转发耗时在 1-2 毫秒级别，而传统的 Bash Shims 通常需要 50-100 毫秒。

## 总结与最佳实践

- **日常终端开发**：请务必配置 `unirtm activate`，享受零延迟的 PATH 劫持。
- **IDE (如 VSCode / WebStorm / Cursor)**：如果在 IDE 终端中遇到找不到版本的问题，请确保 IDE 配置了继承终端环境，或者将 UniRTM 的 shim 路径添加到 IDE 的全局配置中。
