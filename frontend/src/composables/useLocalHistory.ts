// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref } from "vue";
import { useWorkspaceBus } from "./useWorkspaceBus";

export interface HistorySnapshot {
  id: string;
  path: string;
  timestamp: number;
  content: string;
  size: number;
  label?: string;
}

const MAX_SNAPSHOTS_PER_FILE = 20;
const snapshotsMap = ref<Map<string, HistorySnapshot[]>>(new Map());

const activeHistoryModal = ref<{
  visible: boolean;
  path: string;
}>({
  visible: false,
  path: "",
});

export function useLocalHistory() {
  /**
   * Records a snapshot for the specified file.
   * Skips recording if content is unchanged from latest snapshot.
   */
  function recordSnapshot(path: string, content: string, label?: string): HistorySnapshot | null {
    if (!path) return null;

    const list = snapshotsMap.value.get(path) || [];
    const latest = list[0];

    if (latest && latest.content === content) {
      return latest;
    }

    const snapshot: HistorySnapshot = {
      id: `hist-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      path,
      timestamp: Date.now(),
      content,
      size: new Blob([content]).size,
      label,
    };

    const nextList = [snapshot, ...list].slice(0, MAX_SNAPSHOTS_PER_FILE);
    const nextMap = new Map(snapshotsMap.value);
    nextMap.set(path, nextList);
    snapshotsMap.value = nextMap;

    return snapshot;
  }

  /**
   * Retrieves snapshots for given file path, sorted descending by timestamp.
   */
  function getSnapshots(path: string): HistorySnapshot[] {
    if (!path) return [];
    return snapshotsMap.value.get(path) || [];
  }

  /**
   * Clears snapshots for a given path or all if omitted.
   */
  function clearSnapshots(path?: string): void {
    if (path) {
      const nextMap = new Map(snapshotsMap.value);
      nextMap.delete(path);
      snapshotsMap.value = nextMap;
    } else {
      snapshotsMap.value = new Map();
    }
  }

  /**
   * Opens the local history modal for a target file.
   */
  function openLocalHistory(path: string): void {
    if (!path) return;
    activeHistoryModal.value = {
      visible: true,
      path,
    };
  }

  function closeLocalHistory(): void {
    activeHistoryModal.value = {
      visible: false,
      path: "",
    };
  }

  /**
   * Compares the selected snapshot with current working file via main editor Diff tab.
   */
  function compareWithCurrent(snapshot: HistorySnapshot, currentContent: string): void {
    const { openDiff } = useWorkspaceBus();
    const fileName = snapshot.path.split("/").pop() || snapshot.path;
    const timeStr = new Date(snapshot.timestamp).toLocaleTimeString();

    openDiff({
      path: snapshot.path,
      name: `${fileName} (本地快照 ${timeStr} ↔ 当前版本)`,
      originalContent: snapshot.content,
      modifiedContent: currentContent,
      isStaged: false,
    });
  }

  return {
    snapshotsMap,
    activeHistoryModal,
    recordSnapshot,
    getSnapshots,
    clearSnapshots,
    openLocalHistory,
    closeLocalHistory,
    compareWithCurrent,
  };
}
