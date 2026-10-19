// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import { getLanguageForFile } from "../services/editor/monacoLoader";
import { useToast } from "./useToast";
import { useWorkspaceBus } from "./useWorkspaceBus";

export interface DiffTabMeta {
  originalPath?: string;
  originalContent: string;
  modifiedPath: string;
  modifiedContent: string;
  isStaged?: boolean;
  commitHash?: string;
  parentHash?: string;
}

export interface OpenDiffOptions {
  path: string;
  name?: string;
  originalContent: string;
  modifiedContent: string;
  isStaged?: boolean;
  commitHash?: string;
  parentHash?: string;
}

export interface EditorTab {
  id: string; // File relative path or "diff:..."
  path: string;
  name: string;
  language: string;
  content: string;
  originalContent: string;
  isDirty: boolean;
  isLoading: boolean;
  isBinary?: boolean;
  size?: number;
  binaryReason?: "too_large" | "binary" | "unsupported";
  isPinned?: boolean;
  isDiff?: boolean;
  diffMeta?: DiffTabMeta;
  gitStatus?: string;
  initialLine?: number;
  initialColumn?: number;
}


export const KNOWN_BINARY_EXTENSIONS = new Set([
  "exe", "dll", "so", "dylib", "bin", "elf", "wasm",
  "zip", "tar", "gz", "tgz", "bz2", "xz", "7z", "rar", "iso", "dmg", "pkg", "deb", "rpm",
  "o", "a", "lib", "class", "pyc", "pyo",
  "pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
  "mp3", "mp4", "wav", "ogg", "avi", "mov", "flv", "mkv",
  "db", "sqlite", "sqlite3", "db3", "rdb",
  "woff", "woff2", "ttf", "otf", "eot",
]);

const tabs = ref<EditorTab[]>([]);
const activeTabId = ref<string>("");

