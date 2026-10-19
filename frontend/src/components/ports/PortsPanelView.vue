<template>
  <div class="ports-panel-container">
    <!-- Top Filter & Action Bar -->
    <div class="ports-toolbar">
      <div class="toolbar-search">
        <span class="codicon codicon-search search-icon"></span>
        <input
          v-model="searchQuery"
          type="text"
          class="ports-search-input"
          :placeholder="t('ports.filter_placeholder') || '过滤端口、进程或标签...'"
        />
        <button
          v-if="searchQuery"
          class="search-clear-btn"
          @click="searchQuery = ''"
        >
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <div class="toolbar-actions">
        <!-- Add Port Button -->
        <button
          class="ports-tool-btn primary"
          @click="startInlineAdd"
          :title="t('ports.add_port') || '添加端口'"
        >
          <span class="codicon codicon-add"></span>
          <span>{{ t('ports.add_port') || '添加端口' }}</span>
        </button>

        <!-- Refresh Button -->
        <button
          class="ports-tool-btn"
          :class="{ loading: isLoading }"
          :title="t('common.refresh') || '刷新'"
          @click="handleRefresh"
        >
          <span
            class="codicon"
            :class="isLoading ? 'codicon-sync codicon-modifier-spin' : 'codicon-refresh'"
          ></span>
        </button>
      </div>
    </div>

    <!-- Ports Table Wrapper -->
    <div class="ports-table-wrapper" @click="closeContextMenu">
      <table class="ports-table">
        <thead>
          <tr>
            <th class="th-port">{{ t('ports.col_port') || '端口' }}</th>
            <th class="th-address">{{ t('ports.col_address') || '转发地址' }}</th>
            <th class="th-process">{{ t('ports.col_process') || '正在运行的进程' }}</th>
            <th class="th-source">{{ t('ports.col_source') || '源' }}</th>
            <th class="th-visibility">{{ t('ports.col_visibility') || '访问控制' }}</th>
            <th class="th-actions"></th>
          </tr>
        </thead>
        <tbody>
          <!-- Existing Ports Rows -->
          <tr
            v-for="item in filteredPorts"
            :key="item.id"
            class="port-row"
            :class="{
              active: selectedPortId === item.id,
              inactive: item.status === 'inactive'
            }"
            tabindex="0"
            @click="selectedPortId = item.id"
            @contextmenu.prevent="openContextMenu($event, item)"
            @keydown="handleRowKeyDown($event, item)"
          >
            <!-- 1. Port & Status & Label -->
            <td class="td-port">
              <div class="port-cell-content">
                <!-- Status Dot -->
                <span
                  class="port-status-dot"
                  :class="item.status === 'active' ? 'dot-active' : 'dot-inactive'"
                  :title="item.status === 'active' ? (t('ports.status_active') || '运行中') : (t('ports.status_inactive') || '未连接 / 离线')"
                ></span>

                <!-- Port Number -->
                <span class="port-number font-mono">{{ item.port }}</span>

                <!-- Local Port Mapping indication if remapped -->
                <span
                  v-if="item.localPort && item.localPort !== item.port"
                  class="port-mapping-pill font-mono"
                  :title="`远程端口 ${item.port} 映射到本地 ${item.localPort}`"
                >
                  &rarr; {{ item.localPort }}
                </span>

                <!-- Method B: HTTPS Badge -->
                <span
                  v-if="item.probe?.isHttps || item.protocol === 'https'"
                  class="protocol-badge https"
                  title="TLS/HTTPS 加密协议"
                >
                  <span class="codicon codicon-lock"></span>
                  <span>HTTPS</span>
                </span>

                <!-- Method C: Workspace Origin Pill -->
                <span
                  v-if="item.workspaceHint"
                  class="origin-pill"
                  :class="item.workspaceHint.serviceType"
                  :title="`工作区语义: ${item.workspaceHint.sourceFile} (${item.workspaceHint.framework || ''})`"
                >
                  {{ item.workspaceHint.sourceFile.includes('compose') ? 'compose' : 'workspace' }}
                </span>

                <!-- Label / Service Recognition Tag -->
                <template v-if="editingPortId === item.id">
                  <input
                    ref="editLabelInputRef"
                    v-model="editLabelValue"
                    class="inline-edit-input"
                    @keydown.enter="submitEditLabel(item)"
                    @keydown.esc="cancelEditLabel"
                    @blur="submitEditLabel(item)"
                  />
                </template>
                <template v-else>
                  <span
                    class="port-label"
                    :title="t('ports.double_click_to_edit') || '双击编辑标签'"
                    @dblclick="startEditLabel(item)"
                  >
                    {{ item.label }}
                  </span>
                </template>
              </div>
            </td>

            <!-- 2. Forwarded Address & Quick Actions -->
            <td class="td-address">
              <div class="address-cell-content">
                <a
                  class="address-link font-mono"
                  :href="item.previewUrl"
                  target="_blank"
                  @click.stop
                >
                  {{ item.forwardAddress }}
                </a>

                <!-- Hover Quick Actions (aligned with VS Code) -->
                <div class="row-hover-actions">
                  <button
                    class="icon-action-btn"
                    :title="t('ports.edit_label') || '修改标签'"
                    @click.stop="startEditLabel(item)"
                  >
                    <span class="codicon codicon-tag"></span>
                  </button>
                  <button
                    class="icon-action-btn"
                    :title="t('ports.open_browser') || '在浏览器中打开'"
                    @click.stop="openPort(item)"
                  >
                    <span class="codicon codicon-globe"></span>
                  </button>
                  <button
                    class="icon-action-btn"
                    :title="t('ports.copy_address') || '复制转发地址'"
                    @click.stop="copyAddress(item)"
                  >
                    <span class="codicon codicon-copy"></span>
                  </button>
                  <button
                    v-if="item.status === 'active'"
                    class="icon-action-btn"
                    :title="t('ports.stop_forwarding') || '停止转发'"
                    @click.stop="stopPort(item.port)"
                  >
                    <span class="codicon codicon-stop-circle"></span>
                  </button>
                  <button
                    v-else
                    class="icon-action-btn"
                    :title="t('ports.start_forwarding') || '恢复转发'"
                    @click.stop="startPort(item.port)"
                  >
                    <span class="codicon codicon-play-circle"></span>
                  </button>
                  <button
                    class="icon-action-btn danger"
                    :title="t('ports.remove_port') || '删除端口'"
                    @click.stop="removePort(item.port)"
                  >
                    <span class="codicon codicon-close"></span>
                  </button>
                </div>
              </div>
            </td>

            <!-- 3. Running Process & Command line -->
            <td class="td-process">
              <div class="process-cell-wrapper">
                <div
                  class="process-text font-mono"
                  :class="{ 'is-empty': !item.commandLine && !item.process }"
                  :title="item.commandLine || (item.pid ? `PID: ${item.pid}` : '')"
                >
                  {{ item.commandLine || item.process || (item.workspaceHint ? (t('ports.expected_service') || '预期工作区服务') : (t('ports.process_unavailable') || '流程信息不可用')) }}
                </div>
                <!-- Method B Probe Banner if detected -->
                <div
                  v-if="item.probe?.detectedService"
                  class="probe-banner"
                  :title="`探针探测服务: ${item.probe.detectedService} (耗时 ${item.probe.probeDurationMs || 0}ms)`"
                >
                  <span class="codicon codicon-pulse"></span>
                  <span class="probe-text">{{ item.probe.detectedService }}</span>
                </div>
              </div>
            </td>

            <!-- 4. Source Badge (Auto vs User) -->
            <td class="td-source">
              <span class="source-badge" :class="'source-' + item.source">
                {{ item.source === 'auto' ? (t('ports.source_auto') || '自动转发') : (t('ports.source_user') || '用户转发') }}
              </span>
            </td>

            <!-- 5. Visibility (Private vs Public) -->
            <td class="td-visibility">
              <button
                class="visibility-toggle-btn"
                :class="'vis-' + item.visibility"
                :title="t('ports.toggle_visibility') || '点击切换公开/私有访问权限'"
                @click.stop="toggleVisibility(item)"
              >
                <span
                  class="codicon"
                  :class="item.visibility === 'public' ? 'codicon-globe' : 'codicon-lock'"
                ></span>
                <span>{{ item.visibility === 'public' ? (t('ports.vis_public') || '公开') : (t('ports.vis_private') || '私有') }}</span>
              </button>
            </td>

            <!-- 6. More Actions Column -->
            <td class="td-actions">
              <button
                class="more-action-btn"
                :title="t('ports.more_actions') || '更多操作'"
                @click.stop="openContextMenu($event, item)"
              >
                <span class="codicon codicon-ellipsis"></span>
              </button>
            </td>
          </tr>

          <!-- Inline Adding Row -->
          <tr v-if="isAddingInline" class="port-row adding-row">
            <td class="td-port">
              <div class="port-cell-content">
                <span class="port-status-dot dot-inactive"></span>
                <input
                  ref="newPortInputRef"
                  v-model.number="newPortNumber"
                  type="number"
                  class="inline-port-input font-mono"
                  placeholder="端口 (如 3000)"
                  min="1"
                  max="65535"
                  @keydown.enter="submitInlineAdd"
                  @keydown.esc="cancelInlineAdd"
                />
                <input
                  v-model="newPortLabel"
                  type="text"
                  class="inline-label-input"
                  placeholder="标签 (选填)"
                  @keydown.enter="submitInlineAdd"
                  @keydown.esc="cancelInlineAdd"
                />
              </div>
            </td>
            <td class="td-address">
              <span class="placeholder-text font-mono">localhost:{{ newPortNumber || '...' }}</span>
            </td>
            <td class="td-process">
              <span class="placeholder-text">-</span>
            </td>
            <td class="td-source">
              <span class="source-badge source-user">{{ t('ports.source_user') || '用户转发' }}</span>
            </td>
            <td class="td-visibility">
              <span class="source-badge vis-private">{{ t('ports.vis_private') || '私有' }}</span>
            </td>
            <td class="td-actions">
              <div class="inline-submit-group">
                <button
                  class="inline-btn confirm"
                  :title="t('common.confirm') || '确定'"
                  @click.stop="submitInlineAdd"
                >
                  <span class="codicon codicon-check"></span>
                </button>
                <button
                  class="inline-btn cancel"
                  :title="t('common.cancel') || '取消'"
                  @click.stop="cancelInlineAdd"
                >
                  <span class="codicon codicon-close"></span>
                </button>
              </div>
            </td>
          </tr>

          <!-- Empty State -->
          <tr v-if="filteredPorts.length === 0 && !isAddingInline" class="empty-row">
            <td colspan="6">
              <div class="ports-empty-state">
                <span class="codicon codicon-radio-tower empty-icon"></span>
                <div class="empty-title">{{ t('ports.no_forwarded_ports') || '暂无转发的端口' }}</div>
                <div class="empty-desc">
                  {{ t('ports.empty_hint') || '运行任何本地服务时将自动感知并呈现，您也可以手动添加端口。' }}
                </div>
                <button class="empty-add-btn" @click="startInlineAdd">
                  <span class="codicon codicon-add"></span>
                  <span>{{ t('ports.add_port') || '添加端口' }}</span>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Bottom Add Port Row Button (Aligned with VS Code screenshot) -->
      <div v-if="!isAddingInline" class="table-bottom-bar">
        <button class="bottom-add-btn" @click="startInlineAdd">
          <span class="codicon codicon-add"></span>
          <span>{{ t('ports.add_port') || '添加端口' }}</span>
        </button>
      </div>
    </div>

    <!-- Right Click Context Menu (Teleported, 100% aligned with VS Code) -->
    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="ports-context-menu"
        :style="{ left: contextMenuPos.x + 'px', top: contextMenuPos.y + 'px' }"
        @click.stop
      >
        <!-- 1. Open in Browser -->
        <div class="menu-item" @click="handleContextAction('open')">
          <span class="action-label">{{ t('ports.open_browser') || '在浏览器中打开' }}</span>
        </div>

        <!-- 2. Preview in Editor (Simple Browser) -->
        <div class="menu-item" @click="handleContextAction('preview')">
          <span class="action-label">{{ t('ports.preview_in_editor') || '在编辑器中预览' }}</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 3. Set Port Label -->
        <div class="menu-item" @click="handleContextAction('edit-label')">
          <span class="action-label">{{ t('ports.set_label') || '设置端口标签' }}</span>
          <span class="keybinding font-mono">↵</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 4. Copy Local Address -->
        <div class="menu-item" @click="handleContextAction('copy-address')">
          <span class="action-label">{{ t('ports.copy_local_address') || '复制本地地址' }}</span>
          <span class="keybinding font-mono">{{ isMac ? '⌘ C' : 'Ctrl+C' }}</span>
        </div>

        <!-- 5. Change Local Address Port -->
        <div class="menu-item" @click="handleContextAction('change-local-port')">
          <span class="action-label">{{ t('ports.change_local_port') || '更改本地地址端口' }}</span>
        </div>

        <!-- 6. Change Port Protocol > (Submenu trigger) -->
        <div
          class="menu-item has-submenu"
          :class="{ active: isProtocolSubmenuOpen }"
          @mouseenter="openProtocolSubmenu($event)"
          @mouseleave="scheduleCloseProtocolSubmenu"
          @click.stop="openProtocolSubmenu($event)"
        >
          <span class="action-label">{{ t('ports.change_protocol') || '更改端口协议' }}</span>
          <span class="submenu-chevron font-mono">›</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 7. Stop / Resume Forwarding Port -->
        <div class="menu-item" @click="handleContextAction('toggle-status')">
          <span class="action-label">
            {{ contextPort?.status === 'active' ? (t('ports.stop_forwarding_port') || '停止转发端口') : (t('ports.start_forwarding_port') || '转发端口') }}
          </span>
          <span class="keybinding font-mono">{{ isMac ? '⌘ ⌫' : 'Del' }}</span>
        </div>

        <!-- 8. Forward a Port -->
        <div class="menu-item" @click="handleContextAction('forward-port')">
          <span class="action-label">{{ t('ports.forward_port') || '转发端口' }}</span>
        </div>
      </div>
    </Teleport>

    <!-- 6b. Secondary Protocol Submenu (Teleported, 100% aligned with VS Code) -->
    <Teleport to="body">
      <div
        v-if="contextMenuVisible && isProtocolSubmenuOpen"
        class="ports-context-menu ports-protocol-submenu"
        :style="{ left: protocolSubmenuPos.x + 'px', top: protocolSubmenuPos.y + 'px' }"
        @mouseenter="cancelCloseProtocolSubmenu"
        @mouseleave="scheduleCloseProtocolSubmenu"
        @click.stop
      >
        <div class="menu-item check-item" @click="handleSelectProtocol('http')">
          <span class="check-mark font-mono">{{ (contextPort?.protocol === 'http' || !contextPort?.protocol) ? '✓' : '' }}</span>
          <span class="action-label">HTTP</span>
        </div>
        <div class="menu-item check-item" @click="handleSelectProtocol('https')">
          <span class="check-mark font-mono">{{ contextPort?.protocol === 'https' ? '✓' : '' }}</span>
          <span class="action-label">HTTPS</span>
        </div>
      </div>
    </Teleport>

    <!-- Modal: Simple Browser in-editor preview (VS Code aligned) -->
    <Teleport to="body">
      <div v-if="isPreviewModalOpen" class="simple-browser-backdrop" @click.self="isPreviewModalOpen = false">
        <div class="simple-browser-window">
          <div class="simple-browser-header">
            <div class="simple-browser-title-group">
              <span class="codicon codicon-globe browser-icon"></span>
              <span class="browser-title font-mono">{{ previewModalTitle }}</span>
            </div>
            <div class="simple-browser-address-bar">
              <button class="browser-btn" @click="reloadPreviewFrame" title="重新加载">
                <span class="codicon codicon-refresh"></span>
              </button>
              <input type="text" class="browser-url-input font-mono" :value="previewModalUrl" readonly />
              <button class="browser-btn" @click="openExternalFromPreview" title="在外部浏览器中打开">
                <span class="codicon codicon-link-external"></span>
              </button>
            </div>
            <button class="browser-close-btn" @click="isPreviewModalOpen = false" title="关闭预览">
              <span class="codicon codicon-close"></span>
            </button>
          </div>
          <div class="simple-browser-body">
            <iframe v-if="previewModalUrl" :src="previewModalUrl" class="simple-browser-iframe" frameborder="0"></iframe>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Modal: Change Local Address Port (Prompt Dialog) -->
    <Teleport to="body">
      <div v-if="isChangeLocalPortModalOpen" class="change-port-backdrop" @click.self="isChangeLocalPortModalOpen = false">
        <div class="change-port-dialog">
          <div class="change-port-header">
            <span class="change-port-title">{{ t('ports.change_local_port') || '更改本地地址端口' }}</span>
            <button class="dialog-close-btn" @click="isChangeLocalPortModalOpen = false">
              <span class="codicon codicon-close"></span>
            </button>
          </div>
          <div class="change-port-body">
            <p class="change-port-desc">
              为端口 <strong>{{ targetLocalPortItem?.port }}</strong> 指定新的本地映射端口号 (1 - 65535)：
            </p>
            <input
              ref="changeLocalPortInputRef"
              v-model.number="changeLocalPortInput"
              type="number"
              min="1"
              max="65535"
              class="change-port-input font-mono"
              @keydown.enter="submitChangeLocalPort"
              @keydown.esc="isChangeLocalPortModalOpen = false"
            />
          </div>
          <div class="change-port-footer">
            <button class="dialog-btn secondary" @click="isChangeLocalPortModalOpen = false">取消</button>
            <button class="dialog-btn primary" @click="submitChangeLocalPort">确认</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from "vue";
