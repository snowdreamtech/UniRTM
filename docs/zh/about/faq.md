# 常见问题 (FAQ)

## UniRTM 与 mise / asdf 有什么区别？

**asdf** 是这个领域的先驱，但它使用纯 Bash 编写，在 Windows 上支持极差，并且因为使用了 `shims`（垫片）机制，每次执行命令都会带来显著的性能损耗。

**mise** 是对 asdf 的极速重写版，移除了 `shims` 并带来了巨大的性能提升。

**UniRTM** 在架构上向 **mise** 致敬，同样放弃了 `shims` 机制，转而通过拦截终端的 prompt 钩子直接动态修改环境变量（如 `$PATH`），性能表现极为优异。但 UniRTM 的核心超越在于它**不仅仅是一个传统的运行时管理器**。在 AI 辅助编程时代，UniRTM 将 AI 智能体（Cursor, Windsurf 等）的上下文规则体系视为“工具链”中不可或缺的一环，原生内建了对 [SpecKit 工作流](../guide/ide-integration.md) 的支持。

## 为什么一个运行时管理器要管理 AI 工作流（SpecKit）？

在过去，开发环境的一致性仅仅意味着“同样的 Node.js 或 Python 版本”。但在当今的 AI 时代，开发团队面临的新挑战是：“如何保证团队中每个人的 AI 助手（无论是 Cursor, Claude 还是 GitHub Copilot）都在遵循同样的项目规范和提示词上下文？”

UniRTM 认为，**AI Agent 的规则本身就是现代开发工具链的一部分**。因此，除了管理传统的 SDK，我们创造性地引入了对 `.agent/rules` 的管理，实现了真正的“单点事实（Single Source of Truth）”。

## UniRTM 支持 Windows 吗？

**完全支持。**
与 asdf 截然不同，UniRTM 使用 Go 语言编写，将 Windows 视为绝对的一等公民。我们针对 Windows 的路径解析、环境变量处理以及终端（PowerShell / CMD）环境进行了深度的重构和适配。你可以毫无障碍地在 Windows 下管理跨平台工具。

## 我可以使用已有的 `.tool-versions` (asdf) 或 `mise.toml` 吗？

**可以。**
为了让开发者能够无痛迁移，UniRTM 原生支持解析 asdf 的 `.tool-versions` 文件以及 mise 的 `mise.toml` 配置文件。如果你在一个已经使用了 asdf/mise 的老项目中工作，只需安装 UniRTM 即可直接接管项目环境。

## UniRTM 是如何动态修改环境变量的？

UniRTM 不会使用慢速的 `shims`。当你在 `~/.bashrc` 或 `~/.zshrc` 中执行 `eval "$(unirtm activate bash/zsh)"` 时，UniRTM 会向你的 Shell 中注入一个极轻量级的钩子函数（例如 zsh 的 `precmd` 或 bash 的 `PROMPT_COMMAND`）。

每次你切换目录或按下回车时，这个钩子会以纳秒级的速度检查当前目录结构，如果发现有 `.unirtm.toml`，则会在当前 Shell 的上下文中动态重写 `$PATH` 和其他环境变量。这意味着你调用的 `node` 或 `python` 就是磁盘上真实的二进制文件，没有任何中间层的性能损耗。

## 如何配置全局默认版本？

你可以直接修改家目录下的全局配置文件 `~/.config/unirtm/config.toml`，或者使用命令快速设置：

```bash
unirtm use --global node@20
```

## 我该如何卸载 UniRTM？

UniRTM 完全是绿色软件，不会在系统中留下任何顽固的注册表或深层依赖。卸载只需两步：

1. 打开你的 shell 配置文件（如 `~/.zshrc` 或 `~/.bashrc`），删除 `eval "$(unirtm activate ...)"` 和 PATH 相关的行。
2. 删除数据目录和二进制文件：

   ```bash
   rm -rf ~/.local/share/unirtm
   rm -rf ~/.config/unirtm
   rm -f ~/.local/bin/unirtm
   ```
