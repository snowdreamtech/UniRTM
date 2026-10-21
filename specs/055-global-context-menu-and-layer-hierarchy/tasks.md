---
description: "Task list for feature 055-global-context-menu-and-layer-hierarchy"
---

# Tasks: 全局右键上下文菜单与窗口层叠体系架构 (055-global-context-menu-and-layer-hierarchy)

**Input**: Design documents from `/specs/055-global-context-menu-and-layer-hierarchy/`
**Prerequisites**: plan.md, spec.md

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 初始化与基线验证

- [x] T001 验证当前代码基线与 Vitest 测试套件运行状态
- [x] T002 检查当前 `FileContextMenu.vue`、`TerminalContextMenu.vue` 与 `GitFileContextMenu.vue` 的实现差异与定位机制

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 建立 UniRTM 7 层 Z-Index 标准变量与全局层叠体系架构（US3 基础）

- [x] T003 定义全局层叠变量体系在 `frontend/src/styles/theme.css` 中（包含 `--z-base`, `--z-sidebar`, `--z-dropdown`, `--z-context-menu`, `--z-modal`, `--z-toast`, `--z-tooltip` 等标准层级）
- [x] T004 在全局样式或通用菜单样式中加入统一的 `.context-menu-backdrop` 点击遮罩与浮动容器样式规范

**Checkpoint**: 全局层叠变量与菜单基础设施就绪，开始实施具体组件全局传送。

---

## Phase 3: User Story 1 - Explorer File Context Menu Global Teleportation & Edge Visibility (Priority: P1) 🎯 MVP

**Goal**: 解决 Explorer 文件树右键菜单被侧边栏 overflow / backdrop-filter 裁切问题，使菜单传送至 body 并具备视口智能碰撞检测。

**Independent Test**: 打开 `http://localhost:5173/explorer?scope=workspace`，在文件列表任意文件/文件夹右键，验证菜单完整悬浮展示在最上层、文字无截断，且贴近视口右/下边界时自动向内翻转自适应。

### Implementation for User Story 1

- [x] T005 [US1] 在 `frontend/src/components/explorer/FileContextMenu.vue` 中引入 `<Teleport to="body">` 容器包裹
- [x] T006 [US1] 在 `FileContextMenu.vue` 中增加全局透明捕获遮罩 `<div class="context-menu-backdrop">`，支持全局任意区域点击与右键拦截即刻自闭合
- [x] T007 [US1] 在 `FileContextMenu.vue` 中实现基于视口尺寸（`window.innerWidth`, `window.innerHeight`）的智能防溢出翻转计算属性（`adjustedStyle`）
- [x] T008 [US1] 绑定 Esc 按键监听与全局 Resize 监听，窗口发生变化或按 Esc 时平滑关闭菜单
- [x] T009 [US1] 接入 `--z-context-menu` 标准变量，验证新建文件、重命名、删除、在终端打开等所有右键菜单业务操作功能正常

**Checkpoint**: Explorer 文件树右键菜单在任意窗口尺寸下 100% 完整显示，无被裁剪现象，MVP 验证通过。

---

## Phase 4: User Story 2 - Terminal Context Menu Teleportation & Smart Alignment (Priority: P2)

**Goal**: 将终端（Terminal）右键上下文菜单同样规范为 `<Teleport to="body">` 并加上智能防越界保护。

**Independent Test**: 打开底部终端面板，在终端输出区域右下角右键，验证菜单完整显示且在底部自动向上翻转。

### Implementation for User Story 2

- [x] T010 [US2] 在 `frontend/src/components/terminal/TerminalContextMenu.vue` 中加入 `<Teleport to="body">` 与全局点击遮罩
- [x] T011 [US2] 在 `TerminalContextMenu.vue` 中同步升级智能视口边界检测，防止贴底时菜单溢出终端面板
- [x] T012 [US2] 统一 `TerminalContextMenu.vue` 的 z-index 层级至 `var(--z-context-menu-popover, 1001)`

**Checkpoint**: 终端与文件树的右键菜单交互与层级体系达到 100% 架构一致。

---

## Phase 5: User Story 3 - UniRTM 7-Tier Z-Index Stacking Architecture Review (Priority: P2)

**Goal**: 全面评估并校验主要窗口与浮动容器（Sidebar, Drawer, Modal, Dropdown, Tooltip）的层叠优先级。

**Independent Test**: 验证弹窗（Modal）遮罩层高于 ContextMenu，ContextMenu 高于 Sidebar 和 Drawer，Tooltip 高于所有面板。

### Implementation for User Story 3

- [x] T013 [US3] 审查主要组件的 z-index 现状（包括 Modal 对话框、通知 Toast、顶栏下拉菜单等），替换随意散落的硬编码数值
- [x] T014 [US3] 确保各层级不会产生层叠上下文冲突或穿透

---

## Phase 6: Polish & Verification

**Purpose**: 质量把关、自动化测试与提交

- [x] T015 运行前端单元测试套件 `npm run test`（确保 34 个测试文件全部通过）
- [x] T016 运行前端类型检查 `npx vue-tsc --noEmit`（确保 0 类型报错）
- [x] T017 运行 UniRTM 质量守卫检查 `/home/ansible/.local/bin/unirtm check --staged`（确保质量守卫全绿）
- [x] T018 提交 Git Commit：`feat(ui): teleport context menus to body with viewport collision detection`
