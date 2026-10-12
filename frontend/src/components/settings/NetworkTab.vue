<template>
  <div class="tab-content">
    <!-- Section 2A: GitHub Acceleration Mirror -->
    <div class="settings-section">
      <h4 class="section-title">
        <span>⚡ {{ t("settings.github_proxy_title") || "GitHub Acceleration Mirror" }}</span>
      </h4>
      <p class="section-hint">
        {{
          t("settings.github_proxy_desc") ||
          "Accelerates asset downloads and update checks via public GitHub mirror endpoints."
        }}
      </p>

      <div class="form-group">
        <div class="label-row">
          <label class="form-label">{{ t("settings.github_proxy") || "Mirror URL Prefix" }}</label>
          <button v-if="proxyInputUrl" type="button" class="clear-mirror-btn" @click="$emit('setMirror', '')">
            {{ t("settings.proxy_direct") || "Direct" }}
          </button>
        </div>
        <input
          :value="proxyInputUrl"
          type="text"
          class="form-input"
          :placeholder="t('settings.proxy_placeholder') || 'https://proxy.example.com/'"
          @input="onProxyInputUrlChange"
        />
      </div>

      <div class="network-test-row">
        <button class="btn-secondary test-btn" :disabled="isTestingNet" @click="$emit('testConnection')">
          {{
            isTestingNet
              ? t("settings.testing_net") || "Testing Latency..."
              : t("settings.test_net") || "Test GitHub Connectivity"
          }}
        </button>
        <span v-if="netTestResult" class="test-result" :class="netTestSuccess ? 'success' : 'error'">
          {{ netTestResult }}
        </span>
      </div>
    </div>

    <!-- Section 2B: Custom Network Proxy -->
    <div class="settings-section margin-top">
      <h4 class="section-title">
        <span>🔌 {{ t("settings.system_proxy") || "Custom Network Proxy" }}</span>
      </h4>

      <div class="grid-form">
        <div class="form-group span-full">
          <label class="form-label">{{ t("settings.proxy_proto") || "Proxy Protocol" }}</label>
          <div class="protocol-radio-bar">
            <label class="protocol-pill" :class="{ active: proxyProtocol === 'direct' }">
              <input
                type="radio"
                :checked="proxyProtocol === 'direct'"
                value="direct"
                @change="onProtocolChange('direct')"
              />
              {{ t("settings.proxy_direct") || "Direct" }}
            </label>
            <label class="protocol-pill" :class="{ active: proxyProtocol === 'http' }">
              <input type="radio" :checked="proxyProtocol === 'http'" value="http" @change="onProtocolChange('http')" />
              HTTP
            </label>
            <label class="protocol-pill" :class="{ active: proxyProtocol === 'https' }">
              <input
                type="radio"
                :checked="proxyProtocol === 'https'"
                value="https"
                @change="onProtocolChange('https')"
              />
              HTTPS
            </label>
            <label class="protocol-pill" :class="{ active: proxyProtocol === 'socks4' }">
              <input
                type="radio"
                :checked="proxyProtocol === 'socks4'"
                value="socks4"
                @change="onProtocolChange('socks4')"
              />
              SOCKS4
            </label>
            <label class="protocol-pill" :class="{ active: proxyProtocol === 'socks5' }">
              <input
                type="radio"
                :checked="proxyProtocol === 'socks5'"
                value="socks5"
                @change="onProtocolChange('socks5')"
              />
              SOCKS5
            </label>
          </div>
        </div>

        <template v-if="proxyProtocol !== 'direct'">
          <div class="form-group">
            <label class="form-label">{{ t("settings.proxy_host") || "Proxy Host" }}</label>
            <input :value="proxyHost" type="text" class="form-input" placeholder="127.0.0.1" @input="onHostChange" />
          </div>

          <div class="form-group">
            <label class="form-label">{{ t("settings.proxy_port") || "Proxy Port" }}</label>
            <input
              :value="proxyPort"
              type="number"
              class="form-input"
              placeholder="7890"
              min="1"
              max="65535"
              @input="onPortChange"
            />
          </div>

          <div class="form-group">
            <label class="form-label">{{ t("settings.proxyAuthUserLabel") || "Username (Optional)" }}</label>
            <input
              :value="proxyUser"
              type="text"
              class="form-input"
              :placeholder="t('settings.proxyAuthUserPlaceholder') || 'Leave empty if none'"
              @input="onUserChange"
            />
          </div>

          <div class="form-group">
            <label class="form-label">{{ t("settings.proxyAuthPassLabel") || "Password (Optional)" }}</label>
            <input
              :value="proxyPassword"
              type="password"
              class="form-input"
              :placeholder="t('settings.proxyAuthPassPlaceholder') || 'Leave empty if none'"
              @input="onPasswordChange"
            />
          </div>
        </template>
      </div>

      <div class="network-test-row">
        <button class="btn-secondary test-btn" :disabled="isTestingProxy" @click="$emit('testNetworkProxy')">
          {{
            isTestingProxy
              ? t("settings.testingProxy") || "Testing Proxy..."
              : t("settings.testProxyConn") || "Test Proxy Connection"
          }}
        </button>
        <span v-if="proxyTestResult" class="test-result" :class="proxyTestSuccess ? 'success' : 'error'">
          {{ proxyTestResult }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from "../../i18n";

type ProxyProtocol = "direct" | "http" | "https" | "socks4" | "socks5";

defineProps<{
  proxyInputUrl: string;
  proxyProtocol: ProxyProtocol;
  proxyHost: string;
  proxyPort: number;
  proxyUser: string;
  proxyPassword: string;
  isTestingNet: boolean;
  netTestResult: string;
  netTestSuccess: boolean;
  isTestingProxy: boolean;
  proxyTestResult: string;
  proxyTestSuccess: boolean;
}>();

const emit = defineEmits<{
  (e: "update:proxyInputUrl", val: string): void;
  (e: "update:proxyProtocol", val: ProxyProtocol): void;
  (e: "update:proxyHost", val: string): void;
  (e: "update:proxyPort", val: number): void;
  (e: "update:proxyUser", val: string): void;
  (e: "update:proxyPassword", val: string): void;
  (e: "setMirror", val: string): void;
  (e: "testConnection"): void;
  (e: "testNetworkProxy"): void;
  (e: "change"): void;
}>();

function onProxyInputUrlChange(e: Event) {
  const val = (e.target as HTMLInputElement).value;
  emit("update:proxyInputUrl", val);
  emit("change");
}

function onProtocolChange(proto: ProxyProtocol) {
  emit("update:proxyProtocol", proto);
  emit("change");
}

function onHostChange(e: Event) {
  const val = (e.target as HTMLInputElement).value;
  emit("update:proxyHost", val);
  emit("change");
}

function onPortChange(e: Event) {
  const val = Number((e.target as HTMLInputElement).value) || 0;
  emit("update:proxyPort", val);
  emit("change");
}

function onUserChange(e: Event) {
  const val = (e.target as HTMLInputElement).value;
  emit("update:proxyUser", val);
  emit("change");
}

function onPasswordChange(e: Event) {
  const val = (e.target as HTMLInputElement).value;
  emit("update:proxyPassword", val);
  emit("change");
}
</script>

<style scoped>
.tab-content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.settings-section {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.25rem;
}

.settings-section.margin-top {
  margin-top: 0.5rem;
}

.section-title {
  font-size: 0.95rem;
  font-weight: 700;
  margin-bottom: 1rem;
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

.span-full {
  grid-column: span 2;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  font-weight: 600;
}

.section-hint {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-top: -0.5rem;
  margin-bottom: 0.75rem;
  line-height: 1.4;
}

.form-input {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  color: var(--text-main);
  padding: 0.55rem 0.75rem;
  font-size: 0.85rem;
  outline: none;
}

[data-theme="light"] .form-input {
  background: #ffffff;
  border-color: #cbd5e1;
}

.form-input::placeholder {
  color: rgba(148, 163, 184, 0.42);
  opacity: 1;
  font-size: 0.82rem;
  font-weight: 400;
  transition: color 0.2s ease;
}

.form-input:focus::placeholder {
  color: rgba(148, 163, 184, 0.22);
}

.form-input:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 10px var(--accent-cyan-glow);
}

.protocol-radio-bar {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 0.2rem;
}

.protocol-pill {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.8rem;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  transition: all 0.2s ease;
}

.protocol-pill input {
  display: none;
}

.protocol-pill:hover {
  border-color: var(--accent-cyan);
  color: var(--text-main);
}

.protocol-pill.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  font-weight: 700;
}

[data-theme="light"] .protocol-pill {
  background: #ffffff;
  color: #475569;
  border-color: #cbd5e1;
}

[data-theme="light"] .protocol-pill.active {
  background: #0284c7;
  border-color: #0284c7;
  color: #ffffff;
  font-weight: 700;
}

.label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.clear-mirror-btn {
  background: none;
  border: none;
  color: var(--accent-cyan);
  font-size: 0.775rem;
  font-weight: 500;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
  transition: opacity 0.2s;
}

.clear-mirror-btn:hover {
  opacity: 0.8;
}

.network-test-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 1rem;
}

.test-btn {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

[data-theme="light"] .test-btn {
  background: #ffffff;
  color: #0f172a;
  border-color: #cbd5e1;
}

.test-btn:hover:not(:disabled) {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 10px var(--accent-cyan-glow);
}

.test-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.test-result {
  font-size: 0.775rem;
}

.test-result.success {
  color: var(--success);
}

.test-result.error {
  color: #ef4444;
}

@media (max-width: 600px) {
  .grid-form {
    grid-template-columns: 1fr;
  }
  .span-full {
    grid-column: span 1;
  }
}
</style>
