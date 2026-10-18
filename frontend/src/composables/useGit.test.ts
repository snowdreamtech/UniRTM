// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, vi, beforeEach } from "vitest";
import { useGit } from "./useGit";

describe("useGit composable", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("fetchStatus parses response properly and updates reactive status", async () => {
    const mockStatusData = {
      branch: "dev",
      upstream: "origin/dev",
      ahead: 1,
      behind: 0,
      staged: [
        { path: "src/main.ts", status: "M", label: "modified" },
      ],
      unstaged: [
        { path: "src/test.ts", status: "A", label: "added" },
      ],
      totalChanges: 2,
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockStatusData,
    });

    const { status, currentBranch, totalChanges, hasStagedChanges, fetchStatus } = useGit();
    await fetchStatus();

    expect(status.value.branch).toBe("dev");
    expect(currentBranch.value).toBe("dev");
    expect(status.value.staged).toHaveLength(1);
    expect(status.value.unstaged).toHaveLength(1);
    expect(totalChanges.value).toBe(2);
    expect(hasStagedChanges.value).toBe(true);
  });

  it("stage sends POST request to /api/git/stage", async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === "/api/git/stage") {
        return Promise.resolve({
          ok: true,
          json: async () => ({ success: true }),
        });
      }
      return Promise.resolve({
        ok: true,
        json: async () => ({ branch: "dev", staged: [], unstaged: [] }),
      });
    });

    const { stage } = useGit();
    const result = await stage(["src/foo.ts"]);
    expect(result).toBe(true);
    expect(global.fetch).toHaveBeenCalledWith("/api/git/stage", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ paths: ["src/foo.ts"] }),
    }));
  });

  it("commit sends commit message", async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === "/api/git/commit") {
        return Promise.resolve({
          ok: true,
          json: async () => ({ success: true, commit: "feat: add scm" }),
        });
      }
      return Promise.resolve({
        ok: true,
        json: async () => ({ branch: "dev", staged: [], unstaged: [] }),
      });
    });

    const { commit } = useGit();
    const result = await commit("feat: add scm");
    expect(result).toBe(true);
    expect(global.fetch).toHaveBeenCalledWith("/api/git/commit", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ message: "feat: add scm", amend: false }),
    }));
  });

  it("fetchDiff returns diff payload", async () => {
    const mockDiff = {
      path: "test.txt",
      staged: false,
      original: "line 1\n",
      modified: "line 1\nline 2\n",
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockDiff,
    });

    const { fetchDiff } = useGit();
    const diff = await fetchDiff("test.txt", false);
    expect(diff).toEqual(mockDiff);
    expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining("/api/git/diff?path=test.txt&staged=false"));
  });

  it("fetchLog fetches commit history and updates commits ref", async () => {
    const mockCommits = [
      { hash: "abc12345", shortHash: "abc1234", author: "dev", email: "dev@test.com", timestamp: 123456789, message: "init" },
    ];

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, commits: mockCommits }),
    });

    const { commits, fetchLog } = useGit();
    const result = await fetchLog();
    expect(result).toHaveLength(1);
    expect(commits.value[0].message).toBe("init");
  });

  it("sync sends POST to /api/git/sync and refreshes status", async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === "/api/git/sync") {
        return Promise.resolve({
          ok: true,
          json: async () => ({ success: true }),
        });
      }
      return Promise.resolve({
        ok: true,
        json: async () => ({ branch: "dev", staged: [], unstaged: [] }),
      });
    });

    const { sync } = useGit();
    const res = await sync();
    expect(res).toBe(true);
    expect(global.fetch).toHaveBeenCalledWith("/api/git/sync", expect.objectContaining({ method: "POST" }));
  });

  it("getGitStatus and getDirGitCount return status decorations properly", async () => {
    const mockStatusData = {
      branch: "dev",
      staged: [{ path: "src/file1.ts", status: "M" }],
      unstaged: [{ path: "src/sub/file2.ts", status: "U" }],
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockStatusData,
    });

    const { fetchStatus, getGitStatus, getDirGitCount } = useGit();
    await fetchStatus();

    expect(getGitStatus("src/file1.ts")).toBe("M");
    expect(getGitStatus("src/sub/file2.ts")).toBe("U");
    expect(getGitStatus("src/unknown.ts")).toBeUndefined();
    expect(getDirGitCount("src")).toBe(2);
    expect(getDirGitCount("src/sub")).toBe(1);
    expect(getDirGitCount("other")).toBe(0);
  });

  it("fetchGraph fetches topological DAG commits properly", async () => {
    const mockGraphCommits = [
      {
        hash: "c2",
        shortHash: "c2",
        parents: ["c1"],
        author: "snowdream",
        email: "test@example.com",
        timestamp: 1792334000,
        message: "feat: add git graph",
        refs: [{ name: "dev", type: "branch", current: true }],
      },
    ];

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, commits: mockGraphCommits }),
    });

    const { graphCommits, fetchGraph } = useGit();
    const res = await fetchGraph({ branch: "dev" });
    expect(res).toHaveLength(1);
    expect(graphCommits.value[0].shortHash).toBe("c2");
    expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining("/api/git/graph?branch=dev&limit=200"));
  });

  it("fetchCommitDetail fetches detailed commit and changed files", async () => {
    const mockDetail = {
      hash: "c2",
      shortHash: "c2",
      parents: ["c1"],
      author: "snowdream",
      email: "test@example.com",
      timestamp: 1792334000,
      committer: "snowdream",
      committerEmail: "test@example.com",
      committerTimestamp: 1792334000,
      message: "feat: add git graph",
      files: [{ path: "src/main.ts", status: "M", additions: 10, deletions: 2 }],
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, commit: mockDetail }),
    });

    const { fetchCommitDetail, selectedCommitDetail } = useGit();
    const res = await fetchCommitDetail("c2");
    expect(res).toBeDefined();
    expect(selectedCommitDetail.value?.files).toHaveLength(1);
    expect(global.fetch).toHaveBeenCalledWith("/api/git/commit-detail?hash=c2");
  });

  it("fetchCommitDiff fetches commit diff with parent", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        success: true,
        path: "src/main.ts",
        hash: "c2",
        parent: "c1",
        original: "const a = 1;",
        modified: "const a = 2;",
      }),
    });

    const { fetchCommitDiff, activeCommitDiff } = useGit();
    const res = await fetchCommitDiff("c2", "src/main.ts", "c1");
    expect(res?.original).toBe("const a = 1;");
    expect(res?.modified).toBe("const a = 2;");
    expect(activeCommitDiff.value?.hash).toBe("c2");
  });

  it("createTag, createBranchFromCommit, cherryPick, revertCommit, resetToCommit call endpoints", async () => {
    global.fetch = vi.fn().mockImplementation(() => {
      return Promise.resolve({
        ok: true,
        json: async () => ({ success: true, commits: [] }),
      });
    });

    const {
      createTag,
      createBranchFromCommit,
      cherryPick,
      revertCommit,
      resetToCommit,
    } = useGit();

    await createTag("v1.0.0", "c2", "first release");
    expect(global.fetch).toHaveBeenCalledWith("/api/git/tag", expect.objectContaining({ method: "POST" }));

    await createBranchFromCommit("feature/test", "c2", true);
    expect(global.fetch).toHaveBeenCalledWith("/api/git/branch/create", expect.objectContaining({ method: "POST" }));

    await cherryPick("c2");
    expect(global.fetch).toHaveBeenCalledWith("/api/git/cherry-pick", expect.objectContaining({ method: "POST" }));

    await revertCommit("c2");
    expect(global.fetch).toHaveBeenCalledWith("/api/git/revert", expect.objectContaining({ method: "POST" }));

    await resetToCommit("c2", "hard");
    expect(global.fetch).toHaveBeenCalledWith("/api/git/reset", expect.objectContaining({ method: "POST" }));
  });
});

