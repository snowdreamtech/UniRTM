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

export interface GuardRuleReport {
  rule_id?: string;
  rule_name: string;
  category: string;
  phase: string;
  duration_ms: number;
  duration?: number;
  issues_count: number;
  errors_count?: number;
  warnings_count?: number;
  fixed_count?: number;
  status: "passed" | "warn" | "error" | "fixed" | "fail";
}

export interface GuardStatus {
  mode: string;
  scanned: number;
  clean: boolean;
  errors_count: number;
  warnings_count: number;
  fixable_count: number;
  rules_count?: number;
  duration_ms?: number;
  issues: GuardIssue[];
  rule_reports?: GuardRuleReport[];
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

export interface ToolSlotItem {
  id: string;
  description: string;
  category: string;
  phase: string;
  default_tool_name: string;
  default_command: string;
  fix_command?: string;
  default_extensions?: string[];
  standard_config_files?: string[];
  enabled?: boolean;
}

export type GuardSlotMeta = ToolSlotItem;

export interface SlotConfigInput {
  id: string;
  tool?: string;
  command?: string;
  fix_command?: string;
  extensions?: string[];
  category?: string;
  phase?: string;
  enabled?: boolean;
}

const status = ref<GuardStatus>({
  mode: "staged",
  scanned: 0,
  clean: true,
  errors_count: 0,
  warnings_count: 0,
  fixable_count: 0,
  rules_count: 38,
  duration_ms: 0,
  issues: [],
  logs: [],
});

const builtinRules = ref<GuardRuleMeta[]>([]);
const slots = ref<ToolSlotItem[]>([]);
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

  async function runCheck(opts?: { mode?: "staged" | "all"; files?: string[] }): Promise<boolean> {
    isLoading.value = true;
    const mode = opts?.mode || "staged";
    try {
      const res = await fetch("/api/guard/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ mode, fix: false, files: opts?.files || [] }),
      });
      if (res.ok) {
        await fetchStatus(mode);
        return true;
      }
    } catch (e) {
      console.error("Failed to run guard check:", e);
    } finally {
      isLoading.value = false;
    }
    return false;
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

  async function fetchSlots(): Promise<ToolSlotItem[]> {
    try {
      const res = await fetch("/api/guard/slots");
      if (res.ok) {
        const data = await res.json();
        if (data.slots) {
          slots.value = data.slots;
          return data.slots;
        }
      }
    } catch (e) {
      console.error("Failed to fetch tool slots:", e);
    }
    // Fallback to fetchRules if /api/guard/slots is not directly populated
    await fetchRules();
    return slots.value;
  }

  async function saveSlotConfig(slot: SlotConfigInput): Promise<boolean> {
    try {
      const res = await fetch("/api/guard/slots", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(slot),
      });
      if (res.ok) {
        const data = await res.json();
        if (data.slots) {
          slots.value = data.slots;
        }
        await fetchRules();
        return true;
      }
    } catch (e) {
      console.error("Failed to save slot config:", e);
    }
    return false;
  }

  async function deleteSlotConfig(slotId: string): Promise<boolean> {
    try {
      const res = await fetch(`/api/guard/slots?id=${encodeURIComponent(slotId)}`, {
        method: "DELETE",
      });
      if (res.ok) {
        const data = await res.json();
        if (data.slots) {
          slots.value = data.slots;
        }
        await fetchRules();
        return true;
      }
    } catch (e) {
      console.error("Failed to delete slot config:", e);
    }
    return false;
  }

  async function fetchRawTomlConfig(): Promise<string> {
    try {
      const res = await fetch("/api/guard/toml");
      if (res.ok) {
        const data = await res.json();
        return data.content || "";
      }
    } catch (e) {
      console.error("Failed to fetch raw TOML config:", e);
    }
    return "";
  }

  async function saveRawTomlConfig(content: string): Promise<boolean> {
    try {
      const res = await fetch("/api/guard/toml", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content }),
      });
      if (res.ok) {
        await fetchRules();
        await fetchStatus(status.value.mode);
        return true;
      }
    } catch (e) {
      console.error("Failed to save raw TOML config:", e);
    }
    return false;
  }

  return {
    // Standard and alias properties
    status,
    guardStatus: status,
    builtinRules,
    guardBuiltinRules: builtinRules,
    slots,
    guardSlots: slots,
    configFiles,
    guardConfigFiles: configFiles,
    isLoading,
    isChecking: isLoading,
    isFixing,
    // Operations
    fetchStatus,
    runCheck,
    runFix,
    fetchRules,
    fetchSlots,
    saveSlotConfig,
    deleteSlotConfig,
    fetchRawTomlConfig,
    saveRawTomlConfig,
  };
}
