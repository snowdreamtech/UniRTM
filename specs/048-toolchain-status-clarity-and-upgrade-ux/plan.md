# 架构与实现方案设计：工具链状态清晰化与低噪音升级系统 (Plan)

## 1. 架构目标与职责分离 (Architecture Goals)

1. **状态计算解耦**：
   - 提取通用的状态计算器 `getToolDisplayState(rt, currentScope)`；
   - 统一由卡片视图与列表视图复用，确保状态判断规则单一真实源（Single Source of Truth）；
2. **样式解耦与语义归位**：
   - 将现有 `.runtime-version-badge.is-active` 的绿色硬编码重构成中性底色；
   - 新增专用于升级指示的 `.runtime-upgrade-badge`（专属高亮绿 `--success`）；
   - 新增专用于缺失安装的 `.runtime-missing-badge`（琥珀警告色 `--warning`）；
3. **按钮自适应上下文**：
   - 遇到 `missing` 状态时，快速主按钮自动绑定直接安装缺失版本。

---

## 2. 技术设计细节 (Technical Details)

### 2.1 状态判断逻辑
```typescript
export type ToolDisplayState = "ready" | "upgradable" | "missing" | "unconfigured";

export interface ToolStateResult {
  state: ToolDisplayState;
  configuredVersion: string;
  upgradableVersion: string | null;
  isInstalled: boolean;
}

export function computeToolState(rt: RuntimeItem, scope: "user" | "workspace"): ToolStateResult {
  const configured = scope === "user" ? (rt.global_version || "") : (rt.active_version || rt.global_version || "");
  if (!configured) {
    return { state: "unconfigured", configuredVersion: "", upgradableVersion: null, isInstalled: false };
  }

  const isInstalled = isAlreadyInstalled(rt.id, configured);
  const upgradable = getRemoteUpgradableVersion(rt.id);

  if (!isInstalled) {
    return { state: "missing", configuredVersion: configured, upgradableVersion: upgradable, isInstalled: false };
  }

  if (upgradable) {
    return { state: "upgradable", configuredVersion: configured, upgradableVersion: upgradable, isInstalled: true };
  }

  return { state: "ready", configuredVersion: configured, upgradableVersion: null, isInstalled: true };
}
```

### 2.2 视觉与 CSS Token 规范
- **就绪态 (`.runtime-version-badge.is-ready`)**：
  - 深色主题：`background: rgba(255, 255, 255, 0.08); color: #f1f5f9; border: 1px solid rgba(255, 255, 255, 0.15);`
  - 浅色主题：`background: #f8fafc; color: #0f172a; border: 1px solid #cbd5e1;`
  - **不展示任何小图标**，保持纯净。
- **升级态 (`.runtime-upgrade-badge`)**：
  - 专属高亮绿：`background: rgba(16, 185, 129, 0.15); color: #10b981; border: 1px solid rgba(16, 185, 129, 0.4);`
  - 包含动态箭头或可升级文案，一眼可辨。
- **缺失态 (`.runtime-version-badge.is-missing`)**：
  - 琥珀警告色：`background: rgba(245, 158, 11, 0.15); color: #f59e0b; border: 1px solid rgba(245, 158, 11, 0.4);`
  - 显示 `未下载` 或 `待安装`。

---

## 3. 测试与验证策略 (Testing & Validation)

1. **单元测试**：针对 `computeToolState` 或工具状态工具函数编写完备的单测用例覆盖 4 种状态；
2. **多语言测试**：`i18n.test.ts` 确保 53 种语言包含所有状态相关键值；
3. **视觉与构建校验**：运行 `npm test` 与 `npm run build`，确保类型完备、0 错误。
