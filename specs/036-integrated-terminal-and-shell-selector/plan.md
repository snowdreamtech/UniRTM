# 技术方案与实施计划：集成终端与多Shell检测选择器 (Plan)

## 1. 架构设计与技术方案

```
+------------------------------------------------------------------------------+
|                              UniRTM Web Workbench                            |
+------------------------------------------------------------------------------+
| [SettingsView.vue]                           [BottomDrawer.vue]              |
|  - 终端设置卡片                               - 新增 [终端] Tab              |
|  - 扫描结果列表                               - 终端选择下拉框 [Bash ▼]      |
|  - 点击单选持久化                             - 清屏 / 终止 / 运行状态展示   |
+------------------------------------------------------------------------------+
                                      │
                         REST API     │  (fetch /api/terminals, /api/config)
                                      ▼
+------------------------------------------------------------------------------+
|                     Backend (Go cmd/64.ui.go & vite.config.ts)               |
|                                                                              |
|  1. GET /api/terminals:                                                      |
|     - Linux/macOS: 读取 /etc/shells + 扫描 exec.LookPath (bash/zsh/fish...)  |
|     - Windows: 扫描 PowerShell / CMD / Git Bash                              |
|     - 匹配 os.Getenv("SHELL") / COMSPEC 确定系统默认                         |
|     - 读取用户 settings.terminal 确定当前生效终端                            |
|                                                                              |
|  2. POST /api/config:                                                        |
|     - 接收 settings.terminal 持久化至 ~/.config/unirtm/ 或用户配置           |
+------------------------------------------------------------------------------+
```

---

## 2. 核心模块与实现细节

### 2.1 后端终端扫描模块 (Go & Vite Dev Server)
1. **探测逻辑**：
   - 优先读取 `/etc/shells`（如果存在），提取出有效的 Shell 路径；
   - 补充常见 Shell 二进制名称：`bash`、`zsh`、`sh`、`fish`、`pwsh`、`dash`、`tcsh`、`tmux`，在 `$PATH` 中执行查找；
   - Windows 平台下查找 `powershell.exe`、`pwsh.exe`、`cmd.exe`、`bash.exe`（Git Bash 安装路径）；
   - 对查找到的路径进行去重，并解析其友好展示名称（如 `/bin/bash` -> `Bash`，`/bin/zsh` -> `Zsh`）；
   - 判断 `is_default`：与 `os.Getenv("SHELL")` 或 `COMSPEC` 路径对比。
2. **持久化与 Schema**：
   - 更新 `schemas/unirtm.schema.json` 中的 `settings.properties.terminal`；
   - 用户配置保存后即时生效。

### 2.2 前端状态与组合式函数 (`useTerminal.ts`)
- 封装 `useTerminal` composable：
  - `terminals`: 扫描到的可用终端列表；
  - `defaultTerminal`: 系统默认终端对象；
  - `currentTerminal`: 当前选中的终端对象；
  - `isLoading`: 扫描状态；
  - `fetchTerminals()`: 请求后端接口；
  - `setTerminal(path: string)`: 切换并持久化保存。

### 2.3 底栏抽屉增强 (`BottomDrawer.vue`)
- 新增 `terminal` Tab 选项卡（图标采用 `workbench:terminal`，支持徽章提示）；
- 在抽屉顶部操作栏左侧/中间放置终端快速切换器下拉菜单：
  - 显示当前终端名称；
  - 点击弹出下拉列表，列出所有可用终端；
  - 标记 `(系统默认)` 和当前选中的 Checkmark；
- 顶部右侧提供清屏、刷新会话、跳转设置快捷键；
- 主区域渲染终端视窗容器（终端 prompt、会话信息、快速命令输入与执行结果展示）。

### 2.4 设置视图 (`SettingsView.vue`)
- 在「常规设置」或新设「集成终端」分类中新增可视化终端配置卡片：
  - 展示当前系统支持的全部 Shell；
  - 点击单选按钮即可即时切换默认终端；
  - 提供重新检测（Rescan）按钮。

---

## 3. 风险评估与应对策略

| 潜在风险 | 影响程度 | 应对策略 |
| :--- | :--- | :--- |
| **容器或受限环境无 `/etc/shells`** | 低 | 兜底直接使用 `sh` 与 `bash`，确保任何环境均有可用终端返回 |
| **配置的终端被意外删除** | 低 | 前端与后端进行存在性校验，若配置的路径不存在则自动回退至系统默认终端 |
| **测试与跨环境构建** | 低 | 遵循 Vitest 单元测试全覆盖，不引入重型不可编译依赖 |
