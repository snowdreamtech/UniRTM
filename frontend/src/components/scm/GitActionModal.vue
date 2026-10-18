<template>
  <div v-if="visible" class="modal-backdrop" @click.self="handleCancel">
    <div class="modal-dialog" role="dialog" :aria-labelledby="titleId">
      <div class="modal-header">
        <div class="modal-title-group">
          <span class="codicon" :class="modalIcon"></span>
          <h3 :id="titleId" class="modal-title">{{ modalTitle }}</h3>
        </div>
        <button class="modal-close-btn" title="关闭" @click="handleCancel">
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <div class="modal-body">
        <!-- Target Commit Info Header -->
        <div class="target-commit-pill">
          <span class="codicon codicon-git-commit"></span>
          <span class="commit-short font-mono">{{ targetCommit?.shortHash }}</span>
          <span class="commit-msg-preview">{{ targetCommit?.message }}</span>
        </div>

        <!-- 1. Create Tag Form -->
        <div v-if="actionType === 'create-tag'" class="form-group">
          <label class="form-label" for="input-tag-name">标签名称 (Tag Name) <span class="required">*</span></label>
          <input
            id="input-tag-name"
            v-model="tagName"
            type="text"
            class="form-input"
            placeholder="例如: v1.0.0, release-20261018"
            autofocus
            @keydown.enter="handleConfirm"
          />

          <label class="form-label mt-3" for="input-tag-msg">说明备注 (Optional Message)</label>
          <textarea
            id="input-tag-msg"
            v-model="tagMessage"
            class="form-textarea"
            rows="2"
            placeholder="附注标签说明信息 (可选)"
          ></textarea>
        </div>

        <!-- 2. Create Branch Form -->
        <div v-else-if="actionType === 'create-branch'" class="form-group">
          <label class="form-label" for="input-branch-name">新分支名称 (Branch Name) <span class="required">*</span></label>
          <input
            id="input-branch-name"
            v-model="branchName"
            type="text"
            class="form-input"
            placeholder="例如: feature/awesome-feature, bugfix/scm-fix"
            autofocus
            @keydown.enter="handleConfirm"
          />

          <label class="checkbox-label mt-3">
            <input v-model="checkoutBranch" type="checkbox" />
            <span>创建后立即检出并切换到该分支 (Checkout Branch)</span>
          </label>
        </div>

        <!-- 3. Reset To Commit Form -->
        <div v-else-if="actionType === 'reset'" class="form-group">
          <p class="form-hint">请选择重置模式。将当前分支的 HEAD 移动到目标提交：</p>
          <div class="radio-options">
            <label class="radio-card" :class="{ selected: resetMode === 'soft' }">
              <input v-model="resetMode" type="radio" value="soft" />
              <div class="radio-card-content">
                <span class="radio-title">Soft (软重置 --soft)</span>
                <span class="radio-desc">保留工作区与暂存区的所有更改，仅撤销该提交之后的 Commit 记录。</span>
              </div>
            </label>

            <label class="radio-card" :class="{ selected: resetMode === 'mixed' }">
              <input v-model="resetMode" type="radio" value="mixed" />
              <div class="radio-card-content">
                <span class="radio-title">Mixed (默认混合模式 --mixed)</span>
                <span class="radio-desc">保留工作区修改，但重置暂存区（撤销 git add）。</span>
              </div>
            </label>

            <label class="radio-card danger" :class="{ selected: resetMode === 'hard' }">
              <input v-model="resetMode" type="radio" value="hard" />
              <div class="radio-card-content">
                <span class="radio-title">Hard (硬重置 --hard - 危险)</span>
                <span class="radio-desc">丢弃所有自该提交以来的工作区和暂存区更改！</span>
              </div>
            </label>
          </div>

          <div v-if="resetMode === 'hard'" class="danger-alert">
            <span class="codicon codicon-warning"></span>
            <span>注意：硬重置将不可恢复地丢弃自此提交之后的所有未提交修改！</span>
          </div>
        </div>

        <!-- 4. Revert Confirm -->
        <div v-else-if="actionType === 'revert'" class="confirm-message">
          <p>确定要撤销 (Revert) 此提交的更改吗？</p>
          <p class="text-muted text-sm mt-1">Git 将创建一个新的反向提交来撤销此次变更引入的代码。</p>
        </div>

        <!-- 5. Cherry Pick Confirm -->
        <div v-else-if="actionType === 'cherry-pick'" class="confirm-message">
          <p>确定将此提交拣选 (Cherry-pick) 应用到当前分支吗？</p>
          <p class="text-muted text-sm mt-1">该提交引入的变动将被应用并作为新提交追加在当前分支顶部。</p>
        </div>

        <!-- 6. Checkout Confirm -->
        <div v-else-if="actionType === 'checkout'" class="confirm-message">
          <p>确定要检出 (Checkout) 提交 <span class="font-mono">{{ targetCommit?.shortHash }}</span> 吗？</p>
          <p class="text-muted text-sm mt-1">检出特定 Commit 会将 HEAD 处于分离状态 (Detached HEAD)。</p>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="handleCancel">取消</button>
        <button
          class="btn"
          :class="isDangerAction ? 'btn-danger' : 'btn-primary'"
          :disabled="isConfirmDisabled || isSubmitting"
          @click="handleConfirm"
        >
          <span v-if="isSubmitting" class="codicon codicon-loading codicon-modifier-spin"></span>
          <span>{{ confirmButtonText }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import type { GitGraphCommit } from "../../composables/useGit";

export type GitModalActionType =
  | "create-tag"
  | "create-branch"
  | "reset"
  | "revert"
  | "cherry-pick"
  | "checkout";

const props = defineProps<{
  visible: boolean;
  actionType: GitModalActionType | null;
  targetCommit: GitGraphCommit | null;
}>();

const emit = defineEmits<{
  (e: "cancel"): void;
  (
    e: "confirm",
    payload: {
      action: GitModalActionType;
      commit: GitGraphCommit;
      data?: any;
    }
  ): void;
}>();

const titleId = "git-action-modal-title";
const tagName = ref("");
const tagMessage = ref("");
const branchName = ref("");
const checkoutBranch = ref(true);
const resetMode = ref<"soft" | "mixed" | "hard">("mixed");
const isSubmitting = ref(false);

watch(
  () => props.visible,
  (val) => {
    if (val) {
      tagName.value = "";
      tagMessage.value = "";
      branchName.value = "";
      checkoutBranch.value = true;
      resetMode.value = "mixed";
      isSubmitting.value = false;
    }
  }
);

const modalTitle = computed(() => {
  switch (props.actionType) {
    case "create-tag":
      return "为提交创建标签 (Create Tag)";
    case "create-branch":
      return "从此提交新建分支 (Create Branch)";
    case "reset":
      return "重置分支到此提交 (Reset Branch)";
    case "revert":
      return "撤销此提交 (Revert Commit)";
    case "cherry-pick":
      return "拣选此提交 (Cherry-Pick)";
    case "checkout":
      return "检出此提交 (Checkout Commit)";
    default:
      return "Git 操作";
  }
});

const modalIcon = computed(() => {
  switch (props.actionType) {
    case "create-tag":
      return "codicon-tag";
    case "create-branch":
      return "codicon-git-branch";
    case "reset":
      return "codicon-discard";
    case "revert":
      return "codicon-history";
    case "cherry-pick":
      return "codicon-git-pull-request";
    case "checkout":
      return "codicon-check";
    default:
      return "codicon-git-commit";
  }
});

const isDangerAction = computed(() => {
  return props.actionType === "reset" && resetMode.value === "hard";
});

const isConfirmDisabled = computed(() => {
  if (props.actionType === "create-tag") {
    return !tagName.value.trim();
  }
  if (props.actionType === "create-branch") {
    return !branchName.value.trim();
  }
  return false;
});

const confirmButtonText = computed(() => {
  switch (props.actionType) {
    case "create-tag":
      return "创建标签";
    case "create-branch":
      return "创建分支";
    case "reset":
      return resetMode.value === "hard" ? "强制重置 (丢弃修改)" : "确认重置";
    case "revert":
      return "确认撤销";
    case "cherry-pick":
      return "执行拣选";
    case "checkout":
      return "检出提交";
    default:
      return "确认";
  }
});

function handleCancel() {
  emit("cancel");
}

function handleConfirm() {
  if (isConfirmDisabled.value || !props.targetCommit || !props.actionType) return;
  isSubmitting.value = true;

  let data: any = undefined;
  if (props.actionType === "create-tag") {
    data = { name: tagName.value.trim(), message: tagMessage.value.trim() };
  } else if (props.actionType === "create-branch") {
    data = { name: branchName.value.trim(), checkout: checkoutBranch.value };
  } else if (props.actionType === "reset") {
    data = { mode: resetMode.value };
  }

  emit("confirm", {
    action: props.actionType,
    commit: props.targetCommit,
    data,
  });
}
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.15s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.modal-dialog {
  background: var(--card-bg, #182234);
  border: 1px solid var(--card-border, rgba(0, 229, 255, 0.2));
  border-radius: 8px;
  width: 90%;
  max-width: 520px;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--text-main);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--card-border);
  background: var(--section-bg, rgba(255, 255, 255, 0.02));
}

