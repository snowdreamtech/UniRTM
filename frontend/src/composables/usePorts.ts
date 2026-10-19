// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import type {
  PortForwardItem,
  PortForwardPayload,
  SystemListeningPort,
  PortStatus,
  PortVisibility,
  PortProtocol,
} from "../types/ports";
import { recognizePortService } from "../utils/portServiceRecognizer";
import { useToast } from "./useToast";

const STORAGE_KEY = "unirtm_user_forwarded_ports";

const ports = ref<PortForwardItem[]>([]);
const listeningPorts = ref<SystemListeningPort[]>([]);
const userDefinedPorts = ref<Map<number, Partial<PortForwardItem>>>(new Map());
const userStoppedPorts = ref<Set<number>>(new Set());
const userIgnoredAutoPorts = ref<Set<number>>(new Set());
const isLoading = ref<boolean>(false);
const searchQuery = ref<string>("");
const autoDetectEnabled = ref<boolean>(true);

// Load persisted user-defined ports from localStorage
function loadPersistedPorts() {
  if (typeof localStorage === "undefined") return;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const items: Array<Partial<PortForwardItem> & { port: number }> = JSON.parse(raw);
      userDefinedPorts.value.clear();
      for (const item of items) {
        if (typeof item.port === "number") {
          userDefinedPorts.value.set(item.port, item);
        }
      }
    }
  } catch (e) {
    console.error("Failed to load persisted forwarded ports:", e);
  }
}

// Save persisted user-defined ports to localStorage
function savePersistedPorts() {
  if (typeof localStorage === "undefined") return;
  try {
    const arr = Array.from(userDefinedPorts.value.values());
    localStorage.setItem(STORAGE_KEY, JSON.stringify(arr));
  } catch (e) {
    console.error("Failed to save forwarded ports:", e);
  }
}

