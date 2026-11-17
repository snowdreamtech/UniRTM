import { ref, computed } from "vue";
import { useTerminal, TerminalItem } from "./useTerminal";
import { useTerminalPersistence } from "./useTerminalPersistence";
import { useTerminalProcessWatcher } from "./useTerminalProcessWatcher";
import { useTerminalNotifier } from "./useTerminalNotifier";

export type SplitDirection = "none" | "vertical" | "horizontal";

export interface TerminalLine {
  text: string;
  type: "info" | "input" | "output" | "error" | "system";
}

export interface TerminalPane {
  id: string;
  shellName: string;
  shellPath: string;
  lines: TerminalLine[];
  input: string;
  history: string[];
  historyIndex: number;
}

export type TerminalLayoutNode = TerminalLayoutSplitNode | TerminalLayoutPaneNode;

export interface TerminalLayoutSplitNode {
  id: string;
  type: "split";
  direction: "vertical" | "horizontal";
  children: [TerminalLayoutNode, TerminalLayoutNode];
}

export interface TerminalLayoutPaneNode {
  type: "pane";
  paneId: string;
}

export interface TerminalSession {
  id: string;
  title: string;
  splitDirection: SplitDirection;
  panes: TerminalPane[];
  activePaneId: string;
  layoutTree?: TerminalLayoutNode;
}

export function findNodeAndParent(
  root: TerminalLayoutNode,
  targetPaneId: string,
  parent: TerminalLayoutSplitNode | null = null
): { node: TerminalLayoutPaneNode; parent: TerminalLayoutSplitNode | null; indexInParent: number } | null {
  if (root.type === "pane") {
    if (root.paneId === targetPaneId) {
      return { node: root, parent, indexInParent: -1 };
    }
    return null;
  }
  for (let i = 0; i < root.children.length; i++) {
    const child = root.children[i];
    if (child.type === "pane" && child.paneId === targetPaneId) {
      return { node: child, parent: root, indexInParent: i };
    }
    if (child.type === "split") {
      const found = findNodeAndParent(child, targetPaneId, root);
      if (found) return found;
    }
  }
  return null;
}

export function findParentOf(
  root: TerminalLayoutNode,
  targetNode: TerminalLayoutNode
): { parent: TerminalLayoutSplitNode; indexInParent: number } | null {
  if (root.type !== "split") return null;
  for (let i = 0; i < root.children.length; i++) {
    if (root.children[i] === targetNode) {
      return { parent: root, indexInParent: i };
    }
    if (root.children[i].type === "split") {
      const found = findParentOf(root.children[i], targetNode);
      if (found) return found;
    }
  }
  return null;
}

const sessions = ref<TerminalSession[]>([]);
const activeSessionId = ref<string>("");
let sessionCounter = 0;
let paneCounter = 0;

type TerminalDataSender = (data: string) => void;
const activeTerminalSenders = new Map<string, TerminalDataSender>();

export function registerTerminalSender(paneId: string, sender: TerminalDataSender) {
  activeTerminalSenders.set(paneId, sender);
}

export function unregisterTerminalSender(paneId: string) {
  activeTerminalSenders.delete(paneId);
}

export function sendToTerminal(paneId: string | undefined, data: string): boolean {
  if (paneId && activeTerminalSenders.has(paneId)) {
    activeTerminalSenders.get(paneId)!(data);
    return true;
  }
  for (const sender of activeTerminalSenders.values()) {
    sender(data);
    return true;
  }
  return false;
}

