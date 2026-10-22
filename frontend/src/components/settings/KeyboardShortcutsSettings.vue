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
          <span class="btn-text">{{ isZh ? "打开 keybindings.json" : "Open JSON" }}</span>
        </button>
        <button
          v-if="hasCustomKeybindings"
          class="btn btn-outline-danger"
          :title="isZh ? '重置当前范围所有自定义快捷键为系统默认' : 'Reset all shortcuts in current scope to defaults'"
          @click="handleResetAll"
        >
          <span class="codicon codicon-discard"></span>
          <span class="btn-text">{{ isZh ? "恢复全部默认" : "Reset All" }}</span>
        </button>
      </div>
    </div>

    <!-- Scope Switcher: Workspace vs User -->
    <div class="scope-tabs-bar">
      <div class="scope-tabs">
        <button
          class="scope-tab"
          :class="{ active: currentScope === 'workspace' }"
          @click="handleScopeChange('workspace')"
        >
          <span class="codicon codicon-folder"></span>
          <span class="scope-tab-title">{{ isZh ? "工作区快捷键 (当前项目)" : "Workspace Keybindings" }}</span>
          <span class="scope-tab-path font-mono">.unirtm/keybindings.json</span>
        </button>
        <button class="scope-tab" :class="{ active: currentScope === 'user' }" @click="handleScopeChange('user')">
          <span class="codicon codicon-account"></span>
          <span class="scope-tab-title">{{ isZh ? "用户全局快捷键" : "User Global Keybindings" }}</span>
          <span class="scope-tab-path font-mono">~/.config/unirtm/keybindings.json</span>
        </button>
      </div>
    </div>

    <!-- Keymap Presets Bar (VS Code / JetBrains / Vim / Sublime) -->
    <div class="presets-bar">
      <div class="presets-label">
        <span class="codicon codicon-layout"></span>
        <span>{{ isZh ? "预设快捷键方案 (Keymap Presets):" : "Keymap Presets:" }}</span>
      </div>
      <div class="preset-buttons">
        <button
          v-for="preset in keymapPresets"
          :key="preset.id"
          class="preset-btn"
          :title="isZh ? preset.descZh : preset.descEn"
          @click="handleApplyPreset(preset.id)"
        >
          <span class="codicon" :class="getPresetIcon(preset.id)"></span>
          <span class="preset-name">{{ isZh ? preset.nameZh : preset.nameEn }}</span>
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
          {{ isZh ? "自定义 (" + customCount + ")" : "Custom (" + customCount + ")" }}
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
                {{
                  isZh
                    ? currentScope === "workspace"
                      ? "Workspace (项目)"
                      : "User (用户)"
                    : currentScope === "workspace"
                      ? "Workspace"
                      : "User"
                }}
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

          <tr v-if="filteredList.length === 0">
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

            <!-- Scope info in modal -->
            <div class="modal-scope-hint font-mono">
              <span class="codicon codicon-info"></span>
              <span
                >{{ isZh ? "将保存至: " : "Will save to: " }}<strong>{{ keybindingsPath }}</strong></span
              >
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
            <div v-if="jsonError" class="json-error-banner font-mono">
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
              {{ isZh ? "取消" : "Cancel" }}
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
import { ref, computed, watch, onMounted } from "vue";
import {
  useKeyboardShortcuts,
  normalizeKeyComboFromEvent,
  type ActiveShortcutItem,
} from "../../composables/useKeyboardShortcuts";
import { selectedLangSetting } from "../../i18n";
import { useToast } from "../../composables/useToast";

const props = withDefaults(
  defineProps<{
    activeScope?: "workspace" | "user";
  }>(),
  {
    activeScope: "workspace",
  }
);

const {
  activeShortcuts,
  currentScope,
  keybindingsPath,
  rawKeybindingsJson,
  keymapPresets,
  fetchKeybindings,
  switchScope,
  saveKeybinding,
  resetKeybinding,
  resetAllKeybindings,
  saveRawKeybindings,
  applyPreset,
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

watch(
  () => props.activeScope,
  (newScope) => {
    if (newScope && newScope !== currentScope.value) {
      switchScope(newScope);
    }
  }
);

function getPresetIcon(id: string): string {
  switch (id) {
    case "jetbrains":
      return "codicon-symbol-class";
    case "vim":
      return "codicon-terminal-linux";
    case "sublime":
      return "codicon-file-code";
    default:
      return "codicon-vscode";
  }
}

async function handleScopeChange(scope: "workspace" | "user") {
  await switchScope(scope);
  showToast(
    isZh.value
      ? `已切换到 ${scope === "workspace" ? "工作区快捷键 (.unirtm/keybindings.json)" : "用户全局快捷键 (~/.config/unirtm/keybindings.json)"}`
      : `Switched to ${scope === "workspace" ? "Workspace Keybindings" : "User Global Keybindings"}`,
    "info"
  );
}

async function handleApplyPreset(presetId: "vscode" | "jetbrains" | "vim" | "sublime") {
  const preset = keymapPresets.find((p) => p.id === presetId);
  const name = isZh.value ? preset?.nameZh : preset?.nameEn;
  const msg = isZh.value
    ? `确定要将当前范围 (${currentScope.value === "workspace" ? "工作区" : "全局用户"}) 快捷键切换为「${name}」预设方案吗？`
    : `Apply "${name}" keymap preset to ${currentScope.value} scope?`;

  if (!confirm(msg)) return;

  const ok = await applyPreset(presetId, currentScope.value);
  if (ok) {
    showToast(isZh.value ? `已成功应用「${name}」方案！` : `Successfully applied "${name}" preset!`, "success");
  } else {
    showToast(isZh.value ? "应用预设方案失败" : "Failed to apply preset", "error");
  }
}

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
  if (
    !confirm(
      isZh.value ? "确定要恢复当前范围的所有快捷键为默认设置吗？" : "Reset keybindings in this scope to defaults?"
    )
  ) {
    return;
  }
  const ok = await resetAllKeybindings();
  if (ok) {
    showToast(isZh.value ? "当前范围快捷键已恢复为默认设置" : "Keybindings reset to defaults", "success");
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
  fetchKeybindings(props.activeScope);
});
</script>

