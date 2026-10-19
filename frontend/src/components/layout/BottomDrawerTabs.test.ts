import { describe, it, expect, beforeEach } from "vitest";

type DrawerTab = "terminal" | "output" | "tasks" | "doctor" | "ports";

const DRAWER_TAB_ORDER_KEY = "unirtm_drawer_tab_order";
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

describe("BottomDrawer Tab Drag and Drop Reordering", () => {
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
    // 假设用户只存了 ports 和 terminal，应该自动补齐 output, tasks, doctor
    storageMock[DRAWER_TAB_ORDER_KEY] = JSON.stringify(["ports", "terminal", "invalid_tab"]);
    const order = getInitialTabOrder(storageMock);
    expect(order).toEqual(["ports", "terminal", "output", "tasks", "doctor"]);
  });

  it("reorders tabs correctly when dragged before target", () => {
    const initial: DrawerTab[] = ["terminal", "output", "tasks", "doctor", "ports"];
    // 将 'ports' 拖拽到 'terminal' 之前
    const result = reorderTabs(initial, "ports", "terminal", "before");
    expect(result).toEqual(["ports", "terminal", "output", "tasks", "doctor"]);
  });

  it("reorders tabs correctly when dragged after target", () => {
    const initial: DrawerTab[] = ["terminal", "output", "tasks", "doctor", "ports"];
    // 将 'terminal' 拖拽到 'output' 之后
    const result = reorderTabs(initial, "terminal", "output", "after");
    expect(result).toEqual(["output", "terminal", "tasks", "doctor", "ports"]);
  });

  it("reordering to itself does not change order", () => {
    const initial: DrawerTab[] = ["terminal", "output", "tasks", "doctor", "ports"];
    const result = reorderTabs(initial, "terminal", "terminal", "after");
    expect(result).toEqual(initial);
  });

  it("supports reset to default tab order", () => {
    storageMock[DRAWER_TAB_ORDER_KEY] = JSON.stringify(["ports", "doctor", "tasks", "output", "terminal"]);
    expect(getInitialTabOrder(storageMock)).toEqual(["ports", "doctor", "tasks", "output", "terminal"]);

    // 重置
    delete storageMock[DRAWER_TAB_ORDER_KEY];
    expect(getInitialTabOrder(storageMock)).toEqual(DEFAULT_TAB_ORDER);
  });
});
