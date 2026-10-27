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
        <span class="codicon codicon-package project-icon"></span>
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
  background: var(--input-bg, rgba(0, 0, 0, 0.15));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  border-radius: 9999px;
  padding: 1px 2px;
  user-select: none;
  -webkit-user-select: none;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.08);
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
}

:root[data-theme="light"] .project-pill-group,
[data-theme="light"] .project-pill-group {
  background: var(--input-bg, #f8fafc);
  border-color: var(--card-border, #cbd5e1);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.05);
}

.project-pill-group:hover {
  border-color: var(--accent-cyan, #0284c7);
}

.project-selector-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: none;
  padding: 2px 8px;
  border-radius: 9999px;
  font-size: 0.72rem;
  font-weight: 500;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
  outline: none;
  line-height: 1.4;
  white-space: nowrap;
}

:root[data-theme="light"] .project-selector-btn,
[data-theme="light"] .project-selector-btn {
  color: var(--text-muted, #475569);
}

.project-selector-btn:hover {
  color: var(--text-main, #f8fafc);
  background: var(--hover-bg, rgba(255, 255, 255, 0.06));
}

:root[data-theme="light"] .project-selector-btn:hover,
[data-theme="light"] .project-selector-btn:hover {
  color: var(--text-main, #0f172a);
  background: var(--hover-bg, rgba(0, 0, 0, 0.05));
}

.project-selector-btn.is-open {
  background: var(--card-bg, #1e1e1e);
  color: var(--accent-cyan, #00e5ff);
  font-weight: 600;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.18);
}

:root[data-theme="light"] .project-selector-btn.is-open,
[data-theme="light"] .project-selector-btn.is-open {
  background: #ffffff;
  color: var(--accent-cyan, #0284c7);
  box-shadow: 0 1px 4px rgba(15, 23, 42, 0.12);
}

.project-icon {
  font-size: 0.8125rem;
  color: var(--accent-cyan, #00e5ff);
  opacity: 0.9;
}

:root[data-theme="light"] .project-icon,
[data-theme="light"] .project-icon {
  color: var(--accent-cyan, #0284c7);
}

.project-label {
  letter-spacing: 0.2px;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.twistie {
  font-size: 0.625rem;
  transition: transform 0.2s ease;
  opacity: 0.75;
}

.twistie.rotated {
  transform: rotate(180deg);
}

.project-pin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 20px;
  border-radius: 9999px;
  font-size: 0.6875rem;
  color: var(--text-muted, #94a3b8);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
  margin-right: 1px;
}

:root[data-theme="light"] .project-pin-btn,
[data-theme="light"] .project-pin-btn {
  color: var(--text-muted, #64748b);
}

.project-pin-btn:hover {
  color: var(--text-main, #f8fafc);
  background: var(--hover-bg, rgba(255, 255, 255, 0.08));
}

:root[data-theme="light"] .project-pin-btn:hover,
[data-theme="light"] .project-pin-btn:hover {
  color: var(--text-main, #0f172a);
  background: var(--hover-bg, rgba(0, 0, 0, 0.06));
}

.project-pin-btn.pinned {
  color: var(--accent-cyan, #00e5ff);
  background: rgba(0, 229, 255, 0.15);
}

:root[data-theme="light"] .project-pin-btn.pinned,
[data-theme="light"] .project-pin-btn.pinned {
  color: var(--accent-cyan, #0284c7);
  background: #e0f2fe;
}

/* Dropdown Popover */
.project-dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 250px;
  background: var(--modal-bg, #0d1626);
  border: 1px solid var(--card-border, rgba(0, 229, 255, 0.2));
  border-radius: 8px;
  box-shadow:
    0 14px 36px rgba(0, 0, 0, 0.5),
    0 4px 12px rgba(0, 0, 0, 0.25);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  backdrop-filter: blur(16px);
}

:root[data-theme="light"] .project-dropdown-menu,
[data-theme="light"] .project-dropdown-menu {
  background: #ffffff;
  border-color: var(--card-border, #cbd5e1);
  box-shadow:
    0 12px 30px rgba(15, 23, 42, 0.12),
    0 4px 10px rgba(15, 23, 42, 0.06);
}

.dropdown-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.75rem;
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--text-muted, #94a3b8);
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  background: var(--header-bg, rgba(255, 255, 255, 0.03));
}

:root[data-theme="light"] .dropdown-header,
[data-theme="light"] .dropdown-header {
  color: var(--text-muted, #64748b);
  border-bottom-color: var(--card-border, #e2e8f0);
  background: var(--header-bg, #f8fafc);
}

.header-count {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-muted, #94a3b8);
  padding: 1px 6px;
  border-radius: 9999px;
  font-size: 0.625rem;
}

:root[data-theme="light"] .header-count,
[data-theme="light"] .header-count {
  background: var(--btn-secondary-bg, #e2e8f0);
  color: var(--text-muted, #475569);
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
  gap: 0.6rem;
  padding: 0.5rem 0.75rem;
  background: transparent;
  border: none;
  color: var(--text-main, #f8fafc);
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
  width: 100%;
}

:root[data-theme="light"] .dropdown-item,
[data-theme="light"] .dropdown-item {
  color: var(--text-main, #0f172a);
}

.dropdown-item:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.06));
}

:root[data-theme="light"] .dropdown-item:hover,
[data-theme="light"] .dropdown-item:hover {
  background: var(--hover-bg, #f1f5f9);
}

.dropdown-item.active {
  background: var(--vscode-menu-selectionBackground, rgba(0, 229, 255, 0.15));
  color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .dropdown-item.active,
[data-theme="light"] .dropdown-item.active {
  background: var(--vscode-menu-selectionBackground, #e0f2fe);
  color: var(--accent-cyan, #0284c7);
}

.item-icon {
  font-size: 0.9375rem;
  flex-shrink: 0;
  color: var(--text-muted, #94a3b8);
}

:root[data-theme="light"] .item-icon,
[data-theme="light"] .item-icon {
  color: var(--text-muted, #64748b);
}

.dropdown-item.active .item-icon {
  color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .dropdown-item.active .item-icon {
  color: var(--accent-cyan, #0284c7);
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
  color: inherit;
}

.item-sub {
  font-size: 0.625rem;
  color: var(--text-muted, #64748b);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:root[data-theme="light"] .item-sub,
[data-theme="light"] .item-sub {
  color: var(--text-muted, #64748b);
}

.dropdown-item.active .item-sub {
  color: var(--accent-cyan, #00e5ff);
  opacity: 0.85;
}

:root[data-theme="light"] .dropdown-item.active .item-sub {
  color: #0369a1;
  opacity: 1;
}

.active-check {
  font-size: 0.8125rem;
  color: var(--accent-cyan, #00e5ff);
  flex-shrink: 0;
}

:root[data-theme="light"] .active-check,
[data-theme="light"] .active-check {
  color: var(--accent-cyan, #0284c7);
}

.dropdown-footer {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.45rem 0.75rem;
  font-size: 0.6875rem;
  color: var(--text-muted, #94a3b8);
  background: var(--header-bg, rgba(0, 0, 0, 0.2));
  border-top: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
}

:root[data-theme="light"] .dropdown-footer,
[data-theme="light"] .dropdown-footer {
  color: var(--text-muted, #64748b);
  background: var(--header-bg, #f8fafc);
  border-top-color: var(--card-border, #e2e8f0);
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