import { usePorts } from "../../composables/usePorts";
import { t } from "../../i18n";
import type { PortForwardItem, PortProtocol } from "../../types/ports";

const isMac = typeof navigator !== "undefined" && /(Mac|iPhone|iPod|iPad)/i.test(navigator.platform || navigator.userAgent);

const {
  filteredPorts,
  isLoading,
  searchQuery,
  fetchPorts,
  addPort,
  updatePort,
  stopPort,
  startPort,
  removePort,
  copyAddress,
  openPort,
} = usePorts();

const selectedPortId = ref<string | null>(null);

// Inline Add State
const isAddingInline = ref<boolean>(false);
const newPortNumber = ref<number | undefined>(undefined);
const newPortLabel = ref<string>("");
const newPortInputRef = ref<HTMLInputElement | null>(null);

// Inline Label Edit State
const editingPortId = ref<string | null>(null);
const editLabelValue = ref<string>("");
const editLabelInputRef = ref<HTMLInputElement | null>(null);

// Context Menu State
const contextMenuVisible = ref<boolean>(false);
const contextMenuPos = ref({ x: 0, y: 0 });
const contextPort = ref<PortForwardItem | null>(null);

// Protocol Submenu State
const isProtocolSubmenuOpen = ref<boolean>(false);
const protocolSubmenuPos = ref({ x: 0, y: 0 });
let protocolSubmenuTimer: any = null;

