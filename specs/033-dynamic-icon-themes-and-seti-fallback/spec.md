# Feature Specification: 033-dynamic-icon-themes-and-seti-fallback

**Feature Branch**: `033-dynamic-icon-themes-and-seti-fallback`  
**Created**: 2026-10-15  
**Status**: Draft  
**Input**: 用户需求："通过speckit制定计划：1. 将Unirtm WEB中的图标全面接入图标集，方便我们根据图标集进行动态更换。 2. 如果图标不够，fallback到默认图标，Seti UI。"

---

## 1. 概述与核心目标 (Summary & Objectives)

在 UniRTM WEB 端中，目前部分视图（如 Runtimes 与 Tools 列表）已经接入了基础的 VS Code 图标解析系统，但整个工作台（ActivityBar 导航栏、Sidebar 面板、Dashboard 概览、Doctor 诊断、Tasks 任务运行器、Marketplace 市场、Settings 设置页及 BottomStatusBar 状态栏）仍散落着大量硬编码的 Unicode Emoji（如 📊、🌐、🛠️、⚡、⚙️、ℹ️ 等）以及静态符号。这导致在切换第三方图标主题时，界面无法实现整体统一的视觉换肤。

本特性旨在：
1. **全面接入图标集 (Universal Icon Theme Wiring)**：将 UniRTM WEB 端所有视图与导航组件中的图标全面接入图标主题引擎（统一通过 `<TechIcon>` / `<UniIcon>` 动态渲染）。建立语义化工作台图标映射表，使工作台各类图标均可根据当前激活的图标主题包动态更换。
2. **多层级回退保障机制 (Tiered Seti UI Fallback Engine)**：当用户激活的第三方图标主题（如来自于 OpenVSX 的 Material Icon Theme 或第三方图标包）未定义某些语言、工具或工作台图标时，系统**无缝、优雅地逐级回退至默认图标集 Seti UI**（优先回退到 Seti UI 中该类型/语言的具体图标，若仍无匹配则回退到 Seti UI 的默认文件图标 `_default`）。确保绝对不会出现空白方块、损坏图标或布局错乱。

---

## 2. 用户场景与测试用例 (User Scenarios & Testing)

### User Story 1 - 全面图标集接入与全局动态更换 (Priority: P1)

**作为** UniRTM 用户与开发者，  
**我希望** 工作台所有模块（ActivityBar、Sidebar 面板、Dashboard 卡片、Doctor 检查项、Tasks 任务、Marketplace 扩展项、Settings 导航项、BottomStatusBar 指示器）中的图标都由图标集驱动，  
**以便于** 当我在市场中安装并切换任何第三方图标主题包（或在浅色/深色图标主题间切换）时，整个 Web 工作台的图标能够瞬时响应、统一更换风格。

- **独立可测性**:
  - 检查工作台所有界面，不存在硬编码的杂色 Emoji；
  - 在设置或市场中切换图标集（如从 Seti UI 切换为 Material Icons 或自定义图标集），ActivityBar、Sidebar 及各主要卡片图标立刻同步变换为其对应主题风格的图标。

- **验收场景**:
  1. **Given** 用户处于 UniRTM Web 工作台，**When** 打开任何视图（如 Dashboard、Tasks、Doctor、Settings），**Then** 所有功能图标均通过 `<TechIcon :id="identifier" />` 动态渲染，无任何原生系统 Emoji。
  2. **Given** 用户从市场安装并激活了一个新的图标集，**When** 图标主题应用生效，**Then** 工作台所有图标在 16ms（下一帧）内全部无刷新热替换为新图标集的对应图标。

---

### User Story 2 - 第三方图标缺失时优雅回退到 Seti UI (Priority: P1)

**作为** UniRTM 用户，  
**我希望** 当我使用第三方图标集时，如果该图标集覆盖不全（例如缺少某些小众编程语言、UniRTM 专属配置文件或工作台特定标识），  
**能够** 自动、平滑地 fallback 到默认图标集 Seti UI，  
**使得** 界面始终美观完整，永远不出现裂图、空白字符或方块乱码。

- **独立可测性**:
  - 激活一个只包含常见语言图标的简化第三方图标集；
  - 访问一个未被该第三方主题覆盖的工具/文件（如 `unirtm.lock`、`unirtm.toml`、`zig` 或某个工作台导航项）；
  - 观察其渲染结果：自动呈现 Seti UI 对应的规范矢量字体图标，且颜色完全符合 Seti UI 调色板。

