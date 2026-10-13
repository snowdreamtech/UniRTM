<template>
  <header class="titlebar" id="app-titlebar">
    <!-- Left: Brand & Workspace info -->
    <div class="titlebar-left">
      <div class="brand-badge">
        <img src="/logo.png" alt="UniRTM" class="titlebar-logo" />
        <span class="brand-title">UniRTM</span>
      </div>
      <div class="workspace-pill" :title="'Current Workspace: /UniRTM'">
        <span class="folder-icon">📁</span>
        <span class="workspace-name">UniRTM</span>
        <span class="branch-tag">dev</span>
      </div>
    </div>

    <!-- Center: Global Command Palette Search Bar -->
    <div class="titlebar-center">
      <div class="command-search-box" @click="$emit('open-palette')">
        <span class="search-icon">🔍</span>
        <span class="search-placeholder">
          {{ searchPlaceholderText }}
        </span>
        <kbd class="shortcut-kbd">Ctrl+K</kbd>
      </div>
    </div>

    <!-- Right: Quick actions (Language, Theme, Console toggle, Settings, About) -->
    <div class="titlebar-right">
      <!-- Quick Language Switcher Dropdown -->
      <div class="lang-selector-header" ref="langDropdownRef">
        <button
          class="titlebar-pill-btn"
          id="btn-lang-switcher"
          :title="t('settings.language') || 'Language'"
          @click.stop="isLangMenuOpen = !isLangMenuOpen"
        >
          <span class="lang-icon">🌐</span>
          <span class="lang-label">{{ currentLangLabel }}</span>
          <span class="dropdown-caret">▾</span>
        </button>

        <transition name="dropdown-fade">
          <div v-if="isLangMenuOpen" class="lang-dropdown-menu" @click.stop>
            <button
              v-for="opt in langOptions"
              :key="opt.value"
              class="lang-option"
              :class="{ active: currentLang === opt.value }"
              @click="selectLanguage(opt.value)"
            >
              <span class="opt-text">{{ opt.label }}</span>
              <span v-if="currentLang === opt.value" class="opt-check">✓</span>
            </button>
          </div>
        </transition>
      </div>

      <!-- Quick Theme Switcher Button -->
      <button
        class="titlebar-icon-btn"
        id="btn-theme-toggle"
        :title="isLight ? (t('theme.toggleDark') || 'Switch to Dark') : (t('theme.toggleLight') || 'Switch to Light')"
        @click="$emit('toggle-theme')"
      >
        <span>{{ isLight ? "🌙" : "☀️" }}</span>
      </button>

      <!-- Toggle Console Drawer Button -->
      <button
        class="titlebar-icon-btn"
        id="btn-console-drawer"
        :class="{ active: isDrawerOpen }"
        :title="'Toggle Console Drawer (Ctrl+` )'"
        @click="$emit('toggle-drawer')"
      >
        <span>📜</span>
      </button>

      <!-- Settings Button -->
      <button
        class="titlebar-icon-btn"
        id="btn-open-settings"
        :title="t('settings.title') || 'Settings'"
        @click="$emit('open-settings')"
      >
        <span>⚙️</span>
      </button>

      <!-- About Button -->
      <button
        class="titlebar-icon-btn"
        id="btn-open-about"
        :title="t('about.title') || 'About'"
        @click="$emit('open-about')"
      >
        <span>ℹ️</span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { SUPPORTED_LANGUAGES, t } from "../../i18n";

const props = defineProps<{
  currentLang: string;
  currentTheme: string;
  isDrawerOpen: boolean;
}>();

const emit = defineEmits<{
  (e: "toggle-theme"): void;
  (e: "open-settings"): void;
  (e: "open-about"): void;
  (e: "toggle-drawer"): void;
  (e: "open-palette"): void;
  (e: "select-lang", lang: string): void;
}>();

const isLangMenuOpen = ref(false);
const langDropdownRef = ref<HTMLElement | null>(null);

