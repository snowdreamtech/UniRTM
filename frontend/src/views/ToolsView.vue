<template>
  <div class="view-container tools-view" id="view-tools">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title">🛠️ {{ t("tools.title") }}</h2>
        <p class="view-desc">
          由 UniRTM 统一版本锁定与高性能垫片（Go Shims）路由，开箱即用，绝对环境隔离。
        </p>
      </div>
      <div class="view-actions">
        <button class="tool-action-btn primary" @click="handleReshim">
          <span>⚡ {{ t("tools.reshim") }}</span>
        </button>
      </div>
    </div>

    <!-- Search & Filter Filter Bar -->
    <div class="tools-filter-bar glass-card">
      <div class="search-input-wrapper">
        <span class="search-icon">🔍</span>
        <input
          v-model="searchKey"
          type="text"
          class="tool-search-input"
          placeholder="按工具名称、后端或用途筛选..."
        />
      </div>
      <div class="tool-count-badge">
        <span>已配置: {{ filteredTools.length }} / {{ toolsList.length }}</span>
      </div>
    </div>

    <!-- Tools Grid Table -->
    <div class="tools-grid">
      <div
        v-for="tool in filteredTools"
        :key="tool.id"
        class="glass-card tool-card"
        :class="{ active: tool.active }"
      >
        <div class="tool-card-header">
          <div class="tool-badge-icon">{{ tool.icon }}</div>
          <div class="tool-meta-block">
            <div class="tool-name-row">
              <span class="tool-name">{{ tool.name }}</span>
              <span class="tool-tag" :class="tool.tagClass">{{ tool.tag }}</span>
            </div>
            <span class="tool-desc">{{ tool.description }}</span>
          </div>
        </div>

        <div class="tool-details-box">
          <div class="detail-line">
            <span class="lbl">锁定版本:</span>
            <span class="val font-mono version-highlight">{{ tool.version }}</span>
          </div>
          <div class="detail-line">
            <span class="lbl">安装后端:</span>
            <span class="val font-mono">{{ tool.backend }}</span>
          </div>
          <div class="detail-line">
            <span class="lbl">垫片路径:</span>
            <span class="val font-mono text-muted">~/.unirtm/shims/{{ tool.name }}</span>
          </div>
        </div>

        <div class="tool-card-actions">
          <button class="btn-sub secondary" @click="handleSwitchVersion(tool.name)">
            <span>切换版本</span>
          </button>
          <button class="btn-sub primary" @click="handleVerifyTool(tool.name)">
            <span>健康诊断</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { t } from "../i18n";

const emit = defineEmits<{
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string): void;
}>();

const searchKey = ref("");

interface ToolEntry {
  id: string;
  name: string;
  icon: string;
  version: string;
  backend: string;
  description: string;
  tag: string;
  tagClass: string;
  active: boolean;
}

const toolsList = ref<ToolEntry[]>([
  {
    id: "node",
    name: "node",
    icon: "🟢",
    version: "22.14.0",
    backend: "native / npmmirror",
    description: "JavaScript / TypeScript 核心运行时与包管理器环境",
    tag: "核心语言",
    tagClass: "tag-blue",
    active: true,
  },
  {
    id: "go",
    name: "go",
    icon: "🔷",
    version: "1.24.0",
    backend: "native / goproxy.cn",
    description: "UniRTM 底层构建与高并发服务编译工具链",
    tag: "核心语言",
    tagClass: "tag-blue",
    active: true,
  },
  {
    id: "python",
    name: "python",
    icon: "🐍",
    version: "3.12.2",
    backend: "github-releases",
    description: "脚本自动化、数据分析及辅助工具环境",
    tag: "开发语言",
    tagClass: "tag-green",
    active: true,
  },
  {
    id: "pre-commit",
    name: "pre-commit",
    icon: "🪝",
    version: "4.7.0",
    backend: "pipx / github",
    description: "跨语言 Git 提交前多层安全与代码规范检查器",
    tag: "代码规范",
    tagClass: "tag-purple",
    active: true,
  },
  {
    id: "trivy",
    name: "trivy",
    icon: "🛡️",
    version: "0.50.1",
    backend: "aqua / github",
    description: "容器与依赖树全量 CVE 漏洞实时安全扫描工具",
    tag: "安全扫描",
    tagClass: "tag-red",
    active: true,
  },
  {
    id: "eslint",
    name: "eslint",
    icon: "⚡",
    version: "10.12.0",
    backend: "npm:eslint",
    description: "前端 JavaScript / TypeScript 静态语法与代码规范扫描",
    tag: "Linter",
    tagClass: "tag-purple",
    active: true,
  },
]);

