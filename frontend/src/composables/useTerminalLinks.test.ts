// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, vi, beforeEach } from "vitest";
import { parseTerminalLine, createTerminalLinkProvider } from "./useTerminalLinks";
import { useWorkspaceBus } from "./useWorkspaceBus";

describe("useTerminalLinks", () => {
  const bus = useWorkspaceBus();

  beforeEach(() => {
    bus._clearAllListeners();
  });

  describe("parseTerminalLine", () => {
    it("should return empty array for empty or blank lines", () => {
      expect(parseTerminalLine("")).toEqual([]);
      expect(parseTerminalLine("   ")).toEqual([]);
    });

    it("should parse standard web URLs", () => {
      const line = "Listening on http://localhost:5173/ and https://unirtm.dev/docs.";
      const links = parseTerminalLine(line);

      expect(links.length).toBe(2);
      expect(links[0]).toMatchObject({
        type: "url",
        text: "http://localhost:5173/",
        url: "http://localhost:5173/",
      });
      expect(links[1]).toMatchObject({
        type: "url",
        text: "https://unirtm.dev/docs",
        url: "https://unirtm.dev/docs",
      });
    });

    it("should parse file path without line numbers", () => {
      const line = "Built file dist/assets/index.js successfully in 320ms";
      const links = parseTerminalLine(line);

      expect(links.length).toBe(1);
      expect(links[0]).toMatchObject({
        type: "file",
        filePath: "dist/assets/index.js",
        line: undefined,
        column: undefined,
      });
    });

    it("should parse file path with line number", () => {
      const line = "Error in src/components/App.vue:42";
      const links = parseTerminalLine(line);

      expect(links.length).toBe(1);
      expect(links[0]).toMatchObject({
        type: "file",
        filePath: "src/components/App.vue",
        line: 42,
        column: undefined,
      });
    });

    it("should parse file path with line and column numbers and strip leading ./", () => {
      const line = "  at Module (./frontend/src/views/ExplorerView.vue:105:12)";
      const links = parseTerminalLine(line);

      expect(links.length).toBe(1);
      expect(links[0]).toMatchObject({
        type: "file",
        filePath: "frontend/src/views/ExplorerView.vue",
        line: 105,
        column: 12,
      });
    });

    it("should parse Python stack trace format", () => {
      const line = '  File "backend/app/main.py", line 55, in start_server';
      const links = parseTerminalLine(line);

      expect(links.length).toBe(1);
      expect(links[0]).toMatchObject({
        type: "file",
        filePath: "backend/app/main.py",
        line: 55,
        column: 1,
      });
    });

    it("should handle multiple links on the same line ordered by position", () => {
      const line = "See src/foo.ts:10 and check https://github.com/snowdreamtech";
      const links = parseTerminalLine(line);

      expect(links.length).toBe(2);
      expect(links[0].type).toBe("file");
      expect(links[0].filePath).toBe("src/foo.ts");
      expect(links[0].line).toBe(10);

      expect(links[1].type).toBe("url");
      expect(links[1].url).toBe("https://github.com/snowdreamtech");
    });
  });

  describe("createTerminalLinkProvider", () => {
    it("should register link provider and invoke openInEditor on file link click", () => {
      let openInEditorPayload: any = null;
      bus.onOpenInEditor((req) => {
        openInEditorPayload = req;
      });

      const mockLine = {
        translateToString: () => "Failed at src/main.go:42:5",
      };

      const mockTerm: any = {
        buffer: {
          active: {
            getLine: (idx: number) => (idx === 0 ? mockLine : null),
          },
        },
      };

      const provider = createTerminalLinkProvider(mockTerm);
      let capturedLinks: any[] | undefined;

      provider.provideLinks(1, (links) => {
        capturedLinks = links;
      });

      expect(capturedLinks).toBeDefined();
      expect(capturedLinks?.length).toBe(1);
      expect(capturedLinks![0].text).toBe("src/main.go:42:5");

      // Simulate click
      capturedLinks![0].activate({} as MouseEvent, capturedLinks![0].text);

      expect(openInEditorPayload).not.toBeNull();
      expect(openInEditorPayload.path).toBe("src/main.go");
      expect(openInEditorPayload.line).toBe(42);
      expect(openInEditorPayload.column).toBe(5);
    });

    it("should support custom onUrlClick option", () => {
      const onUrlClick = vi.fn();
      const mockLine = {
        translateToString: () => "Visit https://google.com for info",
      };

      const mockTerm: any = {
        buffer: {
          active: {
            getLine: (idx: number) => (idx === 0 ? mockLine : null),
          },
        },
      };

      const provider = createTerminalLinkProvider(mockTerm, { onUrlClick });
      let capturedLinks: any[] | undefined;

      provider.provideLinks(1, (links) => {
        capturedLinks = links;
      });

      expect(capturedLinks?.length).toBe(1);
      capturedLinks![0].activate({} as MouseEvent, capturedLinks![0].text);

      expect(onUrlClick).toHaveBeenCalledWith("https://google.com");
    });
  });
});
