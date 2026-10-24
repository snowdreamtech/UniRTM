// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import { useToast } from "./useToast";
import { t } from "../i18n";
import { useWorkspaceBus } from "./useWorkspaceBus";
import { useEditorTabs } from "./useEditorTabs";

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

export interface GitGpgInfo {
  status: "verified" | "unverified" | "bad" | "expired" | "revoked" | "none";
  keyId?: string;
  signer?: string;
  fingerprint?: string;
}

export interface GitCommitItem {
  hash: string;
  shortHash: string;
  author: string;
  email: string;
  timestamp: number;
  message: string;
  gpg?: GitGpgInfo;
}

export interface GitGraphRef {
  name: string;
  type: "head" | "branch" | "remote" | "tag" | "stash";
  current: boolean;
}

export interface GitGraphCommit {
  hash: string;
  shortHash: string;
  parents: string[];
  author: string;
  email: string;
  timestamp: number;
  message: string;
  refs: GitGraphRef[];
  gpg?: GitGpgInfo;
  isStash?: boolean;
}

export interface GitCommitFileItem {
  path: string;
  oldPath?: string;
  status: string;
  additions: number;
  deletions: number;
}

export interface GitCommitDetail {
  hash: string;
  shortHash: string;
  parents: string[];
  author: string;
  email: string;
  timestamp: number;
  committer: string;
  committerEmail: string;
  committerTimestamp: number;
  message: string;
  gpg?: GitGpgInfo;
  files: GitCommitFileItem[];
}

export interface GitCommitDiffResult {
  path: string;
  hash: string;
  parent?: string;
  original: string;
  modified: string;
}

export interface GitCompareCommitInfo {
  hash: string;
  shortHash: string;
  author: string;
  timestamp: number;
  message: string;
}

export interface GitRepositoryItem {
  id: string;
  name: string;
  rootPath: string;
  branch: string;
  upstream?: string;
  ahead: number;
  behind: number;
  isDirty: boolean;
  totalChanges: number;
}

export interface GitCompareResult {
  base: GitCompareCommitInfo;
  target: GitCompareCommitInfo;
  files: GitCommitFileItem[];
  totalAdditions: number;
  totalDeletions: number;
}

export const GIT_ACTIVE_REPO_STORAGE_KEY = "unirtm_active_git_repo";

export function getPersistedActiveRepoId(): string | null {
  if (typeof localStorage === "undefined") return null;
  try {
    return localStorage.getItem(GIT_ACTIVE_REPO_STORAGE_KEY);
  } catch {
    return null;
  }
}

export function persistActiveRepo(repo: GitRepositoryItem | null): void {
  if (typeof localStorage === "undefined") return;
  try {
    if (!repo) {
      localStorage.removeItem(GIT_ACTIVE_REPO_STORAGE_KEY);
    } else {
      localStorage.setItem(GIT_ACTIVE_REPO_STORAGE_KEY, repo.id || repo.rootPath || repo.name);
    }
  } catch {
    // ignore
  }
}

const repositories = ref<GitRepositoryItem[]>([]);
const activeRepo = ref<GitRepositoryItem | null>(null);
const isLoadingRepositories = ref<boolean>(false);

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
const graphCommits = ref<GitGraphCommit[]>([]);
const selectedCommit = ref<GitGraphCommit | null>(null);
const selectedCommitDetail = ref<GitCommitDetail | null>(null);
const activeCommitDiff = ref<GitCommitDiffResult | null>(null);
const activeCompare = ref<GitCompareResult | null>(null);
const graphFilterBranches = ref<string[]>(["all"]);
const graphFilterBranch = computed<string>({
  get: () => graphFilterBranches.value[0] || "all",
  set: (val: string) => {
    graphFilterBranches.value = val ? [val] : ["all"];
  },
});
const graphFilterPath = ref<string | null>(null);
const graphSearchQuery = ref<string>("");
const graphLimit = ref<number>(200);
const hasMoreCommits = ref<boolean>(true);
const graphShowStashes = ref<boolean>(true);
const isLoadingGraph = ref<boolean>(false);
const isLoadingMore = ref<boolean>(false);
const commitMessage = ref<string>("");
const isLoading = ref<boolean>(false);
const isCommitting = ref<boolean>(false);
const isSyncing = ref<boolean>(false);
const lastError = ref<string | null>(null);

