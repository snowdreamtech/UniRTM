<template>
  <div
    class="scope-switcher"
    id="app-scope-switcher"
    role="radiogroup"
    aria-label="Scope Switcher"
    :title="shortcutTitle"
  >
    <button
      type="button"
      role="radio"
      :aria-checked="activeScope === 'workspace'"
      class="scope-pill-btn"
      :class="{ active: activeScope === 'workspace' }"
      id="btn-scope-workspace"
      :title="workspaceTooltip"
      @click="handleSelectScope('workspace')"
    >
      <span class="scope-icon">
        <TechIcon id="workbench:folder" :size="13" :colored="false" />
      </span>
      <span class="scope-label">{{ workspaceLabel }}</span>
      <span v-if="activeScope === 'workspace'" class="scope-dot"></span>
    </button>

    <button
      type="button"
      role="radio"
      :aria-checked="activeScope === 'user'"
      class="scope-pill-btn"
      :class="{ active: activeScope === 'user' }"
      id="btn-scope-user"
      :title="userTooltip"
      @click="handleSelectScope('user')"
    >
      <span class="scope-icon">
        <TechIcon id="workbench:settings" :size="13" :colored="false" />
      </span>
      <span class="scope-label">{{ userLabel }}</span>
      <span v-if="activeScope === 'user'" class="scope-dot"></span>
    </button>

    <span class="scope-shortcut-hint" title="快捷键: Alt+S 快速切换作用域">Alt+S</span>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import TechIcon from "./TechIcon.vue";
import { useAppScope, type AppScope } from "../../composables/useAppScope";
import { t } from "../../i18n";

const router = useRouter();
const { activeScope, setScope, toggleScope } = useAppScope();

const workspaceLabel = computed(() => t("settings.scope_workspace") || "工作区");
const userLabel = computed(() => t("settings.scope_user") || "用户");

const workspaceTooltip = computed(
  () => "工作区环境：配置当前项目专属环境 (.unirtm.toml)"
);
const userTooltip = computed(
  () => "用户全局环境：配置当前用户所有项目共用的默认兜底环境 (~/.unirtm)"
);
const shortcutTitle = computed(
  () => `当前作用域: ${activeScope.value === "user" ? userLabel.value : workspaceLabel.value} (按 Alt+S 快速切换)`
);

function handleSelectScope(scope: AppScope) {
  setScope(scope, router);
}

function handleKeydown(e: KeyboardEvent) {
  // Alt+S or Option+S
  if (e.altKey && (e.key === "s" || e.key === "S" || e.code === "KeyS")) {
    e.preventDefault();
    toggleScope(router);
  }
}

onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
});
</script>

<style scoped>
.scope-switcher {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  background: var(--input-bg, rgba(0, 0, 0, 0.15));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  padding: 2px 3px;
  border-radius: 9999px;
  position: relative;
  user-select: none;
  -webkit-user-select: none;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.1);
}

.scope-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: none;
  padding: 2px 9px;
  border-radius: 9999px;
  font-size: 0.72rem;
  font-weight: 500;
  color: var(--text-muted, #888);
  cursor: pointer;
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
  outline: none;
  line-height: 1.4;
  white-space: nowrap;
}

.scope-pill-btn:hover {
  color: var(--text-main, #eee);
  background: rgba(255, 255, 255, 0.05);
}

.scope-pill-btn.active {
  background: var(--card-bg, #1e1e1e);
  color: var(--accent-cyan, #0284c7);
  font-weight: 600;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.18), 0 0 0 1px var(--card-border, rgba(255, 255, 255, 0.1));
}

.scope-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  opacity: 0.85;
}

.scope-pill-btn.active .scope-icon {
  opacity: 1;
}

.scope-label {
  letter-spacing: 0.2px;
}

.scope-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--accent-cyan, #0284c7);
  box-shadow: 0 0 6px var(--accent-cyan, #0284c7);
}

.scope-shortcut-hint {
  font-size: 0.62rem;
  font-family: var(--font-mono, monospace);
  padding: 1px 5px;
  margin-left: 2px;
  color: var(--text-muted, #777);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 4px;
  opacity: 0.7;
}

.scope-switcher:hover .scope-shortcut-hint {
  opacity: 1;
}
</style>
