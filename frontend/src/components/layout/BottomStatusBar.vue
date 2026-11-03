<template>
  <footer class="status-bar" id="app-status-bar" role="contentinfo">
    <!-- Left Section: Overall Health & Diagnostics -->
    <div class="status-left">
      <!-- Git Branch & Sync Indicator -->
      <div
        class="status-pill status-git"
        id="status-git-branch"
        :title="
          t('statusbar.git_tooltip', {
            repo: activeRepo ? activeRepo.name + ' • ' : '',
            branch: currentBranch,
            upstream: status.upstream ? ' -> ' + status.upstream : '',
            behind: status.behind,
            ahead: status.ahead,
          }) ||
          `Git: ${activeRepo ? activeRepo.name + ' • ' : ''}${currentBranch}${status.upstream ? ' -> ' + status.upstream : ''} (${status.behind}↓ ${status.ahead}↑) • Click to open Source Control`
        "
        @click="$emit('select-tab', 'scm')"
      >
        <span class="status-icon"><TechIcon id="workbench:git" :size="13" :colored="false" /></span>
        <span v-if="repositories.length > 1 && activeRepo" class="status-repo-name font-mono">{{
          activeRepo.name
        }}</span>
        <span class="status-text font-mono"
          >{{ currentBranch }}{{ status.staged.length > 0 || status.unstaged.length > 0 ? "*" : "" }}</span
        >
        <button
          v-if="status.ahead > 0 || status.behind > 0"
          class="git-sync-btn"
          :title="
            t('statusbar.sync_tooltip', { behind: status.behind, ahead: status.ahead }) ||
            `Sync changes: ${status.behind} behind, ${status.ahead} ahead (Click to sync)`
          "
          :disabled="isSyncing"
          @click.stop="handleSync"
        >
          <span class="codicon" :class="isSyncing ? 'codicon-sync codicon-modifier-spin' : 'codicon-sync'"></span>
          <span>{{ status.behind }}↓ {{ status.ahead }}↑</span>
        </button>
      </div>

      <!-- Git Identity Profile Switcher -->
      <GitProfileStatus @open-settings="$emit('open-settings', $event)" />

      <div
        class="status-pill status-ready"
        id="status-health"
        :title="'System & Environment Status: Ready'"
        @click="$emit('select-tab', 'doctor')"
      >
        <span class="status-dot">●</span>
        <span class="status-text">{{ statusReadyText }}</span>
      </div>

      <div class="status-pill status-warnings" id="status-warnings" :title="'0 Security Warnings'">
        <span class="status-icon"><TechIcon id="workbench:warning" :size="14" /></span>
        <span class="status-text">0</span>
      </div>
    </div>

    <!-- Center Section: Active Background Task / Dynamic Notification -->
    <div class="status-center">
      <div v-if="activeTaskName" class="status-task-pill" id="status-running-task" @click="$emit('toggle-drawer')">
        <TechIcon id="workbench:sync" :size="13" class="spin-icon" />
        <span class="task-running-text">{{ runningTaskText }}: {{ activeTaskName }}</span>
      </div>
    </div>

    <!-- Right Section: Scope, Tool Count, Mirror Latency, Lockfile, Version -->
    <div class="status-right">
      <!-- Active Scope Switcher -->
      <button
        class="status-btn"
        id="btn-status-scope"
        :title="
          activeScope === 'user'
            ? t('statusbar.scope_user_tooltip') ||
              'Current Scope: User (~/.unirtm) • Click to switch to Workspace (Alt+S)'
            : t('statusbar.scope_workspace_tooltip') ||
              'Current Scope: Workspace (.unirtm.toml) • Click to switch to User (Alt+S)'
        "
        @click="toggleScope(router)"
      >
        <span class="btn-icon">
          <TechIcon
            :id="activeScope === 'user' ? 'workbench:settings' : 'workbench:folder'"
            :size="13"
            :colored="false"
          />
        </span>
        <span>{{
          activeScope === "user" ? t("settings.scope_user") || "User" : t("settings.scope_workspace") || "Workspace"
        }}</span>
      </button>

      <!-- Active Toolchain Count -->
      <button
        class="status-btn"
        id="btn-status-tools"
        :title="
          activeScope === 'user'
            ? t('statusbar.user_tools_count', { count: 1 }) || '1 tool configured in user global'
            : t('statusbar.workspace_tools_count', { count: 5 }) || '5 tools configured in workspace'
        "
        @click="$emit('select-tab', 'tools')"
      >
        <span class="btn-icon"><TechIcon id="workbench:tools" :size="14" :colored="false" /></span>
        <span>{{ activeScope === "user" ? 1 : 5 }} {{ toolsLabelText }}</span>
      </button>

      <!-- Network Mirror & Latency -->
      <button
        class="status-btn"
        id="btn-status-network"
        :title="'Mirror: npmmirror • Latency: ' + (latencyMs ? latencyMs + 'ms' : 'Ready')"
        @click="$emit('open-network-settings')"
      >
        <span class="btn-icon"><TechIcon id="workbench:runtimes" :size="14" :colored="false" /></span>
        <span>{{ mirrorLabel }}</span>
      </button>

      <!-- Lockfile status -->
      <div class="status-pill" id="status-lockfile" :title="'unirtm.lock is in sync with active dependencies'">
        <span class="btn-icon"><TechIcon id="unirtm.lock" :size="12" /></span>
        <span>{{ lockStatusText }}</span>
      </div>

      <!-- Keyboard Shortcuts Cheatsheet -->
      <button
        class="status-btn"
        id="btn-status-shortcuts"
        :title="
          isMac
            ? t('statusbar.shortcuts_mac') || 'Keyboard Shortcuts (⌘K ⌘S)'
            : t('statusbar.shortcuts_win') || 'Keyboard Shortcuts (Ctrl+K Ctrl+S)'
        "
        @click="$emit('open-shortcuts')"
      >
        <span class="btn-icon"><span class="codicon codicon-keyboard"></span></span>
        <span>{{ t("statusbar.shortcuts") || "Shortcuts" }}</span>
      </button>

      <!-- Core Version -->
      <div class="status-version font-mono" title="UniRTM Core Version">v0.35.0</div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import TechIcon from "../common/TechIcon.vue";
