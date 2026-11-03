# 任务拆分与执行跟踪清单：多存储库源代码管理系统 (Tasks)

## 任务执行进度概览
- **当前阶段**：Phase 5 - 全量回归与生产封箱 (100% 达成)
- **总体进度**：16 / 16 完成 (100%)

---

## Phase 0: 方案设计与基线建立 (Docs Baseline)
- [x] Task 0.1: 创建 `specs/050-multi-repository-source-control/` 规范、方案与任务文档
- [x] Task 0.2: 检出特性分支 `050-multi-repository-source-control` 并提交文档基线

---

## Phase 1: 多存储库数据模型与后端扫描 API (Data Model & Backend API)
- [x] Task 1.1: 在 Go 后端 (`cmd/64.ui.go`) 实现仓库目录解析助手与 `/api/git/repositories` 扫描端点
- [x] Task 1.2: 改造 Go 后端所有 `/api/git/*` 接口支持 `repo` 查询或负载参数
- [x] Task 1.3: 在 `frontend/vite.config.ts` 开发服务器中同步支持 `/api/git/repositories` 及 `repo` 参数
- [x] Task 1.4: 在 `useGit.ts` 中增加 `repositories`、`activeRepo`、`fetchRepositories()` 及参数透传
- [x] Task 1.5: 编写多存储库 Go 与前端状态管理单元测试
- [x] Task 1.6: 阶段性 Git 提交：`feat(scm): implement multi-repository discovery and repo-scoped APIs`

---

## Phase 2: SCM 侧边栏多存储库展示与交互 (Sidebar Repositories Section)
- [x] Task 2.1: 在 `SourceControlSidebar.vue` 中新增“存储库”可折叠分组与项目列表
- [x] Task 2.2: 实现存储库状态展示（分支徽标、脏状态标记 `*`、领先/落后指示、操作菜单）
- [x] Task 2.3: 实现活动存储库点击切换与下方“更改”区域联动刷新
- [x] Task 2.4: 提交信息输入框动态对齐当前选中存储库分支名称
- [x] Task 2.5: 阶段性 Git 提交：`feat(scm): add repositories section and active repo switching in sidebar`

---

## Phase 3: Git Graph 与底部状态栏多存储库联动 (Graph & Status Bar)
- [x] Task 3.1: 在 Git Graph 页面 (`SourceControlView.vue` / `GitGraphView.vue`) 顶栏工具栏增加 `Repo:` 仓库切换下拉选择器
- [x] Task 3.2: 改造底部状态栏 (`BottomStatusBar.vue`) 显示当前活动存储库名称与分支状态
- [x] Task 3.3: 阶段性 Git 提交：`feat(scm): integrate repo selector in git graph and status bar`

---

## Phase 4: 53 种国际化本地化与单测覆盖 (i18n & Localization)
- [x] Task 4.1: 在 53 种语言包中同步新增多存储库相关文案（`scm.repositories`, `scm.active_repo`, `scm.sidebar.commit_placeholder_branch`, `scm.graph.repo_label`）
- [x] Task 4.2: 运行国际化完整性单测确保 53 种语言完全一致
- [x] Task 4.3: 阶段性 Git 提交：`feat(i18n): complete 53 locales for multi-repository scm`

---

## Phase 5: 全量回归与生产封箱 (Regression & Release)
- [x] Task 5.1: 运行全量单元测试（`npm run test` 与 Go 测试）全部绿灯通过
- [x] Task 5.2: 运行类型检查与生产构建确保 0 警告 0 报错
- [x] Task 5.3: 阶段性 Git 提交：`docs(specs): finalize tasks tracking for multi-repository source control`
