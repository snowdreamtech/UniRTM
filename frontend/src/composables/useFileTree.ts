// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import { useToast } from "./useToast";
import { useEditorTabs } from "./useEditorTabs";
import { useWorkspaceBus } from "./useWorkspaceBus";
import { useGit } from "./useGit";
import { t } from "../i18n";

export interface FSNode {
  name: string;
  path: string;
  is_dir: boolean;
  size: number;
  mod_time: string;
  is_ignored?: boolean;
  children?: FSNode[];
}

export interface InlineAction {
  type: "new-file" | "new-folder" | "rename";
  parentPath: string;
  targetPath?: string;
  defaultValue?: string;
}

export interface ContextMenuState {
  visible: boolean;
  x: number;
  y: number;
  node: FSNode | null;
}

const tree = ref<FSNode | null>(null);
const isLoading = ref<boolean>(false);
const error = ref<string | null>(null);
const selectedNode = ref<FSNode | null>(null);
const expandedPaths = ref<Set<string>>(new Set([""]));
const searchQuery = ref<string>("");
const inlineAction = ref<InlineAction | null>(null);
const contextMenu = ref<ContextMenuState>({
  visible: false,
  x: 0,
  y: 0,
  node: null,
});
const deleteConfirmModal = ref<{ visible: boolean; node: FSNode | null }>({
  visible: false,
  node: null,
});

const activeFileContent = ref<string>("");
const isFileLoading = ref<boolean>(false);

function normalizeNode(raw: any): FSNode {
  const rawChildren = raw.entries || raw.Entries || raw.children || raw.Children;
  return {
    name: raw.name || raw.Name || "",
    path: raw.path || raw.Path || "",
    is_dir: Boolean(
      raw.is_dir !== undefined
        ? raw.is_dir
        : raw.IsDir !== undefined
        ? raw.IsDir
        : Array.isArray(rawChildren) && rawChildren.length > 0
    ),
    size: raw.size || raw.Size || 0,
    mod_time: raw.mod_time || raw.ModTime || "",
    is_ignored: Boolean(raw.is_ignored || raw.IsIgnored),
    children: Array.isArray(rawChildren) ? rawChildren.map(normalizeNode) : undefined,
  };
}

