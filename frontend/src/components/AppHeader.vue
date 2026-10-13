<template>
  <header class="app-header">
    <div class="brand">
      <img src="/logo.png" alt="UniRTM" class="logo-img" />
      <div>
        <h1>{{ t("app.title") || "UniRTM" }}</h1>
        <span class="sub-brand">{{ t("app.subtitle") || "Dev tools, env vars, and tasks in one CLI with built-in security" }}</span>
      </div>
    </div>

    <div class="header-actions">
      <!-- Quick Language Switcher Dropdown -->
      <div class="lang-selector-header" ref="langDropdownRef">
        <button class="lang-pill-btn" :title="t('settings.language') || 'Language'" @click.stop="toggleLangMenu">
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
        class="icon-action-btn"
        :title="
          isLight ? t('theme.toggleDark') || 'Switch to Dark Theme' : t('theme.toggleLight') || 'Switch to Light Theme'
        "
        @click="emit('toggle-theme')"
      >
        <span>{{ isLight ? "🌙" : "☀️" }}</span>
      </button>

      <!-- Log Viewer Button -->
      <button class="icon-action-btn" :title="t('log.title') || 'Logs'" @click="emit('open-logs')">
        <span>📜</span>
      </button>

      <!-- Settings Button -->
      <button class="icon-action-btn" :title="t('settings.title') || 'Settings'" @click="emit('open-settings')">
        <span>⚙️</span>
      </button>

      <!-- About Button -->
      <button class="icon-action-btn" :title="t('about.title') || 'About'" @click="emit('open-about')">
        <span>ℹ️</span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { SUPPORTED_LANGUAGES, t } from "../i18n";

const props = defineProps<{
  currentLang: string;
  currentTheme: string;
}>();

const emit = defineEmits<{
  (e: "toggle-theme"): void;
  (e: "open-settings"): void;
  (e: "open-about"): void;
  (e: "open-logs"): void;
  (e: "select-lang", lang: string): void;
}>();

const isLangMenuOpen = ref(false);
const langDropdownRef = ref<HTMLElement | null>(null);

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

const isLight = computed(() => {
  if (props.currentTheme === "light") return true;
  if (props.currentTheme === "dark") return false;
  return typeof document !== "undefined" && document.documentElement.getAttribute("data-theme") === "light";
});

function toggleLangMenu() {
  isLangMenuOpen.value = !isLangMenuOpen.value;
}

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
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  user-select: none;
  -webkit-user-select: none;
}

.brand {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logo-img {
  width: 48px;
  height: 48px;
  object-fit: contain;
  filter: drop-shadow(0 4px 6px rgba(0, 0, 229, 0.2));
}

.brand h1 {
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0;
  line-height: 1.2;
  color: var(--text-main);
  background: linear-gradient(135deg, var(--text-main) 60%, var(--accent-cyan));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.sub-brand {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  position: relative;
}

.lang-selector-header {
  position: relative;
}

.lang-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.4rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(8px);
}

.lang-pill-btn:hover {
  border-color: var(--card-border-active);
  background: var(--input-bg);
  transform: translateY(-1px);
}

.lang-icon {
  font-size: 0.95rem;
}

.lang-label {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dropdown-caret {
  font-size: 0.7rem;
  opacity: 0.6;
}

.lang-dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 200px;
  max-height: 320px;
  overflow-y: auto;
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(16px);
  z-index: 1000;
  padding: 0.35rem;
}

.lang-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.45rem 0.75rem;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--text-main);
  font-size: 0.82rem;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease;
}

.lang-option:hover {
  background: var(--input-bg);
}

.lang-option.active {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
  font-weight: 600;
}

[data-theme="light"] .lang-option.active {
  background: rgba(2, 132, 199, 0.12);
}

.opt-check {
  font-size: 0.8rem;
  font-weight: bold;
}

.icon-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  color: var(--text-main);
  font-size: 1.1rem;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  backdrop-filter: blur(8px);
}

.icon-action-btn:hover {
  border-color: var(--card-border-active);
  background: var(--input-bg);
  transform: translateY(-1px);
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
