// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, vi, beforeEach } from "vitest";
import { useActiveProject } from "./useActiveProject";

describe("useActiveProject", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    const storage: Record<string, string> = {};
    (global as any).localStorage = {
      getItem: (k: string) => storage[k] || null,
      setItem: (k: string, v: string) => {
        storage[k] = v;
      },
      removeItem: (k: string) => {
        delete storage[k];
      },
      clear: () => {},
      length: 0,
      key: () => null,
    };
  });

  it("loads and resolves projects from workspace roots", async () => {
    const mockRoots = {
      roots: [
        { id: "root-UniRTM", name: "UniRTM", path: "UniRTM" },
        { id: "root-go", name: "go", path: "go" },
      ],
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue(mockRoots),
    });

    const { loadProjects, availableProjects, resolveProjectFromPath, onFileFocused, activeProject, isPinned } =
      useActiveProject();

    await loadProjects();

    expect(availableProjects.value.length).toBe(2);

    // Test resolving file paths
    const matchedGo = resolveProjectFromPath("/home/ansible/workspace/snowdreamtech/go/docker/1.27/Dockerfile");
    expect(matchedGo?.name).toBe("go");

    const matchedUni = resolveProjectFromPath("UniRTM/cmd/main.go");
    expect(matchedUni?.name).toBe("UniRTM");

    // Test auto-switching when file focused
    expect(isPinned.value).toBe(false);
    onFileFocused("go/main.go");
    expect(activeProject.value?.name).toBe("go");

    onFileFocused("UniRTM/internal/server.go");
    expect(activeProject.value?.name).toBe("UniRTM");
  });

  it("respects pin lock when user manually locks project", async () => {
    const { setActiveProject, togglePin, isPinned, activeProject, onFileFocused } = useActiveProject();

    setActiveProject("go", true);
    expect(isPinned.value).toBe(true);
    expect(activeProject.value?.name).toBe("go");

    // Try focusing a file from UniRTM while pinned
    onFileFocused("UniRTM/cmd/64.ui.go");
    // Should NOT switch because isPinned is true
    expect(activeProject.value?.name).toBe("go");

    // Unpin
    togglePin();
    expect(isPinned.value).toBe(false);

    // Now focusing UniRTM should switch
    onFileFocused("UniRTM/cmd/64.ui.go");
    expect(activeProject.value?.name).toBe("UniRTM");
  });
});
