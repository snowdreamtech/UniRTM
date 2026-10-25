<template>
  <div class="git-identity-settings" id="git-identity-settings-panel">
    <!-- Header -->
    <div class="settings-section-header">
      <div class="section-title">
        <span class="codicon codicon-shield"></span>
        <h3>{{ t("git_identity.title") || "Git Identity & Credential Vault" }}</h3>
      </div>
      <p class="section-desc">
        {{
          t("git_identity.desc") ||
          "Manage multiple Git identities locally isolated per repository. Zero private key touching, system agent first, and hardened one-time nonce IPC."
        }}
      </p>
    </div>

    <!-- Security Agents Status Cards -->
    <div class="agent-status-row">
      <div class="status-card" :class="{ active: agentStatus.sshAgentActive }">
        <div class="card-header">
          <span class="codicon codicon-key"></span>
          <span class="card-title">SSH Agent</span>
          <span class="status-indicator" :class="agentStatus.sshAgentActive ? 'online' : 'offline'">
            {{
              agentStatus.sshAgentActive
                ? t("git_identity.active") || "Active"
                : t("git_identity.inactive") || "Inactive"
            }}
          </span>
        </div>
        <div class="card-body">
          <span v-if="agentStatus.sshAgentActive">
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
          <span class="codicon codicon-lock"></span>
          <span class="card-title">GPG Agent</span>
          <span class="status-indicator" :class="agentStatus.gpgAgentActive ? 'online' : 'offline'">
            {{
              agentStatus.gpgAgentActive
                ? t("git_identity.active") || "Active"
                : t("git_identity.inactive") || "Inactive"
            }}
          </span>
        </div>
        <div class="card-body">
          <span v-if="agentStatus.gpgAgentActive">{{
            t("git_identity.gpg_ready") || "gpg-agent running and responsive"
          }}</span>
          <span v-else class="text-muted">{{ t("git_identity.gpg_not_found") || "gpg-agent not responding" }}</span>
        </div>
      </div>

      <div class="status-card vault-control">
        <div class="card-header">
          <span class="codicon codicon-clear-all"></span>
          <span class="card-title">{{ t("git_identity.memory_vault") || "Session Vault" }}</span>
        </div>
        <div class="card-body">
          <button class="clear-vault-btn" @click="handleClearKeyring">
            <span class="codicon codicon-trash"></span>
            <span>{{ t("git_identity.clear_keyring") || "Purge In-Memory Keyring" }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Profiles List -->
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
              <span v-if="prof.id === activeProfileId" class="badge active-badge">{{
                t("git_identity.active_workspace") || "Active in Workspace"
              }}</span>
              <span v-if="prof.isDefault" class="badge default-badge">{{
                t("git_identity.default") || "Default"
              }}</span>
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

    <!-- Edit / Add Modal -->
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

    <!-- Security Audit Trail Accordion -->
    <div class="audit-trail-section">
      <div class="section-toolbar pointer" @click="isAuditExpanded = !isAuditExpanded">
        <div class="title-with-icon">
          <span class="codicon" :class="isAuditExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"></span>
          <h4>{{ t("git_identity.audit_trail") || "Security & Credential Audit Trail" }}</h4>
          <span class="audit-count">({{ auditRecords.length }})</span>
        </div>
        <button class="refresh-btn" @click.stop="loadAuditRecords">
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
  gap: 1.2rem;
  padding: 1rem 0;
}

