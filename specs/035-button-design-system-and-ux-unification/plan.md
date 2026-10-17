# 技术架构设计：全局按钮设计系统与实现规划 (Plan)

## 1. 架构目标与设计原则
在不引入庞大第三方组件库的前提下，依托已有的 CSS 变量系统与 `TechIcon` 图标体系，建立一套低耦合、高性能且易于各组件直接采用的**原子化按钮体系**。

### 1.1 设计系统层级收敛表

```
┌─────────────────────────────────────────────────────────────┐
│                      Button Design System                   │
├─────────────────┬─────────────────────┬─────────────────────┤
│ 类别 (Variants) │ 尺寸 (Sizes)        │ 语义情景 (Semantics)│
├─────────────────┼─────────────────────┼─────────────────────┤
│ • .btn-primary  │ • .btn-md (32px)    │ • Default (中性)    │
│ • .btn-secondary│ • .btn-sm (26px)    │ • Success (成功/生效)│
│ • .btn-ghost    │ • .btn-icon-sm (26) │ • Danger (警告/删除)│
│ • .btn-danger   │ • .btn-icon-md (30) │ • Accent (高亮品牌) │
│ • .btn-capsule  │                     │                     │
└─────────────────┴─────────────────────┴─────────────────────┘
```

---

## 2. 全局样式规划 (`frontend/src/index.css`)

```css
/* Base Button Anchor */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-family: inherit;
  font-weight: 500;
  border: 1px solid transparent;
  cursor: pointer;
  user-select: none;
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
  box-sizing: border-box;
  line-height: 1;
}

.btn:active:not(:disabled) {
  transform: scale(0.96);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Primary */
.btn-primary {
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  color: #ffffff;
  border-color: rgba(2, 132, 199, 0.4);
  box-shadow: 0 2px 8px rgba(0, 229, 255, 0.18);
}
.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #00f0ff, #0369a1);
  box-shadow: 0 4px 14px rgba(0, 229, 255, 0.32);
}

/* Secondary */
.btn-secondary {
  background: var(--input-bg);
  border-color: var(--card-border);
  color: var(--text-main);
}
.btn-secondary:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
  border-color: var(--text-muted);
}

/* Ghost / Action Icon Button (行内小操作) */
.btn-icon {
  width: 26px;
  height: 26px;
  padding: 0;
  border-radius: 6px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
}
.btn-icon:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
}
.btn-icon.danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}
.btn-icon.success:hover:not(:disabled) {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
}
.btn-icon.accent:hover:not(:disabled) {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
}
```

---

## 3. 淘汰硬编码裸字符替换映射表

| 原有裸手写字符 | 所在模块 | 规范替换方案 |
| :--- | :--- | :--- |
| `↺` (恢复基准值) | `EnvironmentsView` | `<TechIcon id="workbench:sync" :size="13" />` 或 `history`，样式 `.btn-icon.success` |
| `✕` (删除 Profile) | `EnvironmentsView` | `<TechIcon id="workbench:close" :size="11" />`，样式 `.btn-icon.danger` |
| `AA` (全大写) | `EnvironmentModal` | `<TechIcon id="workbench:edit" :size="13" />` 或带统一边框的 `.btn.btn-sm` |
| `✕` (弹窗右上角关闭) | 各 Modal Header | `<TechIcon id="workbench:close" :size="14" />`，样式 `.btn-icon` |

---

## 4. 分步实施执行计划

- **Task 1**: 在 `frontend/src/index.css` 中沉淀全局按钮标准类；
- **Task 2**: 重构 `EnvironmentsView.vue` 与 `EnvironmentModal.vue`（彻底替换 `↺`、`✕`、`AA`，统一卡片/列表操作）；
- **Task 3**: 重构 `MarketplaceView.vue`（统一操作列、已安装卸载、应用、各分页与卡片操作）；
- **Task 4**: 重构 `ToolsView.vue` 与 `TasksView.vue`（统一运行、安装、刷新、升级操作）；
- **Task 5**: 重构 `SettingsView.vue` 与其他弹窗组件；
- **Task 6**: 运行全量测试、类型检查、构建打包与最终验收。
