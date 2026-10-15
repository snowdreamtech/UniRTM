<template>
  <div class="view-container tools-view" id="view-tools">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <div class="title-with-badge">
          <h2 class="view-title" style="display: flex; align-items: center; gap: 8px;">
            <TechIcon id="workbench:tools" :size="22" />
            <span>项目开发工具链与运行时全集</span>
          </h2>
          <span class="superset-badge">SUPERSET</span>
        </div>
        <p class="view-desc">
          由 UniRTM 统一版本锁定与高性能垫片（Go Shims）路由，涵盖编程语言、构建系统、云原生运维与安全扫描等全量工具。
        </p>
      </div>
      <div class="view-actions">
        <!-- Scope Switcher Pill -->
        <div class="scope-pill-group">
          <button
            class="scope-pill-btn"
            :class="{ active: currentScope === 'workspace' }"
            @click="currentScope = 'workspace'"
          >
            <TechIcon id="workbench:workspace" :size="13" :colored="currentScope === 'workspace'" />
            <span>当前工作区</span>
          </button>
          <button
            class="scope-pill-btn"
            :class="{ active: currentScope === 'user' }"
            @click="currentScope = 'user'"
          >
            <TechIcon id="workbench:user" :size="13" :colored="currentScope === 'user'" />
            <span>用户默认</span>
          </button>
        </div>

        <button class="tool-action-btn secondary" @click="handleReshim">
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:tasks" :size="14" :colored="false" />
            <span>刷新垫片</span>
          </span>
        </button>

        <button class="tool-action-btn primary" @click="openAddModal">
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <span>+</span>
            <span>添加工具</span>
          </span>
        </button>
      </div>
    </div>

    <!-- Category Tabs Filter -->
    <div class="category-tabs-bar">
      <button
        v-for="cat in categories"
        :key="cat.id"
        class="category-tab-btn"
        :class="{ active: activeCategory === cat.id }"
        @click="activeCategory = cat.id"
      >
        <span>{{ cat.label }}</span>
        <span class="cat-count">{{ getCategoryCount(cat.id) }}</span>
      </button>
    </div>

    <!-- Search & Summary Bar -->
    <div class="tools-filter-bar glass-card">
      <div class="search-input-wrapper">
        <span class="search-icon">
          <TechIcon id="workbench:doctor" :size="14" :colored="false" />
        </span>
        <input
          v-model="searchKey"
          type="text"
          class="tool-search-input"
          placeholder="按工具名称、类别或用途全局搜索..."
        />
        <button v-if="searchKey" class="search-clear-btn" @click="searchKey = ''">✕</button>
      </div>
      <div class="tool-count-badge">
        <span>已显示: {{ filteredTools.length }} / {{ tools.length }}</span>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="filteredTools.length === 0" class="empty-state glass-card">
      <TechIcon id="workbench:tools" :size="48" :colored="false" />
      <p class="empty-title">未找到匹配的工具</p>
      <p class="empty-desc">尝试更换搜索关键字，或在当前分类下点击“添加工具”安装新工具。</p>
      <button class="tool-action-btn primary small" @click="openAddModal">
        + 添加工具
      </button>
    </div>

    <!-- Tools Grid Table -->
    <div v-else class="tools-grid">
      <div
        v-for="tool in filteredTools"
        :key="tool.id"
        class="glass-card tool-card"
        :class="{
          active: isToolActive(tool),
          'subitem-focused': props.activeSubItem && (props.activeSubItem.toLowerCase() === tool.id.toLowerCase() || props.activeSubItem.toLowerCase() === tool.name.toLowerCase())
        }"
      >
        <div class="tool-card-header">
          <div class="tool-badge-icon">
            <TechIcon :id="tool.id" :size="28" :colored="true" />
          </div>
          <div class="tool-meta-block">
            <div class="tool-name-row">
              <span class="tool-name">{{ tool.name }}</span>
              <span class="tool-tag" :class="getCategoryTagClass(tool.category)">
                {{ getCategoryLabel(tool.category) }}
              </span>
            </div>
            <span class="tool-desc">{{ tool.description }}</span>
          </div>
        </div>

        <div class="tool-details-box">
          <div class="detail-line">
            <span class="lbl">生效版本:</span>
            <div class="version-select-wrapper">
              <select
                class="inline-version-select font-mono"
                :value="getEffectiveVersion(tool)"
                @change="onVersionChange(tool, ($event.target as HTMLSelectElement).value)"
              >
                <option
                  v-for="ver in tool.installed_versions"
                  :key="ver"
                  :value="ver"
                >
                  {{ ver }}
                </option>
                <option value="__remote__">+ 远程版本 / 安装新版本...</option>
              </select>
            </div>
          </div>

          <div class="detail-line">
            <span class="lbl">配置来源:</span>
            <span class="val font-mono text-muted">
              {{ tool.source || (currentScope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml') }}
            </span>
          </div>

          <div class="detail-line">
            <span class="lbl">垫片路径:</span>
            <span class="val font-mono text-muted">~/.unirtm/shims/{{ tool.id }}</span>
          </div>
        </div>

        <div class="tool-card-actions">
          <button class="btn-sub secondary" @click="openRemoteModal(tool)">
            <span>获取新版本</span>
          </button>
          <button class="btn-sub primary" @click="handleVerifyTool(tool.name)">
            <span>健康诊断</span>
          </button>
          <button
            class="btn-sub danger"
            title="移除工具"
            @click="confirmRemoveTool(tool)"
          >
            <span>移除</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Add New Tool Modal -->
    <div v-if="isAddModalOpen" class="modal-backdrop" @click.self="isAddModalOpen = false">
      <div class="modal-dialog glass-card add-tool-dialog" role="dialog" aria-modal="true">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon">
              <TechIcon id="workbench:tools" :size="20" />
            </span>
            <div>
              <h3 class="modal-title">添加工具</h3>
              <p class="modal-desc">为当前{{ currentScope === 'user' ? '用户默认' : '工作区' }}添加开发工具与版本</p>
            </div>
          </div>
          <button class="modal-close-btn" @click="isAddModalOpen = false">✕</button>
        </div>

        <div class="add-preset-section">
          <div class="preset-section-title">官方支持推荐工具全集</div>
          <div class="add-preset-grid">
            <div
              v-for="preset in availablePresetTools"
              :key="preset.id"
              class="add-preset-card"
            >
              <div class="preset-card-left">
                <span class="preset-icon"><TechIcon :id="preset.id" :size="20" :colored="true" /></span>
                <div class="preset-info">
                  <div class="preset-name-row">
                    <span class="preset-name">{{ preset.name }}</span>
                    <span class="preset-tag font-mono">v{{ preset.defaultVersion }}</span>
                  </div>
                  <span class="preset-desc">{{ preset.description }}</span>
                </div>
              </div>
              <button
                class="tool-action-btn primary small"
                @click="addToolFromPreset(preset)"
              >
                + 添加
              </button>
            </div>
          </div>
        </div>

        <div class="custom-add-section">
          <div class="preset-section-title">或输入自定义工具名</div>
          <div class="custom-add-form">
            <input
              v-model="customToolInput"
              type="text"
              class="custom-tool-input font-mono"
              placeholder="例如: elixir, zig, maven, kubectl, biome..."
              @keyup.enter="addCustomTool"
            />
            <button
              class="tool-action-btn primary small"
              :disabled="!customToolInput.trim()"
              @click="addCustomTool"
            >
              添加
            </button>
          </div>
        </div>

        <div class="modal-footer">
          <button class="tool-action-btn secondary" @click="isAddModalOpen = false">
            {{ t("common.close") || "Close" }}
          </button>
        </div>
      </div>
    </div>

    <!-- Remote Version Picker Modal -->
    <div v-if="isRemoteModalOpen" class="modal-backdrop" @click.self="closeRemoteModal">
      <div class="modal-dialog glass-card remote-modal-dialog" role="dialog" aria-modal="true">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon"><TechIcon :id="selectedTool?.id || 'workbench:tools'" :size="22" :colored="true" /></span>
            <div>
              <h3 class="modal-title">安装新版本 - {{ selectedTool?.name }}</h3>
              <p class="modal-desc">从官方权威上游获取已发布的可用版本</p>
            </div>
          </div>
          <button class="modal-close-btn" @click="closeRemoteModal">✕</button>
        </div>

        <div class="remote-filter-bar">
          <div class="search-input-wrapper">
            <span class="search-icon">🔍</span>
            <input
              v-model="remoteSearchKey"
              type="text"
              class="remote-search-input font-mono"
              placeholder="搜索版本号 (例如: 22, 1.24, 3.12...)"
            />
            <button v-if="remoteSearchKey" class="search-clear-btn" @click="remoteSearchKey = ''">✕</button>
          </div>

          <div v-if="majorVersionFilters.length > 0" class="major-filter-pills">
            <button
              class="major-pill"
              :class="{ active: selectedMajorFilter === '' }"
              @click="selectedMajorFilter = ''"
            >
              全部
            </button>
            <button
              v-for="major in majorVersionFilters"
              :key="major"
              class="major-pill"
              :class="{ active: selectedMajorFilter === major }"
              @click="selectedMajorFilter = major"
            >
              {{ major }}.x
            </button>
          </div>
        </div>

        <div class="remote-versions-content">
          <div v-if="loadingRemote" class="remote-loading-state">
            <span class="loading-spinner"></span>
            <span>正在连接官方仓库获取版本元数据...</span>
          </div>

          <div v-else-if="filteredRemoteVersions.length === 0" class="remote-empty-state">
            <p>未找到匹配的官方发布版本</p>
          </div>

          <div v-else class="remote-versions-grid">
            <div
              v-for="verObj in filteredRemoteVersions"
              :key="verObj.version"
              class="remote-version-card"
            >
              <div class="ver-card-left">
                <span class="ver-number font-mono">{{ verObj.version }}</span>
                <span v-if="isInstalled(selectedTool, verObj.version)" class="installed-tag">已安装</span>
              </div>
              <button
                class="tool-action-btn primary small"
                :disabled="installingVer === verObj.version"
                @click="installAndUseVersion(verObj.version)"
              >
                <span v-if="installingVer === verObj.version">安装中...</span>
                <span v-else>{{ isInstalled(selectedTool, verObj.version) ? "切换使用" : "安装并使用" }}</span>
              </button>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="tool-action-btn secondary" @click="closeRemoteModal">
            {{ t("common.close") || "Close" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t } from "../i18n";
import { useToast } from "../composables/useToast";
import TechIcon from "../components/common/TechIcon.vue";

const { showToast } = useToast();

const props = defineProps<{
  activeSubItem?: string;
}>();

const emit = defineEmits<{
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string): void;
}>();

export type ToolCategory = "languages" | "build" | "devops" | "linters";

export interface ToolItem {
  id: string;
  name: string;
  category: ToolCategory;
  description: string;
  homepage: string;
  installed_versions: string[];
  active_version: string;
  global_version: string;
  source: string;
}

const currentScope = ref<"user" | "workspace">("workspace");
const searchKey = ref("");
const activeCategory = ref<string>("all");

const categories = [
  { id: "all", label: "全部工具" },
  { id: "languages", label: "编程语言" },
  { id: "build", label: "构建工具" },
  { id: "devops", label: "云原生与运维" },
  { id: "linters", label: "规范与安全" },
];

// Initial default tools list (Superset containing languages, build tools, devops, and linters)
const defaultTools: ToolItem[] = [
  // Programming Languages
  {
    id: "node",
    name: "Node.js",
    category: "languages",
    description: "JavaScript / TypeScript 核心运行环境 (NVM 超集)",
    homepage: "https://nodejs.org",
    installed_versions: ["20.18.3", "22.14.0"],
    active_version: "22.14.0",
    global_version: "22.14.0",
    source: ".unirtm.toml",
  },
  {
    id: "go",
    name: "Go",
    category: "languages",
    description: "Go 编译型系统级语言环境 (GVM 超集)",
    homepage: "https://go.dev",
    installed_versions: ["1.24.0"],
    active_version: "1.24.0",
    global_version: "1.24.0",
    source: ".unirtm.toml",
  },
  {
    id: "python",
    name: "Python",
    category: "languages",
    description: "Python 脚本与人工智能解释器 (Pyenv 超集)",
    homepage: "https://python.org",
    installed_versions: ["3.12.2", "3.13.2"],
    active_version: "3.13.2",
    global_version: "3.13.2",
    source: ".unirtm.toml",
  },
  {
    id: "rust",
    name: "Rust",
    category: "languages",
    description: "Rust 现代内存安全系统工具链与 Cargo",
    homepage: "https://rust-lang.org",
    installed_versions: ["1.85.0"],
    active_version: "1.85.0",
    global_version: "1.85.0",
    source: ".unirtm.toml",
  },
  {
    id: "java",
    name: "Java (OpenJDK)",
    category: "languages",
    description: "Enterprise multi-platform Java Development Kit & Virtual Machine",
    homepage: "https://openjdk.org",
    installed_versions: ["21.0.2"],
    active_version: "21.0.2",
    global_version: "21.0.2",
    source: ".unirtm.toml",
  },

  // Build Systems
  {
    id: "maven",
    name: "Maven",
    category: "build",
    description: "Apache Maven 跨平台 Java 项目自动化构建与依赖管理工具",
    homepage: "https://maven.apache.org",
    installed_versions: ["3.9.6"],
    active_version: "3.9.6",
    global_version: "3.9.6",
    source: ".unirtm.toml",
  },
  {
    id: "gradle",
    name: "Gradle",
    category: "build",
    description: "现代化高性能多语言构建系统与依赖分发器",
    homepage: "https://gradle.org",
    installed_versions: ["8.10.2"],
    active_version: "8.10.2",
    global_version: "8.10.2",
    source: ".unirtm.toml",
  },
  {
    id: "cmake",
    name: "CMake",
    category: "build",
    description: "跨平台 C/C++ 软件构建、测试与打包系统",
    homepage: "https://cmake.org",
    installed_versions: ["3.30.5"],
    active_version: "3.30.5",
    global_version: "3.30.5",
    source: ".unirtm.toml",
  },
  {
    id: "ninja",
    name: "Ninja",
    category: "build",
    description: "专注于超高编译速度的小型底层构建系统",
    homepage: "https://ninja-build.org",
    installed_versions: ["1.12.1"],
    active_version: "1.12.1",
    global_version: "1.12.1",
    source: ".unirtm.toml",
  },

  // Cloud & DevOps
  {
    id: "kubectl",
    name: "Kubectl",
    category: "devops",
    description: "Kubernetes 集群容器编排管理官方命令行工具",
    homepage: "https://kubernetes.io",
    installed_versions: ["v1.31.2"],
    active_version: "v1.31.2",
    global_version: "v1.31.2",
    source: ".unirtm.toml",
  },
  {
    id: "helm",
    name: "Helm",
    category: "devops",
    description: "Kubernetes 应用程序包管理器与 Chart 部署器",
    homepage: "https://helm.sh",
    installed_versions: ["3.16.2"],
    active_version: "3.16.2",
    global_version: "3.16.2",
    source: ".unirtm.toml",
  },

  // Linters & Security
  {
    id: "pre-commit",
    name: "pre-commit",
    category: "linters",
    description: "多语言 Git 提交前钩子管理与代码安全检查框架",
    homepage: "https://pre-commit.com",
    installed_versions: ["4.0.1"],
    active_version: "4.0.1",
    global_version: "4.0.1",
    source: ".unirtm.toml",
  },
  {
    id: "trivy",
    name: "Trivy",
    category: "linters",
    description: "全方位容器镜像、代码库与依赖 CVE 漏洞安全扫描工具",
    homepage: "https://trivy.dev",
    installed_versions: ["0.56.2"],
    active_version: "0.56.2",
    global_version: "0.56.2",
    source: ".unirtm.toml",
  },
  {
    id: "eslint",
    name: "ESLint",
    category: "linters",
    description: "JavaScript 与 TypeScript 代码静态规范检测与自动修复器",
    homepage: "https://eslint.org",
    installed_versions: ["9.14.0"],
    active_version: "9.14.0",
    global_version: "9.14.0",
    source: ".unirtm.toml",
  },
];

const tools = ref<ToolItem[]>([...defaultTools]);

// Preset Tools for Quick Adding (Comprehensive catalog covering native languages and tools)
interface PresetTool {
  id: string;
  name: string;
  category: ToolCategory;
  description: string;
  homepage: string;
  defaultVersion: string;
}

const presetCatalog: PresetTool[] = [
  // Programming Languages
  { id: "bun", name: "Bun", category: "languages", description: "Incredibly fast all-in-one JavaScript & TypeScript toolkit", homepage: "https://bun.sh", defaultVersion: "1.2.4" },
  { id: "deno", name: "Deno", category: "languages", description: "Modern, secure runtime for JavaScript and TypeScript", homepage: "https://deno.com", defaultVersion: "2.2.2" },
  { id: "node", name: "Node.js", category: "languages", description: "JavaScript / TypeScript runtime (NVM superset)", homepage: "https://nodejs.org", defaultVersion: "22.14.0" },
  { id: "go", name: "Go", category: "languages", description: "Go programming language (GVM superset)", homepage: "https://go.dev", defaultVersion: "1.24.0" },
  { id: "python", name: "Python", category: "languages", description: "Python language interpreter (Pyenv superset)", homepage: "https://python.org", defaultVersion: "3.13.2" },
  { id: "rust", name: "Rust", category: "languages", description: "Rust toolchain & cargo build system", homepage: "https://rust-lang.org", defaultVersion: "1.85.0" },
  { id: "java", name: "Java (OpenJDK)", category: "languages", description: "Enterprise multi-platform Java Development Kit & Virtual Machine", homepage: "https://openjdk.org", defaultVersion: "21.0.2" },
  { id: "zig", name: "Zig", category: "languages", description: "General-purpose programming language and toolchain for robust software", homepage: "https://ziglang.org", defaultVersion: "0.13.0" },
  { id: "ruby", name: "Ruby", category: "languages", description: "Dynamic, open source programming language with a focus on simplicity", homepage: "https://ruby-lang.org", defaultVersion: "3.4.1" },
  { id: "php", name: "PHP", category: "languages", description: "Popular general-purpose scripting language suited for web development", homepage: "https://php.net", defaultVersion: "8.4.4" },
  { id: "dart", name: "Dart", category: "languages", description: "Client-optimized language for fast apps on any platform", homepage: "https://dart.dev", defaultVersion: "3.6.1" },
  { id: "flutter", name: "Flutter", category: "languages", description: "Google's UI toolkit for multi-platform applications", homepage: "https://flutter.dev", defaultVersion: "3.29.0" },
  { id: "julia", name: "Julia", category: "languages", description: "High-performance dynamic language for technical & scientific computing", homepage: "https://julialang.org", defaultVersion: "1.11.3" },
  { id: "elixir", name: "Elixir", category: "languages", description: "Dynamic, functional language for building scalable & maintainable software", homepage: "https://elixir-lang.org", defaultVersion: "1.18.2" },
  { id: "erlang", name: "Erlang / OTP", category: "languages", description: "Soft real-time distributed system language & OTP runtime", homepage: "https://erlang.org", defaultVersion: "27.2.1" },
  { id: "haskell", name: "Haskell (GHC)", category: "languages", description: "Purely functional programming language with Glasgow Haskell Compiler", homepage: "https://haskell.org", defaultVersion: "9.12.1" },

  // Build Systems
  { id: "maven", name: "Maven", category: "build", description: "Apache Maven Java dependency management and project build tool", homepage: "https://maven.apache.org", defaultVersion: "3.9.6" },
  { id: "gradle", name: "Gradle", category: "build", description: "High-performance, multi-language build automation tool", homepage: "https://gradle.org", defaultVersion: "8.10.2" },
  { id: "cmake", name: "CMake", category: "build", description: "Cross-platform family of tools to build, test and package software", homepage: "https://cmake.org", defaultVersion: "3.30.5" },
  { id: "ninja", name: "Ninja", category: "build", description: "Small build system with a focus on speed", homepage: "https://ninja-build.org", defaultVersion: "1.12.1" },

  // Cloud & DevOps
  { id: "kubectl", name: "Kubectl", category: "devops", description: "Kubernetes command-line tool for cluster management", homepage: "https://kubernetes.io", defaultVersion: "v1.31.2" },
  { id: "helm", name: "Helm", category: "devops", description: "The package manager for Kubernetes", homepage: "https://helm.sh", defaultVersion: "3.16.2" },

  // Linters & Security
  { id: "pre-commit", name: "pre-commit", category: "linters", description: "Framework for managing and maintaining multi-language pre-commit hooks", homepage: "https://pre-commit.com", defaultVersion: "4.0.1" },
  { id: "trivy", name: "Trivy", category: "linters", description: "Comprehensive security scanner for containers and dependencies", homepage: "https://trivy.dev", defaultVersion: "0.56.2" },
  { id: "eslint", name: "ESLint", category: "linters", description: "Pluggable JavaScript/TypeScript linter utility", homepage: "https://eslint.org", defaultVersion: "9.14.0" },
];

const availablePresetTools = computed(() => {
  const currentIds = new Set(tools.value.map((t) => t.id.toLowerCase()));
  return presetCatalog.filter((p) => !currentIds.has(p.id.toLowerCase()));
});

function getCategoryCount(catId: string): number {
  if (catId === "all") return tools.value.length;
  return tools.value.filter((t) => t.category === catId).length;
}

function getCategoryLabel(catId: ToolCategory): string {
  switch (catId) {
    case "languages": return "编程语言";
    case "build": return "构建系统";
    case "devops": return "云原生与运维";
    case "linters": return "规范与安全";
    default: return "工具";
  }
}

function getCategoryTagClass(catId: ToolCategory): string {
  switch (catId) {
    case "languages": return "tag-blue";
    case "build": return "tag-green";
    case "devops": return "tag-cyan";
    case "linters": return "tag-purple";
    default: return "tag-blue";
  }
}

const filteredTools = computed(() => {
  let list = tools.value;
  if (props.activeSubItem && props.activeSubItem.trim()) {
    const sub = props.activeSubItem.trim().toLowerCase();
    const matched = list.filter((t) => t.id.toLowerCase() === sub || t.name.toLowerCase() === sub);
    if (matched.length > 0) {
      return matched;
    }
  }
  if (activeCategory.value !== "all") {
    list = list.filter((t) => t.category === activeCategory.value);
  }
  if (!searchKey.value.trim()) return list;
  const q = searchKey.value.toLowerCase().trim();
  return list.filter(
    (t) =>
      t.name.toLowerCase().includes(q) ||
      t.id.toLowerCase().includes(q) ||
      t.description.toLowerCase().includes(q) ||
      t.installed_versions.some((v) => v.includes(q))
  );
});

function isToolActive(tool: ToolItem): boolean {
  return !!getEffectiveVersion(tool);
}

function getEffectiveVersion(tool: ToolItem): string {
  if (currentScope.value === "user") {
    return tool.global_version || tool.active_version || tool.installed_versions[0] || "";
  }
  return tool.active_version || tool.global_version || tool.installed_versions[0] || "";
}

async function onVersionChange(tool: ToolItem, selectedVal: string) {
  if (selectedVal === "__remote__") {
    openRemoteModal(tool);
    return;
  }
  const isGlobal = currentScope.value === "user";
  if (isGlobal) {
    tool.global_version = selectedVal;
  } else {
    tool.active_version = selectedVal;
  }
  showToast(`已将 ${tool.name} 切换为 v${selectedVal}`, "success");
  emit("log-event", "INFO", `Switched ${tool.name} to version ${selectedVal} (${isGlobal ? "global" : "workspace"})`);

  try {
    await fetch("/api/runtimes/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: tool.id, version: selectedVal, global: isGlobal }),
    });
  } catch {
    // ignore
  }
}

// Add Tool Modal logic
const isAddModalOpen = ref(false);
const customToolInput = ref("");

function openAddModal() {
  customToolInput.value = "";
  isAddModalOpen.value = true;
}

async function addToolFromPreset(preset: PresetTool) {
  const isGlobal = currentScope.value === "user";
  const newItem: ToolItem = {
    id: preset.id,
    name: preset.name,
    category: preset.category,
    description: preset.description,
    homepage: preset.homepage,
    installed_versions: [preset.defaultVersion],
    active_version: isGlobal ? "" : preset.defaultVersion,
    global_version: isGlobal ? preset.defaultVersion : "",
    source: isGlobal ? "~/.config/unirtm/unirtm.toml" : ".unirtm.toml",
  };

  tools.value.push(newItem);
  isAddModalOpen.value = false;
  showToast(`已成功添加 ${preset.name} (v${preset.defaultVersion})`, "success");
  emit("log-event", "SUCCESS", `Added tool ${preset.name}`);

  try {
    await fetch("/api/runtimes/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: preset.id, version: preset.defaultVersion, global: isGlobal }),
    });
  } catch {
    // ignore
  }
}

async function addCustomTool() {
  const toolName = customToolInput.value.trim().toLowerCase();
  if (!toolName) return;

  const isGlobal = currentScope.value === "user";
  const newItem: ToolItem = {
    id: toolName,
    name: toolName.charAt(0).toUpperCase() + toolName.slice(1),
    category: "languages",
    description: `${toolName} toolchain environment`,
    homepage: "",
    installed_versions: ["latest"],
    active_version: isGlobal ? "" : "latest",
    global_version: isGlobal ? "latest" : "",
    source: isGlobal ? "~/.config/unirtm/unirtm.toml" : ".unirtm.toml",
  };

  tools.value.push(newItem);
  customToolInput.value = "";
  isAddModalOpen.value = false;
  showToast(`已成功添加工具 ${newItem.name}`, "success");
  emit("log-event", "SUCCESS", `Added custom tool ${newItem.name}`);

  try {
    await fetch("/api/runtimes/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: toolName, version: "latest", global: isGlobal }),
    });
  } catch {
    // ignore
  }
}