<style scoped>
.shortcuts-settings-wrapper {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.section-title-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
}

.section-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: rgba(0, 229, 255, 0.1);
  color: #00e5ff;
  font-size: 20px;
}

:root[data-theme="light"] .section-icon {
  background: rgba(2, 132, 199, 0.1);
  color: #0284c7;
}

.section-title-wrap h2 {
  font-size: 18px;
  font-weight: 700;
  margin: 0;
  color: #ffffff;
}

:root[data-theme="light"] .section-title-wrap h2 {
  color: #0f172a;
}

.section-sub {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
  display: inline-block;
}

:root[data-theme="light"] .section-sub {
  color: #64748b;
}

.file-path-badge {
  color: #00e5ff;
  background: rgba(0, 229, 255, 0.1);
  padding: 1px 6px;
  border-radius: 4px;
}

:root[data-theme="light"] .file-path-badge {
  color: #0284c7;
  background: rgba(2, 132, 199, 0.1);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Scope Tabs Bar */
.scope-tabs-bar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.scope-tabs {
  display: flex;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 3px;
  gap: 4px;
}

:root[data-theme="light"] .scope-tabs {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
}

.scope-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
}

:root[data-theme="light"] .scope-tab {
  color: #64748b;
}

.scope-tab:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.04);
}

:root[data-theme="light"] .scope-tab:hover {
  color: #0f172a;
  background: #e2e8f0;
}

.scope-tab.active {
  background: #00e5ff;
  color: #0b101e;
  font-weight: 600;
}

:root[data-theme="light"] .scope-tab.active {
  background: #0284c7;
  color: #ffffff;
}

.scope-tab-title {
  white-space: nowrap;
}

.scope-tab-path {
  font-size: 11px;
  opacity: 0.8;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.2);
}

.scope-tab.active .scope-tab-path {
  background: rgba(0, 0, 0, 0.15);
  color: inherit;
}

/* Presets Bar */
.presets-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

:root[data-theme="light"] .presets-bar {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}

.presets-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
  white-space: nowrap;
}

:root[data-theme="light"] .presets-label {
  color: #475569;
}

