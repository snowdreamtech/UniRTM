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

export interface FileRenameEvent {
  oldPath: string;
  newPath: string;
  isDir?: boolean;
}

export interface OpenInEditorRequest {
  path: string;
  line?: number;
  column?: number;
  timestamp: number;
}

export interface SendToTerminalRequest {
  text: string;
  addNewline?: boolean;
}

export interface RevealInTerminalRequest {
  directoryPath: string;
}

export interface ToolActionRequest {
  action: "switch" | "install" | "toggle-pin" | "delete";
  toolId: string;
  toolName?: string;
  tool?: any;
}

export interface TaskActionRequest {
  action: "run" | "edit" | "delete";
  taskName: string;
  command?: string;
  taskCmd?: string;
  task?: any;
}

type RevealListener = (req: RevealRequest) => void;
type GitStatusChangeListener = () => void;
type OpenDiffListener = (req: WorkspaceDiffRequest) => void;
type FileReloadListener = (paths: string[]) => void;
type FileRenameListener = (evt: FileRenameEvent) => void;
type FileDeleteListener = (path: string) => void;
type OpenInEditorListener = (req: OpenInEditorRequest) => void;
type SendToTerminalListener = (req: SendToTerminalRequest) => void;
type RevealInTerminalListener = (req: RevealInTerminalRequest) => void;
type ToolActionListener = (req: ToolActionRequest) => void;
type ToolsChangeListener = () => void;
type TaskActionListener = (req: TaskActionRequest) => void;
type TasksChangeListener = () => void;
type DoctorChangeListener = () => void;

