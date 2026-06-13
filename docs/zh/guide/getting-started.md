# 快速开始 (Getting Started)

欢迎使用 UniRTM！UniRTM 是一个完全使用 Go 语言编写的通用运行时管理器。它不仅仅是一个多语言工具管理器，还是一个环境变量加载器和任务运行器。

如果你曾经使用过 `asdf`、`mise`、`nvm`、`pyenv` 或 `rbenv`，你会感到非常熟悉——但 UniRTM 提供了压倒性的性能优势和开箱即用的安全审计功能。

## 为什么选择 UniRTM？

1. **100% 原生架构 (极致极速)**: 纯 Go 编写，零外部 Bash 插件依赖。彻底告别 Bash/Ruby 类工具带来的环境污染和 Shell 启动卡顿。
2. **绝对零污染**: 保持你的 `.bashrc` 和 `.zshrc` 干净如初。UniRTM 能够动态注入环境，而无需在每次切换目录时执行任意的 Shell 脚本。
3. **无缝安全集成**: 深度集成业界标准的 Trivy、Syft 和 Gitleaks，轻松扫描工具链漏洞和环境变量中的敏感凭证，为软件供应链护航。
4. **原生跨平台**: 预编译的原生二进制文件，将 Windows、macOS 和 Linux 一视同仁，提供完全一致的体验。
5. **All-in-One**: 一个工具取代你的版本管理器、环境变量加载器（如 `direnv`）和任务运行器（如 `make`），同时开箱即用地为 AI Agent 提供内置 MCP 服务器。

## 1. 安装 UniRTM

最简单的安装方式是使用一键安装脚本：

```bash
curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
```

*如果你想通过 Homebrew、Cargo 或 APT 安装，请参阅 [安装 UniRTM](./installing-unirtm.md)。*

## 2. 注入你的 Shell

为了让 UniRTM 在你执行 `cd` 切换目录时自动为你切换工具版本和环境变量，你需要将它注入到你的 Shell 配置中。

```bash
# 对于 Zsh
echo 'eval "$(unirtm env --shell zsh)"' >> ~/.zshrc

# 对于 Bash
echo 'eval "$(unirtm env --shell bash)"' >> ~/.bashrc
```

重启终端（或执行 `source ~/.zshrc`）使配置生效。

## 3. 安装你的第一个工具

进入你的项目目录，并声明你需要使用的开发工具：

```bash
$ unirtm use node@20 python@3.11 go@1.22
✓ wrote .unirtm.toml

$ unirtm install
✓ installed 3 tools
```

验证它们是否已激活：

```bash
$ node -v
v20.x.x
$ go version
go version go1.22.x
```

## 4. 配置环境变量

在项目根目录创建一个 `.env` 文件：

```bash
echo "DATABASE_URL=postgres://localhost:5432/mydb" > .env
```

当你进入该目录时，UniRTM 会自动为你加载这个变量；当你离开目录时，它会被自动卸载，保持全局环境的绝对干净。

## 5. 定义任务

打开你项目中的 `.unirtm.toml`，添加一个任务：

```toml
[tasks.test]
description = "运行单元测试"
run = "go test ./..."
```

执行它：

```bash
unirtm run test
```

## 下一步

- 阅读 [核心特性漫游](./walkthrough.md) 了解更多高级玩法。
- 深入了解 [.unirtm.toml 配置文件](../configuration/unirtm-toml.md)。
- 探索 [任务运行器](../tasks/overview.md)。
