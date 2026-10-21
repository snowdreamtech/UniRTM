<template>
  <div class="guard-view-container" id="quality-guard-view">
    <!-- Top Action Header -->
    <header class="guard-header">
      <div class="header-left">
        <div class="title-with-badge">
          <span class="codicon codicon-shield header-shield-icon"></span>
          <h1 class="header-title">{{ t("guard.title") || "Quality Guard" }}</h1>
          <span class="status-pill" :class="guardStatus.clean ? 'status-pill-clean' : 'status-pill-warning'">
            <span class="codicon" :class="guardStatus.clean ? 'codicon-pass' : 'codicon-warning'"></span>
            <span>{{
              guardStatus.clean
                ? t("guard.status.healthy") || "All Rules Healthy"
                : `${guardStatus.issues.length} Issues Found`
            }}</span>
          </span>
        </div>
        <p class="header-subtitle">
          {{
            t("guard.subtitle") ||
            "Automated code quality harness, file-extension pipelines, and pre-commit protection."
          }}
        </p>
      </div>

      <div class="header-right-actions">
        <!-- Run Check Staged -->
        <button
          class="btn-action primary"
          :disabled="isChecking"
          @click="handleCheck('staged')"
          title="Run quick staged scan (default pre-commit behavior)"
        >
          <span class="codicon" :class="isChecking ? 'codicon-loading spin' : 'codicon-play'"></span>
          <span>{{
            isChecking
              ? t("guard.actions.checking") || "Checking..."
              : t("guard.actions.check_staged") || "Check Staged"
          }}</span>
        </button>

        <!-- Run Check All -->
        <button
          class="btn-action secondary"
          :disabled="isChecking"
          @click="handleCheck('all')"
          title="Run full repository scan"
        >
          <span class="codicon codicon-run-all"></span>
          <span>{{ t("guard.actions.check_all") || "Check Entire Repo" }}</span>
        </button>

        <!-- One-click Auto Fix -->
        <button
          class="btn-action secondary"
          :disabled="isFixing"
          @click="handleFix"
          title="Automatically fix formatting and whitespace errors"
        >
          <span class="codicon" :class="isFixing ? 'codicon-loading spin' : 'codicon-wand'"></span>
          <span>{{
            isFixing ? t("guard.actions.fixing") || "Fixing..." : t("guard.actions.auto_fix") || "Auto-Fix"
          }}</span>
        </button>

        <!-- Add Tool/Extension Slot Button -->
        <button class="btn-action accent" @click="openAddSlotModal" title="Map a new file extension or external tool">
          <span class="codicon codicon-plus"></span>
          <span>{{ t("guard.actions.add_mapping") || "Add Extension Mapping" }}</span>
        </button>

        <!-- Fallback: View/Edit .unirtm.toml Button -->
        <button
          class="btn-action ghost"
          @click="openTomlFallback"
          title="Directly view or edit .unirtm.toml configuration"
        >
          <span class="codicon codicon-code"></span>
          <span>{{ t("guard.actions.edit_toml") || "Edit .unirtm.toml" }}</span>
        </button>
      </div>
    </header>

    <!-- Key Metrics Stat Row -->
    <div class="metrics-row">
      <div class="metric-card">
        <div class="metric-icon-box bg-blue">
          <span class="codicon codicon-symbol-ruler"></span>
        </div>
        <div class="metric-info">
          <span class="metric-val font-mono">{{ guardStatus.rules_count || 38 }}</span>
          <span class="metric-label">{{ t("guard.metrics.active_rules") || "Active Rules & Slots" }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon-box" :class="guardStatus.clean ? 'bg-green' : 'bg-amber'">
          <span class="codicon" :class="guardStatus.clean ? 'codicon-verified' : 'codicon-warning'"></span>
        </div>
        <div class="metric-info">
          <span class="metric-val font-mono" :class="guardStatus.clean ? 'text-green' : 'text-amber'">
            {{ guardStatus.clean ? "100%" : `${guardStatus.issues.length} Issues` }}
          </span>
          <span class="metric-label">{{ t("guard.metrics.quality_health") || "Quality Health" }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon-box bg-purple">
          <span class="codicon codicon-watch"></span>
        </div>
        <div class="metric-info">
          <span class="metric-val font-mono">{{
            guardStatus.duration_ms ? `${guardStatus.duration_ms}ms` : "--"
          }}</span>
          <span class="metric-label">{{ t("guard.metrics.scan_latency") || "Inspection Latency" }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon-box bg-cyan">
          <span class="codicon codicon-layers"></span>
        </div>
        <div class="metric-info">
          <span class="metric-val font-mono">{{ extensionRows.length }}</span>
          <span class="metric-label">{{ t("guard.metrics.pipelines") || "Extension Pipelines" }}</span>
        </div>
      </div>
    </div>

    <!-- View Switcher Tabs -->
    <div class="guard-view-tabs-bar">
      <div class="tabs-group">
        <button class="view-tab-btn" :class="{ active: activeTab === 'matrix' }" @click="activeTab = 'matrix'">
          <span class="codicon codicon-table"></span>
          <span>{{ t("guard.tabs.matrix") || "Extension Pipelines (Matrix)" }}</span>
        </button>

        <button class="view-tab-btn" :class="{ active: activeTab === 'slots' }" @click="activeTab = 'slots'">
          <span class="codicon codicon-symbol-class"></span>
          <span>{{ t("guard.tabs.slots") || "Tool Slots (CLI)" }}</span>
          <span class="tab-badge font-mono">{{ guardSlots.length }}</span>
        </button>

        <button class="view-tab-btn" :class="{ active: activeTab === 'reports' }" @click="activeTab = 'reports'">
          <span class="codicon codicon-checklist"></span>
          <span>{{ t("guard.tabs.reports") || "Live Streamed Checklist" }}</span>
          <span v-if="guardStatus.rule_reports && guardStatus.rule_reports.length > 0" class="tab-badge font-mono">
            {{ guardStatus.rule_reports.length }}
          </span>
        </button>
      </div>

      <!-- Quick Search Filter in Tabs -->
      <div class="filter-search-box">
        <span class="codicon codicon-search"></span>
        <input
          v-model="searchKeyword"
          type="text"
          class="search-input"
          :placeholder="t('guard.search_placeholder') || 'Filter extensions (.go, .ts) or tools...'"
        />
      </div>
    </div>

    <!-- Main Tab View Content Area -->
    <main class="guard-content-body">
      <!-- 1. Extension Pipelines Matrix Table (Primary View) -->
      <section v-if="activeTab === 'matrix'" class="tab-section matrix-section">
        <div class="table-responsive-wrapper">
          <table class="pipeline-matrix-table" id="table-extension-pipelines">
            <thead>
              <tr>
                <th class="col-ext">{{ t("guard.table.extension") || "File Type / Extension" }}</th>
                <th class="col-linter">{{ t("guard.table.linter") || "Linter (Static Audit)" }}</th>
                <th class="col-formatter">{{ t("guard.table.formatter") || "Formatter (Styling)" }}</th>
                <th class="col-fixer">{{ t("guard.table.fixer") || "Fixer (Auto Healing)" }}</th>
                <th class="col-security">{{ t("guard.table.security") || "Security & Audit" }}</th>
                <th class="col-actions text-right">{{ t("guard.table.actions") || "Actions" }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in filteredExtensionRows" :key="row.extension" class="pipeline-row">
                <!-- Extension Column -->
                <td class="col-ext">
                  <div class="ext-badge-group">
                    <span class="ext-tag font-mono">{{ row.extension }}</span>
                    <span class="lang-label">{{ row.languages.join(", ") }}</span>
                  </div>
                </td>

                <!-- Linter Column -->
                <td class="col-linter">
                  <div v-if="row.linter" class="tool-cell" @click="editSlot(row.linter)">
                    <span class="tool-name">{{ row.linter.default_tool_name }}</span>
                    <code class="tool-cmd font-mono">{{ row.linter.default_command }}</code>
                  </div>
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Formatter Column -->
                <td class="col-formatter">
                  <div v-if="row.formatter" class="tool-cell" @click="editSlot(row.formatter)">
                    <span class="tool-name">{{ row.formatter.default_tool_name }}</span>
                    <code class="tool-cmd font-mono">{{ row.formatter.default_command }}</code>
                  </div>
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Fixer Column -->
                <td class="col-fixer">
                  <div v-if="row.fixer" class="tool-cell" @click="editSlot(row.fixer)">
                    <span class="tool-name text-green">{{ row.fixer.default_tool_name }}</span>
                    <code class="tool-cmd font-mono text-green">{{
                      row.fixer.fix_command || row.fixer.default_command
                    }}</code>
                  </div>
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Security Column -->
                <td class="col-security">
                  <div v-if="row.auditor" class="tool-cell" @click="editSlot(row.auditor)">
                    <span class="tool-name text-purple">{{ row.auditor.default_tool_name }}</span>
                    <code class="tool-cmd font-mono">{{ row.auditor.default_command }}</code>
                  </div>
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Action Buttons -->
                <td class="col-actions text-right">
                  <div class="action-btn-row">
                    <button
                      class="row-btn edit-btn"
                      title="Edit mappings for this extension"
                      @click="handleEditRow(row)"
                    >
                      <span class="codicon codicon-edit"></span>
                    </button>
                    <button
                      class="row-btn add-btn"
                      title="Add tool slot to this extension"
                      @click="handleAddToolToExt(row.extension)"
                    >
                      <span class="codicon codicon-plus"></span>
                    </button>
                  </div>
                </td>
              </tr>

              <tr v-if="filteredExtensionRows.length === 0" class="empty-row">
                <td colspan="6" class="text-center py-4">
                  <span class="codicon codicon-search empty-icon"></span>
                  <p class="empty-text">
                    {{ t("guard.table.no_match") || "No extension pipeline matches the search query." }}
                  </p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- 2. Tool Slots Grid (CLI Tool Cards) -->
      <section v-if="activeTab === 'slots'" class="tab-section slots-section">
        <div class="slots-cards-grid">
          <div v-for="slot in filteredSlots" :key="slot.id" class="slot-card">
            <div class="slot-card-header">
              <div class="slot-title-wrap">
                <span class="slot-tool-name">{{ slot.default_tool_name }}</span>
                <span class="slot-id font-mono">{{ slot.id }}</span>
              </div>
              <div class="slot-header-tags">
                <span class="phase-pill" :class="'phase-' + slot.phase">{{ slot.phase }}</span>
                <button class="icon-edit-btn" title="Edit Slot" @click="editSlot(slot)">
                  <span class="codicon codicon-edit"></span>
                </button>
              </div>
            </div>

            <p class="slot-desc">{{ slot.description }}</p>

            <div class="slot-exts-wrap">
              <span class="ext-label">{{ t("guard.slots.extensions") || "Extensions:" }}</span>
              <div class="ext-tags-row">
                <span
                  v-for="ext in slot.default_extensions && slot.default_extensions.length > 0
                    ? slot.default_extensions
                    : ['*']"
                  :key="ext"
                  class="ext-chip font-mono"
                >
                  {{ ext }}
                </span>
              </div>
            </div>

            <div class="slot-commands-box">
              <div class="cmd-line" v-if="slot.default_command">
                <span class="cmd-label">{{ t("guard.slots.check") || "Check:" }}</span>
                <code class="cmd-text font-mono">{{ slot.default_command }}</code>
              </div>
              <div class="cmd-line" v-if="slot.fix_command">
                <span class="cmd-label text-green">{{ t("guard.slots.fix") || "Fix:" }}</span>
                <code class="cmd-text font-mono text-green">{{ slot.fix_command }}</code>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 3. Live Streamed Checklist & Reports -->
      <section v-if="activeTab === 'reports'" class="tab-section reports-section">
        <div class="reports-container">
          <div class="reports-header">
            <h3 class="reports-title">{{ t("guard.diagnostics.checklist") || "Transparent Execution Checklist" }}</h3>
            <span class="reports-sub">
              {{ t("guard.reports_desc") || "Real-time streaming validation status and execution duration per rule." }}
            </span>
          </div>

          <div v-if="guardStatus.rule_reports && guardStatus.rule_reports.length > 0" class="checklist-items-grid">
            <div
              v-for="report in guardStatus.rule_reports"
              :key="report.rule_name"
              class="checklist-item"
              :class="
                report.status === 'error' ? 'item-error' : report.status === 'warning' ? 'item-warning' : 'item-pass'
              "
            >
              <div class="item-left">
                <span
                  class="codicon"
                  :class="
                    report.status === 'error'
                      ? 'codicon-error text-red'
                      : report.status === 'warning'
                        ? 'codicon-warning text-amber'
                        : 'codicon-pass text-green'
                  "
                ></span>
                <span class="item-rule-name font-mono">{{ report.rule_name }}</span>
              </div>
              <div class="item-right">
                <span class="item-phase-badge" :class="'phase-' + report.phase">{{ report.phase }}</span>
                <span class="item-duration font-mono">{{ report.duration_ms }}ms</span>
              </div>
            </div>
          </div>

          <div v-else class="empty-reports-box">
            <span class="codicon codicon-checklist empty-reports-icon"></span>
            <p>
              {{
                t("guard.no_reports") ||
                "No inspection reports yet. Run 'Check Staged' or 'Check Entire Repo' to see live results."
              }}
            </p>
          </div>
        </div>
      </section>
    </main>

    <!-- Extension / Tool Slot CRUD Modal -->
    <ExtensionSlotModal
      :isOpen="isSlotModalOpen"
      :initialSlot="activeEditingSlot"
      :targetExtension="targetExtForAdd"
      @close="isSlotModalOpen = false"
      @save="onSaveSlot"
      @delete="onDeleteSlot"
    />

    <!-- Toml Config Fallback Drawer -->
    <TomlConfigDrawer :isOpen="isTomlDrawerOpen" @close="isTomlDrawerOpen = false" @saved="onTomlSaved" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t } from "../i18n";
import { useGuard, type ToolSlotItem, type SlotConfigInput } from "../composables/useGuard";
import ExtensionSlotModal from "../components/guard/ExtensionSlotModal.vue";
import TomlConfigDrawer from "../components/guard/TomlConfigDrawer.vue";

const {
  guardStatus,
  guardSlots,
  isChecking,
  isFixing,
  fetchStatus,
  fetchSlots,
  runCheck,
  runFix,
  saveSlotConfig,
  deleteSlotConfig,
} = useGuard();

const isSlotModalOpen = ref(false);
const activeEditingSlot = ref<ToolSlotItem | null>(null);
const targetExtForAdd = ref<string>("");

const isTomlDrawerOpen = ref(false);

const activeTab = ref<"matrix" | "slots" | "reports">("matrix");
const searchKeyword = ref("");

export interface ExtensionPipelineRow {
  extension: string;
  languages: string[];
  linter?: ToolSlotItem;
  formatter?: ToolSlotItem;
  fixer?: ToolSlotItem;
  auditor?: ToolSlotItem;
  allSlots: ToolSlotItem[];
}

onMounted(() => {
  fetchStatus().catch(() => {});
  fetchSlots().catch(() => {});
});

function handleCheck(mode: "staged" | "all") {
  runCheck({ mode }).catch(() => {});
}

function handleFix() {
  runFix().catch(() => {});
}

// Map extensions to human readable language tags
function getLanguagesForExt(ext: string): string[] {
  switch (ext) {
    case ".go":
      return ["Go"];
    case ".ts":
    case ".tsx":
      return ["TypeScript"];
    case ".js":
    case ".jsx":
      return ["JavaScript"];
    case ".vue":
      return ["Vue"];
    case ".sh":
    case ".bash":
      return ["Shell"];
    case ".toml":
      return ["TOML"];
    case ".json":
      return ["JSON"];
    case ".md":
      return ["Markdown"];
    case "*":
      return ["Global / Any"];
    default:
      return [ext.replace(".", "").toUpperCase()];
  }
}

// Compute aggregate rows grouped by file extension
const extensionRows = computed<ExtensionPipelineRow[]>(() => {
  const rowMap = new Map<string, ExtensionPipelineRow>();

  // Ensure primary supported extensions exist by default
  const defaultExts = [".go", ".ts", ".vue", ".sh", ".toml", "*"];
  for (const ext of defaultExts) {
    rowMap.set(ext, {
      extension: ext,
      languages: getLanguagesForExt(ext),
      allSlots: [],
    });
  }

  // Iterate over guardSlots and associate with extension
  for (const slot of guardSlots.value) {
    const exts = slot.default_extensions && slot.default_extensions.length > 0 ? slot.default_extensions : ["*"];

    for (const ext of exts) {
      if (!rowMap.has(ext)) {
        rowMap.set(ext, {
          extension: ext,
          languages: getLanguagesForExt(ext),
          allSlots: [],
        });
      }

      const row = rowMap.get(ext)!;
      row.allSlots.push(slot);

      if (slot.category === "format") {
        row.formatter = slot;
        if (slot.fix_command || slot.phase === "phase1_fixer") {
          row.fixer = slot;
        }
      } else if (slot.category === "lint") {
        row.linter = slot;
      } else if (slot.category === "secret" || slot.category === "security") {
        row.auditor = slot;
      }
    }
  }

  return Array.from(rowMap.values());
});

const filteredExtensionRows = computed(() => {
  if (!searchKeyword.value.trim()) return extensionRows.value;
  const q = searchKeyword.value.toLowerCase().trim();
  return extensionRows.value.filter((row) => {
    return (
      row.extension.toLowerCase().includes(q) ||
      row.languages.some((l) => l.toLowerCase().includes(q)) ||
      (row.linter && row.linter.default_tool_name.toLowerCase().includes(q)) ||
      (row.formatter && row.formatter.default_tool_name.toLowerCase().includes(q)) ||
      (row.fixer && row.fixer.default_tool_name.toLowerCase().includes(q)) ||
      (row.auditor && row.auditor.default_tool_name.toLowerCase().includes(q))
    );
  });
});

const filteredSlots = computed(() => {
  if (!searchKeyword.value.trim()) return guardSlots.value;
  const q = searchKeyword.value.toLowerCase().trim();
  return guardSlots.value.filter((slot) => {
    return (
      slot.id.toLowerCase().includes(q) ||
      slot.default_tool_name.toLowerCase().includes(q) ||
      (slot.default_extensions && slot.default_extensions.some((e) => e.toLowerCase().includes(q)))
    );
  });
});

function openAddSlotModal() {
  activeEditingSlot.value = null;
  targetExtForAdd.value = "";
  isSlotModalOpen.value = true;
}

function openTomlFallback() {
  isTomlDrawerOpen.value = true;
}

function editSlot(slot: ToolSlotItem) {
  activeEditingSlot.value = slot;
  targetExtForAdd.value = "";
  isSlotModalOpen.value = true;
}

function handleEditRow(row: ExtensionPipelineRow) {
  const primarySlot = row.linter || row.formatter || row.fixer || row.auditor || row.allSlots[0];
  if (primarySlot) {
    editSlot(primarySlot);
  } else {
    handleAddToolToExt(row.extension);
  }
}

function handleAddToolToExt(ext: string) {
  activeEditingSlot.value = null;
  targetExtForAdd.value = ext;
  isSlotModalOpen.value = true;
}

async function onSaveSlot(slot: SlotConfigInput) {
  const ok = await saveSlotConfig(slot);
  if (ok) {
    isSlotModalOpen.value = false;
    await fetchSlots();
  }
}

async function onDeleteSlot(slotId: string) {
  const ok = await deleteSlotConfig(slotId);
  if (ok) {
    isSlotModalOpen.value = false;
    await fetchSlots();
  }
}

async function onTomlSaved() {
  await fetchSlots();
  await fetchStatus();
}
</script>

<style scoped>
.guard-view-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  padding: 1.25rem 1.5rem;
  box-sizing: border-box;
  overflow-y: auto;
  gap: 1.25rem;
}

/* Header */
.guard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--card-border, #333);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-shield-icon {
  font-size: 1.5rem;
  color: var(--vscode-button-background, #388bfd);
}

.header-title {
  font-size: 1.375rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-primary, #fff);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.625rem;
  border-radius: 9999px;
}

.status-pill-clean {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
  border: 1px solid rgba(46, 160, 67, 0.3);
}

.status-pill-warning {
  background: rgba(248, 81, 73, 0.15);
  color: #f85149;
  border: 1px solid rgba(248, 81, 73, 0.3);
}

.header-subtitle {
  font-size: 0.8125rem;
  color: var(--text-secondary, #8b949e);
  margin: 0;
}

.header-right-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4375rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  border: 1px solid transparent;
}

.btn-action.primary {
  background: var(--vscode-button-background, #0078d4);
  color: var(--vscode-button-foreground, #fff);
}

.btn-action.primary:hover:not(:disabled) {
  background: var(--vscode-button-hoverBackground, #026ec1);
}

.btn-action.secondary {
  background: var(--vscode-button-secondaryBackground, #3a3d41);
  color: var(--vscode-button-secondaryForeground, #fff);
  border: 1px solid var(--card-border, #45494e);
}

.btn-action.secondary:hover:not(:disabled) {
  background: var(--vscode-button-secondaryHoverBackground, #45494e);
}

.btn-action.accent {
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
  border: 1px solid rgba(56, 139, 253, 0.3);
}

.btn-action.accent:hover:not(:disabled) {
  background: rgba(56, 139, 253, 0.25);
}

.btn-action.ghost {
  background: transparent;
  color: var(--text-secondary, #8b949e);
  border: 1px solid var(--card-border, #333);
}

.btn-action.ghost:hover:not(:disabled) {
  background: var(--hover-bg, rgba(255, 255, 255, 0.05));
  color: var(--text-primary, #fff);
}

.btn-action:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Metrics Row */
.metrics-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.metric-card {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 0.875rem 1rem;
  background: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, #2d2d2d);
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.metric-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  font-size: 1.25rem;
}

.bg-blue {
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
}
.bg-green {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
}
.bg-amber {
  background: rgba(210, 153, 34, 0.15);
  color: #d29922;
}
.bg-purple {
  background: rgba(187, 128, 255, 0.15);
  color: #bc8cff;
}
.bg-cyan {
  background: rgba(57, 197, 207, 0.15);
  color: #39c5cf;
}

.metric-info {
  display: flex;
  flex-direction: column;
}

.metric-val {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--text-primary, #fff);
}

.metric-label {
  font-size: 0.75rem;
  color: var(--text-secondary, #8b949e);
}

/* View Switcher Tabs Bar */
.guard-view-tabs-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--card-border, #333);
  padding-bottom: 0.5rem;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.tabs-group {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.view-tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4375rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-secondary, #8b949e);
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.view-tab-btn:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.05));
  color: var(--text-primary, #fff);
}

.view-tab-btn.active {
  background: var(--selected-bg, rgba(0, 120, 212, 0.15));
  color: var(--vscode-button-background, #388bfd);
  font-weight: 600;
}

.tab-badge {
  font-size: 0.6875rem;
  padding: 0.1rem 0.375rem;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 9999px;
  margin-left: 0.25rem;
}

.filter-search-box {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.3125rem 0.625rem;
  background: var(--input-bg, #181818);
  border: 1px solid var(--input-border, #333);
  border-radius: 6px;
  min-width: 260px;
}

.search-input {
  background: transparent;
  border: none;
  color: var(--text-primary, #fff);
  font-size: 0.8125rem;
  outline: none;
  width: 100%;
}

/* Main Content Body */
.guard-content-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Table Design */
.table-responsive-wrapper {
  overflow-x: auto;
  background: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, #2d2d2d);
  border-radius: 8px;
}

.pipeline-matrix-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
  text-align: left;
}

.pipeline-matrix-table th {
  padding: 0.75rem 1rem;
  background: var(--header-bg, #252526);
  color: var(--text-secondary, #8b949e);
  font-weight: 600;
  border-bottom: 1px solid var(--card-border, #333);
}

.pipeline-matrix-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--card-border, #282828);
  vertical-align: middle;
}

.pipeline-row:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.02));
}

.ext-badge-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.ext-tag {
  font-size: 0.8125rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
  border-radius: 4px;
}

.lang-label {
  font-size: 0.75rem;
  color: var(--text-secondary, #8b949e);
}

.tool-cell {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  cursor: pointer;
  padding: 0.25rem 0.4rem;
  border-radius: 4px;
  transition: background 0.15s ease;
}

.tool-cell:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.05));
}

.tool-name {
  font-weight: 600;
  color: var(--text-primary, #fff);
}

.tool-cmd {
  font-size: 0.6875rem;
  color: var(--text-secondary, #8b949e);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
}

.cell-empty {
  color: var(--text-secondary, #555);
}

.action-btn-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.375rem;
}

.row-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  background: transparent;
  border: 1px solid var(--card-border, #3a3a3a);
  color: var(--text-secondary, #8b949e);
  cursor: pointer;
  transition: all 0.15s ease;
}

.row-btn:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.1));
  color: var(--text-primary, #fff);
}

/* Slots Grid */
.slots-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1rem;
}

.slot-card {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  padding: 1rem;
  background: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, #2d2d2d);
  border-radius: 8px;
  transition: border-color 0.2s ease;
}

.slot-card:hover {
  border-color: var(--card-border-hover, #444);
}

.slot-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.slot-title-wrap {
  display: flex;
  flex-direction: column;
}

.slot-tool-name {
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--text-primary, #fff);
}

.slot-id {
  font-size: 0.6875rem;
  color: var(--text-secondary, #8b949e);
}

.slot-header-tags {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.phase-pill {
  font-size: 0.6875rem;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  font-weight: 600;
}

.phase-phase1_fixer {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
}
.phase-phase2_formatter {
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
}
.phase-phase3_linter {
  background: rgba(210, 153, 34, 0.15);
  color: #d29922;
}

.icon-edit-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary, #8b949e);
  cursor: pointer;
  padding: 0.2rem;
  border-radius: 4px;
}

.icon-edit-btn:hover {
  color: var(--text-primary, #fff);
  background: var(--hover-bg, rgba(255, 255, 255, 0.1));
}

.slot-desc {
  font-size: 0.75rem;
  color: var(--text-secondary, #8b949e);
  margin: 0;
  line-height: 1.4;
}

.slot-exts-wrap {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.ext-label {
  font-size: 0.6875rem;
  color: var(--text-secondary, #8b949e);
}

.ext-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.ext-chip {
  font-size: 0.6875rem;
  padding: 0.1rem 0.3rem;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 3px;
  color: var(--text-primary, #ccc);
}

.slot-commands-box {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.5rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
}

.cmd-line {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.6875rem;
}

.cmd-label {
  font-weight: 600;
  color: var(--text-secondary, #888);
}

.cmd-text {
  color: var(--text-primary, #eee);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Reports Section */
.reports-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.reports-header {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.reports-title {
  font-size: 1rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary, #fff);
}

.reports-sub {
  font-size: 0.75rem;
  color: var(--text-secondary, #8b949e);
}

.checklist-items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 0.625rem;
}

.checklist-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.625rem 0.75rem;
  background: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, #2d2d2d);
  border-radius: 6px;
}

.item-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.item-rule-name {
  font-size: 0.75rem;
  color: var(--text-primary, #fff);
}

.item-right {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.item-phase-badge {
  font-size: 0.625rem;
  padding: 0.1rem 0.25rem;
  border-radius: 3px;
}

.item-duration {
  font-size: 0.6875rem;
  color: var(--text-secondary, #8b949e);
}

.empty-reports-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  background: var(--card-bg, #1e1e1e);
  border: 1px dashed var(--card-border, #333);
  border-radius: 8px;
  color: var(--text-secondary, #8b949e);
  font-size: 0.8125rem;
  gap: 0.5rem;
}

.empty-reports-icon {
  font-size: 2rem;
  opacity: 0.5;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% {
    transform: rotate(360deg);
  }
}
</style>