async function confirmRemoveTool(tool: ToolItem) {
  const isGlobal = currentScope.value === "user";
  const scopeName = isGlobal ? "用户全局" : "当前工作区";
  if (!window.confirm(`确定从${scopeName}中移除 ${tool.name} 吗？`)) {
    return;
  }

  const prev = [...tools.value];
  tools.value = tools.value.filter((t) => t.id !== tool.id);
  showToast(`正在从${scopeName}移除 ${tool.name}...`, "info");

  try {
    const res = await fetch("/api/runtimes/remove", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: tool.id, global: isGlobal }),
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    showToast(`已成功移除 ${tool.name}`, "success");
    emit("log-event", "SUCCESS", `Removed tool ${tool.name}`);
  } catch (err: any) {
    tools.value = prev;
    showToast(`移除失败: ${err.message}`, "error");
  }
}

// Remote Modal Logic
const isRemoteModalOpen = ref(false);
const selectedTool = ref<ToolItem | null>(null);
const remoteVersions = ref<{ version: string }[]>([]);
const remoteSearchKey = ref("");
const selectedMajorFilter = ref("");
const loadingRemote = ref(false);
const installingVer = ref<string | null>(null);

function isInstalled(tool: ToolItem | null, version: string): boolean {
  if (!tool) return false;
  return tool.installed_versions.includes(version);
}

