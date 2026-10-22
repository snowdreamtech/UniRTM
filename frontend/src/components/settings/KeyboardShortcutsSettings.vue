<template>
  <div class="shortcuts-settings-wrapper">
    <!-- Header -->
    <div class="section-header">
      <div class="section-title-wrap">
        <span class="section-icon">
          <span class="codicon codicon-keyboard"></span>
        </span>
        <div>
          <h2>{{ isZh ? "键盘快捷键 (Keyboard Shortcuts)" : "Keyboard Shortcuts" }}</h2>
          <span class="section-sub font-mono">
            {{ isZh ? "快捷键自定义与覆盖映射 • 持久化于 " : "Customize and override shortcuts • Persisted at " }}
            <span class="file-path-badge">{{ keybindingsPath }}</span>
          </span>
        </div>
      </div>

      <div class="header-actions">
        <button
          class="btn btn-secondary"
          :title="isZh ? '打开 keybindings.json 原始文件' : 'Open keybindings.json'"
          @click="openRawJsonModal"
        >
          <span class="codicon codicon-json"></span>
          <span>{{ isZh ? "打开 keybindings.json" : "Open JSON" }}</span>
        </button>
        <button
          v-if="hasCustomKeybindings"
          class="btn btn-outline-danger"
          :title="isZh ? '重置所有自定义快捷键为系统默认' : 'Reset all shortcuts to defaults'"
          @click="handleResetAll"
        >
          <span class="codicon codicon-discard"></span>
          <span>{{ isZh ? "恢复全部默认" : "Reset All" }}</span>
        </button>
      </div>
    </div>

    <!-- Filter & Search Toolbar -->
    <div class="shortcuts-toolbar">
      <div class="search-input-wrap">
        <span class="codicon codicon-search search-icon"></span>
        <input
          v-model="searchQuery"
          type="text"
          class="filter-input"
          :placeholder="isZh ? '按命令名称、命令 ID 或快捷键过滤...' : 'Filter by command name, ID or key...'"
        />
        <button v-if="searchQuery" class="clear-btn" @click="searchQuery = ''">
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <div class="source-filter-pills">
        <button class="filter-pill" :class="{ active: sourceFilter === 'all' }" @click="sourceFilter = 'all'">
          {{ isZh ? "全部 (" + activeShortcuts.length + ")" : "All (" + activeShortcuts.length + ")" }}
        </button>
        <button class="filter-pill" :class="{ active: sourceFilter === 'user' }" @click="sourceFilter = 'user'">
          {{ isZh ? "用户自定义 (" + customCount + ")" : "User (" + customCount + ")" }}
        </button>
        <button class="filter-pill" :class="{ active: sourceFilter === 'default' }" @click="sourceFilter = 'default'">
          {{
            isZh
              ? "系统内置 (" + (activeShortcuts.length - customCount) + ")"
              : "Default (" + (activeShortcuts.length - customCount) + ")"
          }}
        </button>
      </div>
    </div>

    <!-- Shortcuts Table -->
    <div class="shortcuts-table-wrapper">
      <table class="shortcuts-table">
        <thead>
          <tr>
            <th class="col-cmd">{{ isZh ? "命令与功能" : "Command" }}</th>
            <th class="col-id">{{ isZh ? "命令标识符 (Command ID)" : "Command ID" }}</th>
            <th class="col-key">{{ isZh ? "快捷键 (Keybinding)" : "Keybinding" }}</th>
            <th class="col-source">{{ isZh ? "来源" : "Source" }}</th>
            <th class="col-actions">{{ isZh ? "操作" : "Actions" }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in filteredList"
            :key="item.id"
            class="shortcut-row"
            :class="{ 'is-custom': item.isCustom }"
            @dblclick="openRecordModal(item)"
          >
            <!-- Command -->
            <td class="col-cmd">
              <div class="cmd-content">
                <span class="cmd-title">{{ isZh ? item.titleZh : item.titleEn }}</span>
                <span class="cmd-desc">{{ isZh ? item.descZh : item.descEn }}</span>
              </div>
            </td>

            <!-- Command ID -->
            <td class="col-id">
              <span class="cmd-id-pill font-mono">{{ item.id }}</span>
            </td>

            <!-- Keybinding -->
            <td class="col-key">
              <div class="keys-display">
                <template v-for="(chord, cIdx) in splitKeys(item.effectiveKey)" :key="cIdx">
                  <span v-if="cIdx > 0" class="chord-sep">+</span>
                  <kbd class="keycap" :class="{ 'custom-keycap': item.isCustom }">{{ chord }}</kbd>
                </template>
              </div>
            </td>

            <!-- Source -->
            <td class="col-source">
              <span v-if="item.isCustom" class="badge-source badge-user font-mono">
                {{ isZh ? "User (用户)" : "User" }}
              </span>
              <span v-else class="badge-source badge-default font-mono">
                {{ isZh ? "System (内置)" : "System" }}
              </span>
            </td>

            <!-- Actions -->
            <td class="col-actions">
              <div class="action-buttons">
                <button
                  class="action-btn edit-btn"
                  :title="isZh ? '修改此快捷键' : 'Change Keybinding'"
                  @click="openRecordModal(item)"
                >
                  <span class="codicon codicon-edit"></span>
                  <span class="btn-text">{{ isZh ? "编辑" : "Edit" }}</span>
                </button>

                <button
                  v-if="item.isCustom"
                  class="action-btn reset-btn"
                  :title="isZh ? '恢复为此命令的默认快捷键' : 'Reset to default'"
                  @click="handleResetOne(item.id)"
                >
                  <span class="codicon codicon-discard"></span>
                  <span class="btn-text">{{ isZh ? "重置" : "Reset" }}</span>
                </button>
              </div>
            </td>
          </tr>

          <tr v-if="filteredList.length === 0" class="empty-state-row">
            <td colspan="5">
              <div class="empty-state">
                <span class="codicon codicon-search empty-icon"></span>
                <p>{{ isZh ? "未找到符合条件的快捷键" : "No matching keybindings found" }}</p>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 1. Key Recording Modal (VS Code Aligned) -->
    <Teleport to="body">
      <div
        v-if="isRecordingModalOpen"
        class="recording-modal-overlay"
        @click.self="closeRecordModal"
        @keydown="handleRecordKeydown"
        tabindex="0"
      >
        <div class="recording-modal-card" role="dialog" aria-modal="true">
          <div class="modal-card-header">
            <span class="codicon codicon-keyboard modal-icon"></span>
            <h3>{{ isZh ? "修改快捷键" : "Change Keybinding" }}</h3>
            <button class="modal-close-btn" @click="closeRecordModal">
              <span class="codicon codicon-close"></span>
            </button>
          </div>

          <div class="modal-card-body">
            <div class="target-cmd-info">
              <span class="target-title">{{ isZh ? recordingTarget?.titleZh : recordingTarget?.titleEn }}</span>
              <span class="target-id font-mono">{{ recordingTarget?.id }}</span>
            </div>

            <div class="recording-box" :class="{ listening: true }">
              <div v-if="recordedCombo" class="recorded-keys">
                <kbd class="keycap active-recording">{{ recordedCombo }}</kbd>
              </div>
              <div v-else class="recording-prompt">
                <span class="codicon codicon-record pulse-dot"></span>
                <span>{{
                  isZh
                    ? "请按下所需的按键组合，然后按 Enter 保存..."
                    : "Press desired key combination, then press Enter..."
                }}</span>
              </div>
            </div>

            <!-- Conflict warning -->
            <div v-if="conflictCommand" class="conflict-warning">
              <span class="codicon codicon-warning"></span>
              <span>
                {{ isZh ? "提示: 此快捷键已被用于 " : "Note: Key already bound to " }}
                <strong>{{ isZh ? conflictCommand.titleZh : conflictCommand.titleEn }}</strong>
                ({{ conflictCommand.id }})
              </span>
            </div>

            <div class="recording-tips">
              <span>{{
                isZh
                  ? "支持 Ctrl, Alt, Shift, Cmd 组合按键 • Esc 取消录制"
                  : "Supports Ctrl, Alt, Shift, Cmd combinations • Esc to cancel"
              }}</span>
            </div>
          </div>

          <div class="modal-card-footer">
            <button class="btn btn-secondary" @click="closeRecordModal">
              {{ isZh ? "取消" : "Cancel" }}
            </button>
            <button class="btn btn-primary" :disabled="!recordedCombo" @click="applyRecordedKey">
              {{ isZh ? "保存生效 (Enter)" : "Save (Enter)" }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 2. Raw keybindings.json Editor Modal -->
    <Teleport to="body">
      <div v-if="isJsonModalOpen" class="recording-modal-overlay" @click.self="isJsonModalOpen = false">
        <div class="raw-json-modal-card" role="dialog" aria-modal="true">
          <div class="modal-card-header">
            <div class="header-left">
              <span class="codicon codicon-json modal-icon"></span>
              <div>
                <h3>keybindings.json</h3>
                <span class="file-path font-mono">{{ keybindingsPath }}</span>
              </div>
            </div>
            <button class="modal-close-btn" @click="isJsonModalOpen = false">
              <span class="codicon codicon-close"></span>
            </button>
          </div>

          <div class="raw-json-body">
            <textarea
              v-model="editingJsonText"
              class="json-textarea font-mono"
              spellcheck="false"
              :placeholder="jsonPlaceholder"
            ></textarea>
            <div v-if="jsonError" class="json-error-banner">
              <span class="codicon codicon-error"></span>
              <span>{{ jsonError }}</span>
            </div>
          </div>

          <div class="modal-card-footer">
            <button class="btn btn-secondary" @click="handleFormatJson">
              {{ isZh ? "格式化 JSON" : "Format JSON" }}
            </button>
            <div style="flex: 1"></div>
            <button class="btn btn-secondary" @click="isJsonModalOpen = false">
              {{ isZh ? "关闭" : "Close" }}
            </button>
            <button class="btn btn-primary" @click="handleSaveRawJson">
              {{ isZh ? "保存并生效" : "Save Changes" }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import {
  useKeyboardShortcuts,
  normalizeKeyComboFromEvent,
  type ActiveShortcutItem,
} from "../../composables/useKeyboardShortcuts";
import { selectedLangSetting } from "../../i18n";
import { useToast } from "../../composables/useToast";

const {
  activeShortcuts,
  keybindingsPath,
  rawKeybindingsJson,
  fetchKeybindings,
  saveKeybinding,
  resetKeybinding,
  resetAllKeybindings,
  saveRawKeybindings,
} = useKeyboardShortcuts();

const { showToast } = useToast();

const isZh = computed(() => {
  const lang = selectedLangSetting.value || "zh-CN";
  return lang.startsWith("zh") || lang === "auto";
});

const searchQuery = ref("");
const sourceFilter = ref<"all" | "user" | "default">("all");

const isRecordingModalOpen = ref(false);
const recordingTarget = ref<ActiveShortcutItem | null>(null);
const recordedCombo = ref("");

const isJsonModalOpen = ref(false);
const editingJsonText = ref("");
const jsonError = ref("");
const jsonPlaceholder = computed(() => '[\n  {\n    "command": "workbench.view.scm",\n    "key": "ctrl+alt+g"\n  }\n]');

const customCount = computed(() => activeShortcuts.value.filter((s) => s.isCustom).length);
const hasCustomKeybindings = computed(() => customCount.value > 0);

function splitKeys(keys: string): string[] {
  return keys.split(" ");
}

const filteredList = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  const filter = sourceFilter.value;

  return activeShortcuts.value.filter((item) => {
    if (filter === "user" && !item.isCustom) return false;
    if (filter === "default" && item.isCustom) return false;

    if (!q) return true;

    return (
      item.titleZh.toLowerCase().includes(q) ||
      item.titleEn.toLowerCase().includes(q) ||
      item.id.toLowerCase().includes(q) ||
      item.effectiveKey.toLowerCase().includes(q)
    );
  });
});

const conflictCommand = computed(() => {
  if (!recordedCombo.value || !recordingTarget.value) return null;
  const combo = recordedCombo.value.toLowerCase();
  return (
    activeShortcuts.value.find((s) => s.id !== recordingTarget.value?.id && s.effectiveKey.toLowerCase() === combo) ||
    null
  );
});

function openRecordModal(item: ActiveShortcutItem) {
  recordingTarget.value = item;
  recordedCombo.value = "";
  isRecordingModalOpen.value = true;
}

function closeRecordModal() {
  isRecordingModalOpen.value = false;
  recordingTarget.value = null;
  recordedCombo.value = "";
}

function handleRecordKeydown(e: KeyboardEvent) {
  e.preventDefault();
  e.stopPropagation();

  if (e.key === "Escape") {
    closeRecordModal();
    return;
  }

  if (e.key === "Enter") {
    if (recordedCombo.value) {
      applyRecordedKey();
    }
    return;
  }

  const combo = normalizeKeyComboFromEvent(e);
  // Don't update if only modifier key was pressed
  if (!["Ctrl", "Alt", "Shift", "Cmd"].includes(combo)) {
    recordedCombo.value = combo;
  }
}

async function applyRecordedKey() {
  if (!recordingTarget.value || !recordedCombo.value) return;
  const cmdId = recordingTarget.value.id;
  const newKey = recordedCombo.value;

  const ok = await saveKeybinding(cmdId, newKey);
  if (ok) {
    showToast(isZh.value ? `已成功绑定 ${newKey}` : `Bound ${newKey} to ${cmdId}`, "success");
    closeRecordModal();
  } else {
    showToast(isZh.value ? "保存快捷键失败" : "Failed to save keybinding", "error");
  }
}

async function handleResetOne(cmdId: string) {
  const ok = await resetKeybinding(cmdId);
  if (ok) {
    showToast(isZh.value ? "已恢复出厂默认按键" : "Reset keybinding to default", "info");
  }
}

async function handleResetAll() {
  if (!confirm(isZh.value ? "确定要恢复所有快捷键为系统默认值吗？" : "Reset all keybindings to defaults?")) {
    return;
  }
  const ok = await resetAllKeybindings();
  if (ok) {
    showToast(isZh.value ? "所有快捷键已恢复为默认设置" : "All keybindings reset to defaults", "success");
  }
}

function openRawJsonModal() {
  editingJsonText.value = rawKeybindingsJson.value || "[]\n";
  jsonError.value = "";
  isJsonModalOpen.value = true;
}

function handleFormatJson() {
  try {
    const parsed = JSON.parse(editingJsonText.value);
    editingJsonText.value = JSON.stringify(parsed, null, 2) + "\n";
    jsonError.value = "";
  } catch (err: any) {
    jsonError.value = err.message;
  }
}

async function handleSaveRawJson() {
  try {
    const parsed = JSON.parse(editingJsonText.value);
    if (!Array.isArray(parsed)) {
      jsonError.value = "JSON 根节点必须为数组 (Root must be an array of keybinding objects)";
      return;
    }
    const ok = await saveRawKeybindings(editingJsonText.value);
    if (ok) {
      showToast(isZh.value ? "keybindings.json 保存成功" : "keybindings.json saved successfully", "success");
      isJsonModalOpen.value = false;
    }
  } catch (err: any) {
    jsonError.value = err.message;
  }
}

onMounted(() => {
  fetchKeybindings();
});
</script>

<style scoped>
.shortcuts-settings-wrapper {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.file-path-badge {
  color: #00e5ff;
  background: rgba(0, 229, 255, 0.08);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
}

:root[data-theme="light"] .file-path-badge {
  color: #0284c7;
  background: #e0f2fe;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Toolbar */
.shortcuts-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
}

:root[data-theme="light"] .shortcuts-toolbar {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}

.search-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 260px;
}

.search-icon {
  position: absolute;
  left: 10px;
  color: #94a3b8;
  font-size: 14px;
  pointer-events: none;
}

.filter-input {
  width: 100%;
  padding: 8px 32px 8px 32px;
  background: #111a2e;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #ffffff;
  font-size: 13px;
  outline: none;
  transition: all 0.2s ease;
}

:root[data-theme="light"] .filter-input {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

.filter-input:focus {
  border-color: #00e5ff;
  box-shadow: 0 0 0 2px rgba(0, 229, 255, 0.2);
}

:root[data-theme="light"] .filter-input:focus {
  border-color: #0284c7;
  box-shadow: 0 0 0 2px rgba(2, 132, 199, 0.2);
}

.clear-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
}

.source-filter-pills {
  display: flex;
  gap: 6px;
}

.filter-pill {
  padding: 6px 12px;
  font-size: 12px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.15s ease;
}

:root[data-theme="light"] .filter-pill {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  color: #64748b;
}

.filter-pill.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: #00e5ff;
  color: #00e5ff;
  font-weight: 500;
}

