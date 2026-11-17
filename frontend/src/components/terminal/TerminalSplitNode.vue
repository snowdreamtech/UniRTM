<script lang="ts">
export default {
  name: "TerminalSplitNode",
};
</script>

<script setup lang="ts">
import type { TerminalLayoutNode, TerminalSession } from "../../composables/useTerminalSessions";
import type { ActiveProcessInfo } from "../../composables/useTerminalProcessWatcher";
import XTerminalPane from "./XTerminalPane.vue";
import { t } from "../../i18n";

defineProps<{
  node: TerminalLayoutNode;
  session: TerminalSession;
  activePaneId: string;
  getActiveProcess: (paneId: string) => ActiveProcessInfo | undefined;
}>();

const emit = defineEmits<{
  (e: "setActivePane", paneId: string): void;
  (e: "splitPane", direction: "vertical" | "horizontal", paneId: string): void;
  (e: "closePane", paneId: string): void;
}>();
</script>

<template>
  <!-- Recursive Split Container -->
  <div
    v-if="node.type === 'split'"
    class="terminal-split-container"
    :class="node.direction === 'vertical' ? 'split-vertical' : 'split-horizontal'"
  >
    <div class="terminal-split-child">
      <TerminalSplitNode
        :node="node.children[0]"
        :session="session"
        :active-pane-id="activePaneId"
        :get-active-process="getActiveProcess"
        @set-active-pane="emit('setActivePane', $event)"
        @split-pane="(dir, id) => emit('splitPane', dir, id)"
        @close-pane="emit('closePane', $event)"
      />
    </div>
    <div class="terminal-split-child">
      <TerminalSplitNode
        :node="node.children[1]"
        :session="session"
        :active-pane-id="activePaneId"
        :get-active-process="getActiveProcess"
        @set-active-pane="emit('setActivePane', $event)"
        @split-pane="(dir, id) => emit('splitPane', dir, id)"
        @close-pane="emit('closePane', $event)"
      />
    </div>
  </div>

  <!-- Terminal Leaf Pane -->
  <div
    v-else-if="node.type === 'pane' && session.panes.find((p) => p.id === node.paneId)"
    class="terminal-pane"
    :class="{
      active: node.paneId === activePaneId,
      'is-split': (session.panes.length || 0) > 1,
    }"
    @click="emit('setActivePane', node.paneId)"
  >
    <!-- Pane Top Header / Status Banner -->
    <div class="terminal-pane-banner">
      <div class="pane-banner-left">
        <span class="term-status-dot"></span>
        <span
          v-if="getActiveProcess(node.paneId)"
          class="term-process-badge"
          :style="{
            backgroundColor: `${getActiveProcess(node.paneId)?.color}25`,
            color: getActiveProcess(node.paneId)?.color,
            borderColor: `${getActiveProcess(node.paneId)?.color}60`,
          }"
          :title="`Running: ${getActiveProcess(node.paneId)?.name}`"
        >
          {{ getActiveProcess(node.paneId)?.name }}
        </span>
        <span class="term-shell-info font-mono">
          {{ session.panes.find((p) => p.id === node.paneId)?.shellName }}
        </span>
        <span class="term-shell-path font-mono">
          ({{ session.panes.find((p) => p.id === node.paneId)?.shellPath }})
        </span>
        <span v-if="(session.panes.length || 0) > 1" class="term-split-pill">
          {{
            session.splitDirection === "vertical"
              ? t("drawer.split_vertical") || "Split Vertical"
              : t("drawer.split_horizontal") || "Split Horizontal"
          }}
        </span>
      </div>
      <div class="pane-banner-right">
        <!-- Split Horizontal (top/bottom) -->
        <button
          class="pane-banner-btn"
          :title="t('terminal.split_horizontal') || 'Split Terminal Horizontally'"
          @click.stop="emit('splitPane', 'horizontal', node.paneId)"
        >
          <svg
            viewBox="0 0 16 16"
            width="11"
            height="11"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <rect x="2.5" y="2" width="11" height="12" rx="1.5" />
            <path d="M2.5 8h11" />
          </svg>
        </button>

        <!-- Split Vertical (left/right) -->
        <button
          class="pane-banner-btn"
          :title="t('terminal.split_vertical') || 'Split Terminal Vertically'"
          @click.stop="emit('splitPane', 'vertical', node.paneId)"
        >
          <svg
            viewBox="0 0 16 16"
            width="11"
            height="11"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <rect x="2" y="2.5" width="12" height="11" rx="1.5" />
            <path d="M8 2.5v11" />
          </svg>
        </button>

        <!-- Kill Pane Button -->
        <button
          v-if="(session.panes.length || 0) > 1"
          class="pane-close-btn"
          :title="t('terminal.kill_terminal') || 'Kill Terminal'"
          @click.stop="emit('closePane', node.paneId)"
        >
          <svg viewBox="0 0 16 16" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3.5 3.5l9 9M12.5 3.5l-9 9" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Pane Screen Output and Interactive xterm.js PTY Canvas -->
    <div class="terminal-screen">
      <XTerminalPane
        :pane-id="node.paneId"
        :shell-name="session.panes.find((p) => p.id === node.paneId)?.shellName || 'bash'"
        :shell-path="session.panes.find((p) => p.id === node.paneId)?.shellPath || '/bin/bash'"
        :is-active="node.paneId === activePaneId"
        @focus="emit('setActivePane', node.paneId)"
      />
    </div>
  </div>
</template>