// Simple Browser In-Editor Preview Modal State
const isPreviewModalOpen = ref<boolean>(false);
const previewModalUrl = ref<string>("");
const previewModalTitle = ref<string>("Simple Browser");

// Change Local Port Prompt Dialog State
const isChangeLocalPortModalOpen = ref<boolean>(false);
const changeLocalPortInput = ref<number | undefined>(undefined);
const targetLocalPortItem = ref<PortForwardItem | null>(null);
const changeLocalPortInputRef = ref<HTMLInputElement | null>(null);

let pollTimer: any = null;

onMounted(async () => {
  await fetchPorts();
  // Poll listening ports every 3.5 seconds
  pollTimer = setInterval(() => {
    fetchPorts();
  }, 3500);

  window.addEventListener("click", closeContextMenu);
});

onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer);
  window.removeEventListener("click", closeContextMenu);
});

async function handleRefresh() {
  await fetchPorts();
}

function startInlineAdd() {
  isAddingInline.value = true;
  newPortNumber.value = undefined;
  newPortLabel.value = "";
  nextTick(() => {
    newPortInputRef.value?.focus();
  });
}

function cancelInlineAdd() {
  isAddingInline.value = false;
  newPortNumber.value = undefined;
  newPortLabel.value = "";
}

async function submitInlineAdd() {
  if (!newPortNumber.value) return;
  const ok = await addPort({
    port: newPortNumber.value,
    label: newPortLabel.value,
  });
  if (ok) {
    cancelInlineAdd();
  }
}

