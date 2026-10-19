import { describe, it, expect, beforeEach, vi } from "vitest";
import { useEditorTabs } from "./useEditorTabs";

describe("useEditorTabs", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    const { closeAllTabs } = useEditorTabs();
    closeAllTabs();
  });

  it("opens a file tab and fetches its content", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce({
      ok: true,
      json: async () => ({ content: "package main\n\nfunc main() {}\n" }),
    } as any);

    const { tabs, activeTabId, activeTab, openTab } = useEditorTabs();

    await openTab({ path: "cmd/main.go", name: "main.go" });

    expect(tabs.value.length).toBe(1);
    expect(activeTabId.value).toBe("cmd/main.go");
    expect(activeTab.value?.name).toBe("main.go");
    expect(activeTab.value?.language).toBe("go");
    expect(activeTab.value?.content).toBe("package main\n\nfunc main() {}\n");
    expect(activeTab.value?.isDirty).toBe(false);
  });

  it("switches to already opened tab without re-fetching", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({ content: "{}" }),
    } as any);

    const { tabs, activeTabId, openTab, switchTab } = useEditorTabs();

    await openTab({ path: "a.json", name: "a.json" });
    await openTab({ path: "b.json", name: "b.json" });

    expect(tabs.value.length).toBe(2);
    expect(activeTabId.value).toBe("b.json");
    expect(fetchSpy).toHaveBeenCalledTimes(2);

    await openTab({ path: "a.json", name: "a.json" });
    expect(activeTabId.value).toBe("a.json");
    expect(fetchSpy).toHaveBeenCalledTimes(2); // no extra fetch

    switchTab("b.json");
    expect(activeTabId.value).toBe("b.json");
  });

  it("tracks dirty state on content modifications", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce({
      ok: true,
      json: async () => ({ content: "initial" }),
    } as any);

    const { activeTab, openTab, updateContent } = useEditorTabs();

    await openTab({ path: "test.txt", name: "test.txt" });
    expect(activeTab.value?.isDirty).toBe(false);

    updateContent("test.txt", "modified content");
    expect(activeTab.value?.isDirty).toBe(true);

    updateContent("test.txt", "initial");
    expect(activeTab.value?.isDirty).toBe(false);
  });

  it("saves tab content and clears dirty flag", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ content: "old" }),
      } as any)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      } as any);

    const { activeTab, openTab, updateContent, saveTab } = useEditorTabs();

    await openTab({ path: "doc.md", name: "doc.md" });
    updateContent("doc.md", "new");
    expect(activeTab.value?.isDirty).toBe(true);

    const success = await saveTab();
    expect(success).toBe(true);
    expect(activeTab.value?.isDirty).toBe(false);
    expect(activeTab.value?.originalContent).toBe("new");
  });

  it("handles tab close, close others, and close all", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({ content: "" }),
    } as any);

    const { tabs, activeTabId, openTab, closeTab, closeOtherTabs, closeAllTabs } = useEditorTabs();

    await openTab({ path: "1.js", name: "1.js" });
    await openTab({ path: "2.js", name: "2.js" });
    await openTab({ path: "3.js", name: "3.js" });

    expect(tabs.value.length).toBe(3);
    expect(activeTabId.value).toBe("3.js");

    // Close active tab 3.js, should fallback to adjacent tab 2.js
    closeTab("3.js");
    expect(tabs.value.length).toBe(2);
    expect(activeTabId.value).toBe("2.js");

    // Close other tabs
    closeOtherTabs("2.js");
    expect(tabs.value.length).toBe(1);
    expect(activeTabId.value).toBe("2.js");

    // Close all
    closeAllTabs();
    expect(tabs.value.length).toBe(0);
    expect(activeTabId.value).toBe("");
  });

  it("handles oversized and binary files gracefully without throwing errors", async () => {
    // 1. Oversized file via pre-known size (e.g. 63MB unirtm executable)
    const { tabs, openTab } = useEditorTabs();
    await openTab({ path: "unirtm", name: "unirtm", size: 66258944 });

    expect(tabs.value.length).toBe(1);
    expect(tabs.value[0].isBinary).toBe(true);
    expect(tabs.value[0].binaryReason).toBe("too_large");
    expect(tabs.value[0].size).toBe(66258944);

    // 2. Known binary extension (e.g. package.zip)
    await openTab({ path: "dist/bundle.zip", name: "bundle.zip" });
    expect(tabs.value.length).toBe(2);
    expect(tabs.value[1].isBinary).toBe(true);
    expect(tabs.value[1].binaryReason).toBe("binary");

    // 3. Oversized file via HTTP 400 response
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: async () => ({ error: "File too large (exceeds 10MB)", size: 15000000 }),
    } as any);

    await openTab({ path: "huge.dat", name: "huge.dat" });
    expect(tabs.value.length).toBe(3);
    expect(tabs.value[2].isBinary).toBe(true);
    expect(tabs.value[2].binaryReason).toBe("too_large");
    expect(tabs.value[2].size).toBe(15000000);
  });

  it("supports closeTabsToRight, closeSavedTabs, and tab pinning", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({ content: "" }),
    } as any);

    const {
      tabs,
      openTab,
      updateContent,
      closeTabsToRight,
      closeSavedTabs,
      closeAllTabs,
      togglePinTab,
    } = useEditorTabs();

    await openTab({ path: "1.js", name: "1.js" });
    await openTab({ path: "2.js", name: "2.js" });
    await openTab({ path: "3.js", name: "3.js" });
    await openTab({ path: "4.js", name: "4.js" });

    expect(tabs.value.length).toBe(4);

    // Pin tab 4.js, it moves to the front
    togglePinTab("4.js");
    expect(tabs.value[0].id).toBe("4.js");
    expect(tabs.value[0].isPinned).toBe(true);

    // Close tabs to right of 2.js (tabs: [4.js (pinned), 1.js, 2.js, 3.js])
    closeTabsToRight("2.js");
    // 3.js is closed, 4.js is retained because it's pinned
    expect(tabs.value.map((t) => t.id)).toEqual(["4.js", "1.js", "2.js"]);

    // Make 1.js dirty
    updateContent("1.js", "modified");
    expect(tabs.value.find((t) => t.id === "1.js")?.isDirty).toBe(true);

    // Close saved tabs: 2.js should close, 1.js (dirty) and 4.js (pinned) stay
    closeSavedTabs();
    expect(tabs.value.map((t) => t.id)).toEqual(["4.js", "1.js"]);

    // Normal closeAllTabs retains pinned tabs
    closeAllTabs(false);
    expect(tabs.value.length).toBe(1);
    expect(tabs.value[0].id).toBe("4.js");

    // Force closeAllTabs removes everything
    closeAllTabs(true);
    expect(tabs.value.length).toBe(0);
  });

  it("opens and switches diff tabs correctly", () => {
    const { tabs, activeTabId, activeTab, openDiffTab } = useEditorTabs();

    openDiffTab({
      path: "cmd/main.go",
      originalContent: "package main\n\n// old",
      modifiedContent: "package main\n\n// new",
      isStaged: false,
    });

    expect(tabs.value.length).toBe(1);
    expect(activeTabId.value).toBe("diff:working:cmd/main.go");
    expect(activeTab.value?.isDiff).toBe(true);
    expect(activeTab.value?.name).toContain("Working Tree");
    expect(activeTab.value?.diffMeta?.modifiedContent).toBe("package main\n\n// new");

    // Opening again with same path switches to existing tab
    openDiffTab({
      path: "cmd/main.go",
      originalContent: "package main\n\n// old 2",
      modifiedContent: "package main\n\n// new 2",
      isStaged: false,
    });
    expect(tabs.value.length).toBe(1);
    expect(activeTab.value?.diffMeta?.modifiedContent).toBe("package main\n\n// new 2");

    // Opening staged diff creates a separate staged diff tab
    openDiffTab({
      path: "cmd/main.go",
      originalContent: "package main",
      modifiedContent: "package main\n// staged",
      isStaged: true,
    });
    expect(tabs.value.length).toBe(2);
    expect(activeTabId.value).toBe("diff:staged:cmd/main.go");
  });

  it("sets tab git status correctly on matching normal tabs", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({ content: "test content" }),
    } as any);

    const { tabs, openTab, setTabGitStatus } = useEditorTabs();
    await openTab({ path: "foo.ts", name: "foo.ts" });

    expect(tabs.value[0].gitStatus).toBeUndefined();

    setTabGitStatus("foo.ts", "M");
    expect(tabs.value[0].gitStatus).toBe("M");
  });

  it("safely saves dirty tabs matching specified paths", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce({ ok: true, json: async () => ({ content: "init-1" }) } as any)
      .mockResolvedValueOnce({ ok: true, json: async () => ({ content: "init-2" }) } as any)
      .mockResolvedValueOnce({ ok: true, json: async () => ({ success: true }) } as any);

    const { tabs, openTab, updateContent, saveDirtyTabsForPaths } = useEditorTabs();

    await openTab({ path: "file1.ts", name: "file1.ts" });
    await openTab({ path: "file2.ts", name: "file2.ts" });

    updateContent("file1.ts", "dirty-1");
    updateContent("file2.ts", "dirty-2");

    expect(tabs.value[0].isDirty).toBe(true);
    expect(tabs.value[1].isDirty).toBe(true);

    // Save only file1.ts
    await saveDirtyTabsForPaths(["file1.ts"]);

    expect(tabs.value[0].isDirty).toBe(false);
    expect(tabs.value[1].isDirty).toBe(true);
  });

  it("reloads tabs matching specified paths and resets dirty status", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce({ ok: true, json: async () => ({ content: "old-content" }) } as any)
      .mockResolvedValueOnce({ ok: true, json: async () => ({ content: "reloaded-content" }) } as any);

    const { tabs, openTab, updateContent, reloadTabsForPaths } = useEditorTabs();

    await openTab({ path: "file.txt", name: "file.txt" });
    updateContent("file.txt", "uncommitted-dirty");
    await reloadTabsForPaths(["file.txt"]);

    expect(tabs.value[0].content).toBe("reloaded-content");
    expect(tabs.value[0].originalContent).toBe("reloaded-content");
    expect(tabs.value[0].isDirty).toBe(false);
  });

  it("seamlessly migrates open tabs on file or directory rename", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce({ ok: true, json: async () => ({ content: "const a = 1;" }) } as any)
      .mockResolvedValueOnce({ ok: true, json: async () => ({ content: "const b = 2;" }) } as any);

    const { tabs, activeTabId, openTab, handleFileRenamed } = useEditorTabs();

    await openTab({ path: "src/oldName.ts", name: "oldName.ts" });
    await openTab({ path: "src/components/Header.vue", name: "Header.vue" });

    // Single file rename
    handleFileRenamed({ oldPath: "src/oldName.ts", newPath: "src/newName.ts" });
    expect(tabs.value[0].path).toBe("src/newName.ts");
    expect(tabs.value[0].name).toBe("newName.ts");
    expect(tabs.value[0].language).toBe("typescript");

    // Directory rename
    handleFileRenamed({ oldPath: "src/components", newPath: "src/widgets" });
    expect(tabs.value[1].path).toBe("src/widgets/Header.vue");
    expect(tabs.value[1].name).toBe("Header.vue");
    expect(activeTabId.value).toBe("src/widgets/Header.vue");
  });

  it("automatically closes tabs when their files are deleted", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce({ ok: true, json: async () => ({ content: "f1" }) } as any)
      .mockResolvedValueOnce({ ok: true, json: async () => ({ content: "f2" }) } as any);

    const { tabs, openTab, handleFileDeleted } = useEditorTabs();

    await openTab({ path: "temp/a.log", name: "a.log" });
    await openTab({ path: "src/keep.ts", name: "keep.ts" });

    expect(tabs.value.length).toBe(2);

    handleFileDeleted("temp/a.log");
    expect(tabs.value.length).toBe(1);
    expect(tabs.value[0].path).toBe("src/keep.ts");
  });
});


