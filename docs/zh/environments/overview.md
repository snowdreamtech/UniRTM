# 环境变量管理概览 (Environment Variables Overview)

现代应用程序极少在真空环境中运行。它们需要令牌、数据库连接字符串、端口绑定以及 API 密钥。UniRTM 将环境变量管理视为与工具版本控制同等重要的核心一等公民。

## 传统 `.env` 方案的痛点

传统上，开发者依赖 `dotenv`、`direnv` 等实用工具，或在 `.bashrc` 中硬编码 `export` 来管理环境变量。这些方法存在显著缺陷：

1. **强依赖 Shell**: `direnv` 必须安装 Shell 钩子才能工作。
2. **强依赖语言**: 使用 `dotenv` 意味着你必须在代码仓库中为你使用的每种语言（Node, Python, Go 等）单独引入对应的解析库。
3. **作用域污染**: 全局 `export` 会污染你的系统环境，并可能导致密钥泄露给无关的后台应用。

## UniRTM 的解决方案：上下文隔离的环境

UniRTM 会在执行特定工具或任务的瞬间，动态地将环境变量注入到进程中。这意味着环境变量的作用域被严格限制在即将派生的目标进程内，绝不外泄。

### 多层级解析策略 (Multi-Tier Resolution)

UniRTM 采用严格的优先级层级来计算环境变量（从高到低）：

1. **CLI 临时覆盖**: `PORT=8080 unirtm run start`
2. **任务专属变量**: 定义在 `.unirtm.toml` 特定的 `[tasks.start.env]` 块内部的变量。
3. **项目级变量 (`.unirtm.toml`)**: 定义在项目 `.unirtm.toml` 文件全局 `[env]` 块中的变量。
4. **本地 Env 文件 (`.env`)**: 项目目录中标准的 `.env` 文件。
5. **全局 UniRTM 变量**: 定义在 `~/.config/unirtm/config.toml` 中的变量。
6. **系统环境变量**: 从宿主操作系统继承的默认变量。

### 示例：`.unirtm.toml` 的 `[env]` 块

你可以直接在配置文件中原生定义变量：

```toml
[env]
NODE_ENV = "development"
API_ENDPOINT = "https://api.staging.example.com"
# 你甚至可以使用变量插值语法
DEBUG_URL = "${API_ENDPOINT}/debug"
```

### 支持的语法格式

UniRTM 原生集成了解析引擎，完美支持：

- 标准的 `key=value` 格式。
- 引号处理：`"value"` 与 `'value'`。
- Export 语法：`export KEY=value` （前缀的 `export` 关键字会被安全地忽略）。
- Bash 风格的参数展开：`${VAR:-default}`。

### 与密钥管理系统集成 (Secrets Management)

将生产环境的敏感密钥或 Token 提交到 `.unirtm.toml` 或 `.env` 中是极高风险的行为。
请参阅 [密钥管理机制](./secrets.md) 指南，了解 UniRTM 如何无缝对接 1Password CLI、AWS Secrets Manager 或 HashiCorp Vault，实现运行时密钥的动态安全拉取。
