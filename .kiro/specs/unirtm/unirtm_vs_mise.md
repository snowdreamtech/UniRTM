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

6. **深度环境变量管理与注入 (Advanced Env Management)**
   - **计划**: 引入 `unirtm env set/unset`，允许在 `.unirtm.toml` 中精细化配置跨平台的环境变量（例如动态解析路径、读取 `.env` 文件），并在用户进入目录时，以安全的隔离方式自动注入这些变量，成为统一的项目环境管理器（替代传统的 `direnv`）。

7. **自动化垃圾回收与磁盘优化 (Smart Garbage Collection)**
   - **计划**: 引入 `unirtm gc`。基于 SQLite 的审计日志记录，UniRTM 可以分析出长时间未被激活过的旧版本工具（基于 LRU 策略），并智能推荐或自动执行清理，释放磁盘空间。

8. **构建证明与企业级供应链安全 (SLSA & SBOM)**
   - **计划**: 在 GPG 签名校验的基础上，安装工具时自动拉取并校验工具的 **SLSA Provenance (构建来源证明)**，并能够一键导出当前项目所有工具栈的 **SBOM (软件物料清单)**，满足企业级零信任架构的合规需求。

9. **WASM 与容器化降级执行 (WASM / Docker Fallbacks)**
   - **计划**: 若某个工具在当前平台（如 Windows ARM）缺失预编译二进制，UniRTM 可自动降级去拉取其 WebAssembly (WASM) 版本（通过内置的 Wasm 运行时执行），或者静默拉取 Docker 镜像作为 shim 运行，真正实现“Write Once, Run Anywhere”。

10. **生命周期钩子机制 (Lifecycle Hooks)**
    - **计划**: 在 `.unirtm.toml` 中支持 `postinstall`、`preactivate` 等生命周期钩子。例如：当 Node.js 安装完成后自动执行 `corepack enable`，或在切换 Python 版本时自动运行 `poetry install`。

11. **可视化管理面板 (Local Web UI / TUI)**
    - **计划**: 提供 `unirtm ui`，启动一个轻量级的本地 Web Dashboard（或 TUI 终端界面），提供直观的监控：管理各个项目安装的版本、查看 SQLite 审计图表、点击升级版本、并可视化查看任务依赖拓扑图。

12. **离线缓存池与内网镜像源 (Local Mirror / Air-gapped Support)**
    - **计划**: 针对严格物理隔离（Air-gapped）的企业内网环境，提供 `unirtm mirror`。可将项目所需的全部依赖及工具包一键打包为 `离线缓存池 (Cache Pool)`，通过内网分发，实现无网环境下的瞬间装载。同时支持原生配置企业级自定义下载镜像源。

13. **零开销 Native Shim 与 eBPF 注入 (Zero-Overhead Shim)**
    - **计划**: 当前拦截依赖于 Shell 脚本（会带来毫秒级延迟）。未来将探索使用纯原生 Go 编译二进制 Shim，甚至在 Linux 下结合 eBPF 技术，在内核态无缝拦截并重定向工具执行路径，实现真正的**零延迟（Zero-latency）**环境切换。

14. **智能故障诊断与 AI 自动修复 (AI-Powered Doctor & Healing)**
    - **计划**: 增强 `unirtm doctor`，引入更强大的本地启发式规则或可选的 AI 分析。当工具因缺失系统底层依赖（如 `libssl-dev` 或特定 `glibc` 版本）安装失败时，能自动定位根本原因，并给出针对当前 OS 的确切修复命令（`apt/brew/yum`），甚至提示一键修复。

15. **企业级版本治理与安全管控 (Version Governance & Policy)**
    - **计划**: 为企业团队提供安全策略文件（如 `.unirtm.policy.toml`）。允许管理员配置黑白名单，拦截包含已知高危 CVE 漏洞的版本安装，或强制锁定在特定的 LTS 版本域内，防止私自升级导致生产故障。

16. **环境一键打包与快照分发 (Environment Bundling)**
    - **计划**: 提供 `unirtm bundle` 命令。不仅仅锁定配置文件，还能将已安装好的二进制本体、缓存、以及环境上下文打包为跨机器可移植的快照（Tarball）。该快照可直接载入 Docker 基础镜像或 VDI 云桌面中解压即用，大幅降低 CI 部署耗时。

17. **插件沙箱执行机制 (Plugin Sandbox)**
    - **计划**: 考虑到第三方工具插件存在供应链投毒风险，未来计划将不受信任的下载脚本或插件逻辑放置于严密的隔离沙箱（如 WASM Runtime 或 gVisor）中执行，确保核心文件系统的绝对安全。

18. **分布式编译与缓存共享网络 (Distributed Cache Network)**
    - **计划**: 针对需要源码编译安装的语言（如 Python、Ruby），引入远程构建缓存（Remote Caching）。当团队中某位开发者或 CI 机器完成编译后，编译产物及哈希将被上传至企业私有缓存服务器。其他成员只需秒级拉取复用，免去重复漫长的本地编译过程。

19. **透明的网络代理与根证书注入 (Transparent Proxy & CA Injection)**
    - **计划**: 针对企业内网复杂的代理和自签发证书（如 Zscaler 拦截导致的 SSL 报错），UniRTM 在激活环境时，不仅接管 PATH，还能智能识别并自动为 npm、pip、cargo 等工具链注入全局代理变量（`HTTP_PROXY`）和企业 Root CA 证书路径，彻底根除环境相关的网络故障。

20. **工具链 CVE 漏洞扫描与健康度审计 (Vulnerability Scanning)**
    - **计划**: 引入 `unirtm audit`。结合 OSV (Open Source Vulnerabilities) 等漏洞数据库，定期扫描 `.unirtm.toml` 及本地已安装的二进制文件，若发现如 Node.js 或 Python 版本存在严重安全漏洞，则主动发出警告并推荐升级到安全的 Patch 版本。

21. **原生 Monorepo 多体拓扑编排 (Polyglot Workspace Orchestration)**
    - **计划**: 深度优化对巨型 Monorepo 的支持。通过 `unirtm workspace` 分析多包代码库的跨语言环境依赖树，不仅支持按子目录激活，还允许以最优并发度在根目录一键初始化所有微服务底层依赖（Go + Node + Python 混合架构）。

22. **环境配置漂移检测 (Configuration Drift Detection)**
    - **计划**: 引入 `unirtm drift` 命令。长期开发中，本地状态可能与 `.unirtm.toml` 声明的期望状态发生偏离（如手动替换过底层文件、Shim 丢失等）。Drift 检测可以通过对比文件哈希与 SQLite 数据库记录，精准定位并修复环境不一致性。

23. **自适应底层资源分配调度 (Adaptive Resource Scheduling)**
    - **计划**: 在执行解压、并发下载或本地编译任务时，UniRTM 能够动态感知当前系统负载。当检测到开发者正在高负载使用 IDE 甚至开会时，自动将 CPU 密集型任务分配给低功耗核心（如 Apple Silicon 的 E-core）或调低 `nice` 优先级，实现“无感静默安装”。