export function useFileTree() {
  const { showToast } = useToast();

  async function fetchTree(): Promise<void> {
    isLoading.value = true;
    error.value = null;
    try {
      const res = await fetch("/api/fs/tree");
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}: ${res.statusText}`);
      }
      const data = await res.json();
      if (data && (data.entries || data.Entries || data.children || data.Children || data.data || data.name || data.Name)) {
        const rootRaw = data.data || data;
        const rootChildren = rootRaw.entries || rootRaw.Entries || rootRaw.children || rootRaw.Children || [];
        tree.value = {
          name: rootRaw.name || rootRaw.Name || "UNIRTM",
          path: rootRaw.path || rootRaw.Path || "",
          is_dir: true,
          size: 0,
          mod_time: new Date().toISOString(),
          is_ignored: false,
          children: rootChildren.map(normalizeNode),
        };
        // Keep root path expanded by default
        if (!expandedPaths.value.has("")) {
          expandedPaths.value.add("");
        }
      } else {
        throw new Error(data.error || "Failed to load directory tree");
      }
    } catch (err: any) {
      error.value = err.message || "Failed to load directory tree";
      console.error("[useFileTree] fetchTree failed:", err);
    } finally {
      isLoading.value = false;
    }
  }

  function toggleExpand(path: string): void {
    const next = new Set(expandedPaths.value);
    if (next.has(path)) {
      next.delete(path);
    } else {
      next.add(path);
    }
    expandedPaths.value = next;
  }

  function expandPath(path: string): void {
    const next = new Set(expandedPaths.value);
    next.add(path);
    // Also expand parent directories
    const parts = path.split("/").filter(Boolean);
    let current = "";
    for (let i = 0; i < parts.length - 1; i++) {
      current = current ? `${current}/${parts[i]}` : parts[i];
      next.add(current);
    }
    expandedPaths.value = next;
  }

  function revealPath(path: string): FSNode | null {

    if (!path) return null;
    expandPath(path);

    const findNode = (curr: FSNode): FSNode | null => {
      if (curr.path === path) return curr;
      if (curr.children) {
        for (const child of curr.children) {
          const res = findNode(child);
          if (res) return res;
        }
      }
      return null;
    };

    if (tree.value) {
      const target = findNode(tree.value);
      if (target) {
        selectedNode.value = target;
      }
      return target;
    }
    return null;
  }

  function collapseAll(): void {

    // Keep only root folder expanded so user sees the workspace container
    expandedPaths.value = new Set([""]);
  }

  async function selectNode(node: FSNode): Promise<void> {
    selectedNode.value = node;
    if (node.is_dir) {
      toggleExpand(node.path);
    } else {
      const { openTab } = useEditorTabs();
      await openTab({ path: node.path, name: node.name, size: node.size });
    }
  }

  async function loadFileContent(path: string): Promise<void> {
    isFileLoading.value = true;
    try {
      const res = await fetch(`/api/fs/read?path=${encodeURIComponent(path)}`);
      if (!res.ok) {
        if (res.status === 400) {
          // Binary or oversized file: silently skip without error toast
          return;
        }
        throw new Error(`HTTP ${res.status}: ${res.statusText}`);
      }
      const data = await res.json();
      const content = data.content !== undefined ? data.content : data.data?.content;
      if (content !== undefined) {
        activeFileContent.value = content;
      } else {
        throw new Error(data.error || "Failed to read file");
      }
    } catch (err: any) {
      activeFileContent.value = `// Error loading file: ${err.message}`;
    } finally {
      isFileLoading.value = false;
    }
  }

  async function createFile(parentPath: string, name: string): Promise<boolean> {
    const cleanName = name.trim();
    if (!cleanName) return false;
    const targetPath = parentPath ? `${parentPath}/${cleanName}` : cleanName;

    try {
      const res = await fetch("/api/fs/create", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ path: targetPath, is_dir: false }),
      });
      const data = await res.json();
      if (res.ok && !data.error && data.success !== false) {
        showToast(t("explorer.created") || "文件创建成功", "success");
        inlineAction.value = null;
        expandPath(parentPath);
        await fetchTree();
        return true;
      } else {
        throw new Error(data.error || "Create failed");
      }
    } catch (err: any) {
      showToast(err.message, "error");
      return false;
    }
  }

  async function createFolder(parentPath: string, name: string): Promise<boolean> {
    const cleanName = name.trim();
    if (!cleanName) return false;
    const targetPath = parentPath ? `${parentPath}/${cleanName}` : cleanName;

    try {
      const res = await fetch("/api/fs/create", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ path: targetPath, is_dir: true }),
      });
      const data = await res.json();
      if (res.ok && !data.error && data.success !== false) {
        showToast(t("explorer.created") || "文件夹创建成功", "success");
        inlineAction.value = null;
        expandPath(parentPath);
        await fetchTree();
        return true;
      } else {
        throw new Error(data.error || "Create folder failed");
      }
    } catch (err: any) {
      showToast(err.message, "error");
      return false;
    }
  }

  async function renameItem(oldPath: string, newName: string): Promise<boolean> {
    const cleanName = newName.trim();
    if (!cleanName) return false;

    const parts = oldPath.split("/").filter(Boolean);
    parts.pop();
    const parentPath = parts.join("/");
    const newPath = parentPath ? `${parentPath}/${cleanName}` : cleanName;

    if (newPath === oldPath) {
      inlineAction.value = null;
      return true;
    }

    try {
      const res = await fetch("/api/fs/rename", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ old_path: oldPath, new_path: newPath }),
      });
      const data = await res.json();
      if (res.ok && !data.error && data.success !== false) {
        showToast(t("explorer.renamed") || "重命名成功", "success");
        inlineAction.value = null;
        if (selectedNode.value?.path === oldPath) {
          selectedNode.value.path = newPath;
          selectedNode.value.name = cleanName;
        }

        const bus = useWorkspaceBus();
        bus.notifyFileRenamed({ oldPath, newPath });
        bus.notifyGitStatusChanged();

        await fetchTree();
        return true;
      } else {
        throw new Error(data.error || "Rename failed");
      }
    } catch (err: any) {
      showToast(err.message, "error");
      return false;
    }
  }

  async function deleteItem(path: string): Promise<boolean> {
    try {
      const res = await fetch("/api/fs/delete", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ path }),
      });
      const data = await res.json();
      if (res.ok && !data.error && data.success !== false) {
        showToast(t("explorer.deleted") || "删除成功", "success");
        if (selectedNode.value?.path === path || selectedNode.value?.path.startsWith(`${path}/`)) {
          selectedNode.value = null;
          activeFileContent.value = "";
        }
        deleteConfirmModal.value = { visible: false, node: null };

        const bus = useWorkspaceBus();
        bus.notifyFileDeleted(path);
        bus.notifyGitStatusChanged();

        await fetchTree();
        return true;
      } else {
        throw new Error(data.error || "Delete failed");
      }
    } catch (err: any) {
      showToast(err.message, "error");
      return false;
    }
  }

  async function saveFile(path: string, content: string): Promise<boolean> {
    try {
      const res = await fetch("/api/fs/write", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ path, content }),
      });
      const data = await res.json();
      if (res.ok && !data.error && data.success !== false) {
        showToast("文件已保存", "success");
        activeFileContent.value = content;
        return true;
      } else {
        throw new Error(data.error || "Save failed");
      }
    } catch (err: any) {
      showToast(err.message, "error");
      return false;
    }
  }

  async function copyRelativePath(path: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(path);
      showToast("已复制相对路径", "info");
    } catch {
      showToast("复制失败", "error");
    }
  }

  async function copyFileName(name: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(name);
      showToast("已复制文件名", "info");
    } catch {
      showToast("复制失败", "error");
    }
  }

  function startInlineCreate(parentPath: string, type: "new-file" | "new-folder"): void {
    expandPath(parentPath);
    inlineAction.value = {
      type,
      parentPath,
      defaultValue: "",
    };
  }

  function startInlineRename(node: FSNode): void {
    inlineAction.value = {
      type: "rename",
      parentPath: "",
      targetPath: node.path,
      defaultValue: node.name,
    };
  }

  function cancelInlineAction(): void {
    inlineAction.value = null;
  }

  function openContextMenu(e: MouseEvent, node: FSNode): void {
    e.preventDefault();
    e.stopPropagation();
    selectedNode.value = node;
    contextMenu.value = {
      visible: true,
      x: e.clientX,
      y: e.clientY,
      node,
    };
  }

  function closeContextMenu(): void {
    contextMenu.value.visible = false;
  }

  function promptDelete(node: FSNode): void {
    deleteConfirmModal.value = {
      visible: true,
      node,
    };
  }

  // Filter helper
  function filterNode(node: FSNode, query: string): FSNode | null {
    const q = query.toLowerCase();
    const isSelfMatch = node.name.toLowerCase().includes(q);

    if (node.is_dir && node.children) {
      const filteredChildren: FSNode[] = [];
      for (const child of node.children) {
        const filtered = filterNode(child, query);
        if (filtered) {
          filteredChildren.push(filtered);
        }
      }
      if (filteredChildren.length > 0 || isSelfMatch) {
        return {
          ...node,
          children: filteredChildren,
        };
      }
      return null;
    }

    return isSelfMatch ? node : null;
  }

  const isGitFilterActive = ref<boolean>(false);

  function getChangedPathsSet(): Set<string> {
    const { status } = useGit();
    const set = new Set<string>();
    status.value.staged.forEach((item) => set.add(item.path));
    status.value.unstaged.forEach((item) => set.add(item.path));
    return set;
  }

  function filterGitNode(node: FSNode, changedPaths: Set<string>): FSNode | null {
    const isSelfChanged = changedPaths.has(node.path);

    if (node.is_dir && node.children) {
      const filteredChildren: FSNode[] = [];
      for (const child of node.children) {
        const filtered = filterGitNode(child, changedPaths);
        if (filtered) {
          filteredChildren.push(filtered);
        }
      }
      if (filteredChildren.length > 0 || isSelfChanged) {
        return {
          ...node,
          children: filteredChildren,
        };
      }
      return null;
    }

    return isSelfChanged ? node : null;
  }

  function toggleGitFilter(): void {
    isGitFilterActive.value = !isGitFilterActive.value;
    if (isGitFilterActive.value && tree.value) {
      // 开启 Git 聚焦模式时，自动展开含有变动的目录分支
      const gitPaths = getChangedPathsSet();
      gitPaths.forEach((p) => {
        expandPath(p);
      });
    }
  }

  const filteredTree = computed(() => {
    if (!tree.value) return null;
    let base = tree.value;

    if (isGitFilterActive.value) {
      const gitPaths = getChangedPathsSet();
      const gitFiltered = filterGitNode(base, gitPaths);
      if (!gitFiltered) {
        // 无任何改动时返回空子节点的根，避免树结构崩溃
        return {
          ...base,
          children: [],
        };
      }
      base = gitFiltered;
    }

    const query = searchQuery.value.trim();
    if (!query) return base;
    return filterNode(base, query);
  });

  function getNodeGitChangeCount(node: FSNode): number {
    if (!node.is_dir) {
      const gitPaths = getChangedPathsSet();
      return gitPaths.has(node.path) ? 1 : 0;
    }
    const { getDirGitCount } = useGit();
    return getDirGitCount(node.path);
  }

  return {
    tree,
    filteredTree,
    isGitFilterActive,
    toggleGitFilter,
    getNodeGitChangeCount,
    isLoading,
    error,
    selectedNode,
    expandedPaths,
    searchQuery,
    inlineAction,
    contextMenu,
    deleteConfirmModal,
    activeFileContent,
    isFileLoading,
    fetchTree,
    toggleExpand,
    expandPath,
    revealPath,
    collapseAll,

    selectNode,
    loadFileContent,
    createFile,
    createFolder,
    renameItem,
    deleteItem,
    saveFile,
    copyRelativePath,
    copyFileName,
    startInlineCreate,
    startInlineRename,
    cancelInlineAction,
    openContextMenu,
    closeContextMenu,
    promptDelete,
  };
}
