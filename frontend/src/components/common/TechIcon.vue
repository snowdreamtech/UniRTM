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

const resolved = computed(() => {
  return resolveIcon(props.id);
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
