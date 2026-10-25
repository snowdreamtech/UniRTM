# 功能规范：Git 多身份隔离与凭证安全保管箱系统 (Git Identity & Credential Security Vault)

## 1. 背景与业务价值 (Background & Value)

在以 AI Agent 自主结对编程、分步原子化执行与代码提交为主导的现代化智能研发平台中，Git 提交（`git commit -S`）与远程同步（`git push` / `git fetch`）是最核心的底层动作。

面对复杂的开发场景（企业内部工作、开源社区协作、个人项目交替），开发者与 AI Agent 频繁面临三大痛点与信任危机：
1. **AI Agent 分步提交频频受阻**：AI 代理运行在后台或无头环境中，若底层 Git 遇到未解锁的 SSH/GPG 私钥且缺少安全预热机制，进程将被卡死或超时报错退出；如果直接在提示词或配置中暴露明文密码，存在极高安全泄露风险。
2. **多身份冲突与全局污染风险**：传统工具往往修改全局 `~/.gitconfig` 或默认使用单一 `~/.ssh/id_rsa`，导致将工作邮箱错误提交到开源仓库，或因密钥不匹配被远程仓库拒绝。
3. **用户对 AI 与工具链的信任危机**：“这款工具会不会扫描并上传我的 `~/.ssh/id_rsa` 私钥？”、“同机其他进程能否刺探我的凭据？”。必须在架构层面实现**彻底透明、物理免疫与安全可证明**。

---

## 2. 核心架构与信任准则 (Guiding Principles)

1. **零私钥接触原则 (Zero Private Key Touch)**：
   - UniRTM 系统（包括前端、后端守护进程、AI Agent）**绝不读取、不解析、不存储任何用户的 SSH 私钥或 GPG 私钥文件**。
   - 所有私钥严格保留在用户系统原生目录（如 `~/.ssh/`、`~/.gnupg/`），权限始终维持操作系统的 `600`。
   - 密钥解密和签名完全由系统原生的 `ssh` 和 `gpg` 二进制执行。UniRTM 仅充当标准的 AskPass / Loopback 通信管道。

2. **零全局污染原则 (Zero Global Pollution)**：
   - 多身份 Profile（包含 `user.name`、`user.email`、`user.signingkey`、`core.sshCommand` 等）的应用，**严格限定在当前打开仓库的局部配置 `.git/config`**（即 `git config --local`）。
   - 严禁擅自修改全局 `~/.gitconfig`，保证多仓库并行时身份互不干扰。

3. **原生系统 Agent 优先 (Native Agent First)**：
   - 自动探测宿主系统的 `ssh-agent`（`$SSH_AUTH_SOCK`）与 `gpg-agent`。
   - 若系统原生 Agent 中已缓存了解锁密钥，底层 Git 操作直接走内核 Agent 瞬时完成，UniRTM 彻底零弹窗、零拦截。

4. **动态 Nonce IPC 免疫防御 (Hardened Nonce IPC)**：
   - UniRTM 派生 Git 进程与本地 AskPass 脚本时，生成不可预测的单次动态 Nonce Token (`UNIRTM_CREDENTIAL_NONCE`)。
   - AskPass 请求后端索取密码时必须携带该 Nonce，且**一次性使用即焚（Single-use & Expired）**。同机其他恶意进程无法伪造请求或截获凭证。

5. **内存受限暂存与全透明审计 (Secure Memory Vault & Audit Trail)**：
   - 若用户勾选会话记住密码，密码仅保存在后端 Go 守护进程受保护的内存映射中，带有严密 TTL（默认 2 小时），进程退出或主动锁定时彻底覆写擦除。
   - 内存审计日志记录每次触发调用的时刻、关联 Profile、密钥类型指纹、调用方 PID 与结果，前端可实时查看，完全透明。

---

## 3. 核心用户故事 (User Stories)

- **US-1 (当前仓库多身份一键切换)**：作为在企业与开源项目兼顾的开发者，我可以在 UniRTM 底部状态栏一键将当前仓库切换为 “Company Profile” 或 “Personal OpenSource”，仓库的 `user.name`、`user.email`、SSH 密钥与 GPG 签名密钥立即局部对齐生效。
- **US-2 (AI Agent 自动化分步无阻提交)**：作为依赖 AI Agent 自动写代码并分步提交的用户，AI Agent 在执行多次原子 commit 时，通过预热或会话级内存凭证缓存，连续提交秒级通过且每一条 commit 都具有合法的已验证签名（Verified GPG Signature）。
- **US-3 (零接触与系统 Agent 静默复用)**：作为已经启动了系统原生 `ssh-agent` / `gpg-agent` 的极客开发者，UniRTM 自动检测到系统 Agent，commit 和 push 完全静默通过，用户确信 UniRTM 没有复制或私自嗅探任何私钥。
- **US-4 (AskPass 动态 Nonce 防攻击)**：作为注重安全的企业安全员，同机其他未授权进程即便尝试向 UniRTM 本地 API 发送凭证提取请求，因缺少单次动态 Nonce Token 被立即拒绝并记录在安全审计日志中。
- **US-5 (透明安全审计跟踪)**：作为用户，我可以在安全设置面板中点击“凭证活动审计日志”，清晰查看到最近所有 Git 凭据解锁记录与调用细节，心安理得，彻底建立对工具链的信任。

