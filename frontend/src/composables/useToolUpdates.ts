import { ref, computed } from "vue";

export interface ToolUpdateItem {
  id: string;
  name: string;
  current_version: string;
  latest_version: string;
  has_update: boolean;
  pinned: boolean;
  pin_reason?: string;
  homepage?: string;
  error?: string;
}

const updateResults = ref<ToolUpdateItem[]>([]);
const isChecking = ref(false);
const lastCheckedAt = ref<number | null>(null);

export function useToolUpdates() {
  const updatableList = computed(() =>
    updateResults.value.filter((r) => r.has_update && !r.pinned)
  );

  const updatableCount = computed(() => updatableList.value.length);

  const updatableToolIds = computed(() => {
    const ids = new Set<string>();
    for (const r of updatableList.value) {
      const keys = [r.id, r.name].filter(Boolean);
      for (const k of keys) {
        ids.add(k);
        if (k.includes(":")) {
          const bare = k.split(":").slice(1).join(":");
          ids.add(bare);
          if (bare.includes("/")) {
            const short = bare.split("/").pop();
            if (short) ids.add(short);
          }
        }
        if (k.includes("/")) {
          const short = k.split("/").pop();
          if (short) ids.add(short);
        }
      }
    }
    return ids;
  });

  function hasUpdate(toolId: string): boolean {
    if (!toolId) return false;
    if (updatableToolIds.value.has(toolId)) return true;
    if (toolId.includes(":")) {
      const bare = toolId.split(":").slice(1).join(":");
      if (updatableToolIds.value.has(bare)) return true;
      if (bare.includes("/")) {
        const short = bare.split("/").pop();
        if (short && updatableToolIds.value.has(short)) return true;
      }
    }
    if (toolId.includes("/")) {
      const short = toolId.split("/").pop();
      if (short && updatableToolIds.value.has(short)) return true;
    }
    return false;
  }

  function syncFromToolsList(tools: any[]) {
    if (!tools || !Array.isArray(tools)) return;
    const currentResults = [...updateResults.value];
    let changed = false;

    const semverGt = (a: string, b: string) => {
      const pa = a.replace(/^[vV]+/, "").split(".").map((n) => parseInt(n, 10));
      const pb = b.replace(/^[vV]+/, "").split(".").map((n) => parseInt(n, 10));
      for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
        const na = isNaN(pa[i]) ? 0 : pa[i];
        const nb = isNaN(pb[i]) ? 0 : pb[i];
        if (na !== nb) return na > nb;
      }
      return false;
    };

    for (const tool of tools) {
      if (!tool || tool.pinned) continue;
      const cur = tool.active_version || tool.global_version;
      const installed = tool.installed_versions;
      if (!cur || !installed || !Array.isArray(installed) || installed.length === 0) continue;

      const higherInstalled = installed.filter((v: string) => v && semverGt(v, cur));
      if (higherInstalled.length > 0) {
        higherInstalled.sort((a: string, b: string) => (semverGt(a, b) ? -1 : semverGt(b, a) ? 1 : 0));
        const bestVer = higherInstalled[0];
        const toolId = tool.id || tool.name;
        const existingIdx = currentResults.findIndex((r) => r.id === toolId || r.name === toolId);
        if (existingIdx !== -1) {
          if (!currentResults[existingIdx].has_update) {
            currentResults[existingIdx] = {
              ...currentResults[existingIdx],
              has_update: true,
              latest_version: currentResults[existingIdx].latest_version || bestVer,
            };
            changed = true;
          }
        } else {
          currentResults.push({
            id: toolId,
            name: tool.name || toolId,
            current_version: cur,
            latest_version: bestVer,
            has_update: true,
            pinned: false,
          });
          changed = true;
        }
      }
    }

    if (changed) {
      updateResults.value = currentResults;
    }
  }

  async function checkUpdates(isGlobal: boolean = false): Promise<ToolUpdateItem[]> {
    isChecking.value = true;
    try {
      const res = await fetch(`/api/tools/check-updates?global=${isGlobal}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      const rawList = data.results || data.updates || [];
      updateResults.value = rawList.map((item: any) => ({
        id: item.id || item.tool || "",
        name: item.name || item.tool || "",
        current_version: item.current_version || "",
        latest_version: item.latest_version || "",
        has_update: Boolean(item.has_update),
        pinned: Boolean(item.pinned ?? item.is_pinned ?? false),
        pin_reason: item.pin_reason || "",
        homepage: item.homepage,
        error: item.error,
      }));
      lastCheckedAt.value = Date.now();
      return updateResults.value;
    } catch (err) {
      console.error("Failed to check tool updates:", err);
      return [];
    } finally {
      isChecking.value = false;
    }
  }

  function markToolUpdated(toolId: string) {
    updateResults.value = updateResults.value.map((item) =>
      item.id === toolId || item.name === toolId
        ? { ...item, has_update: false }
        : item
    );
  }

  function setUpdateResults(items: ToolUpdateItem[]) {
    updateResults.value = items;
  }

  function clearUpdates() {
    updateResults.value = [];
  }

  return {
    updateResults,
    isChecking,
    lastCheckedAt,
    updatableList,
    updatableCount,
    updatableToolIds,
    hasUpdate,
    syncFromToolsList,
    checkUpdates,
    markToolUpdated,
    setUpdateResults,
    clearUpdates,
  };
}
