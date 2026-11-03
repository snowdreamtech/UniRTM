<template>
  <div class="tab-content">
    <!-- Card 1: Appearance & Language -->
    <div class="settings-section appearance-card">
      <h4 class="section-title">
        <span>🎨 {{ t("settings.section_appearance") || "Appearance & Language" }}</span>
      </h4>

      <div class="grid-form">
        <!-- Language Selection -->
        <div class="form-group highlight-form-group">
          <label class="form-label highlight-label">🌐 {{ t("settings.language") || "Language" }}</label>
          <CustomSelect :model-value="language" :options="languageOptions" @change="onLanguageChange" />
        </div>

        <!-- Theme Selection -->
        <div class="form-group">
          <label class="form-label">🎨 {{ t("settings.theme") || "Appearance Theme" }}</label>
          <CustomSelect :model-value="theme" :options="themeOptions" @change="onThemeChange" />
        </div>
      </div>
    </div>

    <!-- Card 2: Software Updates -->
    <div class="settings-section">
      <h4 class="section-title">
        <span>🔄 {{ t("settings.app_update") || "Software Updates" }}</span>
      </h4>

      <div class="card-options">
        <label class="option-card" :class="{ selected: autoCheckUpdate === true }">
          <input
            type="radio"
            name="autoCheckUpdate"
            :value="true"
            :checked="autoCheckUpdate === true"
            @change="onAutoCheckUpdateChange(true)"
          />
          <div class="option-content">
            <span class="option-title">{{ t("settings.update_auto") || "Automatically check on startup" }}</span>
            <span class="option-desc">
              {{ t("settings.update_auto_desc") || "Check for new releases in the background when app launches" }}
            </span>
          </div>
        </label>

        <label class="option-card" :class="{ selected: autoCheckUpdate === false }">
          <input
            type="radio"
            name="autoCheckUpdate"
            :value="false"
            :checked="autoCheckUpdate === false"
            @change="onAutoCheckUpdateChange(false)"
          />
          <div class="option-content">
            <span class="option-title">{{ t("settings.update_manual") || "Manual check only" }}</span>
            <span class="option-desc">
              {{ t("settings.update_manual_desc") || "Only check for new versions when manually triggered" }}
            </span>
          </div>
        </label>
      </div>
    </div>

    <!-- Card 3: System Tray & Window Behavior -->
    <div class="settings-section">
      <h4 class="section-title">
        <span>📌 {{ t("settings.tray_section") || "System Tray & Window Behavior" }}</span>
      </h4>

      <div class="tray-content">
        <div class="switch-row">
          <div class="switch-label-group">
            <span class="switch-title">{{ t("settings.enable_tray") || "Enable System Tray" }}</span>
            <span class="switch-desc">
              {{ t("settings.enable_tray_desc") || "Keep app alive in system tray / menu bar (disabled by default)" }}
            </span>
          </div>
          <label class="toggle-switch">
            <input
              type="checkbox"
              :checked="enableTray === true"
              @change="onEnableTrayChange(($event.target as HTMLInputElement).checked)"
            />
            <span class="slider"></span>
          </label>
        </div>

        <transition name="slide-fade">
          <div v-if="enableTray" class="close-action-section">
            <label class="form-label sub-label">
              {{ t("settings.close_action") || "When clicking window close button:" }}
            </label>
            <div class="card-options horizontal">
              <label class="option-card compact" :class="{ selected: closeAction === 'quit' }">
                <input
                  type="radio"
                  name="closeAction"
                  value="quit"
                  :checked="closeAction === 'quit'"
                  @change="onCloseActionChange('quit')"
                />
                <div class="option-content">
                  <span class="option-title">{{ t("settings.close_action_quit") || "Quit application" }}</span>
                </div>
              </label>

              <label class="option-card compact" :class="{ selected: closeAction === 'minimize_to_tray' }">
                <input
                  type="radio"
                  name="closeAction"
                  value="minimize_to_tray"
                  :checked="closeAction === 'minimize_to_tray'"
                  @change="onCloseActionChange('minimize_to_tray')"
                />
                <div class="option-content">
                  <span class="option-title">{{
                    t("settings.close_action_minimize") || "Minimize to system tray"
                  }}</span>
                </div>
              </label>
            </div>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from "../../i18n";
