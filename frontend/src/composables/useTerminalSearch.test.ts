// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, vi } from "vitest";
import { ref } from "vue";
import { useTerminalSearch } from "./useTerminalSearch";

describe("useTerminalSearch", () => {
  function createMockSearchAddon() {
    let resultsCallback: ((e: any) => void) | null = null;
    return {
      findNext: vi.fn(),
      findPrevious: vi.fn(),
      clearDecorations: vi.fn(),
      onDidChangeResults: vi.fn((cb) => {
        resultsCallback = cb;
        return { dispose: vi.fn() };
      }),
      _triggerResults(resultIndex: number, resultCount: number) {
        if (resultsCallback) {
          resultsCallback({ resultIndex, resultCount });
        }
      },
    };
  }

  it("initializes with default values", () => {
    const mockAddon = createMockSearchAddon();
    const addonRef = ref(mockAddon as any);
    const search = useTerminalSearch(addonRef);

    expect(search.query.value).toBe("");
    expect(search.matchCase.value).toBe(false);
    expect(search.matchWholeWord.value).toBe(false);
    expect(search.useRegex.value).toBe(false);
    expect(search.resultIndex.value).toBe(-1);
    expect(search.resultCount.value).toBe(0);
  });

  it("handles query change and triggers findNext with incremental flag", () => {
    const mockAddon = createMockSearchAddon();
    const addonRef = ref(mockAddon as any);
    const search = useTerminalSearch(addonRef);

    search.handleQueryChange("compile error");
    expect(search.query.value).toBe("compile error");
    expect(mockAddon.findNext).toHaveBeenCalledWith(
      "compile error",
      expect.objectContaining({
        caseSensitive: false,
        wholeWord: false,
        regex: false,
        incremental: true,
      })
    );
  });

  it("clears decorations when query is emptied", () => {
    const mockAddon = createMockSearchAddon();
    const addonRef = ref(mockAddon as any);
    const search = useTerminalSearch(addonRef);

    search.handleQueryChange("initial");
    search.handleQueryChange("");

    expect(mockAddon.clearDecorations).toHaveBeenCalled();
    expect(search.resultIndex.value).toBe(-1);
    expect(search.resultCount.value).toBe(0);
  });

  it("toggles search options and triggers search if query is non-empty", () => {
    const mockAddon = createMockSearchAddon();
    const addonRef = ref(mockAddon as any);
    const search = useTerminalSearch(addonRef);

    search.query.value = "target";

    search.toggleMatchCase();
    expect(search.matchCase.value).toBe(true);
    expect(mockAddon.findNext).toHaveBeenCalledWith("target", expect.objectContaining({ caseSensitive: true }));

    search.toggleMatchWholeWord();
    expect(search.matchWholeWord.value).toBe(true);
    expect(mockAddon.findNext).toHaveBeenCalledWith("target", expect.objectContaining({ wholeWord: true }));

    search.toggleUseRegex();
    expect(search.useRegex.value).toBe(true);
    expect(mockAddon.findNext).toHaveBeenCalledWith("target", expect.objectContaining({ regex: true }));
  });

  it("findNext and findPrevious call search addon methods", () => {
    const mockAddon = createMockSearchAddon();
    const addonRef = ref(mockAddon as any);
    const search = useTerminalSearch(addonRef);

    search.query.value = "keyword";

    search.findNext();
    expect(mockAddon.findNext).toHaveBeenCalledWith("keyword", expect.objectContaining({ incremental: false }));

    search.findPrevious();
    expect(mockAddon.findPrevious).toHaveBeenCalledWith("keyword", expect.objectContaining({ incremental: false }));
  });

  it("receives results update from search addon", () => {
    const mockAddon = createMockSearchAddon();
    const addonRef = ref(mockAddon as any);
    const search = useTerminalSearch(addonRef);

    mockAddon._triggerResults(2, 5);
    expect(search.resultIndex.value).toBe(2);
    expect(search.resultCount.value).toBe(5);
  });

  it("clears all state on clearSearch", () => {
    const mockAddon = createMockSearchAddon();
    const addonRef = ref(mockAddon as any);
    const search = useTerminalSearch(addonRef);

    search.query.value = "anything";
    search.resultIndex.value = 1;
    search.resultCount.value = 3;

    search.clearSearch();
    expect(search.query.value).toBe("");
    expect(search.resultIndex.value).toBe(-1);
    expect(search.resultCount.value).toBe(0);
    expect(mockAddon.clearDecorations).toHaveBeenCalled();
  });
});
