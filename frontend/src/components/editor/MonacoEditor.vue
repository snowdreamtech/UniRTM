<template>
  <div class="monaco-wrapper">
    <div v-if="isLoading" class="monaco-loading-overlay">
      <span class="codicon codicon-loading codicon-modifier-spin"></span>
      <span>加载编辑器中...</span>
    </div>
    <div v-else-if="mountError" class="monaco-error-overlay">
      <span class="codicon codicon-error error-icon"></span>
      <div class="error-msg">
        <p class="error-title">编辑器加载异常</p>
        <p class="error-detail">{{ mountError }}</p>
      </div>
    </div>
    <div ref="editorContainerRef" class="monaco-host-container"></div>
    <!-- Inline Diff Peek Widget -->
    <div
      v-if="activeDiffPeek"
      class="inline-diff-peek"
      :style="{ top: `${activeDiffPeek.top}px`, left: `${activeDiffPeek.left}px` }"
      @click.stop
    >
      <div class="peek-header">
        <span class="peek-title">
          <span class="codicon codicon-history"></span>
          原版本代码 (第 {{ activeDiffPeek.diff.startLineNumber }} 行)
        </span>
        <div class="peek-actions">
          <button class="peek-btn stage" title="暂存此代码块改动" @click="stageActiveDiffRange">
            <span class="codicon codicon-add"></span>
            暂存此块
          </button>
          <button class="peek-btn revert" title="回滚该代码块至原版 (Ctrl+Z 可撤销)" @click="revertActiveDiffRange">
            <span class="codicon codicon-discard"></span>
            还原此块
          </button>
          <button class="peek-btn close" title="关闭" @click="closeDiffPeek">
            <span class="codicon codicon-close"></span>
          </button>
        </div>
      </div>
      <div class="peek-content">
        <pre v-if="activeDiffPeek.diff.originalText" class="peek-pre deleted-code">{{ activeDiffPeek.diff.originalText }}</pre>
        <span v-else class="peek-empty-tip">(新增代码，原版本无此内容)</span>
      </div>
    </div>
    <!-- Merge Conflict Banner Widget -->
    <div v-if="detectedConflicts.length > 0 && activeConflict" class="conflict-banner">
      <div class="conflict-banner-info">
        <span class="codicon codicon-warning conflict-icon"></span>
        <span class="conflict-text">
          代码合并冲突 (第 {{ activeConflictIdx + 1 }}/{{ detectedConflicts.length }} 处 · 第 {{ activeConflict.startLine }} 行)
        </span>
      </div>
      <div class="conflict-actions">
        <button class="conflict-action-btn current" title="采用当前分支的修改" @click="resolveActiveConflict('current')">
          采用当前更改 ({{ activeConflict.currentLabel }})
        </button>
        <button class="conflict-action-btn incoming" title="采用传入分支的修改" @click="resolveActiveConflict('incoming')">
          采用传入更改 ({{ activeConflict.incomingLabel }})
        </button>
        <button class="conflict-action-btn both" title="同时保留双方更改" @click="resolveActiveConflict('both')">
          保留两者
        </button>
        <div class="conflict-nav" v-if="detectedConflicts.length > 1">
          <button class="conflict-nav-btn" title="上一处冲突" @click="prevConflict">
            <span class="codicon codicon-arrow-up"></span>
          </button>
          <button class="conflict-nav-btn" title="下一处冲突" @click="nextConflict">
            <span class="codicon codicon-arrow-down"></span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from "vue";
import type * as MonacoType from "monaco-editor";
import "monaco-editor/min/vs/editor/editor.main.css";
import { loadMonaco, getLanguageForFile } from "../../services/editor/monacoLoader";
import { computeLineDiff, type DiffLineChange } from "../../utils/diffGutter";
import type { EditorTab } from "../../composables/useEditorTabs";
import { useGit } from "../../composables/useGit";
import {
  parseMergeConflicts,
  getConflictResolutionText,
  type MergeConflictRegion,
  type ConflictResolution,
} from "../../utils/mergeConflict";

