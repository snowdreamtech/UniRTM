<template>
  <div class="monaco-diff-wrapper">
    <!-- Top Diff View Toolbar -->
    <div class="diff-view-subbar">
      <div class="diff-subbar-left">
        <span class="codicon codicon-diff"></span>
        <span class="diff-title-text">{{ tab.name }}</span>
        <span v-if="tab.diffMeta?.isStaged" class="diff-badge staged">已暂存</span>
        <span v-else class="diff-badge unstaged">工作区更改</span>
      </div>
      <div class="diff-subbar-right">
        <button
          type="button"
          class="diff-mode-btn"
          :title="isInline ? '切换为并排对比模式' : '切换为行内对比模式'"
          @click="toggleInlineMode"
        >
          <span class="codicon" :class="isInline ? 'codicon-split-horizontal' : 'codicon-list-flat'"></span>
          <span>{{ isInline ? '并排对比' : '行内对比' }}</span>
        </button>
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="isLoading" class="monaco-loading-overlay">
      <span class="codicon codicon-loading codicon-modifier-spin"></span>
      <span>加载差异对比中...</span>
    </div>

    <!-- Error Overlay -->
    <div v-else-if="mountError" class="monaco-error-overlay">
      <span class="codicon codicon-error error-icon"></span>
      <div class="error-msg">
        <p class="error-title">差异对比加载失败</p>
        <p class="error-detail">{{ mountError }}</p>
      </div>
    </div>

    <!-- Monaco Diff Host -->
    <div ref="containerRef" class="monaco-diff-host"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
import type * as MonacoType from "monaco-editor";
import "monaco-editor/min/vs/editor/editor.main.css";
import { loadMonaco, getLanguageForFile } from "../../services/editor/monacoLoader";
import { setupEditorContextMenuLocalization } from "../../services/editor/monacoLocalization";
import type { EditorTab } from "../../composables/useEditorTabs";

const props = defineProps<{
  tab: EditorTab;
}>();

const containerRef = ref<HTMLDivElement | null>(null);
const isLoading = ref<boolean>(true);
const mountError = ref<string>("");
const isInline = ref<boolean>(false);

let monaco: typeof MonacoType | null = null;
let diffEditorInstance: MonacoType.editor.IStandaloneDiffEditor | null = null;
let originalModel: MonacoType.editor.ITextModel | null = null;
let modifiedModel: MonacoType.editor.ITextModel | null = null;
let resizeObserver: ResizeObserver | null = null;

function toggleInlineMode(): void {
  isInline.value = !isInline.value;
  if (diffEditorInstance) {
    diffEditorInstance.updateOptions({
      renderSideBySide: !isInline.value,
    });
  }
}

function updateModels(): void {
  if (!monaco || !diffEditorInstance) return;

  if (originalModel) {
    originalModel.dispose();
    originalModel = null;
  }
  if (modifiedModel) {
    modifiedModel.dispose();
    modifiedModel = null;
  }

  const lang = props.tab.language || getLanguageForFile(props.tab.path);
  const origText = props.tab.diffMeta?.originalContent ?? props.tab.originalContent ?? "";
  const modText = props.tab.diffMeta?.modifiedContent ?? props.tab.content ?? "";

  originalModel = monaco.editor.createModel(origText, lang);
  modifiedModel = monaco.editor.createModel(modText, lang);

  diffEditorInstance.setModel({
    original: originalModel,
    modified: modifiedModel,
  });
}

function getPreferredTheme(): string {
  const isLight = document.documentElement.getAttribute("data-theme") === "light";
  return isLight ? "unirtm-light" : "unirtm-dark";
}

onMounted(async () => {
  try {
    isLoading.value = true;
    monaco = await loadMonaco();

    if (!containerRef.value) return;

    diffEditorInstance = monaco.editor.createDiffEditor(containerRef.value, {
      renderSideBySide: !isInline.value,
      readOnly: true,
      automaticLayout: true,
      theme: getPreferredTheme(),
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      lineNumbers: "on",
      fontSize: 13,
      fontFamily: "'Fira Code', 'JetBrains Mono', Consolas, 'Courier New', monospace",
      fixedOverflowWidgets: true,
      useShadowDOM: false,
    });

    const origEd = diffEditorInstance.getOriginalEditor();
    const modEd = diffEditorInstance.getModifiedEditor();
    if (origEd) setupEditorContextMenuLocalization(origEd, containerRef.value);
    if (modEd) setupEditorContextMenuLocalization(modEd, containerRef.value);

    updateModels();
    isLoading.value = false;

    resizeObserver = new ResizeObserver(() => {
      diffEditorInstance?.layout();
    });
    resizeObserver.observe(containerRef.value);
  } catch (err: any) {
    mountError.value = err?.message || "Monaco diff editor failed to initialize";
    isLoading.value = false;
  }
});

watch(
  () => [props.tab.diffMeta?.originalContent, props.tab.diffMeta?.modifiedContent],
  () => {
    updateModels();
  }
);

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  if (originalModel) {
    originalModel.dispose();
    originalModel = null;
  }
  if (modifiedModel) {
    modifiedModel.dispose();
    modifiedModel = null;
  }
  if (diffEditorInstance) {
    diffEditorInstance.dispose();
    diffEditorInstance = null;
  }
});
</script>

<style scoped>
.monaco-diff-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background-color: var(--color-bg-primary, #1e1e1e);
}

.diff-view-subbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 14px;
  background-color: var(--color-bg-secondary, #252526);
  border-bottom: 1px solid var(--color-border, #3c3c3c);
  font-size: 12px;
  user-select: none;
}

.diff-subbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.diff-title-text {
  font-family: monospace;
  font-weight: 500;
  color: var(--color-text-primary, #cccccc);
}

.diff-badge {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 600;
}

.diff-badge.staged {
  background-color: rgba(34, 197, 94, 0.15);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.diff-badge.unstaged {
  background-color: rgba(234, 179, 8, 0.15);
  color: #eab308;
  border: 1px solid rgba(234, 179, 8, 0.3);
}

.diff-mode-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  background-color: transparent;
  border: 1px solid var(--color-border, #3c3c3c);
  color: var(--color-text-primary, #cccccc);
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  transition: all 0.15s ease;
}

.diff-mode-btn:hover {
  background-color: var(--color-bg-hover, #2a2d2e);
  color: #ffffff;
  border-color: var(--color-primary, #007acc);
}

.monaco-diff-host {
  flex: 1;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.monaco-loading-overlay,
.monaco-error-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background-color: rgba(30, 30, 30, 0.85);
  z-index: 10;
  color: var(--color-text-primary, #cccccc);
}

.monaco-error-overlay .error-icon {
  color: #ef4444;
  font-size: 24px;
}
</style>
