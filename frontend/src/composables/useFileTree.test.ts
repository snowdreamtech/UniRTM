import { describe, it, expect, beforeEach, vi } from "vitest";
import { useFileTree } from "./useFileTree";
import { useGit } from "./useGit";

describe("useFileTree", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    const { collapseAll, searchQuery, cancelInlineAction, selectedNode } = useFileTree();
    collapseAll();
    searchQuery.value = "";
    cancelInlineAction();
    selectedNode.value = null;
  });

  it("handles directory expansion, collapseAll and nested path expansion", () => {
    const { expandedPaths, toggleExpand, expandPath, collapseAll } = useFileTree();

    expect(expandedPaths.value.has("")).toBe(true);

    toggleExpand("src");
    expect(expandedPaths.value.has("src")).toBe(true);

    toggleExpand("src");
    expect(expandedPaths.value.has("src")).toBe(false);

    expandPath("src/components/layout");
    expect(expandedPaths.value.has("src")).toBe(true);
    expect(expandedPaths.value.has("src/components")).toBe(true);
    expect(expandedPaths.value.has("src/components/layout")).toBe(true);

    collapseAll();
    expect(expandedPaths.value.has("src")).toBe(false);
    expect(expandedPaths.value.has("")).toBe(true);
  });

  it("manages inline action lifecycle (create & rename)", () => {
    const {
      inlineAction,
      startInlineCreate,
      startInlineRename,
      cancelInlineAction,
    } = useFileTree();

    startInlineCreate("src", "new-file");
    expect(inlineAction.value).toEqual({
      type: "new-file",
      parentPath: "src",
      defaultValue: "",
    });

    startInlineRename({
      name: "App.vue",
      path: "src/App.vue",
      is_dir: false,
      size: 1024,
      mod_time: new Date().toISOString(),
    });
    expect(inlineAction.value).toEqual({
      type: "rename",
      parentPath: "",
      targetPath: "src/App.vue",
      defaultValue: "App.vue",
    });

    cancelInlineAction();
    expect(inlineAction.value).toBeNull();
  });

  it("filters file tree based on search query", () => {
    const { tree, filteredTree, searchQuery } = useFileTree();

    tree.value = {
      name: "UniRTM",
      path: "",
      is_dir: true,
      size: 0,
      mod_time: "2026-10-18T00:00:00Z",
      children: [
        {
          name: "cmd",
          path: "cmd",
          is_dir: true,
          size: 0,
          mod_time: "2026-10-18T00:00:00Z",
          children: [
            {
              name: "main.go",
              path: "cmd/main.go",
              is_dir: false,
              size: 500,
              mod_time: "2026-10-18T00:00:00Z",
            },
          ],
        },
        {
          name: "package.json",
          path: "package.json",
          is_dir: false,
          size: 120,
          mod_time: "2026-10-18T00:00:00Z",
        },
      ],
    };

    searchQuery.value = "main";
    expect(filteredTree.value).not.toBeNull();
    expect(filteredTree.value?.children?.length).toBe(1);
    expect(filteredTree.value?.children?.[0].name).toBe("cmd");
    expect(filteredTree.value?.children?.[0].children?.[0].name).toBe("main.go");

    searchQuery.value = "nonexistent";
    expect(filteredTree.value).toBeNull();

    searchQuery.value = "";
    expect(filteredTree.value?.children?.length).toBe(2);
  });

  it("fetches directory tree from API", async () => {
    const mockTree = {
      name: "UniRTM",
      path: "",
      is_dir: true,
      size: 0,
      mod_time: "2026-10-18T00:00:00Z",
      children: [],
    };

    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true, data: mockTree }),
    } as any);

    const { tree, fetchTree, isLoading } = useFileTree();
    await fetchTree();

    expect(isLoading.value).toBe(false);
    expect(tree.value?.name).toBe("UniRTM");
  });

  it("normalizes backend { root, name, entries } response format", async () => {
    const rawBackendResp = {
      root: "/workspace",
      name: "UniRTM",
      path: "",
      entries: [
        {
          name: "cmd",
          path: "cmd",
          is_dir: true,
          size: 0,
          mod_time: "2026-10-18T00:00:00Z",
          children: [
            {
              name: "main.go",
              path: "cmd/main.go",
              is_dir: false,
              size: 400,
              mod_time: "2026-10-18T00:00:00Z",
            },
          ],
        },
      ],
    };

    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce({
      ok: true,
      json: async () => rawBackendResp,
    } as any);

    const { tree, fetchTree } = useFileTree();
    await fetchTree();

    expect(tree.value?.name).toBe("UniRTM");
    expect(tree.value?.children?.length).toBe(1);
    expect(tree.value?.children?.[0].name).toBe("cmd");
    expect(tree.value?.children?.[0].children?.[0].name).toBe("main.go");
  });

  it("reveals nested path and selects the target node", () => {
    const { tree, expandedPaths, selectedNode, revealPath } = useFileTree();

    tree.value = {
      name: "UniRTM",
      path: "",
      is_dir: true,
      size: 0,
      mod_time: "2026-10-18T00:00:00Z",
      children: [
        {
          name: "src",
          path: "src",
          is_dir: true,
          size: 0,
          mod_time: "2026-10-18T00:00:00Z",
          children: [
            {
              name: "App.vue",
              path: "src/App.vue",
              is_dir: false,
              size: 200,
              mod_time: "2026-10-18T00:00:00Z",
            },
          ],
        },
      ],
    };

    const target = revealPath("src/App.vue");
    expect(target?.name).toBe("App.vue");
    expect(selectedNode.value?.path).toBe("src/App.vue");
    expect(expandedPaths.value.has("src")).toBe(true);
    expect(expandedPaths.value.has("src/App.vue")).toBe(true);
  });

  it("filters tree to show only Git changed files in Git Focus mode", () => {
    const { tree, filteredTree, isGitFilterActive, toggleGitFilter } = useFileTree();

    tree.value = {
      name: "UniRTM",
      path: "",
      is_dir: true,
      size: 0,
      mod_time: "2026-10-18T00:00:00Z",
      children: [
        {
          name: "src",
          path: "src",
          is_dir: true,
          size: 0,
          mod_time: "2026-10-18T00:00:00Z",
          children: [
            {
              name: "changed.ts",
              path: "src/changed.ts",
              is_dir: false,
              size: 200,
              mod_time: "2026-10-18T00:00:00Z",
            },
            {
              name: "unchanged.ts",
              path: "src/unchanged.ts",
              is_dir: false,
              size: 300,
              mod_time: "2026-10-18T00:00:00Z",
            },
          ],
        },
        {
          name: "docs",
          path: "docs",
          is_dir: true,
          size: 0,
          mod_time: "2026-10-18T00:00:00Z",
          children: [
            {
              name: "readme.md",
              path: "docs/readme.md",
              is_dir: false,
              size: 100,
              mod_time: "2026-10-18T00:00:00Z",
            },
          ],
        },
      ],
    };

    // Before Git filter: all children visible
    expect(filteredTree.value?.children?.length).toBe(2);

    // Mock useGit status
    const { status } = useGit();
    status.value.unstaged = [
      { path: "src/changed.ts", status: "M", label: "src/changed.ts" },
    ];

    toggleGitFilter();
    expect(isGitFilterActive.value).toBe(true);

    // After Git filter: docs directory is pruned, only src/changed.ts remains
    expect(filteredTree.value?.children?.length).toBe(1);
    expect(filteredTree.value?.children?.[0].name).toBe("src");
    expect(filteredTree.value?.children?.[0].children?.length).toBe(1);
    expect(filteredTree.value?.children?.[0].children?.[0].name).toBe("changed.ts");

    // Toggle off
    toggleGitFilter();
    expect(isGitFilterActive.value).toBe(false);
    expect(filteredTree.value?.children?.length).toBe(2);
  });

  it("calculates recursive Git change count for directory and file nodes", () => {
    const { getNodeGitChangeCount } = useFileTree();
    const { status } = useGit();

    status.value.unstaged = [
      { path: "src/utils/math.ts", status: "M", label: "src/utils/math.ts" },
      { path: "src/components/Header.vue", status: "M", label: "src/components/Header.vue" },
    ];
    status.value.staged = [
      { path: "src/App.vue", status: "A", label: "src/App.vue" },
    ];

    expect(
      getNodeGitChangeCount({
        name: "src",
        path: "src",
        is_dir: true,
        size: 0,
        mod_time: "2026-10-18T00:00:00Z",
      })
    ).toBe(3);

    expect(
      getNodeGitChangeCount({
        name: "components",
        path: "src/components",
        is_dir: true,
        size: 0,
        mod_time: "2026-10-18T00:00:00Z",
      })
    ).toBe(1);

    expect(
      getNodeGitChangeCount({
        name: "App.vue",
        path: "src/App.vue",
        is_dir: false,
        size: 100,
        mod_time: "2026-10-18T00:00:00Z",
      })
    ).toBe(1);

    expect(
      getNodeGitChangeCount({
        name: "docs",
        path: "docs",
        is_dir: true,
        size: 0,
        mod_time: "2026-10-18T00:00:00Z",
      })
    ).toBe(0);
  });

  it("handles getAbsolutePath, copyPath, copyRelativePath, and copyFileName", async () => {
    const { tree, getAbsolutePath, copyPath, copyRelativePath, copyFileName } = useFileTree();

    tree.value = {
      name: "root",
      path: "/home/user/project",
      is_dir: true,
      size: 0,
      mod_time: "2026-10-18T00:00:00Z",
      children: [],
    };

    // getAbsolutePath
    expect(getAbsolutePath("src/App.vue")).toBe("/home/user/project/src/App.vue");
    expect(getAbsolutePath("/absolute/path/file.ts")).toBe("/absolute/path/file.ts");
    expect(getAbsolutePath("")).toBe("/home/user/project");

    // Mock clipboard
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText: writeTextMock },
      writable: true,
      configurable: true,
    });

    const fullPathResult = await copyPath("src/main.ts");
    expect(fullPathResult).toBe("/home/user/project/src/main.ts");
    expect(writeTextMock).toHaveBeenCalledWith("/home/user/project/src/main.ts");

    const relResult = await copyRelativePath("src/main.ts");
    expect(relResult).toBe("src/main.ts");
    expect(writeTextMock).toHaveBeenCalledWith("src/main.ts");

    const nameResult = await copyFileName("main.ts");
    expect(nameResult).toBe("main.ts");
    expect(writeTextMock).toHaveBeenCalledWith("main.ts");
  });
});

