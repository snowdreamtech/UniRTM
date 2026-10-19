# 技术方案与架构设计：自定义工具前缀选择与固定版本号录入 (Plan)

## 1. 架构与模块设计

### 1.1 前缀与工具 Spec 解析器工具函数 (`frontend/src/utils/toolSpecParser.ts`)
实现统一的纯函数解析与拼装逻辑：
- `parseToolSpec(raw: string): { prefix: string; name: string; version: string }`
  - 支持解析：
    - `cargo:ripgrep@14.1.0` -> `{ prefix: "cargo:", name: "ripgrep", version: "14.1.0" }`
    - `npm:@biomejs/biome@1.9.4` -> `{ prefix: "npm:", name: "@biomejs/biome", version: "1.9.4" }`
    - `zig@0.13.0` -> `{ prefix: "", name: "zig", version: "0.13.0" }`
    - `python` -> `{ prefix: "", name: "python", version: "" }`
- `composeToolSpec(prefix: string, name: string, version: string): { toolKey: string; version: string; fullSpec: string }`
  - 生成 `.unirtm.toml` 标准键名（如 `npm:@biomejs/biome`）与版本号（如 `1.9.4`）。
- `SUPPORTED_TOOL_PREFIXES`: 常用前缀元数据清单（包含 label, value, icon, placeholder）。

### 1.2 复合分段输入栏组件/模板重构 (`frontend/src/views/ToolsView.vue`)
- 将原有的 `custom-add-section` 升级为分段一体式控件组：
  - `<select v-model="customPrefix" class="custom-prefix-select">`
  - `<input v-model="customToolName" class="custom-name-input" @paste="handlePasteSpec">`
  - `<span class="version-at-symbol">@</span>`
  - `<input v-model="customToolVersion" class="custom-version-input" placeholder="版本号, 如: 1.0.0">`
  - `<button class="tool-action-btn primary" :disabled="!isCustomFormValid" @click="addCustomRuntime">添加</button>`
- 实时配置预览：
  `<div v-if="customPreviewKey" class="custom-add-preview">📝 配置预览: <code>{{ customPreviewKey }} = "{{ customToolVersion }}"</code></div>`

### 1.3 提交逻辑改造
- 废弃 `latest` 默认逻辑；
- 调用 `quickSwitch(toolKey, customToolVersion, isGlobal)` 并更新前端 `runtimes` 响应式列表。

---

## 2. 任务分期
- **Phase 0**: 规范、方案设计与基线提交
- **Phase 1**: 工具函数 `toolSpecParser.ts` 及其单元测试 `toolSpecParser.test.ts`
- **Phase 2**: `ToolsView.vue` 模板与脚本升级（三段式输入、粘贴解析、语法预览、样式微调）
- **Phase 3**: 53 种语言包国际化补充与全量测试回归验证
