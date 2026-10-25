<template>
  <div class="git-profile-status-wrapper" ref="wrapperRef">
    <div class="status-pill status-git-profile" id="status-git-profile" :title="tooltipText" @click="toggleDropdown">
      <span class="status-icon codicon codicon-shield"></span>
      <span class="status-profile-name font-mono">{{ activeProfileName }}</span>
      <span v-if="activeProfile?.signingKey" class="gpg-badge" title="GPG Signing Enabled">
        <span class="codicon codicon-key"></span>
      </span>
    </div>

    <!-- QuickPick Dropdown -->
    <div v-if="isOpen" class="git-profile-dropdown" id="git-profile-quickpick">
      <div class="dropdown-header">
        <span class="header-title">{{ t("git_identity.switch_profile") || "Switch Git Identity" }}</span>
        <span class="header-hint">{{
          t("git_identity.local_isolation_hint") || "Applied locally to .git/config"
        }}</span>
      </div>

      <div class="profile-list">
        <div
          v-for="prof in profiles"
          :key="prof.id"
          class="profile-item"
          :class="{ active: prof.id === activeProfileId }"
          @click="selectProfile(prof.id)"
        >
          <div class="profile-info">
            <div class="profile-title">
              <span class="profile-name">{{ prof.name }}</span>
              <span v-if="prof.id === activeProfileId" class="active-badge">{{
                t("git_identity.active") || "Active"
              }}</span>
            </div>
            <div class="profile-details">
              <span>{{ prof.userName || "Anonymous" }} &lt;{{ prof.userEmail || "no-email" }}&gt;</span>
              <span v-if="prof.signingKey" class="detail-gpg">GPG: {{ prof.signingKey.slice(-8) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="dropdown-footer">
        <button class="manage-btn" @click="openSettings">
          <span class="codicon codicon-settings"></span>
          <span>{{ t("git_identity.manage_profiles") || "Manage Git Profiles & Vault..." }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useGitIdentity } from "../../composables/useGitIdentity";
import { t } from "../../i18n";

const emit = defineEmits<{
  (e: "open-settings", tab?: string): void;
}>();

const { profiles, activeProfileId, activeProfile, loadProfiles, applyProfile } = useGitIdentity();

const isOpen = ref(false);
const wrapperRef = ref<HTMLElement | null>(null);

const activeProfileName = computed(() => {
  if (activeProfile.value) {
    return activeProfile.value.name;
  }
  return t("git_identity.default") || "Default Profile";
});

const tooltipText = computed(() => {
  if (!activeProfile.value) {
    return t("git_identity.click_to_switch") || "Git Identity: Default (Click to switch)";
  }
  const email = activeProfile.value.userEmail ? ` <${activeProfile.value.userEmail}>` : "";
  const sign = activeProfile.value.signingKey ? " • GPG Signed" : "";
  return `Git Identity: ${activeProfile.value.name} (${activeProfile.value.userName}${email})${sign} • Click to switch`;
});

const toggleDropdown = () => {
  isOpen.value = !isOpen.value;
};

const selectProfile = async (id: string) => {
  await applyProfile(id);
  isOpen.value = false;
};

const openSettings = () => {
  isOpen.value = false;
  emit("open-settings", "git-identity");
};

const handleClickOutside = (e: MouseEvent) => {
  if (wrapperRef.value && !wrapperRef.value.contains(e.target as Node)) {
    isOpen.value = false;
  }
};

onMounted(() => {
  loadProfiles();
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>

<style scoped>
.git-profile-status-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.status-git-profile {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  padding: 0 6px;
  height: 22px;
  border-radius: 3px;
  user-select: none;
  font-size: 11px;
  transition: background 0.15s ease;
}

.status-git-profile:hover {
  background: var(--vscode-statusBarItem-hoverBackground, rgba(255, 255, 255, 0.12));
}

.gpg-badge {
  color: #4ec9b0;
  display: inline-flex;
  align-items: center;
}

.git-profile-dropdown {
  position: absolute;
  bottom: 26px;
  left: 0;
  min-width: 280px;
  background: var(--vscode-dropdown-background, #252526);
  border: 1px solid var(--vscode-dropdown-border, #3c3c3c);
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.35);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dropdown-header {
  padding: 8px 10px;
  border-bottom: 1px solid var(--vscode-panel-border, #333);
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.header-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--vscode-foreground, #ccc);
}

.header-hint {
  font-size: 10px;
  color: var(--vscode-descriptionForeground, #888);
}

.profile-list {
  max-height: 200px;
  overflow-y: auto;
  padding: 4px 0;
}

.profile-item {
  padding: 6px 10px;
  cursor: pointer;
  transition: background 0.12s ease;
}

.profile-item:hover {
  background: var(--vscode-list-hoverBackground, rgba(255, 255, 255, 0.08));
}

.profile-item.active {
  background: var(--vscode-list-activeSelectionBackground, #094771);
  color: #fff;
}

.profile-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.profile-name {
  font-size: 12px;
  font-weight: 500;
}

.active-badge {
  font-size: 9px;
  padding: 1px 4px;
  background: #388a34;
  color: #fff;
  border-radius: 2px;
  text-transform: uppercase;
}

.profile-details {
  font-size: 10px;
  color: var(--vscode-descriptionForeground, #aaa);
  display: flex;
  justify-content: space-between;
  margin-top: 2px;
}

.detail-gpg {
  color: #4ec9b0;
  font-family: monospace;
}

.dropdown-footer {
  padding: 6px 10px;
  border-top: 1px solid var(--vscode-panel-border, #333);
  background: rgba(0, 0, 0, 0.15);
}

.manage-btn {
  width: 100%;
  background: transparent;
  border: none;
  color: var(--vscode-textLink-foreground, #3794ff);
  font-size: 11px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
}

.manage-btn:hover {
  text-decoration: underline;
}
</style>
