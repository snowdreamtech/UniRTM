// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, beforeEach, vi } from "vitest";
import { useWorkspaceBus } from "./useWorkspaceBus";

describe("useWorkspaceBus", () => {
  beforeEach(() => {
    const bus = useWorkspaceBus();
    bus._clearAllListeners();
  });

  it("handles revealInExplorer requests and notifies subscribers", () => {
    const bus = useWorkspaceBus();
    const listener = vi.fn();
    const unsubscribe = bus.onRevealFile(listener);

    bus.revealInExplorer("src/components/BottomDrawer.vue");

    expect(listener).toHaveBeenCalledTimes(1);
    expect(listener).toHaveBeenCalledWith(
      expect.objectContaining({
        path: "src/components/BottomDrawer.vue",
        timestamp: expect.any(Number),
      })
    );
    expect(bus.lastRevealRequest.value?.path).toBe("src/components/BottomDrawer.vue");

    // Unsubscribe
    unsubscribe();
    bus.revealInExplorer("src/App.vue");
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("handles git status change broadcasts", () => {
    const bus = useWorkspaceBus();
    const listener = vi.fn();
    const unsubscribe = bus.onGitStatusChanged(listener);

    bus.notifyGitStatusChanged();
    expect(listener).toHaveBeenCalledTimes(1);

    unsubscribe();
    bus.notifyGitStatusChanged();
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("handles openDiff requests", () => {
    const bus = useWorkspaceBus();
    const listener = vi.fn();
    const unsubscribe = bus.onOpenDiff(listener);

    const diffReq = {
      path: "cmd/main.go",
      originalContent: "func old()",
      modifiedContent: "func new()",
      isStaged: false,
    };

    bus.openDiff(diffReq);
    expect(listener).toHaveBeenCalledTimes(1);
    expect(listener).toHaveBeenCalledWith(diffReq);

    unsubscribe();
  });

  it("handles file reload notifications", () => {
    const bus = useWorkspaceBus();
    const listener = vi.fn();
    const unsubscribe = bus.onFileReloadNeeded(listener);

    bus.notifyFileReloadNeeded(["src/main.ts", "package.json"]);
    expect(listener).toHaveBeenCalledTimes(1);
    expect(listener).toHaveBeenCalledWith(["src/main.ts", "package.json"]);

    unsubscribe();
    bus.notifyFileReloadNeeded(["src/main.ts"]);
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("handles file renamed and deleted broadcasts", () => {
    const bus = useWorkspaceBus();
    const renameListener = vi.fn();
    const deleteListener = vi.fn();

    const unsubRename = bus.onFileRenamed(renameListener);
    const unsubDelete = bus.onFileDeleted(deleteListener);

    bus.notifyFileRenamed({ oldPath: "old.ts", newPath: "new.ts" });
    expect(renameListener).toHaveBeenCalledTimes(1);
    expect(renameListener).toHaveBeenCalledWith({ oldPath: "old.ts", newPath: "new.ts" });

    bus.notifyFileDeleted("deprecated.ts");
    expect(deleteListener).toHaveBeenCalledTimes(1);
    expect(deleteListener).toHaveBeenCalledWith("deprecated.ts");

    unsubRename();
    unsubDelete();
    bus.notifyFileRenamed({ oldPath: "a", newPath: "b" });
    bus.notifyFileDeleted("c");
    expect(renameListener).toHaveBeenCalledTimes(1);
    expect(deleteListener).toHaveBeenCalledTimes(1);
  });

  it("records and updates active file path", () => {
    const bus = useWorkspaceBus();
    expect(bus.currentActiveFilePath.value).toBe("");

    bus.setActiveFilePath("frontend/src/types/credential.ts");
    expect(bus.currentActiveFilePath.value).toBe("frontend/src/types/credential.ts");
  });
});
