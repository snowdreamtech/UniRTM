# 需求规范：UniRTM 工程质量守卫与质检 Harness 体系重构 (Spec 053)

## 1. 业务背景与痛点分析
在 UniRTM 的 L5「质量守卫」架构中，工程守护者的职责是：**“装得动、装得对、写得一致、提交不炸”**。
此前项目借助 `scripts/pre-commit/*.py` 约 31 个 Python 脚本和外部 `pre-commit` 框架进行提交检查，存在如下重大痛点：
1. **执行笨重与环境强依赖**：每次 Git 提交都要启动数十次 Python 解释器，耗时达数秒至十秒，且强制要求开发者本地安装了 Python 和 pipx；
2. **缺乏时序流水线**：Fixer（自动修复）、Formatter（排版对齐）、Linter（静态审查）混杂在一起，无序执行导致修复后的代码可能再次违规；
3. **工具死锁不可平替**：工具与脚本强绑定。当某个工具停更（如 golint/gosec/tslint）或社区出现 100 倍速的新工具（如 ruff, biome, oxlint）时，无法无痛平替；
4. **缺乏场景解耦 (Local vs CI)**：未能清晰区分“本地极速防线（防泄密、防 Git 污染）”与“CI 深度审计（CVE 漏洞库、SBOM）”；
5. **缺乏文件自适应与过滤**：缺少对扫描后缀、指定目录白名单及 `node_modules`/`dist`/`vendor` 等重型构建目录的直切剪枝；
6. **Web 协同缺失**：Web 界面缺少可视化的守卫状态展现、提交拦截结构化诊断和一键修复能力。

---

## 2. 核心系统架构与功能规范

### 2.1 六大领域分类 (The 6 Pillars)
所有规则与工具按职责划分为 6 大领域：
1. **`hygiene` (文本与 Git 仓库卫生)**：防大文件 (`check-added-large-files`)、防未解决合并冲突 (`check-merge-conflict`)、防大小写冲突 (`check-case-conflict`)、防 Lockfile 冲突，以及行尾空白 (`trailing-whitespace`) 与末尾换行 (`end-of-file`)；
2. **`format` (代码排版统一)**：`gofmt`, `prettier`, `taplo`, `shfmt`, `ruff format`；
3. **`lint` (静态代码分析 / SAST)**：`golangci-lint`, `eslint`, `ruff check`, `shellcheck`, `actionlint`；
4. **`secret` (敏感凭证拦截 - 本地必须绝对前置)**：`gitleaks protect --staged`, `detect-private-key`, `check-env-safety`；
5. **`security` (依赖漏洞与供应链安全 / SCA - CI 深度审计)**：`trivy`, `osv-scanner`, `govulncheck`, `npm audit`, `pip-audit`, `zizmor`；
6. **`policy` (流程与开源合规)**：`license-check` (文件头开源协议检查), `commitlint` (提交信息规范)。

### 2.2 三阶确定性执行流水线 (Pipeline: Fixer ➔ Formatter ➔ Linter)
1. **Phase 1: Fixer (自动修复)**：执行文本卫生修正（BOM、空白、EOF）及语法级修复（`goimports`, `ruff --fix` 等），直接原地修改文件；
2. **Phase 2: Formatter (格式对齐)**：对修复后的代码进行视觉样式“熨烫整平”，消除缩进、单双引号和换行噪音；
3. **Phase 3: Linter (只读裁决守卫)**：以最终文件状态进行只读语义与质量检查，发现违规直接退出非 0 并给出可读报告，绝不篡改代码。

### 2.3 两种扫描模式 (Scan Modes)
- **Git 变动模式 (`changed` / `staged`)**：
  - 默认模式。仅针对 Git 修改文件或暂存区（`--staged`）扫描，毫秒级响应，专用于本地 `git commit`；
- **全仓库模式 (`all`)**：
  - 递归遍历全仓库文件，用于 CI 验证（`unirtm check --all`）或初次接入全量审查。

