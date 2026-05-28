# 开发工具管理概览

UniRTM 自动处理系统中开发工具的安装和版本切换。无论你使用的是 `node`、`python`、`go` 还是发布到 GitHub 上的原生二进制文件，UniRTM 都能无缝管理它们，且完全无需依赖拖慢系统的 Shell 脚本。

## 工作原理

当运行 `unirtm use node@20` 时，会发生以下两件事：

1. **解析与下载**：UniRTM 从其配置的后端解析版本号，将二进制文件下载到全局缓存目录（通常位于 `~/.local/share/unirtm/installs`），并赋予执行权限。
2. **注册**：它将此依赖项写入你本地的 `.unirtm.toml` 文件中。

当你在终端中导航到此目录时，UniRTM 会检测到 `.unirtm.toml` 文件，并将其内部的 `bin` 目录动态注入到系统的 `$PATH` 前端。这保证了当你键入 `node` 时，执行的绝对是该项目指定的版本。

## 后端体系 (Backends)

UniRTM 与传统工具管理器的一个核心区别在于它获取软件的方式。UniRTM 采用了可插拔的后端架构：

- **核心插件 (Core Plugins)**：UniRTM 附带了直接编译进其 Go 二进制文件中的核心插件。Node.js、Go 和 Python 等工具拥有原生支持，提供极致的下载速度和校验和（checksum）验证。
- **GitHub 后端**：你可以直接从 GitHub Releases 安装几乎任何工具，而无需编写自定义插件。
  ```bash
  unirtm use github:cli/cli@latest
  ```
- **Cargo/NPM/Pip 后端**：UniRTM 可以通过其他包管理器代理安装过程，并在不同项目之间独立管理所生成的二进制文件版本。

## 工具解析顺序

UniRTM 通过（按顺序）检查以下位置来解析工具版本：

1. `UNIRTM_NODE_VERSION` （系统环境变量）
2. `.unirtm.toml` （当前目录）
3. `.nvmrc` 或 `.node-version` （传统版本文件，如果开启了兼容模式）
4. `.unirtm.toml` （父级目录，向上递归到根目录）
5. `~/.config/unirtm/config.toml` （全局配置）

如果一个工具被声明了但是尚未在本地安装，UniRTM 会优雅地降级回退，或者提示你运行 `unirtm install` 进行安装。
