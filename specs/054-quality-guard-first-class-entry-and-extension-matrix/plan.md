# 实现计划：Quality Guard 一级入口升级与文件扩展名管线矩阵 (Plan 054)

## 1. 架构总览与技术方案

```
┌─────────────────────────────────────────────────────────────┐
│                       Web UI (Frontend)                     │
│  - ActivityBar: Guard 一级护盾导航图标 (/guard)               │
│  - GuardView: 一级主控制台                                  │
│    ├─ Metric Summary (检查耗时、规则统计、健康评分)          │
│    ├─ Extension Matrix Table (按文件后缀管线表格)             │
│    ├─ Tool Slot Cards (底层槽位详情)                         │
│    ├─ Live Streamed Checklist (实时流式打勾面板)             │
│    └─ TomlConfigModal (查看/编辑 .unirtm.toml 兜底抽屉)       │
│  - ExtensionSlotModal: 可视化增删改查表单弹窗               │
│    └─ 从 Tools 列表中快速选取已安装工具                      │
└──────────────────────────────┬──────────────────────────────┘
                               │ RESTful API
┌──────────────────────────────▼──────────────────────────────┐
│                       Go Backend (cmd / internal)           │
│  - cmd/64.ui_guard.go:                                      │
│    ├─ GET /api/guard/slots                                  │
│    ├─ POST /api/guard/slots (安全写入 .unirtm.toml)         │
│    └─ DELETE /api/guard/slots (安全删除/重置)               │
│  - internal/guard/config.go: 扩展名反查与聚合逻辑           │
│  - internal/config/edit.go: 保证格式完整的 TOML 回写          │
└─────────────────────────────────────────────────────────────┘
```

## 2. 详细技术设计

### 2.1 路由与导航（ActivityBar / Router）
1. 在 `frontend/src/router/index.ts` 中注册路由 `/guard`，关联组件 `views/GuardView.vue`；
2. 在 `frontend/src/components/layout/ActivityBar.vue` 的 `primaryTabs` 中增加：
   ```ts
   {
     id: "guard",
     icon: "workbench:shield",
     label: t("nav.guard") || "Quality Guard",
   }
   ```
3. 在 `SourceControlView.vue` 中保留轻量级跳转横幅或按钮，平滑承接既有习惯。

### 2.2 扩展名管线聚合数据模型（Extension Pipeline Model）
前端通过 `useGuard.ts` 将扁平的 Slots 槽位列表自动聚合成按文件扩展名维度的行数据：
```ts
export interface ExtensionPipelineRow {
  extension: string;          // e.g. ".go", ".ts", ".vue", "*"
  languages: string[];        // e.g. ["Go"], ["TypeScript", "Vue"]
  linter?: ToolSlotItem;      // e.g. "go vet" / "golangci-lint" / "eslint"
  formatter?: ToolSlotItem;   // e.g. "gofmt" / "prettier"
  fixer?: ToolSlotItem;       // e.g. "gofmt -w" / "prettier --write"
  auditor?: ToolSlotItem;     // e.g. "gitleaks"
  allSlots: ToolSlotItem[];
}
```

### 2.3 可视化 CRUD 与兜底机制
1. **新建/编辑弹窗 `ExtensionSlotModal.vue`**：
   - 包含文件后缀标签输入器（支持预设标签如 `.go`, `.ts`, `.sh` 点击快速添加）；
   - “从已安装 Tools 中选取”下拉框（联动 `useTools.ts`）；
   - 完整的命令定制（`command`, `fix_command`, `phase`, `category`, `enabled`）；
2. **兜底 TOML 抽屉 `TomlFallbackModal.vue`**：
   - 提取当前 `.unirtm.toml` 中 `[guard]` 相关文本；
   - 支持只读查看与直接代码编辑，点击“保存”直写配置文件并重载 Guard 状态。

### 2.4 后端持久化方案
利用 Go `github.com/pelletier/go-toml/v2` 解析并结合 `internal/config/edit.go`，将 `guard.slots.<slot-id>` 安全注入项目根目录 `.unirtm.toml`，确保空行和注释不会被意外损毁。
同时，在 Vite 开发服务器 `vite.config.ts` 中编写对应的 mock 路由，保证前端热更新与独立测试均 100% 畅通。
