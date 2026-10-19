# 架构与实现方案设计：开发环境工具链状态折叠分组面板 (Plan)

## 1. 架构目标与职责分离 (Architecture Goals)

1. **状态分类引擎适配 (`toolStatus.ts` & `useToolUpdates.ts`)**：
   - 复用现有的 `computeToolStatus(tool, scope, updates)` 纯函数算法；
   - 在 `SidebarPanel.vue` 中获取 `rawTools` 及更新候选列表 `toolUpdates`，为每个工具项注入计算后的 `statusState: "missing" | "upgradable" | "ready"`；
2. **多组别聚合与响应式派生**：
   - 派生计算属性：
     - `missingTools`: 过滤出 `statusState === "missing"` 且命中搜索关键字的工具集合；
     - `upgradableTools`: 过滤出 `statusState === "upgradable"` 且命中搜索关键字的工具集合；
     - `readyTools`: 过滤出 `statusState === "ready"` 且命中搜索关键字的工具集合；
3. **UI 折叠面板（Accordion）组件渲染 (`SidebarPanel.vue`)**：
   - 针对 `missingTools`（若长度 > 0）：渲染“⚠️ 待安装 (N)”分组；
   - 针对 `upgradableTools`（若长度 > 0）：渲染“🟢 可升级 (N)”分组；
   - 针对 `readyTools`：渲染“已就绪 / 已安装 (N)”分组；
   - 各自拥有独立折叠状态变量：`isMissingExpanded`, `isUpgradableExpanded`, `isReadyExpanded`；
4. **视觉层次与主题适配**：
   - 待安装组：采用琥珀色警告徽标与警示高光；
   - 可升级组：采用绿色升级圆点与高亮；
   - 已就绪组：精致极简的既有风格。

---

## 2. 关键流程设计 (Key Flows)

### 2.1 状态计算与分组渲染流程
```
[加载 / 切换作用域] -> 拉取 /api/tools/status 及 /api/tools/updates
                  -> computeToolStatus(tool, scope, updates)
                  -> 派生 missingTools / upgradableTools / readyTools
                  -> UI 响应式挂载对应折叠面板
                  -> 用户在搜索框输入时，各分组内即时过滤
```