function startEditLabel(item: PortForwardItem) {
  editingPortId.value = item.id;
  editLabelValue.value = item.label;
  nextTick(() => {
    editLabelInputRef.value?.focus();
  });
}

function cancelEditLabel() {
  editingPortId.value = null;
  editLabelValue.value = "";
}

async function submitEditLabel(item: PortForwardItem) {
  if (!editingPortId.value) return;
  const newLabel = editLabelValue.value.trim();
  if (newLabel && newLabel !== item.label) {
    await updatePort(item.port, { label: newLabel });
  }
  cancelEditLabel();
}

async function toggleVisibility(item: PortForwardItem) {
  const nextVis = item.visibility === "public" ? "private" : "public";
  await updatePort(item.port, { visibility: nextVis });
}

function openContextMenu(event: MouseEvent, item: PortForwardItem) {
  contextPort.value = item;
  contextMenuPos.value = {
    x: Math.min(event.clientX, window.innerWidth - 240),
    y: Math.min(event.clientY, window.innerHeight - 300),
  };
  contextMenuVisible.value = true;
  isProtocolSubmenuOpen.value = false;
}

function closeContextMenu() {
  contextMenuVisible.value = false;
  isProtocolSubmenuOpen.value = false;
}

function openProtocolSubmenu(event: MouseEvent) {
  cancelCloseProtocolSubmenu();
  const target = (event.currentTarget as HTMLElement).getBoundingClientRect();
  const screenWidth = window.innerWidth;
  let x = target.right;
  if (x + 130 > screenWidth) {
    x = target.left - 130;
  }
  protocolSubmenuPos.value = {
    x,
    y: target.top - 4,
  };
  isProtocolSubmenuOpen.value = true;
}

function scheduleCloseProtocolSubmenu() {
  protocolSubmenuTimer = setTimeout(() => {
    isProtocolSubmenuOpen.value = false;
  }, 220);
}

function cancelCloseProtocolSubmenu() {
  if (protocolSubmenuTimer) {
    clearTimeout(protocolSubmenuTimer);
    protocolSubmenuTimer = null;
  }
}

async function handleSelectProtocol(proto: PortProtocol) {
  if (!contextPort.value) return;
  const item = contextPort.value;
  closeContextMenu();
  await updatePort(item.port, { protocol: proto });
}

function openInEditorPreview(item: PortForwardItem) {
  const url = item.previewUrl || `http://localhost:${item.localPort || item.port}/`;
  previewModalUrl.value = url;
  previewModalTitle.value = `Simple Browser: ${item.label || item.port}`;
  isPreviewModalOpen.value = true;
}

function reloadPreviewFrame() {
  const current = previewModalUrl.value;
  previewModalUrl.value = "";
  nextTick(() => {
    previewModalUrl.value = current;
  });
}

function openExternalFromPreview() {
  if (previewModalUrl.value) {
    window.open(previewModalUrl.value, "_blank");
  }
}

