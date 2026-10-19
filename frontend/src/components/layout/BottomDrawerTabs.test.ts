import { describe, it, expect, beforeEach } from "vitest";

type DrawerTab = "terminal" | "output" | "tasks" | "doctor" | "ports";

const DRAWER_TAB_ORDER_KEY = "unirtm_drawer_tab_order";
const DRAWER_TAB_HIDDEN_KEY = "unirtm_drawer_tab_hidden";
const DEFAULT_TAB_ORDER: DrawerTab[] = ["terminal", "output", "tasks", "doctor", "ports"];

// 模拟纯逻辑测试与持久化
function getInitialTabOrder(storageMock: Record<string, string>): DrawerTab[] {
  try {
    const saved = storageMock[DRAWER_TAB_ORDER_KEY];
    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed)) {
        const valid = parsed.filter((id): id is DrawerTab => DEFAULT_TAB_ORDER.includes(id as DrawerTab));
        const missing = DEFAULT_TAB_ORDER.filter((id) => !valid.includes(id));
        return [...valid, ...missing];
      }
    }
  } catch {}
  return [...DEFAULT_TAB_ORDER];
}

function getInitialHiddenTabs(storageMock: Record<string, string>): Set<DrawerTab> {
  try {
    const saved = storageMock[DRAWER_TAB_HIDDEN_KEY];
    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed)) {
        return new Set(parsed.filter((id): id is DrawerTab => DEFAULT_TAB_ORDER.includes(id as DrawerTab)));
      }
    }
  } catch {}
  return new Set();
}

function reorderTabs(
  currentOrder: DrawerTab[],
  sourceId: DrawerTab,
  targetId: DrawerTab,
  position: "before" | "after"
): DrawerTab[] {
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

function toggleTabVisibility(
  hiddenTabs: Set<DrawerTab>,
  tabOrder: DrawerTab[],
  tabId: DrawerTab,
  currentTab: DrawerTab
): { newHidden: Set<DrawerTab>; newCurrentTab: DrawerTab } {
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

describe("BottomDrawer Tab Drag, Drop & Visibility Management", () => {
  let storageMock: Record<string, string>;

  beforeEach(() => {
    storageMock = {};
  });

  it("returns default tab order when localStorage is empty", () => {
    const order = getInitialTabOrder(storageMock);
    expect(order).toEqual(["terminal", "output", "tasks", "doctor", "ports"]);
  });

  it("restores custom tab order from localStorage", () => {
    storageMock[DRAWER_TAB_ORDER_KEY] = JSON.stringify(["ports", "terminal", "output", "tasks", "doctor"]);
    const order = getInitialTabOrder(storageMock);
    expect(order).toEqual(["ports", "terminal", "output", "tasks", "doctor"]);
  });

  it("handles missing or corrupted tabs gracefully by appending missing tabs", () => {
    storageMock[DRAWER_TAB_ORDER_KEY] = JSON.stringify(["ports", "terminal", "invalid_tab"]);
    const order = getInitialTabOrder(storageMock);
    expect(order).toEqual(["ports", "terminal", "output", "tasks", "doctor"]);
  });

  it("reorders tabs correctly when dragged before target", () => {
    const initial: DrawerTab[] = ["terminal", "output", "tasks", "doctor", "ports"];
    const result = reorderTabs(initial, "ports", "terminal", "before");
    expect(result).toEqual(["ports", "terminal", "output", "tasks", "doctor"]);
  });

  it("reorders tabs correctly when dragged after target", () => {
    const initial: DrawerTab[] = ["terminal", "output", "tasks", "doctor", "ports"];
    const result = reorderTabs(initial, "terminal", "output", "after");
    expect(result).toEqual(["output", "terminal", "tasks", "doctor", "ports"]);
  });

  it("reordering to itself does not change order", () => {
    const initial: DrawerTab[] = ["terminal", "output", "tasks", "doctor", "ports"];
    const result = reorderTabs(initial, "terminal", "terminal", "after");
    expect(result).toEqual(initial);
  });

  it("supports hiding a tab and auto-switching if current tab is hidden", () => {
    const initialHidden = new Set<DrawerTab>();
    const tabOrder = [...DEFAULT_TAB_ORDER];
    
    // 隐藏当前正在查看的 terminal
    const res = toggleTabVisibility(initialHidden, tabOrder, "terminal", "terminal");
    expect(res.newHidden.has("terminal")).toBe(true);
    // 自动切换到剩余可见的第一个 tab (output)
    expect(res.newCurrentTab).toBe("output");
  });

  it("prevents hiding the last remaining visible tab", () => {
    // 假设已经隐藏了 4 个
    const almostAllHidden = new Set<DrawerTab>(["output", "tasks", "doctor", "ports"]);
    const tabOrder = [...DEFAULT_TAB_ORDER];

    // 尝试隐藏仅剩的 terminal
    const res = toggleTabVisibility(almostAllHidden, tabOrder, "terminal", "terminal");
    // 保护生效，terminal 不被隐藏
    expect(res.newHidden.has("terminal")).toBe(false);
    expect(res.newCurrentTab).toBe("terminal");
  });

  it("supports unhiding a previously hidden tab", () => {
    const hidden = new Set<DrawerTab>(["doctor"]);
    const tabOrder = [...DEFAULT_TAB_ORDER];

    const res = toggleTabVisibility(hidden, tabOrder, "doctor", "terminal");
    expect(res.newHidden.has("doctor")).toBe(false);
  });

  it("supports reset to default tab order and all visible", () => {
    storageMock[DRAWER_TAB_ORDER_KEY] = JSON.stringify(["ports", "doctor", "tasks", "output", "terminal"]);
    storageMock[DRAWER_TAB_HIDDEN_KEY] = JSON.stringify(["tasks", "doctor"]);

    expect(getInitialTabOrder(storageMock)).toEqual(["ports", "doctor", "tasks", "output", "terminal"]);
    expect(getInitialHiddenTabs(storageMock).has("tasks")).toBe(true);

    // 重置
    delete storageMock[DRAWER_TAB_ORDER_KEY];
    delete storageMock[DRAWER_TAB_HIDDEN_KEY];

    expect(getInitialTabOrder(storageMock)).toEqual(DEFAULT_TAB_ORDER);
    expect(getInitialHiddenTabs(storageMock).size).toBe(0);
  });
});