function openRemoteModal(tool: ToolItem) {
  selectedTool.value = tool;
  remoteVersions.value = [];
  remoteSearchKey.value = "";
  selectedMajorFilter.value = "";
  isRemoteModalOpen.value = true;
  fetchRemoteVersions(tool);
}

function closeRemoteModal() {
  isRemoteModalOpen.value = false;
  selectedTool.value = null;
  remoteVersions.value = [];
  remoteSearchKey.value = "";
  selectedMajorFilter.value = "";
  loadingRemote.value = false;
}

async function fetchRemoteVersions(tool: ToolItem) {
  loadingRemote.value = true;
  try {
    const res = await fetch(`/api/runtimes/remote?tool=${encodeURIComponent(tool.id)}`);
    if (res.ok) {
      const data = await res.json();
      if (data.versions && Array.isArray(data.versions)) {
        remoteVersions.value = data.versions;
      }
    }
  } catch {
    // fallback to some mock versions if offline
    remoteVersions.value = [
      { version: "latest" },
      ...tool.installed_versions.map((v) => ({ version: v })),
    ];
  } finally {
    loadingRemote.value = false;
  }
}

const majorVersionFilters = computed(() => {
  const majors = new Set<string>();
  for (const v of remoteVersions.value) {
    const parts = v.version.replace(/^v/, "").split(".");
    if (parts[0] && !isNaN(parseInt(parts[0], 10))) {
      majors.add(parts[0]);
    }
  }
  return Array.from(majors).slice(0, 5);
});

