# 任务清单：Git 多身份隔离与凭证安全保管箱系统 (Tasks)

## 阶段 0：SpecKit 规范定义与准备 (Spec & Blueprint)
- [x] 0.1 编写 `specs/056-git-identity-and-credential-security-vault/spec.md` 需求规范文档
- [x] 0.2 编写 `specs/056-git-identity-and-credential-security-vault/plan.md` 技术架构方案
- [x] 0.3 编写 `specs/056-git-identity-and-credential-security-vault/tasks.md` 分步执行任务清单
- [x] 0.4 **执行提交 0**：`docs(spec): add specification and tasks for 056-git-identity-and-credential-security-vault`

---

## 阶段 1：后端 Git Profile 隔离管理与持久化 (Backend Identity Manager)
- [x] 1.1 在 `internal/server/` 中定义 `GitProfile`、`GitProfileManager` 数据模型与存储层（持久化在工作区受保护配置或应用私有配置）
- [x] 1.2 实现 `ApplyProfile(workspaceDir, profileID)`：仅通过 `git config --local` 修改局部 `.git/config`，严格杜绝全局污染
- [x] 1.3 实现公钥与 GPG 列表只读发现机制（仅扫描 `~/.ssh/*.pub` 与 `gpg --list-public-keys`，绝不触碰私钥）
- [x] 1.4 实现 HTTP 路由：`GET/POST/DELETE /api/git/profiles`，`POST /api/git/profile/apply`，`GET /api/git/system-keys`
- [x] 1.5 编写 `internal/server/git_profile_test.go` 单元测试，验证 Profile 保存与局部仓库配置写入隔离性
- [x] 1.6 **执行提交 1**：`feat(server): add git identity profile management with local repository isolation`

---

## 阶段 2：原生 Agent 探测与 Hardened AskPass Nonce IPC (Agent & Hardened Broker)
- [x] 2.1 实现系统原生 `ssh-agent`（`$SSH_AUTH_SOCK`）与 `gpg-agent` 探测器，支持通过 `GET /api/git/agent-status` 返回健康状态
- [x] 2.2 在凭证协调器中引入密码学安全的单次 Nonce 机制：Git 进程派生时注入 `UNIRTM_CREDENTIAL_NONCE`
- [x] 2.3 在 AskPass 校验逻辑中严格核验 Nonce，一次性使用即焚，防范同机未授权进程刺探
- [x] 2.4 实现安全审计流水记录器 `CredentialAuditTrail`，记录凭据调用事件，暴露 `GET /api/git/credential-audit`
- [x] 2.5 编写 Go 单元测试覆盖 Nonce 鉴权、即焚防御与审计流水
- [x] 2.6 **执行提交 2**：`feat(server): enforce one-time nonce token in askpass ipc and audit trail`

---

## 阶段 3：前端 Git Identity 状态管理与类型 (Frontend Composable & Types)
- [x] 3.1 在 `frontend/src/types/git.ts`（或 `credential.ts`）中定义 `GitProfile`、`SystemKeys`、`NativeAgentStatus`、`CredentialAuditRecord` 接口
- [x] 3.2 创建 `frontend/src/composables/useGitIdentity.ts` 组合式函数，管理 Profiles、切换、Agent 状态与审计流水
- [x] 3.3 编写 `frontend/src/composables/useGitIdentity.test.ts` 单元测试
- [x] 3.4 **执行提交 3**：`feat(frontend): implement useGitIdentity composable and types with tests`

---

## 阶段 4：状态栏 Profile 切换器与安全设置面板 (UI Components)
- [ ] 4.1 创建 `frontend/src/components/common/GitProfileStatus.vue` 状态栏组件，展示当前活跃 Profile 并支持 QuickPick 切换
- [ ] 4.2 在 `frontend/src/components/layout/StatusBar.vue` 集成 `GitProfileStatus` 组件
- [ ] 4.3 在设置中心或独立模块中增加 Git Identity 与凭证安全卡片（Profile 表单、Agent 指示灯、内存缓存清空、安全审计日志抽屉）
- [ ] 4.4 编写或更新前端 UI 单元测试
- [ ] 4.5 **执行提交 4**：`feat(ui): add GitProfileStatus status bar switcher and identity settings`

---

## 阶段 5：全 53 国多语言同步与全量质量门禁 (i18n & Quality Gate)
- [ ] 5.1 更新 `frontend/src/i18n/types.ts`、`zh-CN.ts` 与 `en-US.ts`，定义多身份与安全保管箱相关词条
- [ ] 5.2 自动同步词条至全部 53 种语言文件，确保 `i18n.test.ts` 0 报错
- [ ] 5.3 运行全量 `npm run test` 与 `go test ./internal/server`，运行构建并验证 UniRTM 质量门禁
- [ ] 5.4 **执行提交 5**：`feat(i18n): sync multi-language keys and finalize git identity security vault`
