// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref } from "vue";

export interface GuardIssue {
  file_path: string;
  line: number;
  column: number;
  rule_id: string;
  category: string;
  phase: string;
  severity: "error" | "warning" | "info";
  message: string;
  fixable: boolean;
  fixed?: boolean;
}

export interface GuardStatus {
  mode: string;
  scanned: number;
  clean: boolean;
  errors_count: number;
  warnings_count: number;
  fixable_count: number;
  issues: GuardIssue[];
  logs?: string[];
}

export interface ConfigFileMeta {
  name: string;
  title: string;
  description: string;
  category: string;
  phase: string;
  exists: boolean;
  path: string;
}

export interface GuardRuleMeta {
  id: string;
  category: string;
  phase: string;
  extensions: string[] | null;
}

export interface GuardSlotMeta {
  id: string;
  description: string;
  category: string;
  phase: string;
  default_tool_name: string;
  default_command: string;
  fix_command?: string;
  default_extensions?: string[];
  standard_config_files?: string[];
}

const status = ref<GuardStatus>({
  mode: "staged",
  scanned: 0,
  clean: true,
  errors_count: 0,
  warnings_count: 0,
  fixable_count: 0,
  issues: [],
  logs: [],
});

const builtinRules = ref<GuardRuleMeta[]>([]);
const slots = ref<GuardSlotMeta[]>([]);
const configFiles = ref<ConfigFileMeta[]>([]);
const isLoading = ref(false);
const isFixing = ref(false);

export function useGuard() {
  async function fetchStatus(mode = "staged"): Promise<GuardStatus | null> {
    isLoading.value = true;
    try {
      const res = await fetch(`/api/guard/status?mode=${encodeURIComponent(mode)}`);
      if (res.ok) {
        const data = await res.json();
        status.value = data;
        return data;
      }
    } catch (e) {
      console.error("Failed to fetch guard status:", e);
    } finally {
      isLoading.value = false;
    }
    return null;
  }

  async function runFix(mode = "staged", files: string[] = []): Promise<boolean> {
    isFixing.value = true;
    try {
      const res = await fetch("/api/guard/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ mode, fix: true, files }),
      });
      if (res.ok) {
        // Re-fetch status to get updated clean metrics
        await fetchStatus(mode);
        return true;
      }
    } catch (e) {
      console.error("Failed to run guard fix:", e);
    } finally {
      isFixing.value = false;
    }
    return false;
  }

  async function fetchRules() {
    try {
      const res = await fetch("/api/guard/rules");
      if (res.ok) {
        const data = await res.json();
        builtinRules.value = data.builtin_rules || [];
        slots.value = data.slots || [];
        configFiles.value = data.config_files || [];
      }
    } catch (e) {
      console.error("Failed to fetch guard rules:", e);
    }
  }

  return {
    status,
    builtinRules,
    slots,
    configFiles,
    isLoading,
    isFixing,
    fetchStatus,
    runFix,
    fetchRules,
  };
}
