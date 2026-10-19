# 任务清单：VS Code 对齐端口转发与智能云端转发系统 (Tasks)

## 阶段 1：后端系统级端口探测器与 CRUD API 实现
- [x] 1.1 在 `cmd/64.ui.go` 中实现跨平台系统监听端口探测器（支持 Linux `/proc/net/tcp`、`ss`，返回端口、协议、PID 与命令行）
- [x] 1.2 在 `cmd/64.ui.go` 中实现云端智能反向代理网关 `/api/ports/preview/:port/*`（支持 HTTP / WebSocket 穿透）
- [x] 1.3 在 `cmd/64.ui.go` 中实现完整的端口管理 REST API 接口（`/api/ports/list`, `/api/ports/forward`, `/api/ports/update`, `/api/ports/stop`, `/api/ports/unforward`）
- [x] 1.4 在 `frontend/vite.config.ts` 中实现同构的开发环境中间件，确保本地 dev 与生产打包表现一致
- [x] 1.5 编写 Go 单元测试 `cmd/64.ui_ports_test.go`，验证 API 正确性与安全路径隔离

## 阶段 2：前端状态管理 Composable 与智能服务识别引擎
- [x] 2.1 创建 `frontend/src/types/ports.ts`，定义严格的 TypeScript 接口与状态枚举
- [x] 2.2 创建 `frontend/src/utils/portServiceRecognizer.ts`，实现智能服务指纹识别器（端口、进程名与技术栈图标关联）
- [x] 2.3 创建 `frontend/src/composables/usePorts.ts`，实现自动发现轮询、用户持久化端口维护、端口冲突自动重映射、CRUD 操作
- [x] 2.4 在 `usePorts.test.ts` 中编写完整的状态管理单元测试（覆盖自动发现、手动 CRUD、冲突解决与重映射）

## 阶段 3：底部抽屉「端口」UI 面板开发（1:1 对齐 VS Code）
- [x] 3.1 创建 `frontend/src/components/ports/PortsPanelView.vue`，实现 VS Code 原生设计风格的端口表格：
  - 状态指示灯（实心绿灯 `●` / 空心圆 `○`）
  - 端口号与标签名（支持内联双击重命名）
  - 转发地址（`localhost:port`，悬停操作：复制链接、在内置/外部浏览器打开）
  - 正在运行的进程（命令行、PID、悬停 Tooltip）
  - 来源标识（「自动转发」与「用户转发」）
  - 访问控制切换（私有 / 公开）
  - 单行操作按钮（停止转发 / 启动转发、删除端口）
  - 底部操作栏「添加端口」按钮与内联添加行
- [x] 3.2 在 `frontend/src/components/layout/BottomDrawer.vue` 中挂载「端口 (Ports)」Tab，展示活跃端口数量角标 Badge
- [x] 3.3 在 `frontend/src/components/common/TechIcon.vue` 中增加 `workbench:ports` 图标支持

## 阶段 4：终端 URL 智能嗅探与高级/云端特性集成
- [x] 4.1 在 `frontend/src/components/terminal/XTerminalPane.vue` 中集成终端输出 URL 嗅探器，检测到本地服务启动时自动注册到端口列表并触发 Toast 提示
- [x] 4.2 实现云端安全预览 URL 生成与扫码/一键打开体验（支持反向代理路径与 Token）
- [x] 4.3 实现端口右键上下文菜单（复制地址、复制端口、停止转发、删除端口、切换可见性）

## 阶段 5：全 53 种语言国际化与全量验证
- [x] 5.1 在全 53 种语言字典（`frontend/src/i18n/locales/*.ts`）中注入端口模块多语言词条
- [x] 5.2 运行 `npm run test` 与 `npx vue-tsc --noEmit`，确保 0 类型报错与全测试通过
- [x] 5.3 运行 `go vet ./...` 与 `go test ./cmd/...` 确保后端稳健
