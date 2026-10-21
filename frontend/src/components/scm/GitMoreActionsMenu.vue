<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="git-more-menu-overlay"
      @click="handleOverlayClick"
      @contextmenu.prevent="handleOverlayClick"
    >
      <div class="git-more-menu" :style="menuStyle" @click.stop>
        <!-- 1. View Git Graph -->
        <div class="menu-item" @mouseenter="clearSubmenu" @click="triggerAction('view-graph')">
          <span class="codicon codicon-git-commit item-icon"></span>
          <span class="item-label">{{ t("scm.menu.view_graph") }}</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 2. Primary Top-level Actions -->
        <div class="menu-item" @mouseenter="clearSubmenu" @click="triggerAction('pull')">
          <span class="codicon codicon-arrow-down item-icon"></span>
          <span class="item-label">{{ t("scm.menu.pull") }}</span>
        </div>
        <div class="menu-item" @mouseenter="clearSubmenu" @click="triggerAction('push')">
          <span class="codicon codicon-arrow-up item-icon"></span>
          <span class="item-label">{{ t("scm.menu.push") }}</span>
        </div>
        <div class="menu-item" @mouseenter="clearSubmenu" @click="triggerAction('clone')">
          <span class="codicon codicon-repo-clone item-icon"></span>
          <span class="item-label">{{ t("scm.menu.clone") }}</span>
        </div>
        <div class="menu-item" @mouseenter="clearSubmenu" @click="triggerAction('checkout')">
          <span class="codicon codicon-git-branch item-icon"></span>
          <span class="item-label">{{ t("scm.menu.checkout") }}</span>
        </div>
        <div class="menu-item" @mouseenter="clearSubmenu" @click="triggerAction('fetch')">
          <span class="codicon codicon-sync item-icon"></span>
          <span class="item-label">{{ t("scm.menu.fetch") }}</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 3. Submenu: Commit ▶ -->
        <div
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'commit' }"
          @mouseenter="openSubmenu('commit', $event)"
        >
          <span class="codicon codicon-check item-icon"></span>
          <span class="item-label">{{ t("scm.menu.commit") }}</span>
          <span class="codicon codicon-chevron-right submenu-arrow"></span>
        </div>

        <!-- 4. Submenu: Changes ▶ -->
        <div
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'changes' }"
          @mouseenter="openSubmenu('changes', $event)"
        >
          <span class="codicon codicon-diff-modified item-icon"></span>
          <span class="item-label">{{ t("scm.menu.changes") }}</span>
          <span class="codicon codicon-chevron-right submenu-arrow"></span>
        </div>

        <!-- 5. Submenu: Pull, Push ▶ -->
        <div
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'pull-push' }"
          @mouseenter="openSubmenu('pull-push', $event)"
        >
          <span class="codicon codicon-sync item-icon"></span>
          <span class="item-label">{{ t("scm.menu.pull_push") }}</span>
          <span class="codicon codicon-chevron-right submenu-arrow"></span>
        </div>

        <!-- 6. Submenu: Branch ▶ -->
        <div
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'branch' }"
          @mouseenter="openSubmenu('branch', $event)"
        >
          <span class="codicon codicon-git-branch item-icon"></span>
          <span class="item-label">{{ t("scm.menu.branch") }}</span>
          <span class="codicon codicon-chevron-right submenu-arrow"></span>
        </div>

        <!-- 7. Submenu: Remote ▶ -->
        <div
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'remote' }"
          @mouseenter="openSubmenu('remote', $event)"
        >
          <span class="codicon codicon-cloud item-icon"></span>
          <span class="item-label">{{ t("scm.menu.remote") }}</span>
          <span class="codicon codicon-chevron-right submenu-arrow"></span>
        </div>

        <!-- 8. Submenu: Stash ▶ -->
        <div
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'stash' }"
          @mouseenter="openSubmenu('stash', $event)"
        >
          <span class="codicon codicon-archive item-icon"></span>
          <span class="item-label">{{ t("scm.menu.stash") }}</span>
          <span class="codicon codicon-chevron-right submenu-arrow"></span>
        </div>

        <!-- 9. Submenu: Tags ▶ -->
        <div
          class="menu-item has-submenu"
          :class="{ active: activeSubmenu === 'tags' }"
          @mouseenter="openSubmenu('tags', $event)"
        >
          <span class="codicon codicon-tag item-icon"></span>
          <span class="item-label">{{ t("scm.menu.tags") }}</span>
          <span class="codicon codicon-chevron-right submenu-arrow"></span>
        </div>

        <div class="menu-separator"></div>

        <!-- 10. Show Git Output -->
        <div class="menu-item" @mouseenter="clearSubmenu" @click="triggerAction('show-output')">
          <span class="codicon codicon-output item-icon"></span>
          <span class="item-label">{{ t("scm.menu.show_output") }}</span>
        </div>
      </div>

      <!-- Nested Submenu Layer -->
      <div v-if="activeSubmenu" class="git-more-submenu" :style="submenuStyle" @click.stop>
        <!-- Commit Submenu -->
        <template v-if="activeSubmenu === 'commit'">
          <div class="menu-item" @click="triggerAction('commit')">
            {{ t("scm.menu.commit") }}
          </div>
          <div class="menu-item" @click="triggerAction('commit-amend')">
            {{ t("scm.menu.commit_amend") }}
          </div>
          <div class="menu-item" @click="triggerAction('commit-staged')">
            {{ t("scm.menu.commit_staged") }}
          </div>
          <div class="menu-item" @click="triggerAction('commit-all')">
            {{ t("scm.menu.commit_all") }}
          </div>
        </template>

        <!-- Changes Submenu -->
        <template v-else-if="activeSubmenu === 'changes'">
          <div class="menu-item" @click="triggerAction('stage-all')">
            {{ t("scm.menu.stage_all") }}
          </div>
          <div class="menu-item" @click="triggerAction('unstage-all')">
            {{ t("scm.menu.unstage_all") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item danger" @click="triggerAction('discard-all')">
            {{ t("scm.menu.discard_all") }}
          </div>
        </template>

        <!-- Pull, Push Submenu -->
        <template v-else-if="activeSubmenu === 'pull-push'">
          <div class="menu-item" @click="triggerAction('sync')">
            {{ t("scm.menu.sync") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="triggerAction('pull')">
            {{ t("scm.menu.pull") }}
          </div>
          <div class="menu-item" @click="triggerAction('pull-rebase')">
            {{ t("scm.menu.pull_rebase") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="triggerAction('push')">
            {{ t("scm.menu.push") }}
          </div>
          <div class="menu-item danger" @click="triggerAction('push-force')">
            {{ t("scm.menu.push_force") }}
          </div>
        </template>

        <!-- Branch Submenu -->
        <template v-else-if="activeSubmenu === 'branch'">
          <div class="menu-item" @click="triggerAction('create-branch')">
            {{ t("scm.menu.create_branch") }}
          </div>
          <div class="menu-item" @click="triggerAction('checkout')">
            {{ t("scm.menu.checkout") }}
          </div>
          <div class="menu-item" @click="triggerAction('rename-branch')">
            {{ t("scm.menu.rename_branch") }}
          </div>
          <div class="menu-item danger" @click="triggerAction('delete-branch')">
            {{ t("scm.menu.delete_branch") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="triggerAction('merge-branch')">
            {{ t("scm.menu.merge_branch") }}
          </div>
          <div class="menu-item" @click="triggerAction('rebase-branch')">
            {{ t("scm.menu.rebase_branch") }}
          </div>
        </template>

        <!-- Remote Submenu -->
        <template v-else-if="activeSubmenu === 'remote'">
          <div class="menu-item" @click="triggerAction('fetch')">
            {{ t("scm.menu.fetch_all") }}
          </div>
          <div class="menu-item" @click="triggerAction('push')">
            {{ t("scm.menu.push_remote") }}
          </div>
        </template>

        <!-- Stash Submenu -->
        <template v-else-if="activeSubmenu === 'stash'">
          <div class="menu-item" @click="triggerAction('stash')">
            {{ t("scm.menu.stash_changes") }}
          </div>
          <div class="menu-item" @click="triggerAction('stash-untracked')">
            {{ t("scm.menu.stash_include_untracked") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item danger" @click="triggerAction('reset-uncommitted')">
            {{ t("scm.menu.reset_uncommitted") }}
          </div>
          <div class="menu-item danger" @click="triggerAction('clean-untracked')">
            {{ t("scm.menu.clean_untracked") }}
          </div>
        </template>

        <!-- Tags Submenu -->
        <template v-else-if="activeSubmenu === 'tags'">
          <div class="menu-item" @click="triggerAction('create-tag')">
            {{ t("scm.menu.create_tag") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="triggerAction('push-tags')">
            {{ t("scm.menu.push_tags") }}
          </div>
          <div class="menu-item danger" @click="triggerAction('push-tags-force')">
            {{ t("scm.menu.push_tags_force") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item danger" @click="triggerAction('delete-tag')">
            {{ t("scm.menu.delete_tag") }}
          </div>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { t } from "../../i18n";

const props = defineProps<{
  visible: boolean;
  position: { x: number; y: number };
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "action", action: string): void;
}>();

type SubmenuType = "commit" | "changes" | "pull-push" | "branch" | "remote" | "stash" | "tags" | null;

const activeSubmenu = ref<SubmenuType>(null);
const submenuPos = ref({ x: 0, y: 0 });

const MENU_WIDTH = 240;
const SUBMENU_WIDTH = 220;

const menuStyle = computed(() => {
  let left = props.position.x;
  let top = props.position.y;

  // Prevent right edge overflow
  if (left + MENU_WIDTH > window.innerWidth - 10) {
    left = window.innerWidth - MENU_WIDTH - 10;
  }
  // Prevent bottom edge overflow
  const estimatedHeight = 440;
  if (top + estimatedHeight > window.innerHeight - 10) {
    top = Math.max(10, window.innerHeight - estimatedHeight - 10);
  }

  return {
    left: `${Math.max(10, left)}px`,
    top: `${Math.max(10, top)}px`,
  };
});

const submenuStyle = computed(() => {
  return {
    left: `${submenuPos.value.x}px`,
    top: `${submenuPos.value.y}px`,
  };
});

function openSubmenu(type: SubmenuType, event: MouseEvent) {
  activeSubmenu.value = type;
  const target = event.currentTarget as HTMLElement | null;
  if (!target) return;

  const rect = target.getBoundingClientRect();
  let x = rect.right + 2;
  let y = rect.top - 4;

  // If opening on the right overflows window, open to the left
  if (x + SUBMENU_WIDTH > window.innerWidth - 10) {
    x = rect.left - SUBMENU_WIDTH - 2;
  }

  // Prevent bottom overflow
  const estimatedSubHeight = 220;
  if (y + estimatedSubHeight > window.innerHeight - 10) {
    y = Math.max(10, window.innerHeight - estimatedSubHeight - 10);
  }

  submenuPos.value = { x, y };
}

function clearSubmenu() {
  activeSubmenu.value = null;
}

function triggerAction(action: string) {
  emit("action", action);
  emit("close");
}

function handleOverlayClick() {
  emit("close");
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && props.visible) {
    emit("close");
  }
}

onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
  window.addEventListener("resize", handleOverlayClick);
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", handleKeydown);
  window.removeEventListener("resize", handleOverlayClick);
});
</script>

<style scoped>
.git-more-menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: transparent;
}

.git-more-menu,
.git-more-submenu {
  position: fixed;
  background: var(--card-bg, #0d1626);
  border: 1px solid var(--card-border, #1f2e48);
  border-radius: 6px;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.45);
  padding: 5px 0;
  font-size: 12px;
  color: var(--text-main, #e2e8f0);
  user-select: none;
  z-index: 10000;
  min-width: 200px;
}

:root[data-theme="light"] .git-more-menu,
:root[data-theme="light"] .git-more-submenu {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.12);
  color: #1e293b;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  cursor: pointer;
  position: relative;
  transition: background 0.12s ease;
}

.menu-item:hover,
.menu-item.active {
  background: var(--hover-bg, rgba(56, 189, 248, 0.12));
  color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .menu-item:hover,
:root[data-theme="light"] .menu-item.active {
  background: #f1f5f9;
  color: #0284c7;
}

.item-icon {
  font-size: 14px;
  width: 16px;
  text-align: center;
  color: var(--text-muted, #94a3b8);
}

.menu-item:hover .item-icon,
.menu-item.active .item-icon {
  color: inherit;
}

.item-label {
  flex: 1;
  white-space: nowrap;
}

.submenu-arrow {
  font-size: 12px;
  color: var(--text-muted, #64748b);
  margin-left: 12px;
}

.menu-separator {
  height: 1px;
  background: var(--card-border, #1e293b);
  margin: 4px 0;
}

:root[data-theme="light"] .menu-separator {
  background: #e2e8f0;
}

.menu-item.danger {
  color: #f87171;
}

.menu-item.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}
</style>