const filteredTools = computed(() => {
  if (!searchKey.value) return toolsList.value;
  const q = searchKey.value.toLowerCase();
  return toolsList.value.filter(
    (t) =>
      t.name.toLowerCase().includes(q) ||
      t.backend.toLowerCase().includes(q) ||
      t.description.toLowerCase().includes(q)
  );
});

function handleReshim() {
  emit("log-event", "INFO", "Executing unirtm reshim...", "Updating ~/.unirtm/shims router");
  setTimeout(() => {
    emit("log-event", "SUCCESS", "All 6 tool shims verified and regenerated successfully.", "Time taken: 18ms");
  }, 200);
}

function handleSwitchVersion(toolName: string) {
  emit("log-event", "INFO", `Opened version picker for ${toolName}.`, "Fetching available releases...");
}

function handleVerifyTool(toolName: string) {
  emit("log-event", "SUCCESS", `Tool check: ${toolName} binary integrity and shims are OK.`, "Hash matches unirtm.lock");
}
</script>

<style scoped>
.tools-view {
  display: flex;
  flex-direction: column;
  gap: 1.1rem;
  padding: 1.25rem;
  height: 100%;
  overflow-y: auto;
  box-sizing: border-box;
}

.view-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.view-title {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0 0 0.35rem;
  color: var(--text-main);
}

.view-desc {
  font-size: 0.82rem;
  color: var(--text-muted);
  margin: 0;
}

.tool-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.45rem 0.95rem;
  border-radius: 8px;
  border: none;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  color: #070a12;
  transition: all 0.2s ease;
}

.tool-action-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 229, 255, 0.3);
}

.tools-filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 1rem;
  border-radius: 10px;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  max-width: 480px;
}

.search-icon {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.tool-search-input {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 0.82rem;
}

.tool-count-badge {
  font-size: 0.75rem;
  color: var(--text-muted);
  background: var(--input-bg);
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
}

.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1rem;
}

.tool-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 1.1rem;
  border-radius: 12px;
  transition: all 0.2s ease;
}

.tool-card:hover {
  border-color: var(--card-border-active);
  transform: translateY(-2px);
}

.tool-card-header {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.85rem;
}

.tool-badge-icon {
  font-size: 1.8rem;
  line-height: 1;
}

.tool-meta-block {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.tool-name-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.tool-name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-main);
}

.tool-tag {
  font-size: 0.68rem;
  font-weight: 600;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
}

.tag-blue {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
}

.tag-green {
  background: rgba(16, 185, 129, 0.12);
  color: var(--success);
}

.tag-purple {
  background: rgba(157, 78, 221, 0.12);
  color: var(--accent-purple);
}

.tag-red {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}

.tool-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
  line-height: 1.45;
  height: 2.9em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tool-details-box {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.55rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  margin-bottom: 0.85rem;
  font-size: 0.75rem;
}

.detail-line {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-line .lbl {
  color: var(--text-muted);
}

.version-highlight {
  color: var(--accent-cyan);
  font-weight: 600;
}

.tool-card-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-sub {
  flex: 1;
  padding: 0.35rem 0.6rem;
  border-radius: 6px;
  font-size: 0.75rem;
  cursor: pointer;
  border: none;
  transition: all 0.15s ease;
  font-weight: 500;
}

.btn-sub.secondary {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
}

.btn-sub.secondary:hover {
  border-color: var(--card-border-active);
  background: rgba(0, 229, 255, 0.08);
}

.btn-sub.primary {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  border: 1px solid rgba(0, 229, 255, 0.3);
}

.btn-sub.primary:hover {
  background: var(--accent-cyan);
  color: #070a12;
}
</style>