export function useTerminalSessions() {
  const { currentTerminal, defaultTerminal } = useTerminal();

  function getEffectiveShell(shell?: { name: string; path: string }): { name: string; path: string } {
    if (shell && shell.path) {
      return shell;
    }
    if (currentTerminal.value) {
      return {
        name: currentTerminal.value.name || "Shell",
        path: currentTerminal.value.path || "/bin/sh",
      };
    }
    if (defaultTerminal.value) {
      return {
        name: defaultTerminal.value.name || "Shell",
        path: defaultTerminal.value.path || "/bin/sh",
      };
    }
    return { name: "bash", path: "/bin/bash" };
  }

  function createPane(shell?: { name: string; path: string }): TerminalPane {
    paneCounter += 1;
    const effectiveShell = getEffectiveShell(shell);
    return {
      id: `pane-${Date.now()}-${paneCounter}`,
      shellName: effectiveShell.name,
      shellPath: effectiveShell.path,
      lines: [
        {
          text: `UniRTM Terminal Session (${effectiveShell.name})`,
          type: "system",
        },
      ],
      input: "",
      history: [],
      historyIndex: -1,
    };
  }

  function createSession(shell?: { name: string; path: string }): TerminalSession {
    sessionCounter += 1;
    const pane = createPane(shell);
    const session: TerminalSession = {
      id: `session-${Date.now()}-${sessionCounter}`,
      title: `${sessionCounter}: ${pane.shellName}`,
      splitDirection: "none",
      panes: [pane],
      activePaneId: pane.id,
      layoutTree: {
        type: "pane",
        paneId: pane.id,
      },
    };

    sessions.value.push(session);
    activeSessionId.value = session.id;
    return session;
  }

  const activeSession = computed<TerminalSession | null>(() => {
    if (sessions.value.length === 0) return null;
    const found = sessions.value.find((s) => s.id === activeSessionId.value);
    return found || sessions.value[0] || null;
  });

  const activePane = computed<TerminalPane | null>(() => {
    const session = activeSession.value;
    if (!session || session.panes.length === 0) return null;
    const found = session.panes.find((p) => p.id === session.activePaneId);
    return found || session.panes[0] || null;
  });

  function splitActiveSession(
    direction: "vertical" | "horizontal",
    shell?: { name: string; path: string },
    targetPaneId?: string
  ): TerminalPane | null {
    if (sessions.value.length === 0) {
      createSession(shell);
    }
    const session = activeSession.value;
    if (!session) return null;

    const targetId = targetPaneId || session.activePaneId || (session.panes.length > 0 ? session.panes[0].id : "");

    const newPane = createPane(shell);
    session.panes.push(newPane);

    if (!session.layoutTree) {
      session.layoutTree = {
        type: "pane",
        paneId: targetId || newPane.id,
      };
    }

    const found = findNodeAndParent(session.layoutTree, targetId);
    const splitNode: TerminalLayoutSplitNode = {
      id: `split-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      type: "split",
      direction,
      children: [
        { type: "pane", paneId: targetId },
        { type: "pane", paneId: newPane.id },
      ],
    };

    if (found) {
      if (!found.parent) {
        // Target was root node
        session.layoutTree = splitNode;
      } else {
        // Target was child of a split node
        found.parent.children[found.indexInParent] = splitNode;
      }
    } else {
      // Fallback: wrap whole tree with new pane
      session.layoutTree = {
        id: `split-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
        type: "split",
        direction,
        children: [session.layoutTree, { type: "pane", paneId: newPane.id }],
      };
    }

    session.splitDirection = direction;
    session.activePaneId = newPane.id;
    return newPane;
  }

  function setActiveSession(sessionId: string) {
    const target = sessions.value.find((s) => s.id === sessionId);
    if (target) {
      activeSessionId.value = target.id;
    }
  }

  function setActivePane(paneId: string, sessionId?: string) {
    const targetSession = sessionId ? sessions.value.find((s) => s.id === sessionId) : activeSession.value;
    if (targetSession) {
      const pane = targetSession.panes.find((p) => p.id === paneId);
      if (pane) {
        targetSession.activePaneId = pane.id;
        activeSessionId.value = targetSession.id;
      }
    }
  }

  function cleanupPaneResources(paneId: string): void {
    if (!paneId) return;
    unregisterTerminalSender(paneId);
    try {
      useTerminalPersistence().clearBuffer(paneId);
    } catch {}
    try {
      const watcher = useTerminalProcessWatcher();
      watcher.activeProcesses.delete(paneId);
      watcher.customTitles.delete(paneId);
    } catch {}
    try {
      useTerminalNotifier().clearTracked(paneId);
    } catch {}
  }

  function closePane(paneId: string, sessionId?: string) {
    const targetSession = sessionId
      ? sessions.value.find((s) => s.id === sessionId)
      : sessions.value.find((s) => s.panes.some((p) => p.id === paneId));

    if (!targetSession) return;

    cleanupPaneResources(paneId);

    if (targetSession.panes.length > 1) {
      targetSession.panes = targetSession.panes.filter((p) => p.id !== paneId);

      if (targetSession.layoutTree) {
        const found = findNodeAndParent(targetSession.layoutTree, paneId);
        if (found && found.parent) {
          const siblingIndex = found.indexInParent === 0 ? 1 : 0;
          const siblingNode = found.parent.children[siblingIndex];

          const parentFound = findParentOf(targetSession.layoutTree, found.parent);
          if (!parentFound) {
            // found.parent was the root node
            targetSession.layoutTree = siblingNode;
          } else {
            parentFound.parent.children[parentFound.indexInParent] = siblingNode;
          }
        }
      }

      if (targetSession.panes.length === 1) {
        targetSession.splitDirection = "none";
        targetSession.layoutTree = {
          type: "pane",
          paneId: targetSession.panes[0].id,
        };
      } else if (targetSession.layoutTree && targetSession.layoutTree.type === "split") {
        targetSession.splitDirection = targetSession.layoutTree.direction;
      }

      if (targetSession.activePaneId === paneId) {
        targetSession.activePaneId = targetSession.panes[0]?.id || "";
      }
    } else {
      closeSession(targetSession.id);
    }
  }

  function closeSession(sessionId: string) {
    const idx = sessions.value.findIndex((s) => s.id === sessionId);
    if (idx === -1) return;

    const removedSession = sessions.value[idx];
    if (removedSession && Array.isArray(removedSession.panes)) {
      for (const p of removedSession.panes) {
        cleanupPaneResources(p.id);
      }
    }

    sessions.value.splice(idx, 1);

    if (sessions.value.length === 0) {
      createSession();
    } else {
      const nextSession = sessions.value[Math.max(0, idx - 1)] || sessions.value[0];
      activeSessionId.value = nextSession.id;
    }
  }

  function closeActiveTerminal() {
    const pane = activePane.value;
    const session = activeSession.value;
    if (pane && session) {
      closePane(pane.id, session.id);
    }
  }

  function runCommand(paneId: string, commandText: string) {
    const session = sessions.value.find((s) => s.panes.some((p) => p.id === paneId));
    if (!session) return;
    const pane = session.panes.find((p) => p.id === paneId);
    if (!pane) return;

    const trimmed = commandText.trim();
    if (!trimmed) return;

    pane.lines.push({ text: trimmed, type: "input" });
    pane.history.push(trimmed);
    pane.historyIndex = -1;
    pane.input = "";

    if (trimmed === "clear") {
      pane.lines = [];
      return;
    }

    if (trimmed === "exit") {
      closePane(pane.id, session.id);
      return;
    }

    if (trimmed === "unirtm" || trimmed === "unirtm --help") {
      pane.lines.push({
        text: "UniRTM - Unified Runtime & Tool Manager (v0.35.0)",
        type: "output",
      });
      pane.lines.push({
        text: "Commands: run, test, install, doctor, marketplace",
        type: "output",
      });
      return;
    }

    if (trimmed.startsWith("echo ")) {
      pane.lines.push({
        text: trimmed.slice(5),
        type: "output",
      });
      return;
    }

    pane.lines.push({
      text: `[${pane.shellName}] executed: ${trimmed}`,
      type: "output",
    });

    // Send command directly to live WebSocket PTY shell
    sendToTerminal(paneId, `${trimmed}\r`);
  }

  function clearActivePane() {
    const pane = activePane.value;
    if (pane) {
      pane.lines = [];
    }
  }

  // Ensure at least one session exists if initialized
  function ensureInitialSession(shell?: TerminalItem | null) {
    if (sessions.value.length === 0) {
      createSession(shell || undefined);
    }
  }

  return {
    sessions,
    activeSessionId,
    activeSession,
    activePane,
    createSession,
    splitActiveSession,
    setActiveSession,
    setActivePane,
    closePane,
    closeSession,
    closeActiveTerminal,
    runCommand,
    sendToTerminal,
    clearActivePane,
    ensureInitialSession,
  };
}
