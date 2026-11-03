# 架构设计与实施计划：集成终端会话持久化、动态状态感知与高级体验系统 (Plan)

## 1. 架构总览与技术设计

本系统通过在前端终端管理层构建**会话持久化**、**动态进程嗅探器**与**后台长任务监听器**三层增强体系，实现媲美专业本地 IDE 的丝滑韧性体验。

```mermaid
graph TD
    subgraph Terminal Core
        XP[XTerminalPane.vue]
        XT[@xterm/xterm]
        WGL[@xterm/addon-webgl]
    end

    subgraph State & Persistence Layer
        TP[useTerminalPersistence.ts<br/>Ring Buffer & Storage]
        PW[useTerminalProcessWatcher.ts<br/>Process & Title Sniffer]
        TN[useTerminalNotifier.ts<br/>Long-Running Task Toast]
    end

    subgraph UI & Workbench
        BDT[BottomDrawerTabs.vue<br/>Dynamic Title & Badges]
        TCM[TerminalContextMenu.vue<br/>Export Log & Clear Buffer]
        TOAST[useToast.ts]
    end

    XP --> XT
    XP --> WGL
    XP <--> TP
    XP --> PW
    PW --> TN
    TN --> TOAST

    PW --> BDT
    XP --> TCM
    TCM --> TP
```

---

## 2. 模块细化设计

### 2.1 会话历史持久化缓存器 (`useTerminalPersistence.ts`)
- **存储设计**：
  - 每个 `paneId` 分配独立的固定容量 Ring Buffer（默认保留最新 500KB / 5,000 行纯文本与 ANSI 转义流）；
  - 数据在写入 xterm 的同时追加进入 Ring Buffer，持久化到 `sessionStorage` 的分段存储槽中；
  - 提供 `exportTerminalLog(paneId, title)`：清洗 ANSI 转义控制字符或保留原始格式，生成带 ISO 时间戳的 `.log` 文件并通过 `Blob` 触发下载；
  - 提供 `clearTerminalHistory(paneId)`：清空缓存，并与终端 `term.clear()` 联动。

### 2.2 前台子进程与技术栈嗅探器 (`useTerminalProcessWatcher.ts`)
- **识别机制**：
  - 监听终端输出字符流与用户输入命中；
  - 支持特征正则匹配表：
    - `Vite / Node`: `vite`, `npm run`, `pnpm dev`, `next dev`, `node` -> 图标：`codicon-symbol-property`，高亮绿
    - `Python`: `python`, `pytest`, `uvicorn`, `flask` -> 图标：`codicon-symbol-variable`，高亮蓝
    - `Go`: `go run`, `go test`, `air` -> 图标：`codicon-symbol-structure`，高亮青
    - `Rust`: `cargo run`, `cargo build`, `cargo test` -> 图标：`codicon-gear`，高亮橙
    - `Docker / Cloud`: `docker`, `kubectl` -> 图标：`codicon-server`，高亮蓝
    - `Git`: `git push`, `git fetch`, `git clone` -> 图标：`codicon-git-commit`，高亮紫
  - 状态输出：
    - `detectedProcess`: 当前正在前台执行的指令名（若空闲则回退显示默认 Shell）；
    - `startTime`: 命令启动时间戳；
    - `status`: `'running' | 'idle' | 'failed'`。

### 2.3 后台长任务完成智能通知器 (`useTerminalNotifier.ts`)
- **逻辑流**：
  - 当检测到进程结束信号（如返回提示符或输出退出信息），计算耗时 `durationMs = Date.now() - startTime`；
  - 若 `durationMs >= 10,000ms`（大于 10 秒）且当前终端处于后台（非活跃 Tab 或抽屉折叠），调用 `showToast` 弹出通知：
    - 成功：`[终端 1] "npm run build" 执行完成，耗时 42s`
    - 失败：`[终端 1] "cargo test" 执行失败 (异常退出)，耗时 15s`
  - 点击通知气泡可直接拉起并聚焦到对应的终端 Tab。

### 2.4 Tab 动态呈现与自定义重命名 (`BottomDrawerTabs.vue` 升级)
- 终端标签页实时展示当前进程徽标与名字；
- 支持双击 Tab 标题进入可编辑状态，自定义设置别名（如“编译服务”、“本地数据库”等），并在会话中持久化。

### 2.5 WebGL 加速与连字渲染
- 在 `XTerminalPane.vue` 中尝试加载 `@xterm/addon-webgl`；
- 注册 `onContextLoss` 监听器，在不支持 WebGL 或 GPU 崩溃时平滑回退至标准 Canvas 渲染器，保证 100% 稳定性；
- 配置 `fontFamily` 优先级支持主流编程连字字体。

---

## 3. 分阶段实施策略

- **Phase 0：Spec Kit 规划与基线建立**
  - 创建规范、计划与任务拆解清单，提交文档基线。
- **Phase 1：终端会话持久化与日志导出功能**
  - 实现 `useTerminalPersistence.ts`；
  - `XTerminalPane.vue` 接入历史恢复机制；
  - `TerminalContextMenu.vue` 新增「导出终端日志」与「清空历史缓存」；
  - 编写单元测试并通过验证，执行分步 Git 提交。
- **Phase 2：动态子进程感知与 Tab 智能重命名**
  - 实现 `useTerminalProcessWatcher.ts` 进程匹配引擎；
  - `BottomDrawerTabs.vue` 接入进程感知与双击重命名；
  - 编写单元测试并通过验证，执行分步 Git 提交。
- **Phase 3：后台长任务监控通知与 WebGL/连字视觉增强**
  - 实现 `useTerminalNotifier.ts`；
  - 在 `XTerminalPane.vue` 中接入 WebGL 硬件加速与后台通知；
  - 编写单元测试并通过验证，执行分步 Git 提交。
- **Phase 4：53 种多语言适配与全量构建发布**
  - 同步新增字典至 53 种语言文件；
  - 全量运行 `npm run test` 与 `npm run build`，封箱提交。
