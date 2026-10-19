<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from "vue";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { getXtermTheme } from "../../utils/terminalTheme";
import { registerTerminalSender, unregisterTerminalSender } from "../../composables/useTerminalSessions";
import { usePorts } from "../../composables/usePorts";
import { createTerminalLinkProvider } from "../../composables/useTerminalLinks";
import { SearchAddon } from "@xterm/addon-search";
import TerminalSearchWidget from "./TerminalSearchWidget.vue";
import TerminalContextMenu from "./TerminalContextMenu.vue";
import { useWorkspaceBus } from "../../composables/useWorkspaceBus";

const { registerDetectedPort } = usePorts();
const detectedPorts = new Set<number>();

function sniffTerminalPorts(text: string) {
  if (!text) return;
  const regex = /(?:https?:\/\/)?(?:localhost|127\.0\.0\.1|0\.0\.0\.0|\[::1\]):(\d{2,5})/gi;
  let match;
  while ((match = regex.exec(text)) !== null) {
    const port = parseInt(match[1], 10);
    if (port && port > 0 && port <= 65535 && !detectedPorts.has(port)) {
      detectedPorts.add(port);
      registerDetectedPort(port);
    }
  }
}

const props = defineProps<{
  paneId: string;
  shellName: string;
  shellPath: string;
  isActive: boolean;
}>();

const emit = defineEmits<{
  (e: "focus"): void;
}>();

const terminalContainer = ref<HTMLDivElement | null>(null);
const isConnected = ref(false);
const isConnecting = ref(true);
const connectionError = ref<string | null>(null);
const isSearchVisible = ref(false);
const searchAddonInstance = ref<SearchAddon | null>(null);

const contextMenuVisible = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const hasSelection = ref(false);

const busCleanups: (() => void)[] = [];

function closeSearch() {
  isSearchVisible.value = false;
  term?.focus();
}

function openSearch() {
  isSearchVisible.value = true;
}

function handleContextMenu(e: MouseEvent) {
  contextMenuX.value = e.clientX;
  contextMenuY.value = e.clientY;
  hasSelection.value = !!term?.hasSelection();
  contextMenuVisible.value = true;
}

async function handleCopy() {
  if (!term) return;
  const selection = term.getSelection();
  if (selection && typeof navigator !== "undefined" && navigator.clipboard) {
    try {
      await navigator.clipboard.writeText(selection);
    } catch (err) {
      console.error("[XTerminalPane] Failed to copy selection:", err);
    }
  }
}

async function handlePaste() {
  if (!term || !ws || ws.readyState !== WebSocket.OPEN) return;
  if (typeof navigator !== "undefined" && navigator.clipboard) {
    try {
      const text = await navigator.clipboard.readText();
      if (text) {
        ws.send(JSON.stringify({ type: "input", data: text }));
      }
    } catch (err) {
      console.error("[XTerminalPane] Failed to paste text:", err);
    }
  }
}

function handleSelectAll() {
  term?.selectAll();
}

function handleClear() {
  term?.clear();
}

let term: Terminal | null = null;
let fitAddon: FitAddon | null = null;
let ws: WebSocket | null = null;
let resizeObserver: ResizeObserver | null = null;
let themeObserver: MutationObserver | null = null;

function checkIsDark(): boolean {
  const root = document.documentElement;
  const theme = root.getAttribute("data-theme");
  if (theme === "dark") return true;
  if (theme === "light") return false;
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

function updateTerminalTheme() {
  if (!term) return;
  const isDark = checkIsDark();
  term.options.theme = getXtermTheme(isDark);
}

function connectWebSocket() {
  if (!term) return;
  isConnecting.value = true;
  connectionError.value = null;

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.host;
  const cols = term.cols || 80;
  const rows = term.rows || 24;

  const url = `${protocol}//${host}/api/terminal/ws?id=${encodeURIComponent(props.paneId)}&shell=${encodeURIComponent(props.shellPath)}&cols=${cols}&rows=${rows}`;

  try {
    ws = new WebSocket(url);
  } catch (err: any) {
    isConnecting.value = false;
    connectionError.value = err.message || "Failed to create WebSocket";
    return;
  }

  ws.onopen = () => {
    isConnected.value = true;
    isConnecting.value = false;
    connectionError.value = null;
    registerTerminalSender(props.paneId, (data: string) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "input", data }));
      }
    });
    if (fitAddon) {
      fitAddon.fit();
      sendResize();
    }
  };

  ws.onmessage = (event) => {
    if (!term) return;
    try {
      const msg = JSON.parse(event.data);
      if (msg.type === "output" && typeof msg.data === "string") {
        term.write(msg.data);
        sniffTerminalPorts(msg.data);
      }
    } catch {
      // Raw string fallback
      term.write(event.data);
      sniffTerminalPorts(String(event.data));
    }
  };

  ws.onerror = () => {
    isConnected.value = false;
    isConnecting.value = false;
    connectionError.value = "WebSocket connection error";
  };

  ws.onclose = () => {
    isConnected.value = false;
    isConnecting.value = false;
  };
}

function sendResize() {
  if (!term || !ws || ws.readyState !== WebSocket.OPEN) return;
  ws.send(
    JSON.stringify({
      type: "resize",
      cols: term.cols,
      rows: term.rows,
    })
  );
}

function reconnect() {
  if (ws) {
    try {
      ws.close();
    } catch {}
    ws = null;
  }
  if (term) {
    term.reset();
  }
  connectWebSocket();
}

