# 实现计划：全局右键上下文菜单与窗口层叠体系架构 (Plan 055)

**Feature Branch**: `055-global-context-menu-and-layer-hierarchy`
**Spec Reference**: `specs/055-global-context-menu-and-layer-hierarchy/spec.md`
**Status**: Ready for Implementation

---

## 1. 架构总览与技术方案

```
┌────────────────────────────────────────────────────────────────────────┐
│                        Document Body (Viewport)                        │
│                                                                        │
│  ┌───────────────────────┐         ┌────────────────────────────────┐  │
│  │     #app (Root)       │         │  Teleport Target: <body>       │  │
│  │                       │         │                                │  │
│  │ ┌───────────────────┐ │         │  ┌──────────────────────────┐  │  │
│  │ │  .sidebar-panel   │ │         │  │ .file-context-menu       │  │  │
│  │ │  backdrop-filter  │ │ ───────┼─>│ z-index: 1000            │  │  │
│  │ │ ┌───────────────┐ │ │Teleport │  │ Smart Placement:         │  │  │
│  │ │ │.content-list  │ │ │ (body)  │  │   left: min(x, W - w - 8)│  │  │
│  │ │ │overflow:hidden│ │ │         │  │   top: min(y, H - h - 8) │  │  │
│  │ │ │  [FileNode] ──┼─┼─┼─────────┘  │ 100% 完整无裁切展示!     │  │  │
│  │ │ └───────────────┘ │ │            └──────────────────────────┘  │  │
│  │ └───────────────────┘ │                                          │  │
│  └───────────────────────┘                                          │  │
└────────────────────────────────────────────────────────────────────────┘
```

### 核心解决策略：
1. **脱壳传送（Teleportation）**：将 `FileContextMenu.vue` 和 `TerminalContextMenu.vue` 通过 `<Teleport to="body">` 直接传送至 `document.body` 根节点，从物理结构上彻底摆脱父级 `.sidebar-panel` 的 `backdrop-filter: blur(...)` 局部包含块与 `.sidebar-content-list` 的 `overflow: hidden` 截断。
2. **视口边界碰撞与动态定位（Smart Placement）**：引入双向视口碰撞保护算法：
   $$\text{left} = \begin{cases} x, & x + \text{width} \le W - 8 \\ \max(8, W - \text{width} - 8), & \text{otherwise} \end{cases}$$
   $$\text{top} = \begin{cases} y, & y + \text{height} \le H - 8 \\ \max(8, H - \text{height} - 8), & \text{otherwise} \end{cases}$$
3. **全局层叠管理体系（Global Layer Hierarchy）**：在前端全局样式中正式收敛 Z-Index 变量与层叠金字塔（Layer 0: 内容基底, Layer 1: 固定布局, Layer 2: 抽屉面板, Layer 3: 普通下拉, Layer 4: 右键上下文菜单 1000, Layer 5: 模态遮罩 1500, Layer 6: 系统通知 2500+）。

---

## 2. 详细技术设计

### 2.1 FileContextMenu 全局化与视口智能计算
在 `frontend/src/components/explorer/FileContextMenu.vue` 中：
1. 使用 `<Teleport to="body">` 包裹根节点；
2. 注入全局捕获遮罩 `<div class="context-menu-backdrop" @click="closeContextMenu" @contextmenu.prevent="closeContextMenu">`，实现任何区域（包括 iframe、Monaco、终端、状态栏）点击或再次右键均能秒级收起；
3. 将位置属性从朴素的 `left: contextMenu.x, top: contextMenu.y` 升级为动态响应式 `computed` 坐标，确保在大屏、小屏、靠近右边缘或底边缘时均完美可见。

### 2.2 TerminalContextMenu 对齐规范
在 `frontend/src/components/terminal/TerminalContextMenu.vue` 中：
1. 同样包裹 `<Teleport to="body">`；
2. 对齐全局 `z-index: 1000` 与 `context-menu-backdrop`，消除在半屏/全屏切换或滚动时可能产生的边界问题。

### 2.3 UniRTM 全局 Z-Index 层叠规范沉淀
在 `frontend/src/styles/` 中明确标准化变量体系：
```css
:root {
  --z-base: 1;
  --z-sidebar: 20;
  --z-activity-bar: 30;
  --z-title-bar: 40;
  --z-status-bar: 40;
  --z-drawer: 50;
  --z-dropdown: 500;
  --z-context-menu: 1000;
  --z-modal: 1500;
  --z-toast: 2500;
  --z-tooltip: 3000;
}
```

---

## 3. 测试与验证方案

1. **功能验收测试**：
   - 在左侧 Explorer 文件树中右键点击文件与目录，验证菜单 100% 完整显示在最上层，右侧无任何截断；
   - 在屏幕右侧边缘右键，验证向左自适应贴边；
   - 点击外部任意位置，验证菜单立即销毁关闭；
   - 测试菜单内置快捷操作（新建文件、重命名、删除、在终端打开、Git 历史）功能正常执行。
2. **自动化测试套件**：
   - 运行 `npx vitest run`（确保 34 个测试文件全部通过）；
   - 运行 `npx vue-tsc --noEmit`（0 类型错误）。
3. **UniRTM 质量守卫**：
   - 运行 `/home/ansible/.local/bin/unirtm check --staged`（31 条守卫规则全绿）。