const props = defineProps<{
  tab: EditorTab;
}>();

const emit = defineEmits<{
  (e: "change", content: string): void;
  (e: "cursor-change", line: number, column: number): void;
  (e: "save"): void;
}>();

const editorContainerRef = ref<HTMLDivElement | null>(null);
const isLoading = ref<boolean>(true);
const mountError = ref<string>("");

let monaco: typeof MonacoType | null = null;
let editorInstance: MonacoType.editor.IStandaloneCodeEditor | null = null;
let resizeObserver: ResizeObserver | null = null;
let gutterDecorationIds: string[] = [];
let diffDebounceTimer: ReturnType<typeof setTimeout> | null = null;

interface ActiveDiffPeek {
  diff: DiffLineChange;
  top: number;
  left: number;
}

const activeDiffPeek = ref<ActiveDiffPeek | null>(null);
let currentDiffs: DiffLineChange[] = [];

const detectedConflicts = ref<MergeConflictRegion[]>([]);
const activeConflictIdx = ref<number>(0);

const activeConflict = computed(() => {
  if (detectedConflicts.value.length === 0) return null;
  const idx = Math.min(Math.max(0, activeConflictIdx.value), detectedConflicts.value.length - 1);
  return detectedConflicts.value[idx];
});

function scanConflicts(): void {
  if (!editorInstance) return;
  const model = editorInstance.getModel();
  if (!model) return;
  const text = model.getValue();
  detectedConflicts.value = parseMergeConflicts(text);
  if (activeConflictIdx.value >= detectedConflicts.value.length) {
    activeConflictIdx.value = Math.max(0, detectedConflicts.value.length - 1);
  }
}

function scrollToActiveConflict(): void {
  if (!editorInstance || !activeConflict.value) return;
  editorInstance.revealLineInCenter(activeConflict.value.startLine);
}

function nextConflict(): void {
  if (detectedConflicts.value.length <= 1) return;
  activeConflictIdx.value = (activeConflictIdx.value + 1) % detectedConflicts.value.length;
  scrollToActiveConflict();
}

function prevConflict(): void {
  if (detectedConflicts.value.length <= 1) return;
  activeConflictIdx.value = (activeConflictIdx.value - 1 + detectedConflicts.value.length) % detectedConflicts.value.length;
  scrollToActiveConflict();
}

function resolveActiveConflict(resolution: ConflictResolution): void {
  if (!editorInstance || !monaco || !activeConflict.value) return;
  const conflict = activeConflict.value;
  const model = editorInstance.getModel();
  if (!model) return;

  const replaceText = getConflictResolutionText(conflict, resolution);
  const totalLines = model.getLineCount();

  let range: MonacoType.Range;
  let textToInsert = replaceText;
  if (conflict.endLine < totalLines) {
    range = new monaco.Range(conflict.startLine, 1, conflict.endLine + 1, 1);
    textToInsert = replaceText ? `${replaceText}\n` : "";
  } else {
    range = new monaco.Range(conflict.startLine, 1, conflict.endLine, model.getLineLength(conflict.endLine) + 1);
  }

  editorInstance.executeEdits("resolve-conflict", [
    {
      range,
      text: textToInsert,
      forceMoveMarkers: true,
    },
  ]);

  emit("change", model.getValue());
  scheduleGutterDiffUpdate();
  scanConflicts();
}

function closeDiffPeek(): void {
  activeDiffPeek.value = null;
}

