# 技术实现方案：Monaco Editor 异步离线集成 (Plan)

## 1. 技术栈与架构设计

- **核心依赖**：`monaco-editor`（官方 npm 包，完整能力不剔除任何内置语言）。
- **打包与离线策略**：
  - Vite `rollupOptions.output.manualChunks` 独立拆分 `vendor-monaco`，确保首屏体积零膨胀；
  - 本地离线 Web Worker 构建，使用原生 ESM 动态模块 `new Worker(new URL('monaco-editor/esm/vs/editor/editor.worker.js', import.meta.url), { type: 'module' })`，100% 离线运行，绝对禁止引入网络 CDN。
- **状态层**：封装 `useEditorTabs.ts` 集中管理多文件标签页模型与缓存。
- **视图层**：
  - `MonacoEditor.vue`：Monaco 容器组件，负责模型绑定、生命周期、自适应布局 (`ResizeObserver`) 与主题监听；
  - `ExplorerView.vue`：升级为多 Tab 工作区容器，协调侧边栏文件树点击与编辑器联动。

---

## 2. 核心模块与实现细节

### 2.1 本地离线 Worker 配置与动态按需加载
在 `frontend/src/services/editor/monacoLoader.ts` 中配置全局 `MonacoEnvironment`：
```ts
import * as monaco from "monaco-editor";

// 配置本地离线 Worker 构建
self.MonacoEnvironment = {
  getWorker(_, label) {
    if (label === "json") {
      return new Worker(new URL("monaco-editor/esm/vs/language/json/json.worker.js", import.meta.url), { type: "module" });
    }
    if (label === "css" || label === "scss" || label === "less") {
      return new Worker(new URL("monaco-editor/esm/vs/language/css/css.worker.js", import.meta.url), { type: "module" });
    }
    if (label === "html" || label === "handlebars" || label === "razor") {
      return new Worker(new URL("monaco-editor/esm/vs/language/html/html.worker.js", import.meta.url), { type: "module" });
    }
    if (label === "typescript" || label === "javascript") {
      return new Worker(new URL("monaco-editor/esm/vs/language/typescript/ts.worker.js", import.meta.url), { type: "module" });
    }
    return new Worker(new URL("monaco-editor/esm/vs/editor/editor.worker.js", import.meta.url), { type: "module" });
  },
};
```
当首次挂载 `MonacoEditor.vue` 时通过异步 `await import(...)` 触发加载，并在加载期展示平滑加载态。

### 2.2 多标签页状态管理 (`useEditorTabs.ts`)
```ts
export interface EditorTab {
  id: string; // 文件相对路径
  path: string;
  name: string;
  language: string;
  content: string;
  originalContent: string;
  isDirty: boolean;
  model?: monaco.editor.ITextModel;
}
```
- 每次打开新文件时，若已存在相同 `id` 则直接切换至该 Tab；若不存在则自动拉取 `/api/fs/read` 并创建新 Tab；
- 为每个 Tab 创建并复用 `ITextModel`，用户在多 Tab 间来回切换时**完全保留撤销/重做历史 (Undo/Redo Stack) 与光标位置**；
- 监听内容变化对比 `originalContent` 计算 `isDirty`。

### 2.3 主题同步适配
- 监听当前 UniRTM 主题系统：
  - 默认暗黑主题应用 `vs-dark`；
  - 浅色主题应用 `vs`；
  - 基于当前 CSS 变量自动向 Monaco 注册定制主题（如匹配 `--bg-color`、`--card-bg`、`--text-main`），确保编辑器边框与底色与外层工作区严丝合缝。

### 2.4 快捷键与生命周期管理
- 绑定 `monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS` 触发快捷保存；
- 使用 `ResizeObserver` 监听父容器尺寸变化（如侧边栏拖拽、底部抽屉拖拽），自动调用 `editor.layout()`。