### 2.4 三层文件过滤系统 (File Filtering System)
- **扫描后缀 (`extensions`)**：按文件扩展名精准匹配（如 `.go`, `.ts`, `.vue`, `.py`）；
- **限制目录 (`include_dirs`)**：指定扫描目录白名单（如 `src`, `cmd`, `internal`）；
- **排除目录与文件 (`exclude_dirs`, `exclude_files`)**：
  - 遇到 `node_modules`, `vendor`, `dist`, `build`, `target`, `.git` 时使用 `filepath.SkipDir` 快速剪枝，杜绝深层递归；
  - 支持排除特定的 glob 模式（如 `*.min.js`, `*.lock`）。

### 2.5 工具插槽与可配置平替 (Pluggable Slot & Adapter Architecture)
- 抽象标准插槽（如 `web.formatter`, `web.linter`, `go.linter`, `go.security`）；
- 当 A 工具停更或团队需要选用 B 工具时，无需修改任何代码，仅需在 `.unirtm.toml` 中切换：
  - 槽位平替：`[guard.slots] "web.formatter" = "biome"`
  - 自定义工具挂载：`[[guard.tools]] name = "oxlint" slot = "web.linter" command = "oxlint {{files}}"`

### 2.6 原生配置文件无缝继承 (Native Config Inheritance)
- 默认自动探测并挂载各生态的原生标准配置文件：
  - `.golangci.yml`, `.prettierrc`, `.eslintrc.*`, `ruff.toml`, `.gitleaks.toml` 等；
  - 支持显式指定配置文件路径。

### 2.7 双画像解耦 (Local Pre-commit vs CI Security Profile)
- **`profile = "pre-commit"`**：
  - 仅扫描暂存区，包含 Hygiene, Fixer, Formatter, 快速 Lint, 以及 `gitleaks protect --staged`（泄露拦截）；
  - 耗时控制在 100~200ms 以内，完全离线；
- **`profile = "ci"`**：
  - 全量扫描，只读裁决，激活深度依赖漏洞库（Trivy, OSV, Govulncheck）与 GitHub Actions 工作流安全审计（Zizmor）。

### 2.8 Web UI 可视化质检与交互中心 (Quality Guard Hub)
- **SCM 侧边栏提交区集成**：
  - 守卫状态指示胶囊（🟢 通过 / 🟡 待修复 / 🔴 阻断警告）；
  - `[⚡ 一键全自动修复]` 与 `[🔍 预跑质检]` 快捷按钮；
  - 提交被拦截时展开结构化诊断列表，支持一键修复或在 Monaco 编辑器中定位；
- **质量大盘 / 规则管理抽屉**：
  - 智能生态卡片（展示已激活的 Go / Web / Python 生态与守卫工具）；
  - 规则独立开关（Switch Toggle），单源真理双向同步回 `.unirtm.toml`；
  - 标签式目录与后缀配置（Include / Exclude Dirs, Extensions）。

---

## 3. 验收标准
1. **Go 原生化替代**：核心通用文本卫生与安全检测（EOF, Whitespace, BOM, 大文件, 冲突标记, 敏感私钥探测）采用纯 Go 原生实现，消除对 30+ Python 脚本的依赖；
2. **三阶执行顺序**：`unirtm fix` / `unirtm check` 严格按照 `Fixer -> Formatter -> Linter` 管道执行；
3. **模式与过滤生效**：支持 `--staged`, `--all`，以及 `extensions`, `include_dirs`, `exclude_dirs`，且对 `node_modules` 做到零深入快速跳过；
4. **工具平替与原生配置**：支持在 `.unirtm.toml` 中配置自定义工具替换已有槽位，并正确读取 `.golangci.yml` 等原生配置文件；
5. **双 Profile 分离**：`pre-commit` 极速前置防泄密，`ci` 深度审计漏洞；
6. **Web 协同闭环**：Web UI 拥有守卫胶囊、一键修复按钮、规则开关及诊断抽屉；
7. **稳定性保障**：所有 Go 单元测试与端到端测试 100% 通过，前端类型检查与单元测试 0 报错。