function updateGutterDecorations(): void {
  if (!editorInstance || !monaco) return;
  const model = editorInstance.getModel();
  if (!model) return;

  const originalText = props.tab.originalContent ?? "";
  const currentText = model.getValue();

  if (originalText === currentText) {
    gutterDecorationIds = editorInstance.deltaDecorations(gutterDecorationIds, []);
    return;
  }

  const diffs = computeLineDiff(originalText, currentText);
  currentDiffs = diffs;
  const newDecorations: MonacoType.editor.IModelDeltaDecoration[] = diffs.map((diff) => ({
    range: new monaco!.Range(diff.startLineNumber, 1, diff.endLineNumber, 1),
    options: {
      isWholeLine: true,
      linesDecorationsClassName: `gutter-diff-${diff.type}`,
    },
  }));

  gutterDecorationIds = editorInstance.deltaDecorations(gutterDecorationIds, newDecorations);
}

function revertActiveDiffRange(): void {
  if (!activeDiffPeek.value || !editorInstance || !monaco) return;
  const { diff } = activeDiffPeek.value;
  const model = editorInstance.getModel();
  if (!model) return;

  if (diff.type === "added") {
    const startLine = diff.startLineNumber;
    const endLine = diff.endLineNumber;
    const lineCount = model.getLineCount();
    let range: MonacoType.Range;
    if (endLine < lineCount) {
      range = new monaco.Range(startLine, 1, endLine + 1, 1);
    } else if (startLine > 1) {
      const prevLineLen = model.getLineLength(startLine - 1);
      range = new monaco.Range(startLine - 1, prevLineLen + 1, endLine, model.getLineLength(endLine) + 1);
    } else {
      range = new monaco.Range(startLine, 1, endLine, model.getLineLength(endLine) + 1);
    }
    editorInstance.executeEdits("revert-diff", [
      {
        range,
        text: "",
        forceMoveMarkers: true,
      },
    ]);
  } else if (diff.type === "modified") {
    const startLine = diff.startLineNumber;
    const endLine = diff.endLineNumber;
    const range = new monaco.Range(startLine, 1, endLine, model.getLineLength(endLine) + 1);
    editorInstance.executeEdits("revert-diff", [
      {
        range,
        text: diff.originalText ?? "",
        forceMoveMarkers: true,
      },
    ]);
  } else if (diff.type === "deleted") {
    const line = diff.startLineNumber;
    const range = new monaco.Range(line, 1, line, 1);
    editorInstance.executeEdits("revert-diff", [
      {
        range,
        text: (diff.originalText ?? "") + "\n",
        forceMoveMarkers: true,
      },
    ]);
  }

  closeDiffPeek();
  emit("change", model.getValue());
  scheduleGutterDiffUpdate();
}

async function stageActiveDiffRange(): Promise<void> {
  if (!activeDiffPeek.value || !props.tab?.path) return;
  const { diff } = activeDiffPeek.value;
  const filePath = props.tab.path;

  const { stageChunk } = useGit();
  const ok = await stageChunk(filePath, diff);
  if (ok) {
    if (editorInstance) {
      const model = editorInstance.getModel();
      if (model) {
        props.tab.originalContent = model.getValue();
      }
    }
    closeDiffPeek();
    updateGutterDecorations();
  }
}

function scheduleGutterDiffUpdate(): void {
  if (diffDebounceTimer) clearTimeout(diffDebounceTimer);
  diffDebounceTimer = setTimeout(() => {
    updateGutterDecorations();
  }, 100);
}

import { localizeEditorActions, setupEditorContextMenuLocalization } from "../../services/editor/monacoLocalization";

