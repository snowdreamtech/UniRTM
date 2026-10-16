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
      if (r.id) ids.add(r.id);
      if (r.name) ids.add(r.name);
      if (r.id && r.id.includes(":")) {
        ids.add(r.id.split(":").slice(1).join(":"));
      }
    }
    return ids;
  });

  function hasUpdate(toolId: string): boolean {
    if (!toolId) return false;
    return updatableToolIds.value.has(toolId);
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
    checkUpdates,
    markToolUpdated,
    setUpdateResults,
    clearUpdates,
  };
}