:root[data-theme="light"] .filter-pill.active {
  background: #e0f2fe;
  border-color: #0284c7;
  color: #0284c7;
}

/* Table */
.shortcuts-table-wrapper {
  overflow-x: auto;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  background: rgba(13, 20, 36, 0.6);
}

:root[data-theme="light"] .shortcuts-table-wrapper {
  border: 1px solid #e2e8f0;
  background: #ffffff;
}

.shortcuts-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.shortcuts-table th {
  padding: 10px 16px;
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0, 0, 0, 0.2);
}

:root[data-theme="light"] .shortcuts-table th {
  color: #64748b;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}

.shortcut-row {
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  transition: background-color 0.15s ease;
}

:root[data-theme="light"] .shortcut-row {
  border-bottom: 1px solid #f1f5f9;
}

.shortcut-row:hover {
  background: rgba(0, 229, 255, 0.04);
}

:root[data-theme="light"] .shortcut-row:hover {
  background: #f8fafc;
}

.shortcut-row td {
  padding: 12px 16px;
  vertical-align: middle;
}

.cmd-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.cmd-title {
  font-size: 13px;
  font-weight: 600;
  color: #ffffff;
}

:root[data-theme="light"] .cmd-title {
  color: #0f172a;
}

.cmd-desc {
  font-size: 12px;
  color: #94a3b8;
}

:root[data-theme="light"] .cmd-desc {
  color: #64748b;
}

.cmd-id-pill {
  font-size: 11px;
  color: #94a3b8;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.05);
}

:root[data-theme="light"] .cmd-id-pill {
  color: #475569;
  background: #f1f5f9;
}

.keys-display {
  display: flex;
  align-items: center;
  gap: 6px;
}

.chord-sep {
  color: #64748b;
  font-size: 11px;
}

.keycap {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.15);
  box-shadow: 0 2px 0 rgba(0, 0, 0, 0.3);
  color: #e2e8f0;
}