function initMonacoThemes(m: typeof MonacoType): void {
  // 浅色编辑器 -> 深色反差右键菜单
  m.editor.defineTheme("unirtm-light", {
    base: "vs",
    inherit: true,
    rules: [],
    colors: {
      "menu.background": "#1e1e1e",
      "menu.foreground": "#cccccc",
      "menu.selectionBackground": "#04395e",
      "menu.selectionForeground": "#ffffff",
      "menu.selectionBorder": "transparent",
      "menu.separatorBackground": "#333333",
      "menu.border": "#3c3c3c",
      "widget.shadow": "rgba(0, 0, 0, 0.55)",
      "editorWidget.background": "#ffffff",
      "editorWidget.border": "#cbd5e1",
      "editorWidget.foreground": "#1e293b",
      "editorSuggestWidget.background": "#ffffff",
      "editorSuggestWidget.border": "#cbd5e1",
      "editorSuggestWidget.selectedBackground": "#e0f2fe",
      "editorSuggestWidget.highlightForeground": "#0060c0",
      "editorHoverWidget.background": "#ffffff",
      "editorHoverWidget.border": "#cbd5e1",
    },
  });

  // 深色编辑器 -> 浅色反差右键菜单
  m.editor.defineTheme("unirtm-dark", {
    base: "vs-dark",
    inherit: true,
    rules: [],
    colors: {
      "menu.background": "#ffffff",
      "menu.foreground": "#1e293b",
      "menu.selectionBackground": "#0060c0",
      "menu.selectionForeground": "#ffffff",
      "menu.selectionBorder": "transparent",
      "menu.separatorBackground": "#e2e8f0",
      "menu.border": "#cbd5e1",
      "widget.shadow": "rgba(0, 0, 0, 0.35)",
      "editorWidget.background": "#1e1e1e",
      "editorWidget.border": "#3c3c3c",
      "editorWidget.foreground": "#cccccc",
      "editorSuggestWidget.background": "#1e1e1e",
      "editorSuggestWidget.border": "#3c3c3c",
      "editorSuggestWidget.selectedBackground": "#04395e",
      "editorSuggestWidget.highlightForeground": "#38bdf8",
      "editorHoverWidget.background": "#1e1e1e",
      "editorHoverWidget.border": "#3c3c3c",
    },
  });
}

function applyTheme(): void {
  if (!monaco) return;
  const isLight = document.documentElement.getAttribute("data-theme") === "light";
  monaco.editor.setTheme(isLight ? "unirtm-light" : "unirtm-dark");
}

function getOrCreateModel(tab: EditorTab): MonacoType.editor.ITextModel | null {
  if (!monaco) return null;
  const uri = monaco.Uri.file(tab.path);
  let model = monaco.editor.getModel(uri);
  const lang = tab.language || getLanguageForFile(tab.name);

  if (!model) {
    model = monaco.editor.createModel(tab.content, lang, uri);
  } else {
    monaco.editor.setModelLanguage(model, lang);
    if (!tab.isDirty && model.getValue() !== tab.content) {
      model.setValue(tab.content);
    }
  }

  return model;
}

onMounted(async () => {
  try {
    monaco = await loadMonaco();
    initMonacoThemes(monaco);
    applyTheme();

    if (!editorContainerRef.value) return;

    const model = getOrCreateModel(props.tab);

    editorInstance = monaco.editor.create(editorContainerRef.value, {
      model,
      automaticLayout: true,
      fontSize: 13,
      fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
      minimap: { enabled: true, maxColumn: 80 },
      scrollBeyondLastLine: false,
      renderWhitespace: "selection",
      lineNumbers: "on",
      tabSize: 2,
      cursorBlinking: "smooth",
      smoothScrolling: true,
      contextmenu: true,
      fixedOverflowWidgets: true,
      useShadowDOM: false,
    });

    localizeEditorActions(editorInstance);
    setupEditorContextMenuLocalization(editorInstance, editorContainerRef.value);
    scanConflicts();

    // Content change
    editorInstance.onDidChangeModelContent(() => {
      if (editorInstance) {
        emit("change", editorInstance.getValue());
        scheduleGutterDiffUpdate();
        scanConflicts();
      }
    });

    // Mouse down on gutter diff decorations to show peek
    editorInstance.onMouseDown((e) => {
      if (
        e.target.type === monaco!.editor.MouseTargetType.GUTTER_LINE_DECORATIONS ||
        e.target.type === monaco!.editor.MouseTargetType.GUTTER_GLYPH_MARGIN ||
        e.target.type === monaco!.editor.MouseTargetType.GUTTER_LINE_NUMBERS
      ) {
        const line = e.target.position?.lineNumber;
        if (line) {
          const matchingDiff = currentDiffs.find(
            (d) => line >= d.startLineNumber && line <= d.endLineNumber
          );
          if (matchingDiff) {
            const browserEvt = e.event.browserEvent as MouseEvent;
            const containerRect = editorContainerRef.value?.getBoundingClientRect();
            const top = (browserEvt.clientY - (containerRect?.top || 0)) + 6;
            activeDiffPeek.value = {
              diff: matchingDiff,
              top: Math.max(16, Math.min(top, (containerRect?.height || 400) - 130)),
              left: 56,
            };
            return;
          }
        }
      }
      if (activeDiffPeek.value) {
        activeDiffPeek.value = null;
      }
    });

    // Cursor change
    editorInstance.onDidChangeCursorPosition((e) => {
      emit("cursor-change", e.position.lineNumber, e.position.column);
    });

    // Ctrl+S
    editorInstance.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      emit("save");
    });

    // Resize observer
    resizeObserver = new ResizeObserver(() => {
      editorInstance?.layout();
    });
    resizeObserver.observe(editorContainerRef.value);

    // Immediate layout pass
    setTimeout(() => {
      editorInstance?.layout();
    }, 50);

    // Initial position
    const pos = editorInstance.getPosition();
    if (pos) {
      emit("cursor-change", pos.lineNumber, pos.column);
    }
  } catch (err: any) {
    console.error("[MonacoEditor] Mount failed:", err);
    mountError.value = String(err?.message || err);
  } finally {
    isLoading.value = false;
  }
});

