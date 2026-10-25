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
              <span class="name">{{ prof.name }}</span>
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
              <span class="val font-mono">{{ prof.userName }} &lt;{{ prof.userEmail }}&gt;</span>
            </div>
            <div v-if="prof.sshKeyPath" class="detail-row">
              <span class="label">SSH:</span>
              <span class="val font-mono text-truncate">{{ prof.sshKeyPath }}</span>
            </div>
            <div v-if="prof.signingKey" class="detail-row">
              <span class="label">GPG:</span>
              <span class="val font-mono text-gpg">{{ prof.signingKey }} {{ prof.gpgSign ? "(Auto-sign)" : "" }}</span>
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
            <label>{{ t("git_identity.ssh_key") || "SSH Key (Local Path / Discovered)" }}</label>
            <div class="input-with-select">
              <input
                v-model="editingProfile.sshKeyPath"
                type="text"
                placeholder="e.g. ~/.ssh/id_ed25519"
                class="input-text"
              />
              <select v-if="systemKeys.sshKeys.length > 0" class="select-discovered" @change="onSelectDiscoveredSSH">
                <option value="">{{ t("git_identity.select_discovered") || "-- Select Discovered Key --" }}</option>
                <option v-for="k in systemKeys.sshKeys" :key="k.path" :value="k.path">
                  {{ k.name }} ({{ k.algorithm }})
                </option>
              </select>
            </div>
          </div>

          <!-- GPG Key Selection -->
          <div class="form-group">
            <label>{{ t("git_identity.gpg_signing_key") || "GPG Signing Key ID" }}</label>
            <div class="input-with-select">
              <input
                v-model="editingProfile.signingKey"
                type="text"
                placeholder="e.g. 3AA5C34371567BD2"
                class="input-text"
              />
              <select v-if="systemKeys.gpgKeys.length > 0" class="select-discovered" @change="onSelectDiscoveredGPG">
                <option value="">{{ t("git_identity.select_discovered") || "-- Select GPG Key --" }}</option>
                <option v-for="g in systemKeys.gpgKeys" :key="g.keyId" :value="g.keyId">
                  {{ g.keyId }} ({{ g.userId }})
                </option>
              </select>
            </div>
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
import { ref, onMounted } from "vue";
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

const editingProfile = ref<Partial<GitProfile>>({
  name: "",
  userName: "",
  userEmail: "",
  sshKeyPath: "",
  signingKey: "",
  gpgSign: false,
  isDefault: false,
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
  editingProfile.value = { ...prof };
  isEditing.value = true;
};

const handleSaveProfile = async () => {
  if (!editingProfile.value.name) return;
  await saveProfile(editingProfile.value);
  isEditing.value = false;
};

const onSelectDiscoveredSSH = (e: Event) => {
  const target = e.target as HTMLSelectElement;
  if (target.value) {
    editingProfile.value.sshKeyPath = target.value;
  }
};

const onSelectDiscoveredGPG = (e: Event) => {
  const target = e.target as HTMLSelectElement;
  if (target.value) {
    editingProfile.value.signingKey = target.value;
    editingProfile.value.gpgSign = true;
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
  background: var(--accent-cyan);
  color: #070a12;
  font-weight: 600;
  border: none;
  padding: 5px 12px;
  border-radius: 4px;
  font-size: 0.78rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn.add-btn:hover {
  filter: brightness(1.1);
  box-shadow: 0 0 10px var(--accent-cyan-glow);
}

.profiles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
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
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.profile-name-row .name {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-main);
}

.card-actions {
  display: flex;
  gap: 4px;
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
  gap: 4px;
  font-size: 0.76rem;
}

.detail-row {
  display: flex;
  gap: 8px;
}

.detail-row .label {
  color: var(--text-muted);
  width: 32px;
  flex-shrink: 0;
}

.detail-row .val {
  color: var(--text-main);
}

.text-gpg {
  color: var(--accent-purple);
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
  padding: 6px 14px;
  font-size: 0.8rem;
  border-radius: 4px;
  cursor: pointer;
  border: none;
  font-weight: 500;
}

.btn.primary {
  background: var(--accent-cyan);
  color: #070a12;
  font-weight: 600;
}

.btn.primary:hover {
  filter: brightness(1.1);
}

.btn.secondary {
  background: var(--btn-secondary-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
}

.btn.secondary:hover {
  background: var(--btn-secondary-hover-bg);
}

.max-w-150 {
  max-width: 150px;
}

.text-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
