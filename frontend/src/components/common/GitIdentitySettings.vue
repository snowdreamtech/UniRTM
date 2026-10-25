<template>
  <div class="git-identity-settings" id="git-identity-settings-panel">
    <!-- Header Divider & Title aligned with SettingsView design system -->
    <div class="vault-divider"></div>
    <div class="section-title-wrap">
      <span class="section-icon">
        <span class="codicon codicon-shield"></span>
      </span>
      <div>
        <h3>{{ t("git_identity.title") || "Git Identity & Credential Vault" }}</h3>
        <span class="section-sub">
          {{
            t("git_identity.desc") ||
            "Manage multiple Git identities locally isolated per repository. Zero private key touching, system agent first, and hardened one-time nonce IPC."
          }}
        </span>
      </div>
    </div>

    <!-- Security Agents Status Cards (Matches sec-card system) -->
    <div class="agent-status-row">
      <div class="status-card" :class="{ active: agentStatus.sshAgentActive }">
        <div class="card-header">
          <div class="header-left">
            <span class="codicon codicon-key"></span>
            <span class="card-title">SSH Agent</span>
          </div>
          <span class="sec-badge" :class="agentStatus.sshAgentActive ? 'green' : 'gray'">
            {{
              agentStatus.sshAgentActive
                ? t("git_identity.active") || "Active"
                : t("git_identity.inactive") || "Inactive"
            }}
          </span>
        </div>
        <div class="card-body">
          <span v-if="agentStatus.sshAgentActive" class="body-text">
            {{
              t("git_identity.loaded_keys", { count: agentStatus.sshAgentKeys }) ||
              `${agentStatus.sshAgentKeys} keys loaded in system agent`
            }}
          </span>
          <span v-else class="text-muted">{{
            t("git_identity.agent_not_found") || "No active SSH_AUTH_SOCK detected"
          }}</span>
        </div>
      </div>

      <div class="status-card" :class="{ active: agentStatus.gpgAgentActive }">
        <div class="card-header">
          <div class="header-left">
            <span class="codicon codicon-lock"></span>
            <span class="card-title">GPG Agent</span>
          </div>
          <span class="sec-badge" :class="agentStatus.gpgAgentActive ? 'green' : 'gray'">
            {{
              agentStatus.gpgAgentActive
                ? t("git_identity.active") || "Active"
                : t("git_identity.inactive") || "Inactive"
            }}
          </span>
        </div>
        <div class="card-body">
          <span v-if="agentStatus.gpgAgentActive" class="body-text">{{
            t("git_identity.gpg_ready") || "gpg-agent running and responsive"
          }}</span>
          <span v-else class="text-muted">{{ t("git_identity.gpg_not_found") || "gpg-agent not responding" }}</span>
        </div>
      </div>

      <div class="status-card vault-control">
        <div class="card-header">
          <div class="header-left">
            <span class="codicon codicon-clear-all"></span>
            <span class="card-title">{{ t("git_identity.memory_vault") || "Session Vault" }}</span>
          </div>
        </div>
        <div class="card-body">
          <button class="clear-vault-btn" @click="handleClearKeyring">
            <span class="codicon codicon-trash"></span>
            <span>{{ t("git_identity.clear_keyring") || "Purge In-Memory Keyring" }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Profiles Section -->
    <div class="profiles-section">
      <div class="section-toolbar">
        <h4>{{ t("git_identity.profiles_list") || "Identity Profiles" }}</h4>
        <button class="action-btn add-btn" @click="openAddForm">
          <span class="codicon codicon-plus"></span>
          <span>{{ t("git_identity.new_profile") || "New Profile" }}</span>
        </button>
      </div>

      <div class="profiles-grid">
        <div
          v-for="prof in profiles"
          :key="prof.id"
          class="profile-card"
          :class="{ current: prof.id === activeProfileId }"
        >
          <div class="card-top">
            <div class="profile-name-row">
              <span class="name">{{ formatProfileName(prof) }}</span>
              <span v-if="prof.id === activeProfileId" class="sec-badge cyan">{{
                t("git_identity.active_workspace") || "Active in Workspace"
              }}</span>
              <span v-if="prof.isDefault" class="sec-badge gray">{{ t("git_identity.default") || "Default" }}</span>
            </div>
            <div class="card-actions">
              <button
                v-if="prof.id !== activeProfileId"
                class="icon-btn apply-btn"
                :title="t('git_identity.apply_tooltip') || 'Apply to current repository (.git/config)'"
                @click="applyProfile(prof.id)"
              >
                <span class="codicon codicon-check"></span>
              </button>
              <button class="icon-btn" :title="t('git_identity.edit') || 'Edit'" @click="editProfile(prof)">
                <span class="codicon codicon-edit"></span>
              </button>
              <button
                class="icon-btn delete-btn"
                :title="t('git_identity.delete') || 'Delete'"
                @click="deleteProfile(prof.id)"
              >
                <span class="codicon codicon-trash"></span>
              </button>
            </div>
          </div>

          <div class="card-details">
            <div class="detail-row">
              <span class="label">User:</span>
              <span class="val font-mono text-truncate" :title="`${prof.userName} <${prof.userEmail}>`">
                {{ prof.userName }} &lt;{{ prof.userEmail }}&gt;
              </span>
            </div>
            <div v-if="prof.sshKeyPath" class="detail-row">
              <span class="label">SSH:</span>
              <span class="val font-mono text-truncate" :title="prof.sshKeyPath">{{ prof.sshKeyPath }}</span>
            </div>
            <div v-if="prof.signingKey" class="detail-row">
              <span class="label">GPG:</span>
              <span class="val font-mono text-gpg" :title="`Key ID / Fingerprint: ${prof.signingKey}`">
                <span class="gpg-key-id">{{ formatGpgKey(prof.signingKey) }}</span>
                <span v-if="prof.gpgSign" class="auto-sign-badge">
                  {{ t("git_identity.auto_sign_tag") || "(Auto-sign)" }}
                </span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Security Audit Trail Accordion -->
    <div class="audit-trail-section">
      <div class="section-toolbar pointer" @click="isAuditExpanded = !isAuditExpanded">
        <div class="title-with-icon">
          <span class="codicon" :class="isAuditExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"></span>
          <h4>{{ t("git_identity.audit_trail") || "Security & Credential Audit Trail" }}</h4>
          <span class="audit-count">({{ auditRecords.length }})</span>
        </div>
        <button class="refresh-btn icon-btn" @click.stop="loadAuditRecords">
          <span class="codicon codicon-refresh"></span>
        </button>
      </div>

      <div v-if="isAuditExpanded" class="audit-table-wrapper">
        <table v-if="auditRecords.length > 0" class="audit-table">
          <thead>
            <tr>
              <th>Time</th>
              <th>Action</th>
              <th>Type</th>
              <th>Key / Target</th>
              <th>Result</th>
              <th>Client IP</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rec in auditRecords.slice().reverse()" :key="rec.id">
              <td class="font-mono">{{ new Date(rec.timestamp).toLocaleTimeString() }}</td>
              <td>
                <span class="action-tag">{{ rec.action }}</span>
              </td>
              <td>
                <span class="type-tag" :class="rec.type">{{ rec.type }}</span>
              </td>
              <td class="font-mono text-truncate max-w-150" :title="rec.targetKey">{{ rec.targetKey || "-" }}</td>
              <td>
                <span class="result-badge" :class="rec.result">{{ rec.result }}</span>
              </td>
              <td class="font-mono">{{ rec.clientIp }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-audit">
          {{ t("git_identity.no_audit_events") || "No credential access events recorded yet in this session." }}
        </div>
      </div>
    </div>

    <!-- Edit / Add Modal Dialog -->
    <div v-if="isEditing" class="profile-modal-overlay">
      <div class="profile-modal">
        <div class="modal-header">
          <h3>
            {{
              editingProfile.id
                ? t("git_identity.edit_profile") || "Edit Profile"
                : t("git_identity.create_profile") || "Create Profile"
            }}
          </h3>
          <button class="close-btn" @click="isEditing = false">
            <span class="codicon codicon-close"></span>
          </button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>{{ t("git_identity.profile_name") || "Profile Name" }} *</label>
            <input v-model="editingProfile.name" type="text" placeholder="e.g. Work / Corporate" class="input-text" />
          </div>

          <div class="form-row">
            <div class="form-group flex-1">
              <label>{{ t("git_identity.user_name") || "user.name" }} *</label>
              <input v-model="editingProfile.userName" type="text" placeholder="e.g. Alice Doe" class="input-text" />
            </div>
            <div class="form-group flex-1">
              <label>{{ t("git_identity.user_email") || "user.email" }} *</label>
              <input
                v-model="editingProfile.userEmail"
                type="email"
                placeholder="e.g. alice@company.com"
                class="input-text"
              />
            </div>
          </div>

          <!-- SSH Key Selection -->
          <div class="form-group">
            <div class="field-label-row">
              <label>{{ t("git_identity.ssh_key") || "SSH Key (Local Path / Discovered)" }}</label>
              <span v-if="agentStatus.sshAgentActive" class="agent-hint-tag">
                <span class="codicon codicon-pass"></span>
                {{ t("git_identity.ssh_agent_ready") || "SSH Agent Active (Password-less)" }}
              </span>
            </div>
            <div class="input-with-select">
              <input
                v-model="editingProfile.sshKeyPath"
                type="text"
                :placeholder="t('git_identity.ssh_key_placeholder') || 'Leave empty to use SSH Agent (password-less)'"
                class="input-text"
              />
              <select
                v-if="systemKeys.sshKeys.length > 0"
                :value="selectedSshKeyOption"
                class="select-discovered"
                @change="onSelectDiscoveredSSH"
              >
                <option value="">
                  {{ t("git_identity.use_ssh_agent_default") || "-- Use SSH Agent / System Default --" }}
                </option>
                <option v-for="k in systemKeys.sshKeys" :key="k.path" :value="k.path">
                  {{ k.name }} ({{ k.algorithm }})
                </option>
              </select>
            </div>
            <small class="field-helper-text">
              {{
                t("git_identity.ssh_key_hint") ||
                "When left empty, Git will automatically use SSH Agent and system SSH configuration for password-less authentication."
              }}
            </small>
          </div>

          <!-- GPG Key Selection -->
          <div class="form-group">
            <div class="field-label-row">
              <label>{{ t("git_identity.gpg_signing_key") || "GPG Signing Key ID" }}</label>
              <span v-if="agentStatus.gpgAgentActive" class="agent-hint-tag">
                <span class="codicon codicon-pass"></span>
                {{ t("git_identity.gpg_agent_ready") || "GPG Agent Ready" }}
              </span>
            </div>
            <div class="input-with-select">
              <input
                v-model="editingProfile.signingKey"
                type="text"
                placeholder="e.g. 3AA5C34371567BD2"
                class="input-text"
              />
              <select
                v-if="systemKeys.gpgKeys.length > 0"
                :value="selectedGpgKeyOption"
                class="select-discovered"
                @change="onSelectDiscoveredGPG"
              >
                <option value="">
                  {{ t("git_identity.select_discovered_gpg") || "-- Select GPG Key --" }}
                </option>
                <option v-for="g in systemKeys.gpgKeys" :key="g.keyId" :value="g.keyId">
                  {{ formatGpgKey(g.keyId) }} ({{ g.userId }})
                </option>
              </select>
            </div>
            <small class="field-helper-text">
              {{
                t("git_identity.gpg_key_hint") ||
                "GPG Agent caches your passphrase in memory for the session duration, enabling seamless signing without repeated password prompts."
              }}
            </small>
          </div>

          <div class="form-row check-row">
            <label class="checkbox-label">
              <input v-model="editingProfile.gpgSign" type="checkbox" />
              <span>{{
                t("git_identity.enable_auto_sign") || "Automatically sign commits (commit.gpgsign = true)"
              }}</span>
            </label>
            <label class="checkbox-label">
              <input v-model="editingProfile.isDefault" type="checkbox" />
              <span>{{ t("git_identity.set_as_default") || "Set as default profile" }}</span>
            </label>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn secondary" @click="isEditing = false">{{ t("common.cancel") || "Cancel" }}</button>
          <button class="btn primary" :disabled="!editingProfile.name" @click="handleSaveProfile">
            {{ t("common.save") || "Save Profile" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useGitIdentity } from "../../composables/useGitIdentity";
import type { GitProfile } from "../../types/credential";
import { t } from "../../i18n";

const {
  profiles,
  activeProfileId,
  systemKeys,
  agentStatus,
  auditRecords,
  loadProfiles,
  saveProfile,
  deleteProfile,
  applyProfile,
  loadSystemKeys,
  loadAgentStatus,
  loadAuditRecords,
  clearKeyringCache,
} = useGitIdentity();

const isEditing = ref(false);
const isAuditExpanded = ref(false);

function formatProfileName(prof: GitProfile | Partial<GitProfile>): string {
  if (!prof) return t("git_identity.default_profile") || "Default Profile";
  if (prof.id === "default" || prof.name === "Default Profile" || prof.name === "Default") {
    return t("git_identity.default_profile") || "Default Profile";
  }
  return prof.name || t("git_identity.default_profile") || "Default Profile";
}

function formatGpgKey(key?: string): string {
  if (!key) return "";
  if (key.length > 16) {
    return key.slice(-16);
  }
  return key;
}

const editingProfile = ref<Partial<GitProfile>>({
  name: "",
  userName: "",
  userEmail: "",
  sshKeyPath: "",
  signingKey: "",
  gpgSign: false,
  isDefault: false,
});

const selectedSshKeyOption = computed(() => {
  if (!editingProfile.value.sshKeyPath) return "";
  const path = editingProfile.value.sshKeyPath.trim();
  const matched = systemKeys.value.sshKeys.find(
    (k) => k.path === path || path.endsWith(k.name) || k.path.endsWith(path)
  );
  return matched ? matched.path : path;
});

const selectedGpgKeyOption = computed(() => {
  if (!editingProfile.value.signingKey) return "";
  const signKey = editingProfile.value.signingKey.trim().toLowerCase();
  const matched = systemKeys.value.gpgKeys.find(
    (g) =>
      g.keyId.toLowerCase() === signKey ||
      signKey.endsWith(g.keyId.toLowerCase()) ||
      g.keyId.toLowerCase().endsWith(signKey) ||
      (g.shortKeyId && signKey.endsWith(g.shortKeyId.toLowerCase()))
  );
  return matched ? matched.keyId : "";
});

const openAddForm = () => {
  editingProfile.value = {
    name: "",
    userName: "",
    userEmail: "",
    sshKeyPath: "",
    signingKey: "",
    gpgSign: false,
    isDefault: false,
  };
  isEditing.value = true;
};

const editProfile = (prof: GitProfile) => {
  editingProfile.value = {
    ...prof,
    name: formatProfileName(prof),
  };
  isEditing.value = true;
};

const handleSaveProfile = async () => {
  if (!editingProfile.value.name) return;
  await saveProfile(editingProfile.value);
  isEditing.value = false;
};

const onSelectDiscoveredSSH = (e: Event) => {
  const target = e.target as HTMLSelectElement;
  editingProfile.value.sshKeyPath = target.value;
};

const onSelectDiscoveredGPG = (e: Event) => {
  const target = e.target as HTMLSelectElement;
  if (target.value) {
    editingProfile.value.signingKey = target.value;
    editingProfile.value.gpgSign = true;
  } else {
    editingProfile.value.signingKey = "";
  }
};

const handleClearKeyring = async () => {
  await clearKeyringCache();
  await loadAuditRecords();
};

onMounted(() => {
  loadProfiles();
  loadSystemKeys();
  loadAgentStatus();
  loadAuditRecords();
});
</script>

<style scoped>
.git-identity-settings {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  margin-top: 1rem;
}

.vault-divider {
  height: 1px;
  background: var(--card-border);
  margin: 0.5rem 0 0.8rem 0;
  opacity: 0.6;
}

.section-title-wrap {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.section-icon {
  font-size: 1.2rem;
  color: var(--accent-cyan);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: rgba(0, 229, 255, 0.1);
  border: 1px solid var(--card-border);
  flex-shrink: 0;
}

.section-title-wrap h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0 0 4px 0;
}

.section-sub {
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.4;
  display: block;
}

/* Agent Status Row */
.agent-status-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 14px;
}

.status-card {
  padding: 14px;
  border-radius: 6px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: all 0.2s ease;
}

.status-card:hover {
  border-color: var(--card-border-active);
}

.status-card.active {
  border-color: rgba(16, 185, 129, 0.35);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-main);
  font-size: 0.85rem;
  font-weight: 600;
}

.sec-badge {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 4px;
  text-transform: uppercase;
}

.sec-badge.green {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.sec-badge.cyan {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
}

.sec-badge.purple {
  background: rgba(157, 78, 221, 0.15);
  color: var(--accent-purple);
}

.sec-badge.gray {
  background: var(--hover-bg);
  color: var(--text-muted);
  border: 1px solid var(--card-border);
}

.card-body {
  font-size: 0.78rem;
}

.card-body .body-text {
  color: var(--text-main);
}

.card-body .text-muted {
  color: var(--text-muted);
}

.clear-vault-btn {
  background: var(--hover-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 4px 10px;
  font-size: 0.75rem;
  border-radius: 4px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.clear-vault-btn:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.35);
}

/* Profiles Section */
.profiles-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.section-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.section-toolbar h4 {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0;
}

.action-btn.add-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: #0284c7;
  color: #ffffff;
  font-weight: 500;
  border: 1px solid rgba(56, 189, 248, 0.35);
  padding: 5px 12px;
  border-radius: 4px;
  font-size: 0.78rem;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(2, 132, 199, 0.25);
  transition: all 0.2s ease;
}

.action-btn.add-btn:hover {
  background: #0369a1;
  border-color: #38bdf8;
  box-shadow: 0 3px 10px rgba(2, 132, 199, 0.35);
}

.profiles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 12px;
}