// Watch tab switch
watch(
  () => props.tab.id,
  () => {
    if (!editorInstance || !monaco) return;
    const model = getOrCreateModel(props.tab);
    if (model) {
      editorInstance.setModel(model);
      editorInstance.layout();
      updateGutterDecorations();
      const pos = editorInstance.getPosition();
      if (pos) {
        emit("cursor-change", pos.lineNumber, pos.column);
      }
      editorInstance.focus();
    }
  }
);

// Watch tab content updates
watch(
  () => props.tab.content,
  (newContent) => {
    if (!editorInstance || !monaco) return;
    const model = editorInstance.getModel();
    if (model && model.getValue() !== newContent && !props.tab.isDirty) {
      model.setValue(newContent);
    }
    scheduleGutterDiffUpdate();
  }
);

// Watch external content reset (e.g., save or re-fetch)
watch(
  () => props.tab.originalContent,
  (newOrig) => {
    if (!editorInstance || props.tab.isDirty) return;
    const model = editorInstance.getModel();
    if (model && model.getValue() !== newOrig) {
      model.setValue(newOrig);
    }
    updateGutterDecorations();
  }
);

// Watch theme change attribute on <html>
const themeObserver = new MutationObserver(() => {
  applyTheme();
});

if (typeof document !== "undefined") {
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ["data-theme", "style"],
  });
}

onBeforeUnmount(() => {
  themeObserver.disconnect();
  resizeObserver?.disconnect();
  editorInstance?.dispose();
  editorInstance = null;
});
</script>

<style scoped>
.monaco-wrapper {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--bg-color);
}

.monaco-host-container {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.monaco-loading-overlay,
.monaco-error-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  background-color: var(--bg-color);
  color: var(--text-muted);
  font-size: 13px;
  z-index: 10;
}

.monaco-error-overlay {
  flex-direction: column;
  padding: 24px;
  text-align: center;
}

.error-icon {
  font-size: 28px;
  color: #ef4444;
}

.error-title {
  font-weight: 600;
  color: var(--text-main);
  margin-bottom: 4px;
}

.error-detail {
  font-size: 12px;
  color: var(--text-muted);
  font-family: monospace;
}

:deep(.gutter-diff-added) {
  border-left: 3px solid #2ea043 !important;
  margin-left: 3px;
  cursor: pointer;
}

:deep(.gutter-diff-modified) {
  border-left: 3px solid #007acc !important;
  margin-left: 3px;
  cursor: pointer;
}

