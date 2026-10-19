# 任务拆分与执行跟踪清单：开发环境工具链状态折叠分组面板 (Tasks)

## 任务执行进度概览
- **当前阶段**：全部完成 (100%)
- **总体进度**：12 / 12 完成 (100%)

---

## Phase 0: 方案设计与基线建立 (Docs Baseline)
- [x] Task 0.1: 创建 `specs/051-tools-status-accordion-sidebar/` 规范、方案与任务文档
- [x] Task 0.2: 检出特性分支 `051-tools-status-accordion-sidebar` 并提交文档基线

---

## Phase 1: 状态分类派生与计算逻辑 (Tool Status Categorization)
- [x] Task 1.1: 在 `SidebarPanel.vue` 中集成 `computeToolStatus` 计算每个工具的状态类别（`missing` / `upgradable` / `ready`）
- [x] Task 1.2: 派生各分组响应式计算属性：`missingTools`, `upgradableTools`, `readyTools` 及搜索匹配
- [x] Task 1.3: 编写工具状态分组与派生算法的单元测试
- [x] Task 1.4: 阶段性 Git 提交：`feat(tools): implement status-based categorization for sidebar tools`

---

## Phase 2: 折叠面板 UI 渲染与交互动效 (Accordion UI & Interaction)
- [x] Task 2.1: 在 `SidebarPanel.vue` 模板中重构工具区域为独立 Accordion 折叠分组
- [x] Task 2.2: 实现“⚠️ 待安装 (N)”分组（N > 0 时置顶呈现，琥珀色警告视觉）
- [x] Task 2.3: 实现“🟢 可升级 (N)”分组（N > 0 时展示，绿色升级小圆点）
- [x] Task 2.4: 实现“已就绪 / 已安装 (N)”分组与展开/折叠交互
- [x] Task 2.5: 阶段性 Git 提交：`feat(sidebar): render collapsible accordion groups for tools by status`

---

## Phase 3: 53 种国际化本地化与全量回归 (i18n & Regression)
- [x] Task 3.1: 在 53 种语言包中增加状态分组标题文案（`sidebar.tools_missing`, `sidebar.tools_upgradable`, `sidebar.tools_ready`）
- [x] Task 3.2: 运行全量单元测试与类型检查确保 0 报错
- [x] Task 3.3: 阶段性 Git 提交：`feat(i18n): complete 53 locales and finalize tools status accordion`