.profile-card {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
  overflow: hidden;
  box-sizing: border-box;
  transition: all 0.2s ease;
}

.profile-card:hover {
  border-color: var(--card-border-active);
}

.profile-card.current {
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.05);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  min-width: 0;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  min-width: 0;
}

.profile-name-row .name {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-main);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.icon-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  padding: 3px 6px;
  cursor: pointer;
  border-radius: 3px;
  transition: all 0.15s ease;
}

.icon-btn:hover {
  background: var(--hover-bg);
  color: var(--text-main);
}

.apply-btn {
  color: var(--success);
}

.apply-btn:hover {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.delete-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.15);
}

.card-details {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 0.76rem;
  min-width: 0;
  overflow: hidden;
}

.detail-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
  overflow: hidden;
}

.detail-row .label {
  color: var(--text-muted);
  width: 34px;
  flex-shrink: 0;
  font-size: 0.74rem;
}

.detail-row .val {
  color: var(--text-main);
  min-width: 0;
  flex: 1;
}

.detail-row .val.text-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-row .val.text-gpg {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  min-width: 0;
  color: var(--accent-purple);
}

.gpg-key-id {
  font-family: monospace;
  word-break: break-all;
}

.auto-sign-badge {
  font-size: 0.7rem;
  opacity: 0.85;
  white-space: nowrap;
}

