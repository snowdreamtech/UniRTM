<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->
<script setup lang="ts">
import { ref, toRef, watch, nextTick } from "vue";
import type { SearchAddon } from "@xterm/addon-search";
import { useTerminalSearch } from "../../composables/useTerminalSearch";
import { t } from "../../i18n";

const props = defineProps<{
  visible: boolean;
  searchAddon?: SearchAddon | null;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "search", query: string): void;
}>();

const inputRef = ref<HTMLInputElement | null>(null);
const searchAddonRef = toRef(props, "searchAddon") as any;

const {
  query,
  matchCase,
  matchWholeWord,
  useRegex,
  resultIndex,
  resultCount,
  findNext,
  findPrevious,
  handleQueryChange,
  toggleMatchCase,
  toggleMatchWholeWord,
  toggleUseRegex,
  clearSearch,
} = useTerminalSearch(searchAddonRef);

function onInput(e: Event) {
  const val = (e.target as HTMLInputElement).value;
  handleQueryChange(val);
  emit("search", val);
}

function handleFindNext() {
  findNext();
  emit("search", query.value);
}

function handleFindPrevious() {
  findPrevious();
  emit("search", query.value);
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    e.preventDefault();
    e.stopPropagation();
    emit("close");
  } else if (e.key === "Enter") {
    e.preventDefault();
    if (e.shiftKey) {
      handleFindPrevious();
    } else {
      handleFindNext();
    }
  }
}

watch(
  () => props.visible,
  (val) => {
    if (val) {
      nextTick(() => {
        inputRef.value?.focus();
        inputRef.value?.select();
        if (query.value) {
          handleQueryChange(query.value);
        }
      });
    } else {
      clearSearch();
    }
  }
);
</script>

<template>
  <transition name="search-slide">
    <div
      v-if="visible"
      class="terminal-search-widget"
      role="search"
      aria-label="Terminal Search"
      @keydown="handleKeyDown"
    >
      <div class="search-input-box">
        <input
          ref="inputRef"
          v-model="query"
          type="text"
          class="search-input"
          :placeholder="t('terminal.search.placeholder')"
          @input="onInput"
        />

        <!-- Match Counter -->
        <span v-if="query" class="match-count">
          {{ resultCount > 0 ? `${resultIndex + 1}/${resultCount}` : t('terminal.search.no_results') }}
        </span>

        <!-- Option Toggles -->
        <div class="search-options">
          <button
            type="button"
            class="option-btn"
            :class="{ active: matchCase }"
            :title="t('terminal.search.match_case')"
            @click="toggleMatchCase"
          >
            Aa
          </button>
          <button
            type="button"
            class="option-btn"
            :class="{ active: matchWholeWord }"
            :title="t('terminal.search.match_whole_word')"
            @click="toggleMatchWholeWord"
          >
            \b
          </button>
          <button
            type="button"
            class="option-btn"
            :class="{ active: useRegex }"
            :title="t('terminal.search.use_regex')"
            @click="toggleUseRegex"
          >
            .*
          </button>
        </div>
      </div>

      <!-- Navigation & Close Buttons -->
      <div class="search-actions">
        <button
          type="button"
          class="action-btn"
          :title="t('terminal.search.previous')"
          :disabled="!query || resultCount === 0"
          @click="handleFindPrevious"
        >
          <svg viewBox="0 0 16 16" width="14" height="14" fill="currentColor">
            <path d="M8 4.5l-5 5h10l-5-5z" />
          </svg>
        </button>
        <button
          type="button"
          class="action-btn"
          :title="t('terminal.search.next')"
          :disabled="!query || resultCount === 0"
          @click="handleFindNext"
        >
          <svg viewBox="0 0 16 16" width="14" height="14" fill="currentColor">
            <path d="M8 11.5l5-5H3l5 5z" />
          </svg>
        </button>
        <button
          type="button"
          class="action-btn close-btn"
          :title="t('terminal.search.close')"
          @click="emit('close')"
        >
          <svg viewBox="0 0 16 16" width="14" height="14" fill="currentColor">
            <path d="M3.72 3.72a.75.75 0 011.06 0L8 6.94l3.22-3.22a.75.75 0 111.06 1.06L9.06 8l3.22 3.22a.75.75 0 11-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 01-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 010-1.06z" />
          </svg>
        </button>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.terminal-search-widget {
  position: absolute;
  top: 8px;
  right: 18px;
  z-index: 30;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  background: var(--vscode-editorWidget-background, #252526);
  border: 1px solid var(--vscode-editorWidget-border, #454545);
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-size: 12px;
  color: var(--vscode-editorWidget-foreground, #cccccc);
}

.search-input-box {
  display: flex;
  align-items: center;
  position: relative;
  background: var(--vscode-input-background, #3c3c3c);
  border: 1px solid var(--vscode-input-border, #3c3c3c);
  border-radius: 3px;
  padding: 2px 4px;
}

.search-input-box:focus-within {
  border-color: var(--vscode-focusBorder, #007fd4);
}

.search-input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--vscode-input-foreground, #cccccc);
  font-size: 12px;
  width: 150px;
  padding: 2px 4px;
}

.search-input::placeholder {
  color: var(--vscode-input-placeholderForeground, #8c8c8c);
}

.match-count {
  font-size: 11px;
  color: var(--vscode-descriptionForeground, #999999);
  padding: 0 4px;
  white-space: nowrap;
}

.search-options {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-left: 4px;
}

.option-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 2px;
  color: var(--vscode-editorWidget-foreground, #aaaaaa);
  font-size: 11px;
  font-family: monospace;
  font-weight: bold;
  padding: 1px 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.option-btn:hover {
  background: var(--vscode-toolbar-hoverBackground, rgba(90, 93, 94, 0.31));
  color: #ffffff;
}

.option-btn.active {
  background: var(--vscode-inputOption-activeBackground, rgba(0, 127, 212, 0.4));
  border-color: var(--vscode-inputOption-activeBorder, #007fd4);
  color: var(--vscode-inputOption-activeForeground, #ffffff);
}

.search-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  background: transparent;
  border: none;
  border-radius: 3px;
  color: var(--vscode-icon-foreground, #c5c5c5);
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn:hover:not(:disabled) {
  background: var(--vscode-toolbar-hoverBackground, rgba(90, 93, 94, 0.31));
  color: #ffffff;
}

.action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.close-btn:hover {
  color: #f87171;
}

/* Slide Transition */
.search-slide-enter-active,
.search-slide-leave-active {
  transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.18s ease;
}

.search-slide-enter-from,
.search-slide-leave-to {
  transform: translateY(-8px);
  opacity: 0;
}
</style>
