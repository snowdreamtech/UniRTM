<template>
  <img
    v-if="resolved.type === 'svg' && resolved.svgPath"
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
