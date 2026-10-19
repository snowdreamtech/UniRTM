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

// 默认耗时任务提醒阈值：10 秒
const DEFAULT_THRESHOLD_MS = 10000;
let notificationThresholdMs = DEFAULT_THRESHOLD_MS;

const trackedTasks = reactive<Map<string, TrackedTaskInfo>>(new Map());

// 正则检测 shell 提示符（表示命令执行完毕返回交互状态）
const SHELL_PROMPT_REGEX = /(?:[$#>]\s*$|\(base\)\s*[$#>])/;

// 常见命令失败关键字嗅探
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
   * 启动对某个窗格中命令的计时监控
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
   * 记录输出 chunk，侦测错误并在检测到提示符返回时结算
   */
  function feedOutput(paneId: string, chunk: string): TaskCompletionResult | null {
    if (!paneId || !chunk) return null;
    const task = trackedTasks.get(paneId);
    if (!task) return null;

    // 检查是否有错误关键字
    for (const pattern of ERROR_PATTERNS) {
      if (pattern.test(chunk)) {
        task.hasError = true;
        break;
      }
    }

    // 检查是否返回 shell 提示符
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