.preset-buttons {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.preset-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(15, 23, 42, 0.4);
  color: #cbd5e1;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

:root[data-theme="light"] .preset-btn {
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #334155;
}

.preset-btn:hover {
  border-color: #00e5ff;
  color: #00e5ff;
  background: rgba(0, 229, 255, 0.08);
}

:root[data-theme="light"] .preset-btn:hover {
  border-color: #0284c7;
  color: #0284c7;
  background: #e0f2fe;
}

/* Toolbar */
.shortcuts-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.search-input-wrap {
  position: relative;
  flex: 1;
  min-width: 260px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #64748b;
  font-size: 14px;
}

.filter-input {
  width: 100%;
  padding: 8px 36px;
  background: rgba(13, 20, 36, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #ffffff;
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s ease;
}

:root[data-theme="light"] .filter-input {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

.filter-input:focus {
  border-color: #00e5ff;
}

:root[data-theme="light"] .filter-input:focus {
  border-color: #0284c7;
}

.clear-btn {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
}

.clear-btn:hover {
  color: #ffffff;
}

:root[data-theme="light"] .clear-btn:hover {
  color: #0f172a;
}

.source-filter-pills {
  display: flex;
  gap: 6px;
}

.filter-pill {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

:root[data-theme="light"] .filter-pill {
  border: 1px solid #cbd5e1;
  color: #64748b;
}

.filter-pill:hover {
  color: #ffffff;
  border-color: rgba(255, 255, 255, 0.2);
}

:root[data-theme="light"] .filter-pill:hover {
  color: #0f172a;
  border-color: #94a3b8;
}

.filter-pill.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: #00e5ff;
  color: #00e5ff;
  font-weight: 600;
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
  min-width: 820px;
  border-collapse: collapse;
  text-align: left;
  table-layout: fixed;
}

.shortcuts-table th {
  padding: 10px 16px;
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0, 0, 0, 0.2);
  white-space: nowrap;
}

:root[data-theme="light"] .shortcuts-table th {
  color: #64748b;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}

/* Column Widths & Layout */
.col-cmd {
  width: 32%;
  min-width: 220px;
}

.col-id {
  width: 26%;
  min-width: 180px;
}

.col-key {
  width: 20%;
  min-width: 140px;
}

.col-source {
  width: 10%;
  min-width: 80px;
}

.col-actions {
  width: 12%;
  min-width: 135px;
  text-align: right !important;
  white-space: nowrap;
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
  font-size: 11px;
  color: #64748b;
}

:root[data-theme="light"] .cmd-desc {
  color: #64748b;
}

.cmd-id-pill {
  font-size: 11px;
  color: #94a3b8;
  background: rgba(255, 255, 255, 0.05);
  padding: 3px 8px;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:root[data-theme="light"] .cmd-id-pill {
  color: #475569;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
}

.keys-display {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
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
  white-space: nowrap;
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
  white-space: nowrap;
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

/* Action Buttons Styling - Prevents wrapping */
.action-buttons {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  white-space: nowrap;
  flex-wrap: nowrap;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
  flex-shrink: 0;
  line-height: 1.2;
}

:root[data-theme="light"] .action-btn {
  border: 1px solid #cbd5e1;
  color: #475569;
}

.btn-text {
  white-space: nowrap;
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
  backdrop-filter: blur(4px);
}

.recording-modal-card {
  width: 90%;
  max-width: 480px;
  background: #0f172a;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

:root[data-theme="light"] .recording-modal-card {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  box-shadow: 0 20px 40px rgba(15, 23, 42, 0.15);
}

.modal-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .modal-card-header {
  border-bottom: 1px solid #e2e8f0;
}

.modal-icon {
  font-size: 18px;
  color: #00e5ff;
}

:root[data-theme="light"] .modal-icon {
  color: #0284c7;
}

.modal-card-header h3 {
  flex: 1;
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #ffffff;
}

:root[data-theme="light"] .modal-card-header h3 {
  color: #0f172a;
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  font-size: 16px;
}

.modal-close-btn:hover {
  color: #ffffff;
}

:root[data-theme="light"] .modal-close-btn:hover {
  color: #0f172a;
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
  font-size: 11px;
  color: #64748b;
}

.recording-box {
  min-height: 80px;
  border: 2px dashed rgba(0, 229, 255, 0.4);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 229, 255, 0.03);
  padding: 16px;
}

:root[data-theme="light"] .recording-box {
  border-color: #38bdf8;
  background: #f0f9ff;
}

.recording-prompt {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #94a3b8;
  font-size: 13px;
}

:root[data-theme="light"] .recording-prompt {
  color: #475569;
}

.pulse-dot {
  color: #ef4444;
  animation: pulse 1.2s infinite ease-in-out;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.4;
    transform: scale(0.85);
  }
}

.keycap.active-recording {
  font-size: 16px;
  padding: 6px 14px;
  background: #00e5ff;
  color: #0b101e;
  border: none;
  box-shadow: 0 4px 12px rgba(0, 229, 255, 0.3);
}

.modal-scope-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: #94a3b8;
}

:root[data-theme="light"] .modal-scope-hint {
  color: #64748b;
}

.conflict-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  background: rgba(234, 179, 8, 0.1);
  border: 1px solid rgba(234, 179, 8, 0.25);
  color: #eab308;
  font-size: 12px;
}

.recording-tips {
  font-size: 11px;
  color: #64748b;
  text-align: center;
}

.modal-card-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0, 0, 0, 0.15);
}

:root[data-theme="light"] .modal-card-footer {
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
}

/* Raw JSON Modal */
.raw-json-modal-card {
  width: 90%;
  max-width: 640px;
  background: #0f172a;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

:root[data-theme="light"] .raw-json-modal-card {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  box-shadow: 0 20px 40px rgba(15, 23, 42, 0.15);
}

.raw-json-modal-card .header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.raw-json-modal-card .file-path {
  font-size: 11px;
  color: #64748b;
}

.raw-json-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.json-textarea {
  width: 100%;
  height: 320px;
  background: #0b101e;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  padding: 12px;
  color: #e2e8f0;
  font-size: 12px;
  line-height: 1.5;
  resize: vertical;
  outline: none;
}

:root[data-theme="light"] .json-textarea {
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

.json-textarea:focus {
  border-color: #00e5ff;
}

:root[data-theme="light"] .json-textarea:focus {
  border-color: #0284c7;
}

.json-error-banner {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border-radius: 4px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #ef4444;
  font-size: 12px;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 36px 16px;
  color: #64748b;
  gap: 8px;
}

.empty-icon {
  font-size: 28px;
}
</style>
