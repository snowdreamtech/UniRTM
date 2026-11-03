# 架构与实现方案设计：多存储库源代码管理系统 (Plan)

## 1. 架构目标与职责分离 (Architecture Goals)

1. **多仓扫描与后端适配层**：
   - 依赖工作区多根配置（`.unirtm/workspaces.json`）与当前工作目录，遍历定位包含 `.git` 目录的工程；
   - 封装仓库路径解析助手：`resolveRepoDir(repo string) string`，支持根据 repo 名称或路径解析安全物理目录；
   - 暴露 `/api/git/repositories`，支持各端点注入 `repo` 参数；
2. **前端多仓状态管理 (`useGit.ts`)**：
   - 新增 `repositories: Ref<GitRepositoryItem[]>` 与 `activeRepo: Ref<GitRepositoryItem | null>`；
   - 提供 `setActiveRepo(repo: GitRepositoryItem | string)` 与 `fetchRepositories()`；
   - 所有现有操作（`fetchStatus`, `stage`, `commit`, `fetchGraph`, 等）自动携带 `activeRepo.value?.rootPath`；
3. **SCM 侧边栏重构 (`SourceControlSidebar.vue`)**：
   - 引入“存储库”卡片列表，展示仓库名称、分支、脏状态徽标与上下游同步数；
   - 点击切换活动仓库，触发 SCM 变更列表与输入框状态重载；
4. **Git Graph 与底部状态栏联动 (`SourceControlView.vue` & `BottomStatusBar.vue`)**：
   - Git Graph 增加存储库选择下拉框；
   - 状态栏适配多仓显示。

---

## 2. 关键流程设计 (Key Flows)

### 2.1 存储库发现与初始化
```
[工作区加载] -> useGit.fetchRepositories()
             -> GET /api/git/repositories
             -> 后端遍历 workspaceRoots + 当前工作目录
             -> 针对每个包含 .git 的目录执行 git status 快速摘要
             -> 返回 repositories 数组
             -> 若未选定 activeRepo，默认选中第一项或与当前工作区匹配的仓库
```

### 2.2 切换活动存储库与操作隔离
```
[用户在侧边栏点击仓库 B] -> useGit.setActiveRepo(B)
                      -> 更新 activeRepo
                      -> 重新拉取 B 的 status 与 branches
                      -> 下方“更改”区域刷新为 B 的更改文件列表
                      -> 提交按钮与快捷键仅作用于仓库 B
```