/* Audit Trail */
.audit-trail-section {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 12px 14px;
  transition: all 0.2s ease;
}

.section-toolbar.pointer {
  cursor: pointer;
  user-select: none;
}

.title-with-icon {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-main);
}

.title-with-icon h4 {
  font-size: 0.88rem;
  font-weight: 600;
  margin: 0;
}

.audit-count {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.audit-table-wrapper {
  margin-top: 10px;
  overflow-x: auto;
}

.audit-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.75rem;
}

.audit-table th {
  padding: 8px;
  text-align: left;
  border-bottom: 1px solid var(--card-border);
  color: var(--text-muted);
  background: var(--header-bg);
  font-weight: 600;
}

.audit-table td {
  padding: 7px 8px;
  text-align: left;
  border-bottom: 1px solid var(--card-border);
  color: var(--text-main);
}

.action-tag {
  background: var(--hover-bg);
  border: 1px solid var(--card-border);
  padding: 2px 5px;
  border-radius: 3px;
}

.type-tag.ssh {
  color: var(--accent-cyan);
}

.type-tag.gpg {
  color: var(--accent-purple);
}

.type-tag.security_violation {
  color: #ef4444;
  font-weight: 600;
}

.result-badge.success {
  color: var(--success);
  font-weight: 600;
}