.settings-section-header {
  border-bottom: 1px solid var(--vscode-panel-border, #333);
  padding-bottom: 0.8rem;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.1rem;
  color: var(--vscode-foreground, #fff);
}

.section-desc {
  font-size: 0.82rem;
  color: var(--vscode-descriptionForeground, #888);
  margin-top: 4px;
}

.agent-status-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 10px;
}

.status-card {
  background: var(--vscode-editor-background, #1e1e1e);
  border: 1px solid var(--vscode-panel-border, #333);
  border-radius: 4px;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.status-card.active {
  border-color: #2e6930;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
}

.status-indicator {
  margin-left: auto;
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 2px;
  text-transform: uppercase;
}

.status-indicator.online {
  background: #235a25;
  color: #7ee787;
}

.status-indicator.offline {
  background: #3e3e3e;
  color: #888;
}

.card-body {
  font-size: 11px;
  color: var(--vscode-descriptionForeground, #aaa);
}

.clear-vault-btn {
  background: transparent;
  border: 1px solid var(--vscode-button-secondaryBorder, #444);
  color: var(--vscode-foreground, #ccc);
  padding: 3px 8px;
  font-size: 11px;
  border-radius: 3px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
}

.clear-vault-btn:hover {
  background: var(--vscode-button-secondaryHoverBackground, rgba(255, 255, 255, 0.08));
}

.section-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.section-toolbar.pointer {
  cursor: pointer;
  user-select: none;
}

.title-with-icon {
  display: flex;
  align-items: center;
  gap: 6px;
}

.audit-count {
  font-size: 11px;
  color: var(--vscode-descriptionForeground, #888);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--vscode-button-background, #0e639c);
  color: #fff;
  border: none;
  padding: 4px 10px;
  border-radius: 3px;
  font-size: 11px;
  cursor: pointer;
}

.action-btn:hover {
  background: var(--vscode-button-hoverBackground, #1177bb);
}

.profiles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 10px;
}

.profile-card {
  background: var(--vscode-editor-background, #1e1e1e);
  border: 1px solid var(--vscode-panel-border, #333);
  border-radius: 4px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.profile-card.current {
  border-color: #0e639c;
  background: rgba(14, 99, 156, 0.08);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.profile-name-row .name {
  font-size: 13px;
  font-weight: 600;
}

.badge {
  font-size: 9px;
  padding: 1px 4px;
  border-radius: 2px;
  text-transform: uppercase;
}

.active-badge {
  background: #0e639c;
  color: #fff;
}

.default-badge {
  background: #3e3e3e;
  color: #ccc;
}

.card-actions {
  display: flex;
  gap: 4px;
}

.icon-btn {
  background: transparent;
  border: none;
  color: var(--vscode-foreground, #aaa);
  padding: 2px 4px;
  cursor: pointer;
  border-radius: 3px;
}

.icon-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.apply-btn {
  color: #7ee787;
}

.delete-btn:hover {
  color: #f85149;
}

.card-details {
  display: flex;
  flex-direction: column;
  gap: 3px;
  font-size: 11px;
}

.detail-row {
  display: flex;
  gap: 6px;
}

.detail-row .label {
  color: var(--vscode-descriptionForeground, #888);
  width: 32px;
}

.text-gpg {
  color: #4ec9b0;
}

/* Modal */
.profile-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.profile-modal {
  width: 480px;
  background: var(--vscode-editor-background, #252526);
  border: 1px solid var(--vscode-dropdown-border, #454545);
  border-radius: 6px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 10px 14px;
  border-bottom: 1px solid var(--vscode-panel-border, #333);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-header h3 {
  font-size: 13px;
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--vscode-foreground, #888);
  cursor: pointer;
}

.modal-body {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-group label {
  font-size: 11px;
  color: var(--vscode-foreground, #ccc);
}

.form-row {
  display: flex;
  gap: 8px;
}

.flex-1 {
  flex: 1;
}

.input-text {
  background: var(--vscode-input-background, #3c3c3c);
  border: 1px solid var(--vscode-input-border, #555);
  color: var(--vscode-input-foreground, #fff);
  padding: 5px 8px;
  font-size: 12px;
  border-radius: 3px;
  outline: none;
}

.input-with-select {
  display: flex;
  gap: 6px;
}

.input-with-select .input-text {
  flex: 1;
}

.select-discovered {
  background: var(--vscode-dropdown-background, #3c3c3c);
  border: 1px solid var(--vscode-dropdown-border, #555);
  color: var(--vscode-dropdown-foreground, #fff);
  font-size: 11px;
  border-radius: 3px;
  max-width: 170px;
}

.check-row {
  flex-direction: column;
  gap: 6px;
  margin-top: 4px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  cursor: pointer;
}

.modal-footer {
  padding: 10px 14px;
  border-top: 1px solid var(--vscode-panel-border, #333);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn {
  padding: 5px 12px;
  font-size: 12px;
  border-radius: 3px;
  cursor: pointer;
  border: none;
}

.btn.primary {
  background: var(--vscode-button-background, #0e639c);
  color: #fff;
}

.btn.secondary {
  background: var(--vscode-button-secondaryBackground, #3a3d41);
  color: #fff;
}

/* Audit table */
.audit-trail-section {
  background: var(--vscode-editor-background, #1e1e1e);
  border: 1px solid var(--vscode-panel-border, #333);
  border-radius: 4px;
  padding: 10px;
}

.audit-table-wrapper {
  margin-top: 8px;
  overflow-x: auto;
}

.audit-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 11px;
}

.audit-table th,
.audit-table td {
  padding: 5px 8px;
  text-align: left;
  border-bottom: 1px solid var(--vscode-panel-border, #2a2a2a);
}

.action-tag {
  background: #333;
  padding: 1px 4px;
  border-radius: 2px;
}

.type-tag.ssh {
  color: #79c0ff;
}

.type-tag.gpg {
  color: #7ee787;
}

.type-tag.security_violation {
  color: #f85149;
  font-weight: bold;
}

.result-badge.success {
  color: #7ee787;
}

.result-badge.denied_invalid_nonce {
  color: #f85149;
  font-weight: 600;
}

.empty-audit {
  padding: 12px;
  text-align: center;
  font-size: 11px;
  color: var(--vscode-descriptionForeground, #777);
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
