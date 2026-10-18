// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import { useToast } from "./useToast";

export interface GitChangeItem {
  path: string;
  origPath?: string;
  status: "M" | "A" | "D" | "R" | "C" | "U" | "?" | string;
  label: string;
}

export interface GitStatusData {
  branch: string;
  upstream?: string;
  ahead: number;
  behind: number;
  staged: GitChangeItem[];
  unstaged: GitChangeItem[];
  totalChanges: number;
}

export interface GitBranchItem {
  name: string;
  current: boolean;
  remote: boolean;
}

export interface GitDiffResult {
  path: string;
  staged: boolean;
  original: string;
  modified: string;
}

export interface GitCommitItem {
  hash: string;
  shortHash: string;
  author: string;
  email: string;
  timestamp: number;
  message: string;
}

const status = ref<GitStatusData>({
  branch: "main",
  ahead: 0,
  behind: 0,
  staged: [],
  unstaged: [],
  totalChanges: 0,
});

const branches = ref<GitBranchItem[]>([]);
const commits = ref<GitCommitItem[]>([]);
const commitMessage = ref<string>("");
const isLoading = ref<boolean>(false);
const isCommitting = ref<boolean>(false);
const isSyncing = ref<boolean>(false);
const lastError = ref<string | null>(null);

export function useGit() {
  const { showToast } = useToast();

  const currentBranch = computed(() => status.value.branch || "HEAD");
  const totalChanges = computed(() => (status.value.staged.length + status.value.unstaged.length));
  const hasStagedChanges = computed(() => status.value.staged.length > 0);
  const hasUnstagedChanges = computed(() => status.value.unstaged.length > 0);

  const gitChangeMap = computed<Map<string, string>>(() => {
    const map = new Map<string, string>();
    for (const item of status.value.unstaged) {
      map.set(item.path, item.status);
    }
    for (const item of status.value.staged) {
      map.set(item.path, item.status);
    }
    return map;
  });

  function getGitStatus(path: string): string | undefined {
    return gitChangeMap.value.get(path);
  }

  function getDirGitCount(dirPath: string): number {
    let count = 0;
    const prefix = dirPath.endsWith("/") ? dirPath : dirPath + "/";
    for (const [p] of gitChangeMap.value) {
      if (p.startsWith(prefix)) count++;
    }
    return count;
  }

  /**
   * Fetches latest git status from server.
   */
  async function fetchStatus(): Promise<GitStatusData> {
    isLoading.value = true;
    lastError.value = null;
    try {
      const res = await fetch("/api/git/status");
      if (!res.ok) {
        throw new Error(`Git status failed: HTTP ${res.status}`);
      }
      const data = await res.json();
      const stagedList = Array.isArray(data.staged) ? data.staged : [];
      const unstagedList = Array.isArray(data.unstaged) ? [...data.unstaged] : [];
      if (Array.isArray(data.untracked)) {
        for (const u of data.untracked) {
          if (!unstagedList.some((i) => i.path === u.path)) {
            unstagedList.push(u);
          }
        }
      }
      status.value = {
        branch: data.branch || "HEAD",
        upstream: data.upstream || "",
        ahead: data.ahead || 0,
        behind: data.behind || 0,
        staged: stagedList,
        unstaged: unstagedList,
        totalChanges: data.totalChanges ?? (stagedList.length + unstagedList.length),
      };
      return status.value;
    } catch (err: any) {
      lastError.value = err.message || String(err);
      return status.value;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Fetches list of local and remote branches.
   */
  async function fetchBranches(): Promise<GitBranchItem[]> {
    try {
      const res = await fetch("/api/git/branches");
      if (!res.ok) {
        throw new Error(`Git branches failed: HTTP ${res.status}`);
      }
      const data = await res.json();
      branches.value = Array.isArray(data.branches) ? data.branches : [];
      return branches.value;
    } catch (err: any) {
      console.error("fetchBranches error:", err);
      return [];
    }
  }

  /**
   * Stage files (git add). If paths is empty or omitted, stages all changes.
   */
  async function stage(paths?: string[]): Promise<boolean> {
    isLoading.value = true;
    try {
      const res = await fetch("/api/git/stage", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ paths: paths || [] }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || "暂存操作失败");
      }
      await fetchStatus();
      return true;
    } catch (err: any) {
      showToast(err.message || "暂存失败", "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Unstage files (git reset / restore --staged). If paths is empty, unstages all.
   */
  async function unstage(paths?: string[]): Promise<boolean> {
    isLoading.value = true;
    try {
      const res = await fetch("/api/git/unstage", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ paths: paths || [] }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || "取消暂存失败");
      }
      await fetchStatus();
      return true;
    } catch (err: any) {
      showToast(err.message || "取消暂存失败", "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Discard unstaged changes for given files (git restore / checkout / clean).
   */
  async function discard(paths: string[]): Promise<boolean> {
    if (!paths || paths.length === 0) return false;
    isLoading.value = true;
    try {
      const res = await fetch("/api/git/discard", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ paths }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || "放弃更改失败");
      }
      showToast(`已放弃 ${paths.length} 个文件的更改`, "info");
      await fetchStatus();
      return true;
    } catch (err: any) {
      showToast(err.message || "放弃更改失败", "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Commit changes.
   */
  async function commit(msg?: string, amend: boolean = false): Promise<boolean> {
    const message = (msg !== undefined ? msg : commitMessage.value).trim();
    if (!message && !amend) {
      showToast("请输入提交信息 (Commit message)", "warning");
      return false;
    }
    isCommitting.value = true;
    try {
      const res = await fetch("/api/git/commit", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message, amend }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || "提交失败");
      }
      commitMessage.value = "";
      showToast(`提交成功: ${message.slice(0, 30)}`, "success");
      await fetchStatus();
      return true;
    } catch (err: any) {
      showToast(err.message || "提交失败", "error");
      return false;
    } finally {
      isCommitting.value = false;
    }
  }

  /**
   * Checkout branch or switch to a newly created branch.
   */
  async function checkout(branch: string, create: boolean = false): Promise<boolean> {
    if (!branch) return false;
    isLoading.value = true;
    try {
      const res = await fetch("/api/git/checkout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ branch, create }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || "切换分支失败");
      }
      showToast(`已切换至分支: ${branch}`, "success");
      await Promise.all([fetchStatus(), fetchBranches()]);
      return true;
    } catch (err: any) {
      showToast(err.message || "切换分支失败", "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Fetch diff for a file (staged or unstaged) returning original and modified content.
   */
  async function fetchDiff(path: string, staged: boolean = false): Promise<GitDiffResult | null> {
    try {
      const q = new URLSearchParams({ path, staged: String(staged) });
      const res = await fetch(`/api/git/diff?${q.toString()}`);
      if (!res.ok) {
        throw new Error(`Git diff failed: HTTP ${res.status}`);
      }
      return await res.json();
    } catch (err: any) {
      console.error("fetchDiff error:", err);
      showToast(`获取文件差异失败: ${err.message}`, "error");
      return null;
    }
  }

  /**
   * Fetch commit log history.
   */
  async function fetchLog(): Promise<GitCommitItem[]> {
    try {
      const res = await fetch("/api/git/log");
      if (!res.ok) throw new Error(`Git log failed: HTTP ${res.status}`);
      const data = await res.json();
      commits.value = Array.isArray(data.commits) ? data.commits : [];
      return commits.value;
    } catch (err: any) {
      console.error("fetchLog error:", err);
      return [];
    }
  }

  /**
   * Sync with remote (pull then push).
   */
  async function sync(): Promise<boolean> {
    isSyncing.value = true;
    try {
      const res = await fetch("/api/git/sync", { method: "POST" });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || "同步失败");
      }
      showToast("Git 仓库已与远程同步完成", "success");
      await fetchStatus();
      return true;
    } catch (err: any) {
      showToast(err.message || "同步失败", "error");
      return false;
    } finally {
      isSyncing.value = false;
    }
  }

  return {
    status,
    branches,
    commits,
    commitMessage,
    isLoading,
    isCommitting,
    isSyncing,
    lastError,
    currentBranch,
    totalChanges,
    hasStagedChanges,
    hasUnstagedChanges,
    gitChangeMap,
    getGitStatus,
    getDirGitCount,
    fetchStatus,
    fetchBranches,
    fetchLog,
    sync,
    stage,
    unstage,
    discard,
    commit,
    checkout,
    fetchDiff,
  };
}
