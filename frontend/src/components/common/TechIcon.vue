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
