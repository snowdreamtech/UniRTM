// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, beforeEach } from "vitest";
import { useTerminalProcessWatcher, detectProcessFromChunk } from "./useTerminalProcessWatcher";

describe("useTerminalProcessWatcher", () => {
  const watcher = useTerminalProcessWatcher();

  beforeEach(() => {
    watcher._clearAllForTest();
  });

  describe("detectProcessFromChunk", () => {
    it("returns null on empty or unrecognized input", () => {
      expect(detectProcessFromChunk("")).toBeNull();
      expect(detectProcessFromChunk("ls -la /tmp")).toBeNull();
    });

    it("detects Node / Vite stack correctly", () => {
      const res = detectProcessFromChunk("[vite] dev server running at: http://localhost:5173/");
      expect(res).not.toBeNull();
      expect(res?.name).toBe("Node");
      expect(res?.color).toBe("#22c55e");
    });

    it("detects Python / pytest correctly", () => {
      const res = detectProcessFromChunk("pytest tests/test_service.py -v");
      expect(res).not.toBeNull();
      expect(res?.name).toBe("Python");
      expect(res?.color).toBe("#3b82f6");
    });

    it("detects Go build / test correctly", () => {
      const res = detectProcessFromChunk("go test ./pkg/broker/...");
      expect(res).not.toBeNull();
      expect(res?.name).toBe("Go");
      expect(res?.color).toBe("#06b6d4");
    });

    it("detects Rust cargo run correctly", () => {
      const res = detectProcessFromChunk("cargo run --release");
      expect(res).not.toBeNull();
      expect(res?.name).toBe("Rust");
      expect(res?.color).toBe("#f97316");
    });

    it("detects Docker commands correctly", () => {
      const res = detectProcessFromChunk("docker compose up -d");
      expect(res).not.toBeNull();
      expect(res?.name).toBe("Docker");
    });
  });

  describe("recordChunk and title computation", () => {
    it("records running process and computes effective title", () => {
      expect(watcher.getEffectiveTitle("pane-1", "bash")).toBe("bash");

      watcher.recordChunk("pane-1", "npm run dev");
      const process = watcher.getActiveProcess("pane-1");
      expect(process).not.toBeNull();
      expect(process?.name).toBe("Node");

      expect(watcher.getEffectiveTitle("pane-1", "bash")).toBe("Node (bash)");

      // Shell prompt finishes the process
      watcher.recordChunk("pane-1", "user@host:~$ ");
      expect(watcher.getActiveProcess("pane-1")).toBeNull();
      expect(watcher.getEffectiveTitle("pane-1", "bash")).toBe("bash");
    });

    it("gives priority to custom user title", () => {
      watcher.setCustomTitle("session-1", "Dev Server");
      expect(watcher.getEffectiveTitle("session-1", "bash")).toBe("Dev Server");

      // Custom title should override even when a process is running
      watcher.recordChunk("session-1", "python main.py");
      expect(watcher.getEffectiveTitle("session-1", "bash")).toBe("Dev Server");
    });
  });
});
