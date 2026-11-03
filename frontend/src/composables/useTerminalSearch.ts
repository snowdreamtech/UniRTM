// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, watch, type Ref } from "vue";
import type { SearchAddon, ISearchOptions } from "@xterm/addon-search";

export function useTerminalSearch(searchAddonRef: Ref<SearchAddon | null>) {
  const query = ref<string>("");
  const matchCase = ref<boolean>(false);
  const matchWholeWord = ref<boolean>(false);
  const useRegex = ref<boolean>(false);

  const resultIndex = ref<number>(-1);
  const resultCount = ref<number>(0);

  let resultsListener: { dispose: () => void } | null = null;

  function getSearchOptions(incremental = false): ISearchOptions {
    return {
      caseSensitive: matchCase.value,
      wholeWord: matchWholeWord.value,
      regex: useRegex.value,
      incremental,
      decorations: {
        matchBackground: "rgba(234, 179, 8, 0.35)",
        matchBorder: "rgba(234, 179, 8, 0.7)",
        activeMatchBackground: "rgba(249, 115, 22, 0.65)",
        activeMatchBorder: "#f97316",
        matchOverviewRuler: "#eab308",
        activeMatchColorOverviewRuler: "#f97316",
      },
    };
  }

  function findNext(): void {
    if (!searchAddonRef.value || !query.value) return;
    searchAddonRef.value.findNext(query.value, getSearchOptions(false));
  }

  function findPrevious(): void {
    if (!searchAddonRef.value || !query.value) return;
    searchAddonRef.value.findPrevious(query.value, getSearchOptions(false));
  }

  function handleQueryChange(newQuery: string): void {
    query.value = newQuery;
    if (!searchAddonRef.value) return;

    if (!newQuery) {
      searchAddonRef.value.clearDecorations?.();
      resultIndex.value = -1;
      resultCount.value = 0;
      return;
    }

    searchAddonRef.value.findNext(newQuery, getSearchOptions(true));
  }

  function toggleMatchCase(): void {
    matchCase.value = !matchCase.value;
    if (query.value) {
      handleQueryChange(query.value);
    }
  }

  function toggleMatchWholeWord(): void {
    matchWholeWord.value = !matchWholeWord.value;
    if (query.value) {
      handleQueryChange(query.value);
    }
  }

  function toggleUseRegex(): void {
    useRegex.value = !useRegex.value;
    if (query.value) {
      handleQueryChange(query.value);
    }
  }

  function clearSearch(): void {
    query.value = "";
    resultIndex.value = -1;
    resultCount.value = 0;
    searchAddonRef.value?.clearDecorations?.();
  }

  function attachResultsListener(): void {
    if (resultsListener) {
      resultsListener.dispose();
      resultsListener = null;
    }
    const addon = searchAddonRef.value;
    if (addon && typeof addon.onDidChangeResults === "function") {
      resultsListener = addon.onDidChangeResults((e: { resultIndex: number; resultCount: number }) => {
        resultIndex.value = e.resultIndex;
        resultCount.value = e.resultCount;
      });
    }
  }

  watch(
    () => searchAddonRef.value,
    () => {
      attachResultsListener();
    },
    { immediate: true }
  );

  function dispose(): void {
    if (resultsListener) {
      resultsListener.dispose();
      resultsListener = null;
    }
  }

  return {
    query,
    matchCase,
    matchWholeWord,
    useRegex,
    resultIndex,
    resultCount,
    getSearchOptions,
    findNext,
    findPrevious,
    handleQueryChange,
    toggleMatchCase,
    toggleMatchWholeWord,
    toggleUseRegex,
    clearSearch,
    dispose,
  };
}
