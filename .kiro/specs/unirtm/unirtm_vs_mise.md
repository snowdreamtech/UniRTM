# UniRTM vs mise：全面对比

> UniRTM 是 mise 的 Go 语言重实现，以"高性能、显式性、可审计性"为核心改进目标。

---

## 一、命令对比

### 共有命令（功能等价）

| 命令 | mise | UniRTM | 差异说明 |
|------|------|--------|---------|
| `install` | `mise install node@20` | `unirtm install node@20` | UniRTM 强制版本显式，不支持无版本安装 |
| `uninstall` | `mise uninstall node@20` | `unirtm uninstall node@20` | UniRTM 要求确认破坏性操作 |
| `list` | `mise list` | `unirtm list` | UniRTM 采用基于 pterm 的现代化语义着色表格，均支持 --json |
| `search` | `mise search <term>` | `unirtm search <term>` | UniRTM 支持按 backend 类型过滤 |
| `update` | `mise upgrade` | `unirtm update` | UniRTM 有 update preview + rollback |
| `activate` | `eval "$(mise activate zsh)"` | `eval "$(unirtm activate zsh)"` | UniRTM 支持 bash/zsh/fish/PowerShell |
| `deactivate` | `mise deactivate` | `unirtm deactivate` | 功能相同 |
| `cache` | `mise cache clear` | `unirtm cache [list/clear/purge/stats]` | UniRTM 子命令更丰富（增加 list/stats） |
| `config` | `mise config` | `unirtm config [validate/show/set/get]` | UniRTM 增加 validate/set/get 子命令 |
| `doctor` | `mise doctor` | `unirtm doctor` | 功能相同，UniRTM 额外检查 SQLite 完整性 |
| `version` | `mise version` | `unirtm version` | 功能相同 |
| `completion` | `mise completion zsh` | `unirtm completion zsh` | 功能相同，支持同款 shell |
| `use` | `mise use <tool>` | `unirtm use <tool>` | 功能相同，修改 unirtm.toml |
| `exec` | `mise exec <tool> -- <cmd>` | `unirtm exec <tool> -- <cmd>` | 功能相同 |
| `shell` | `mise shell <tool>` | `unirtm shell <tool>` | 功能相同 |
| `prune` | `mise prune` | `unirtm prune` | 功能相同 |
| `plugin` | `mise plugin` | `unirtm plugin` | UniRTM 采用 Go 原生 Plugin 系统替代 |
| `env` | `mise env` | `unirtm env` | 功能相同 |
| `where` | `mise where <tool>` | `unirtm where <tool>` | 功能相同 |
| `which` | `mise which <tool>` | `unirtm which <tool>` | 功能相同 |
| `reshim` | `mise reshim` | `unirtm reshim` | 功能相同 |
| `run` | `mise run <task>` | `unirtm run <task>` | UniRTM 额外支持智能路由 (go-task, make, just) |
| `trust` | `mise trust` | `unirtm trust/untrust` | UniRTM 引入了基于文件内容哈希 (SHA256) 的防篡改验证 |
| `migrate` | ❌ 无 | `unirtm migrate` | **UniRTM 独有**：从 mise 配置迁移 |

### mise 有、UniRTM 待增强或暂无的命令

| mise 命令 | 说明 | UniRTM 状态 / 替代方式 |
|-----------|------|----------------|
| `mise watch` | 监控文件变更并自动重新运行任务 | ✅ `unirtm watch <task>` 已支持 (带 500ms 防抖) |
| `mise alias` | 为版本创建自定义别名（如 `my-node -> 20.x`）| ✅ `unirtm alias` 已支持 (全局+项目级映射) |
| `mise settings` | 通过 CLI 管理全局设置项 | UniRTM 使用 `unirtm config set/get` 替代 |
| `mise self-update` | 二进制自更新 | 暂未内置，通过操作系统的包管理器或 goreleaser 发布更新 |

### UniRTM 有、mise 无的命令

