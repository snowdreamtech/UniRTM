// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import {
  computeToolStatus,
  isVersionLocallyInstalled,
  compareSemverDesc,
  categorizeTools,
  type ToolStatusItem,
  type ToolUpdateCandidate,
} from "./toolStatus";

describe("toolStatus utility", () => {
  describe("compareSemverDesc", () => {
    it("sorts correctly in descending order", () => {
      expect(compareSemverDesc("20.18.3", "20.18.2")).toBeLessThan(0); // v1 > v2
      expect(compareSemverDesc("18.0.0", "20.0.0")).toBeGreaterThan(0); // v1 < v2
      expect(compareSemverDesc("v1.2.3", "1.2.3")).toBe(0); // equal ignoring prefix
    });
  });

  describe("isVersionLocallyInstalled", () => {
    it("matches exact and prefixed versions", () => {
      const installed = ["20.18.3", "v21.0.0"];
      expect(isVersionLocallyInstalled(installed, "20.18.3")).toBe(true);
      expect(isVersionLocallyInstalled(installed, "v20.18.3")).toBe(true);
      expect(isVersionLocallyInstalled(installed, "21.0.0")).toBe(true);
      expect(isVersionLocallyInstalled(installed, "18.0.0")).toBe(false);
      expect(isVersionLocallyInstalled([], "20.18.3")).toBe(false);
      expect(isVersionLocallyInstalled(null, "20.18.3")).toBe(false);
    });
  });

  describe("computeToolStatus", () => {
    it("returns unconfigured when no version is specified", () => {
      const tool: ToolStatusItem = {
        id: "nodejs",
        active_version: "",
        global_version: "",
        installed_versions: ["20.18.3"],
      };
      const res = computeToolStatus(tool, "workspace");
      expect(res.state).toBe("unconfigured");
      expect(res.configuredVersion).toBe("");
      expect(res.isInstalled).toBe(false);
    });

    it("returns missing when configured in workspace but not installed locally", () => {
      const tool: ToolStatusItem = {
        id: "nodejs",
        name: "Node.js",
        active_version: "22.1.0",
        installed_versions: ["20.18.3", "20.18.2"], // 磁盘上没有 22.1.0
      };
      const res = computeToolStatus(tool, "workspace");
      expect(res.state).toBe("missing");
      expect(res.configuredVersion).toBe("22.1.0");
      expect(res.isInstalled).toBe(false);
      expect(res.upgradableVersion).toBeNull();
    });

    it("returns ready when configured and installed without available updates", () => {
      const tool: ToolStatusItem = {
        id: "go",
        name: "Go",
        active_version: "1.23.1",
        installed_versions: ["1.23.1", "1.22.0"],
      };
      const updates: ToolUpdateCandidate[] = [{ id: "go", latest_version: "1.23.1", has_update: false }];
      const res = computeToolStatus(tool, "workspace", updates);
      expect(res.state).toBe("ready");
      expect(res.configuredVersion).toBe("1.23.1");
      expect(res.isInstalled).toBe(true);
      expect(res.upgradableVersion).toBeNull();
    });

    it("returns upgradable when configured, installed, and a strictly newer version is available", () => {
      const tool: ToolStatusItem = {
        id: "python",
        name: "Python",
        active_version: "3.12.0",
        installed_versions: ["3.12.0"],
      };
      const updates: ToolUpdateCandidate[] = [{ id: "python", latest_version: "3.13.1", has_update: true }];
      const res = computeToolStatus(tool, "workspace", updates);
      expect(res.state).toBe("upgradable");
      expect(res.configuredVersion).toBe("3.12.0");
      expect(res.upgradableVersion).toBe("3.13.1");
      expect(res.isInstalled).toBe(true);
    });

    it("respects pinned flag and suppresses upgradable state", () => {
      const tool: ToolStatusItem = {
        id: "rust",
        name: "Rust",
        active_version: "1.80.0",
        installed_versions: ["1.80.0"],
        pinned: true, // 锁定版本
      };
      const updates: ToolUpdateCandidate[] = [{ id: "rust", latest_version: "1.82.0", has_update: true }];
      const res = computeToolStatus(tool, "workspace", updates);
      expect(res.state).toBe("ready"); // 不触发 upgradable
      expect(res.upgradableVersion).toBeNull();
    });

    it("evaluates user scope correctly", () => {
      const tool: ToolStatusItem = {
        id: "bun",
        global_version: "1.1.20",
        active_version: "1.2.0", // workspace 独占
        installed_versions: ["1.1.20"],
      };
      const resUser = computeToolStatus(tool, "user");
      expect(resUser.state).toBe("ready");
      expect(resUser.configuredVersion).toBe("1.1.20");

      const resWorkspace = computeToolStatus(tool, "workspace");
      expect(resWorkspace.state).toBe("missing"); // 本地没有 1.2.0
      expect(resWorkspace.configuredVersion).toBe("1.2.0");
    });
  });

  describe("categorizeTools", () => {
    const mockTools = [
      {
        id: "node",
        name: "Node.js",
        active_version: "22.1.0",
        installed_versions: ["20.18.0"], // 缺少 22.1.0 -> missing
      },
      {
        id: "go",
        name: "Go",
        active_version: "1.23.0",
        installed_versions: ["1.23.0"], // 存在升级 1.24.0 -> upgradable
      },
      {
        id: "python",
        name: "Python",
        active_version: "3.12.0",
        installed_versions: ["3.12.0"], // 已就绪 -> ready
      },
      {
        id: "rust",
        name: "Rust",
        active_version: "", // 未配置 -> unconfigured / skipped
        installed_versions: ["1.80.0"],
      },
    ];

    const mockUpdates: ToolUpdateCandidate[] = [
      {
        id: "go",
        name: "Go",
        latest_version: "1.24.0",
        has_update: true,
      },
    ];

    it("categorizes tools into missing, upgradable, and ready correctly", () => {
      const result = categorizeTools(mockTools, "workspace", mockUpdates);

      expect(result.missing).toHaveLength(1);
      expect(result.missing[0].id).toBe("node");
      expect(result.missing[0].state).toBe("missing");

      expect(result.upgradable).toHaveLength(1);
      expect(result.upgradable[0].id).toBe("go");
      expect(result.upgradable[0].state).toBe("upgradable");
      expect(result.upgradable[0].upgradableVersion).toBe("1.24.0");

      expect(result.ready).toHaveLength(1);
      expect(result.ready[0].id).toBe("python");
      expect(result.ready[0].state).toBe("ready");

      expect(result.allFiltered).toHaveLength(3);
    });

    it("filters tools matching searchQuery", () => {
      const result = categorizeTools(mockTools, "workspace", mockUpdates, "node");

      expect(result.missing).toHaveLength(1);
      expect(result.upgradable).toHaveLength(0);
      expect(result.ready).toHaveLength(0);
      expect(result.allFiltered).toHaveLength(1);
    });

    it("returns empty arrays when no tools match filter", () => {
      const result = categorizeTools(mockTools, "workspace", mockUpdates, "nonexistent");

      expect(result.missing).toHaveLength(0);
      expect(result.upgradable).toHaveLength(0);
      expect(result.ready).toHaveLength(0);
      expect(result.allFiltered).toHaveLength(0);
    });
  });
});
