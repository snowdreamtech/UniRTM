# 需求规范：集成终端与多Shell检测选择器 (Integrated Terminal & Shell Selector)

## 1. 概述与背景
在 UniRTM 工作台中，用户执行任务（如 `unirtm run dev`、构建脚本、测试套件）或排查环境工具版本（如 `node -v`、`go version`）时，当前底栏仅提供单向静态日志流，无法支持 ANSI 彩色高亮、实时键盘交互（stdin 响应 `[y/n]`、密码输入）、信号控制（`Ctrl+C` 终止进程）及自由命令执行。

本特性旨在为 UniRTM 打造紧贴 VS Code 体验的**集成终端（Integrated Terminal）**系统：
1. **自动检测客户端所有可用 Shell/终端**：全自动扫描系统环境已安装的 Shell，展示详细名称与路径，并精准标记系统默认终端。
2. **终端偏好设置与持久化**：允许用户在界面上随心指定工作台默认使用的终端，并持久化存储于 UniRTM 的全局配置中（`settings.terminal`）。
3. **底栏工作台「终端」Tab 深度集成**：在底栏四大面板（终端、任务、输出、诊断）中加入终端视图，支持便捷下拉切 Shell、清屏、终止进程及任务联动运行。

---

## 2. 用户场景与体验目标

### 2.1 场景 1：自动检测与智能识别系统默认 Shell
- **现状**：缺乏终端概念，用户无法获知本机已安装的 Shell 环境。
- **目标**：进入终端或设置页面时，自动扫描宿主环境（Linux/macOS 下扫描 `/etc/shells` 与 `$PATH`；Windows 下扫描 PowerShell/CMD/Git-Bash 等），按优先级列出所有可用终端；自动读取系统环境变量 `$SHELL`（或 Windows `COMSPEC`），将系统默认项打上 `(系统默认)` 徽章并作为缺省选择。

### 2.2 场景 2：底栏终端视图切换与会话控制
- **现状**：底栏仅有输出、任务、诊断 3 个 Tab。
- **目标**：在 `BottomDrawer` 中引入 **[ 终端 (Terminal) ]** Tab。
  - 顶部工具栏提供 Shell 下拉切换菜单（如 `[ Bash (系统默认) ▼ ]`、`[ Zsh ]`、`[ Fish ]`）；
  - 提供清屏（Clear）、终止当前运行任务（Kill / SIGINT）、进入终端设置的快捷按钮；
  - 面板内渲染高质量终端输出或交互界面，完美贴合深浅色主题。

### 2.3 场景 3：终端偏好持久化（设置中心与快捷切换双向同步）
- **现状**：无终端持久化配置项。
- **目标**：
  - 用户在底栏下拉框切换终端时，可一键设为默认终端；
  - 在 `SettingsView` 设置视图中提供专属的「终端设置 (Terminal Settings)」卡片，罗列扫描到的全部终端、执行路径与系统状态，用户选择后通过 `/api/config` 写入持久化存储（`settings.terminal`），重启或跨页面访问不丢失。

---

## 3. 功能需求详细规范

### FR-01: 客户端终端扫描与检测 API (`GET /api/terminals`)
- **实现范围**：
  - 在 Go 后端（`cmd/64.ui.go`）与前端本地模拟层（`vite.config.ts`）提供标准端点 `GET /api/terminals`；
  - 返回数据结构：
    ```json
    {
      "default": "/bin/bash",
      "current": "/bin/bash",
      "terminals": [
        { "id": "bash", "name": "Bash", "path": "/bin/bash", "is_default": true },
        { "id": "zsh", "name": "Zsh", "path": "/usr/bin/zsh", "is_default": false },
        { "id": "sh", "name": "POSIX sh", "path": "/bin/sh", "is_default": false },
        { "id": "fish", "name": "Fish", "path": "/usr/bin/fish", "is_default": false },
        { "id": "pwsh", "name": "PowerShell", "path": "/usr/bin/pwsh", "is_default": false }
      ]
    }
    ```
  - 支持跨平台检测（Linux、macOS、Windows）。

### FR-02: 终端偏好持久化存储 (`POST /api/config` & `schemas/unirtm.schema.json`)
- 在 `schemas/unirtm.schema.json` 中的 `settings` 声明 `terminal` 字段（`type: string`，描述默认终端路径或 ID）；
- 用户选定终端后，调用 `/api/config` 将 `settings.terminal` 持久化写入用户级配置文件；
- 在获取配置与终端列表时，优先取 `settings.terminal` 作为 `current`，若未配置则回落到 `default`。

### FR-03: 底栏面板集成「终端」Tab (`BottomDrawer.vue`)
- 在 `BottomDrawer.vue` 标签栏加入 `<button class="drawer-tab" ...>`（图标：`workbench:terminal`，标签：`终端` / `Terminal`）；
- 抽屉头部右侧/中间加入终端切换下拉菜单，实时展示当前生效的终端名称；
- 底部内容区渲染集成终端主界面（支持 ANSI 格式化、终端输入提示、命令行操作历史与状态提示）；
- 支持全屏放大（Maximize）、清屏（Clear）、关闭（Close）。

### FR-04: 设置中心「终端设置」卡片 (`SettingsView.vue`)
- 在偏好设置中增加「集成终端」设置项；
- 以可视化卡片/列表形式展示本机探测出的终端环境（包含终端名称、绝对路径、是否为系统默认、是否为当前选中）；
- 允许用户单选切换，实时保存并提示 Toast 反馈。

---

## 4. 非功能需求与验收标准
1. **可靠的系统容错**：若扫描到的自定义终端被意外卸载，系统能平滑回退至系统默认 Shell，不发生崩溃或白屏；
2. **多语言国际化**：所有新增文案（终端、系统默认、检测中、切换终端、终端设置等）同步覆盖中英文及多语言字典；
3. **单元测试与构建**：所有 Vitest 测试用例保持 100% 通过，`npm run build` 零编译错误；
4. **代码规范与提交**：遵循 Git 提交规范并强制执行 GPG 签名。
