import { describe, it, expect, beforeEach, vi } from "vitest";
import { useFileTree } from "./useFileTree";

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
});
