// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { reactive } from "vue";
import { useToast } from "./useToast";

export interface TrackedTaskInfo {
  paneId: string;
  command: string;
  startTime: number;
  hasError: boolean;
}

export interface TaskCompletionResult {
  paneId: string;
  command: string;
  durationSec: number;
  success: boolean;
  notified: boolean;
}

// Default long-running command notification threshold: 10 seconds
const DEFAULT_THRESHOLD_MS = 10000;
let notificationThresholdMs = DEFAULT_THRESHOLD_MS;

const trackedTasks = reactive<Map<string, TrackedTaskInfo>>(new Map());

// Regex to detect shell prompt return (command completion)
const SHELL_PROMPT_REGEX = /(?:[$#>]\s*$|\(base\)\s*[$#>])/;

// Common command failure keywords
const ERROR_PATTERNS = [
  /\b(?:command\s+not\s+found|syntax\s+error|permission\s+denied)\b/i,
  /\b(?:error:\s+|failed\s+with\s+exit\s+code\s+[1-9]\d*)\b/i,
  /\b(?:fatal:\s+|panic:\s+|build\s+failed)\b/i,
  /\b(?:npm\s+ERR!|ELIFECYCLE)\b/i,
];

export function useTerminalNotifier() {
  const { showToast } = useToast();

  function setThreshold(ms: number): void {
    if (ms >= 0) {
      notificationThresholdMs = ms;
    }
  }

  function getThreshold(): number {
    return notificationThresholdMs;
  }

  /**
   * Start timing monitor for command in pane
   */
  function startTracking(paneId: string, command: string): void {
    if (!paneId || !command || !command.trim()) return;
    trackedTasks.set(paneId, {
      paneId,
      command: command.trim(),
      startTime: Date.now(),
      hasError: false,
    });
  }

  /**
   * Record output chunk, detect errors and settle when prompt returns
   */
  function feedOutput(paneId: string, chunk: string): TaskCompletionResult | null {
    if (!paneId || !chunk) return null;
    const task = trackedTasks.get(paneId);
    if (!task) return null;

    // Check for error keywords
    for (const pattern of ERROR_PATTERNS) {
      if (pattern.test(chunk)) {
        task.hasError = true;
        break;
      }
    }

    // Check if shell prompt returned
    if (SHELL_PROMPT_REGEX.test(chunk)) {
      const durationMs = Date.now() - task.startTime;
      const durationSec = Math.round(durationMs / 1000);
      const isSuccess = !task.hasError;
      const isLongTask = durationMs >= notificationThresholdMs;

      trackedTasks.delete(paneId);

      if (isLongTask) {
        const cmdSummary = task.command.length > 25 ? task.command.slice(0, 22) + "..." : task.command;
        if (isSuccess) {
          showToast(`Task '${cmdSummary}' completed (${durationSec}s)`, "success");
        } else {
          showToast(`Task '${cmdSummary}' finished with errors (${durationSec}s)`, "error");
        }
      }

      return {
        paneId,
        command: task.command,
        durationSec,
        success: isSuccess,
        notified: isLongTask,
      };
    }

    return null;
  }

  function getActiveTrackedTask(paneId: string): TrackedTaskInfo | null {
    return trackedTasks.get(paneId) || null;
  }

  function clearTracked(paneId: string): void {
    trackedTasks.delete(paneId);
  }

  function _clearAllForTest(): void {
    trackedTasks.clear();
    notificationThresholdMs = DEFAULT_THRESHOLD_MS;
  }

  return {
    startTracking,
    feedOutput,
    setThreshold,
    getThreshold,
    getActiveTrackedTask,
    clearTracked,
    _clearAllForTest,
  };
}