import GitProfileStatus from "../common/GitProfileStatus.vue";
import { useAppScope } from "../../composables/useAppScope";
import { useGit } from "../../composables/useGit";
import { isMac } from "../../utils/platform";
import { t } from "../../i18n";

const router = useRouter();
const { activeScope, toggleScope } = useAppScope();
const { currentBranch, status, isSyncing, sync, activeRepo, repositories } = useGit();

async function handleSync() {
  await sync();
}

const props = defineProps<{
  activeTaskName?: string;
  latencyMs?: number | null;
  currentLang?: string;
}>();

defineEmits<{
  (e: "select-tab", tabId: string): void;
  (e: "open-network-settings"): void;
  (e: "toggle-drawer"): void;
  (e: "open-shortcuts"): void;
  (e: "open-settings", tab?: string): void;
}>();

const statusReadyText = computed(() => t("statusbar.status_ready") || "Ready");
const runningTaskText = computed(() => t("statusbar.status_running") || "Running");
const toolsLabelText = computed(() => t("statusbar.tools_label") || "tools");
const lockStatusText = computed(() => t("statusbar.lock_sync") || "Lock Sync");

const mirrorLabel = computed(() => {
  const lat = props.latencyMs ? ` (${props.latencyMs}ms)` : "";
  return `npmmirror${lat}`;
});
</script>

<style scoped>
.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 26px;
  min-height: 26px;
  background: var(--card-bg);
  border-top: 1px solid var(--card-border);
  padding: 0 0.65rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  user-select: none;
  -webkit-user-select: none;
  box-sizing: border-box;
  z-index: 60;
  backdrop-filter: blur(8px);
}

.status-left,
.status-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.status-git {
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  padding: 1px 6px;
  border-radius: 3px;
  transition: background-color 0.15s ease;
}

.status-git:hover {
  background-color: var(--item-hover-bg, rgba(128, 128, 128, 0.12));
}

.status-repo-name {
  font-weight: 600;
  margin-right: 2px;
  color: var(--text-main);
  opacity: 0.95;
}

.git-sync-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  background: transparent;
  border: none;
  color: inherit;
  font-size: 11px;
  cursor: pointer;
  padding: 0 4px;
  border-radius: 3px;
}

.git-sync-btn:hover:not(:disabled) {
  background: var(--item-hover-bg, rgba(128, 128, 128, 0.15));
}

.status-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
}

.status-ready {
  color: var(--success);
  cursor: pointer;
  font-weight: 500;
}

.status-ready:hover {
  background: rgba(16, 185, 129, 0.12);
}

.status-dot {
  font-size: 0.75rem;
  line-height: 1;
}

.status-warnings {
  color: var(--text-muted);
}

.status-task-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: rgba(0, 229, 255, 0.14);
  color: var(--accent-cyan);
  padding: 0.1rem 0.55rem;
  border-radius: 9999px;
  font-weight: 600;
  cursor: pointer;
  animation: pulse-task 2s infinite ease-in-out;
}

@keyframes pulse-task {
  0%,
  100% {
    opacity: 0.9;
  }
  50% {
    opacity: 0.6;
  }
}

.spin-icon {
  display: inline-block;
  animation: spin 1.2s infinite linear;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.status-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.72rem;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.status-btn:hover {
  background: var(--input-bg);
  color: var(--text-main);
}

.btn-icon {
  font-size: 0.75rem;
}

.status-version {
  font-size: 0.68rem;
  opacity: 0.7;
  padding-left: 0.25rem;
}
</style>