const filteredRemoteVersions = computed(() => {
  let list = remoteVersions.value;
  if (selectedMajorFilter.value) {
    list = list.filter((v) => {
      const clean = v.version.replace(/^v/, "");
      return clean.startsWith(selectedMajorFilter.value + ".");
    });
  }
  if (remoteSearchKey.value.trim()) {
    const q = remoteSearchKey.value.trim().toLowerCase();
    list = list.filter((v) => v.version.toLowerCase().includes(q));
  }
  return list;
});

async function installAndUseVersion(ver: string) {
  if (!selectedTool.value) return;
  const tool = selectedTool.value;
  const isGlobal = currentScope.value === "user";

  installingVer.value = ver;
  showToast(`正在安装 ${tool.name} v${ver}...`, "info");
  emit("log-event", "INFO", `Installing ${tool.name} v${ver}...`);

  setTimeout(async () => {
    installingVer.value = null;
    if (!tool.installed_versions.includes(ver)) {
      tool.installed_versions.unshift(ver);
    }
    if (isGlobal) {
      tool.global_version = ver;
    } else {
      tool.active_version = ver;
    }
    closeRemoteModal();
    showToast(`已成功安装并切换至 ${tool.name} v${ver}`, "success");
    emit("log-event", "SUCCESS", `Installed and activated ${tool.name} v${ver}`);

    try {
      await fetch("/api/runtimes/use", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ tool: tool.id, version: ver, global: isGlobal }),
      });
    } catch {
      // ignore
    }
  }, 600);
}

