<template>
  <svg
    v-if="id === 'workbench:delete'"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px`, color: colored ? '#ef4444' : 'inherit' }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path fill-rule="evenodd" clip-rule="evenodd" d="M10 3h3v1h-1v9a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V4H3V3h3V2a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v1zM7 2v1h2V2H7zm2 3H8v7h1V5zm-2 0H6v7h1V5z"/>
  </svg>
  <svg
    v-else-if="id === 'workbench:workspace' || id === 'workbench:folder'"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px`, color: colored ? '#3b82f6' : 'inherit' }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path fill-rule="evenodd" clip-rule="evenodd" d="M14.5 3H7.71l-.85-.85L6.51 2h-5l-.5.5v11l.5.5h13l.5-.5v-10l-.5-.5zm-.5 10H2V3h4.29l.85.85.35.15H14v9z"/>
  </svg>
  <svg
    v-else-if="id === 'workbench:user'"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px`, color: colored ? '#10b981' : 'inherit' }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path fill-rule="evenodd" clip-rule="evenodd" d="M8 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6zm2-3a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm4 8.5v.5H2v-.5c0-2.3 3.1-3.5 6-3.5s6 1.2 6 3.5zm-1 .5c-.4-1.4-2.8-2.5-5-2.5s-4.6 1.1-5 2.5h10z"/>
  </svg>
  <svg
    v-else-if="id === 'workbench:tools'"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px`, color: 'inherit' }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path fill-rule="evenodd" clip-rule="evenodd" d="M14.85 2.15a2 2 0 0 0-2.83 0L9.4 4.77l1.83 1.83 2.62-2.62a1 1 0 0 1 1.41 1.41l-2.62 2.63 1.83 1.83 2.62-2.62a2 2 0 0 0 0-2.83l-1.84-1.87zm-6.15 3.32L2.15 12a2 2 0 0 0 0 2.83l1 1a2 2 0 0 0 2.83 0l6.53-6.55-3.81-3.81zM3.56 14.41a1 1 0 0 1-1.41-1.41l5.44-5.45 1.41 1.41-5.44 5.45z"/>
  </svg>
  <svg
    v-else-if="id === 'workbench:settings'"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px`, color: 'inherit' }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path fill-rule="evenodd" clip-rule="evenodd" d="M9.1 4.4L8.6 2H7.4l-.5 2.4-.7.3-2-1.3-.9.8 1.3 2-.2.7-2.4.5v1.2l2.4.5.3.8-1.3 2 .8.8 2-1.3.8.3.4 2.3h1.2l.5-2.4.8-.3 2 1.3.8-.8-1.3-2 .3-.8 2.3-.4V7.4l-2.4-.5-.3-.7 1.3-2-.8-.9-2 1.3-.7-.2zM8 11a3 3 0 1 1 0-6 3 3 0 0 1 0 6z"/>
  </svg>
  <svg
    v-else-if="id === 'workbench:terminal'"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px`, color: 'inherit' }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path fill-rule="evenodd" clip-rule="evenodd" d="M1 3.5A1.5 1.5 0 0 1 2.5 2h11A1.5 1.5 0 0 1 15 3.5v9a1.5 1.5 0 0 1-1.5 1.5h-11A1.5 1.5 0 0 1 1 12.5v-9zM2.5 3a.5.5 0 0 0-.5.5v9a.5.5 0 0 0 .5.5h11a.5.5 0 0 0 .5-.5v-9a.5.5 0 0 0-.5-.5h-11zM4.146 5.146a.5.5 0 0 1 .708 0l2.5 2.5a.5.5 0 0 1 0 .708l-2.5 2.5a.5.5 0 0 1-.708-.708L6.293 8 4.146 5.854a.5.5 0 0 1 0-.708zM8.5 10.5a.5.5 0 0 1 .5-.5h3a.5.5 0 0 1 0 1H9a.5.5 0 0 1-.5-.5z"/>
  </svg>
  <svg
    v-else-if="id === 'workbench:about'"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px`, color: 'inherit' }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path fill-rule="evenodd" clip-rule="evenodd" d="M8 1.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13zM0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8zm7.25-2a.75.75 0 1 1 1.5 0 .75.75 0 0 1-1.5 0zM7 7.75A.75.75 0 0 1 7.75 7h.5a.75.75 0 0 1 .75.75v3.5a.75.75 0 0 1-1.5 0V8.5h-.25A.75.75 0 0 1 7 7.75z"/>
  </svg>
  <svg
    v-else-if="isNpm"
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
    <rect width="16" height="16" rx="2" fill="#CB3837"/>
    <path fill="#ffffff" d="M3 3h10v10H8V5.5H5.5V13H3V3z"/>
  </svg>
  <svg
    v-else-if="isGithub"
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    viewBox="0 0 16 16"
    fill="currentColor"
    class="tech-icon tech-icon-svg"
    :class="className"
    :style="{ width: `${size}px`, height: `${size}px`, minWidth: `${size}px`, color: colored ? 'var(--text-main, #cbd5e1)' : 'inherit' }"
    role="img"
    :title="id"
    :aria-label="id"
  >
    <path fill-rule="evenodd" clip-rule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
  </svg>
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
  <span
    v-else
    class="tech-icon"
    :class="className"
    :style="iconStyles"
    :title="id"
    role="img"
    :aria-label="id"
  >{{ resolved.character }}</span>
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
