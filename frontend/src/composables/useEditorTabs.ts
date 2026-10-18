// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import { getLanguageForFile } from "../services/editor/monacoLoader";
import { useToast } from "./useToast";

export interface EditorTab {
  id: string; // File relative path
  path: string;
  name: string;
  language: string;
  content: string;
  originalContent: string;
  isDirty: boolean;
  isLoading: boolean;
}

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

  async function openTab(file: { path: string; name: string }): Promise<void> {
    const existing = tabs.value.find((t) => t.id === file.path);
    if (existing) {
      activeTabId.value = existing.id;
      return;
    }

    const language = getLanguageForFile(file.name);
    const newTab: EditorTab = {
      id: file.path,
      path: file.path,
      name: file.name,
      language,
      content: "",
      originalContent: "",
      isDirty: false,
      isLoading: true,
    };

    tabs.value.push(newTab);
    activeTabId.value = newTab.id;

    try {
      const res = await fetch(`/api/fs/read?path=${encodeURIComponent(file.path)}`);
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}: ${res.statusText}`);
      }
      const data = await res.json();
      const content = data.content !== undefined ? data.content : data.data?.content ?? "";
      newTab.content = content;
      newTab.originalContent = content;
      newTab.isDirty = false;
    } catch (err: any) {
      newTab.content = `// Error loading file: ${err.message}`;
      newTab.originalContent = newTab.content;
      showToast(`读取文件失败: ${err.message}`, "error");
    } finally {
      newTab.isLoading = false;
    }
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
    tabs.value = tabs.value.filter((t) => t.id === tabId);
    activeTabId.value = tabId;
  }

  function closeAllTabs(): void {
    tabs.value = [];
    activeTabId.value = "";
  }

  return {
    tabs,
    activeTabId,
    activeTab,
    hasDirtyTabs,
    openTab,
    switchTab,
    updateContent,
    saveTab,
    saveAllTabs,
    closeTab,
    closeOtherTabs,
    closeAllTabs,
  };
}