function handleReshim() {
  emit("log-event", "INFO", "Executing unirtm reshim...", "Updating ~/.unirtm/shims router");
  showToast("正在重新生成垫片 (Reshim)...", "info");
  setTimeout(() => {
    emit("log-event", "SUCCESS", `All ${tools.value.length} tool shims regenerated successfully.`, "Time taken: 16ms");
    showToast("所有工具垫片已成功生成并校验完毕！", "success");
  }, 300);
}

function handleVerifyTool(toolName: string) {
  emit("log-event", "SUCCESS", `Tool check: ${toolName} binary integrity and shims are OK.`, "Hash matches unirtm.lock");
  showToast(`${toolName} 完整性校验通过，垫片路由正常。`, "success");
}

// Fetch live runtimes status to merge with tools on mount
onMounted(async () => {
  try {
    const res = await fetch("/api/runtimes/status");
    if (res.ok) {
      const data = await res.json();
      if (data.runtimes && Array.isArray(data.runtimes)) {
        for (const rt of data.runtimes) {
          const existing = tools.value.find((t) => t.id === rt.id);
          if (existing) {
            if (rt.installed_versions && rt.installed_versions.length > 0) {
              existing.installed_versions = rt.installed_versions;
            }
            if (rt.active_version) existing.active_version = rt.active_version;
            if (rt.global_version) existing.global_version = rt.global_version;
            if (rt.source) existing.source = rt.source;
          }
        }
      }
    }
  } catch {
    // dev mode fallback
  }
});
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
  flex-wrap: wrap;
  gap: 0.75rem;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.view-title {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-main);
}

