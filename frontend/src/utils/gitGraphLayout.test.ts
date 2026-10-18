// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { layoutGitGraph, getLaneColor, GRAPH_COLORS } from "./gitGraphLayout";
import type { GitGraphCommit } from "../composables/useGit";

describe("gitGraphLayout", () => {
  it("returns default empty result for empty commits list", () => {
    const res = layoutGitGraph([]);
    expect(res.nodes).toEqual([]);
    expect(res.paths).toEqual([]);
    expect(res.maxLane).toBe(0);
  });

  it("handles a single linear commit", () => {
    const commits: GitGraphCommit[] = [
      {
        hash: "c1",
        shortHash: "c1",
        parents: [],
        author: "Alice",
        email: "alice@example.com",
        timestamp: 1000,
        message: "Initial commit",
        refs: [{ name: "main", type: "branch", current: true }],
      },
    ];

    const res = layoutGitGraph(commits, { rowHeight: 30, laneWidth: 20, startX: 10 });
    expect(res.nodes.length).toBe(1);
    expect(res.nodes[0].hash).toBe("c1");
    expect(res.nodes[0].lane).toBe(0);
    expect(res.nodes[0].x).toBe(10);
    expect(res.nodes[0].y).toBe(15);
    expect(res.nodes[0].isHead).toBe(true);
    expect(res.paths.length).toBe(0);
  });

  it("draws vertical straight path for linear sequence", () => {
    const commits: GitGraphCommit[] = [
      {
        hash: "c2",
        shortHash: "c2",
        parents: ["c1"],
        author: "Alice",
        email: "alice@example.com",
        timestamp: 2000,
        message: "Second commit",
        refs: [{ name: "HEAD", type: "head", current: true }],
      },
      {
        hash: "c1",
        shortHash: "c1",
        parents: [],
        author: "Alice",
        email: "alice@example.com",
        timestamp: 1000,
        message: "First commit",
        refs: [],
      },
    ];

    const res = layoutGitGraph(commits, { rowHeight: 30, laneWidth: 20, startX: 10 });
    expect(res.nodes.length).toBe(2);
    expect(res.paths.length).toBe(1);
    expect(res.paths[0].id).toBe("c2-c1");
    expect(res.paths[0].d).toContain("L");
    expect(res.paths[0].fromLane).toBe(0);
    expect(res.paths[0].toLane).toBe(0);
  });

  it("draws bezier curve for branch merge", () => {
    const commits: GitGraphCommit[] = [
      {
        hash: "c3",
        shortHash: "c3",
        parents: ["c2", "b1"], // Merge commit
        author: "Alice",
        email: "alice@example.com",
        timestamp: 3000,
        message: "Merge branch feature",
        refs: [],
      },
      {
        hash: "b1",
        shortHash: "b1",
        parents: ["c1"],
        author: "Bob",
        email: "bob@example.com",
        timestamp: 2500,
        message: "Feature commit",
        refs: [],
      },
      {
        hash: "c2",
        shortHash: "c2",
        parents: ["c1"],
        author: "Alice",
        email: "alice@example.com",
        timestamp: 2000,
        message: "Main commit",
        refs: [],
      },
      {
        hash: "c1",
        shortHash: "c1",
        parents: [],
        author: "Alice",
        email: "alice@example.com",
        timestamp: 1000,
        message: "Root commit",
        refs: [],
      },
    ];

    const res = layoutGitGraph(commits);
    expect(res.nodes.length).toBe(4);
    expect(res.paths.length).toBeGreaterThan(1);
    // Merge curve should contain 'C' (cubic bezier)
    const curvePath = res.paths.find((p) => p.d.includes("C"));
    expect(curvePath).toBeDefined();
  });

  it("getLaneColor wraps around colors properly", () => {
    expect(getLaneColor(0)).toBe(GRAPH_COLORS[0]);
    expect(getLaneColor(GRAPH_COLORS.length)).toBe(GRAPH_COLORS[0]);
  });
});
