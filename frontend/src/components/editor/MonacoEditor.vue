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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
import type * as MonacoType from "monaco-editor";
import { loadMonaco, getLanguageForFile } from "../../services/editor/monacoLoader";
import type { EditorTab } from "../../composables/useEditorTabs";

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

function applyTheme(): void {
  if (!monaco) return;
  const isLight = document.documentElement.getAttribute("data-theme") === "light";
  monaco.editor.setTheme(isLight ? "vs" : "vs-dark");
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
    });

    // Content change
    editorInstance.onDidChangeModelContent(() => {
      if (editorInstance) {
        emit("change", editorInstance.getValue());
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
</style>
