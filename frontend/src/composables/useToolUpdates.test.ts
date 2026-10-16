import { describe, it, expect, beforeEach } from "vitest";
import { useToolUpdates } from "./useToolUpdates";

describe("useToolUpdates", () => {
  const {
    updatableCount,
    hasUpdate,
    setUpdateResults,
    markToolUpdated,
    clearUpdates,
  } = useToolUpdates();

  beforeEach(() => {
    clearUpdates();
  });

  it("should have zero updatable count initially and no dots", () => {
    expect(updatableCount.value).toBe(0);
    expect(hasUpdate("go")).toBe(false);
    expect(hasUpdate("node")).toBe(false);
  });

  it("should ignore tools that are pinned even if has_update is true", () => {
    setUpdateResults([
      {
        id: "go",
        name: "Go",
        current_version: "1.24.0",
        latest_version: "1.27.2",
        has_update: true,
        pinned: true,
        pin_reason: "Production compatibility",
      },
    ]);

    // Pinned tools must NOT trigger update dot or increase badge count
    expect(hasUpdate("go")).toBe(false);
    expect(updatableCount.value).toBe(0);
  });

  it("should recognize unpinned tools with update available", () => {
    setUpdateResults([
      {
        id: "github:aquasecurity/trivy",
        name: "trivy",
        current_version: "0.75.0",
        latest_version: "1.0.0",
        has_update: true,
        pinned: false,
      },
      {
        id: "node",
        name: "Node.js",
        current_version: "26.8.1",
        latest_version: "26.8.1",
        has_update: false,
        pinned: false,
      },
    ]);

    expect(hasUpdate("github:aquasecurity/trivy")).toBe(true);
    expect(hasUpdate("node")).toBe(false);
    expect(updatableCount.value).toBe(1);
  });

  it("should clear update status when tool is marked updated", () => {
    setUpdateResults([
      {
        id: "github:astral-sh/ruff",
        name: "ruff",
        current_version: "0.16.6",
        latest_version: "1.0.0",
        has_update: true,
        pinned: false,
      },
    ]);

    expect(hasUpdate("github:astral-sh/ruff")).toBe(true);
    expect(updatableCount.value).toBe(1);

    markToolUpdated("github:astral-sh/ruff");

    expect(hasUpdate("github:astral-sh/ruff")).toBe(false);
    expect(updatableCount.value).toBe(0);
  });

  it("should match tool updates by full ID, repo name, and short name", () => {
    setUpdateResults([
      {
        id: "github:astral-sh/ruff",
        name: "github:astral-sh/ruff",
        current_version: "0.16.6",
        latest_version: "0.16.10",
        has_update: true,
        pinned: false,
      },
    ]);

    expect(hasUpdate("github:astral-sh/ruff")).toBe(true);
    expect(hasUpdate("astral-sh/ruff")).toBe(true);
    expect(hasUpdate("ruff")).toBe(true);
    expect(hasUpdate("other-tool")).toBe(false);
  });

  it("should sync local tool updates when a higher version is already installed", () => {
    const { syncFromToolsList } = useToolUpdates();
    syncFromToolsList([
      {
        id: "github:astral-sh/ruff",
        name: "github:astral-sh/ruff",
        active_version: "0.16.6",
        installed_versions: ["0.16.10", "0.16.6"],
        pinned: false,
      },
      {
        id: "github:checkmake/checkmake",
        name: "checkmake",
        active_version: "v0.3.2",
        installed_versions: ["v0.3.2"],
        pinned: false,
      },
      {
        id: "node",
        name: "Node.js",
        active_version: "26.8.1",
        installed_versions: ["26.8.1", "24.0.0"],
        pinned: false,
      },
    ]);

    expect(hasUpdate("github:astral-sh/ruff")).toBe(true);
    expect(hasUpdate("ruff")).toBe(true);
    expect(hasUpdate("github:checkmake/checkmake")).toBe(false);
    expect(hasUpdate("node")).toBe(false);
    expect(updatableCount.value).toBe(1);
  });
});