:root[data-theme="light"] .keycap {
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  box-shadow: 0 2px 0 rgba(15, 23, 42, 0.1);
  color: #0f172a;
}

.keycap.custom-keycap {
  color: #00e5ff;
  border-color: rgba(0, 229, 255, 0.4);
}

:root[data-theme="light"] .keycap.custom-keycap {
  color: #0284c7;
  border-color: #38bdf8;
}

.badge-source {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
}

.badge-user {
  background: rgba(0, 229, 255, 0.12);
  color: #00e5ff;
  border: 1px solid rgba(0, 229, 255, 0.25);
}

:root[data-theme="light"] .badge-user {
  background: #e0f2fe;
  color: #0284c7;
  border: 1px solid #bae6fd;
}

.badge-default {
  background: rgba(255, 255, 255, 0.05);
  color: #94a3b8;
}

:root[data-theme="light"] .badge-default {
  background: #f1f5f9;
  color: #64748b;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.15s ease;
}

:root[data-theme="light"] .action-btn {
  border: 1px solid #cbd5e1;
  color: #475569;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

:root[data-theme="light"] .action-btn:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.reset-btn:hover {
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.4);
}

/* Modals */
.recording-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(6px);
  outline: none;
}

.recording-modal-card {
  width: 90vw;
  max-width: 480px;
  background: #0d1424;
  border: 1px solid rgba(0, 229, 255, 0.3);
  border-radius: 12px;
  box-shadow:
    0 20px 40px rgba(0, 0, 0, 0.8),
    0 0 20px rgba(0, 229, 255, 0.15);
  overflow: hidden;
  color: #e2e8f0;
}