import CustomSelect from "../CustomSelect.vue";

defineProps<{
  language: string;
  theme: string;
  autoCheckUpdate: boolean;
  enableTray: boolean;
  closeAction: string;
  languageOptions: Array<{ value: string; label: string }>;
  themeOptions: Array<{ value: string; label: string }>;
}>();

const emit = defineEmits<{
  (e: "update:language", val: string): void;
  (e: "update:theme", val: string): void;
  (e: "update:autoCheckUpdate", val: boolean): void;
  (e: "update:enableTray", val: boolean): void;
  (e: "update:closeAction", val: string): void;
  (e: "languageChange", val: string): void;
  (e: "themeChange", val: string): void;
  (e: "change"): void;
}>();

function onLanguageChange(val: string) {
  emit("update:language", val);
  emit("languageChange", val);
}

function onThemeChange(val: string) {
  emit("update:theme", val);
  emit("themeChange", val);
}

function onAutoCheckUpdateChange(val: boolean) {
  emit("update:autoCheckUpdate", val);
  emit("change");
}

function onEnableTrayChange(val: boolean) {
  emit("update:enableTray", val);
  emit("change");
}

function onCloseActionChange(val: string) {
  emit("update:closeAction", val);
  emit("change");
}
</script>

<style scoped>
.tab-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.settings-section {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.15rem 1.25rem;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.settings-section.appearance-card {
  position: relative;
  z-index: 10;
}

.section-title {
  font-size: 0.925rem;
  font-weight: 700;
  margin-bottom: 0.85rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--accent-cyan);
}

.grid-form {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-group.highlight-form-group {
  position: relative;
  z-index: 20;
}

.form-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  font-weight: 600;
}

.sub-label {
  font-size: 0.78rem;
  margin-bottom: 0.15rem;
}

.highlight-label {
  color: var(--accent-cyan);
}

/* Card Options (Radios as Cards) */
.card-options {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.card-options.horizontal {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.option-card {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem 0.95rem;
  border-radius: 10px;
  border: 1px solid var(--card-border);
  background: rgba(255, 255, 255, 0.02);
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.option-card:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(0, 229, 255, 0.35);
}

.option-card.selected {
  background: rgba(0, 229, 255, 0.07);
  border-color: var(--accent-cyan);
  box-shadow: 0 0 10px rgba(0, 229, 255, 0.1);
}

.option-card.compact {
  padding: 0.65rem 0.85rem;
  align-items: center;
}

.option-card input[type="radio"] {
  accent-color: var(--accent-cyan);
  cursor: pointer;
  width: 15px;
  height: 15px;
  margin-top: 0.15rem;
  flex-shrink: 0;
}

.option-card.compact input[type="radio"] {
  margin-top: 0;
}

.option-content {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.option-title {
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--text-main);
  line-height: 1.3;
}

.option-desc {
  font-size: 0.735rem;
  color: var(--text-muted);
  line-height: 1.35;
}

/* Tray setting card */
.tray-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.switch-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.switch-label-group {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.switch-title {
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--text-main);
}

.switch-desc {
  font-size: 0.735rem;
  color: var(--text-muted);
}

.close-action-section {
  padding-top: 0.75rem;
  border-top: 1px dashed var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

/* Slide Fade Transition */
.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.22s ease-out;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* Modern Toggle Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
  flex-shrink: 0;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(148, 163, 184, 0.3);
  transition: 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 22px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--accent-cyan);
}

input:checked + .slider:before {
  transform: translateX(18px);
}

@media (max-width: 600px) {
  .grid-form {
    grid-template-columns: 1fr;
  }
  .card-options.horizontal {
    grid-template-columns: 1fr;
  }
}
</style>