.modal-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-main);
}

.modal-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-close-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.modal-body {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.target-commit-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--input-bg, rgba(0, 0, 0, 0.2));
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12px;
}

.commit-short {
  color: var(--accent-cyan, #00e5ff);
  font-weight: 600;
}

.commit-msg-preview {
  flex: 1;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-main);
}

.required {
  color: var(--error, #ef4444);
}

.form-input,
.form-textarea {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s ease;
}

.form-input:focus,
.form-textarea:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 1px var(--accent-cyan);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-main);
  cursor: pointer;
}

.form-hint {
  font-size: 12px;
  color: var(--text-muted);
  margin: 0 0 6px 0;
}

.radio-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.radio-card {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid var(--card-border);
  border-radius: 6px;
  background: var(--input-bg);
  cursor: pointer;
  transition: all 0.15s ease;
}

.radio-card:hover {
  border-color: var(--accent-cyan);
}

.radio-card.selected {
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.05);
}

.radio-card.danger.selected {
  border-color: var(--error, #ef4444);
  background: rgba(239, 68, 68, 0.08);
}

.radio-card input {
  margin-top: 3px;
}

.radio-card-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.radio-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
}

.radio-desc {
  font-size: 11px;
  color: var(--text-muted);
  line-height: 1.4;
}

.danger-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 6px;
  color: var(--error, #ef4444);
  font-size: 12px;
}

.confirm-message {
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-main);
}

.mt-1 {
  margin-top: 4px;
}

.mt-3 {
  margin-top: 12px;
}

.text-muted {
  color: var(--text-muted);
}

.text-sm {
  font-size: 12px;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 18px;
  background: var(--section-bg, rgba(255, 255, 255, 0.02));
  border-top: 1px solid var(--card-border);
}

.btn {
  padding: 6px 16px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid transparent;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s ease;
}

.btn-secondary {
  background: transparent;
  border-color: var(--card-border);
  color: var(--text-main);
}

.btn-secondary:hover {
  background: rgba(128, 128, 128, 0.15);
}

.btn-primary {
  background: var(--accent-cyan);
  color: #070a12;
}

:root[data-theme="light"] .btn-primary {
  color: #ffffff;
}

.btn-primary:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn-danger {
  background: var(--error, #ef4444);
  color: #ffffff;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