- **验收场景**:
  1. **Given** 激活的第三方主题未定义 `unirtm.lock` 或 `unirtm.toml`，**When** 渲染该项，**Then** 解析器在第三方主题中查找未命中后，自动在 `defaultSetiManifest` 中匹配到 Seti UI 的 `_lock` (`\uE05E`) 或 `_config` (`\uE019`) 并精准渲染。
  2. **Given** 用户输入一个完全未知的自定义工具名（如 `my-custom-cli`），无论当前主题还是 Seti UI 都没有对应语言项，**When** 执行解析，**Then** 最终回退至 Seti UI 的 `_default` 文件图标 (`\uE023`, `#d4d7d6`)。

---

### User Story 3 - 统一的 `<TechIcon>` / `<UniIcon>` 语义化调用与开发者体验 (Priority: P2)

**作为** UniRTM 前端工程师，  
**我希望** 可以在任何 Vue 组件中通过简明统一的组件调用图标，既支持技术/语言/文件名（如 `id="go"`、`id="Cargo.toml"`），也支持工作台语义标记（如 `id="workbench:settings"`、`id="workbench:doctor"`），  
**以便于** 后续维护工作台视图时无需关心底层是 SVG、Font-Face 还是 Seti Fallback。

- **独立可测性**:
  - 在组件中引入 `<TechIcon id="workbench:settings" :size="18" />`，查看渲染 DOM 树：生成语义化标签并具备完善的 fallback 链路，无任何控制台警告。

- **验收场景**:
  1. **Given** 传入 `id="workbench:tasks"`，**When** 解析器处理，**Then** 映射至语义定义 `task` / `sh`，在当前主题查找，未定义则在 Seti UI 中呈现 shell 符号。
  2. **Given** 组件属性 `:colored="false"`，**When** 渲染在特定高亮按钮中，**Then** 图标颜色继承父级 `currentColor`。

---

## 3. 功能需求清单 (Functional Requirements)

- **FR-001 (全面排查与替换)**: 系统必须对 `frontend/src` 下所有 Vue 视图和组件进行全面扫描与重构，将所有硬编码的 Emoji 图标（导航、卡片、状态、列表项）替换为统一的 `<TechIcon>` 组件。
- **FR-002 (工作台语义映射表)**: 系统必须建立结构化的工作台语义图标映射字典（`workbenchSemanticIcons`），将工作台操作与状态（如 `dashboard`, `runtimes`, `tools`, `tasks`, `marketplace`, `doctor`, `settings`, `about`, `terminal`, `success`, `warning`, `error` 等）映射为标准的主题检索标识符。
- **FR-003 (双层级回退解析机制)**: `IconThemeService.resolveIcon` 必须实现严格的双层级回退算法：
  1. **第 1 层**: 检查当前激活主题的 `fileNames` -> `fileExtensions` -> `languageIds`；
  2. **第 2 层 (Seti UI 回退)**: 若第 1 层未命中或无效，自动检索内置 `defaultSetiManifest` 的 `fileNames` -> `fileExtensions` -> `languageIds`；
  3. **终极兜底**: 若 Seti UI 亦未精确命中，回退到 Seti UI 的默认图标 `_default` (`\uE023`)。
- **FR-004 (混合渲染支持)**: 系统必须完美支持当前主题为 SVG 图标（如 Material Icon Theme）而回退图标为 Seti 字体图标（Woff2）的混合渲染场景，样式与尺寸绝对对齐。
- **FR-005 (缓存与性能)**: 图标解析器必须引入基于 LRU/Map 的内存缓存，确保同一标识符在同一次主题生命周期内解析耗时小于 0.01ms，对渲染性能零损耗。
- **FR-006 (国际化与无障碍)**: 所有渲染的图标必须包含规范的 `role="img"` 及对应上下文的 `title` / `aria-label`。

---

## 4. 边界与异常处理 (Edge Cases)

1. **第三方主题格式不规范或缺失字体**: 若第三方主题未定义 `fonts` 且未提供 `iconPath`，直接命中 Seti UI Fallback。
2. **离线与断网环境**: 内置的 Seti UI 字体与定义全部为本地打包资产，在任何无网络、内网穿透或离线开发机环境下 100% 正常显示。
3. **极端大小缩放**: 支持 `:size` 任意数值（12px 到 64px），字体与 SVG 均自适应容器居中，不破坏现有 Flex/Grid 布局。
