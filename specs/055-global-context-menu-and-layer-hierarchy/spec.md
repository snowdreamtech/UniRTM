# Feature Specification: 055-global-context-menu-and-layer-hierarchy

**Feature Branch**: `055-global-context-menu-and-layer-hierarchy`
**Created**: 2026-10-22
**Status**: In Planning
**Input**: User description: "评估各主要窗口的层叠优先级，将左侧列表等右键菜单做成全局 Teleport，彻底解决被侧边栏 overflow / backdrop-filter 裁剪的问题"

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Explorer File Context Menu Global Teleportation & Edge Visibility (Priority: P1)

作为 UniRTM 开发者，在资源管理器（Explorer）文件树中的任何文件或文件夹上点击右键时，弹出的完整操作菜单（新建文件、新建文件夹、复制路径、重命名、删除、在终端打开、添加文件夹到工作区、Git 差异对比等）必须完整显示在最上层，不能被侧边栏边界截断。

**Why this priority**:
这是用户截图中直接暴露的核心体验缺陷。当前菜单向右延伸部分被 `SidebarPanel` 的局部包含块与 `overflow-x: hidden` 截断，导致用户无法辨识右侧选项文字，严重影响核心代码与文件管理操作。

**Independent Test**:
打开 `http://localhost:5173/explorer?scope=workspace`，在文件树最右侧或底部节点点击右键，验证菜单完整呈现、文字毫无截断，且无论侧边栏如何收窄、拖拽伸缩，菜单均能在最顶层展示。

**Acceptance Scenarios**:
1. **Given** 侧边栏宽度较窄（如 240px），**When** 用户在长文件名节点上点击右键，**Then** 菜单完整悬浮在编辑器内容区上方，文字清晰完整，无任何右侧边缘截断。
2. **Given** 鼠标点击位置接近屏幕右边缘，**When** 触发右键菜单，**Then** 菜单自动触发视口碰撞检测并向左翻转展示，菜单不溢出屏幕外。
3. **Given** 右键菜单处于打开状态，**When** 用户在编辑器任意空白处点击鼠标左键或右键，**Then** 菜单顺畅收起，无 DOM 残留。

---

### User Story 2 - Terminal & Global Context Menus Teleportation & Smart Flipping (Priority: P2)

作为用户，在底部抽屉终端（Terminal）或其它局部工作区组件中点击右键时，同样享有全局层叠与智能视口防越界保护，避免因终端面板滚动或贴边导致菜单显示不全。

**Why this priority**:
与 Explorer 文件树右键菜单同属上下文交互体系。必须保持全应用交互范式的一致性，防止同类问题在终端等其他区域再次复现。

**Independent Test**:
打开底部终端面板，在终端输出区域右键，验证菜单以 Teleport 挂载到 body，右键复制、粘贴、清屏等功能正常触发并自闭合。

**Acceptance Scenarios**:
1. **Given** 底部终端面板处于半屏高度，**When** 在终端右下角点击右键，**Then** 菜单自动向上翻转，所有菜单项均在视口可见区域内。

---

### User Story 3 - UniRTM 7-Tier Z-Index Stacking Architecture System (Priority: P2)

作为 UniRTM 前端架构，定义一套严密的 7 层层叠架构系统规范（Tokens 与 CSS 变量），统筹规范 Content, Layout, Drawers, Dropdowns, Context Menus, Modals, Tooltips 的 z-index 层级，杜绝随意硬编码 `z-index: 9999` 导致的层级踩踏与穿透。

**Why this priority**:
治标（单个菜单 Teleport）同时治本（系统级规约）。建立系统化规范才能保障未来新增 AI Agent 浮窗、代码补全 Widget 时不会破坏现有层叠关系。

**Independent Test**:
通过浏览器自动化或样式断言，验证 ContextMenu 位于 Modal 之下、Drawer 与 Sidebar 之上，Tooltip 与 Toast 处于最高优先级。

**Acceptance Scenarios**:
1. **Given** 打开模态弹窗（如删除确认弹窗），**Then** 模态层（z-index: 1500）覆盖在普通面板与右键菜单之上。
2. **Given** 鼠标悬浮在侧边栏图标上弹出 Tooltip，**Then** Tooltip（z-index: 3000）始终能清晰展现在最上层。

---

## Edge Cases

1. **多屏幕与超小视口**：当浏览器窗口被缩小到极端尺寸（例如宽 400px、高 300px）时，右键菜单计算位置后通过 `Math.max(8, ...)` 保证至少不越出屏幕负坐标，并支持滚动查看。
2. **多层嵌套右键连续触发**：用户先在一个文件上右键，不关闭菜单直接在另一个文件上右键，旧菜单应立即被平滑替换并重新定位，不能出现两个菜单重叠。
3. **Esc 快捷键与窗口 Resize**：用户按下 `Escape` 键或改变浏览器窗口大小时，活跃上下文菜单应自动平滑关闭。

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `FileContextMenu.vue` 必须使用 `<Teleport to="body">` 挂载到文档最顶层 DOM。
- **FR-002**: `FileContextMenu.vue` 必须具备基于视口 `window.innerWidth` 与 `window.innerHeight` 的智能碰撞检测（Smart Boundary Detection），超出右边界或底边界时自动向左/向上翻转或偏移。
- **FR-003**: 必须配置全局捕获遮罩或全局事件监听，支持全局任意区域点击（Click Outside）或再次右键即刻关闭菜单。
- **FR-004**: `TerminalContextMenu.vue` 必须同步升级为 `<Teleport to="body">`，确保全应用右键菜单机制规范统一。
- **FR-005**: 在前端样式系统（如 `src/styles/variables.css` 或核心样式表）中正式确立统一的 Z-Index 层叠规范（Layer 0 ~ Layer 6）。
- **FR-006**: 菜单中所有动作触发（新建、复制、删除、打开终端、差异比对）必须继续保持 100% 原有行为与业务参数传递，无回归破坏。

---

## Non-Functional Requirements

- **NFR-001**: 菜单弹出延迟低于 16ms（保持 60fps），无明显的跳动或重新计算抖动（FOUC）。
- **NFR-002**: 必须通过所有前端单元测试（`npm run test`）与类型检查（`vue-tsc`），确保 0 报错。
- **NFR-003**: 严格遵循 UniRTM 质量守卫（Quality Guard 38 条规则）。
