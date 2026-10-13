# Feature Specification: Network Profiles, Domestic Mirror Matrix & Cascading Overrides

**Feature Identifier**: `028-network-profile-and-mirror-matrix`  
**Created**: 2026-10-13  
**Status**: Draft  
**Input**: "针对编译前网络与环境准备，建立完整网络管理体系：支持多语言生态国内官方/大厂镜像矩阵（可内置下拉选择）；坚决杜绝内置任何第三方代理服务器与GitHub代理（默认官方直连）；支持Profile完整生命周期（新建、修改、另存为、恢复出厂默认值、删除）；支持单语言生态（npm, go, python, rust, ruby等）独立自定义覆盖（Overrides > Profile）；实现CLI与Web功能完全对齐，全方位母语化、本地化与国际化支持。"

---

## 1. 核心背景与架构定位 (Background & Motivation)

### 1.1 开发者工具生态位差异
现代开发工具链存在明确的阶段分工：
- **VS Code / Antigravity / AI Agent**：位于**编码、语法分析与编译阶段**（Code & Build Phase），致力于代码生成、补全与构建触发。
- **UniRTM (Universal Runtime Manager)**：位于**编译前的基础设施与网络准备阶段**（Pre-Flight & Environment Setup）。
- **痛点**：在 UniRTM 之前，开发者面临繁琐的手工配置（手动安装各语言管理工具、频繁被网络墙阻塞 `go get` / `npm install` / `pip install`、各环境缺少统一 Shims 路由）。若编译前环境无法就绪，再先进的 IDE 和 AI Agent 也无法完成编译。

### 1.2 安全红线与架构铁律 (Security Boundary & Inviolable Principles)
1. **公开镜像源可内置、可下拉**：阿里云、腾讯云、华为云、清华大学、中科大、npmmirror、goproxy.cn、Ruby China 等公开合规的镜像站属于基础设施，系统提供官方内置列表与下拉快速选择。
2. **严禁内置任何代理服务器与第三方 GitHub 代理**：
   - 代理服务器是用户本地开发环境（如 127.0.0.1:7890）或企业内网网关的私有资产，输入框默认必须为空、开关默认必须关闭。
   - 严禁内置任何形如 `ghproxy` 等第三方自建或不知名 GitHub 代理服务，杜绝中间人攻击 (MITM) 劫持、代码篡改及后门注入风险。GitHub 默认必须永远走官方直连通道。
3. **彻底母语化、本地化（拒绝中英文杂合）**：
   - 所有界面展示与 CLI 输出必须基于用户当前选定语言进行纯正母语化呈现，坚决杜绝 `关于 (About)` 或中英文括号夹杂的陈旧设计。

---

## 2. 用户场景与验收准则 (User Scenarios & Acceptance Criteria)

### User Story 1 - Profile 完整生命周期管理 (Priority: P1) 🎯 MVP

作为跨网络环境协作的开发者，我希望能够在不同的网络档案（Profile）之间一键切换，并且支持新建、编辑、另存为以及恢复出厂默认值，以便在家庭网络、办公网络、海外 CI 流水线之间快速对齐。

**独立测试准则**：
通过 CLI 执行 `unirtm network profile use cn` 以及在 Web 仪表盘点击卡片，验证全生态源瞬间对齐；通过 `unirtm network profile reset cn` 验证内置 Profile 无论被如何修改都能一键还原出厂值。

**验收场景**：
1. **Given** 处于全新环境，**When** 运行 `unirtm network profile list` 或打开 Web 设置，**Then** 展示内置的 `official`（官方直连）与 `cn`（国内镜像加速矩阵）档案。
2. **Given** 用户当前生效的配置，**When** 调用 `unirtm network profile save-as office-net` 或 Web 点击“另存为新档案”，**Then** 系统创建独立的自定义 Profile 并持久化至配置文件。
3. **Given** 用户修改了内置的 `cn` 档案内容导致不可用，**When** 执行 `unirtm network profile reset cn` 或点击“恢复出厂默认值”，**Then** 系统瞬间还原为原厂推荐的公共大厂高校镜像源。
4. **Given** 用户创建的自定义 Profile，**When** 执行 `unirtm network profile delete <name>` 或 Web 点击删除，**Then** 该自定义档案被安全移除，激活状态自动平滑回退至默认 Profile。

