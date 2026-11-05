# 环境隔离配置 (Environments)

在现代软件工程中，开发人员通常需要根据运行环境的不同（例如本地 `development`、持续集成 `ci`、预发布 `staging` 或生产 `production`）使用不同的工具链与环境变量。

UniRTM 采用清晰的**多文件级联覆盖规范**（Multi-File Cascading Overlay，`.unirtm.<env>.toml`），使不同环境的配置彻底解耦、模块化，并彻底杜绝单文件膨胀与锁文件冲突。

## 架构与加载优先级

UniRTM 按以下层级顺序加载并合并配置（数字越大优先级越高，后加载的内容覆盖先加载的内容）：

1. **全局用户配置**：`~/.config/unirtm/unirtm.toml`
2. **项目基础配置**：`.unirtm.toml`（开发者日常开发工具与基准任务）
3. **环境覆盖配置**：`.unirtm.<env>.toml`（例如 `.unirtm.ci.toml`、`.unirtm.production.toml`）
4. **本地私有覆盖**：`.unirtm.local.toml`（个人机器专属配置，加入 `.gitignore` 且不入库）

## 定义环境覆盖配置

无需在单个文件中堆叠冗长的 `[environments]` 表格，只需在项目根目录下创建对应环境的 `.unirtm.<env>.toml` 文件。

### 示例：项目基础配置 (`.unirtm.toml`)

声明本地开发所需的常用工具和基础环境变量：

```toml
[tools]
node = "20.18.0"
go = "1.23.0"

[env]
APP_ENV = "development"
PORT = 3000

[settings]
cache_ttl = 3600
```

### 示例：CI 专属环境覆盖 (`.unirtm.ci.toml`)

声明仅在 CI 流水线中执行的代码合规检测、安全扫描与发布打包工具，本地开发者机器无需下载这些沉重的二进制：

```toml
[tools]
"github:aquasecurity/trivy" = "0.76.1"
"github:google/osv-scanner" = "2.7.0"
"github:goreleaser/goreleaser" = "2.19.0"

[env]
APP_ENV = "ci"
CI = "true"
```

## 激活指定环境

可以通过以下任意一种方式激活指定环境：

1. **命令行参数 (`-E` 或 `--env`)**：
   ```bash
   unirtm -E ci run verify
   unirtm --env production run start
   ```

2. **环境变量 (`UNIRTM_ENV`)**：
   ```bash
   export UNIRTM_ENV=ci
   unirtm run verify
   ```

## 全局锁文件生成 (`unirtm lock`)

UniRTM 多文件环境规范的一大核心优势在于统一的依赖锁定机制：

执行 `unirtm lock` 时，UniRTM 会自动扫描项目根目录下的 `.unirtm.toml` 以及**所有** `.unirtm.<env>.toml` 覆盖文件，合并全部工具声明并生成全局统一的 `unirtm.lock` 锁文件。彻底避免不同环境下的 `unirtm lock` 互相覆盖或清空非当前环境依赖工具的问题。
