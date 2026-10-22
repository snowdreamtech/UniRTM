// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, vi, beforeEach } from "vitest";
import { useKeyboardShortcuts, SHORTCUT_DEFINITIONS } from "./useKeyboardShortcuts";

function createKeyEvent(opts: {
  ctrlKey?: boolean;
  shiftKey?: boolean;
  altKey?: boolean;
  metaKey?: boolean;
  key: string;
}) {
  return {
    ctrlKey: opts.ctrlKey ?? false,
    shiftKey: opts.shiftKey ?? false,
    altKey: opts.altKey ?? false,
    metaKey: opts.metaKey ?? false,
    key: opts.key,
    preventDefault: vi.fn(),
  } as unknown as KeyboardEvent;
}

describe("useKeyboardShortcuts", () => {
  const mockHandlers = {
    openCommandPalette: vi.fn(),
    closeCommandPalette: vi.fn(),
    isCommandPaletteOpen: vi.fn(() => false),
    openSettings: vi.fn(),
    toggleSidebar: vi.fn(),
    toggleDrawer: vi.fn(),
    openDrawerTab: vi.fn(),
    newTerminalSession: vi.fn(),
    navigateToTab: vi.fn(),
    toggleTheme: vi.fn(),
    toggleScope: vi.fn(),
    isDrawerOpen: vi.fn(() => false),
    closeDrawer: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("exports comprehensive VS Code aligned shortcut definitions", () => {
    expect(SHORTCUT_DEFINITIONS.length).toBeGreaterThanOrEqual(15);
    const ids = SHORTCUT_DEFINITIONS.map((s) => s.id);
    expect(ids).toContain("workbench.action.showCommands");
    expect(ids).toContain("workbench.action.quickOpen");
    expect(ids).toContain("workbench.action.openSettings");
    expect(ids).toContain("workbench.action.openGlobalKeybindings");
    expect(ids).toContain("workbench.action.toggleSidebarVisibility");
    expect(ids).toContain("workbench.action.togglePanel");
    expect(ids).toContain("workbench.action.terminal.toggleTerminal");
    expect(ids).toContain("workbench.action.terminal.new");
    expect(ids).toContain("workbench.view.explorer");
    expect(ids).toContain("workbench.view.scm");
    expect(ids).toContain("workbench.view.doctor");
    expect(ids).toContain("workbench.view.extensions");
    expect(ids).toContain("workbench.action.files.save");
    expect(ids).toContain("workbench.action.closeActiveEditor");
  });

  it("handles modal visibility toggling", () => {
    const { isShortcutsModalOpen, openShortcutsModal, closeShortcutsModal } = useKeyboardShortcuts();
    expect(isShortcutsModalOpen.value).toBe(false);
    openShortcutsModal();
    expect(isShortcutsModalOpen.value).toBe(true);
    closeShortcutsModal();
    expect(isShortcutsModalOpen.value).toBe(false);
  });

  it("triggers Command Palette via Ctrl+Shift+P, F1, and Ctrl+P", () => {
    const { createKeydownHandler } = useKeyboardShortcuts();
    const handleKeydown = createKeydownHandler(mockHandlers);

    // Ctrl+Shift+P
    const eventP = createKeyEvent({ ctrlKey: true, shiftKey: true, key: "p" });
    handleKeydown(eventP);
    expect(mockHandlers.openCommandPalette).toHaveBeenCalledTimes(1);

    // F1
    const eventF1 = createKeyEvent({ key: "F1" });
    handleKeydown(eventF1);
    expect(mockHandlers.openCommandPalette).toHaveBeenCalledTimes(2);

    // Ctrl+P
    const eventQuick = createKeyEvent({ ctrlKey: true, key: "p" });
    handleKeydown(eventQuick);
    expect(mockHandlers.openCommandPalette).toHaveBeenCalledTimes(3);
  });

  it("triggers Sidebar toggle via Ctrl+B", () => {
    const { createKeydownHandler } = useKeyboardShortcuts();
    const handleKeydown = createKeydownHandler(mockHandlers);

    const eventB = createKeyEvent({ ctrlKey: true, key: "b" });
    handleKeydown(eventB);
    expect(mockHandlers.toggleSidebar).toHaveBeenCalledTimes(1);
  });

  it("triggers Bottom Panel toggle via Ctrl+J", () => {
    const { createKeydownHandler } = useKeyboardShortcuts();
    const handleKeydown = createKeydownHandler(mockHandlers);

    const eventJ = createKeyEvent({ ctrlKey: true, key: "j" });
    handleKeydown(eventJ);
    expect(mockHandlers.toggleDrawer).toHaveBeenCalledTimes(1);
  });

  it("triggers Integrated Terminal toggle via Ctrl+`", () => {
    const { createKeydownHandler } = useKeyboardShortcuts();
    const handleKeydown = createKeydownHandler(mockHandlers);

    const eventTerm = createKeyEvent({ ctrlKey: true, key: "`" });
    handleKeydown(eventTerm);
    expect(mockHandlers.openDrawerTab).toHaveBeenCalledWith("terminal");
  });

  it("triggers New Terminal session via Ctrl+Shift+`", () => {
    const { createKeydownHandler } = useKeyboardShortcuts();
    const handleKeydown = createKeydownHandler(mockHandlers);

    const eventNewTerm = createKeyEvent({ ctrlKey: true, shiftKey: true, key: "`" });
    handleKeydown(eventNewTerm);
    expect(mockHandlers.newTerminalSession).toHaveBeenCalledTimes(1);
  });

  it("navigates to views via Ctrl+Shift+E, G, D, X", () => {
    const { createKeydownHandler } = useKeyboardShortcuts();
    const handleKeydown = createKeydownHandler(mockHandlers);

    handleKeydown(createKeyEvent({ ctrlKey: true, shiftKey: true, key: "e" }));
    expect(mockHandlers.navigateToTab).toHaveBeenCalledWith("explorer");

    handleKeydown(createKeyEvent({ ctrlKey: true, shiftKey: true, key: "g" }));
    expect(mockHandlers.navigateToTab).toHaveBeenCalledWith("scm");

    handleKeydown(createKeyEvent({ ctrlKey: true, shiftKey: true, key: "d" }));
    expect(mockHandlers.navigateToTab).toHaveBeenCalledWith("doctor");

    handleKeydown(createKeyEvent({ ctrlKey: true, shiftKey: true, key: "x" }));
    expect(mockHandlers.navigateToTab).toHaveBeenCalledWith("marketplace");
  });

  it("switches tabs via Alt+1 through Alt+9", () => {
    const { createKeydownHandler } = useKeyboardShortcuts();
    const handleKeydown = createKeydownHandler(mockHandlers);

    handleKeydown(createKeyEvent({ altKey: true, key: "1" }));
    expect(mockHandlers.navigateToTab).toHaveBeenCalledWith("dashboard");

    handleKeydown(createKeyEvent({ altKey: true, key: "2" }));
    expect(mockHandlers.navigateToTab).toHaveBeenCalledWith("explorer");

    handleKeydown(createKeyEvent({ altKey: true, key: "3" }));
    expect(mockHandlers.navigateToTab).toHaveBeenCalledWith("scm");
  });

  it("toggles scope via Alt+S", () => {
    const { createKeydownHandler } = useKeyboardShortcuts();
    const handleKeydown = createKeydownHandler(mockHandlers);

    handleKeydown(createKeyEvent({ altKey: true, key: "s" }));
    expect(mockHandlers.toggleScope).toHaveBeenCalledTimes(1);
  });

  it("supports custom keybinding overrides and dynamic matching", async () => {
    // Mock global fetch
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        success: true,
        path: "/mock/keybindings.json",
        raw: '[{"command":"workbench.view.scm","key":"Ctrl+Alt+G"}]',
        keybindings: [{ command: "workbench.view.scm", key: "Ctrl+Alt+G" }],
      }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const { activeShortcuts, saveKeybinding, resetKeybinding, resetAllKeybindings, createKeydownHandler } =
      useKeyboardShortcuts();

    // Save custom keybinding
    await saveKeybinding("workbench.view.scm", "Ctrl+Alt+G");
    expect(fetchMock).toHaveBeenCalledWith("/api/keybindings", expect.objectContaining({ method: "POST" }));

    const scmItem = activeShortcuts.value.find((s) => s.id === "workbench.view.scm");
    expect(scmItem?.effectiveKey).toBe("Ctrl+Alt+G");
    expect(scmItem?.isCustom).toBe(true);
    expect(scmItem?.source).toBe("user");

    // Dynamic keydown test with custom key
    const handleKeydown = createKeydownHandler(mockHandlers);
    handleKeydown(createKeyEvent({ ctrlKey: true, altKey: true, key: "g" }));
    expect(mockHandlers.navigateToTab).toHaveBeenCalledWith("scm");

    // Reset single keybinding
    await resetKeybinding("workbench.view.scm");
    const scmItemAfter = activeShortcuts.value.find((s) => s.id === "workbench.view.scm");
    expect(scmItemAfter?.isCustom).toBe(false);
    expect(scmItemAfter?.source).toBe("default");

    // Reset all keybindings
    await resetAllKeybindings();
    expect(fetchMock).toHaveBeenCalledWith("/api/keybindings", expect.objectContaining({ method: "POST" }));

    vi.unstubAllGlobals();
  });
});