:root[data-theme="light"] .recording-modal-card {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  box-shadow: 0 20px 40px rgba(15, 23, 42, 0.15);
  color: #0f172a;
}

.modal-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .modal-card-header {
  border-bottom: 1px solid #e2e8f0;
}

.modal-icon {
  font-size: 18px;
  color: #00e5ff;
  margin-right: 8px;
}

:root[data-theme="light"] .modal-icon {
  color: #0284c7;
}

.modal-card-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  flex: 1;
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 4px;
}

.modal-card-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.target-cmd-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.target-title {
  font-size: 14px;
  font-weight: 600;
  color: #ffffff;
}

:root[data-theme="light"] .target-title {
  color: #0f172a;
}

.target-id {
  font-size: 12px;
  color: #94a3b8;
}

.recording-box {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80px;
  border: 2px dashed rgba(0, 229, 255, 0.3);
  border-radius: 8px;
  background: rgba(0, 229, 255, 0.04);
}

:root[data-theme="light"] .recording-box {
  border-color: #38bdf8;
  background: #f0f9ff;
}

.keycap.active-recording {
  font-size: 16px;
  padding: 6px 16px;
  color: #00e5ff;
  border-color: #00e5ff;
  box-shadow: 0 0 12px rgba(0, 229, 255, 0.3);
}

