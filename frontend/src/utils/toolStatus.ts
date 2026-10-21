// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export type ToolDisplayState = "ready" | "upgradable" | "missing" | "unconfigured";

export interface ToolStatusItem {
  id: string;
  name?: string;
  installed_versions?: string[];
  active_version?: string;
  global_version?: string;
  pinned?: boolean;
}

export interface ToolUpdateCandidate {
  id: string;
  name?: string;
  latest_version?: string;
  has_update?: boolean;
  pinned?: boolean;
}

export interface ToolStatusResult {
  state: ToolDisplayState;
  configuredVersion: string;
  upgradableVersion: string | null;
  isInstalled: boolean;
}

function cleanVer(v: string | undefined | null): string {
  if (!v) return "";
  return v.trim().replace(/^[vV]+/, "");
}

/**
 * 比较两个语义化版本（降序：v1 > v2 返回负数，v1 < v2 返回正数，相等返回 0）
 */
export function compareSemverDesc(v1: string, v2: string): number {
  const p1 = cleanVer(v1)
    .split(".")
    .map((n) => parseInt(n, 10) || 0);
  const p2 = cleanVer(v2)
    .split(".")
    .map((n) => parseInt(n, 10) || 0);
  const maxLen = Math.max(p1.length, p2.length);

  for (let i = 0; i < maxLen; i++) {
    const a = p1[i] || 0;
    const b = p2[i] || 0;
    if (a !== b) return b - a;
  }
  return 0;
}

/**
 * 检查指定版本是否已在本地安装列表中
 */
export function isVersionLocallyInstalled(installedVersions: string[] | undefined | null, version: string): boolean {
  if (!installedVersions || !version) return false;
  const targetClean = cleanVer(version);
  if (!targetClean) return false;

  return installedVersions.some((iv) => cleanVer(iv) === targetClean);
}

/**
 * 计算工具在指定作用域下的核心显示状态 (Ready / Upgradable / Missing / Unconfigured)
 */
export function computeToolStatus(
  tool: ToolStatusItem,
  scope: "user" | "workspace",
  updates: ToolUpdateCandidate[] = []
): ToolStatusResult {
  const configured = scope === "user" ? tool.global_version || "" : tool.active_version || tool.global_version || "";

  if (!configured.trim()) {
    return {
      state: "unconfigured",
      configuredVersion: "",
      upgradableVersion: null,
      isInstalled: false,
    };
  }

  const isInstalled = isVersionLocallyInstalled(tool.installed_versions, configured);

  // 1. 如果工作区配置了该版本，但本地磁盘中根本没有该版本 -> Missing (待下载)
  if (!isInstalled) {
    return {
      state: "missing",
      configuredVersion: configured,
      upgradableVersion: null,
      isInstalled: false,
    };
  }

  // 2. 检查是否有更高可用升级版本（且未被锁定）
  if (!tool.pinned) {
    const normId = (id: string) => (id || "").replace(/^github:/, "").toLowerCase();
    const shortName = (id: string) => normId(id).split("/").pop() || "";
    const toolNorm = normId(tool.id);
    const toolShort = shortName(tool.id);

    const matchUpdate = updates.find((u) => {
      if (u.pinned) return false;
      if (u.id === tool.id || (u.name && tool.name && u.name === tool.name)) return true;
      const uNorm = normId(u.id);
      const uShort = shortName(u.id);
      return (
        (toolNorm && uNorm === toolNorm) ||
        (toolShort && (uShort === toolShort || (u.name || "").toLowerCase() === toolShort))
      );
    });

    if (matchUpdate && matchUpdate.has_update && matchUpdate.latest_version) {
      const cleanLatest = cleanVer(matchUpdate.latest_version);
      const cleanCur = cleanVer(configured);
      // cleanLatest 严格大于 cleanCur 时成立
      if (cleanLatest && cleanLatest !== cleanCur && compareSemverDesc(cleanLatest, cleanCur) < 0) {
        return {
          state: "upgradable",
          configuredVersion: configured,
          upgradableVersion: matchUpdate.latest_version,
          isInstalled: true,
        };
      }
    }
  }

  // 3. 正常已配置且已安装在本地 -> Ready
  return {
    state: "ready",
    configuredVersion: configured,
    upgradableVersion: null,
    isInstalled: true,
  };
}

export interface SidebarToolItem {
  id: string;
  name: string;
  icon: string;
  version: string;
  state: ToolDisplayState;
  upgradableVersion: string | null;
  raw: any;
}

export interface CategorizedTools {
  missing: SidebarToolItem[];
  upgradable: SidebarToolItem[];
  ready: SidebarToolItem[];
  allFiltered: SidebarToolItem[];
}

/**
 * 将工具列表根据计算出的状态（missing / upgradable / ready）分类归入不同面板分组
 * 并同时支持 searchQuery 过滤
 */
export function categorizeTools(
  tools: any[],
  scope: "user" | "workspace",
  updates: ToolUpdateCandidate[] = [],
  searchQuery: string = ""
): CategorizedTools {
  const q = searchQuery.toLowerCase().trim();
  const missing: SidebarToolItem[] = [];
  const upgradable: SidebarToolItem[] = [];
  const ready: SidebarToolItem[] = [];
  const allFiltered: SidebarToolItem[] = [];

  for (const tool of tools) {
    if (!tool) continue;
    const hasConfig =
      scope === "user"
        ? Boolean(tool.global_version && String(tool.global_version).trim())
        : Boolean(
            (tool.active_version && String(tool.active_version).trim()) ||
            tool.is_legacy ||
            (tool.global_version && String(tool.global_version).trim())
          );
    if (!hasConfig) continue;

    const status = computeToolStatus(tool, scope, updates);
    if (status.state === "unconfigured") continue;

    const item: SidebarToolItem = {
      id: tool.id || tool.name,
      name: tool.name || tool.id,
      icon: tool.id || tool.name,
      version: status.configuredVersion,
      state: status.state,
      upgradableVersion: status.upgradableVersion,
      raw: tool,
    };

    if (q) {
      const matchName = item.name.toLowerCase().includes(q);
      const matchId = item.id.toLowerCase().includes(q);
      const matchVer = item.version.toLowerCase().includes(q);
      if (!matchName && !matchId && !matchVer) {
        continue;
      }
    }

    allFiltered.push(item);
    if (item.state === "missing") {
      missing.push(item);
    } else if (item.state === "upgradable") {
      upgradable.push(item);
    } else if (item.state === "ready") {
      ready.push(item);
    }
  }

  return { missing, upgradable, ready, allFiltered };
}