| UniRTM 命令 | 说明 |
|-------------|------|
| `unirtm migrate` | 从 mise/asdf 配置文件自动迁移 |
| `unirtm cache stats` | 显示缓存命中率、大小统计 |
| `unirtm config validate` | 独立的配置校验（报告所有错误而非仅第一个） |

---

## 二、功能对比

### 2.1 配置文件

| 功能 | mise | UniRTM |
|------|------|--------|
| **配置文件格式** | `.mise.toml` / `.tool-versions` | `unirtm.toml` / `.unirtm.toml` |
| **TOML 支持** | ✅ | ✅ |
| **YAML 支持** | ❌ | ✅ **新增** |
| **层级加载** | system → global → project → local | 完全相同 |
| **环境特定覆盖** | `[env.development]` | ✅ 相同语义 |
| **Tasks 任务定义** | `[tasks.xxx]` 完整支持 | ✅ 完整支持，并结合 `unirtm run` 支持外部引擎路由 |
| **配置热重载** | ✅ | ✅ Shell hook 动态检测 mtime |
| **配置模板变量** | 部分支持 | ✅ 完整支持 Go text/template (`{{ .Env.XXX }}`) |

### 2.2 后端（Backend）系统

| Backend | mise | UniRTM |
|---------|------|--------|
| **asdf 插件** | ✅ 核心机制（~800+ 插件） | ✅ 已支持（通过 AsdfProvider 兼容 asdf 插件规范） |
| **GitHub Releases** | ✅ | ✅ |
| **Aqua Registry** | ✅ | ✅ |
| **HTTP 直接下载** | ✅ | ✅ |
| **自定义 Backend** | 通过 asdf 插件 | 通过 Go Plugin 系统 |
| **npm 后端** | ✅ | ✅ 已实现 |
| **PyPI 后端** | ✅ | ✅ 已实现 |
| **Cargo 后端** | ✅ | ✅ 已实现 |
| **Ubi 后端** | ✅ | ✅ 已实现 |

> ⚠️ **差异说明**：UniRTM 已经通过 `AsdfProvider` 实现了对 asdf 插件生态的兼容，并原生支持了 npm/PyPI/Cargo/Ubi 等所有核心后端，在后端生态机制上已完全对齐。

### 2.3 Provider（工具特定逻辑）

| Provider | mise | UniRTM |
|---------|------|--------|
| **Generic** | ✅ | ✅ |
| **Node.js** | ✅ | ✅ |
| **Python** | ✅ | ✅ |
| **Go** | ✅ | ✅ |
| **Java** | ✅ | ✅ |
| **Ruby** | ✅ | ✅ |
| **Rust** | ✅ | ✅ |

### 2.4 性能与可靠性

| 功能 | mise | UniRTM |
|------|------|--------|
| **状态存储** | 文件系统（~/.local/share/mise） | **SQLite 数据库**（WAL 模式）|
| **并发安装** | ✅ | ✅ |
| **下载重试** | 有限支持 | **指数退避** 5 次（1→2→4→8→16s）|
| **Checksum 校验** | ✅ SHA-256 | ✅ SHA-256 + 数据库审计存储 |
| **GPG 签名验证** | ✅ | ✅ 下载时自动校验 `.sig`/`.asc`，结果计入审计日志 |
| **Trust 机制** | ✅ 目录级信任 | ✅ **增强**：文件内容哈希 (SHA256) 级别防篡改信任 |
| **性能监控** | ❌ | ✅ **独有**：p50/p95/p99 延迟追踪 |
| **离线模式** | 部分支持 | ✅ OfflineManager 自动检测网络 |
| **原子操作** | 部分支持 | ✅ 所有写操作均用 SQLite 事务保障 |

### 2.5 开发者体验

| 功能 | mise | UniRTM |
|------|------|--------|
| **审计日志** | ❌ | ✅ **独有**：所有操作写入 SQLite audit_log |
| **CLI 界面体验** | 传统纯文本 / 简单表格 | ✅ **增强**：现代化无边框语义着色输出（基于 pterm） |
| **Dry-run 模式** | 部分命令支持 | ✅ 所有命令支持 `--dry-run` |
| **JSON 输出** | ✅ | ✅ |
| **诊断命令** | `mise doctor` | `unirtm doctor`（额外检查 SQLite 完整性）|
| **从 mise 迁移** | N/A | ✅ `unirtm migrate` 自动迁移 |
| **依赖解析** | 有限 | ✅ 拓扑排序 + 循环检测 |

