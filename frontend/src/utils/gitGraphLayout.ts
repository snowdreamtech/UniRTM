// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import type { GitGraphCommit } from "../composables/useGit";

export interface GraphNode {
  hash: string;
  x: number;
  y: number;
  lane: number;
  color: string;
  isHead: boolean;
}

export interface GraphPath {
  id: string;
  d: string;
  color: string;
  fromLane: number;
  toLane: number;
}

export interface GraphLayoutResult {
  nodes: GraphNode[];
  paths: GraphPath[];
  width: number;
  height: number;
  rowHeight: number;
  maxLane: number;
}

export const GRAPH_COLORS = [
  "#38bdf8", // Sky Blue
  "#4ade80", // Emerald Green
  "#fbbf24", // Amber Yellow
  "#c084fc", // Purple
  "#f472b6", // Rose Pink
  "#22d3ee", // Cyan
  "#fb923c", // Orange
  "#818cf8", // Indigo
];

export function getLaneColor(lane: number): string {
  return GRAPH_COLORS[lane % GRAPH_COLORS.length];
}

export interface GraphLayoutOptions {
  rowHeight?: number;
  laneWidth?: number;
  startX?: number;
  expandedIndex?: number;
  expandedHeight?: number;
  maxLanes?: number;
  maxWidth?: number;
}

/**
 * Computes topological lanes and SVG paths for Git Graph DAG commits.
 */
export function layoutGitGraph(commits: GitGraphCommit[], options: GraphLayoutOptions = {}): GraphLayoutResult {
  const rowHeight = options.rowHeight ?? 34;
  const laneWidth = options.laneWidth ?? 18;
  const startX = options.startX ?? 16;
  const maxLanes = options.maxLanes ?? 10;
  const maxWidth = options.maxWidth ?? 220;
  const expandedIndex = options.expandedIndex;
  const expandedHeight = expandedIndex !== undefined && expandedIndex >= 0 ? (options.expandedHeight ?? 280) : 0;

  function getRowTop(idx: number): number {
    if (expandedIndex === undefined || expandedIndex < 0 || idx <= expandedIndex) {
      return idx * rowHeight;
    }
    return idx * rowHeight + expandedHeight;
  }

  function getNodeY(idx: number): number {
    return getRowTop(idx) + rowHeight / 2;
  }

  if (!commits || commits.length === 0) {
    return {
      nodes: [],
      paths: [],
      width: 40,
      height: 40,
      rowHeight,
      maxLane: 0,
    };
  }

  const commitIndexMap = new Map<string, number>();
  commits.forEach((c, idx) => commitIndexMap.set(c.hash, idx));

  // Lanes: index -> commit hash that occupies or is expected on this lane
  const activeLanes: (string | null)[] = [];
  const commitLanes = new Map<string, number>();
  let maxLane = 0;

  for (let i = 0; i < commits.length; i++) {
    const commit = commits[i];
    let lane = activeLanes.indexOf(commit.hash);

    if (lane === -1) {
      // Find first empty lane or create new lane
      lane = activeLanes.indexOf(null);
      if (lane === -1) {
        lane = Math.min(maxLanes - 1, activeLanes.length);
        if (lane === activeLanes.length) {
          activeLanes.push(null);
        }
      }
    }

    lane = Math.min(maxLanes - 1, lane);
    commitLanes.set(commit.hash, lane);
    maxLane = Math.max(maxLane, lane);

    // Free current commit from lane
    activeLanes[lane] = null;

    // Allocate lanes only for parents that exist in the commit list (prevents lane leakage on filtered commits)
    const parents = (commit.parents || []).filter((p) => commitIndexMap.has(p));
    if (parents.length > 0) {
      // First parent inherits current lane if not already assigned
      const firstParent = parents[0];
      if (!activeLanes.includes(firstParent)) {
        activeLanes[lane] = firstParent;
      }

      // Other parents (merge sources)
      for (let pIdx = 1; pIdx < parents.length; pIdx++) {
        const parent = parents[pIdx];
        if (!activeLanes.includes(parent)) {
          let emptyLane = activeLanes.indexOf(null);
          if (emptyLane === -1) {
            emptyLane = Math.min(maxLanes - 1, activeLanes.length);
            if (emptyLane === activeLanes.length) {
              activeLanes.push(null);
            }
          }
          activeLanes[emptyLane] = parent;
          maxLane = Math.max(maxLane, emptyLane);
        }
      }
    }
  }

  // Generate SVG nodes and connecting paths
  const nodes: GraphNode[] = [];
  const paths: GraphPath[] = [];

  for (let i = 0; i < commits.length; i++) {
    const commit = commits[i];
    const rawLane = commitLanes.get(commit.hash) ?? 0;
    const lane = Math.min(maxLanes - 1, rawLane);
    const x = startX + lane * laneWidth;
    const y = getNodeY(i);
    const color = getLaneColor(lane);
    const isHead = commit.refs?.some((r) => r.type === "head" || r.current) ?? false;

    nodes.push({
      hash: commit.hash,
      x,
      y,
      lane,
      color,
      isHead,
    });

    const parents = commit.parents || [];
    for (let pIdx = 0; pIdx < parents.length; pIdx++) {
      const parentHash = parents[pIdx];
      const parentIdx = commitIndexMap.get(parentHash);

      if (parentIdx !== undefined) {
        const rawParentLane = commitLanes.get(parentHash) ?? lane;
        const parentLane = Math.min(maxLanes - 1, rawParentLane);
        const parentX = startX + parentLane * laneWidth;
        const parentY = getNodeY(parentIdx);
        const pathColor = pIdx === 0 ? color : getLaneColor(parentLane);

        if (lane === parentLane) {
          // Straight vertical connection
          paths.push({
            id: `${commit.hash}-${parentHash}`,
            d: `M ${x} ${y} L ${parentX} ${parentY}`,
            color: pathColor,
            fromLane: lane,
            toLane: parentLane,
          });
        } else {
          // Smooth cubic bezier curve for branching / merging
          const deltaY = parentY - y;
          const curveOffset = Math.min(rowHeight * 0.75, Math.abs(deltaY) * 0.4);

          paths.push({
            id: `${commit.hash}-${parentHash}`,
            d: `M ${x} ${y} C ${x} ${y + curveOffset}, ${parentX} ${parentY - curveOffset}, ${parentX} ${parentY}`,
            color: pathColor,
            fromLane: lane,
            toLane: parentLane,
          });
        }
      }
    }
  }

  const effectiveMaxLane = Math.min(maxLanes - 1, maxLane);
  const calculatedWidth = startX + (effectiveMaxLane + 1) * laneWidth + 12;
  const width = Math.min(maxWidth, Math.max(48, calculatedWidth));
  const height = Math.max(rowHeight, commits.length * rowHeight + expandedHeight);

  return {
    nodes,
    paths,
    width,
    height,
    rowHeight,
    maxLane: effectiveMaxLane,
  };
}
