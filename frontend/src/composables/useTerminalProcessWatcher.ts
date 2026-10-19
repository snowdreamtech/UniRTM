// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { reactive } from "vue";

export interface DetectedProcessInfo {
  type: string;
  name: string;
  icon: string;
  color: string;
  startTime: number;
}

interface ProcessRule {
  type: string;
  name: string;
  icon: string;
  color: string;
  regex: RegExp;
}

const PROCESS_RULES: ProcessRule[] = [
  {
    type: "node",
    name: "Node",
    icon: "codicon-symbol-property",
    color: "#22c55e",
    regex: /(?:^|[\s[])(?:node|npm|pnpm|yarn|bun|vite|nuxt|next|nest|ts-node|tsx)(?:[\s\]./]|$)/i,
  },
  {
    type: "python",
    name: "Python",
    icon: "codicon-symbol-variable",
    color: "#3b82f6",
    regex: /(?:^|[\s[])(?:python3?|pytest|uvicorn|flask|django|gunicorn|pip)(?:[\s\]./]|$)/i,
  },
  {
    type: "go",
    name: "Go",
    icon: "codicon-symbol-structure",
    color: "#06b6d4",
    regex: /\bgo\s+(?:run|test|build)\b|\bair\b|\bdlv\b/i,
  },
  {
    type: "rust",
    name: "Rust",
    icon: "codicon-gear",
    color: "#f97316",
    regex: /\bcargo(?:\s|$|\.|\/)|rustc\b/i,
  },
  {
    type: "docker",
    name: "Docker",
    icon: "codicon-server",
    color: "#38bdf8",
    regex: /\bdocker\b|\bkubectl\b|\bpodman\b|\bcompose\b/i,
  },
  {
    type: "git",
    name: "Git",
    icon: "codicon-git-commit",
    color: "#a855f7",
    regex: /\bgit\s+(?:push|pull|clone|fetch|merge|rebase|commit|checkout|switch)\b/i,
  },
  {
    type: "build",
    name: "Build",
    icon: "codicon-tools",
    color: "#eab308",
    regex: /\bmake\b|\bcmake\b|\bninja\b|\bmvn\b|\bgradle\b/i,
  },
];

// 正则检测 shell 提示符返回（表示命令已结束回到空闲态）
const SHELL_PROMPT_REGEX = /(?:[$#>]\s*$|\(base\)\s*[$#>])/;

const activeProcesses = reactive<Map<string, DetectedProcessInfo>>(new Map());
const customTitles = reactive<Map<string, string>>(new Map());

/**
 * 从文本流中嗅探当前运行的进程
 */
export function detectProcessFromChunk(text: string): DetectedProcessInfo | null {
  if (!text) return null;

  for (const rule of PROCESS_RULES) {
    if (rule.regex.test(text)) {
      return {
        type: rule.type,
        name: rule.name,
        icon: rule.icon,
        color: rule.color,
        startTime: Date.now(),
      };
    }
  }

  return null;
}

export function useTerminalProcessWatcher() {
  /**
   * 记录输出并智能更新对应 pane 的活动进程
   */
  function recordChunk(paneId: string, chunk: string): void {
    if (!paneId || !chunk) return;

    // 1. 如果包含 shell 提示符，说明前台进程可能已退回交互状态
    if (SHELL_PROMPT_REGEX.test(chunk)) {
      activeProcesses.delete(paneId);
      return;
    }

    // 2. 匹配规则
    const detected = detectProcessFromChunk(chunk);
    if (detected) {
      const existing = activeProcesses.get(paneId);
      // 如果类型相同，保留原有的 startTime 以准确统计任务耗时
      if (existing && existing.type === detected.type) {
        return;
      }
      activeProcesses.set(paneId, detected);
    }
  }

  /**
   * 获取指定 pane 正在运行的进程信息
   */
  function getActiveProcess(paneId: string): DetectedProcessInfo | null {
    return activeProcesses.get(paneId) || null;
  }

  /**
   * 设置自定义终端标题
   */
  function setCustomTitle(id: string, title: string): void {
    if (!id) return;
    if (!title || !title.trim()) {
      customTitles.delete(id);
    } else {
      customTitles.set(id, title.trim());
    }
  }

  /**
   * 获取自定义终端标题
   */
  function getCustomTitle(id: string): string {
    return customTitles.get(id) || "";
  }

  /**
   * 计算综合有效标题
   */
  function getEffectiveTitle(id: string, defaultName: string): string {
    const custom = customTitles.get(id);
    if (custom) return custom;

    const process = activeProcesses.get(id);
    if (process) {
      return `${process.name} (${defaultName})`;
    }

    return defaultName;
  }

  function _clearAllForTest(): void {
    activeProcesses.clear();
    customTitles.clear();
  }

  return {
    recordChunk,
    getActiveProcess,
    setCustomTitle,
    getCustomTitle,
    getEffectiveTitle,
    activeProcesses,
    customTitles,
    _clearAllForTest,
  };
}