export function usePorts() {
  const { showToast } = useToast();

  loadPersistedPorts();

  // Active listening ports count
  const activePortsCount = computed(() => {
    return ports.value.filter((p) => p.status === "active").length;
  });

  // Filtered ports list based on search query
  const filteredPorts = computed(() => {
    const q = searchQuery.value.trim().toLowerCase();
    if (!q) return ports.value;
    return ports.value.filter((p) => {
      return (
        String(p.port).includes(q) ||
        String(p.localPort).includes(q) ||
        p.label.toLowerCase().includes(q) ||
        (p.process && p.process.toLowerCase().includes(q)) ||
        (p.commandLine && p.commandLine.toLowerCase().includes(q)) ||
        p.forwardAddress.toLowerCase().includes(q)
      );
    });
  });

  /**
   * Helper to find next available local port to avoid collision.
   */
  function findAvailableLocalPort(preferredPort: number, excludeId?: string): number {
    const usedPorts = new Set<number>();
    for (const p of ports.value) {
      if (p.id !== excludeId) {
        usedPorts.add(p.localPort);
      }
    }
    let candidate = preferredPort;
    while (usedPorts.has(candidate)) {
      candidate++;
    }
    return candidate;
  }

  /**
   * Merge system listening ports and user-defined ports into reactive ports list.
   */
  function syncPortsList() {
    const mergedMap = new Map<number, PortForwardItem>();
    const listeningMap = new Map<number, SystemListeningPort>();

    for (const lp of listeningPorts.value) {
      listeningMap.set(lp.port, lp);
    }

    // 1. Process User-Defined Ports (highest precedence)
    for (const [portNum, userConfig] of userDefinedPorts.value.entries()) {
      const isListening = listeningMap.has(portNum);
      const isStopped = userStoppedPorts.value.has(portNum);
      const lp = listeningMap.get(portNum);

      const status: PortStatus = isStopped ? "inactive" : isListening ? "active" : "inactive";
      const recognized = recognizePortService(portNum, lp?.process, lp?.commandLine);

      const localPort = userConfig.localPort || portNum;
      const protocol: PortProtocol = userConfig.protocol || recognized.defaultProtocol || "http";
      const visibility: PortVisibility = userConfig.visibility || "private";
      const label = userConfig.label || recognized.label;

      const item: PortForwardItem = {
        id: `port-${portNum}`,
        port: portNum,
        localPort,
        label,
        protocol,
        source: "user",
        status,
        visibility,
        process: lp?.process,
        commandLine: lp?.commandLine,
        pid: lp?.pid,
        forwardAddress: `localhost:${localPort}`,
        previewUrl: `${protocol}://localhost:${localPort}/`,
        icon: userConfig.icon || recognized.icon,
      };

      mergedMap.set(portNum, item);
    }

    // 2. Process Auto-Discovered Listening Ports
    if (autoDetectEnabled.value) {
      for (const lp of listeningPorts.value) {
        if (userDefinedPorts.value.has(lp.port)) continue;
        if (userIgnoredAutoPorts.value.has(lp.port)) continue;

        const isStopped = userStoppedPorts.value.has(lp.port);
        const recognized = recognizePortService(lp.port, lp.process, lp.commandLine);
        const localPort = lp.port;
        const protocol: PortProtocol = recognized.defaultProtocol || "http";

        const item: PortForwardItem = {
          id: `port-${lp.port}`,
          port: lp.port,
          localPort,
          label: recognized.label,
          protocol,
          source: "auto",
          status: isStopped ? "inactive" : "active",
          visibility: "private",
          process: lp.process,
          commandLine: lp.commandLine,
          pid: lp.pid,
          forwardAddress: `localhost:${localPort}`,
          previewUrl: `${protocol}://localhost:${localPort}/`,
          icon: recognized.icon,
        };

        mergedMap.set(lp.port, item);
      }
    }

    // Sort by port number ascending
    ports.value = Array.from(mergedMap.values()).sort((a, b) => a.port - b.port);
  }

  /**
   * Fetch system listening ports from backend API.
   */
  async function fetchPorts(): Promise<PortForwardItem[]> {
    isLoading.value = true;
    try {
      const res = await fetch("/api/ports/list");
      if (res.ok) {
        const data = await res.json();
        if (data.success && Array.isArray(data.listening)) {
          listeningPorts.value = data.listening;
        }
      }
    } catch (err: any) {
      console.warn("fetchPorts request failed, falling back to cached state:", err);
    } finally {
      isLoading.value = false;
      syncPortsList();
    }
    return ports.value;
  }

  /**
   * Create or update a port forwarding item (User Forward).
   */
  async function addPort(payload: PortForwardPayload): Promise<boolean> {
    const port = Number(payload.port);
    if (!port || port <= 0 || port > 65535) {
      showToast("端口号无效，请输入 1-65535 之间的有效端口", "error");
      return false;
    }

    const assignedLocalPort = payload.localPort || findAvailableLocalPort(port, `port-${port}`);
    const recognized = recognizePortService(port);

    const userItem: Partial<PortForwardItem> & { port: number } = {
      port,
      localPort: assignedLocalPort,
      label: payload.label?.trim() || recognized.label,
      protocol: payload.protocol || recognized.defaultProtocol,
      visibility: payload.visibility || "private",
      icon: recognized.icon,
    };

    userDefinedPorts.value.set(port, userItem);
    userStoppedPorts.value.delete(port);
    userIgnoredAutoPorts.value.delete(port);
    savePersistedPorts();

    // Call backend API if available
    try {
      await fetch("/api/ports/forward", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(userItem),
      });
    } catch (_) {}

    syncPortsList();
    showToast(`端口 ${port} 转发已添加`, "success");
    return true;
  }

  /**
   * Update an existing port configuration.
   */
  async function updatePort(
    port: number,
    updates: {
      label?: string;
      localPort?: number;
      protocol?: PortProtocol;
      visibility?: PortVisibility;
    }
  ): Promise<boolean> {
    const existing = userDefinedPorts.value.get(port) || {};
    const updated: Partial<PortForwardItem> & { port: number } = {
      ...existing,
      port,
      ...(updates.label !== undefined && { label: updates.label }),
      ...(updates.localPort !== undefined && { localPort: updates.localPort }),
      ...(updates.protocol !== undefined && { protocol: updates.protocol }),
      ...(updates.visibility !== undefined && { visibility: updates.visibility }),
    };

    userDefinedPorts.value.set(port, updated);
    savePersistedPorts();

    try {
      await fetch("/api/ports/update", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updated),
      });
    } catch (_) {}

    syncPortsList();
    showToast(`端口 ${port} 配置已更新`, "success");
    return true;
  }

  /**
   * Stop/Pause forwarding on a port.
   */
  async function stopPort(port: number): Promise<boolean> {
    userStoppedPorts.value.add(port);
    try {
      await fetch("/api/ports/stop", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ port }),
      });
    } catch (_) {}

    syncPortsList();
    showToast(`已停止端口 ${port} 的转发`, "info");
    return true;
  }

  /**
   * Start/Resume forwarding on a stopped port.
   */
  async function startPort(port: number): Promise<boolean> {
    userStoppedPorts.value.delete(port);
    try {
      await fetch("/api/ports/forward", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ port }),
      });
    } catch (_) {}

    syncPortsList();
    showToast(`已恢复端口 ${port} 的转发`, "success");
    return true;
  }

  /**
   * Remove/Delete a port forward item.
   */
  async function removePort(port: number): Promise<boolean> {
    if (userDefinedPorts.value.has(port)) {
      userDefinedPorts.value.delete(port);
      savePersistedPorts();
    }
    userIgnoredAutoPorts.value.add(port);
    userStoppedPorts.value.delete(port);

    try {
      await fetch("/api/ports/unforward", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ port }),
      });
    } catch (_) {}

    syncPortsList();
    showToast(`已移除端口 ${port}`, "info");
    return true;
  }

  /**
   * Terminal Output Sniffer: automatically detected port from terminal output text.
   */
  function registerDetectedPort(
    port: number,
    options?: { label?: string; protocol?: PortProtocol }
  ) {
    if (!port || port <= 0 || port > 65535) return;
    if (userIgnoredAutoPorts.value.has(port)) return;

    const exists = ports.value.find((p) => p.port === port);
    if (exists && exists.status === "active") return;

    // Check if in listening ports, else add virtual listening port
    const inListening = listeningPorts.value.find((lp) => lp.port === port);
    if (!inListening) {
      listeningPorts.value.push({
        port,
        protocol: options?.protocol || "http",
        localAddress: `127.0.0.1:${port}`,
        process: options?.label || "Web Server",
      });
    }

    syncPortsList();
    showToast(`侦测到本地端口 ${port} 已就绪，已自动加入端口列表`, "info");
  }

  /**
   * Copy forwarded address to clipboard.
   */
  async function copyAddress(item: PortForwardItem): Promise<boolean> {
    try {
      const url = item.previewUrl || `http://${item.forwardAddress}/`;
      await navigator.clipboard.writeText(url);
      showToast(`已复制地址: ${url}`, "success");
      return true;
    } catch (e) {
      showToast("复制地址失败", "error");
      return false;
    }
  }

  /**
   * Copy port number to clipboard.
   */
  async function copyPort(port: number): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(String(port));
      showToast(`已复制端口号: ${port}`, "success");
      return true;
    } catch (e) {
      showToast("复制端口号失败", "error");
      return false;
    }
  }

  /**
   * Open port in system browser or in-IDE tab.
   */
  function openPort(item: PortForwardItem) {
    const url = item.previewUrl || `http://${item.forwardAddress}/`;
    window.open(url, "_blank");
  }

  return {
    ports,
    filteredPorts,
    listeningPorts,
    activePortsCount,
    isLoading,
    searchQuery,
    autoDetectEnabled,
    fetchPorts,
    addPort,
    updatePort,
    stopPort,
    startPort,
    removePort,
    registerDetectedPort,
    copyAddress,
    copyPort,
    openPort,
    findAvailableLocalPort,
  };
}
