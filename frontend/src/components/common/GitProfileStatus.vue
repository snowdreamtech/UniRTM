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
  color: var(--text-muted);
  transition: all 0.15s ease;
}

.status-git-profile:hover {
  color: var(--text-main);
  background: var(--hover-bg);
}

.gpg-badge {
  color: var(--accent-purple);
  display: inline-flex;
  align-items: center;
}

.git-profile-dropdown {
  position: absolute;
  bottom: 26px;
  left: 0;
  min-width: 280px;
  background: var(--modal-bg);
  backdrop-filter: blur(16px);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.22);
  color: var(--text-main);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dropdown-header {
  padding: 10px 12px;
  border-bottom: 1px solid var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: var(--header-bg);
}

.header-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-main);
}

.header-hint {
  font-size: 10px;
  color: var(--text-muted);
}

.profile-list {
  max-height: 200px;
  overflow-y: auto;
  padding: 4px 0;
}

.profile-item {
  padding: 7px 12px;
  cursor: pointer;
  transition: background 0.12s ease;
  color: var(--text-main);
}

.profile-item:hover {
  background: var(--hover-bg);
}

.profile-item.active {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
}

.profile-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.profile-name {
  font-size: 12px;
  font-weight: 600;
}

.active-badge {
  font-size: 9px;
  padding: 1px 5px;
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
  border-radius: 3px;
  font-weight: 700;
  text-transform: uppercase;
}

.profile-details {
  font-size: 10px;
  color: var(--text-muted);
  display: flex;
  justify-content: space-between;
  margin-top: 3px;
}

.detail-gpg {
  color: var(--accent-purple);
  font-family: monospace;
}

.dropdown-footer {
  padding: 8px 12px;
  border-top: 1px solid var(--card-border);
  background: var(--header-bg);
}

.manage-btn {
  width: 100%;
  background: transparent;
  border: none;
  color: var(--accent-cyan);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 2px 0;
}

.manage-btn:hover {
  text-decoration: underline;
}
</style>
