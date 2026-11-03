# 需求规范：Quality Guard 一级入口升级与文件扩展名管线矩阵 (Spec 054)

## 1. 业务背景与用户痛点
在 UniRTM 的工程看门人体系中，已完成从 Python 重型脚本到纯 Go 原生 Guard 引擎的重构，并实现了 ToolSlots 外部工具调度。然而，当前的前端交互仍存在以下痛点：
1. **入口层级不当（寄生于 SCM）**：Quality Guard 目前作为子模块嵌套在“源代码管理（SCM）”中。当开发者工作区无变更时，极少会进入 SCM，导致想要审查代码规范、配置工具管线或执行全仓巡检时难以找到入口；
2. **缺乏文件后缀直观视角（File-centric Gap）**：现有界面仅按工具槽位（ToolSlots）罗列，用户无法一眼看清“我的 `.go` 文件、`.ts` 文件在提交时到底有哪些 Linter、Formatter、Fixer 在守卫”；
3. **配置门槛高（只能手动改文件）**：用户添加新工具或绑定新后缀时，不得不查阅文档去手动修改 `.unirtm.toml`，缺少开箱即用的可视化 CRUD 体验；
4. **底层配置与上层体验脱节**：缺少直接从 UniRTM 已安装工具库（Tools）中挑选后端的交互，缺少多字段定制（具体命令、参数、修复命令），且缺乏把 `.unirtm.toml` 作为高级兜底查看/编辑的操作通道。

---

## 2. 核心功能规范

### 2.1 升级 Quality Guard 为 ActivityBar 一级入口
- 在全局侧边导航条 `ActivityBar` 中新增一级专属图标（科技感护盾 `workbench:shield` / `shield-check`），路由为 `/guard`；
- 独立新建一级主控制台视图 `GuardView.vue`；
- SCM 提交面板与侧边栏保留轻量状态指示胶囊（如 `🛡️ Guard: 38 rules ready`）及深链跳转链接，实现轻量解耦与有机联动。

### 2.2 文件扩展名管线矩阵表格 (Extension Pipeline Matrix Table)
- 以文件扩展名（`.go`, `.ts/.tsx`, `.vue`, `.sh`, `.toml`, `* (全局)`）为行的可视化矩阵大盘；
- 列清晰呈现：
  - **Extension**：文件后缀与语言类型；
  - **Linter**：静态代码检查工具与命令；
  - **Formatter**：格式化工具与命令；
  - **Fixer**：自动修复命令；
  - **Security / Audit**：安全与依赖审计；
  - **Status & Actions**：启停状态切换、✎ 编辑、🗑 删除、↺ 恢复默认。

### 2.3 可视化无代码 CRUD 弹窗与多字段定制
- **新增/编辑弹窗 (`ExtensionSlotModal.vue`)**：
  - 支持从 UniRTM Tools 已安装工具列表中下拉快速挑选（如挑选 `gofmt`, `eslint`, `prettier`, `shfmt`）；
  - 支持多字段精细定制：
    - `ID / Name`：槽位唯一标识；
    - `Extensions`：绑定的文件扩展名标签组（如 `[".ts", ".tsx", ".vue"]`）；
    - `Category`：所属分类（`format`, `lint`, `secret`, `security`, `hygiene`）；
    - `Phase`：执行阶段（Phase 1: Fixer, Phase 2: Formatter, Phase 3: Linter）；
    - `Command`：检查/格式化命令行（支持 `{{files}}` 占位符）；
    - `Fix Command`：自动修复命令行；
    - `Enabled`：启用/禁用状态开关。
- **删除与恢复**：
  - 自定义槽位支持一键删除；
  - 内置槽位支持修改后一键“恢复默认配置”。

### 2.4 保留 `.unirtm.toml` 为兜底操作通道
- 页面右上角保留醒目的 `[ 📄 查看/编辑 .unirtm.toml ]` 兜底操作按钮；
- 点击后可在侧滑抽屉或模态窗口中实时查看当前项目的 Guard 配置段落，支持高级用户直接修改或一键跳转至编辑器，保持单一事实源（Single Source of Truth, SSOT）。

### 2.5 后端 RESTful API 与双向持久化
- `GET /api/guard/slots`：获取当前所有生效的工具槽位定义；
- `POST /api/guard/slots`：创建或更新槽位配置，安全持久化回写至 `.unirtm.toml` 的 `[guard.slots."<slot-id>"]`；
- `DELETE /api/guard/slots?id=<id>`：从 `.unirtm.toml` 中删除指定槽位重写配置。