:deep(.gutter-diff-deleted) {
  border-bottom: 2px solid #f85149 !important;
  width: 6px;
  margin-left: 2px;
  cursor: pointer;
}

:root[data-theme="light"] :deep(.gutter-diff-added) {
  border-left-color: #10b981 !important;
}

:root[data-theme="light"] :deep(.gutter-diff-modified) {
  border-left-color: #0284c7 !important;
}

:root[data-theme="light"] :deep(.gutter-diff-deleted) {
  border-bottom-color: #ef4444 !important;
}

/* Inline Diff Peek Widget */
.inline-diff-peek {
  position: absolute;
  z-index: 40;
  min-width: 300px;
  max-width: 520px;
  background-color: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, rgba(128, 128, 128, 0.35));
  border-radius: 6px;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.45);
  overflow: hidden;
  font-size: 12px;
  backdrop-filter: blur(8px);
}

:root[data-theme="light"] .inline-diff-peek {
  background-color: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.peek-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  background-color: rgba(128, 128, 128, 0.1);
  border-bottom: 1px solid var(--card-border, rgba(128, 128, 128, 0.2));
}

.peek-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: var(--text-main);
  font-size: 11px;
}

.peek-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.peek-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  border-radius: 3px;
  padding: 2px 6px;
  font-size: 11px;
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-muted);
  transition: all 0.15s ease;
}

.peek-btn.stage {
  background-color: rgba(46, 160, 67, 0.14);
  color: #2ea043;
  font-weight: 500;
}

.peek-btn.stage:hover {
  background-color: rgba(46, 160, 67, 0.25);
}

.peek-btn.revert {
  background-color: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  font-weight: 500;
}

.peek-btn.revert:hover {
  background-color: rgba(239, 68, 68, 0.22);
}

.peek-btn.close:hover {
  background-color: rgba(128, 128, 128, 0.2);
  color: var(--text-main);
}

.peek-content {
  padding: 8px 10px;
  max-height: 160px;
  overflow-y: auto;
}

.peek-pre {
  margin: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  color: #ef4444;
  background-color: rgba(239, 68, 68, 0.08);
  padding: 4px 6px;
  border-radius: 3px;
  border-left: 2px solid #ef4444;
}

.peek-empty-tip {
  color: var(--text-muted);
  font-style: italic;
  font-size: 11px;
}

.conflict-banner {
  position: absolute;
  top: 10px;
  right: 20px;
  z-index: 25;
  background-color: var(--card-bg, #1e293b);
  border: 1px solid #f59e0b;
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 12px;
  font-size: 12px;
}

:root[data-theme="light"] .conflict-banner {
  background-color: #fffbeb;
  border-color: #f59e0b;
  color: #92400e;
}

.conflict-banner-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: #f59e0b;
}

.conflict-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.conflict-action-btn {
  border: none;
  border-radius: 4px;
  padding: 3px 8px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.conflict-action-btn.current {
  background-color: rgba(59, 130, 246, 0.18);
  color: #3b82f6;
}

.conflict-action-btn.current:hover {
  background-color: rgba(59, 130, 246, 0.3);
}

.conflict-action-btn.incoming {
  background-color: rgba(16, 185, 129, 0.18);
  color: #10b981;
}

.conflict-action-btn.incoming:hover {
  background-color: rgba(16, 185, 129, 0.3);
}

.conflict-action-btn.both {
  background-color: rgba(168, 85, 247, 0.18);
  color: #a855f7;
}

.conflict-action-btn.both:hover {
  background-color: rgba(168, 85, 247, 0.3);
}

.conflict-nav {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-left: 4px;
  border-left: 1px solid rgba(128, 128, 128, 0.25);
  padding-left: 6px;
}

.conflict-nav-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  color: var(--text-muted);
}

.conflict-nav-btn:hover {
  background: rgba(128, 128, 128, 0.2);
  color: var(--text-main);
}
</style>