---

### User Story 2 - 单语言生态独立覆盖 (Cascading Overrides) (Priority: P1)

作为具有特定语言需求的开发者（例如仅 npm 需要使用企业内网私有库，而 Go 和 Python 依然使用公共国内镜像），我希望能够独立更改单个或多个语言的镜像源，而不必受限于 Profile 整体，同时随时能够重置回跟随 Profile。

**独立测试准则**：
设置 `unirtm network registry set npm https://npm.corp.internal`，验证只有 npm 被独立覆盖，其他语言依然继承当前 Profile；执行 `unirtm network registry reset npm` 验证 npm 重新恢复跟随 Profile。

**验收场景**：
1. **Given** 当前生效 Profile 为 `cn`，**When** 用户单独修改了 `npm` 为私有源地址，**Then** 解析器根据 `Overrides > Profile > Default` 原则，npm 生效私有源，其余语言继续生效 `cn` 镜像源。
2. **Given** 存在单项覆盖状态，**When** 在 Web 端查看时，该语言行展示醒目的 `[单项自定义覆盖]` 徽章，并提供快捷 `[↺ 恢复跟随预设]` 按钮；CLI 在 `unirtm network status` 中明确标记来源。
3. **Given** 用户点击恢复跟随或执行 CLI `reset`，**Then** 单项覆盖被清除，该语言立即重新动态跟随当前激活 Profile 的值。

---

### User Story 3 - 代理服务器安全解耦与原生沙箱隔离 (Priority: P1)

作为注重网络安全的企业开发者，我希望系统对代理配置保持绝对的安全中立，不内置任何不可信公网代理或 GitHub 镜像中继，仅在需要时通过环境变量沙箱将我自有的本地/内网代理注入到执行命令中。

**独立测试准则**：
初始安装后检查配置文件和 Web 界面，代理必须为关闭状态、地址完全为空，GitHub 下载完全走官方直连；配置本地代理后，运行任务时子进程正确接收 `HTTP_PROXY`/`NO_PROXY` 环境变量，宿主系统不被污染。

