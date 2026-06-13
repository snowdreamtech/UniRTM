# IDE 深度集成

UniRTM 的设计初衷之一就是与现代 IDE 无缝协同。传统的版本管理器（如 `asdf` 或 `pyenv`）重度依赖缓慢且复杂的 Bash 垫片（Shims），这往往会让 IDE 和语言服务器（LSP）感到困惑。相比之下，UniRTM 采用了 **原生环境变量解析 (Native Environment Resolution)** 架构。

这意味着你的 IDE 能够瞬间、精准地检测到正确的工具链、代码检查器（Linters）和语言服务器，而无需任何恶心的 Hack 或复杂的 Shell 包装脚本。

## Visual Studio Code

VS Code 与 UniRTM 的结合天衣无缝。由于 UniRTM 能够自动注入环境变量并修改 `PATH`，只要你从终端启动 VS Code，它就能开箱即用地继承所有配置。

### 1. 从终端启动（强烈推荐）

确保 VS Code 完全继承项目环境（包含版本固定和 `.unirtm.toml` 中的环境变量）最稳妥的方法是，在进入项目目录后直接启动：

```bash
cd my-project
unirtm use node@20
code .
```

只要这样做，VS Code 的所有插件（如 ESLint, Prettier, Go 插件等）都会立刻使用 `.unirtm.toml` 中精确定义的工具二进制文件。

### 2. 通过 `settings.json` 直接配置

如果你习惯从系统的 GUI 启动器或快捷方式打开 VS Code，你可以显式地将插件路径指向 UniRTM 的安装目录。

你可以使用 `unirtm which` 获取任意工具的精确绝对路径：

```bash
$ unirtm which node
/Users/username/.local/share/unirtm/installs/node/20.0.0/bin/node
```

然后，在项目的 `.vscode/settings.json` 中明确指定它们：

```json
{
  "eslint.nodePath": "/Users/username/.local/share/unirtm/installs/node/20.0.0/bin/node",
  "go.goroot": "/Users/username/.local/share/unirtm/installs/go/1.22.0/go"
}
```

## JetBrains IDEs (IntelliJ, WebStorm, GoLand, PyCharm)

JetBrains 家族的 IDE 也是在启动时读取环境。和 VS Code 类似，从终端直接启动 IDE 是最省心的方式。

### GUI 手动配置

如果你通过 JetBrains Toolbox 启动 IDE，你可以手动为项目指定 SDK。

1. 打开 **Settings/Preferences (设置)**。
2. 导航到对应语言的 SDK 设置项（例如：**Languages & Frameworks > Node.js** 或 **Go > GOROOT**）。
3. 将执行路径指向由 `unirtm which <tool>` 解析出的路径。

UniRTM 的目录结构具有极高的可预测性，你可以轻松在以下目录中找到对应版本：
`~/.local/share/unirtm/installs/<tool>/<version>`

## Neovim / Vim

如果你使用 Neovim 以及 `nvim-lspconfig`, `mason.nvim` 或 `null-ls` 等生态插件，它们会自动读取系统当前的 `PATH`。

由于 UniRTM 会在进入目录的瞬间将正确的工具链路径前置插入到 `PATH` 中，因此 Neovim 会在打开文件时立刻命中正确的 LSP、Linters 和格式化工具。

```bash
cd my-project
nvim .
```

*(完全不需要任何额外的插件或配置！)*

## 总结

正因为 UniRTM 抛弃了缓慢、臃肿的 Bash 垫片机制，转而采用直接注入原生环境变量的策略，它彻底消除了困扰其他版本管理器的“IDE 兼容性税”。你的 IDE 看到的永远是最原生、最纯粹的二进制文件，保证了 **零性能开销** 和 **绝对的准确性**。
