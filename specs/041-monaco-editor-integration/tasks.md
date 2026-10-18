# 任务拆解：Monaco Editor 代码编辑器集成 (Tasks)

## 任务分组与执行步骤

- [x] 1. **阶段 1：依赖安装、Vite 离线配置与分包策略**
  - [x] 1.1 安装 `monaco-editor` 官方依赖包并更新 `package.json`
  - [x] 1.2 在 `frontend/vite.config.ts` 中配置 `manualChunks` 拆分 `vendor-monaco`，确保首屏体积零膨胀
  - [x] 1.3 创建 `frontend/src/services/editor/monacoLoader.ts`，配置本地 ESM Worker 离线加载环境与全套语言推导映射
  - [x] 1.4 验证 Monaco 离线加载与 TypeScript 类型并执行第一次提交

- [x] 2. **阶段 2：多标签页状态管理 (`useEditorTabs.ts`)**
  - [x] 2.1 封装 `frontend/src/composables/useEditorTabs.ts`，管理标签页集合（打开、激活切换、关闭、关闭其他、关闭全部）
  - [x] 2.2 实现文件读取缓存、双向内容绑定与 `isDirty` 未保存状态计算
  - [x] 2.3 与 `/api/fs/read` 及 `/api/fs/write` 接口打通，支持 `Ctrl+S` 单文件保存与批量保存
  - [x] 2.4 编写 `frontend/src/composables/useEditorTabs.test.ts` 单元测试并执行第二次提交

- [ ] 3. **阶段 3：Monaco 编辑器组件与主画布深度整合**
  - [ ] 3.1 封装 `frontend/src/components/editor/MonacoEditor.vue`，实现按需异步挂载、`ITextModel` 缓存复用、`Ctrl+S` 绑定、`ResizeObserver` 动态布局与主题自适应
  - [ ] 3.2 升级 `frontend/src/views/ExplorerView.vue`，集成 VS Code 风格多 Tab 头部栏、未保存指示小圆点、关闭快捷按钮与快速切换
  - [ ] 3.3 保持非文本文件兼容性（支持图片预览、大文件安全提示与文件夹概览）
  - [ ] 3.4 增强编辑器底部状态栏（光标行号列号、UTF-8 编码、LF 换行符与语言模式标签）
  - [ ] 3.5 联调左侧资源管理器点击文件无缝开 Tab 并执行第三次提交

- [ ] 4. **阶段 4：测试套件、构建与完整性验证**
  - [ ] 4.1 运行前端 Vitest 单元测试 `npm test -- --run`，确保全部测试通过
  - [ ] 4.2 运行 Go 后端测试 `go test -v ./cmd -run TestUIFSApi`
  - [ ] 4.3 执行 `npm run build` 确保生产环境打包成功且零报错
  - [ ] 4.4 检查所有提交日志与文档完整性并交付
