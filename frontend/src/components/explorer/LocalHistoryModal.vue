<template>
  <div v-if="activeHistoryModal.visible" class="local-history-backdrop" @click="closeModal">
    <div class="local-history-modal" @click.stop>
      <div class="modal-header">
        <div class="header-title">
          <span class="codicon codicon-history header-icon"></span>
          <span>本地历史快照 (Local History)</span>
          <span class="file-badge">{{ fileName }}</span>
        </div>
        <button class="close-btn" title="关闭" @click="closeModal">
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="snapshots.length === 0" class="empty-state">
          <span class="codicon codicon-history empty-icon"></span>
          <p>暂无本地历史快照</p>
          <span class="empty-sub">每次手动或自动保存该文件时，系统将自动记录带时间戳的安全快照（最多保留 20 份）</span>
        </div>

        <div v-else class="snapshot-list">
          <div v-for="(item, index) in snapshots" :key="item.id" class="snapshot-item">
            <div class="snapshot-info">
              <div class="snapshot-time-row">
                <span class="snapshot-index">#{{ snapshots.length - index }}</span>
                <span class="snapshot-time">{{ formatTime(item.timestamp) }}</span>
                <span class="snapshot-relative">({{ formatRelativeTime(item.timestamp) }})</span>
                <span v-if="index === 0" class="latest-tag">最新快照</span>
              </div>
              <div class="snapshot-meta-row">
                <span class="snapshot-size">{{ formatSize(item.size) }}</span>
                <span v-if="item.label" class="snapshot-label">{{ item.label }}</span>
              </div>
            </div>

            <div class="snapshot-actions">
              <button class="action-btn diff-btn" title="在主编辑器中与当前版本差异比对" @click="handleCompare(item)">
                <span class="codicon codicon-diff"></span>
                对比差异
              </button>
              <button
                class="action-btn restore-btn"
                title="将当前工作区文件内容还原到此快照版本"
                @click="handleRestore(item)"
              >
                <span class="codicon codicon-discard"></span>
                还原到此版
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <span class="footer-tip">💡 本地历史独立于 Git 仓库，可挽救任何未提交的草稿改动</span>
        <button class="footer-btn" @click="closeModal">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useLocalHistory, type HistorySnapshot } from "../../composables/useLocalHistory";
import { useFileTree } from "../../composables/useFileTree";
import { useToast } from "../../composables/useToast";

const { activeHistoryModal, getSnapshots, closeLocalHistory, compareWithCurrent } = useLocalHistory();
const { activeFileContent, saveFile } = useFileTree();
const { showToast } = useToast();

const currentPath = computed(() => activeHistoryModal.value.path);
const fileName = computed(() => currentPath.value.split("/").pop() || currentPath.value);
const snapshots = computed(() => getSnapshots(currentPath.value));

function closeModal(): void {
  closeLocalHistory();
}

function handleCompare(snapshot: HistorySnapshot): void {
  compareWithCurrent(snapshot, activeFileContent.value);
  closeModal();
}

async function handleRestore(snapshot: HistorySnapshot): Promise<void> {
  const ok = await saveFile(snapshot.path, snapshot.content);
  if (ok) {
    showToast(`已成功还原至 ${formatTime(snapshot.timestamp)} 快照`, "success");
    closeModal();
  }
}

function formatTime(timestamp: number): string {
  const d = new Date(timestamp);
  return d.toLocaleString("zh-CN", {
    hour12: false,
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatRelativeTime(timestamp: number): string {
  const diffSec = Math.floor((Date.now() - timestamp) / 1000);
  if (diffSec < 10) return "刚刚";
  if (diffSec < 60) return `${diffSec}秒前`;
  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) return `${diffMin}分钟前`;
  const diffHour = Math.floor(diffMin / 60);
  if (diffHour < 24) return `${diffHour}小时前`;
  return `${Math.floor(diffHour / 24)}天前`;
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  return `${(bytes / 1024).toFixed(1)} KB`;
}
</script>

<style scoped>
.local-history-backdrop {
  position: fixed;
  inset: 0;
  z-index: 999;
  background-color: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.local-history-modal {
  width: 90%;
  max-width: 620px;
  max-height: 80vh;
  background-color: var(--card-bg, #1e293b);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  border-radius: 8px;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

:root[data-theme="light"] .local-history-modal {
  background-color: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.15);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
  color: var(--text-main);
}

.header-icon {
  color: #38bdf8;
  font-size: 16px;
}

.file-badge {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 12px;
  background-color: rgba(56, 189, 248, 0.14);
  color: #38bdf8;
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
  background-color: rgba(128, 128, 128, 0.18);
  color: var(--text-main);
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  color: var(--text-muted);
  text-align: center;
}

.empty-icon {
  font-size: 36px;
  margin-bottom: 10px;
  opacity: 0.6;
}

.empty-sub {
  font-size: 12px;
  margin-top: 6px;
  max-width: 360px;
}

.snapshot-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.snapshot-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 6px;
  background-color: rgba(128, 128, 128, 0.06);
  border: 1px solid rgba(128, 128, 128, 0.12);
  transition: all 0.15s ease;
}

.snapshot-item:hover {
  background-color: rgba(128, 128, 128, 0.12);
  border-color: rgba(56, 189, 248, 0.3);
}

.snapshot-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.snapshot-time-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main);
}

.snapshot-index {
  color: var(--text-muted);
  font-size: 11px;
}

.snapshot-relative {
  font-size: 12px;
  color: var(--text-muted);
}

.latest-tag {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 4px;
  background-color: rgba(16, 185, 129, 0.15);
  color: #10b981;
  font-weight: 600;
}

.snapshot-meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: var(--text-muted);
}

.snapshot-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  border: none;
  border-radius: 4px;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.15s ease;
}

.action-btn.diff-btn {
  background-color: rgba(56, 189, 248, 0.14);
  color: #38bdf8;
}

.action-btn.diff-btn:hover {
  background-color: rgba(56, 189, 248, 0.25);
}

.action-btn.restore-btn {
  background-color: rgba(245, 158, 11, 0.14);
  color: #f59e0b;
}

.action-btn.restore-btn:hover {
  background-color: rgba(245, 158, 11, 0.25);
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-top: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  font-size: 11px;
  color: var(--text-muted);
}

.footer-btn {
  background-color: rgba(128, 128, 128, 0.15);
  border: none;
  border-radius: 4px;
  padding: 4px 12px;
  font-size: 12px;
  color: var(--text-main);
  cursor: pointer;
}

.footer-btn:hover {
  background-color: rgba(128, 128, 128, 0.25);
}
</style>