function openChangeLocalPortModal(item: PortForwardItem) {
  targetLocalPortItem.value = item;
  changeLocalPortInput.value = item.localPort || item.port;
  isChangeLocalPortModalOpen.value = true;
  nextTick(() => {
    changeLocalPortInputRef.value?.focus();
    changeLocalPortInputRef.value?.select();
  });
}

async function submitChangeLocalPort() {
  if (!targetLocalPortItem.value || !changeLocalPortInput.value) return;
  const newPort = Number(changeLocalPortInput.value);
  if (isNaN(newPort) || newPort < 1 || newPort > 65535) {
    return;
  }
  const item = targetLocalPortItem.value;
  isChangeLocalPortModalOpen.value = false;
  await updatePort(item.port, { localPort: newPort });
}

function handleRowKeyDown(event: KeyboardEvent, item: PortForwardItem) {
  if (editingPortId.value === item.id) return;

  // Enter -> edit label
  if (event.key === "Enter") {
    event.preventDefault();
    startEditLabel(item);
  }
  // ⌘C / Ctrl+C -> copy local address
  else if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "c") {
    event.preventDefault();
    copyAddress(item);
  }
  // ⌘⌫ / Delete -> stop forwarding
  else if ((event.metaKey && (event.key === "Backspace" || event.key === "Delete")) || event.key === "Delete") {
    event.preventDefault();
    if (item.status === "active") {
      stopPort(item.port);
    } else {
      startPort(item.port);
    }
  }
}

async function handleContextAction(action: string) {
  if (!contextPort.value) return;
  const item = contextPort.value;

  switch (action) {
    case "open":
      closeContextMenu();
      openPort(item);
      break;
    case "preview":
      closeContextMenu();
      openInEditorPreview(item);
      break;
    case "edit-label":
      closeContextMenu();
      startEditLabel(item);
      break;
    case "copy-address":
      closeContextMenu();
      await copyAddress(item);
      break;
    case "change-local-port":
      closeContextMenu();
      openChangeLocalPortModal(item);
      break;
    case "toggle-status":
      closeContextMenu();
      if (item.status === "active") {
        await stopPort(item.port);
      } else {
        await startPort(item.port);
      }
      break;
    case "forward-port":
      closeContextMenu();
      startInlineAdd();
      break;
    case "remove":
      closeContextMenu();
      await removePort(item.port);
      break;
  }
}
</script>

