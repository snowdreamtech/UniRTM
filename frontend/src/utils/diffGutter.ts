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
 * 计算原始文本与当前修改文本之间的行级增量变动（基于 Myers/动态规划行比较）
 * 用于在主编辑器行号槽（Gutter）呈现克制的增加（绿色）、修改（蓝色）指示
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

  // 如果原本为空，全部为新增
  if (originalText === "" && modifiedText !== "") {
    return [
      {
        type: "added",
        startLineNumber: 1,
        endLineNumber: Math.max(1, m),
      },
    ];
  }

  // 经典简明 LCS 矩阵（针对单文件编辑规模极速高效）
  // 限制最大矩阵规模，防止百万行极端卡顿
  if (n > 2000 || m > 2000) {
    // 大文件回退至行首行尾剪枝简易比较
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

  // 回溯生成 diff edits
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

  // 合并连续 edits 为 DiffLineChange
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
      // 纯删除行，标记在前一行或第 1 行
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
