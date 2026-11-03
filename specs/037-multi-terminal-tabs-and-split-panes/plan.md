# 特性实施计划：多终端标签与窗格拆分 (Plan)

## 1. 架构与状态流设计

### 1.1 数据结构抽象
```typescript
export type SplitDirection = 'none' | 'vertical' | 'horizontal';

export interface TerminalPane {
  id: string;
  name: string;
  shellPath: string;
  lines: TerminalLine[];
  input: string;
  history: string[];
  historyIndex: number;
}

export interface TerminalSession {
  id: string;
  title: string;
  splitDirection: SplitDirection;
  panes: TerminalPane[];
  activePaneId: string;
}
```

### 1.2 组合式函数封装 (`useTerminalSessions.ts`)
- 维护 `sessions: Ref<TerminalSession[]>`、`activeSessionId: Ref<string>`；
- 提供计算属性：
  - `activeSession`: 当前激活的终端标签会话；
  - `activePane`: 当前会话中处于聚焦状态的终端窗格；
- 提供方法：
  - `createSession(shell?: TerminalItem)`: 创建新标签；
  - `splitActiveSession(direction: 'vertical' | 'horizontal', shell?: TerminalItem)`: 拆分为二；
  - `closePane(paneId: string, sessionId?: string)`: 关闭指定窗格；
  - `closeSession(sessionId: string)`: 关闭整个标签会话；
  - `setActiveSession(sessionId: string)`: 切换标签；
  - `setActivePane(paneId: string)`: 切换激活窗格；
  - `runCommand(paneId: string, cmd: string)`: 运行命令并模拟或执行；
  - `clearActivePane()`: 清空当前激活窗格的输出。

## 2. 界面与交互层实现
- **Top Toolbar 操作区**：
  - 在 `BottomDrawer.vue` 顶部右侧新增 VSCode 风格的操作按钮组：
    - `+` 新建终端按扭及下拉选单；
    - 竖直拆分按扭（图标为垂直分割）；
    - 横向拆分按扭（图标为水平分割）；
    - 关闭当前终端垃圾桶按扭；
- **侧边终端列表 (Terminal Tabs Sidebar)**：
  - 在终端视窗右侧渲染紧凑的侧边栏（宽度约 160px，带折叠/自适应支持）；
  - 显示所有标签列表、图标、标题、拆分提示以及悬浮关闭图标；
- **视窗分屏渲染区**：
  - 当 `splitDirection === 'vertical'` 时，主窗体使用 `flex flex-row`，两个窗格并列；
  - 当 `splitDirection === 'horizontal'` 时，主窗体使用 `flex flex-col`，两个窗格上下；
  - 当 `splitDirection === 'none'` 时，单窗格 100% 占满；
  - 窗格具有激活高亮边框（VSCode 经典蓝色/主题色边框聚焦）。

## 3. 分步实施策略
- **Phase 1**: 编写 `useTerminalSessions.ts` 及其单元测试，验证多标签与分屏状态机；
- **Phase 2**: 在 i18n 中补充多终端与分屏操作文案，并同步所有 53 个语言文件；
- **Phase 3**: 在 `BottomDrawer.vue` 中集成工具栏操作按钮（新建、横向拆分、竖直拆分、垃圾桶关闭）；
- **Phase 4**: 实现 VSCode 风格的侧边终端列表及分屏渲染组件与样式；
- **Phase 5**: 全量 Vitest 单测与生产打包回归，确保质量 100%。
