# 功能规范：Git SSH 与 GPG 私钥密码中继与交互管理系统 (Git SSH & GPG Passphrase Broker)

## 1. 背景与目标 (Background & Goals)

在现代化 Web 研发环境（UniRTM Web UI、云原生 DevContainer、Remote SSH/WSL、以及 AI 自动化编程执行中），Git 提交（`git commit -S`）与远程同步（`git push` / `git fetch` / `git pull`）是开发者与自动化助理最频繁执行的操作。

然而，当开发者的 **SSH 私钥** 或 **GPG 签名私钥** 设置了保护密码（Passphrase）时，现有的 VS Code、无头环境（Headless）以及 AI 智能助手往往频繁遭遇“拦路虎”：
1. **SSH 终端卡死或拒绝访问**：底层 `ssh` 强制读取 `/dev/tty`，在后台执行、Web 进程或 AI Agent 环境中因无 TTY 会直接报错 `Permission denied (publickey)` 或死等挂起；
2. **GPG 签名失败崩溃**：`commit.gpgsign=true` 时，底层通过 `gpg-agent` 启动系统级 `pinentry`（如 `pinentry-curses`/`pinentry-gtk`），在缺少 TTY 和 `$DISPLAY` 图形桌面的环境下直接崩溃抛出 `Inappropriate ioctl for device`、`gpg: signing failed: Inappropriate ioctl for device`，VS Code 原生无法拦截；
3. **缺少会话级凭据缓存**：即使输入了一次密码，后续每次 commit 或 push 都会频繁弹窗，严重阻断心流。

### 核心目标
构建 UniRTM 内置的**端到端通用凭据中继经纪人（Universal Passphrase Broker）**：
1. **全协议覆盖**：100% 解决 Git 场景下的 **SSH 私钥解锁** 与 **GPG 签名私钥解锁**；
2. **Web 优雅交互**：在 Web 界面呈现高科技感、VS Code 风格的密码输入弹窗（`PassphrasePromptModal.vue`），支持密码显隐、快捷键回车确认/Esc取消、错误实时重试；
3. **免密会话安全缓存（Session Keyring Cache）**：支持用户勾选“在此会话期间记住密码”，在 UniRTM 后端内存中安全缓存指定 TTL，后续自动免密通过；
4. **终端兜底无缝集成**：可一键转到底栏集成的 XTerm 终端完成原生交互；
5. **AI 编程友好**：AI 助手在后台执行 Git 操作若遇密码，可触发 Web 浮层提示用户输入，防止 AI 进程挂死超时。

---

## 2. 核心用户故事 (User Stories)

- **US-1 (SSH 私钥 Web 弹窗解锁)**：作为开发者，在 UniRTM 的 SCM 面板点击“同步更改 (Sync / Push)”时，如果我的 SSH 密钥设置了 Passphrase，Web 界面立即弹出友好输入框，输入密码后 push 顺利成功，无需打开终端手动执行。
- **US-2 (GPG 签名透明穿透)**：作为开发者，当我配置了 `git commit -S` 强制签名提交时，在 Web 界面点击“提交 (Commit)”时，系统通过 Loopback 机制捕获 GPG 密码请求，在 Web 界面弹出 GPG 专用解锁框（显示 Key ID 和 User ID），输入后顺利生成已验证签名（Verified Commit）。
- **US-3 (会话级免密记忆)**：作为开发者，我在第一次输入 Passphrase 时勾选了“在本次会话期间记住此密码”，当天后续的所有 commit 和 push 操作均不再弹窗，秒级静默完成。
- **US-4 (智能重试与错误提示)**：作为开发者，如果我不小心输错了密码，系统不发生进程崩溃，而是立即在弹窗内以红色醒目提示“密码错误，请重新输入”，并允许再次尝试。
- **US-5 (终端控制台原生兜底)**：作为偏好终端的极客开发者，在弹窗中可以点击“在终端中输入”，操作将自动桥接到 UniRTM 底部的集成 PTY 终端中完成。

---

## 3. 功能需求矩阵 (Functional Requirements)

### 3.1 凭证请求与拦截核心 (FR-1)
- **FR-1.1 (SSH AskPass 桥接)**：通过注入 `GIT_ASKPASS`、`SSH_ASKPASS` 环境变量，引导底层 ssh 凭证请求进入 UniRTM Broker 管道。
- **FR-1.2 (GPG Loopback 穿透)**：对带签名提交注入 `--pinentry-mode loopback` 并在底层管道建立交互，拦截 `GET_PASSPHRASE` 并转发至前端。
- **FR-1.3 (全双工事件通信)**：通过 WebSocket / Server-Sent Events (SSE) 或挂起等待长连接通道，实现后端与 Web 前端的毫秒级双向凭证索取与应答。

### 3.2 交互界面与视觉系统 (FR-2)
- **FR-2.1 (凭证对话框 PassphraseModal)**：
  - 显示密钥类型标牌（SSH Key 或 GPG Key）、密钥路径/指纹、请求提示说明；
  - 密码输入框（支持小眼睛显隐密码、Caps Lock 大写锁定检测）；
  - `[x] 在本次工作区会话期间记住密码` 复选框；
  - 快捷键支持：`Enter` 提交、`Esc` 取消。
- **FR-2.2 (错误与重试机制)**：捕获密码验证失败反馈，弹窗保持开启并高亮输入框，提示重新输入。

### 3.3 安全性与缓存机制 (FR-3)
- **FR-3.1 (仅内存隔离存储)**：记住的密码绝不写入磁盘明文文件，仅在守护进程受保护的内存中维护。
- **FR-3.2 (会话 TTL 与主动清除)**：支持设定会话过期时间（默认 2 小时），或在用户登出 / 关闭 UniRTM 服务时彻底清空。

---

## 4. 验收标准 (Acceptance Criteria)

1. 配置有密码的 SSH 密钥执行 `git push` 时，Web 界面正常弹出 Passphrase 对话框，正确输入后成功推送。
2. 配置有密码的 GPG 密钥执行 `git commit -S` 时，不出现 `Inappropriate ioctl for device` 报错，Web 界面正确弹出 GPG 解锁框并成功完成签名提交。
3. 勾选“会话记住密码”后，下一次相同密钥的 Git 操作直接免密通过。
4. 全量自动化测试（Vitest + Go test）覆盖核心逻辑，所有测试通过。
