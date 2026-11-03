// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, beforeEach, vi } from "vitest";
import { useTerminalPersistence, stripAnsiCodes } from "./useTerminalPersistence";

describe("useTerminalPersistence", () => {
  const persistence = useTerminalPersistence();

  beforeEach(() => {
    persistence._clearAllForTest();
  });

  describe("stripAnsiCodes", () => {
    it("returns empty string on empty input", () => {
      expect(stripAnsiCodes("")).toBe("");
    });

    it("strips standard color and bold ansi sequences", () => {
      const colored = "\u001b[1m\u001b[32mSUCCESS:\u001b[0m Build passed in \u001b[33m120ms\u001b[0m";
      expect(stripAnsiCodes(colored)).toBe("SUCCESS: Build passed in 120ms");
    });
  });

  describe("buffer storage operations", () => {
    it("appends chunks and retrieves buffer", () => {
      persistence.appendChunk("pane-1", "Hello ");
      persistence.appendChunk("pane-1", "World\n");

      expect(persistence.hasBuffer("pane-1")).toBe(true);
      expect(persistence.getBuffer("pane-1")).toBe("Hello World\n");
    });

    it("isolates buffers by paneId", () => {
      persistence.appendChunk("pane-1", "data 1");
      persistence.appendChunk("pane-2", "data 2");

      expect(persistence.getBuffer("pane-1")).toBe("data 1");
      expect(persistence.getBuffer("pane-2")).toBe("data 2");
    });

    it("clears buffer for a specific pane", () => {
      persistence.appendChunk("pane-1", "active logs");
      expect(persistence.hasBuffer("pane-1")).toBe(true);

      persistence.clearBuffer("pane-1");
      expect(persistence.hasBuffer("pane-1")).toBe(false);
      expect(persistence.getBuffer("pane-1")).toBe("");
    });
  });

  describe("exportLog", () => {
    it("returns false if no buffer exists", () => {
      expect(persistence.exportLog("empty-pane")).toBe(false);
    });

    it("triggers file download when buffer exists", () => {
      persistence.appendChunk("pane-test", "\u001b[31mError 500\u001b[0m in handler");

      const clickMock = vi.fn();
      const mockElement = {
        href: "",
        download: "",
        click: clickMock,
      };

      const originalWindow = (global as any).window;
      const originalDocument = (global as any).document;

      (global as any).window = {};
      (global as any).document = {
        createElement: vi.fn().mockReturnValue(mockElement),
        body: {
          appendChild: vi.fn(),
          removeChild: vi.fn(),
        },
      };

      const origCreate = global.URL?.createObjectURL;
      const origRevoke = global.URL?.revokeObjectURL;
      if (!global.URL) {
        (global as any).URL = {};
      }
      global.URL.createObjectURL = vi.fn().mockReturnValue("blob:mock-url");
      global.URL.revokeObjectURL = vi.fn();

      try {
        const success = persistence.exportLog("pane-test", "bash", true);
        expect(success).toBe(true);
        expect(clickMock).toHaveBeenCalled();
        expect(mockElement.download).toContain("unirtm-bash-");
      } finally {
        (global as any).window = originalWindow;
        (global as any).document = originalDocument;
        if (origCreate) global.URL.createObjectURL = origCreate;
        if (origRevoke) global.URL.revokeObjectURL = origRevoke;
      }
    });
  });
});