export function useEditorTabs() {
  const { showToast } = useToast();

  const activeTab = computed<EditorTab | null>(() => {
    return tabs.value.find((t) => t.id === activeTabId.value) || null;
  });

  const hasDirtyTabs = computed<boolean>(() => {
    return tabs.value.some((t) => t.isDirty);
  });

  async function openTab(file: {
    path: string;
    name: string;
    size?: number;
    line?: number;
    column?: number;
  }): Promise<void> {
    const existing = tabs.value.find((t) => t.id === file.path);
    if (existing) {
      activeTabId.value = existing.id;
      if (file.line !== undefined) existing.initialLine = file.line;
      if (file.column !== undefined) existing.initialColumn = file.column;
      if (!existing.content && !existing.isDirty && !existing.isBinary) {
        try {
          const res = await fetch(`/api/fs/read?path=${encodeURIComponent(file.path)}`);
          if (res.ok) {
            const data = await res.json();
            const content = data.content !== undefined ? data.content : data.data?.content ?? "";
            existing.content = content;
            existing.originalContent = content;
          }
        } catch {}
      }
      return;
    }

    const language = getLanguageForFile(file.name);
    let fileContent = "";
    let isBinary = false;
    let binaryReason: "too_large" | "binary" | "unsupported" | undefined;
    let fileSize = file.size;

    // Check pre-known size or binary extension
    if (fileSize && fileSize > 10 * 1024 * 1024) {
      isBinary = true;
      binaryReason = "too_large";
    } else {
      const ext = file.name.slice(file.name.lastIndexOf(".") + 1).toLowerCase();
      if (file.name.includes(".") && KNOWN_BINARY_EXTENSIONS.has(ext)) {
        isBinary = true;
        binaryReason = "binary";
      }
    }

    // Only read file content if not already known to be binary or too large
    if (!isBinary) {
      try {
        const res = await fetch(`/api/fs/read?path=${encodeURIComponent(file.path)}`);
        if (!res.ok) {
          const errorData = await res.json().catch(() => ({}));
          const errMsg = String(errorData.error || res.statusText || "");
          if (errMsg.toLowerCase().includes("too large") || res.status === 400) {
            isBinary = true;
            binaryReason = errMsg.toLowerCase().includes("too large") ? "too_large" : "binary";
            if (errorData.size) {
              fileSize = errorData.size;
            }
          } else {
            throw new Error(errMsg || `HTTP ${res.status}`);
          }
        } else {
          const data = await res.json();
          fileContent = data.content !== undefined ? data.content : data.data?.content ?? "";
          if (data.size) {
            fileSize = data.size;
          }
          // Detect null bytes in initial chunk (standard binary sniffing)
          if (fileContent.slice(0, 1024).includes("\0")) {
            isBinary = true;
            binaryReason = "binary";
          }
        }
      } catch (err: any) {
        if (!isBinary) {
          fileContent = `// 读取文件失败: ${err.message}`;
          showToast(`读取文件失败: ${err.message}`, "error");
        }
      }
    }

    const newTab: EditorTab = {
      id: file.path,
      path: file.path,
      name: file.name,
      language,
      content: fileContent,
      originalContent: fileContent,
      isDirty: false,
      isLoading: false,
      isBinary,
      size: fileSize,
      binaryReason,
      initialLine: file.line,
      initialColumn: file.column,
    };

    tabs.value.push(newTab);
    activeTabId.value = newTab.id;
  }

  function switchTab(tabId: string): void {
    if (tabs.value.some((t) => t.id === tabId)) {
      activeTabId.value = tabId;
    }
  }

  function updateContent(tabId: string, newContent: string): void {
    const tab = tabs.value.find((t) => t.id === tabId);
    if (tab) {
      tab.content = newContent;
      tab.isDirty = newContent !== tab.originalContent;
    }
  }

  async function saveTab(tabId?: string): Promise<boolean> {
    const targetId = tabId || activeTabId.value;
    const tab = tabs.value.find((t) => t.id === targetId);
    if (!tab) return false;

    try {
      const res = await fetch("/api/fs/write", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ path: tab.path, content: tab.content }),
      });
      const data = await res.json();
      if (res.ok && !data.error && data.success !== false) {
        tab.originalContent = tab.content;
        tab.isDirty = false;
        showToast(`已保存 ${tab.name}`, "success");
        return true;
      } else {
        throw new Error(data.error || "Save failed");
      }
    } catch (err: any) {
      showToast(`保存失败: ${err.message}`, "error");
      return false;
    }
  }

  async function saveAllTabs(): Promise<void> {
    const dirtyTabs = tabs.value.filter((t) => t.isDirty);
    for (const tab of dirtyTabs) {
      await saveTab(tab.id);
    }
  }

  /**
   * 保存指定路径列表中处于未保存（dirty）状态的标签页。
   * 如果 paths 未提供或为空，则保存所有处于 dirty 状态的标签页。
   */
  async function saveDirtyTabsForPaths(paths?: string[]): Promise<void> {
    const pathSet = paths && paths.length > 0 ? new Set(paths) : null;
    const dirtyTabs = tabs.value.filter((t) => t.isDirty && (!pathSet || pathSet.has(t.path)));
    for (const tab of dirtyTabs) {
      await saveTab(tab.id);
    }
  }

  /**
   * 重新拉取指定路径列表在磁盘上的最新内容，用于 Git Discard 或外部文件变更时同步。
   */
  async function reloadTabsForPaths(paths: string[]): Promise<void> {
    if (!paths || paths.length === 0) return;
    const pathSet = new Set(paths);
    const targetTabs = tabs.value.filter((t) => !t.isDiff && pathSet.has(t.path));
    for (const tab of targetTabs) {
      try {
        const res = await fetch(`/api/fs/read?path=${encodeURIComponent(tab.path)}`);
        if (res.ok) {
          const data = await res.json();
          const content = data.content !== undefined ? data.content : data.data?.content ?? "";
          tab.content = content;
          tab.originalContent = content;
          tab.isDirty = false;
        }
      } catch (err) {
        console.warn(`[useEditorTabs] Failed to reload tab ${tab.path}:`, err);
      }
    }
  }

  /**
   * 处理文件或目录重命名，平滑迁移已打开标签页的路径和元数据
   */
  function handleFileRenamed({ oldPath, newPath }: { oldPath: string; newPath: string }): void {
    const isDirRename = oldPath.endsWith("/") || tabs.value.some((t) => t.path.startsWith(`${oldPath}/`));
    tabs.value.forEach((tab) => {
      if (tab.isDiff) return;

      if (tab.path === oldPath) {
        // 单文件重命名
        const newName = newPath.split("/").pop() || newPath;
        tab.path = newPath;
        tab.name = newName;
        tab.language = getLanguageForFile(newName);
        if (tab.id === oldPath) {
          tab.id = newPath;
          if (activeTabId.value === oldPath) {
            activeTabId.value = newPath;
          }
        }
      } else if (isDirRename && tab.path.startsWith(`${oldPath}/`)) {
        // 目录重命名带来的子文件路径迁移
        const relativeSuffix = tab.path.slice(oldPath.length);
        const updatedPath = `${newPath}${relativeSuffix}`;
        tab.path = updatedPath;
        if (tab.id.startsWith(oldPath)) {
          const oldId = tab.id;
          tab.id = `${newPath}${relativeSuffix}`;
          if (activeTabId.value === oldId) {
            activeTabId.value = tab.id;
          }
        }
      }
    });
  }

  /**
   * 处理文件或目录删除，自动关闭相关标签页
   */
  function handleFileDeleted(deletedPath: string): void {
    const toClose: string[] = [];
    tabs.value.forEach((tab) => {
      if (tab.path === deletedPath || tab.path.startsWith(`${deletedPath}/`)) {
        toClose.push(tab.id);
      }
    });
    toClose.forEach((id) => closeTab(id));
  }

  // 监听工作区总线触发的文件重载通知（如 Git Discard、外部变更或分支切换）
  const workspaceBus = useWorkspaceBus();
  workspaceBus.onFileReloadNeeded((paths) => {
    reloadTabsForPaths(paths);
  });

  workspaceBus.onFileRenamed((evt) => {
    handleFileRenamed(evt);
  });

  workspaceBus.onFileDeleted((path) => {
    handleFileDeleted(path);
  });

  function closeTab(tabId: string): void {
    const index = tabs.value.findIndex((t) => t.id === tabId);
    if (index === -1) return;

    if (activeTabId.value === tabId) {
      if (tabs.value.length > 1) {
        const nextActive = index === tabs.value.length - 1 ? tabs.value[index - 1] : tabs.value[index + 1];
        activeTabId.value = nextActive.id;
      } else {
        activeTabId.value = "";
      }
    }

    tabs.value.splice(index, 1);
  }

  function closeOtherTabs(tabId: string): void {
    tabs.value = tabs.value.filter((t) => t.id === tabId || t.isPinned);
    activeTabId.value = tabId;
  }

  function closeTabsToRight(tabId: string): void {
    const index = tabs.value.findIndex((t) => t.id === tabId);
    if (index === -1) return;

    const activeIndex = tabs.value.findIndex((t) => t.id === activeTabId.value);
    if (activeIndex > index) {
      activeTabId.value = tabId;
    }

    tabs.value = tabs.value.filter((tab, idx) => idx <= index || tab.isPinned);
  }

  function closeSavedTabs(): void {
    const retained = tabs.value.filter((t) => t.isDirty || t.isPinned);
    tabs.value = retained;
    if (!retained.some((t) => t.id === activeTabId.value)) {
      activeTabId.value = retained[0]?.id || "";
    }
  }

  function closeAllTabs(forceAll = false): void {
    if (forceAll) {
      tabs.value = [];
      activeTabId.value = "";
    } else {
      const pinned = tabs.value.filter((t) => t.isPinned);
      tabs.value = pinned;
      if (!pinned.some((t) => t.id === activeTabId.value)) {
        activeTabId.value = pinned[0]?.id || "";
      }
    }
  }

  function togglePinTab(tabId: string): void {
    const tab = tabs.value.find((t) => t.id === tabId);
    if (tab) {
      tab.isPinned = !tab.isPinned;
      tabs.value.sort((a, b) => {
        if (a.isPinned && !b.isPinned) return -1;
        if (!a.isPinned && b.isPinned) return 1;
        return 0;
      });
      showToast(tab.isPinned ? `已固定标签页: ${tab.name}` : `已取消固定: ${tab.name}`, "info");
    }
  }

  function openDiffTab(options: OpenDiffOptions): void {
    const diffId = options.commitHash
      ? `diff:commit:${options.commitHash.slice(0, 7)}:${options.path}`
      : options.isStaged
      ? `diff:staged:${options.path}`
      : `diff:working:${options.path}`;

    const existing = tabs.value.find((t) => t.id === diffId);
    if (existing) {
      if (existing.diffMeta) {
        existing.diffMeta.originalContent = options.originalContent;
        existing.diffMeta.modifiedContent = options.modifiedContent;
      }
      activeTabId.value = existing.id;
      return;
    }

    const fileName = options.path.split("/").pop() || options.path;
    const displayName = options.name
      ? options.name
      : options.commitHash
      ? `${fileName} (${options.commitHash.slice(0, 7)})`
      : options.isStaged
      ? `${fileName} (Staged)`
      : `${fileName} (Working Tree)`;

    const language = getLanguageForFile(fileName);

    const diffTab: EditorTab = {
      id: diffId,
      path: options.path,
      name: displayName,
      language,
      content: options.modifiedContent,
      originalContent: options.originalContent,
      isDirty: false,
      isLoading: false,
      isDiff: true,
      diffMeta: {
        originalPath: options.path,
        originalContent: options.originalContent,
        modifiedPath: options.path,
        modifiedContent: options.modifiedContent,
        isStaged: options.isStaged,
        commitHash: options.commitHash,
        parentHash: options.parentHash,
      },
      gitStatus: options.isStaged ? "S" : "M",
    };

    tabs.value.push(diffTab);
    activeTabId.value = diffTab.id;
  }

  function setTabGitStatus(path: string, status: string): void {
    tabs.value.forEach((t) => {
      if (t.path === path && !t.isDiff) {
        t.gitStatus = status;
      }
    });
  }

  return {
    tabs,
    activeTabId,
    activeTab,
    hasDirtyTabs,
    openTab,
    openDiffTab,
    setTabGitStatus,
    switchTab,
    updateContent,
    saveTab,
    saveAllTabs,
    saveDirtyTabsForPaths,
    reloadTabsForPaths,
    handleFileRenamed,
    handleFileDeleted,
    closeTab,
    closeOtherTabs,
    closeTabsToRight,
    closeSavedTabs,
    closeAllTabs,
    togglePinTab,
  };
}