<style scoped>
.ports-panel-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  background: var(--card-bg, #0b1120);
  color: var(--text-main, #f8fafc);
  overflow: hidden;
  font-family: var(--font-sans, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif);
}

/* Toolbar */
.ports-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  border-bottom: 1px solid var(--card-border, #1e293b);
  background: var(--card-bg, #0b1120);
  gap: 12px;
  flex-shrink: 0;
}

.toolbar-search {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1;
  max-width: 320px;
}

.search-icon {
  position: absolute;
  left: 8px;
  font-size: 13px;
  color: var(--text-muted, #94a3b8);
  pointer-events: none;
}

.ports-search-input {
  width: 100%;
  background: var(--input-bg, #0f172a);
  border: 1px solid var(--card-border, #334155);
  border-radius: 4px;
  color: var(--text-main, #f8fafc);
  font-size: 12px;
  padding: 4px 26px 4px 28px;
  outline: none;
  transition: border-color 0.15s ease, background-color 0.15s ease;
}

.ports-search-input:focus {
  border-color: var(--primary, #3b82f6);
}

.search-clear-btn {
  position: absolute;
  right: 6px;
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  padding: 2px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ports-tool-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--input-bg, #1e293b);
  border: 1px solid var(--card-border, #334155);
  border-radius: 4px;
  color: var(--text-main, #e2e8f0);
  padding: 4px 10px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.ports-tool-btn:hover {
  background: rgba(128, 128, 128, 0.2);
  border-color: var(--primary, #3b82f6);
}

.ports-tool-btn.primary {
  background: var(--primary, #3b82f6);
  border-color: var(--primary, #3b82f6);
  color: #fff;
}

.ports-tool-btn.primary:hover {
  filter: brightness(1.1);
}

/* Table Area */
.ports-table-wrapper {
  flex: 1;
  overflow-y: auto;
  position: relative;
  background: var(--card-bg, #0b1120);
}

.ports-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  text-align: left;
}

.ports-table thead th {
  position: sticky;
  top: 0;
  background: var(--header-bg, #0f172a);
  color: var(--text-muted, #94a3b8);
  font-weight: 600;
  padding: 6px 12px;
  border-bottom: 1px solid var(--card-border, #1e293b);
  z-index: 2;
  user-select: none;
}

.th-port { width: 220px; }
.th-address { width: 240px; }
.th-process { width: auto; }
.th-source { width: 100px; }
.th-visibility { width: 100px; }
.th-actions { width: 44px; text-align: center; }

.port-row {
  border-bottom: 1px solid var(--card-border, #1e293b);
  transition: background-color 0.1s ease;
  cursor: pointer;
  background: transparent;
}

.port-row:hover {
  background: rgba(128, 128, 128, 0.08);
}

.port-row.active {
  background: rgba(59, 130, 246, 0.12);
}

.port-row.inactive {
  opacity: 0.72;
}

.ports-table td {
  padding: 7px 12px;
  vertical-align: middle;
}

/* Port Cell */
.port-cell-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.port-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.port-status-dot.dot-active {
  background: #10b981;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
}

.port-status-dot.dot-inactive {
  border: 1.5px solid var(--text-muted, #64748b);
  background: transparent;
}

.port-number {
  font-weight: 600;
  color: var(--text-main, #f8fafc);
}

.port-mapping-pill {
  font-size: 11px;
  background: rgba(59, 130, 246, 0.15);
  color: var(--primary, #3b82f6);
  padding: 1px 5px;
  border-radius: 4px;
}

.protocol-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 5px;
  border-radius: 3px;
  text-transform: uppercase;
}

.protocol-badge.https {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.origin-pill {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  background: rgba(148, 163, 184, 0.15);
  color: var(--text-muted, #94a3b8);
  border: 1px solid rgba(148, 163, 184, 0.25);
  text-transform: lowercase;
}

.origin-pill.compose {
  background: rgba(14, 165, 233, 0.15);
  color: #0284c7;
  border-color: rgba(14, 165, 233, 0.3);
}

.port-label {
  font-size: 11px;
  color: var(--text-muted, #94a3b8);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 140px;
}

.inline-edit-input {
  background: var(--input-bg, #0f172a);
  border: 1px solid var(--primary, #3b82f6);
  border-radius: 3px;
  color: var(--text-main, #f8fafc);
  font-size: 11px;
  padding: 1px 6px;
  outline: none;
}

/* Address Cell */
.address-cell-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: relative;
}

.address-link {
  color: #38bdf8;
  text-decoration: none;
  font-size: 12px;
  font-weight: 500;
}

.address-link:hover {
  text-decoration: underline;
}

.row-hover-actions {
  display: none;
  align-items: center;
  gap: 3px;
  margin-left: auto;
}

.port-row:hover .row-hover-actions {
  display: flex;
}

.icon-action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  padding: 3px;
  border-radius: 3px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.icon-action-btn:hover {
  background: rgba(128, 128, 128, 0.25);
  color: var(--text-main, #f8fafc);
}

.icon-action-btn.danger:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

/* Process Cell */
.process-cell-wrapper {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.process-text {
  font-size: 11px;
  color: var(--text-main, #e2e8f0);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 320px;
}

.process-text.is-empty {
  color: var(--text-muted, #64748b);
  font-style: italic;
}

.probe-banner {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  color: #a855f7;
  background: rgba(168, 85, 247, 0.12);
  border: 1px solid rgba(168, 85, 247, 0.25);
  padding: 1px 5px;
  border-radius: 3px;
  width: fit-content;
}

.probe-banner .codicon {
  font-size: 10px;
}

/* Source & Visibility Badges */
.source-badge {
  font-size: 11px;
  color: var(--text-muted, #94a3b8);
}

.visibility-toggle-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: none;
  font-size: 11px;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
}

.visibility-toggle-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.visibility-toggle-btn.vis-public {
  color: #10b981;
}

/* More Actions */
.more-action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  padding: 3px;
  border-radius: 3px;
}

.more-action-btn:hover {
  background: rgba(128, 128, 128, 0.2);
  color: var(--text-main);
}

/* Bottom Bar */
.table-bottom-bar {
  padding: 8px 12px;
  border-top: 1px solid var(--card-border, #1e293b);
  background: var(--card-bg, #0b1120);
}

.bottom-add-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--primary, #0284c7);
  color: #fff;
  border: none;
  padding: 5px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s ease;
}

.bottom-add-btn:hover {
  filter: brightness(1.1);
}

/* Adding Inline Row */
.adding-row {
  background: rgba(59, 130, 246, 0.08);
}

.inline-port-input {
  width: 90px;
  background: var(--input-bg, #0f172a);
  border: 1px solid var(--primary, #3b82f6);
  border-radius: 3px;
  color: var(--text-main, #f8fafc);
  font-size: 12px;
  padding: 3px 6px;
  outline: none;
}

.inline-label-input {
  width: 130px;
  background: var(--input-bg, #0f172a);
  border: 1px solid var(--card-border, #334155);
  border-radius: 3px;
  color: var(--text-main, #f8fafc);
  font-size: 12px;
  padding: 3px 6px;
  outline: none;
}

.inline-submit-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.inline-btn {
  background: transparent;
  border: none;
  padding: 3px 5px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
}

.inline-btn.confirm {
  color: #10b981;
}

.inline-btn.cancel {
  color: #ef4444;
}

/* Empty State */
.ports-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  gap: 8px;
  color: var(--text-muted, #94a3b8);
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 4px;
  opacity: 0.6;
}

.empty-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-main, #f8fafc);
}

.empty-desc {
  font-size: 12px;
  max-width: 380px;
  text-align: center;
}

.empty-add-btn {
  margin-top: 12px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--primary, #3b82f6);
  color: #fff;
  border: none;
  padding: 6px 14px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}

/* Context Menu (VS Code Native Aesthetic) */
.ports-context-menu {
  position: fixed;
  background: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, #454545);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.45);
  border-radius: 5px;
  padding: 4px 0;
  z-index: 99999;
  min-width: 190px;
  font-size: 13px;
  font-family: var(--font-sans, system-ui, -apple-system, sans-serif);
  user-select: none;
}

.ports-protocol-submenu {
  min-width: 120px;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 16px 4px 16px;
  color: var(--text-main, #cccccc);
  cursor: pointer;
  transition: background 0.08s ease, color 0.08s ease;
  height: 26px;
  box-sizing: border-box;
}

.menu-item .action-label {
  flex: 1;
  white-space: nowrap;
}

.menu-item .keybinding {
  font-size: 11px;
  color: var(--text-muted, #858585);
  margin-left: 20px;
}

.menu-item .submenu-chevron {
  font-size: 14px;
  color: var(--text-muted, #858585);
  margin-left: 14px;
}

.menu-item.check-item {
  padding-left: 10px;
  justify-content: flex-start;
  gap: 6px;
}

.menu-item .check-mark {
  width: 14px;
  display: inline-block;
  font-size: 12px;
  font-weight: 700;
  color: var(--primary, #38bdf8);
}

.menu-item:hover,
.menu-item.active {
  background: var(--primary, #094771);
  color: #ffffff;
}

.menu-item:hover .keybinding,
.menu-item:hover .submenu-chevron,
.menu-item:hover .check-mark {
  color: #ffffff;
}

.menu-item.danger:hover {
  background: #e51400;
  color: #ffffff;
}

.menu-separator {
  height: 1px;
  background: var(--card-border, #333333);
  margin: 4px 0;
}

/* Simple Browser Modal */
.simple-browser-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100000;
}

.simple-browser-window {
  width: 86vw;
  height: 84vh;
  max-width: 1280px;
  background: var(--card-bg, #182234);
  border: 1px solid var(--card-border, #334155);
  border-radius: 8px;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.simple-browser-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  background: var(--header-bg, #0f172a);
  border-bottom: 1px solid var(--card-border, #1e293b);
  gap: 12px;
}

.simple-browser-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-main, #f8fafc);
  font-size: 13px;
  font-weight: 600;
}

.simple-browser-address-bar {
  flex: 1;
  max-width: 600px;
  display: flex;
  align-items: center;
  background: var(--input-bg, #0b1120);
  border: 1px solid var(--card-border, #334155);
  border-radius: 4px;
  padding: 2px 6px;
  gap: 6px;
}

.browser-url-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main, #e2e8f0);
  font-size: 12px;
}

.browser-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  padding: 3px;
  border-radius: 3px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
}

.browser-btn:hover {
  background: rgba(128, 128, 128, 0.2);
  color: var(--text-main, #fff);
}

.browser-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  font-size: 14px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}

.browser-close-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.simple-browser-body {
  flex: 1;
  width: 100%;
  background: #ffffff;
}

.simple-browser-iframe {
  width: 100%;
  height: 100%;
  border: none;
}

/* Change Local Port Prompt Dialog */
.change-port-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(3px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100000;
}

.change-port-dialog {
  width: 380px;
  background: var(--card-bg, #1e293b);
  border: 1px solid var(--card-border, #334155);
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.change-port-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.change-port-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f8fafc);
}

.dialog-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  font-size: 13px;
  padding: 2px;
}

.dialog-close-btn:hover {
  color: var(--text-main, #f8fafc);
}

.change-port-desc {
  font-size: 12px;
  color: var(--text-muted, #cbd5e1);
  margin: 0 0 8px 0;
  line-height: 1.5;
}

.change-port-input {
  width: 100%;
  box-sizing: border-box;
  background: var(--input-bg, #0f172a);
  border: 1px solid var(--card-border, #334155);
  border-radius: 4px;
  color: var(--text-main, #f8fafc);
  font-size: 13px;
  padding: 6px 10px;
  outline: none;
}

.change-port-input:focus {
  border-color: var(--primary, #0284c7);
}

.change-port-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 4px;
}

.dialog-btn {
  padding: 5px 14px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 4px;
  cursor: pointer;
  border: none;
  transition: all 0.15s ease;
}

.dialog-btn.secondary {
  background: var(--input-bg, #334155);
  color: var(--text-main, #e2e8f0);
}

.dialog-btn.secondary:hover {
  background: rgba(128, 128, 128, 0.25);
}

.dialog-btn.primary {
  background: var(--primary, #0284c7);
  color: #ffffff;
}

.dialog-btn.primary:hover {
  filter: brightness(1.1);
}

/* ==========================================================================
   LIGHT THEME POLISH (Aligning with VS Code & UniRTM Light Palette)
   ========================================================================== */
:root[data-theme="light"] .ports-panel-container,
[data-theme="light"] .ports-panel-container {
  background: #ffffff;
  color: #0f172a;
}

:root[data-theme="light"] .ports-toolbar,
[data-theme="light"] .ports-toolbar {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

:root[data-theme="light"] .ports-search-input,
[data-theme="light"] .ports-search-input {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

:root[data-theme="light"] .ports-search-input:focus,
[data-theme="light"] .ports-search-input:focus {
  border-color: #0284c7;
}

:root[data-theme="light"] .ports-tool-btn,
[data-theme="light"] .ports-tool-btn {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #334155;
}

:root[data-theme="light"] .ports-tool-btn:hover,
[data-theme="light"] .ports-tool-btn:hover {
  background: #f1f5f9;
  border-color: #0284c7;
  color: #0284c7;
}

:root[data-theme="light"] .ports-tool-btn.primary,
[data-theme="light"] .ports-tool-btn.primary {
  background: #0284c7;
  border-color: #0284c7;
  color: #ffffff;
}

:root[data-theme="light"] .ports-tool-btn.primary:hover,
[data-theme="light"] .ports-tool-btn.primary:hover {
  background: #0369a1;
}

:root[data-theme="light"] .ports-table-wrapper,
[data-theme="light"] .ports-table-wrapper {
  background: #ffffff;
}

:root[data-theme="light"] .ports-table thead th,
[data-theme="light"] .ports-table thead th {
  background: #f8fafc;
  color: #64748b;
  border-bottom: 1px solid #e2e8f0;
}

:root[data-theme="light"] .port-row,
[data-theme="light"] .port-row {
  border-bottom: 1px solid #f1f5f9;
}

:root[data-theme="light"] .port-row:hover,
[data-theme="light"] .port-row:hover {
  background: rgba(2, 132, 199, 0.04);
}

:root[data-theme="light"] .port-row.active,
[data-theme="light"] .port-row.active {
  background: rgba(2, 132, 199, 0.1);
}

:root[data-theme="light"] .port-number,
[data-theme="light"] .port-number {
  color: #0f172a;
}

:root[data-theme="light"] .port-mapping-pill,
[data-theme="light"] .port-mapping-pill {
  background: #e0f2fe;
  color: #0284c7;
}

:root[data-theme="light"] .origin-pill,
[data-theme="light"] .origin-pill {
  background: #f1f5f9;
  color: #475569;
  border-color: #cbd5e1;
}

:root[data-theme="light"] .origin-pill.compose,
[data-theme="light"] .origin-pill.compose {
  background: #e0f2fe;
  color: #0369a1;
  border-color: #bae6fd;
}

:root[data-theme="light"] .port-label,
[data-theme="light"] .port-label {
  color: #64748b;
}

:root[data-theme="light"] .inline-edit-input,
[data-theme="light"] .inline-edit-input {
  background: #ffffff;
  border-color: #0284c7;
  color: #0f172a;
}

:root[data-theme="light"] .address-link,
[data-theme="light"] .address-link {
  color: #0284c7;
}

:root[data-theme="light"] .address-link:hover,
[data-theme="light"] .address-link:hover {
  color: #0369a1;
}

:root[data-theme="light"] .icon-action-btn,
[data-theme="light"] .icon-action-btn {
  color: #64748b;
}

:root[data-theme="light"] .icon-action-btn:hover,
[data-theme="light"] .icon-action-btn:hover {
  background: #e2e8f0;
  color: #0f172a;
}

:root[data-theme="light"] .process-text,
[data-theme="light"] .process-text {
  color: #334155;
}

:root[data-theme="light"] .process-text.is-empty,
[data-theme="light"] .process-text.is-empty {
  color: #94a3b8;
}

:root[data-theme="light"] .probe-banner,
[data-theme="light"] .probe-banner {
  color: #7c3aed;
  background: #f5f3ff;
  border-color: #ddd6fe;
}

:root[data-theme="light"] .source-badge,
[data-theme="light"] .source-badge {
  color: #64748b;
}

:root[data-theme="light"] .visibility-toggle-btn,
[data-theme="light"] .visibility-toggle-btn {
  color: #64748b;
}

:root[data-theme="light"] .visibility-toggle-btn:hover,
[data-theme="light"] .visibility-toggle-btn:hover {
  background: #f1f5f9;
  color: #0f172a;
}

:root[data-theme="light"] .visibility-toggle-btn.vis-public,
[data-theme="light"] .visibility-toggle-btn.vis-public {
  color: #16a34a;
}

:root[data-theme="light"] .more-action-btn,
[data-theme="light"] .more-action-btn {
  color: #64748b;
}

:root[data-theme="light"] .more-action-btn:hover,
[data-theme="light"] .more-action-btn:hover {
  background: #e2e8f0;
  color: #0f172a;
}

:root[data-theme="light"] .table-bottom-bar,
[data-theme="light"] .table-bottom-bar {
  background: #ffffff;
  border-top: 1px solid #e2e8f0;
}

:root[data-theme="light"] .bottom-add-btn,
[data-theme="light"] .bottom-add-btn {
  background: #0284c7;
  color: #ffffff;
}

:root[data-theme="light"] .bottom-add-btn:hover,
[data-theme="light"] .bottom-add-btn:hover {
  background: #0369a1;
}

:root[data-theme="light"] .adding-row,
[data-theme="light"] .adding-row {
  background: rgba(2, 132, 199, 0.06);
}

:root[data-theme="light"] .inline-port-input,
:root[data-theme="light"] .inline-label-input,
[data-theme="light"] .inline-port-input,
[data-theme="light"] .inline-label-input {
  background: #ffffff;
  color: #0f172a;
  border-color: #cbd5e1;
}

:root[data-theme="light"] .inline-port-input:focus,
:root[data-theme="light"] .inline-label-input:focus,
[data-theme="light"] .inline-port-input:focus,
[data-theme="light"] .inline-label-input:focus {
  border-color: #0284c7;
}

:root[data-theme="light"] .ports-empty-state,
[data-theme="light"] .ports-empty-state {
  color: #64748b;
}

:root[data-theme="light"] .empty-title,
[data-theme="light"] .empty-title {
  color: #0f172a;
}

:global(:root[data-theme="light"] .ports-context-menu),
:global([data-theme="light"] .ports-context-menu) {
  background: #f8fafc !important;
  border-color: #cbd5e1 !important;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12) !important;
}

:global(:root[data-theme="light"] .ports-context-menu .menu-item),
:global([data-theme="light"] .ports-context-menu .menu-item) {
  color: #1e293b !important;
}

:global(:root[data-theme="light"] .ports-context-menu .menu-item .keybinding),
:global([data-theme="light"] .ports-context-menu .menu-item .submenu-chevron) {
  color: #64748b !important;
}

:global(:root[data-theme="light"] .ports-context-menu .menu-item:hover),
:global([data-theme="light"] .ports-context-menu .menu-item:hover .keybinding),
:global([data-theme="light"] .ports-context-menu .menu-item:hover .submenu-chevron),
:global([data-theme="light"] .ports-context-menu .menu-item:hover .check-mark) {
  background: #0284c7 !important;
  color: #ffffff !important;
}

:global(:root[data-theme="light"] .ports-context-menu .menu-item.danger:hover) {
  background: #ef4444 !important;
  color: #ffffff !important;
}

:global(:root[data-theme="light"] .ports-context-menu .menu-separator),
:global([data-theme="light"] .ports-context-menu .menu-separator) {
  background: #e2e8f0 !important;
}

:global(:root[data-theme="light"] .simple-browser-window),
:global([data-theme="light"] .simple-browser-window) {
  background: #ffffff !important;
  border-color: #cbd5e1 !important;
}

:global(:root[data-theme="light"] .simple-browser-header),
:global([data-theme="light"] .simple-browser-header) {
  background: #f8fafc !important;
  border-bottom-color: #e2e8f0 !important;
}

:global(:root[data-theme="light"] .simple-browser-address-bar),
:global([data-theme="light"] .simple-browser-address-bar) {
  background: #ffffff !important;
  border-color: #cbd5e1 !important;
}

:global(:root[data-theme="light"] .browser-url-input),
:global([data-theme="light"] .browser-url-input) {
  color: #0f172a !important;
}

:global(:root[data-theme="light"] .change-port-dialog),
:global([data-theme="light"] .change-port-dialog) {
  background: #ffffff !important;
  border-color: #cbd5e1 !important;
}

:global(:root[data-theme="light"] .change-port-title),
:global([data-theme="light"] .change-port-title) {
  color: #0f172a !important;
}

:global(:root[data-theme="light"] .change-port-desc),
:global([data-theme="light"] .change-port-desc) {
  color: #475569 !important;
}

:global(:root[data-theme="light"] .change-port-input),
:global([data-theme="light"] .change-port-input) {
  background: #ffffff !important;
  border-color: #cbd5e1 !important;
  color: #0f172a !important;
}

:global(:root[data-theme="light"] .dialog-btn.secondary),
:global([data-theme="light"] .dialog-btn.secondary) {
  background: #f1f5f9 !important;
  color: #334155 !important;
}
</style>
