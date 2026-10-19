# 架构与实现方案设计：多根工作区与文件浏览器管理系统 (Plan)

## 1. 架构目标与职责分离 (Architecture Goals)

1. **多根数据模型设计**：
   - 抽象 `WorkspaceFolder` 结构：
     ```typescript
     export interface WorkspaceFolder {
       id: string;
       name: string;
       path: string;
       isRoot: boolean;
     }
     ```
   - 在 `useFileTree` 中从单一 `tree` 扩展为支持 `workspaceRoots: Ref<FSNode[]>` 集合；
   - 单根时 `workspaceRoots` 长度为 1，多根时为 N；
2. **前后端接口对齐**：
   - `GET /api/workspace/roots`：获取当前多根配置列表；
   - `POST /api/workspace/roots/add`：添加新文件夹根目录；
   - `POST /api/workspace/roots/remove`：移出指定的根文件夹；
   - 在开发环境 `vite.config.ts` 中同步提供完整的 API Mock 与状态维护，在 Go 后端提供标准处理函数。
3. **右键上下文菜单（`FileContextMenu.vue`）升级**：
   - 新增“将文件夹添加到工作区...”项；
   - 新增“将文件夹从工作区删除”项，通过 `isWorkspaceRoot(node)` 智能判断可见性；
4. **添加文件夹模态弹窗（`AddFolderToWorkspaceModal.vue`）**：
   - 支持直接输入路径或浏览当前目录；
   - 附带合法性校验与去重判断。

---

## 2. 关键流程设计 (Key Flows)

### 2.1 添加文件夹至工作区
```
[用户触发] -> (右键某子文件夹 或 顶栏操作)
          -> 弹出 "将文件夹添加到工作区" 确认/输入弹窗
          -> POST /api/workspace/roots/add { path: "/path/to/folder" }
          -> 后端写入工作区配置并返回成功
          -> 前端并发拉取对应根目录的 FSNode 并挂载至 workspaceRoots
          -> 界面刷新展示多根树视图，Toast 提示添加成功
```

### 2.2 从工作区删除文件夹
```
[用户触发] -> (右键某顶级工作区根文件夹)
          -> 点击 "将文件夹从工作区删除"
          -> 弹出原生或拟态确认框: "确定要从工作区中移除文件夹 [xxx] 吗？此操作不会删除磁盘上的任何文件。"
          -> POST /api/workspace/roots/remove { path: "/path/to/folder" }
          -> 从 workspaceRoots 过滤移除该根
          -> 若移除后仅剩 1 个根，自动切回单根展示模式
```