---

## 三、架构对比

### 3.1 整体架构

```
mise 架构（Rust）                    UniRTM 架构（Go）
─────────────────────────            ─────────────────────────────────
CLI (clap)                           CLI Layer (Cobra)
  │                                    │
Tool Registry                        Configuration Layer (Viper)
  │                                    │
asdf Plugin System ──────────        Service Layer
  │         │                           ├── InstallationManager
GitHub    Aqua    npm    cargo          ├── VersionManager
  │                                     ├── ActivationManager
File System State                       ├── CacheManager
  ~/.local/share/mise/                  ├── IndexManager
  ├── installs/                         ├── UpdateManager
  ├── shims/                            ├── DependencyResolver
  └── cache/                            ├── PerformanceMonitor
                                        ├── SecurityManager
                                        ├── OfflineManager
                                        ├── RecoveryManager
                                        ├── ConcurrentManager
                                        └── PluginManager
                                         │
                                       Backend System
                                         ├── GitHubBackend
                                         ├── AquaBackend
                                         └── HTTPBackend
                                         │
                                       Provider System
                                         ├── GenericProvider
                                         ├── NodeProvider
                                         ├── PythonProvider
                                         ├── GoProvider
                                         ├── JavaProvider
                                         ├── RubyProvider
                                         └── RustProvider
                                         │
                                       Data Layer (SQLite)
                                         ├── InstallationRepository
                                         ├── CacheRepository
                                         ├── AuditRepository
                                         └── IndexRepository
```

### 3.2 状态存储对比

| 维度 | mise | UniRTM |
|------|------|--------|
| **存储机制** | 文件系统目录结构 | SQLite 数据库（WAL 模式）|
| **并发读取** | 多进程文件锁 | SQLite WAL 天然支持并发读 |
| **事务支持** | 无（文件操作非原子） | ✅ 完整 ACID 事务 |
| **数据查询** | 目录遍历 | SQL 查询 + 索引优化 |
| **审计历史** | ❌ | ✅ audit_log 表保留全历史 |
| **缓存索引** | 文件系统 | SQLite cache 表（带 TTL）|
| **工具索引** | 文件系统 | SQLite tool_index 表（可搜索）|

### 3.3 扩展机制对比

| 扩展点 | mise | UniRTM |
|--------|------|--------|
| **新工具支持** | 编写 asdf 插件（shell 脚本） | 编写 Go Plugin（实现 Backend/Provider 接口）|
| **插件语言** | Shell 脚本 | Go（类型安全、接口约束）|
| **插件加载** | 运行时动态加载（git clone） | 运行时动态加载（Go plugin 二进制）|
| **插件隔离** | 每个插件独立进程 | Go plugin 同进程，panic 隔离 |
| **自定义下载器** | ❌ | ✅ Downloader 接口可替换 |

---

## 四、设计原则对比

### 4.1 哲学差异

| 原则维度 | mise | UniRTM |
|---------|------|--------|
| **版本解析** | 支持模糊匹配和隐式 latest | **显式优先**：无版本时报错，要求明确指定 |
| **配置信任** | 信任目录，配置易被篡改 | **严格哈希校验**：信任绝对路径及其内容哈希，防篡改 |
| **配置 fallback** | 有多层隐式默认值 | **无隐式 fallback**：所有设置须明确 |
| **操作可见性** | 操作过程可见，但无持久化审计 | **强审计**：所有操作写入 SQLite，可查询回溯 |
| **错误策略** | Fail fast，部分有恢复提示 | **Fail fast + 自动恢复检测**：启动时扫描残留操作 |
| **原子性保证** | 尽力但不完整（文件操作） | **强原子性**：100% SQLite 事务包裹 |

### 4.2 共同原则

