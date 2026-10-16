import { describe, it, expect, beforeEach } from "vitest";
import { useAppScope, SCOPE_STORAGE_KEY } from "./useAppScope";

const storageMap = new Map<string, string>();
const localStorageMock = {
  getItem: (key: string) => storageMap.get(key) ?? null,
  setItem: (key: string, val: string) => storageMap.set(key, String(val)),
  removeItem: (key: string) => storageMap.delete(key),
  clear: () => storageMap.clear(),
};

(globalThis as any).window = {
  localStorage: localStorageMock,
  location: { search: "" },
};

describe("useAppScope", () => {
  beforeEach(() => {
    storageMap.clear();
    const { setScope } = useAppScope();
    setScope("workspace");
  });

  it("defaults to workspace scope initially", () => {
    const { activeScope, isWorkspace, isUser } = useAppScope();
    expect(activeScope.value).toBe("workspace");
    expect(isWorkspace.value).toBe(true);
    expect(isUser.value).toBe(false);
  });

  it("switches to user scope and updates local storage", () => {
    const { activeScope, setScope, isWorkspace, isUser } = useAppScope();
    setScope("user");
    expect(activeScope.value).toBe("user");
    expect(isUser.value).toBe(true);
    expect(isWorkspace.value).toBe(false);
    expect(localStorageMock.getItem(SCOPE_STORAGE_KEY)).toBe("user");
  });

  it("toggles scope back and forth", () => {
    const { activeScope, toggleScope } = useAppScope();
    expect(activeScope.value).toBe("workspace");
    toggleScope();
    expect(activeScope.value).toBe("user");
    toggleScope();
    expect(activeScope.value).toBe("workspace");
  });

  it("syncs from route query", () => {
    const { activeScope, syncFromRoute } = useAppScope();
    syncFromRoute({
      query: { scope: "user" },
    } as any);
    expect(activeScope.value).toBe("user");

    syncFromRoute({
      query: { scope: "workspace" },
    } as any);
    expect(activeScope.value).toBe("workspace");
  });
});
