# 任务清单：Git SSH 与 GPG 私钥密码中继与交互管理系统 (Tasks)

## 阶段 1：数据类型与前端凭据状态管理 (Core Types & Composable)
- [x] 1.1 在 `frontend/src/types/credential.ts` 中定义 `CredentialRequest`、`CredentialResponse` 与 `CredentialType` 类型
- [x] 1.2 在 `frontend/src/composables/useCredentialBroker.ts` 中实现凭据协调状态机、轮询与应答函数
- [x] 1.3 编写 `frontend/src/composables/useCredentialBroker.test.ts` 单元测试，验证请求暂存、响应与会话缓存逻辑
- [x] 1.4 **执行提交 1**：`feat(credential): implement credential broker types and composable with tests`

## 阶段 2：VS Code 质感密码解锁模态框 (PassphraseModal.vue UI)
- [x] 2.1 创建 `frontend/src/components/common/PassphraseModal.vue` 弹窗组件（支持 SSH/GPG 区分展示、密码小眼睛显隐、Caps Lock 检测）
- [x] 2.2 增加会话记住复选框 `[x] 在当前工作区会话期间记住密码` 与 `Enter`/`Esc` 快捷键支持
- [x] 2.3 在 `frontend/src/App.vue` 全局挂载 `PassphraseModal`，打通凭据弹窗弹出与关闭链路
- [x] 2.4 **执行提交 2**：`feat(ui): add modern PassphraseModal component with session keyring option`

## 阶段 3：Go 后端凭据协调器与内存安全缓存 (Backend Broker Core)
- [x] 3.1 在后端 `cmd/64.ui.go` 与 `cmd/64.ui_credential.go` 中实现线程安全的内存 Keyring 缓存 `sessionKeyring`（支持 TTL 超时清理）
- [x] 3.2 实现凭据事务挂起等待队列 `credentialCoordinator`（通过 Go channel 协调底层等待与前端应答）
- [x] 3.3 新增 `/api/git/credential-pending`、`/api/git/credential-respond` 与 `/api/git/credential-clear-cache` REST API
- [x] 3.4 编写 Go 端单元测试验证超时保底机制与 channel 唤醒
- [x] 3.5 **执行提交 3**：`feat(backend): implement credential coordinator and in-memory session keyring`

## 阶段 4：Git/SSH/GPG 命令执行链路拦截与注入 (Execution Integration)
- [ ] 4.1 在 `cmd/64.ui.go` 的 `runGitCommand` 中封装统一的 Broker 执行器，自动注入 `GIT_ASKPASS` 与 `SSH_ASKPASS`
- [ ] 4.2 为 `git commit -S` 注入 `--pinentry-mode loopback` 并在底层捕获 GPG passphrase 请求并中继
- [ ] 4.3 结合 `useGit.ts`，在前端触发 Commit / Push / Pull 时无缝挂起并唤醒凭据弹窗
- [ ] 4.4 **执行提交 4**：`feat(git): integrate askpass and gpg loopback broker into git command execution`

## 阶段 5：全 53 国多语言支持与端到端质量验证 (i18n & Verification)
- [ ] 5.1 更新 `frontend/src/i18n/types.ts`、`zh-CN.ts` 与 `en-US.ts`，添加 SSH/GPG 凭据相关文案
- [ ] 5.2 自动同步词条到全 53 国语言字典，确保 `i18n.test.ts` 0 报错
- [ ] 5.3 运行全量 `npm run test` 与 `npm run build`，确保类型检查与生产构建 100% 成功
- [ ] 5.4 **执行提交 5**：`feat(i18n): sync multi-language keys and finalize credential broker validation`
