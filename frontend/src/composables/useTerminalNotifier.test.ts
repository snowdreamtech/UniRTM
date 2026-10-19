// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, beforeEach, vi } from "vitest";
import { useTerminalNotifier } from "./useTerminalNotifier";
import { useToast } from "./useToast";

vi.mock("./useToast", () => {
  const showToast = vi.fn();
  return {
    useToast: () => ({
      showToast,
      toastMessage: { value: "" },
      toastType: { value: "info" },
    }),
  };
});

describe("useTerminalNotifier", () => {
  const notifier = useTerminalNotifier();
  const { showToast } = useToast();

  beforeEach(() => {
    notifier._clearAllForTest();
    vi.clearAllMocks();
  });

  it("ignores tracking when invalid parameters are given", () => {
    notifier.startTracking("", "");
    expect(notifier.getActiveTrackedTask("")).toBeNull();

    notifier.startTracking("pane-1", "   ");
    expect(notifier.getActiveTrackedTask("pane-1")).toBeNull();
  });

  it("starts tracking a command correctly", () => {
    notifier.startTracking("pane-1", "npm run build");
    const task = notifier.getActiveTrackedTask("pane-1");
    expect(task).not.toBeNull();
    expect(task?.command).toBe("npm run build");
    expect(task?.hasError).toBe(false);
  });

  it("detects errors in output stream", () => {
    notifier.startTracking("pane-1", "go build ./...");
    notifier.feedOutput("pane-1", "Compiling package...\r\n");
    let task = notifier.getActiveTrackedTask("pane-1");
    expect(task?.hasError).toBe(false);

    notifier.feedOutput("pane-1", "error: syntax error in main.go\r\n");
    task = notifier.getActiveTrackedTask("pane-1");
    expect(task?.hasError).toBe(true);
  });

  it("completes short task without triggering toast", () => {
    notifier.setThreshold(5000); // 5s 阈值
    notifier.startTracking("pane-1", "ls -la");

    const result = notifier.feedOutput("pane-1", "total 32\r\nuser@host:~$ ");
    expect(result).not.toBeNull();
    expect(result?.notified).toBe(false);
    expect(result?.success).toBe(true);
    expect(showToast).not.toHaveBeenCalled();
    expect(notifier.getActiveTrackedTask("pane-1")).toBeNull();
  });

  it("triggers success toast when long task finishes successfully", () => {
    notifier.setThreshold(100); // 测试用低阈值 100ms
    notifier.startTracking("pane-1", "docker compose up -d");

    const task = notifier.getActiveTrackedTask("pane-1");
    if (task) {
      task.startTime = Date.now() - 15000; // 模拟耗时 15s
    }

    const result = notifier.feedOutput("pane-1", "Container unirtm Started\r\nuser@host:~$ ");
    expect(result).not.toBeNull();
    expect(result?.notified).toBe(true);
    expect(result?.success).toBe(true);
    expect(showToast).toHaveBeenCalledWith(
      expect.stringContaining("completed (15s)"),
      "success"
    );
  });

  it("triggers error toast when long task finishes with errors", () => {
    notifier.setThreshold(100);
    notifier.startTracking("pane-2", "cargo test");

    const task = notifier.getActiveTrackedTask("pane-2");
    if (task) {
      task.startTime = Date.now() - 12000; // 模拟耗时 12s
    }

    notifier.feedOutput("pane-2", "fatal: test suite panic\r\n");
    const result = notifier.feedOutput("pane-2", "user@host:# ");
    expect(result).not.toBeNull();
    expect(result?.notified).toBe(true);
    expect(result?.success).toBe(false);
    expect(showToast).toHaveBeenCalledWith(
      expect.stringContaining("finished with errors (12s)"),
      "error"
    );
  });
});
