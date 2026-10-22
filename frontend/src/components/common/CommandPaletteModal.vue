<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->

<template>
  <transition name="palette-fade">
    <div
      v-if="isOpen"
      class="palette-overlay"
      id="modal-command-palette"
      @click="handleOverlayClick"
      @keydown.esc="close"
    >
      <div class="palette-container glass-card" @click.stop @contextmenu.prevent>
        <!-- Search Input Bar -->
        <div class="palette-search-row">
          <span class="palette-search-icon">
            <TechIcon id="workbench:search" :size="18" />
          </span>
          <input
            ref="inputRef"
            v-model="query"
            type="text"
            class="palette-input"
            :placeholder="t('command_palette.placeholder') || '输入命令或功能名称快速跳转或执行 (Ctrl+K)...'"
            @keydown.down.prevent="moveSelection(1)"
            @keydown.up.prevent="moveSelection(-1)"
            @keydown.enter.prevent="executeSelected"
          />
          <span v-if="query" class="palette-clear-btn" @click="query = ''">×</span>
          <kbd class="palette-esc-hint">ESC</kbd>
        </div>

        <!-- Command List -->
        <div class="palette-list-wrapper">
          <div v-if="filteredCommands.length === 0" class="palette-empty-state">
            <span class="empty-icon"><TechIcon id="workbench:search" :size="32" /></span>
            <span class="empty-text">{{ t("command_palette.no_results") || "未找到匹配的命令或功能" }}</span>
          </div>

          <div v-else class="palette-items-list">
            <div
              v-for="(item, index) in filteredCommands"
              :key="item.id"
              class="palette-item"
              :class="{ selected: selectedIndex === index }"
              @mouseenter="selectedIndex = index"
              @click="executeItem(item)"
            >
              <div class="item-left">
                <span class="item-icon">
                  <TechIcon v-if="item.techIcon" :id="item.techIcon" :size="18" />
                  <span v-else-if="item.codicon" class="codicon" :class="item.codicon"></span>
                  <span v-else class="codicon codicon-terminal"></span>
                </span>
                <div class="item-details">
                  <span class="item-title">{{ item.title }}</span>
                  <span v-if="item.desc" class="item-desc">{{ item.desc }}</span>
                </div>
              </div>

              <div class="item-right">
                <span v-if="item.category" class="category-pill">{{ item.category }}</span>
                <kbd v-if="item.shortcut" class="shortcut-pill">{{ item.shortcut }}</kbd>
                <span class="enter-arrow">↵</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer Bar -->
        <div class="palette-footer">
          <div class="footer-tips">
            <span class="tip-item"><kbd>↑</kbd> <kbd>↓</kbd> 导航</span>
            <span class="tip-item"><kbd>↵</kbd> 执行</span>
            <span class="tip-item"><kbd>ESC</kbd> 关闭</span>
          </div>
          <div class="footer-count">{{ filteredCommands.length }} 个可用命令</div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { useRouter } from "vue-router";
import { t } from "../../i18n";
import TechIcon from "./TechIcon.vue";
import type { DrawerTab } from "../layout/BottomDrawer.vue";

export interface CommandItem {
  id: string;
  title: string;
  desc?: string;
  category: "视图" | "Git" | "诊断" | "系统" | "工具";
  shortcut?: string;
  techIcon?: string;
  codicon?: string;
  action: () => void;
}

const props = defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "select-tab", tabId: string): void;
  (e: "toggle-theme"): void;
  (e: "open-drawer", tab?: DrawerTab): void;
  (e: "toggle-scope"): void;
}>();

const router = useRouter();
const query = ref("");
const selectedIndex = ref(0);
const inputRef = ref<HTMLInputElement | null>(null);

function close() {
  emit("close");
}

function handleOverlayClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    close();
  }
}

watch(
  () => props.isOpen,
  (val) => {
    if (val) {
      query.value = "";
      selectedIndex.value = 0;
      nextTick(() => {
        inputRef.value?.focus();
      });
    }
  }
);