.result-badge.denied_invalid_nonce {
  color: #ef4444;
  font-weight: 600;
}

.empty-audit {
  padding: 14px;
  text-align: center;
  font-size: 0.78rem;
  color: var(--text-muted);
}

/* Modal */
.profile-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.profile-modal {
  width: 480px;
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.25);
  color: var(--text-main);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--card-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-header h3 {
  font-size: 0.95rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-main);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}

.close-btn:hover {
  color: var(--text-main);
  background: var(--hover-bg);
}

.modal-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.form-group label {
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--text-muted);
}

.form-row {
  display: flex;
  gap: 10px;
}

.flex-1 {
  flex: 1;
}

.input-text {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 7px 10px;
  font-size: 0.82rem;
  border-radius: 4px;
  outline: none;
  transition: all 0.15s ease;
}

.input-text:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 2px rgba(0, 229, 255, 0.2);
}

.input-with-select {
  display: flex;
  gap: 8px;
}

.input-with-select .input-text {
  flex: 1;
}

.select-discovered {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.78rem;
  border-radius: 4px;
  padding: 0 8px;
  max-width: 170px;
}

.check-row {
  flex-direction: column;
  gap: 8px;
  margin-top: 6px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.78rem;
  color: var(--text-main);
  cursor: pointer;
}