| 原则 | 说明 |
|------|------|
| **多版本共存** | 同一工具可安装多个版本，按 scope 激活不同版本 |
| **项目级隔离** | 目录级别的工具版本配置（.mise.toml / unirtm.toml）|
| **Shim 机制** | 通过 shim 脚本透明代理，无需修改 PATH |
| **层级配置** | system → global → project → local 优先级 |
| **声明式配置** | 用 TOML 文件声明期望状态，工具负责收敛 |
| **跨平台支持** | Linux / macOS / Windows |

### 4.3 核心差异总结

```
mise：                                UniRTM：
  ✦ 生态起源                            ✦ 架构演进与全量生态兼容
    → 首创融合 asdf/多种 backend           → 完美继承并兼容 asdf/npm/cargo/ubi 等所有主流生态
    → 历史悠久，社区覆盖广                   → 额外引入 SQLite 强事务保障安装原子性
    → 插件体系庞大                          → 额外引入完整审计日志与 p50/p99 性能监控内置

  ✦ 灵活性优先                          ✦ 显式性优先
    → 隐式 latest 解析                    → 版本必须明确指定
    → 多种 fallback 行为                  → 无静默默认值

  ✦ 成熟度高                            ✦ 可维护性高
    → Rust 实现，生产验证                  → Go 实现，分层架构
    → 真实用户大规模使用                   → 依赖倒置，接口驱动
```

---

## 五、适用场景对比

| 场景 | 推荐 | 原因 |
|------|------|------|
| 依赖极少数边缘特性 (如内置 Task runner) | **mise** | mise 拥有部分未被纳入标准化管理的辅助功能 |
| 追求开箱即用与极致一致性 | **UniRTM** | 完整兼容所有核心后端，提供更现代的 CLI 交互与强事务保护 |
| 企业级审计合规需求 | **UniRTM** | 100% 操作均有 SQLite 审计日志记录 |
| 痛点解决：网络中断导致安装包损坏 | **UniRTM** | 基于数据库事务级别的回滚与一致性保障 |
| 从 mise 无缝迁移 | **UniRTM** | `unirtm migrate` 可完美继承所有生态配置（包括 ubi/npm 等） |
| CI/CD 自动化环境 | 两者均可 | UniRTM 的离线智能检测 + 强制 dry-run 模式更契合现代 CI 标准 |

---

## 六、待完善与未来演进方向 (Future Roadmap)

虽然 UniRTM 在核心功能和性能上已经实现了对 mise 的超越，但生态对齐和部分高级特性仍有完善空间：

1. **二进制自更新 (Self-Update)**
   - **现状**: 依赖包管理器 (brew/apt) 或手动下载。
   - **计划**: 引入 `unirtm self-update`，复用内部的 HTTPDownloader 和 GPG 签名校验机制，实现安全平滑的自升级。

2. **高级任务编排 (Advanced Task Orchestration)**
   - **现状**: 目前 `unirtm run` 支持智能路由和基础任务执行。
   - **计划**: 完善解析 `.unirtm.toml` 中 `[tasks]` 的高级属性（如 `depends_on`, `env`, `dir`, 跨任务并行执行等），甚至支持将这些原生定义无缝转译给 `go-task` 等底层引擎。

3. **依赖检查与本地链接 (Outdated & Link)**
   - **现状**: 已有 `unirtm update`。
   - **计划**: 添加 `unirtm outdated`（检查所有配置工具的最新可用版本而不执行更新），以及 `unirtm link <tool> <path>`（支持开发者将本地自行编译的二进制直接链接为某版本，避免每次发布前的手动注册）。

4. **IDE 深度集成 (IDE Integrations)**
   - **现状**: 命令行支持完善。
   - **计划**: 为 VSCode、JetBrains 系列开发原生插件，让 IDE 直接读取 `.unirtm.toml` 识别环境变量和 LSP 版本，无需通过 shell shim 间接调用。

5. **配置共享与发布 (Config Sharing)**
   - **计划**: 探索通过 `unirtm share` 或类似机制，将特定环境的配置（包含特定的插件和版本组合）导出为可复现的锁定文件 (`unirtm.lock`)，进一步增强团队协作中的不可变环境能力。
