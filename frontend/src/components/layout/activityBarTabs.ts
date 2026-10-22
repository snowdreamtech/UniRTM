export type ActivityTabId =
  "dashboard" | "explorer" | "scm" | "guard" | "tools" | "tasks" | "environments" | "marketplace" | "doctor";

export const ACTIVITY_BAR_ORDER_KEY = "unirtm_activity_bar_order";
export const ACTIVITY_BAR_HIDDEN_KEY = "unirtm_activity_bar_hidden";

export const DEFAULT_ACTIVITY_TAB_ORDER: ActivityTabId[] = [
  "dashboard",
  "explorer",
  "scm",
  "guard",
  "tools",
  "tasks",
  "environments",
  "marketplace",
  "doctor",
];

export function getInitialActivityTabOrder(storage?: Storage | Record<string, string>): ActivityTabId[] {
  const s = storage || (typeof localStorage !== "undefined" ? localStorage : null);
  if (!s) return [...DEFAULT_ACTIVITY_TAB_ORDER];
  try {
    const raw =
      typeof (s as Storage).getItem === "function"
        ? (s as Storage).getItem(ACTIVITY_BAR_ORDER_KEY)
        : (s as Record<string, string>)[ACTIVITY_BAR_ORDER_KEY];
    if (raw) {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) {
        const valid = parsed.filter((id): id is ActivityTabId =>
          DEFAULT_ACTIVITY_TAB_ORDER.includes(id as ActivityTabId)
        );
        const missing = DEFAULT_ACTIVITY_TAB_ORDER.filter((id) => !valid.includes(id));
        return [...valid, ...missing];
      }
    }
  } catch {}
  return [...DEFAULT_ACTIVITY_TAB_ORDER];
}

export function getInitialHiddenActivityTabs(storage?: Storage | Record<string, string>): Set<ActivityTabId> {
  const s = storage || (typeof localStorage !== "undefined" ? localStorage : null);
  if (!s) return new Set();
  try {
    const raw =
      typeof (s as Storage).getItem === "function"
        ? (s as Storage).getItem(ACTIVITY_BAR_HIDDEN_KEY)
        : (s as Record<string, string>)[ACTIVITY_BAR_HIDDEN_KEY];
    if (raw) {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) {
        return new Set(
          parsed.filter((id): id is ActivityTabId => DEFAULT_ACTIVITY_TAB_ORDER.includes(id as ActivityTabId))
        );
      }
    }
  } catch {}
  return new Set();
}

export function reorderActivityTabs(
  currentOrder: ActivityTabId[],
  sourceId: ActivityTabId,
  targetId: ActivityTabId,
  position: "before" | "after"
): ActivityTabId[] {
  if (sourceId === targetId) return [...currentOrder];
  const order = [...currentOrder];
  const oldIndex = order.indexOf(sourceId);
  if (oldIndex === -1) return order;
  order.splice(oldIndex, 1);

  const targetIndex = order.indexOf(targetId);
  if (targetIndex !== -1) {
    const insertIndex = position === "after" ? targetIndex + 1 : targetIndex;
    order.splice(insertIndex, 0, sourceId);
  }
  return order;
}

export function toggleActivityTabVisibility(
  hiddenTabs: Set<ActivityTabId>,
  tabOrder: ActivityTabId[],
  tabId: ActivityTabId,
  currentTab: string
): { newHidden: Set<ActivityTabId>; newCurrentTab: string } {
  const newHidden = new Set(hiddenTabs);
  let newCurrent = currentTab;
  const visible = tabOrder.filter((id) => !newHidden.has(id));

  if (newHidden.has(tabId)) {
    newHidden.delete(tabId);
  } else {
    // 保护：至少保留 1 个可见 Tab
    if (visible.length <= 1) {
      return { newHidden, newCurrentTab: newCurrent };
    }
    newHidden.add(tabId);
    if (currentTab === tabId) {
      const remaining = tabOrder.filter((id) => !newHidden.has(id));
      if (remaining.length > 0) {
        newCurrent = remaining[0];
      }
    }
  }

  return { newHidden, newCurrentTab: newCurrent };
}
