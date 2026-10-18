# 任务拆解与进度评估：Git 源代码管理系统 (Tasks)

## 任务分组与执行状态 (Tasks & Execution Status)

- [x] 1. **阶段 1：后端 Git API 接口与双端服务实现 (Backend API Implementation)**
  - [x] 1.1 在 `frontend/vite.config.ts` 中实现 Connect 中间件，支持 `/api/git/*`（`status`, `stage`, `unstage`, `discard`, `commit`, `diff`, `branches`, `checkout`）
  - [x] 1.2 在 `cmd/64.ui.go` 中实现 Go 原生 HTTP 处理函数，保持与 Vite 开发中间件的 JSON 数据结构完全一致
  - [x] 1.3 编写 `cmd/64.ui_test.go` 中的 `TestUIGitAPI` 单元测试，验证状态拉取、分支列出与差异获取
  - [x] 1.4 运行 `go test ./cmd/...` 确保所有后端测试通过

- [x] 2. **阶段 2：前端状态管理与活动栏入口 (Frontend State & ActivityBar)**
  - [x] 2.1 封装 `frontend/src/composables/useGit.ts`，支持响应式状态追踪、暂存/撤销/提交调度
  - [x] 2.2 编写 `frontend/src/composables/useGit.test.ts` 单元测试并通过 Vitest 验证
  - [x] 2.3 在 `frontend/src/components/common/TechIcon.vue` 中注册 `workbench:scm` (`codicon-source-control`) 等图标
  - [x] 2.4 在 `frontend/src/components/layout/ActivityBar.vue` 中挂载 SCM 图标，并配置动态变更数字 Badge
  - [x] 2.5 在 `frontend/src/router/index.ts` 与 `frontend/src/App.vue` 中配置 `/scm` 路由和标题

- [x] 3. **阶段 3：SCM 侧边栏与 Monaco 差异对比编辑器 (SCM Panel & Monaco DiffEditor)**
  - [x] 3.1 封装 `frontend/src/components/scm/SourceControlSidebar.vue`，实现分支切换浮层、提交文本框 (`Ctrl+Enter`)、暂存的更改与工作区更改列表
  - [x] 3.2 在 `frontend/src/components/layout/SidebarPanel.vue` 中集成 `SourceControlSidebar`
  - [x] 3.3 封装 `frontend/src/views/SourceControlView.vue`，实现 Monaco DiffEditor 差异视图与全景工作区概览
  - [x] 3.4 支持并排比对 (Side by Side) 与内联比对 (Inline) 模式切换
  - [x] 3.5 支持在比对视图中快捷暂存、取消暂存、放弃修改及“在代码编辑器中打开”

- [x] 4. **阶段 4：质量验证与自动化测试套件 (Quality Assurance & Validation)**
  - [x] 4.1 运行 `npx vue-tsc --noEmit`，验证 TypeScript 零类型错误
  - [x] 4.2 运行 `npm test -- --run`，验证全部 13 个测试套件 85 个测试用例全部通过
  - [x] 4.3 运行 `go test -v -run TestUIGitAPI ./cmd/...` 验证 Go 后端接口全部通过
  - [x] 4.4 执行 API 实测 (`curl /api/git/status`, `curl /api/git/diff`, `curl /api/git/log`)，验证实时数据流正确性

- [x] 5. **阶段 5：进阶增强与全局联动 (Advanced Features & Full-stack Integration)**
  - [x] 5.1 文件树联动：将 `useGit` 变更文件注入 `FileTreeNode.vue`，实现资源管理器文件树修改高亮（`M` 黄色，`U`/`A` 绿色，`D` 红色）与折叠目录修改数量角标冒泡
  - [x] 5.2 状态栏集成：在 `BottomStatusBar.vue` 底部全局状态栏集成 Git 当前分支名称与 `↓0 ↑1` 同步按钮，支持一键与远程同步及快捷跳转 SCM
  - [x] 5.3 提交历史可视化：在 `SourceControlView.vue` 中集成实时 Git Commit History 列表，展示提交哈希、提交信息、作者与格式化时间戳
  - [ ] 5.4 储藏管理 (Stash)：支持 `git stash` 暂存现场与快速恢复

