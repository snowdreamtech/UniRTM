# 任务拆分与执行跟踪清单：工具链状态清晰化与低噪音升级系统 (Tasks)

## 任务执行进度概览
- **当前阶段**：Phase 5 - 顶部搜索栏待安装统计与一键过滤增强 (100% 封箱交付)
- **总体进度**：17 / 17 已完成 (100%)

---

## Phase 0: 方案设计与基线建立 (Docs Baseline)
- [x] Task 0.1: 创建 `specs/048-toolchain-status-clarity-and-upgrade-ux/` 规范、架构与任务文档
- [x] Task 0.2: 检出特性分支 `048-toolchain-status-clarity-and-upgrade-ux` 并提交文档基线

---

## Phase 1: 工具状态计算核心与测试 (State Machine & Unit Tests)
- [x] Task 1.1: 在前端提取并实现工具状态计算函数 `computeToolStatus`（支持 ready/upgradable/missing/unconfigured）
- [x] Task 1.2: 编写工具状态核心单测文件覆盖 4 种典型状态转换
- [x] Task 1.3: 阶段性 Git 提交：`feat(tools): implement toolchain state machine and unit tests`

---

## Phase 2: 卡片与表格视觉重构与降噪 (UI Refactor & Noise Reduction)
- [x] Task 2.1: 重构 `ToolsView.vue` 卡片头部徽标（就绪态去图标化并变更为中性底色）
- [x] Task 2.2: 注入专属绿色升级徽标（`↑ 可升级`）与琥珀色缺失安装徽标
- [x] Task 2.3: 同步适配表格视图（Table View）中的版本列展示与状态自适应
- [x] Task 2.4: 在缺失（Missing）状态下自适应卡片首选操作为一键安装
- [x] Task 2.5: 阶段性 Git 提交：`feat(tools): refactor tool badges for low noise and dedicated upgrade color`

---

## Phase 3: 53 种多语言本地化与文案对齐 (i18n & Localization)
- [x] Task 3.1: 在 53 种语言包中同步新增状态词条（如 `runtimes.status_missing`, `runtimes.status_upgradable` 等）
- [x] Task 3.2: 运行 `npm test -- src/i18n/i18n.test.ts` 确保全语言键值一致
- [x] Task 3.3: 阶段性 Git 提交：`feat(tools): complete 53 locales for toolchain status badges`

---

## Phase 4: 全量回归与生产封箱发布 (Regression & Release)
- [x] Task 4.1: 运行全量单元测试（`npm run test`）确保所有测试用例绿灯
- [x] Task 4.2: 运行生产环境打包（`npm run build`）确保 0 警告与 0 错误
- [x] Task 4.3: 阶段性 Git 提交：`feat(tools): complete toolchain status clarity and release feature`

---

## Phase 5: 顶部搜索栏待安装统计与一键过滤增强 (Toolbar Missing Count & Quick Filter)
- [x] Task 5.1: 在 `ToolsView.vue` 计算 `scopeMissingCount`，并在搜索工具栏渲染琥珀色 `⚠️ N 个待安装` 统计徽标
- [x] Task 5.2: 支持点击徽标快速切换“待安装”专属过滤筛选（再次点击恢复查看全部）
- [x] Task 5.3: 在 53 种语言包中同步新增 `runtimes.count_missing_suffix` 与 `runtimes.missing_tooltip`
- [x] Task 5.4: 运行全量单元测试与生产构建确保 0 报错
- [x] Task 5.5: 阶段性 Git 提交：`feat(tools): add toolbar missing counter and quick filter toggle`
