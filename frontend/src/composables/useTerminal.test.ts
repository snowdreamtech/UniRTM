import { describe, it, expect, beforeEach, vi } from "vitest";
import { useTerminal } from "./useTerminal";

describe("useTerminal composable", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("fetches and detects terminals successfully", async () => {
    const mockResponse = {
      default: "/bin/bash",
      current: "/bin/bash",
      terminals: [
        { id: "bash", name: "Bash", path: "/bin/bash", is_default: true },
        { id: "zsh", name: "Zsh", path: "/usr/bin/zsh", is_default: false },
      ],
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
    });

    const { fetchTerminals, terminals, defaultTerminal, currentTerminal } = useTerminal();
    await fetchTerminals();

    expect(terminals.value).toHaveLength(2);
    expect(defaultTerminal.value?.path).toBe("/bin/bash");
    expect(currentTerminal.value?.path).toBe("/bin/bash");
  });

  it("updates preferred terminal and posts to /api/config", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true }),
    });

    const { setPreferredTerminal, currentTerminalPath } = useTerminal();
    const result = await setPreferredTerminal("/usr/bin/zsh");

    expect(result).toBe(true);
    expect(currentTerminalPath.value).toBe("/usr/bin/zsh");
    expect(globalThis.fetch).toHaveBeenCalledWith("/api/config", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        scope: "user",
        key: "settings.terminal",
        value: "/usr/bin/zsh",
      }),
    }));
  });
});
