# 持续集成 (CI/CD)

UniRTM 提供了一等公民级别的 GitHub Action——`snowdreamtech/setup-unirtm`，旨在将你的本地开发环境无缝集成到 CI/CD 管道中。这保证了你的笔记本电脑和构建服务器之间的执行环境达到 **100% 的绝对一致**。

## 核心工作流

使用默认配置启动并运行该 Action 是最快的入门方式：

```yaml
name: CI
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup UniRTM & Install Tools
        uses: snowdreamtech/setup-unirtm@v1
        with:
          # 可选：指定 UniRTM 的版本（默认为最新版）
          version: latest
          # 可选：传入 GitHub Token，以防止下载依赖时触发 API 速率限制
          github_token: ${{ secrets.GITHUB_TOKEN }}

      - name: Run Tests
        # .unirtm.toml 中定义的所有工具，现在已经作为原生命令暴露在环境中了
        run: npm test
```

## 深度剖析：`setup-unirtm` 的底层代码原理

`setup-unirtm` 绝不仅仅是一个简单的下载脚本。它是一个专门为临时性 CI 环境量身定制的高性能集成外壳。

让我们深入代码，看看它在底层到底为你做了哪些工作：

### 1. 无摩擦的二进制解析 (Zero-Friction Resolution)

Action 不会从源码编译 UniRTM，也不会依赖缓慢的包管理器。它会利用你提供的 `github_token`（绕过限流）查询 GitHub Releases API，自动探测当前 Runner 的操作系统（Linux, macOS, Windows）和 CPU 架构（amd64, arm64），然后直接拉取极其小巧的编译好的原生二进制文件。最后，它会将该文件路径隐式注入到环境的 `$GITHUB_PATH` 中。

### 2. 智能自动缓存 (Intelligent Auto-Caching)

CI 流程中最耗时的步骤通常是下载和编译开发工具（如构建 Python 或 Ruby）。`setup-unirtm` 在底层拦截了这一痛点：

- 它会提取你的 `.unirtm.toml` 和 `.unirtm.lock` 文件，经过哈希计算生成一个全局唯一的 Cache Key。
- 它调用底层的 `@actions/cache` API 尝试恢复 `~/.local/share/unirtm/installs`（安装目录）和 `downloads`（下载缓存目录）。
- 如果缓存命中（Cache Hit），整个工具链的安装耗时将缩短为 **0 秒**。
- 如果缓存未命中，Action 会自动在后台执行 `unirtm install`，并在整个 Workflow 结束（Post-run）时，将其无缝打包回传至 GitHub 缓存服务器。

### 3. 环境变量桥接 (Environment Variable Bridging)

UniRTM 允许你在 `.unirtm.toml` 的 `[env]` 块中声明项目级别的环境变量。
Action 代码会在内部执行 `unirtm env` 解析这些变量，并将它们提取后原封不动地注入到 GitHub Runner 的 `$GITHUB_ENV` 上下文中。
这意味着，**你本地的所有环境变量会在 CI 的所有后续步骤中瞬间生效**，无需你在 GitHub Secrets 或 Variables 设置里去痛苦地一条条复制粘贴（除了敏感凭证外）。

### 4. 自动 PATH 注入 (Automatic Tool Shimming)

安装工具后，`setup-unirtm` 并没有止步。它会精准提取出被激活工具（如 `node`, `go`）的底层绝对路径，并将这些路径挂载进 `$GITHUB_PATH`。
因此，在后续的 `run` 步骤中，你根本不需要使用繁杂的 `unirtm exec npm -- install`；你可以直接、自然地运行 `npm install`。因为原生的 Runner PATH 已经被完美地“接管”并指向了你固定的工具链。

## 自定义输入参数

你可以通过以下参数自定义 `setup-unirtm` 的行为：

| Input | 默认值 | 描述 |
|---|---|---|
| `version` | `latest` | 安装的 UniRTM 版本 (例如 `v1.2.3`)。 |
| `install` | `true` | 是否在设置好 CLI 后自动运行 `unirtm install`。 |
| `cache` | `true` | 是否启用 GitHub Actions 缓存来持久化保存工具。 |
| `github_token` | <code v-pre>${{ github.token }}</code> | 用于获取发布版本并避免流控的 Token。 |
| `unirtm_toml` | `.unirtm.toml` | 配置文件的相对路径（常用于 Monorepo）。 |

## 其他 CI 平台 (GitLab CI, CircleCI)

如果你没有使用 GitHub Actions，在其他平台集成 UniRTM 同样极其简单，只需要依赖官方的基础 Bash 脚本即可：

```yaml
# GitLab CI 示例
test:
  image: ubuntu:latest
  script:
    # 1. 安装 UniRTM
    - curl -fsSL https://github.com/snowdreamtech/unirtm/raw/main/install.sh | sh
    - export PATH="$HOME/.local/bin:$PATH"

    # 2. 安装项目依赖的工具
    - unirtm install

    # 3. 提取环境变量和 PATH 并注入当前 Session
    - eval "$(unirtm activate bash)"

    # 4. 执行测试
    - npm test
```
