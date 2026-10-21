<template>
  <!-- 1. Tech Brand Icons: NPM -->
  <svg
    v-if="isNpm"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px` }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <rect width="16" height="16" rx="2" fill="#CB3837" />
    <path fill="#ffffff" d="M3 3h10v10H8V5.5H5.5V13H3V3z" />
  </svg>

  <!-- 2. Tech Brand Icons: GitHub -->
  <svg
    v-else-if="isGithub"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{
      width: `${size}px`,
      height: `${size}px`,
      minWidth: `${size}px`,
      color: colored ? 'var(--text-main, #cbd5e1)' : 'inherit',
    }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path
      fill-rule="evenodd"
      clip-rule="evenodd"
      d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"
    />
  </svg>

  <!-- 3. Official VS Code Product Icon Theme (@vscode/codicons) -->
  <span
    v-else-if="codiconClass"
    :class="['codicon', codiconClass, className]"
    :style="codiconStyles"
    :title="id"
    role="img"
    :aria-label="id"
  ></span>

  <!-- 4. Dynamic Theme SVG / Custom Icon Theme File -->
  <img
    v-else-if="resolved.type === 'svg' && resolved.svgPath"
    :src="resolved.svgPath"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="svgStyles"
    :title="id"
    :alt="id"
    role="img"
  />

  <!-- 5. VS Code File Icon Theme (Seti UI baseline font) -->
  <span v-else class="tech-icon" :class="className" :style="iconStyles" :title="id" role="img" :aria-label="id">{{
    resolved.character
  }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useIconTheme } from "../../composables/useIconTheme";

const props = withDefaults(
  defineProps<{
    id: string;
    size?: number;
    colored?: boolean;
    className?: string;
  }>(),
  {
    size: 16,
    colored: true,
    className: "",
  }
);

const { resolveIcon } = useIconTheme();

const WORKBENCH_CODICON_MAP: Record<string, string> = {
  "workbench:dashboard": "codicon-graph",
  "workbench:explorer": "codicon-files",
  "workbench:runtimes": "codicon-server",
  "workbench:tools": "codicon-tools",
  "workbench:tasks": "codicon-tasklist",
  "workbench:marketplace": "codicon-extensions",
  "workbench:doctor": "codicon-pulse",
  "workbench:settings": "codicon-settings-gear",
  "workbench:about": "codicon-info",
  "workbench:terminal": "codicon-terminal",
  "workbench:ports": "codicon-radio-tower",
  "workbench:logs": "codicon-output",
  "workbench:lock": "codicon-lock",
  "workbench:unlock": "codicon-unlock",
  "workbench:environments": "codicon-variable-group",
  "workbench:variable": "codicon-symbol-variable",
  "workbench:key": "codicon-key",
  "workbench:eye": "codicon-eye",
  "workbench:eye-closed": "codicon-eye-closed",
  "workbench:copy": "codicon-copy",
  "workbench:config": "codicon-gear",
  "workbench:success": "codicon-pass",
  "workbench:warning": "codicon-warning",
  "workbench:error": "codicon-error",
  "workbench:info": "codicon-info",
  "workbench:sync": "codicon-sync",
  "workbench:search": "codicon-search",
  "workbench:delete": "codicon-trash",
  "workbench:folder": "codicon-folder",
  "workbench:workspace": "codicon-root-folder",
  "workbench:package": "codicon-package",
  "workbench:globe": "codicon-globe",
  "workbench:speed": "codicon-zap",
  "workbench:star": "codicon-star",
  "workbench:user": "codicon-account",
  "workbench:edit": "codicon-edit",
  "workbench:github": "codicon-github",
  "workbench:git": "codicon-git-branch",
  "workbench:scm": "codicon-source-control",
  "workbench:git-commit": "codicon-check",
  "workbench:git-discard": "codicon-discard",
  "workbench:sparkle": "codicon-sparkle",
  "workbench:cards": "codicon-editor-layout",
  "workbench:list": "codicon-table",
  "workbench:more": "codicon-more",
};

const codiconClass = computed<string | null>(() => {
  const id = props.id || "";
  if (id.startsWith("codicon:")) {
    const name = id.replace("codicon:", "").trim();
    return `codicon-${name}`;
  }
  if (id.startsWith("workbench:")) {
    return WORKBENCH_CODICON_MAP[id] || `codicon-${id.replace("workbench:", "")}`;
  }
  return null;
});

const isNpm = computed(() => {
  const id = props.id?.toLowerCase() || "";
  return id.startsWith("npm:") || id === "npm";
});

const isGithub = computed(() => {
  const id = props.id?.toLowerCase() || "";
  return id.startsWith("github:") || id === "github" || id === "workbench:github";
});

const normalizedId = computed(() => {
  const id = props.id?.toLowerCase() || "";
  if (id.startsWith("go:")) return "go";
  if (id.startsWith("pipx:") || id.startsWith("pip:") || id.startsWith("pypi:")) return "python";
  if (id.startsWith("cargo:")) return "rust";
  return props.id;
});

const resolved = computed(() => {
  return resolveIcon(normalizedId.value);
});

const codiconStyles = computed(() => {
  return {
    fontSize: `${props.size}px`,
    width: `${props.size}px`,
    height: `${props.size}px`,
    lineHeight: `${props.size}px`,
    minWidth: `${props.size}px`,
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    verticalAlign: "middle",
    color:
      props.colored && (props.id === "workbench:delete" || props.id.includes("danger") || props.id.includes("error"))
        ? "var(--error, #ef4444)"
        : props.colored && props.id === "workbench:warning"
          ? "var(--warning, #f59e0b)"
          : props.colored && props.id === "workbench:success"
            ? "var(--success, #10b981)"
            : "inherit",
  };
});

const iconStyles = computed(() => {
  return {
    fontSize: `${props.size}px`,
    width: `${props.size}px`,
    height: `${props.size}px`,
    minWidth: `${props.size}px`,
    color: props.colored ? resolved.value.color || "inherit" : "inherit",
    fontFamily: resolved.value.fontFamily || "seti",
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    lineHeight: "1",
    userSelect: "none" as const,
    WebkitUserSelect: "none" as const,
  };
});

const svgStyles = computed(() => {
  return {
    width: `${props.size}px`,
    height: `${props.size}px`,
    minWidth: `${props.size}px`,
    objectFit: "contain" as const,
    display: "inline-flex",
    verticalAlign: "middle",
  };
});
</script>

<style scoped>
.tech-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  vertical-align: middle;
  font-style: normal;
  font-variant: normal;
}
.tech-icon-svg {
  display: inline-block;
}
</style>