---

## 4. 功能需求矩阵 (Functional Requirements)

### 4.1 Git 身份配置与局部隔离 (FR-1: Identity Management)
- **FR-1.1**: 支持定义与管理多个 Git Profile（Profile ID、Profile Name、User Name、User Email、SSH Key 关联、GPG Key ID、是否默认启用签名等）。
- **FR-1.2**: 支持一键将选中的 Profile 应用于当前工作区仓库（执行 `git config --local`），设置 `user.name`, `user.email`, `user.signingkey`, `commit.gpgsign`, 以及隔离的 `core.sshCommand = "ssh -i <key_path> -o IdentitiesOnly=yes"`。
- **FR-1.3**: 提供扫描读取当前系统可用公钥（`~/.ssh/*.pub`）与 GPG 公钥列表（`gpg --list-public-keys`）的只读只探测接口，便于用户在界面上下拉选择，绝不触碰私钥。

### 4.2 系统 Agent 探测与 AskPass 动态防御 (FR-2: Agent & Hardened IPC)
- **FR-2.1**: 后端启动与 Git 运行时自动探测 `$SSH_AUTH_SOCK` 和运行中的 `gpg-agent` 状态，并在状态栏/设置面板呈现诊断信号灯（Active / Inactive）。
- **FR-2.2**: 派生 Git 进程时生成密码学安全的 32 字节 UUIDv4/CryptoRand 单次 Nonce，注入子进程环境变量 `UNIRTM_CREDENTIAL_NONCE`。
- **FR-2.3**: AskPass 脚本在通过 IPC 向 UniRTM 请求密码时必须验证 Nonce，成功消费后立即从未决 Nonce 池中销毁。

### 4.3 内存凭据保管箱与审计追踪 (FR-3: Vault & Audit Trail)
- **FR-3.1**: 内存 Session Keyring 支持会话级暂存（仅限用户显式勾选 remember），支持 TTL 到期销毁、主动全部清空、单项锁止。
- **FR-3.2**: 记录 `CredentialAuditRecord` 流水（时间戳、类型 SSH/GPG、操作 commit/push/fetch、密钥指纹/标识、状态 Success/Denied/Cancelled、调用方信息），前端可视化呈现。

### 4.4 前端界面与状态栏交互 (FR-4: UI & UX)
- **FR-4.1**: 底部状态栏集成 `GitProfileStatus.vue`，展示当前激活 Profile，点击可弹出快速选择列表。
- **FR-4.2**: 设置页面提供 Git Identity & Security 独立配置卡片，支持 Profile CRUD、Agent 诊断指示、审计日志抽屉。
- **FR-4.3**: 凭据弹窗 `PassphraseModal.vue` 增强 Nonce 与安全认证徽章，清晰告知用户当前由 UniRTM 局部受限 Broker 代理。

---

## 5. 验收标准 (Acceptance Criteria)

1. 在前端切换 Git Profile 后，当前仓库的 `.git/config` 立即更新对应配置，而全局 `~/.gitconfig` 保持原样、无任何改动。
2. 配置独立 SSH 密钥时，`core.sshCommand` 自动注入 `-o IdentitiesOnly=yes`，确保严格使用该密钥连接，避免密钥污染。
3. 若系统环境变量存在有效 `SSH_AUTH_SOCK`，Git Push 不触发 UniRTM 密码弹窗，自动透传系统 Agent 成功完成。
4. AskPass 端点核验 Nonce：若直接通过 curl 或恶意无 Nonce 请求 `/api/git/credential-askpass`，返回 403 Forbidden；合法 Nonce 使用一次后再次使用返回 403。
5. 前端安全审计面板能够准确列出刚刚发生的凭证调用流水，包括时间、操作与密钥标识。
6. 全量通过 Go 单元测试 (`go test ./...`) 与前端测试 (`npm run test`)，32 项质量门禁全部绿灯。
