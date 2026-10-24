// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export type DiffChangeType = "added" | "modified" | "deleted";

export interface DiffLineChange {
  type: DiffChangeType;
  startLineNumber: number;
  endLineNumber: number;
  originalText?: string;
}

/**
 * Compute line-level incremental changes between original and modified text (Myers/DP line diff)
 * Used to display subtle added (green) and modified (blue) indicators in editor gutter
 */
export function computeLineDiff(originalText: string, modifiedText: string): DiffLineChange[] {
  if (originalText === modifiedText) {
    return [];
  }

  const origLines = originalText.split("\n");
  const modLines = modifiedText.split("\n");

  const n = origLines.length;
  const m = modLines.length;

  if (n === 0 && m === 0) return [];

  // If original was empty, mark all as added
  if (originalText === "" && modifiedText !== "") {
    return [
      {
        type: "added",
        startLineNumber: 1,
        endLineNumber: Math.max(1, m),
      },
    ];
  }

  // Concise LCS matrix (efficient for single file editing scale)
  // Limit max matrix size to prevent extreme lag on very large files
  if (n > 2000 || m > 2000) {
    // Fallback to head/tail pruning comparison for huge files
    return computeSimpleLineDiff(origLines, modLines);
  }

  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0));

  for (let i = 1; i <= n; i++) {
    for (let j = 1; j <= m; j++) {
      if (origLines[i - 1] === modLines[j - 1]) {
        dp[i][j] = dp[i - 1][j - 1] + 1;
      } else {
        dp[i][j] = Math.max(dp[i - 1][j], dp[i][j - 1]);
      }
    }
  }

  // Backtrack to generate diff edits
  let i = n;
  let j = m;
  const edits: { op: "same" | "add" | "del"; origIdx: number; modIdx: number }[] = [];

  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && origLines[i - 1] === modLines[j - 1]) {
      edits.push({ op: "same", origIdx: i, modIdx: j });
      i--;
      j--;
    } else if (j > 0 && (i === 0 || dp[i][j - 1] >= dp[i - 1][j])) {
      edits.push({ op: "add", origIdx: i, modIdx: j });
      j--;
    } else {
      edits.push({ op: "del", origIdx: i, modIdx: j });
      i--;
    }
  }

  edits.reverse();

  // Merge contiguous edits into DiffLineChange
  const result: DiffLineChange[] = [];
  let k = 0;

  while (k < edits.length) {
    const edit = edits[k];
    if (edit.op === "same") {
      k++;
      continue;
    }

    const startK = k;
    let addCount = 0;
    let delCount = 0;
    let modStart = 0;
    let modEnd = 0;

    const origReplacedLines: string[] = [];

    while (k < edits.length && edits[k].op !== "same") {
      if (edits[k].op === "add") {
        addCount++;
        const line = edits[k].modIdx;
        if (!modStart || line < modStart) modStart = line;
        if (line > modEnd) modEnd = line;
      } else {
        delCount++;
        const oIdx = edits[k].origIdx;
        if (oIdx >= 1 && oIdx <= origLines.length) {
          origReplacedLines.push(origLines[oIdx - 1]);
        }
      }
      k++;
    }

    const origContent = origReplacedLines.join("\n");

    if (addCount > 0 && delCount > 0) {
      result.push({
        type: "modified",
        startLineNumber: modStart,
        endLineNumber: modEnd,
        originalText: origContent,
      });
    } else if (addCount > 0) {
      result.push({
        type: "added",
        startLineNumber: modStart,
        endLineNumber: modEnd,
        originalText: "",
      });
    } else if (delCount > 0) {
      // Pure deletion: mark on previous line or line 1
      const refLine = Math.min(Math.max(1, edits[startK].modIdx || 1), m);
      result.push({
        type: "deleted",
        startLineNumber: refLine,
        endLineNumber: refLine,
        originalText: origContent,
      });
    }
  }

  return result;
}

function computeSimpleLineDiff(origLines: string[], modLines: string[]): DiffLineChange[] {
  let start = 0;
  while (start < origLines.length && start < modLines.length && origLines[start] === modLines[start]) {
    start++;
  }
  let endOrig = origLines.length - 1;
  let endMod = modLines.length - 1;
  while (endOrig >= start && endMod >= start && origLines[endOrig] === modLines[endMod]) {
    endOrig--;
    endMod--;
  }

  if (start <= endMod) {
    return [
      {
        type: start <= endOrig ? "modified" : "added",
        startLineNumber: start + 1,
        endLineNumber: endMod + 1,
      },
    ];
  }

  return [];
}