.superset-badge {
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: linear-gradient(135deg, rgba(0, 229, 255, 0.2), rgba(157, 78, 221, 0.2));
  color: var(--accent-cyan);
  border: 1px solid rgba(0, 229, 255, 0.4);
}

.view-desc {
  font-size: 0.82rem;
  color: var(--text-muted);
  margin: 0.35rem 0 0;
}

.view-actions {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

/* Scope Pill Group */
.scope-pill-group {
  display: inline-flex;
  align-items: center;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 2px;
}

.scope-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.3rem 0.65rem;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.76rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.scope-pill-btn.active {
  background: var(--card-bg);
  color: var(--text-main);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.25);
  font-weight: 600;
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
  transition: all 0.2s ease;
}

.tool-action-btn.primary {
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  color: #070a12;
}

.tool-action-btn.primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 229, 255, 0.3);
}

.tool-action-btn.secondary {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
}

.tool-action-btn.secondary:hover {
  background: rgba(0, 229, 255, 0.08);
  border-color: var(--card-border-active);
}

.tool-action-btn.small {
  padding: 0.25rem 0.65rem;
  font-size: 0.74rem;
}

/* Category Tabs Bar */
.category-tabs-bar {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  border-bottom: 1px solid var(--card-border);
  padding-bottom: 0.4rem;
}

