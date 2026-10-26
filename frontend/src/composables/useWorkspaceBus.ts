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

export interface EnvActionRequest {
  action: "edit" | "copy-value" | "copy-name" | "restore-base" | "delete";
  envName: string;
  env?: any;
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
type DoctorChangeListener = (report?: any) => void;
type EnvActionListener = (req: EnvActionRequest) => void;
type EnvsChangeListener = () => void;

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
const envActionListeners = new Set<EnvActionListener>();
const envsChangeListeners = new Set<EnvsChangeListener>();

/**
 * useWorkspaceBus: Lightweight cross-component coordination bus
 * Responsible for bidirectional navigation and dispatching between Explorer, Editor, and SCM
 */
export function useWorkspaceBus() {
  /**
   * Reveal and highlight specified file in Explorer
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
   * Register file reveal listener
   */
  function onRevealFile(listener: RevealListener): () => void {
    revealListeners.add(listener);
    return () => {
      revealListeners.delete(listener);
    };
  }

  /**
   * Broadcast Git status change to refresh Explorer tree and SCM list
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
   * Register Git status change listener
   */
  function onGitStatusChanged(listener: GitStatusChangeListener): () => void {
    gitStatusChangeListeners.add(listener);
    return () => {
      gitStatusChangeListeners.delete(listener);
    };
  }

  /**
   * Request opening diff comparison tab in Editor
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
   * Register diff comparison open listener
   */
  function onOpenDiff(listener: OpenDiffListener): () => void {
    openDiffListeners.add(listener);
    return () => {
      openDiffListeners.delete(listener);
    };
  }

  /**
   * Notify specified file path to reload latest content from disk (Git Discard, external change, Checkout)
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
   * Register file reload listener
   */
  function onFileReloadNeeded(listener: FileReloadListener): () => void {
    fileReloadListeners.add(listener);
    return () => {
      fileReloadListeners.delete(listener);
    };
  }

  /**
   * Broadcast file or directory rename event
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
   * Register file or directory rename listener
   */
  function onFileRenamed(listener: FileRenameListener): () => void {
    fileRenameListeners.add(listener);
    return () => {
      fileRenameListeners.delete(listener);
    };
  }

  /**
   * Broadcast file or directory delete event
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
   * Register file or directory delete listener
   */
  function onFileDeleted(listener: FileDeleteListener): () => void {
    fileDeleteListeners.add(listener);
    return () => {
      fileDeleteListeners.delete(listener);
    };
  }

  /**
   * Record currently active focused file in editor
   */
  function setActiveFilePath(path: string): void {
    currentActiveFilePath.value = path;
  }

  /**
   * Open specified file in Editor and navigate to line/column
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
   * Register open file in Editor listener
   */
  function onOpenInEditor(listener: OpenInEditorListener): () => void {
    openInEditorListeners.add(listener);
    return () => {
      openInEditorListeners.delete(listener);
    };
  }

  /**
   * Send command or code to active terminal for execution
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
   * Register send to terminal listener
   */
  function onSendToTerminal(listener: SendToTerminalListener): () => void {
    sendToTerminalListeners.add(listener);
    return () => {
      sendToTerminalListeners.delete(listener);
    };
  }

  /**
   * Open specified working directory in integrated terminal
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
   * Register open directory in terminal listener
   */
  function onRevealInTerminal(listener: RevealInTerminalListener): () => void {
    revealInTerminalListeners.add(listener);
    return () => {
      revealInTerminalListeners.delete(listener);
    };
  }

  /**
   * Trigger toolchain action (switch, install, pin, remove from sidebar menu)
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
   * Register toolchain action listener
   */
  function onToolAction(listener: ToolActionListener): () => void {
    toolActionListeners.add(listener);
    return () => {
      toolActionListeners.delete(listener);
    };
  }

  /**
   * Broadcast toolchain configuration or status change
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
   * Register toolchain change listener
   */
  function onToolsChanged(listener: ToolsChangeListener): () => void {
    toolsChangeListeners.add(listener);
    return () => {
      toolsChangeListeners.delete(listener);
    };
  }

  /**
   * Trigger task action (run, edit, duplicate, remove from sidebar menu)
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
   * Register task action listener
   */
  function onTaskAction(listener: TaskActionListener): () => void {
    taskActionListeners.add(listener);
    return () => {
      taskActionListeners.delete(listener);
    };
  }

  /**
   * Broadcast task list change
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
   * Register task list change listener
   */
  function onTasksChanged(listener: TasksChangeListener): () => void {
    tasksChangeListeners.add(listener);
    return () => {
      tasksChangeListeners.delete(listener);
    };
  }

  /**
   * Broadcast environment doctor diagnosis change
   */
  function notifyDoctorChanged(report?: any): void {
    doctorChangeListeners.forEach((fn) => {
      try {
        fn(report);
      } catch (err) {
        console.error("[WorkspaceBus] Error in doctor change listener:", err);
      }
    });
  }

  /**
   * Register environment doctor change listener
   */
  function onDoctorChanged(listener: DoctorChangeListener): () => void {
    doctorChangeListeners.add(listener);
    return () => {
      doctorChangeListeners.delete(listener);
    };
  }

  /**
   * Trigger environment variable action (edit, copy value, copy name, restore, delete)
   */
  function triggerEnvAction(req: EnvActionRequest): void {
    if (!req || !req.envName) return;
    envActionListeners.forEach((fn) => {
      try {
        fn(req);
      } catch (err) {
        console.error("[WorkspaceBus] Error in env action listener:", err);
      }
    });
  }

  /**
   * Register environment variable action listener
   */
  function onEnvAction(listener: EnvActionListener): () => void {
    envActionListeners.add(listener);
    return () => {
      envActionListeners.delete(listener);
    };
  }

  /**
   * Broadcast environment variable configuration change
   */
  function notifyEnvsChanged(): void {
    envsChangeListeners.forEach((fn) => {
      try {
        fn();
      } catch (err) {
        console.error("[WorkspaceBus] Error in envs change listener:", err);
      }
    });
  }

  /**
   * Register environment variable change listener
   */
  function onEnvsChanged(listener: EnvsChangeListener): () => void {
    envsChangeListeners.add(listener);
    return () => {
      envsChangeListeners.delete(listener);
    };
  }

  /**
   * Clear all listeners (for unit test isolation)
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
    envActionListeners.clear();
    envsChangeListeners.clear();
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
    triggerEnvAction,
    onEnvAction,
    notifyEnvsChanged,
    onEnvsChanged,
    _clearAllListeners,
  };
}