function gitFetch(url: string, init?: RequestInit): Promise<Response> {
  if (!activeRepo.value) {
    return init !== undefined ? fetch(url, init) : fetch(url);
  }
  const separator = url.includes("?") ? "&" : "?";
  const finalUrl = `${url}${separator}repo=${encodeURIComponent(activeRepo.value.id || activeRepo.value.rootPath)}`;
  const headers = new Headers(init?.headers || {});
  headers.set("X-Git-Repo", activeRepo.value.id || activeRepo.value.rootPath);
  return fetch(finalUrl, {
    ...init,
    headers,
  });
}

export function useGit() {
  const { showToast } = useToast();

  const currentBranch = computed(() => status.value.branch || "HEAD");
  const totalChanges = computed(() => status.value.staged.length + status.value.unstaged.length);
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
      const res = await gitFetch("/api/git/status");
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
        totalChanges: data.totalChanges ?? stagedList.length + unstagedList.length,
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
      const res = await gitFetch("/api/git/branches");
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
      // Prevent conflicts: flush unsaved (dirty) editor files safely before staging
      try {
        const { saveDirtyTabsForPaths } = useEditorTabs();
        await saveDirtyTabsForPaths(paths);
      } catch (saveErr) {
        console.warn("[useGit] Failed to pre-save dirty tabs before stage:", saveErr);
      }

      const res = await gitFetch("/api/git/stage", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ paths: paths || [] }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || t("scm.toast.stage_failed") || "Staging operation failed");
      }
      await fetchStatus();
      useWorkspaceBus().notifyGitStatusChanged();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.stage_failed") || "Staging operation failed", "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Stage a specific chunk/diff range for a file.
   * Ensures dirty tabs are safely saved first, then executes staging and notifies workspace bus.
   */
  async function stageChunk(filePath: string, _chunk?: any): Promise<boolean> {
    if (!filePath) return false;
    try {
      const { saveDirtyTabsForPaths } = useEditorTabs();
      await saveDirtyTabsForPaths([filePath]);
    } catch (saveErr) {
      console.warn("[useGit] Failed to pre-save before stageChunk:", saveErr);
    }

    const success = await stage([filePath]);
    if (success) {
      useWorkspaceBus().notifyGitStatusChanged();
    }
    return success;
  }

  /**
   * Unstage files (git reset / restore --staged). If paths is empty, unstages all.
   */
  async function unstage(paths?: string[]): Promise<boolean> {
    isLoading.value = true;
    try {
      const res = await gitFetch("/api/git/unstage", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ paths: paths || [] }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || t("scm.toast.unstage_failed") || "Unstaging failed");
      }
      await fetchStatus();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.unstage_failed") || "Unstaging failed", "error");
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
      const res = await gitFetch("/api/git/discard", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ paths }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || t("scm.toast.discard_failed") || "Failed to discard changes");
      }
      showToast(
        t("scm.toast.discard_success", { count: paths.length }) || `Discarded changes in ${paths.length} files`,
        "info"
      );
      await fetchStatus();

      // Notify editor to reload freshly reverted disk contents, preventing stale dirty state
      const bus = useWorkspaceBus();
      bus.notifyGitStatusChanged();
      bus.notifyFileReloadNeeded(paths);

      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.discard_failed") || "Failed to discard changes", "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Commit changes.
   */
  async function commit(msg?: string, amend: boolean = false, sign?: boolean): Promise<boolean> {
    const message = (msg !== undefined ? msg : commitMessage.value).trim();
    if (!message && !amend) {
      showToast(t("scm.toast.enter_commit_msg") || "Please enter a commit message", "warning");
      return false;
    }
    isCommitting.value = true;
    try {
      const payload: { message: string; amend: boolean; sign?: boolean } = { message, amend };
      if (typeof sign === "boolean") {
        payload.sign = sign;
      }
      const res = await gitFetch("/api/git/commit", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || t("scm.toast.commit_failed") || "Commit failed");
      }
      commitMessage.value = "";
      showToast(
        t("scm.toast.commit_success", { msg: message.slice(0, 30) }) || `Commit successful: ${message.slice(0, 30)}`,
        "success"
      );
      await fetchStatus();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.commit_failed") || "Commit failed", "error");
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
      const res = await gitFetch("/api/git/checkout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ branch, create }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || t("scm.toast.switch_branch_failed") || "Failed to switch branch");
      }
      showToast(t("scm.toast.switch_branch_success", { branch }) || `Switched to branch: ${branch}`, "success");
      await Promise.all([fetchStatus(), fetchBranches()]);

      const bus = useWorkspaceBus();
      bus.notifyGitStatusChanged();
      // Notify all opened tabs to refresh content (files changed after switching branches)
      try {
        const { tabs } = useEditorTabs();
        const openPaths = tabs.value.filter((t) => !t.isDiff).map((t) => t.path);
        if (openPaths.length > 0) {
          bus.notifyFileReloadNeeded(openPaths);
        }
      } catch (tabsErr) {
        console.warn("[useGit] Failed to notify tabs on checkout:", tabsErr);
      }

      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.switch_branch_failed") || "Failed to switch branch", "error");
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
      const res = await gitFetch(`/api/git/diff?${q.toString()}`);
      if (!res.ok) {
        throw new Error(`Git diff failed: HTTP ${res.status}`);
      }
      return await res.json();
    } catch (err: any) {
      console.error("fetchDiff error:", err);
      showToast(
        t("scm.toast.get_diff_failed", { error: err.message }) || `Failed to get file diff: ${err.message}`,
        "error"
      );
      return null;
    }
  }

  /**
   * Fetch commit log history.
   */
  async function fetchLog(): Promise<GitCommitItem[]> {
    try {
      const res = await gitFetch("/api/git/log");
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
      const res = await gitFetch("/api/git/sync", { method: "POST" });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || t("scm.toast.sync_failed") || "Sync failed");
      }
      showToast(t("scm.toast.sync_success") || "Git repository synced with remote", "success");
      await fetchStatus();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.sync_failed") || "Sync failed", "error");
      return false;
    } finally {
      isSyncing.value = false;
    }
  }

  /**
   * Fetches the commit graph for Git Graph DAG view.
   */
  async function fetchGraph(options?: {
    branch?: string | string[];
    path?: string | null;
    limit?: number;
    showStashes?: boolean;
  }): Promise<GitGraphCommit[]> {
    isLoadingGraph.value = true;
    try {
      const b = options?.branch !== undefined ? options.branch : graphFilterBranches.value;
      const p = options?.path !== undefined ? options.path : graphFilterPath.value;
      const limit = options?.limit !== undefined ? options.limit : graphLimit.value;
      graphLimit.value = limit;

      const q = new URLSearchParams();
      if (Array.isArray(b)) {
        if (b.length > 0 && !b.includes("all")) {
          q.set("branch", b.join(","));
        }
      } else if (b && b !== "all") {
        q.set("branch", b);
      }
      if (p) q.set("path", p);
      if (options?.showStashes !== undefined) {
        graphShowStashes.value = options.showStashes;
        q.set("showStashes", String(options.showStashes));
      }
      q.set("limit", String(limit));

      const res = await gitFetch(`/api/git/graph?${q.toString()}`);
      if (!res.ok) throw new Error(`Git graph failed: HTTP ${res.status}`);
      const data = await res.json();
      const rawCommits = Array.isArray(data.commits) ? data.commits : [];
      graphCommits.value = rawCommits;
      const normalCount = rawCommits.filter((c: any) => !c.isStash && c.hash !== "*").length;
      hasMoreCommits.value = normalCount >= limit;
      return graphCommits.value;
    } catch (err: any) {
      console.error("fetchGraph error:", err);
      showToast(err.message || t("scm.toast.get_graph_failed") || "Failed to get Git commit graph", "error");
      return [];
    } finally {
      isLoadingGraph.value = false;
    }
  }

  /**
   * Loads next page of commits in graph DAG view.
   */
  async function loadMoreCommits(step: number = 200): Promise<GitGraphCommit[]> {
    if (isLoadingGraph.value || isLoadingMore.value || !hasMoreCommits.value) return graphCommits.value;
    isLoadingMore.value = true;
    try {
      const nextLimit = graphLimit.value + step;
      return await fetchGraph({ limit: nextLimit, showStashes: graphShowStashes.value });
    } finally {
      isLoadingMore.value = false;
    }
  }

  /**
   * Resets graph limit pagination state.
   */
  function resetGraphLimit(initial: number = 200) {
    graphLimit.value = initial;
    hasMoreCommits.value = true;
  }

  /**
   * Fetches full commit details and changed files with diff stats.
   */
  async function fetchCommitDetail(hash: string): Promise<GitCommitDetail | null> {
    try {
      const res = await gitFetch(`/api/git/commit-detail?hash=${encodeURIComponent(hash)}`);
      if (!res.ok) throw new Error(`Fetch commit detail failed: HTTP ${res.status}`);
      const data = await res.json();
      if (data.success && data.commit) {
        selectedCommitDetail.value = data.commit;
        return data.commit;
      }
      return null;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.get_commit_details_failed") || "Failed to get commit details", "error");
      return null;
    }
  }

  /**
   * Fetches diff for a file in a specific historical commit vs its parent.
   */
  async function fetchCommitDiff(hash: string, path: string, parentHash?: string): Promise<GitCommitDiffResult | null> {
    try {
      const q = new URLSearchParams({ hash, path });
      if (parentHash) q.set("parent", parentHash);
      const res = await gitFetch(`/api/git/commit-diff?${q.toString()}`);
      if (!res.ok) throw new Error(`Fetch commit diff failed: HTTP ${res.status}`);
      const data = await res.json();
      if (data.success) {
        activeCommitDiff.value = {
          path: data.path,
          hash: data.hash,
          parent: data.parent,
          original: data.original || "",
          modified: data.modified || "",
        };
        return activeCommitDiff.value;
      }
      return null;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.get_history_diff_failed") || "Failed to get history diff", "error");
      return null;
    }
  }

  /**
   * Compares two commits and returns file changes and diff statistics.
   */
  async function fetchCompare(base: string, target: string): Promise<GitCompareResult | null> {
    try {
      const q = new URLSearchParams({ base, target });
      const res = await gitFetch(`/api/git/compare?${q.toString()}`);
      if (!res.ok) throw new Error(`Fetch compare failed: HTTP ${res.status}`);
      const data = await res.json();
      if (data.success && data.compare) {
        activeCompare.value = data.compare;
        return data.compare;
      }
      return null;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.compare_failed") || "Failed to compare commits", "error");
      return null;
    }
  }

  /**
   * Creates a Git tag on a specific commit.
   */
  async function createTag(name: string, hash: string, message?: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/tag", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, hash, message }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.create_tag_failed") || "Failed to create tag");
      showToast(t("scm.toast.create_tag_success", { name }) || `Tag created successfully: ${name}`, "success");
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.create_tag_failed") || "Failed to create tag", "error");
      return false;
    }
  }

  /**
   * Creates a new branch pointing to a specific commit.
   */
  async function createBranchFromCommit(name: string, hash: string, checkoutBranch: boolean = false): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/branch/create", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, hash, checkout: checkoutBranch }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.create_branch_failed") || "Failed to create branch");
      showToast(
        checkoutBranch
          ? t("scm.toast.create_branch_and_switched", { name }) || `Branch created and switched: ${name}`
          : t("scm.toast.create_branch_success", { name }) || `Branch created successfully: ${name}`,
        "success"
      );
      await fetchBranches();
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.create_branch_failed") || "Failed to create branch", "error");
      return false;
    }
  }

  /**
   * Cherry-picks a commit onto current HEAD.
   */
  async function cherryPick(hash: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/cherry-pick", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ hash }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.cherry_pick_failed") || "Cherry-pick failed");
      showToast(
        t("scm.toast.cherry_pick_success", { hash: hash.slice(0, 8) }) || `Cherry-picked commit: ${hash.slice(0, 8)}`,
        "success"
      );
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.cherry_pick_failed") || "Cherry-pick failed", "error");
      return false;
    }
  }

  /**
   * Reverts a commit.
   */
  async function revertCommit(hash: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/revert", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ hash }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.revert_failed") || "Revert commit failed");
      showToast(
        t("scm.toast.revert_success", { hash: hash.slice(0, 8) }) || `Reverted commit: ${hash.slice(0, 8)}`,
        "success"
      );
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.revert_failed") || "Revert commit failed", "error");
      return false;
    }
  }

  /**
   * Resets branch to specific commit.
   */
  async function resetToCommit(hash: string, mode: "soft" | "mixed" | "hard" = "mixed"): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/reset", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ hash, mode }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) throw new Error(data.error || t("scm.toast.reset_failed") || "Reset failed");
      showToast(
        t("scm.toast.reset_success", { hash: hash.slice(0, 8), mode }) ||
          `Reset to commit ${hash.slice(0, 8)} (${mode})`,
        "info"
      );
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.reset_failed") || "Reset failed", "error");
      return false;
    }
  }

  /**
   * Renames a branch.
   */
  async function renameBranch(oldName: string, newName: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/branch/rename", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ oldName, newName }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.rename_branch_failed") || "Failed to rename branch");
      await fetchBranches();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.rename_branch_failed") || "Failed to rename branch", "error");
      return false;
    }
  }

  /**
   * Deletes a branch.
   */
  async function deleteBranch(name: string, force: boolean = false): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/branch/delete", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, force }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.delete_branch_failed") || "Failed to delete branch");
      await fetchBranches();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.delete_branch_failed") || "Failed to delete branch", "error");
      return false;
    }
  }

  /**
   * Merges specified branch into current branch.
   */
  async function mergeBranch(branch: string, noFf: boolean = false, squash: boolean = false): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/branch/merge", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ branch, noFf, squash }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.merge_branch_failed") || "Failed to merge branch");
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.merge_branch_failed") || "Failed to merge branch", "error");
      return false;
    }
  }

  /**
   * Rebases current branch onto specified branch.
   */
  async function rebaseBranch(branch: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/branch/rebase", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ branch }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.rebase_branch_failed") || "Failed to rebase branch");
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.rebase_branch_failed") || "Failed to rebase branch", "error");
      return false;
    }
  }

  /**
   * Pushes branch to remote.
   */
  async function pushBranch(branch: string, setUpstream: boolean = false, force: boolean = false): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/branch/push", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ branch, setUpstream, force }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.push_branch_failed") || "Failed to push branch");
      await fetchBranches();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.push_branch_failed") || "Failed to push branch", "error");
      return false;
    }
  }

  /**
   * Deletes a tag.
   */
  async function deleteTag(name: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/tag/delete", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.delete_tag_failed") || "Failed to delete tag");
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.delete_tag_failed") || "Failed to delete tag", "error");
      return false;
    }
  }

  /**
   * Pushes a tag to remote.
   */
  async function pushTag(name: string, force: boolean = false): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/tag/push", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, force }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.push_tag_failed") || "Failed to push tag");
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.push_tag_failed") || "Failed to push tag", "error");
      return false;
    }
  }

  async function downloadArchive(ref: string): Promise<boolean> {
    const filename = `${ref.replace(/\//g, "_")}.zip`;
    showToast(t("scm.branch.archive_generating", { ref }), "info");
    try {
      const res = await gitFetch(`/api/git/archive?ref=${encodeURIComponent(ref)}`);
      if (!res.ok) {
        const data = await res.json().catch(() => ({ error: res.statusText }));
        throw new Error(data.error || t("scm.toast.download_archive_failed") || "Failed to download archive");
      }
      const blob = await res.blob();
      const objectUrl = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = objectUrl;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      setTimeout(() => URL.revokeObjectURL(objectUrl), 60000);
      showToast(t("scm.branch.archive_success", { name: filename }), "success");
      return true;
    } catch (err: any) {
      try {
        window.open(`/api/git/archive?ref=${encodeURIComponent(ref)}`, "_blank");
      } catch (_) {}
      showToast(err.message || t("scm.toast.download_archive_failed") || "Failed to download archive", "error");
      return false;
    }
  }

  function setGraphFilterPath(p: string | null) {
    graphFilterPath.value = p;
    fetchGraph({ path: p });
  }

  function clearGraphFilterPath() {
    graphFilterPath.value = null;
    fetchGraph({ path: null });
  }

  /**
   * Stashes uncommitted changes.
   */
  async function stashUncommitted(message?: string, includeUntracked: boolean = true): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/stash", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message, includeUntracked }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) throw new Error(data.error || t("scm.toast.stash_failed") || "Stash failed");
      showToast(t("scm.uncommitted.stash_success"), "success");
      await fetchStatus();
      await fetchBranches();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.stash_failed") || "Stash failed", "error");
      return false;
    }
  }

  /**
   * Applies a stash to working tree.
   */
  async function stashApply(stash?: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/stash/apply", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ stash }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.apply_stash_failed") || "Failed to apply stash");
      showToast(t("scm.toast.apply_stash_success") || "Stash applied successfully", "success");
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.apply_stash_failed") || "Failed to apply stash", "error");
      return false;
    }
  }

  /**
   * Pops a stash to working tree and removes it from stash list.
   */
  async function stashPop(stash?: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/stash/pop", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ stash }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.pop_stash_failed") || "Failed to pop stash");
      showToast(t("scm.toast.pop_stash_success") || "Stash popped successfully", "success");
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.pop_stash_failed") || "Failed to pop stash", "error");
      return false;
    }
  }

  /**
   * Drops a stash permanently.
   */
  async function stashDrop(stash?: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/stash/drop", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ stash }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.delete_stash_failed") || "Failed to delete stash");
      showToast(t("scm.toast.delete_stash_success") || "Stash deleted", "success");
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.delete_stash_failed") || "Failed to delete stash", "error");
      return false;
    }
  }

  /**
   * Creates a new branch from stash and applies it.
   */
  async function stashBranch(branch: string, stash?: string): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/stash/branch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ branch, stash }),
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.branch_from_stash_failed") || "Failed to create branch from stash");
      showToast(
        t("scm.toast.branch_from_stash_success", { branch }) || `Created branch from stash: ${branch}`,
        "success"
      );
      await fetchBranches();
      await fetchStatus();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(
        err.message || t("scm.toast.branch_from_stash_failed") || "Failed to create branch from stash",
        "error"
      );
      return false;
    }
  }

  /**
   * Resets all tracked uncommitted changes to HEAD.
   */
  async function resetUncommitted(): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/reset-uncommitted", {
        method: "POST",
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.reset_uncommitted_failed") || "Failed to reset uncommitted changes");
      showToast(t("scm.uncommitted.reset_success"), "success");
      await fetchStatus();
      await fetchBranches();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(
        err.message || t("scm.toast.reset_uncommitted_failed") || "Failed to reset uncommitted changes",
        "error"
      );
      return false;
    }
  }

  /**
   * Cleans all untracked files and directories.
   */
  async function cleanUntracked(): Promise<boolean> {
    try {
      const res = await gitFetch("/api/git/clean", {
        method: "POST",
      });
      const data = await res.json();
      if (!res.ok || !data.success)
        throw new Error(data.error || t("scm.toast.clean_untracked_failed") || "Failed to clean untracked files");
      showToast(t("scm.uncommitted.clean_success"), "success");
      await fetchStatus();
      await fetchBranches();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.clean_untracked_failed") || "Failed to clean untracked files", "error");
      return false;
    }
  }

  /**
   * Pulls remote changes (git pull), optionally with rebase.
   */
  async function pull(rebase: boolean = false): Promise<boolean> {
    isLoading.value = true;
    try {
      const res = await gitFetch("/api/git/pull", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ rebase }),
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || t("scm.toast.pull_failed") || "Pull failed");
      }
      showToast(t("scm.actions.pull_success"), "success");
      await fetchStatus();
      await fetchBranches();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.pull_failed") || "Pull failed", "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Fetches remote branches and tags from remotes (git fetch --all --prune).
   */
  async function fetchRemote(): Promise<boolean> {
    isLoading.value = true;
    try {
      const res = await gitFetch("/api/git/fetch", {
        method: "POST",
      });
      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.error || t("scm.toast.fetch_failed") || "Fetch failed");
      }
      showToast(t("scm.actions.fetch_success"), "success");
      await fetchStatus();
      await fetchBranches();
      await fetchGraph();
      return true;
    } catch (err: any) {
      showToast(err.message || t("scm.toast.fetch_failed") || "Fetch failed", "error");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Fetches the list of git repositories in the current multi-project workspace.
   */
  async function fetchRepositories(): Promise<GitRepositoryItem[]> {
    isLoadingRepositories.value = true;
    try {
      const res = await fetch("/api/git/repositories");
      if (!res.ok) {
        throw new Error(`Git repositories failed: HTTP ${res.status}`);
      }
      const data = await res.json();
      repositories.value = Array.isArray(data.repositories) ? data.repositories : [];
      if (repositories.value.length > 0) {
        const savedId = getPersistedActiveRepoId();
        if (savedId) {
          const matched = repositories.value.find(
            (r) => r.id === savedId || r.rootPath === savedId || r.name === savedId
          );
          if (matched) {
            activeRepo.value = matched;
          }
        }
        if (
          !activeRepo.value ||
          !repositories.value.some((r) => r.id === activeRepo.value?.id || r.rootPath === activeRepo.value?.rootPath)
        ) {
          activeRepo.value = repositories.value[0];
          persistActiveRepo(activeRepo.value);
        }
      }
      return repositories.value;
    } catch (err: any) {
      console.warn("[useGit] fetchRepositories error:", err);
      return [];
    } finally {
      isLoadingRepositories.value = false;
    }
  }

  /**
   * Sets the active git repository for source control operations.
   */
  async function setActiveRepo(repoOrId: GitRepositoryItem | string): Promise<void> {
    if (typeof repoOrId === "string") {
      const target = repositories.value.find(
        (r) => r.id === repoOrId || r.name === repoOrId || r.rootPath === repoOrId
      );
      if (target) {
        activeRepo.value = target;
      } else {
        activeRepo.value = {
          id: repoOrId,
          name: repoOrId,
          rootPath: repoOrId,
          branch: "HEAD",
          ahead: 0,
          behind: 0,
          isDirty: false,
          totalChanges: 0,
        };
      }
    } else {
      activeRepo.value = repoOrId;
    }
    persistActiveRepo(activeRepo.value);
    await Promise.all([fetchStatus(), fetchBranches()]);
  }

  return {
    repositories,
    activeRepo,
    isLoadingRepositories,
    fetchRepositories,
    setActiveRepo,
    status,
    branches,
    commits,
    graphCommits,
    selectedCommit,
    selectedCommitDetail,
    activeCommitDiff,
    graphFilterBranch,
    graphFilterBranches,
    graphFilterPath,
    graphSearchQuery,
    graphLimit,
    hasMoreCommits,
    isLoadingMore,
    loadMoreCommits,
    resetGraphLimit,
    isLoadingGraph,
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
    fetchGraph,
    fetchCommitDetail,
    fetchCommitDiff,
    activeCompare,
    fetchCompare,
    createTag,
    createBranchFromCommit,
    renameBranch,
    deleteBranch,
    mergeBranch,
    rebaseBranch,
    pushBranch,
    deleteTag,
    pushTag,
    downloadArchive,
    cherryPick,
    revertCommit,
    resetToCommit,
    stashUncommitted,
    stashApply,
    stashPop,
    stashDrop,
    stashBranch,
    resetUncommitted,
    cleanUntracked,
    pull,
    fetchRemote,
    setGraphFilterPath,
    clearGraphFilterPath,
    sync,
    stage,
    stageChunk,
    unstage,
    discard,
    commit,
    checkout,
    fetchDiff,
  };
}
