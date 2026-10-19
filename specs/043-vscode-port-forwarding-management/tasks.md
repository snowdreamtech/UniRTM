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

## 阶段 6：方式 B 网络层应用协议探针 (Protocol Probing) 实现
- [ ] 6.1 在 `frontend/src/types/ports.ts` 中定义 `PortProbeInfo` 探针数据结构与 `protocol` 协议推断增强
- [ ] 6.2 在 `cmd/64.ui.go` 中实现并发轻量级网络应用协议探针：
  - 严格限制并发超时（<= 40ms），保障系统高吞吐
  - TLS 握手嗅探（发送 TLS ClientHello 探测是否为原生 HTTPS/TLS 服务）
  - HTTP 协议与 Server 标头嗅探（发送 `HEAD /` 提取 `Server`、`X-Powered-By`、`Location`）
  - 常见数据库/中间件握手指纹探测（Redis `PING` / MySQL 握手包）
- [ ] 6.3 扩展后端 API：`GET /api/ports/list?probe=true` 及 `POST /api/ports/probe`
- [ ] 6.4 在 `frontend/vite.config.ts` 中实现同构的本地 Dev 开发中间件探针路由
- [ ] 6.5 在 `frontend/src/utils/portServiceRecognizer.ts` 与 `frontend/src/composables/usePorts.ts` 中联动探针结果，实现自动升级协议为 `https` 及提取精确框架/容器服务名称
- [ ] 6.6 编写单元测试验证探针解析器与协议自适应逻辑

## 阶段 7：方式 C 工作区上下文语义关联推断 (Workspace Context Inference) 实现
- [ ] 7.1 在 `frontend/src/types/ports.ts` 中定义 `WorkspaceServiceHint` 工作区服务推断结构
- [ ] 7.2 在后端 `cmd/64.ui.go` 与前端同构中间件中实现工作区配置解析器（扫描解析 `package.json`、`docker-compose.yml`、`.env` 获得预声明端口和业务服务名）
- [ ] 7.3 提供 `GET /api/ports/workspace-hints` 接口返回工作区上下文预声明服务
- [ ] 7.4 在 `frontend/src/composables/usePorts.ts` 中实现三维多源融合推荐：
  - 融合优先级：`用户手动自定义 > 方式 C 工作区业务名 > 方式 B 探针应用框架 > 方式 A 进程名 > 默认保底`
  - 智能提供“预占位/推荐转发端口”（未启动的服务以空心圆指示）
- [ ] 7.5 编写工作区上下文提取与融合测试用例

## 阶段 8：UI 增强、全 53 国多语言支持与端到端质量验证
- [ ] 8.1 在 `frontend/src/components/ports/PortsPanelView.vue` 中美化呈现：
  - 方式 B 探测出的 HTTPS 锁标、HTTP Server 标签、框架角标
  - 方式 C 关联的工作区文件来源徽标（如 `compose: web`, `pkg: frontend`）
- [ ] 8.2 更新全 53 种语言字典，补齐方式 B/C 新增的多语言文本词条
- [ ] 8.3 运行全量 `npm run test`、`npx vue-tsc --noEmit`、`go vet ./...` 与 `go test ./cmd/...`，确保 0 报错

