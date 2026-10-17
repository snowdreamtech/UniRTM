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

  it("supports splitting terminal vertically (1 into 2) and horizontally", () => {
    const { activeSession, createSession, splitActiveSession } = useTerminalSessions();
    
    createSession({ name: "bash", path: "/bin/bash" });
    expect(activeSession.value?.panes.length).toBe(1);

    // Split vertically
    splitActiveSession("vertical", { name: "zsh", path: "/bin/zsh" });
    expect(activeSession.value?.panes.length).toBe(2);
    expect(activeSession.value?.splitDirection).toBe("vertical");
    expect(activeSession.value?.panes[1].shellName).toBe("zsh");

    // Switching split direction to horizontal
    splitActiveSession("horizontal");
    expect(activeSession.value?.splitDirection).toBe("horizontal");
    expect(activeSession.value?.panes.length).toBe(2);
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
