import { describe, it, expect, beforeEach } from "vitest";
import {
  ActivityTabId,
  ACTIVITY_BAR_ORDER_KEY,
  ACTIVITY_BAR_HIDDEN_KEY,
  DEFAULT_ACTIVITY_TAB_ORDER,
  getInitialActivityTabOrder,
  getInitialHiddenActivityTabs,
  reorderActivityTabs,
  toggleActivityTabVisibility,
} from "./activityBarTabs";

describe("ActivityBar Tab Drag, Drop & Visibility Management", () => {
  let storageMock: Record<string, string>;

  beforeEach(() => {
    storageMock = {};
  });

  it("returns default activity tab order when storage is empty", () => {
    const order = getInitialActivityTabOrder(storageMock);
    expect(order).toEqual(DEFAULT_ACTIVITY_TAB_ORDER);
  });

  it("restores custom activity tab order from storage", () => {
    const customOrder: ActivityTabId[] = [
      "doctor",
      "marketplace",
      "environments",
      "tasks",
      "tools",
      "guard",
      "scm",
      "explorer",
      "dashboard",
    ];
    storageMock[ACTIVITY_BAR_ORDER_KEY] = JSON.stringify(customOrder);
    const order = getInitialActivityTabOrder(storageMock);
    expect(order).toEqual(customOrder);
  });

  it("handles missing new tabs gracefully by appending them to the saved order", () => {
    storageMock[ACTIVITY_BAR_ORDER_KEY] = JSON.stringify(["tools", "dashboard", "explorer"]);
    const order = getInitialActivityTabOrder(storageMock);
    expect(order[0]).toBe("tools");
    expect(order[1]).toBe("dashboard");
    expect(order[2]).toBe("explorer");
    // All default tabs should still be present
    expect(order.length).toBe(DEFAULT_ACTIVITY_TAB_ORDER.length);
    for (const id of DEFAULT_ACTIVITY_TAB_ORDER) {
      expect(order).toContain(id);
    }
  });

  it("filters out invalid tab IDs stored in storage", () => {
    storageMock[ACTIVITY_BAR_ORDER_KEY] = JSON.stringify(["invalid_tab", "tools", "tasks"]);
    const order = getInitialActivityTabOrder(storageMock);
    expect(order).not.toContain("invalid_tab");
    expect(order[0]).toBe("tools");
    expect(order[1]).toBe("tasks");
  });

  it("reorders tabs correctly when dropped 'before' target", () => {
    const initial: ActivityTabId[] = ["dashboard", "explorer", "scm", "tools"];
    const reordered = reorderActivityTabs(initial, "tools", "explorer", "before");
    expect(reordered).toEqual(["dashboard", "tools", "explorer", "scm"]);
  });

  it("reorders tabs correctly when dropped 'after' target", () => {
    const initial: ActivityTabId[] = ["dashboard", "explorer", "scm", "tools"];
    const reordered = reorderActivityTabs(initial, "dashboard", "scm", "after");
    expect(reordered).toEqual(["explorer", "scm", "dashboard", "tools"]);
  });

  it("no-ops if source and target are the same", () => {
    const initial: ActivityTabId[] = ["dashboard", "explorer", "scm"];
    const reordered = reorderActivityTabs(initial, "explorer", "explorer", "before");
    expect(reordered).toEqual(initial);
  });

  it("toggles tab visibility and updates hidden set", () => {
    const hidden = new Set<ActivityTabId>();
    const order = [...DEFAULT_ACTIVITY_TAB_ORDER];

    const res1 = toggleActivityTabVisibility(hidden, order, "doctor", "dashboard");
    expect(res1.newHidden.has("doctor")).toBe(true);
    expect(res1.newCurrentTab).toBe("dashboard");

    const res2 = toggleActivityTabVisibility(res1.newHidden, order, "doctor", "dashboard");
    expect(res2.newHidden.has("doctor")).toBe(false);
  });

  it("switches currentTab if the active tab is hidden", () => {
    const hidden = new Set<ActivityTabId>();
    const order: ActivityTabId[] = ["dashboard", "explorer", "scm"];

    const res = toggleActivityTabVisibility(hidden, order, "dashboard", "dashboard");
    expect(res.newHidden.has("dashboard")).toBe(true);
    expect(res.newCurrentTab).toBe("explorer");
  });

  it("restores hidden tabs from storage", () => {
    storageMock[ACTIVITY_BAR_HIDDEN_KEY] = JSON.stringify(["tasks", "marketplace"]);
    const hidden = getInitialHiddenActivityTabs(storageMock);
    expect(hidden.has("tasks")).toBe(true);
    expect(hidden.has("marketplace")).toBe(true);
    expect(hidden.has("dashboard")).toBe(false);
  });

  it("prevents hiding the last remaining visible tab", () => {
    const hidden = new Set<ActivityTabId>([
      "explorer",
      "scm",
      "guard",
      "tools",
      "tasks",
      "environments",
      "marketplace",
      "doctor",
    ]);
    const order = [...DEFAULT_ACTIVITY_TAB_ORDER];

    const res = toggleActivityTabVisibility(hidden, order, "dashboard", "dashboard");
    // Should not hide dashboard because it is the only visible tab
    expect(res.newHidden.has("dashboard")).toBe(false);
    expect(res.newCurrentTab).toBe("dashboard");
  });
});