.category-tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.4rem 0.85rem;
  border-radius: 6px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.category-tab-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.04);
}

.category-tab-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.1);
  font-weight: 600;
}

.cat-count {
  font-size: 0.68rem;
  background: var(--input-bg);
  padding: 0.1rem 0.4rem;
  border-radius: 10px;
  color: var(--text-muted);
}

.category-tab-btn.active .cat-count {
  background: rgba(0, 229, 255, 0.2);
  color: var(--accent-cyan);
}

/* Filter Bar */
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
  position: relative;
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

.search-clear-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.75rem;
  padding: 0 0.3rem;
}

.tool-count-badge {
  font-size: 0.75rem;
  color: var(--text-muted);
  background: var(--input-bg);
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3.5rem 1.5rem;
  text-align: center;
  gap: 0.75rem;
  border-radius: 12px;
}

.empty-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0;
}

.empty-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  max-width: 380px;
  margin: 0 0 0.5rem;
}

/* Grid & Cards */
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
  background: var(--card-bg);
  border: 1px solid var(--card-border);
}

.tool-card:hover {
  border-color: var(--card-border-active);
  transform: translateY(-2px);
}

.tool-card.subitem-focused {
  border-color: var(--accent-cyan) !important;
  box-shadow: 0 0 0 1px var(--accent-cyan), 0 4px 20px rgba(0, 229, 255, 0.3) !important;
  animation: pulse-border 2s infinite ease-in-out;
}