:root[data-theme="light"] .keycap.active-recording {
  color: #0284c7;
  border-color: #0284c7;
  box-shadow: 0 0 12px rgba(2, 132, 199, 0.2);
}

.recording-prompt {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #94a3b8;
  font-size: 13px;
}

.pulse-dot {
  color: #ef4444;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.4;
    transform: scale(0.9);
  }
}

.conflict-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  background: rgba(234, 179, 8, 0.12);
  border: 1px solid rgba(234, 179, 8, 0.25);
  color: #eab308;
  font-size: 12px;
}

.recording-tips {
  font-size: 12px;
  color: #94a3b8;
  text-align: center;
}

.modal-card-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  padding: 14px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0, 0, 0, 0.2);
}

:root[data-theme="light"] .modal-card-footer {
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
}

/* Raw JSON Modal */
.raw-json-modal-card {
  width: 90vw;
  max-width: 680px;
  height: 80vh;
  display: flex;
  flex-direction: column;
  background: #0d1424;
  border: 1px solid rgba(0, 229, 255, 0.3);
  border-radius: 12px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.8);
  overflow: hidden;
  color: #e2e8f0;
}

:root[data-theme="light"] .raw-json-modal-card {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

.raw-json-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 10px;
  overflow: hidden;
}

.json-textarea {
  flex: 1;
  width: 100%;
  padding: 12px;
  background: #111a2e;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #38bdf8;
  font-size: 13px;
  line-height: 1.5;
  resize: none;
  outline: none;
}

:root[data-theme="light"] .json-textarea {
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

.json-error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.25);
  color: #ef4444;
  font-size: 12px;
}
</style>