const allCommands = computed<CommandItem[]>(() => [
  // 1. Navigation views
  {
    id: "nav-scm",
    title: t("nav.scm") || "源码管理 (Source Control / Git)",
    desc: "查看变更列表、暂存更改、分支切换与提交历史图谱",
    category: "Git",
    shortcut: "Alt+G",
    techIcon: "workbench:git",
    action: () => {
      emit("select-tab", "scm");
      router.push("/scm");
    },
  },
  {
    id: "nav-doctor",
    title: t("nav.doctor") || "系统健康体检 (Doctor)",
    desc: "全景探测核心运行时、Shims 路由、磁盘足迹与 API 配额",
    category: "诊断",
    shortcut: "Alt+D",
    techIcon: "workbench:doctor",
    action: () => {
      emit("select-tab", "doctor");
      router.push("/doctor");
    },
  },
  {
    id: "nav-tools",
    title: t("nav.tools") || "开发环境与工具链 (Tools)",
    desc: "管理 Node.js, Python, Go, Rust 等多版本运行时与编译器",
    category: "工具",
    techIcon: "workbench:tools",
    action: () => {
      emit("select-tab", "tools");
      router.push("/tools");
    },
  },
  {
    id: "nav-environments",
    title: t("nav.environments") || "环境变量管理 (Environments)",
    desc: "配置与管理当前工作区和用户全局环境变量",
    category: "视图",
    techIcon: "workbench:environments",
    action: () => {
      emit("select-tab", "environments");
      router.push("/environments");
    },
  },
  {
    id: "nav-tasks",
    title: t("nav.tasks") || "工程任务自动化 (Tasks)",
    desc: "运行构建、测试、格式化及自定义流水线脚本",
    category: "视图",
    techIcon: "workbench:tasks",
    action: () => {
      emit("select-tab", "tasks");
      router.push("/tasks");
    },
  },
  {
    id: "nav-marketplace",
    title: t("nav.marketplace") || "扩展市场 (Marketplace)",
    desc: "发现并安装主题外观包、插件及后端扩展",
    category: "视图",
    techIcon: "workbench:marketplace",
    action: () => {
      emit("select-tab", "marketplace");
      router.push("/marketplace");
    },
  },
  {
    id: "nav-explorer",
    title: t("nav.explorer") || "文件资源管理器 (Explorer)",
    desc: "浏览工作区项目目录树与编辑代码文件",
    category: "视图",
    techIcon: "workbench:folder",
    action: () => {
      emit("select-tab", "explorer");
      router.push("/explorer");
    },
  },
  {
    id: "nav-dashboard",
    title: t("nav.dashboard") || "项目全景看板 (Dashboard)",
    desc: "UniRTM 项目状态概览、核心版本指标与快捷向导",
    category: "视图",
    techIcon: "workbench:dashboard",
    action: () => {
      emit("select-tab", "dashboard");
      router.push("/dashboard");
    },
  },
  {
    id: "nav-guard",
    title: t("nav.guard") || "代码质量守护 (Quality Guard)",
    desc: "自动化预提交拦截、安全密钥防泄露与格式合规性检查",
    category: "诊断",
    codicon: "codicon-shield",
    action: () => {
      emit("select-tab", "guard");
      router.push("/guard");
    },
  },
  {
    id: "nav-settings",
    title: t("settings.title") || "系统偏好设置 (Settings)",
    desc: "自定义编辑器、外观主题、网络代理与安全设置",
    category: "系统",
    shortcut: "Ctrl+,",
    techIcon: "workbench:settings",
    action: () => {
      emit("select-tab", "settings");
      router.push("/settings");
    },
  },
  {
    id: "nav-about",
    title: t("about.title") || "关于 UniRTM (About)",
    desc: "查看版本信息、构建特征集、系统许可证及更新日志",
    category: "系统",
    techIcon: "workbench:about",
    action: () => {
      emit("select-tab", "about");
      router.push("/about");
    },
  },

  // 2. Global Actions
  {
    id: "act-theme",
    title: "切换浅色 / 深色外观主题 (Toggle Theme)",
    desc: "在现代化 VS Code 暗黑高对比度与浅色极简模式间切换",
    category: "系统",
    codicon: "codicon-color-mode",
    action: () => {
      emit("toggle-theme");
    },
  },
  {
    id: "act-scope",
    title: "切换当前配置生效作用域 (Toggle Scope: Workspace / User)",
    desc: "切换工作区专属配置 (.unirtm.toml) 或用户全局配置 (~/.unirtm)",
    category: "系统",
    shortcut: "Alt+S",
    codicon: "codicon-layers",
    action: () => {
      emit("toggle-scope");
    },
  },
  {
    id: "act-terminal",
    title: "打开集成终端 (Open Integrated Terminal)",
    desc: "在底部抽屉中启动内置 Shell 交互控制台",
    category: "工具",
    shortcut: "Ctrl+`",
    codicon: "codicon-terminal",
    action: () => {
      emit("open-drawer", "terminal");
    },
  },
  {
    id: "act-output",
    title: "查看运行输出与日志 (Show Logs / Output)",
    desc: "检查运行时后台输出、事件分发及构建任务流水日志",
    category: "工具",
    codicon: "codicon-output",
    action: () => {
      emit("open-drawer", "output");
    },
  },
]);