const lastRevealRequest = ref<RevealRequest | null>(null);
const currentActiveFilePath = ref<string>("");
const revealListeners = new Set<RevealListener>();
const gitStatusChangeListeners = new Set<GitStatusChangeListener>();
const openDiffListeners = new Set<OpenDiffListener>();
const fileReloadListeners = new Set<FileReloadListener>();
const fileRenameListeners = new Set<FileRenameListener>();
const fileDeleteListeners = new Set<FileDeleteListener>();
const openInEditorListeners = new Set<OpenInEditorListener>();
const sendToTerminalListeners = new Set<SendToTerminalListener>();
const revealInTerminalListeners = new Set<RevealInTerminalListener>();
const toolActionListeners = new Set<ToolActionListener>();
const toolsChangeListeners = new Set<ToolsChangeListener>();
const taskActionListeners = new Set<TaskActionListener>();
const tasksChangeListeners = new Set<TasksChangeListener>();
const doctorChangeListeners = new Set<DoctorChangeListener>();

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
   * 广播文件或目录重命名事件
   */
  function notifyFileRenamed(evt: FileRenameEvent): void {
    if (!evt || !evt.oldPath || !evt.newPath) return;
    fileRenameListeners.forEach((fn) => {
      try {
        fn(evt);
      } catch (err) {
        console.error("[WorkspaceBus] Error in file renamed listener:", err);
      }
    });
  }

  /**
   * 注册文件或目录重命名监听器
   */
  function onFileRenamed(listener: FileRenameListener): () => void {
    fileRenameListeners.add(listener);
    return () => {
      fileRenameListeners.delete(listener);
    };
  }

  /**
   * 广播文件或目录删除事件
   */
  function notifyFileDeleted(path: string): void {
    if (!path) return;
    fileDeleteListeners.forEach((fn) => {
      try {
        fn(path);
      } catch (err) {
        console.error("[WorkspaceBus] Error in file deleted listener:", err);
      }
    });
  }

  /**
   * 注册文件或目录删除监听器
   */
  function onFileDeleted(listener: FileDeleteListener): () => void {
    fileDeleteListeners.add(listener);
    return () => {
      fileDeleteListeners.delete(listener);
    };
  }

  /**
   * 记录当前编辑器聚焦的活跃文件
   */
  function setActiveFilePath(path: string): void {
    currentActiveFilePath.value = path;
  }

  /**
   * 在主编辑器中打开指定文件并定位到特定行/列
   */
  function openInEditor(req: { path: string; line?: number; column?: number }): void {
    if (!req || !req.path) return;
    const fullReq: OpenInEditorRequest = {
      path: req.path,
      line: req.line,
      column: req.column,
      timestamp: Date.now(),
    };
    openInEditorListeners.forEach((fn) => {
      try {
        fn(fullReq);
      } catch (err) {
        console.error("[WorkspaceBus] Error in openInEditor listener:", err);
      }
    });
  }

  /**
   * 注册在主编辑器中打开文件的监听器
   */
  function onOpenInEditor(listener: OpenInEditorListener): () => void {
    openInEditorListeners.add(listener);
    return () => {
      openInEditorListeners.delete(listener);
    };
  }

  /**
   * 发送指令或代码到当前活动终端执行
   */
  function sendToTerminal(req: SendToTerminalRequest): void {
    if (!req || req.text === undefined) return;
    sendToTerminalListeners.forEach((fn) => {
      try {
        fn(req);
      } catch (err) {
        console.error("[WorkspaceBus] Error in sendToTerminal listener:", err);
      }
    });
  }

  /**
   * 注册发送至终端的监听器
   */
  function onSendToTerminal(listener: SendToTerminalListener): () => void {
    sendToTerminalListeners.add(listener);
    return () => {
      sendToTerminalListeners.delete(listener);
    };
  }

  /**
   * 在集成终端中打开指定工作目录
   */
  function revealInTerminal(directoryPath: string): void {
    if (!directoryPath) return;
    const req: RevealInTerminalRequest = { directoryPath };
    revealInTerminalListeners.forEach((fn) => {
      try {
        fn(req);
      } catch (err) {
        console.error("[WorkspaceBus] Error in revealInTerminal listener:", err);
      }
    });
  }

  /**
   * 注册在集成终端中打开目录的监听器
   */
  function onRevealInTerminal(listener: RevealInTerminalListener): () => void {
    revealInTerminalListeners.add(listener);
    return () => {
      revealInTerminalListeners.delete(listener);
    };
  }

  /**
   * 触发工具链操作（如从侧边栏右键菜单请求主画布执行切换、安装、锁定、删除）
   */
  function triggerToolAction(req: ToolActionRequest): void {
    if (!req || !req.toolId) return;
    toolActionListeners.forEach((fn) => {
      try {
        fn(req);
      } catch (err) {
        console.error("[WorkspaceBus] Error in tool action listener:", err);
      }
    });
  }

  /**
   * 注册工具链操作监听器
   */
  function onToolAction(listener: ToolActionListener): () => void {
    toolActionListeners.add(listener);
    return () => {
      toolActionListeners.delete(listener);
    };
  }

  /**
   * 广播工具链配置或状态发生变动
   */
  function notifyToolsChanged(): void {
    toolsChangeListeners.forEach((fn) => {
      try {
        fn();
      } catch (err) {
        console.error("[WorkspaceBus] Error in tools change listener:", err);
      }
    });
  }

  /**
   * 注册工具链变动监听器
   */
  function onToolsChanged(listener: ToolsChangeListener): () => void {
    toolsChangeListeners.add(listener);
    return () => {
      toolsChangeListeners.delete(listener);
    };
  }

  /**
   * 触发工程自动化任务操作（如从侧边栏右键菜单请求主画布执行运行、编辑、复制、删除）
   */
  function triggerTaskAction(req: TaskActionRequest): void {
    if (!req || !req.taskName) return;
    taskActionListeners.forEach((fn) => {
      try {
        fn(req);
      } catch (err) {
        console.error("[WorkspaceBus] Error in task action listener:", err);
      }
    });
  }

  /**
   * 注册工程自动化任务操作监听器
   */
  function onTaskAction(listener: TaskActionListener): () => void {
    taskActionListeners.add(listener);
    return () => {
      taskActionListeners.delete(listener);
    };
  }

  /**
   * 广播任务列表发生变动
   */
  function notifyTasksChanged(): void {
    tasksChangeListeners.forEach((fn) => {
      try {
        fn();
      } catch (err) {
        console.error("[WorkspaceBus] Error in tasks change listener:", err);
      }
    });
  }

  /**
   * 注册任务列表变动监听器
   */
  function onTasksChanged(listener: TasksChangeListener): () => void {
    tasksChangeListeners.add(listener);
    return () => {
      tasksChangeListeners.delete(listener);
    };
  }

  /**
   * 广播系统环境诊断变动
   */
  function notifyDoctorChanged(): void {
    doctorChangeListeners.forEach((fn) => {
      try {
        fn();
      } catch (err) {
        console.error("[WorkspaceBus] Error in doctor change listener:", err);
      }
    });
  }

  /**
   * 注册系统环境诊断变动监听器
   */
  function onDoctorChanged(listener: DoctorChangeListener): () => void {
    doctorChangeListeners.add(listener);
    return () => {
      doctorChangeListeners.delete(listener);
    };
  }

  /**
   * 清除所有监听器（主要用于单元测试隔离）
   */
  function _clearAllListeners(): void {
    revealListeners.clear();
    gitStatusChangeListeners.clear();
    openDiffListeners.clear();
    fileReloadListeners.clear();
    fileRenameListeners.clear();
    fileDeleteListeners.clear();
    openInEditorListeners.clear();
    sendToTerminalListeners.clear();
    revealInTerminalListeners.clear();
    toolActionListeners.clear();
    toolsChangeListeners.clear();
    taskActionListeners.clear();
    tasksChangeListeners.clear();
    doctorChangeListeners.clear();
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
    notifyFileRenamed,
    onFileRenamed,
    notifyFileDeleted,
    onFileDeleted,
    setActiveFilePath,
    openInEditor,
    onOpenInEditor,
    sendToTerminal,
    onSendToTerminal,
    revealInTerminal,
    onRevealInTerminal,
    triggerToolAction,
    onToolAction,
    notifyToolsChanged,
    onToolsChanged,
    triggerTaskAction,
    onTaskAction,
    notifyTasksChanged,
    onTasksChanged,
    notifyDoctorChanged,
    onDoctorChanged,
    _clearAllListeners,
  };
}