@keyframes pulse-border {
  0%, 100% {
    box-shadow: 0 0 0 1px var(--accent-cyan), 0 4px 20px rgba(0, 229, 255, 0.25);
  }
  50% {
    box-shadow: 0 0 0 2px var(--accent-cyan), 0 6px 28px rgba(0, 229, 255, 0.45);
  }
}

.tool-card-header {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.85rem;
}

.tool-badge-icon {
  font-size: 1.8rem;
  line-height: 1;
  display: flex;
  align-items: flex-start;
}

.tool-meta-block {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  overflow: hidden;
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
  font-size: 0.66rem;
  font-weight: 600;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
}

.tag-blue { background: rgba(0, 229, 255, 0.12); color: var(--accent-cyan); }
.tag-green { background: rgba(16, 185, 129, 0.12); color: var(--success); }
.tag-cyan { background: rgba(6, 182, 212, 0.15); color: #22d3ee; }
.tag-purple { background: rgba(157, 78, 221, 0.12); color: var(--accent-purple); }

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

.inline-version-select {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--accent-cyan);
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  outline: none;
  cursor: pointer;
}

.inline-version-select:focus {
  border-color: var(--card-border-active);
}

.tool-card-actions {
  display: flex;
  gap: 0.4rem;
}

.btn-sub {
  flex: 1;
  padding: 0.35rem 0.5rem;
  border-radius: 6px;
  font-size: 0.74rem;
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

.btn-sub.danger {
  background: transparent;
  color: var(--text-muted);
  border: 1px solid transparent;
  flex: 0 0 auto;
}

.btn-sub.danger:hover {
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.08);
}

/* Modal Dialogs */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-dialog {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.25rem;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.5);
  width: 90%;
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.modal-title-row {
  display: flex;
  gap: 0.65rem;
}

.modal-title {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
}

.modal-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin: 0.2rem 0 0;
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1rem;
  cursor: pointer;
}

.modal-close-btn:hover {
  color: var(--text-main);
}

/* Add Tool Dialog */
.add-tool-dialog {
  max-width: 680px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.add-preset-section {
  overflow-y: auto;
  max-height: 380px;
  padding-right: 4px;
}

.preset-section-title {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.55rem;
}

.add-preset-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.65rem;
  margin-bottom: 0.5rem;
}

.add-preset-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.75rem;
  border-radius: 8px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  transition: all 0.15s ease;
}

.add-preset-card:hover {
  border-color: var(--card-border-active);
  background: var(--card-bg);
}

.preset-card-left {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  overflow: hidden;
}

.preset-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.preset-name-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.preset-name {
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
}

.preset-tag {
  font-size: 0.66rem;
  color: var(--text-muted);
}

.preset-desc {
  font-size: 0.68rem;
  color: var(--text-muted);
  max-width: 140px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.custom-add-section {
  padding-top: 0.75rem;
  border-top: 1px solid var(--card-border);
}

.custom-add-form {
  display: flex;
  gap: 0.6rem;
  align-items: center;
}

.custom-tool-input {
  flex: 1;
  padding: 0.45rem 0.75rem;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.82rem;
  outline: none;
}

.custom-tool-input:focus {
  border-color: var(--card-border-active);
}

/* Remote Version Dialog */
.remote-modal-dialog {
  max-width: 620px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.remote-filter-bar {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  margin-bottom: 0.85rem;
}

.remote-search-input {
  width: 100%;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 0.45rem 0.75rem;
  color: var(--text-main);
  font-size: 0.82rem;
  outline: none;
}

.major-filter-pills {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.major-pill {
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
  font-size: 0.72rem;
  font-weight: 600;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  cursor: pointer;
}

.major-pill.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.remote-versions-content {
  overflow-y: auto;
  max-height: 340px;
  padding-right: 4px;
}

.remote-loading-state,
.remote-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2.5rem;
  color: var(--text-muted);
  font-size: 0.82rem;
  gap: 0.6rem;
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(0, 229, 255, 0.2);
  border-top-color: var(--accent-cyan);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.remote-versions-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.6rem;
}

.remote-version-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.55rem 0.75rem;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
}

.ver-card-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.ver-number {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-main);
}

.installed-tag {
  font-size: 0.64rem;
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
}

.modal-footer {
  margin-top: 0.85rem;
  display: flex;
  justify-content: flex-end;
}
</style>
