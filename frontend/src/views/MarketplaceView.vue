<template>
  <div class="view-container marketplace-view" id="view-marketplace">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title">🧩 扩展与插件市场 (Extensions Marketplace)</h2>
        <p class="view-desc">
          让 UniRTM 内核保持极度轻量与高纯度，把 AI 智能体、专有语言支持与深度安全能力通过市场按需扩展。
        </p>
      </div>
    </div>

    <!-- Category Chips and Search Filter Bar -->
    <div class="market-filter-bar glass-card">
      <div class="search-input-wrapper">
        <span class="search-icon">🔍</span>
        <input
          v-model="searchQuery"
          type="text"
          class="market-search-input"
          placeholder="搜索扩展、插件、工具包或 AI 智能体 (如: agent, rust, security)..."
        />
      </div>

      <div class="category-chips">
        <button
          v-for="cat in categories"
          :key="cat.id"
          class="chip-btn"
          :class="{ active: activeCategory === cat.id }"
          @click="activeCategory = cat.id"
        >
          {{ cat.label }}
        </button>
      </div>
    </div>

    <!-- Extensions Card Grid -->
    <div class="extensions-grid">
      <div
        v-for="ext in filteredExtensions"
        :key="ext.id"
        class="glass-card extension-card"
      >
        <div class="ext-top-row">
          <div class="ext-icon-box">{{ ext.icon }}</div>
          <div class="ext-meta">
            <div class="ext-title-row">
              <span class="ext-name">{{ ext.name }}</span>
              <span class="ext-version font-mono">v{{ ext.version }}</span>
            </div>
            <span class="ext-author">{{ ext.author }}</span>
          </div>
        </div>

        <p class="ext-desc">{{ ext.description }}</p>

        <div class="ext-tags-row">
          <span v-for="tag in ext.tags" :key="tag" class="ext-tag">{{ tag }}</span>
        </div>

        <div class="ext-footer-actions">
          <div class="ext-stats">
            <span class="stat-star">⭐ {{ ext.stars }}</span>
            <span class="stat-downloads">📥 {{ ext.downloads }}</span>
          </div>

          <button
            class="ext-install-btn"
            :class="{ installed: ext.installed }"
            @click="toggleInstall(ext)"
          >
            <span>{{ ext.installed ? "已安装 (Installed)" : "获取并安装" }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";

const emit = defineEmits<{
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string): void;
}>();

const searchQuery = ref("");
const activeCategory = ref("all");

const categories = [
  { id: "all", label: "全部 (All)" },
  { id: "ai", label: "🤖 AI 智能体" },
  { id: "security", label: "🛡️ 安全审计" },
  { id: "runtime", label: "🛠️ 运行时 Provider" },
  { id: "pack", label: "📦 任务套件" },
];

interface ExtensionItem {
  id: string;
  name: string;
  icon: string;
  version: string;
  author: string;
  description: string;
  category: string;
  tags: string[];
  stars: string;
  downloads: string;
  installed: boolean;
}

const extensions = ref<ExtensionItem[]>([
  {
    id: "uniagent",
    name: "UniAgent",
    icon: "🤖",
    version: "0.1.0",
    author: "SnowdreamTech Official",
    description: "专注于项目跑起来、开发起来、编译起来过程中的智能故障诊断与排错助手。独立进程，绝不膨胀 UniRTM 内核。",
    category: "ai",
    tags: ["AI Agent", "Troubleshooter", "MCP Server"],
    stars: "1.2k",
    downloads: "28k",
    installed: true,
  },
  {
    id: "security-suite",
    name: "Security Suite Pack",
    icon: "🛡️",
    version: "1.2.0",
    author: "SnowdreamTech Official",
    description: "一键集成 Trivy, Syft, Gitleaks 深度漏洞检查与软件物料清单 (SBOM) 生成，守护供应链安全。",
    category: "security",
    tags: ["Trivy", "CVE Scan", "SBOM", "Pre-commit"],
    stars: "890",
    downloads: "15k",
    installed: true,
  },
  {
    id: "rust-provider",
    name: "Rust & Cargo Provider",
    icon: "🦀",
    version: "1.82.0",
    author: "Community",
    description: "为 UniRTM 提供原生 Rust / Cargo 极速下载、环境变量隔离与跨平台编译垫片支持。",
    category: "runtime",
    tags: ["Rust", "Cargo", "Tool Provider"],
    stars: "640",
    downloads: "9.2k",
    installed: false,
  },
  {
    id: "datascience-pack",
    name: "Python DataScience Pack",
    icon: "🔬",
    version: "2.0.0",
    author: "Community",
    description: "预先配置 uv, ruff, ipython, jupyter 科学计算全套轻量级任务流与环境配置。",
    category: "pack",
    tags: ["Python", "uv", "Ruff", "Jupyter"],
    stars: "520",
    downloads: "7.1k",
    installed: false,
  },
]);

const filteredExtensions = computed(() => {
  return extensions.value.filter((ext) => {
    const matchCat = activeCategory.value === "all" || ext.category === activeCategory.value;
    const matchQuery =
      !searchQuery.value ||
      ext.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      ext.description.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      ext.tags.some((t) => t.toLowerCase().includes(searchQuery.value.toLowerCase()));
    return matchCat && matchQuery;
  });
});

function toggleInstall(ext: ExtensionItem) {
  if (ext.installed) {
    ext.installed = false;
    emit("log-event", "INFO", `Uninstalled extension: ${ext.name}`, "Removed from ~/.local/share/unirtm/extensions");
  } else {
    ext.installed = true;
    emit("log-event", "SUCCESS", `Installed extension: ${ext.name} v${ext.version}`, "Mounted successfully to UniRTM");
  }
}
</script>

<style scoped>
.marketplace-view {
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

.market-filter-bar {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  border-radius: 12px;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.4rem 0.75rem;
}

.search-icon {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.market-search-input {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 0.82rem;
}

.category-chips {
  display: flex;
  gap: 0.45rem;
  flex-wrap: wrap;
}

.chip-btn {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.chip-btn:hover {
  color: var(--text-main);
  border-color: var(--card-border-active);
}

.chip-btn.active {
  background: rgba(0, 229, 255, 0.14);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  font-weight: 600;
}

.extensions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 1rem;
}

.extension-card {
  display: flex;
  flex-direction: column;
  padding: 1.15rem;
  border-radius: 12px;
  transition: all 0.2s ease;
}

.extension-card:hover {
  border-color: var(--card-border-active);
  transform: translateY(-2px);
}

.ext-top-row {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.65rem;
}

.ext-icon-box {
  font-size: 2rem;
  line-height: 1;
}

.ext-meta {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.ext-title-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.ext-name {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-main);
}

.ext-version {
  font-size: 0.72rem;
  color: var(--text-muted);
}

.ext-author {
  font-size: 0.72rem;
  color: var(--text-muted);
}

.ext-desc {
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.45;
  margin: 0 0 0.85rem;
  flex: 1;
}

.ext-tags-row {
  display: flex;
  gap: 0.35rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.ext-tag {
  font-size: 0.68rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--accent-cyan);
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
}

.ext-footer-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid var(--card-border);
  padding-top: 0.75rem;
}

.ext-stats {
  display: flex;
  gap: 0.65rem;
  font-size: 0.72rem;
  color: var(--text-muted);
}

.ext-install-btn {
  padding: 0.35rem 0.85rem;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  border: none;
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  color: #070a12;
  transition: all 0.2s ease;
}

.ext-install-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 229, 255, 0.3);
}

.ext-install-btn.installed {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--success);
}
</style>