const isLight = computed(() => {
  if (props.currentTheme === "light") return true;
  if (props.currentTheme === "dark") return false;
  return typeof document !== "undefined" && document.documentElement.getAttribute("data-theme") === "light";
});

const langOptions = computed(() => [
  { value: "auto", label: "🌐 " + (t("common.autoDetect") || "Auto Detect") },
  ...SUPPORTED_LANGUAGES.map((item) => ({
    value: item.code,
    label: item.nativeName,
  })),
]);

const currentLangLabel = computed(() => {
  if (props.currentLang === "auto") {
    return t("common.langAuto") || "Auto";
  }
  const opt = langOptions.value.find((o) => o.value === props.currentLang);
  return opt ? opt.label : t("common.lang") || "Language";
});

const searchPlaceholderText = computed(() => {
  return props.currentLang.startsWith("zh")
    ? "搜索工具、任务、扩展与设置... (Ctrl+K)"
    : "Search tools, tasks, extensions & settings... (Ctrl+K)";
});

function selectLanguage(langVal: string) {
  emit("select-lang", langVal);
  isLangMenuOpen.value = false;
}

function handleGlobalClick(event: MouseEvent) {
  if (langDropdownRef.value && !langDropdownRef.value.contains(event.target as Node)) {
    isLangMenuOpen.value = false;
  }
}

onMounted(() => {
  window.addEventListener("click", handleGlobalClick);
});

onUnmounted(() => {
  window.removeEventListener("click", handleGlobalClick);
});
</script>

<style scoped>
.titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 42px;
  min-height: 42px;
  background: var(--card-bg);
  border-bottom: 1px solid var(--card-border);
  padding: 0 0.85rem;
  user-select: none;
  -webkit-user-select: none;
  box-sizing: border-box;
  z-index: 100;
  backdrop-filter: blur(12px);
}

.titlebar-left {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 260px;
}

.brand-badge {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.titlebar-logo {
  width: 22px;
  height: 22px;
  object-fit: contain;
}

.brand-title {
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  background: linear-gradient(135deg, var(--text-main) 60%, var(--accent-cyan));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.workspace-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.6rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.workspace-name {
  color: var(--text-main);
  font-weight: 600;
}

.branch-tag {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
  padding: 0.05rem 0.35rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
}

.titlebar-center {
  flex: 1;
  max-width: 520px;
  margin: 0 1rem;
}

.command-search-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 7px;
  padding: 0.3rem 0.75rem;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  color: var(--text-muted);
  font-size: 0.82rem;
}

.command-search-box:hover {
  border-color: var(--card-border-active);
  background: rgba(0, 229, 255, 0.06);
  color: var(--text-main);
}

.search-placeholder {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shortcut-kbd {
  font-size: 0.7rem;
  font-family: inherit;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 0.1rem 0.35rem;
  color: var(--text-muted);
}

.titlebar-right {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  min-width: 260px;
  justify-content: flex-end;
}

.lang-selector-header {
  position: relative;
}

.titlebar-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.25rem 0.55rem;
  border-radius: 6px;
  font-size: 0.78rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.titlebar-pill-btn:hover {
  border-color: var(--card-border-active);
  background: rgba(0, 229, 255, 0.08);
}

.lang-label {
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dropdown-caret {
  font-size: 0.65rem;
  opacity: 0.6;
}

.lang-dropdown-menu {
  position: absolute;
  top: calc(100% + 5px);
  right: 0;
  width: 190px;
  max-height: 280px;
  overflow-y: auto;
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(16px);
  z-index: 1000;
  padding: 0.3rem;
}

.lang-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.38rem 0.65rem;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.78rem;
  text-align: left;
  cursor: pointer;
}

.lang-option:hover {
  background: var(--input-bg);
}

.lang-option.active {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
  font-weight: 600;
}

.titlebar-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.titlebar-icon-btn:hover {
  border-color: var(--card-border-active);
  background: rgba(0, 229, 255, 0.08);
  transform: translateY(-1px);
}

.titlebar-icon-btn.active {
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.16);
  color: var(--accent-cyan);
}

.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
