# 架构设计与实施计划：Git 多身份隔离与凭证安全保管箱系统 (Plan)

## 1. 架构概览 (Architecture Overview)

```
+---------------------------------------------------------------------------------------+
|                                    UniRTM Frontend                                    |
|  [Status Bar: Active Profile]   [Settings: Git Identity & Vault]   [Passphrase Modal] |
|                                |                                          ^           |
+--------------------------------|------------------------------------------|-----------+
                                 | REST / SSE                               |
+--------------------------------v------------------------------------------v-----------+
|                                  UniRTM Go Backend                                    |
|                                                                                       |
|  +------------------------+  +-------------------------+  +------------------------+  |
|  |   Git Profile Manager  |  |  Native Agent Detector  |  |  Secure Memory Keyring |  |
|  | (Local .git/config)    |  |  (ssh-agent, gpg-agent) |  |  (TTL & Secure Zero)   |  |
|  +------------------------+  +-------------------------+  +------------------------+  |
|               |                                                        |              |
|               v                                                        v              |
|  +---------------------------------------------------------------------------------+  |
|  |             Hardened Credential Coordinator (One-time Nonce IPC)                |  |
|  |     Injects UNIRTM_CREDENTIAL_NONCE, verifies token, logs to Audit Trail        |  |
|  +---------------------------------------------------------------------------------+  |
+---------------------------------------------------------------------------------------+
        |                                                        |
        | Spawns with AskPass / Loopback                         | Localhost IPC
        v                                                        v
+-------------------------------+              +-------------------------------------+
| git commit -S / git push/fetch| ------------>|  unirtm-askpass / gpg-loopback shim |
+-------------------------------+              +-------------------------------------+
        |
        +-----> Calls OS Native ssh / gpg binaries (Zero Private Key Touch)
```

---

## 2. 数据结构设计 (Data Contracts)

### 2.1 Git Profile 实体模型
```go
type GitProfile struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    UserName    string `json:"userName"`
    UserEmail   string `json:"userEmail"`
    SSHKeyPath  string `json:"sshKeyPath,omitempty"` // 指向公钥或已知 key 别名，仅用于 core.sshCommand 构造
    SigningKey  string `json:"signingKey,omitempty"` // GPG Key ID / Fingerprint
    GPGSign     bool   `json:"gpgSign"`              // 是否默认对 commit 强制签名
    IsDefault   bool   `json:"isDefault"`
    UpdatedAt   int64  `json:"updatedAt"`
}
```

### 2.2 安全审计流水记录 (CredentialAuditRecord)
```go
type CredentialAuditRecord struct {
    ID        string `json:"id"`
    Timestamp int64  `json:"timestamp"`
    Action    string `json:"action"`    // "commit_sign", "ssh_push", "ssh_fetch"
    Type      string `json:"type"`      // "ssh", "gpg"
    KeyHint   string `json:"keyHint"`   // 公钥路径截断或 GPG KeyID 后 8 位，脱敏展示
    Result    string `json:"result"`    // "success", "cached", "denied", "agent_pass"
    CallerPID int    `json:"callerPid"` // 触发进程 PID
}
```

### 2.3 系统原生 Agent 状态 (NativeAgentStatus)
```go
type NativeAgentStatus struct {
    SSHAgentActive bool   `json:"sshAgentActive"`
    SSHAgentSock   string `json:"sshAgentSock,omitempty"` // 脱敏展示
    GPGAgentActive bool   `json:"gpgAgentActive"`
}
```

---

## 3. 后端 API 接口设计 (API Specification)

1. `GET /api/git/profiles`
   - 获取用户配置的所有 Git 身份 Profiles 列表，以及当前已激活应用的 Profile。
2. `POST /api/git/profiles`
   - 保存或更新 Git 身份 Profile。
3. `DELETE /api/git/profiles/{id}`
   - 删除指定的 Git Profile。
4. `POST /api/git/profile/apply`
   - 将选中的 Profile 应用到当前打开的工作区仓库。
   - 底层原子化执行：
     - `git config --local user.name "<userName>"`
     - `git config --local user.email "<userEmail>"`
     - 若有 `signingKey`: `git config --local user.signingkey "<signingKey>"`, `git config --local commit.gpgsign true`
     - 若有 `sshKeyPath`: `git config --local core.sshCommand "ssh -i <sshKeyPath> -o IdentitiesOnly=yes"`
5. `GET /api/git/system-keys`
   - 只读枚举系统原生公钥文件（`~/.ssh/*.pub`）和已导入的 GPG 公钥（`gpg --list-public-keys`），辅助前端下拉选择，彻底零触碰私钥。
6. `GET /api/git/agent-status`
   - 查询当前系统原生 `ssh-agent` 与 `gpg-agent` 探测连通状态。
7. `GET /api/git/credential-audit`
   - 获取最近凭证调用与拦截审计流水（保留最近 100 条）。
8. `POST /api/git/credential-askpass` (带有 Nonce 强校验)
   - 供 AskPass 脚本调用的本地特权 IPC。必须在 Header `X-Unirtm-Nonce` 包含派生时生成的未过期一次性 Token，校验成功后立焚。

---

## 4. 前端组件与交互架构 (Frontend Architecture)

1. **`frontend/src/composables/useGitIdentity.ts`**:
   - 维护 Profile 列表、当前激活 Profile、系统 Keys 发现、系统 Agent 状态、切换操作。
2. **`frontend/src/components/common/GitProfileStatus.vue`**:
   - 挂载于底部状态栏（紧邻分支名与 Git 同步状态），清晰显示当前身份徽标。
   - 点击展开 Profile QuickPick，一键切换。
3. **`frontend/src/views/SettingsView.vue` (或独立卡片)**:
   - Git 身份管理：Profile CRUD 表单、绑定 SSH/GPG 公钥。
   - 安全防护中心：原生 Agent 探针状态卡片、一键清除内存 Keyring 缓存按钮。
   - 凭据活动审计抽屉：实时可查的日志流水表格。
4. **多语言词条同步**:
   - 在 `frontend/src/i18n/types.ts`、`zh-CN.ts`、`en-US.ts` 等增加对应词条，并同步支持 53 种语言。

---

## 5. 安全与质量门禁策略 (Security & Quality Gates)

1. **测试策略**:
   - Go 单元测试覆盖：Profile 增删改查、局部 `.git/config` 隔离写入验证、Nonce 校验及单次使用即焚测试、Agent 探测测试。
   - 前端 Vitest 单元测试覆盖：`useGitIdentity.test.ts`、状态栏切换与 Mock API 交互。
2. **UniRTM Quality Guard**:
   - 运行项目全量 32 项自动化检测门禁，确保无任何未捕获错误或断裂。