onMounted(() => {
  if (!terminalContainer.value) return;

  const isDark = checkIsDark();

  term = new Terminal({
    cursorBlink: true,
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Menlo, Monaco, Consolas, monospace",
    fontSize: 13,
    lineHeight: 1.35,
    theme: getXtermTheme(isDark),
    allowProposedApi: true,
    convertEol: true,
  });

  fitAddon = new FitAddon();
  term.loadAddon(fitAddon);

  const searchAddon = new SearchAddon();
  term.loadAddon(searchAddon);
  searchAddonInstance.value = searchAddon;

  term.registerLinkProvider(createTerminalLinkProvider(term));

  term.attachCustomKeyEventHandler((e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "f") {
      e.preventDefault();
      e.stopPropagation();
      openSearch();
      return false;
    }
    return true;
  });

  term.open(terminalContainer.value);

  // Hook user keystrokes to websocket
  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: "input", data }));
    }
  });

  // Fit initial size
  nextTick(() => {
    if (fitAddon) {
      fitAddon.fit();
    }
    connectWebSocket();
    if (props.isActive && term) {
      term.focus();
    }
  });

  // Observe container resize
  resizeObserver = new ResizeObserver(() => {
    if (!fitAddon || !term) return;
    try {
      fitAddon.fit();
      sendResize();
    } catch {}
  });
  resizeObserver.observe(terminalContainer.value);

  // Observe theme changes on documentElement
  themeObserver = new MutationObserver(() => {
    updateTerminalTheme();
  });
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ["data-theme"],
  });

  // Listen to workspace bus events
  const bus = useWorkspaceBus();
  busCleanups.push(
    bus.onSendToTerminal((req) => {
      if (!props.isActive) return;
      if (ws && ws.readyState === WebSocket.OPEN) {
        const textToSend = req.addNewline !== false ? `${req.text}\n` : req.text;
        ws.send(JSON.stringify({ type: "input", data: textToSend }));
      }
    }),
    bus.onRevealInTerminal((req) => {
      if (!props.isActive) return;
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "input", data: `cd "${req.directoryPath}"\n` }));
      }
    })
  );
});

watch(
  () => props.isActive,
  (active) => {
    if (active && term) {
      nextTick(() => {
        fitAddon?.fit();
        sendResize();
        term?.focus();
      });
    }
  }
);

onBeforeUnmount(() => {
  busCleanups.forEach((fn) => fn());
  busCleanups.length = 0;
  unregisterTerminalSender(props.paneId);
  if (resizeObserver) {
    resizeObserver.disconnect();
    resizeObserver = null;
  }
  if (themeObserver) {
    themeObserver.disconnect();
    themeObserver = null;
  }
  if (ws) {
    try {
      ws.send(JSON.stringify({ type: "kill" }));
      ws.close();
    } catch {}
    ws = null;
  }
  if (term) {
    term.dispose();
    term = null;
  }
});
</script>

<template>
  <div
    class="xterm-pane-wrapper"
    @click="emit('focus')"
    @contextmenu.prevent="handleContextMenu"
  >
    <div ref="terminalContainer" class="xterm-container"></div>

    <!-- Floating Search Widget -->
    <TerminalSearchWidget
      :visible="isSearchVisible"
      :search-addon="searchAddonInstance"
      @close="closeSearch"
    />

    <!-- Context Menu -->
    <TerminalContextMenu
      :visible="contextMenuVisible"
      :x="contextMenuX"
      :y="contextMenuY"
      :has-selection="hasSelection"
      @close="contextMenuVisible = false"
      @copy="handleCopy"
      @paste="handlePaste"
      @select-all="handleSelectAll"
      @clear="handleClear"
      @find="openSearch"
      @restart="reconnect"
    />

    <!-- Reconnect / Error Overlay -->
    <div v-if="!isConnected && !isConnecting" class="xterm-overlay">
      <div class="xterm-overlay-card">
        <span class="overlay-icon">⚠️</span>
        <span class="overlay-text">{{ connectionError || '终端会话已断开' }}</span>
        <button class="overlay-reconnect-btn" @click.stop="reconnect">
          重新连接
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.xterm-pane-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
}

[data-theme="dark"] .xterm-pane-wrapper {
  background-color: #18181b;
}

.xterm-container {
  flex: 1;
  width: 100%;
  height: 100%;
  padding: 6px 10px;
  overflow: hidden;
  box-sizing: border-box;
  background-color: transparent;
}

:deep(.xterm) {
  height: 100%;
  padding: 0;
  background-color: transparent !important;
}

:deep(.xterm-viewport) {
  background-color: transparent !important;
  overflow-y: auto !important;
}

:deep(.xterm-screen) {
  background-color: transparent !important;
}

:deep(.xterm:focus),
:deep(.xterm.focus),
:deep(.xterm-helper-textarea:focus) {
  outline: none !important;
  border: none !important;
}

.xterm-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.xterm-overlay-card {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-card, #1e293b);
  border: 1px solid var(--card-border, #334155);
  border-radius: 8px;
  padding: 8px 14px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.overlay-icon {
  font-size: 1.1rem;
}

.overlay-text {
  font-size: 0.8rem;
  color: var(--text-main, #f1f5f9);
}

.overlay-reconnect-btn {
  background: #0284c7;
  color: #ffffff;
  border: none;
  border-radius: 4px;
  padding: 4px 10px;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s ease;
}

.overlay-reconnect-btn:hover {
  background: #0369a1;
}

[data-theme="light"] .xterm-overlay {
  background: rgba(255, 255, 255, 0.6);
}

[data-theme="light"] .xterm-overlay-card {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

[data-theme="light"] .overlay-text {
  color: #0f172a;
}
</style>
