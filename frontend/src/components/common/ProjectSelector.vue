<template>
  <div v-if="availableProjects.length > 0" class="project-selector-container" ref="containerRef" @contextmenu.prevent>
    <div class="project-pill-group">
      <!-- Main Dropdown Trigger -->
      <button
        class="project-selector-btn"
        :class="{ 'is-open': isOpen, 'is-single': !isMultiProject }"
        @click="isMultiProject && (isOpen = !isOpen)"
        :title="
          tr('project.current_tooltip', { name: activeProject?.name || 'All' }) ||
          `Active Project: ${activeProject?.name || 'Workspace'}`
        "
      >
        <span class="codicon codicon-package"></span>
        <span class="project-label">{{ activeProject?.name || tr("project.all") || "Workspace" }}</span>
        <span v-if="isMultiProject" class="codicon codicon-chevron-down twistie" :class="{ rotated: isOpen }"></span>
      </button>

      <!-- Pin / Follow Toggle Button -->
      <button
        v-if="isMultiProject"
        class="project-pin-btn"
        :class="{ pinned: isPinned }"
        @click="togglePin"
        :title="
          isPinned
            ? tr('project.pinned_hint') || 'Project locked (Will NOT switch on file open)'
            : tr('project.unpinned_hint') || 'Project auto-follows current open file'
        "
      >
        <span class="codicon" :class="isPinned ? 'codicon-pinned' : 'codicon-pin'"></span>
      </button>
    </div>

    <!-- Dropdown Menu -->
    <transition name="dropdown-fade">
      <div v-if="isOpen" class="project-dropdown-menu">
        <div class="dropdown-header">
          <span>{{ tr("project.select_title") || "WORKSPACE PROJECTS" }}</span>
          <span class="header-count font-mono">{{ availableProjects.length }}</span>
        </div>

        <div class="dropdown-list">
          <button
            v-for="p in availableProjects"
            :key="p.id"
            class="dropdown-item"
            :class="{ active: activeProject?.id === p.id || activeProject?.path === p.path }"
            @click="handleSelect(p)"
          >
            <span class="item-icon codicon" :class="p.isGit ? 'codicon-repo' : 'codicon-folder'"></span>
            <div class="item-details">
              <span class="item-title">{{ p.name }}</span>
              <span v-if="p.path" class="item-sub font-mono">{{ p.path }}</span>
            </div>
            <span
              v-if="activeProject?.id === p.id || activeProject?.path === p.path"
              class="codicon codicon-check active-check"
            ></span>
          </button>
        </div>

        <div class="dropdown-footer">
          <span class="codicon codicon-info"></span>
          <span>{{
            isPinned
              ? tr("project.footer_pinned") || "Locked: click 📌 to auto-follow files"
              : tr("project.footer_unpinned") || "Auto-follows file & Git changes"
          }}</span>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { t } from "../../i18n";
import { useActiveProject, type ProjectItem } from "../../composables/useActiveProject";

const tr = (key: string, params?: Record<string, string | number>): string => (t as any)(key, params);

const { activeProject, isPinned, availableProjects, isMultiProject, loadProjects, setActiveProject, togglePin } =
  useActiveProject();

const isOpen = ref(false);
const containerRef = ref<HTMLElement | null>(null);

function handleSelect(p: ProjectItem) {
  setActiveProject(p, true);
  isOpen.value = false;
}

function handleDocClick(e: MouseEvent) {
  if (!isOpen.value) return;
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    isOpen.value = false;
  }
}

onMounted(() => {
  loadProjects();
  document.addEventListener("click", handleDocClick);
});

onUnmounted(() => {
  document.removeEventListener("click", handleDocClick);
});
</script>

<style scoped>
.project-selector-container {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.project-pill-group {
  display: inline-flex;
  align-items: center;
  background: var(--card-bg, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--card-border, rgba(128, 128, 128, 0.25));
  border-radius: 9999px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.project-pill-group:hover {
  border-color: var(--accent-cyan, #0284c7);
  background: rgba(2, 132, 199, 0.08);
}

.project-selector-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.5rem 0.2rem 0.65rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-main, #e0e0e0);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: color 0.15s ease;
}

.project-selector-btn:hover {
  color: #fff;
}

.project-label {
  font-weight: 600;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.twistie {
  font-size: 0.6875rem;
  transition: transform 0.2s ease;
  opacity: 0.7;
}

.twistie.rotated {
  transform: rotate(180deg);
}

.project-pin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.2rem 0.45rem;
  font-size: 0.6875rem;
  color: var(--text-muted, #888);
  background: transparent;
  border: none;
  border-left: 1px solid var(--card-border, rgba(128, 128, 128, 0.2));
  cursor: pointer;
  transition: all 0.15s ease;
}

.project-pin-btn:hover {
  color: var(--text-main, #fff);
  background: rgba(255, 255, 255, 0.08);
}

.project-pin-btn.pinned {
  color: #58a6ff;
  background: rgba(56, 139, 253, 0.15);
}

/* Dropdown Menu */
.project-dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 240px;
  background: var(--bg-dropdown, #1e1e1e);
  border: 1px solid var(--card-border, #333);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  backdrop-filter: blur(12px);
}

.dropdown-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.75rem;
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--text-muted, #888);
  border-bottom: 1px solid var(--card-border, #2a2a2a);
  background: rgba(0, 0, 0, 0.15);
}

.header-count {
  background: rgba(255, 255, 255, 0.08);
  padding: 1px 6px;
  border-radius: 9999px;
  font-size: 0.625rem;
}

.dropdown-list {
  display: flex;
  flex-direction: column;
  max-height: 280px;
  overflow-y: auto;
  padding: 0.25rem 0;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.45rem 0.75rem;
  background: transparent;
  border: none;
  color: var(--text-muted, #aaa);
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
  width: 100%;
}

.dropdown-item:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-main, #fff);
}

.dropdown-item.active {
  background: rgba(2, 132, 199, 0.15);
  color: var(--accent-cyan, #388bfd);
}

.item-icon {
  font-size: 0.9375rem;
  flex-shrink: 0;
}

.item-details {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.item-title {
  font-size: 0.75rem;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-sub {
  font-size: 0.625rem;
  opacity: 0.7;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.active-check {
  font-size: 0.8125rem;
  color: var(--accent-cyan, #388bfd);
  flex-shrink: 0;
}

.dropdown-footer {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.4rem 0.75rem;
  font-size: 0.6875rem;
  color: var(--text-muted, #777);
  background: rgba(0, 0, 0, 0.2);
  border-top: 1px solid var(--card-border, #2a2a2a);
}

.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: all 0.15s ease;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
