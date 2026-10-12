<template>
  <div class="custom-select-container" ref="containerRef" :class="{ open: isOpen, disabled: props.disabled }">
    <div class="select-trigger" :class="{ disabled: props.disabled }" @click="!props.disabled && toggleOpen()">
      <span class="selected-label">{{ selectedOption?.label || modelValue }}</span>
      <span class="chevron-icon">▾</span>
    </div>

    <transition name="dropdown-fade">
      <div v-if="isOpen" class="select-dropdown-menu">
        <div
          v-for="opt in options"
          :key="opt.value"
          class="dropdown-item"
          :class="{ selected: opt.value === modelValue }"
          @click="selectOption(opt.value)"
        >
          <span class="item-label">{{ opt.label }}</span>
          <span v-if="opt.value === modelValue" class="check-icon">✓</span>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";

export interface SelectOption {
  value: string;
  label: string;
}

const props = defineProps<{
  modelValue: string;
  options: SelectOption[];
  disabled?: boolean;
}>();

const emit = defineEmits(["update:modelValue", "change"]);

const isOpen = ref(false);
const containerRef = ref<HTMLElement | null>(null);

const selectedOption = computed(() => {
  return props.options.find((o) => o.value === props.modelValue);
});

function toggleOpen() {
  isOpen.value = !isOpen.value;
}

function selectOption(val: string) {
  emit("update:modelValue", val);
  emit("change", val);
  isOpen.value = false;
}

function handleClickOutside(e: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    isOpen.value = false;
  }
}

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>

<style scoped>
.custom-select-container {
  position: relative;
  width: 100%;
  user-select: none;
}

.select-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  color: var(--text-main);
  padding: 0.55rem 0.75rem;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.select-trigger:hover {
  border-color: var(--card-border-active);
}

.select-trigger.disabled {
  opacity: 0.55;
  cursor: not-allowed;
  pointer-events: none;
}

.custom-select-container.open {
  z-index: 1000;
}

.custom-select-container.open .select-trigger {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 10px var(--accent-cyan-glow);
}

.chevron-icon {
  font-size: 0.8rem;
  color: var(--text-muted);
  transition: transform 0.2s ease;
}

.custom-select-container.open .chevron-icon {
  transform: rotate(180deg);
  color: var(--accent-cyan);
}

.select-dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  z-index: 10000;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.35);
  overflow: hidden;
  padding: 4px;
  max-height: 220px;
  overflow-y: auto;
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
}

.dropdown-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  font-size: 0.85rem;
  color: var(--text-main);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.dropdown-item:hover {
  background: rgba(0, 229, 255, 0.1);
  color: var(--accent-cyan);
}

.dropdown-item.selected {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  font-weight: 600;
}

.check-icon {
  font-weight: bold;
  font-size: 0.85rem;
}

/* Transitions */
.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: all 0.18s ease;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* Light Theme Enhancements */
[data-theme="light"] .select-trigger {
  background: #ffffff;
  color: #0f172a;
  border-color: #cbd5e1;
  font-weight: 600;
}

[data-theme="light"] .select-dropdown-menu {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 10px 25px rgba(15, 23, 42, 0.15);
}

[data-theme="light"] .dropdown-item {
  color: #0f172a;
  font-weight: 600;
}

[data-theme="light"] .dropdown-item:hover {
  background: #f0f9ff;
  color: #0284c7;
}

[data-theme="light"] .dropdown-item.selected {
  background: #e0f2fe;
  color: #0284c7;
  font-weight: 700;
}
</style>
