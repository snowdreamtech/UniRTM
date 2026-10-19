// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref } from "vue";

export interface RevealRequest {
  path: string;
  timestamp: number;
}

export interface WorkspaceDiffRequest {
  path: string;
  name?: string;
  originalContent: string;
  modifiedContent: string;
  isStaged?: boolean;
  commitHash?: string;
  parentHash?: string;
}

type RevealListener = (req: RevealRequest) => void;
type GitStatusChangeListener = () => void;
type OpenDiffListener = (req: WorkspaceDiffRequest) => void;
type FileReloadListener = (paths: string[]) => void;

const lastRevealRequest = ref<RevealRequest | null>(null);
const currentActiveFilePath = ref<string>("");
const revealListeners = new Set<RevealListener>();
const gitStatusChangeListeners = new Set<GitStatusChangeListener>();
const openDiffListeners = new Set<OpenDiffListener>();
const fileReloadListeners = new Set<FileReloadListener>();

/**
 * useWorkspaceBus: 轻量级跨组件协同状态总线
 * 负责文件管理器 (Explorer)、主编辑器 (Editor) 与 源代码管理 (SCM) 的双向定位与行为派发
 */
export function useWorkspaceBus() {
  /**
   * 在文件管理器中揭示并高亮指定文件
   */
  function revealInExplorer(path: string): void {
    if (!path) return;
    const req: RevealRequest = { path, timestamp: Date.now() };
    lastRevealRequest.value = req;
    revealListeners.forEach((fn) => {
      try {
        fn(req);
      } catch (err) {
        console.error("[WorkspaceBus] Error in reveal listener:", err);
      }
    });
  }

  /**
   * 注册文件揭示监听器
   */
  function onRevealFile(listener: RevealListener): () => void {
    revealListeners.add(listener);
    return () => {
      revealListeners.delete(listener);
    };
  }

  /**
   * 广播 Git 状态发生变动，驱动 Explorer 树和 SCM 列表协同刷新
   */
  function notifyGitStatusChanged(): void {
    gitStatusChangeListeners.forEach((fn) => {
      try {
        fn();
      } catch (err) {
        console.error("[WorkspaceBus] Error in git status listener:", err);
      }
    });
  }

  /**
   * 注册 Git 状态变动监听器
   */
  function onGitStatusChanged(listener: GitStatusChangeListener): () => void {
    gitStatusChangeListeners.add(listener);
    return () => {
      gitStatusChangeListeners.delete(listener);
    };
  }

  /**
   * 请求在主编辑器中打开差异比对标签页
   */
  function openDiff(req: WorkspaceDiffRequest): void {
    openDiffListeners.forEach((fn) => {
      try {
        fn(req);
      } catch (err) {
        console.error("[WorkspaceBus] Error in open diff listener:", err);
      }
    });
  }

  /**
   * 注册差异比对打开监听器
   */
  function onOpenDiff(listener: OpenDiffListener): () => void {
    openDiffListeners.add(listener);
    return () => {
      openDiffListeners.delete(listener);
    };
  }

  /**
   * 通知指定路径的文件需要从磁盘重新加载最新内容（如执行了 Git Discard、外部变更或 Checkout）
   */
  function notifyFileReloadNeeded(paths: string[]): void {
    if (!paths || paths.length === 0) return;
    fileReloadListeners.forEach((fn) => {
      try {
        fn(paths);
      } catch (err) {
        console.error("[WorkspaceBus] Error in file reload listener:", err);
      }
    });
  }

  /**
   * 注册文件重新加载监听器
   */
  function onFileReloadNeeded(listener: FileReloadListener): () => void {
    fileReloadListeners.add(listener);
    return () => {
      fileReloadListeners.delete(listener);
    };
  }

  /**
   * 记录当前编辑器聚焦的活跃文件
   */
  function setActiveFilePath(path: string): void {
    currentActiveFilePath.value = path;
  }

  /**
   * 清除所有监听器（主要用于单元测试隔离）
   */
  function _clearAllListeners(): void {
    revealListeners.clear();
    gitStatusChangeListeners.clear();
    openDiffListeners.clear();
    fileReloadListeners.clear();
    lastRevealRequest.value = null;
    currentActiveFilePath.value = "";
  }

  return {
    lastRevealRequest,
    currentActiveFilePath,
    revealInExplorer,
    onRevealFile,
    notifyGitStatusChanged,
    onGitStatusChanged,
    openDiff,
    onOpenDiff,
    notifyFileReloadNeeded,
    onFileReloadNeeded,
    setActiveFilePath,
    _clearAllListeners,
  };
}
