# 任务拆解：资源管理器 (File Explorer) 对标 VS Code (Tasks)

## 任务分组与执行步骤

- [x] 1. **阶段 1：导航接入与前后端 API 基础设施**
  - [x] 1.1 在 `ActivityBar.vue` 中将 `explorer` 插入至 `dashboard` 与 `tools` 之间，配置 `workbench:explorer` 图标与高亮指示器
  - [x] 1.2 在 `i18n` 语言包中注册 `nav.explorer` 及 `explorer.*` 常用国际化词条
  - [x] 1.3 在 `frontend/src/router/index.ts` 注册 `/explorer` 路由并绑定至 `ExplorerView.vue`
  - [x] 1.4 在 `frontend/vite.config.ts` 中实现 `/api/fs/tree`、`/api/fs/create`、`/api/fs/rename`、`/api/fs/delete` 开发中间件（带工作区防越界校验）
  - [x] 1.5 在 `cmd/64.ui.go` 中实现 Go 原生 `/api/fs/*` 处理函数与安全工作区边界控制
  - [x] 1.6 编写 Go 单元测试验证 `/api/fs/tree` 与安全边界校验并执行第一次提交

- [ ] 2. **阶段 2：VS Code 风格文件树与 Seti UI 图标系统集成**
  - [ ] 2.1 封装 `frontend/src/composables/useFileTree.ts` 管理文件树展开状态、选中状态与 API 交互
  - [ ] 2.2 封装文件图标获取逻辑，深度对接已有的 `@vscode/seti-ui` 图标引擎，精准映射语言与文件类型
  - [ ] 2.3 构建 `FileTreeNode.vue` 递归组件：支持层级缩进引导线、折叠指示箭头、文件夹开关图标、文件类型图标
  - [ ] 2.4 构建 `SidebarPanel.vue` 中的资源管理器侧边栏面板（含工作区标头、新建文件/文件夹、刷新、全部收起 4 大快捷动作）
  - [ ] 2.5 验证文件树展开/折叠、平滑动画与主题适配并执行第二次提交

- [ ] 3. **阶段 3：文件操作、内联编辑与右键上下文菜单**
  - [ ] 3.1 实现原地内联重命名与内联新建文件/文件夹输入框（支持 Enter 提交、Esc 取消、自动聚焦）
  - [ ] 3.2 实现 VS Code 原生风格右键上下文菜单（复制路径、重命名、删除、在终端中打开等）
  - [ ] 3.3 实现删除确认弹窗，防止误删文件
  - [ ] 3.4 实现主内容区 `ExplorerView.vue`，展示当前选中文件详情（类型、大小、修改时间、相对路径）与快捷操作卡片
  - [ ] 3.5 联调右键菜单“在集成终端中打开”，自动在终端切换到目标目录并执行第三次提交

- [ ] 4. **阶段 4：测试套件、构建与文档完整性验证**
  - [ ] 4.1 编写前端组件与 composable 单元测试（测试文件树状态展开、图标匹配、排序）
  - [ ] 4.2 运行前端单元测试 `npm test -- --run`
  - [ ] 4.3 运行 Go 端所有单元测试
  - [ ] 4.4 执行 `npm run build` 确保生产环境打包零错误，并执行最终提交
