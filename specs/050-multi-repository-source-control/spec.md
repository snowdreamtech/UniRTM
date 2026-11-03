# 规范文档：多存储库源代码管理系统 (Specification)

## 1. 业务背景与问题分析 (Problem Statement)

在现代大型工程协作、微服务架构以及 Monorepo / Multi-repo 混合开发中，工作区通常包含多个独立的 Git 代码库（例如一个多根工作区同时挂载了前端、后端、模板以及依赖基础库）。
根据我们在前序特性（`049-multi-root-workspace-explorer`）中实现的多根工作区支持，工作区中可以同时添加多个工程目录（例如 `UniRTM`, `UniBoot`, `UniGo`, `template` 等）。

### 当前 UniRTM 源代码管理（SCM）现状与痛点：
1. **单存储库固定视角**：当前 `useGit` 与 SCM 侧边栏（`SourceControlSidebar.vue`）默认绑定单一根目录执行 Git 操作，无法感知工作区内挂载的其他 Git 仓库；
2. **缺乏多存储库列表展示**：未能如 VS Code 原生设计那样，在 SCM 侧边栏顶部呈现“存储库”（Repositories）管理区域；
3. **缺少存储库切换与独立作用域**：提交更改（Commit）、暂存（Stage）、放弃更改（Discard）、以及 Git Graph 可视化图谱无法自由切换或聚焦在特定存储库上。

---

## 2. 核心原则与设计目标 (Design Goals)

1. **深度对齐 VS Code 原生设计**：
   - 在源代码管理侧边栏顶部呈现“存储库 (Repositories)”列表区域；
   - 每个存储库项目清晰呈现：存储库名称、当前分支名、是否有未提交改动（`*`）、同步状态（`↑x ↓y`）及操作入口；
   - 点击存储库即可切换为当前活动存储库（Active Repository），下方“更改 (Changes)”区域联动展示该存储库的暂存/未暂存文件与提交输入框；
2. **跨存储库 Git Graph 与状态栏深度联动**：
   - 在 Git Graph 页面顶部工具栏提供 `Repo: [ RepoName ▾ ]` 下拉选择器，与侧边栏活动存储库双向同步；
   - 底部状态栏（`BottomStatusBar.vue`）展示当前活动存储库名称及分支同步状态（例如 `template dev* ↑2 ↓0`）；
3. **前后端解耦与多仓扫描**：
   - 后端基于多根工作区配置（`.unirtm/workspaces.json`）与工作区根路径，自动检测与解析包含 `.git` 的有效存储库；
   - 提供 `GET /api/git/repositories` 统一接口；
   - 所有现有 `/api/git/*` 接口（`status`, `stage`, `unstage`, `discard`, `commit`, `diff`, `branches`, `checkout`, `sync`, `graph`, 等）支持 `repo` 参数指定目标存储库；
4. **全语言与高可靠性**：
   - 覆盖 53 种国际化语言包（提供存储库分组文案、多仓切换提示等）；
   - 编写单元测试（Go + Vitest），保证向后兼容性与稳定性。

---

## 3. 详细功能规范 (Functional Requirements)

### 3.1 存储库扫描与接口规范
| 接口 | 方法 | 请求参数 | 响应数据 | 说明 |
| :--- | :--- | :--- | :--- | :--- |
| `/api/git/repositories` | GET | 无 | `{ success: boolean, repositories: GitRepositoryItem[] }` | 扫描工作区所有根目录中的 Git 仓库并返回状态概要 |
| `/api/git/*` (所有 Git 端点) | GET/POST | `repo` (query 参数或 json body) | 原有对应结构 | 根据 `repo`（路径或仓库名）在指定目录下执行 git 命令，缺省回退到当前主工程 |

`GitRepositoryItem` 数据结构：
```typescript
export interface GitRepositoryItem {
  id: string;          // 唯一标识（绝对路径或根名称）
  name: string;        // 显示名称（如 UniRTM, template）
  rootPath: string;    // 本地文件系统绝对/相对路径
  branch: string;      // 当前分支（如 dev, main）
  upstream?: string;   // 上游分支（如 origin/dev）
  ahead: number;       // 领先提交数
  behind: number;      // 落后提交数
  isDirty: boolean;    // 是否有未暂存或未提交改动
  totalChanges: number;// 改动文件总数
}
```

### 3.2 SCM 侧边栏多存储库呈现
1. **存储库列表（Repositories Section）**：
   - 位于 SCM 侧边栏顶部，包含可折叠分组标题“存储库 (Repositories)”；
   - 罗列已识别的存储库，高亮当前活动存储库；
   - 支持快捷刷新与操作菜单；
2. **更改区域（Changes Section）**：
   - 提交信息输入框占位符动态对齐：`消息(按Enter在“<branch>”提交)`；
   - 针对活动存储库执行提交、暂存、撤销与文件差异比较。

### 3.3 Git Graph 与状态栏集成
- Git Graph 顶栏增加 `Repo:` 下拉切换，切换后即时重新加载并渲染对应存储库的 DAG 图谱；
- 底部状态栏展示格式：`<RepoName> <branch><dirty*> <ahead/behind>`。