**验收场景**：
1. **Given** 全新配置，**When** 检查 `network.proxy`，**Then** `enabled` 为 `false`，`http`、`socks5` 为空字符串，无任何第三方代理服务。
2. **Given** 用户开启代理并填写 `http://127.0.0.1:7890`，**When** 执行 `unirtm run`、shims 调度或下载工具链，**Then** 引擎自动在子进程沙箱注入 `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, `NO_PROXY`。
3. **Given** 任务执行结束退出，**When** 检查系统宿主环境变量和 `~/.bashrc`，**Then** 宿主环境保持未被篡改的洁净状态。

---

### User Story 4 - 实时网络延迟测速矩阵与智能决策 (Priority: P2)

作为开发者，我希望能够直观看到当前配置下各官方源和国内主流镜像站的实际网络往返延迟 (RTT)，帮助我做出最优的网络选择。

**独立测试准则**：
执行 `unirtm network test` 或在 Web 页面点击“实时测速”，系统并发探测各节点延迟并返回带颜色分级的毫秒值。

**验收场景**：
1. **Given** 用户触发测速，**When** 系统并发探测 npmmirror、npmjs、goproxy.cn、golang.org 等端点，**Then** 2 秒内汇总出各端点 RTT 毫秒延迟与状态（极佳 <50ms、良好 <200ms、缓慢 >500ms、超时）。
2. **Given** 检测到官方源大面积超时而国内源通畅，**When** 用户查看 Dashboard 或 CLI，**Then** 给出人性化的智能加速建议引导（“检测到海外依赖拉取较慢，推荐一键切换至国内加速矩阵”）。

---

### User Story 5 - CLI 与 Web 功能完全对齐与全方位母语化 (Priority: P1)

作为多端使用的开发者，我希望无论在 CLI 命令行还是在 Web 工作台界面中，所有操作能力（Profile 增删改查、单项重写、代理设置、测速）完全对齐，并且支持 100% 纯正母语化。

**独立测试准则**：
在 Web 界面执行的操作能被 CLI 实时反映，在 CLI 执行的修改能被 Web 自动重载刷新；切换语言为简体中文、正体中文、英文时，所有网络相关的术语、提示、按钮和错误文案均完整本地化。

---

## 3. 功能需求清单 (Functional Requirements)

- **FR-001 (配置分层结构)**：系统配置文件 MUST 在 `[network]` 根表下严格划分 `proxy`、`profiles`、`overrides` 三大子模块。
- **FR-002 (代理安全默认值)**：代理配置中的 `enabled` MUST 默认为 `false`，`http`、`socks5` MUST 默认为空，严禁内置任何公网代理或 GitHub 第三方中继。
- **FR-003 (内置 Profile 预设)**：系统 MUST 内置不可删除的原厂出厂档案：
  - `official`：全量指向各大语言官方注册表与官方分发 CDN。
  - `cn`：全量指向国内顶级公共镜像（npm -> npmmirror, go -> goproxy.cn, python -> 清华大学, rust -> rsproxy, ruby -> ruby-china）。
- **FR-004 (Profile CRUD 支持)**：系统 MUST 支持用户创建、修改、保存、重命名和删除自定义 Profile。
- **FR-005 (出厂默认值重置)**：系统 MUST 提供针对内置 Profile 的“出厂还原”功能，无论用户如何修改 `cn` 档案，均可一键恢复初始值。
- **FR-006 (单项覆盖优先原则)**：解析器在决定某语言生效的镜像源时，MUST 遵循优先级链：
  `network.overrides[lang] > active_profile.registries[lang] > official_fallback[lang]`。
- **FR-007 (单项覆盖重置)**：系统 MUST 支持单独清除某一语言的覆盖记录，使其立即恢复跟随当前 Profile。
- **FR-008 (Shims & Task 环境变量沙箱注入)**：UniRTM 在调度 shims 二进制或执行任务时，MUST 自动注入对应环境变量：
  - 代理：`HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, `NO_PROXY`。
  - npm：`NPM_CONFIG_REGISTRY`。
  - Go：`GOPROXY`。
  - Python：`PIP_INDEX_URL`。
  - Rust：`CARGO_REGISTRIES_CRATES_IO_INDEX`。
- **FR-009 (CLI 子命令对齐)**：`unirtm network` MUST 提供 `status`, `profile [list|use|create|save-as|set|reset|delete]`, `registry [set|get|reset]`, `test`, `proxy [on|off|set]` 子命令。
- **FR-010 (Web REST API 接口)**：UniRTM 后端服务 MUST 提供 `/api/network/status`, `/api/network/profile`, `/api/network/registry`, `/api/network/test`, `/api/network/proxy` 等标准 RESTful 接口。
- **FR-011 (Web 视觉控制中心)**：Web 设置中心 MUST 提供“网络代理”与“镜像源管理”两个清晰独立的板块；各语言行支持下拉推荐源、自定义输入、覆盖徽章标记与一键撤销。
- **FR-012 (并发 RTT 测速)**：测速模块 MUST 采用并发探测，超时时间严格控制在 3 秒以内，避免阻塞用户界面或命令行。
- **FR-013 (全量母语化与国际化)**：所有新增文本 MUST 录入 i18n 词典，支持简体中文、繁体中文、英文等语言，严禁中英文混合。

---

## 4. 边界与异常情况处理 (Edge Cases)

1. **当前激活的 Profile 被意外删除**：
   若用户在 CLI 删除了当前正在使用的自定义 Profile，系统 MUST 自动安全回退激活 `cn`（或 `official`），并记录警告日志，绝不抛出崩溃。
2. **填写的代理地址格式无效**：
   校验器 MUST 对 `http://`、`https://`、`socks5://` 协议及 `host:port` 格式进行验证，格式错误时阻断保存并提供明确提示。
3. **网络断网或镜像探测全部失败**：
   测速探测遇到完全断网时，优雅标记所有节点为 `TIMEOUT`，绝不造成死锁或崩溃。
4. **单项覆盖输入为空**：
   若用户清空了单项覆盖的输入框并保存，系统自动等价于执行 `reset` 操作，恢复跟随 Profile。
