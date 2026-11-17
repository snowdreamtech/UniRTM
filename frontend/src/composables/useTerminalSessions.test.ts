import { describe, it, expect, beforeEach } from "vitest";
import { useTerminalSessions } from "./useTerminalSessions";

describe("useTerminalSessions composable", () => {
  beforeEach(() => {
    const { sessions } = useTerminalSessions();
    sessions.value = [];
  });

  it("creates initial session and active pane correctly", () => {
    const { sessions, activeSession, activePane, createSession } = useTerminalSessions();

    expect(sessions.value.length).toBe(0);
    expect(activeSession.value).toBeNull();

    createSession({ name: "bash", path: "/bin/bash" });
    expect(sessions.value.length).toBe(1);
    expect(activeSession.value).not.toBeNull();
    expect(activeSession.value?.title).toContain("bash");
    expect(activeSession.value?.splitDirection).toBe("none");
    expect(activePane.value).not.toBeNull();
    expect(activePane.value?.shellName).toBe("bash");
  });

  it("supports creating multiple sessions (tabs) and switching", () => {
    const { sessions, activeSession, createSession, setActiveSession } = useTerminalSessions();

    const s1 = createSession({ name: "bash", path: "/bin/bash" });
    const s2 = createSession({ name: "zsh", path: "/bin/zsh" });

    expect(sessions.value.length).toBe(2);
    expect(activeSession.value?.id).toBe(s2.id);

    setActiveSession(s1.id);
    expect(activeSession.value?.id).toBe(s1.id);
  });

  it("supports binary split tree (vertical, horizontal nested splits) matching VSCode and tmux", () => {
    const { activeSession, createSession, splitActiveSession, closePane } = useTerminalSessions();

    createSession({ name: "bash", path: "/bin/bash" });
    expect(activeSession.value?.panes.length).toBe(1);
    expect(activeSession.value?.layoutTree?.type).toBe("pane");

    // Split vertically (1 into 2)
    const pane2 = splitActiveSession("vertical", { name: "zsh", path: "/bin/zsh" });
    expect(activeSession.value?.panes.length).toBe(2);
    expect(activeSession.value?.splitDirection).toBe("vertical");
    expect(activeSession.value?.panes[1].shellName).toBe("zsh");
    expect(activeSession.value?.layoutTree?.type).toBe("split");

    // Nested split horizontally on active pane (pane2) -> 3 panes in T-shape layout
    const pane3 = splitActiveSession("horizontal", { name: "sh", path: "/bin/sh" });
    expect(activeSession.value?.panes.length).toBe(3);
    expect(activeSession.value?.splitDirection).toBe("horizontal");
    expect(activeSession.value?.panes[2].shellName).toBe("sh");

    // Verify binary split tree structure
    const tree = activeSession.value?.layoutTree;
    expect(tree?.type).toBe("split");
    if (tree?.type === "split") {
      expect(tree.direction).toBe("vertical");
      expect(tree.children[0].type).toBe("pane");
      expect(tree.children[1].type).toBe("split");
      if (tree.children[1].type === "split") {
        expect(tree.children[1].direction).toBe("horizontal");
        expect(tree.children[1].children[0].type).toBe("pane");
        expect(tree.children[1].children[1].type).toBe("pane");
      }
    }

    // Close pane 3: collapses nested split, leaves 2 panes
    closePane(pane3!.id);
    expect(activeSession.value?.panes.length).toBe(2);
    expect(activeSession.value?.layoutTree?.type).toBe("split");

    // Close pane 2: collapses root split, leaves 1 pane and restores splitDirection none
    closePane(pane2!.id);
    expect(activeSession.value?.panes.length).toBe(1);
    expect(activeSession.value?.splitDirection).toBe("none");
    expect(activeSession.value?.layoutTree?.type).toBe("pane");
  });

  it("closes split pane and restores single pane", () => {
    const { activeSession, createSession, splitActiveSession, closePane } = useTerminalSessions();

    createSession({ name: "bash", path: "/bin/bash" });
    const pane2 = splitActiveSession("vertical", { name: "zsh", path: "/bin/zsh" });
    expect(activeSession.value?.panes.length).toBe(2);

    closePane(pane2!.id);
    expect(activeSession.value?.panes.length).toBe(1);
    expect(activeSession.value?.splitDirection).toBe("none");
  });

  it("closes session and automatically ensures at least one session exists if all closed", () => {
    const { sessions, createSession, closeSession } = useTerminalSessions();

    const s1 = createSession({ name: "bash", path: "/bin/bash" });
    expect(sessions.value.length).toBe(1);

    closeSession(s1.id);
    // When last session closed, automatically creates a new one
    expect(sessions.value.length).toBe(1);
  });

  it("executes commands, records history and handles clear/echo", () => {
    const { activePane, createSession, runCommand } = useTerminalSessions();

    createSession({ name: "bash", path: "/bin/bash" });
    const paneId = activePane.value!.id;

    runCommand(paneId, "echo Hello UniRTM");
    expect(activePane.value?.lines.some((l) => l.text === "Hello UniRTM")).toBe(true);

    runCommand(paneId, "clear");
    expect(activePane.value?.lines.length).toBe(0);
  });
});