const filteredCommands = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (!q) return allCommands.value;
  return allCommands.value.filter((cmd) => {
    return (
      cmd.title.toLowerCase().includes(q) ||
      (cmd.desc && cmd.desc.toLowerCase().includes(q)) ||
      cmd.category.toLowerCase().includes(q) ||
      cmd.id.toLowerCase().includes(q)
    );
  });
});

watch(filteredCommands, (newFiltered) => {
  if (selectedIndex.value >= newFiltered.length) {
    selectedIndex.value = Math.max(0, newFiltered.length - 1);
  }
});

function moveSelection(step: number) {
  const total = filteredCommands.value.length;
  if (total === 0) return;
  selectedIndex.value = (selectedIndex.value + step + total) % total;
}

function executeSelected() {
  const item = filteredCommands.value[selectedIndex.value];
  if (item) {
    executeItem(item);
  }
}

function executeItem(item: CommandItem) {
  close();
  item.action();
}
</script>

<style scoped>
.palette-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  z-index: 99999;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 5vh;
}

.palette-container {
  width: 640px;
  max-width: 90vw;
  background: var(--bg-card, #1e1e1e);
  border: 1px solid var(--border-color, #333333);
  border-radius: 10px;
  box-shadow:
    0 16px 48px rgba(0, 0, 0, 0.55),
    0 0 0 1px rgba(255, 255, 255, 0.08);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.palette-search-row {
  display: flex;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-color, #2d2d2d);
  gap: 0.75rem;
  background: rgba(255, 255, 255, 0.02);
}

.palette-search-icon {
  display: flex;
  align-items: center;
  color: var(--color-primary, #00f0ff);
}

.palette-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main, #ffffff);
  font-size: 0.95rem;
  font-weight: 500;
  font-family: inherit;
}

.palette-input::placeholder {
  color: var(--text-muted, #888888);
}

.palette-clear-btn {
  font-size: 1.1rem;
  color: var(--text-muted, #888888);
  cursor: pointer;
  padding: 0 4px;
}

.palette-clear-btn:hover {
  color: var(--text-main, #ffffff);
}

.palette-esc-hint {
  font-size: 0.68rem;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: var(--text-muted, #aaaaaa);
  font-family: var(--font-mono, monospace);
}

.palette-list-wrapper {
  max-height: 380px;
  overflow-y: auto;
  padding: 0.4rem;
}

.palette-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2.5rem 1rem;
  gap: 0.6rem;
  color: var(--text-muted, #888888);
}

.palette-items-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.palette-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.55rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.12s ease;
  user-select: none;
}

.palette-item:hover,
.palette-item.selected {
  background: rgba(0, 240, 255, 0.1);
  color: var(--color-primary, #00f0ff);
}

.item-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  overflow: hidden;
}

.item-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  color: inherit;
  flex-shrink: 0;
}

.item-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow: hidden;
}

.item-title {
  font-size: 0.86rem;
  font-weight: 600;
  color: var(--text-main, #ffffff);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.palette-item.selected .item-title {
  color: var(--color-primary, #00f0ff);
}

.item-desc {
  font-size: 0.72rem;
  color: var(--text-muted, #999999);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.category-pill {
  font-size: 0.68rem;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted, #888888);
}

.shortcut-pill {
  font-size: 0.68rem;
  padding: 1px 5px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: var(--text-muted, #aaaaaa);
  font-family: var(--font-mono, monospace);
}

.enter-arrow {
  opacity: 0;
  font-size: 0.85rem;
  color: var(--color-primary, #00f0ff);
  transition: opacity 0.15s ease;
}

.palette-item.selected .enter-arrow {
  opacity: 1;
}

.palette-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.45rem 1rem;
  background: rgba(0, 0, 0, 0.2);
  border-top: 1px solid var(--border-color, #2d2d2d);
  font-size: 0.72rem;
  color: var(--text-muted, #777777);
}

.footer-tips {
  display: flex;
  gap: 0.75rem;
}

.tip-item kbd {
  font-family: var(--font-mono, monospace);
  background: rgba(255, 255, 255, 0.08);
  padding: 1px 4px;
  border-radius: 3px;
}

.palette-fade-enter-active,
.palette-fade-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.palette-fade-enter-from,
.palette-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