.modal-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--card-border);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  background: var(--header-bg);
}

.btn {
  padding: 6px 16px;
  font-size: 0.8rem;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid transparent;
  font-weight: 500;
  transition: all 0.15s ease;
}

.btn.primary {
  background: #0284c7;
  color: #ffffff;
  border-color: rgba(56, 189, 248, 0.4);
  box-shadow: 0 1px 3px rgba(2, 132, 199, 0.25);
}

.btn.primary:hover:not(:disabled) {
  background: #0369a1;
  border-color: #38bdf8;
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.35);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

.btn.secondary {
  background: transparent;
  border: 1px solid var(--card-border);
  color: var(--text-main);
}

.btn.secondary:hover {
  background: var(--hover-bg);
}

/* Light theme overrides */
:root[data-theme="light"] .action-btn.add-btn,
[data-theme="light"] .action-btn.add-btn,
:root[data-theme="light"] .btn.primary,
[data-theme="light"] .btn.primary {
  background: #0284c7 !important;
  color: #ffffff !important;
  border: 1px solid #0284c7 !important;
  box-shadow: 0 1px 3px rgba(2, 132, 199, 0.2) !important;
}

:root[data-theme="light"] .action-btn.add-btn:hover,
[data-theme="light"] .action-btn.add-btn:hover,
:root[data-theme="light"] .btn.primary:hover:not(:disabled),
[data-theme="light"] .btn.primary:hover:not(:disabled) {
  background: #0369a1 !important;
  border-color: #0369a1 !important;
}

:root[data-theme="light"] .btn.secondary,
[data-theme="light"] .btn.secondary {
  background: #f1f5f9 !important;
  border-color: #cbd5e1 !important;
  color: #475569 !important;
}

:root[data-theme="light"] .btn.secondary:hover,
[data-theme="light"] .btn.secondary:hover {
  background: #e2e8f0 !important;
  color: #0f172a !important;
}

.max-w-150 {
  max-width: 150px;
}

.text-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field-label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.agent-hint-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.7rem;
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  padding: 1px 6px;
  border-radius: 3px;
  font-weight: 500;
}

.field-helper-text {
  font-size: 0.72rem;
  color: var(--text-muted);
  line-height: 1.3;
}
</style>
