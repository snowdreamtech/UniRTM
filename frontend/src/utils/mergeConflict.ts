// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export interface MergeConflictRegion {
  id: string;
  startLine: number; // 1-based line of <<<<<<<
  separatorLine: number; // 1-based line of =======
  endLine: number; // 1-based line of >>>>>>>
  currentLabel: string;
  incomingLabel: string;
  currentContent: string;
  incomingContent: string;
}

export type ConflictResolution = "current" | "incoming" | "both";

/**
 * Parses Git conflict markers (<<<<<<<, =======, >>>>>>>) from text content.
 * Robust against arbitrary indentation, empty blocks, or missing labels.
 */
export function parseMergeConflicts(text: string): MergeConflictRegion[] {
  if (!text || !text.includes("<<<<<<<")) {
    return [];
  }

  const lines = text.split("\n");
  const conflicts: MergeConflictRegion[] = [];

  let inConflict = false;
  let startLine = -1;
  let separatorLine = -1;
  let currentLabel = "";
  let incomingLabel = "";
  let currentLines: string[] = [];
  let incomingLines: string[] = [];
  let conflictIdx = 0;

  for (let i = 0; i < lines.length; i++) {
    const lineNum = i + 1;
    const line = lines[i];

    if (!inConflict) {
      if (line.startsWith("<<<<<<<")) {
        inConflict = true;
        startLine = lineNum;
        separatorLine = -1;
        currentLabel = line.slice(7).trim() || "HEAD (Current Change)";
        incomingLabel = "";
        currentLines = [];
        incomingLines = [];
      }
    } else {
      if (separatorLine === -1) {
        if (line.startsWith("=======")) {
          separatorLine = lineNum;
        } else {
          currentLines.push(line);
        }
      } else {
        if (line.startsWith(">>>>>>>")) {
          const endLine = lineNum;
          incomingLabel = line.slice(7).trim() || "Incoming Change";
          conflictIdx++;
          conflicts.push({
            id: `conflict-${conflictIdx}-${startLine}`,
            startLine,
            separatorLine,
            endLine,
            currentLabel,
            incomingLabel,
            currentContent: currentLines.join("\n"),
            incomingContent: incomingLines.join("\n"),
          });
          inConflict = false;
        } else {
          incomingLines.push(line);
        }
      }
    }
  }

  return conflicts;
}

/**
 * Computes replacement text for a conflict region given resolution strategy.
 */
export function getConflictResolutionText(
  conflict: MergeConflictRegion,
  resolution: ConflictResolution
): string {
  switch (resolution) {
    case "current":
      return conflict.currentContent;
    case "incoming":
      return conflict.incomingContent;
    case "both":
      if (conflict.currentContent && conflict.incomingContent) {
        return `${conflict.currentContent}\n${conflict.incomingContent}`;
      }
      return conflict.currentContent || conflict.incomingContent;
  }
}
