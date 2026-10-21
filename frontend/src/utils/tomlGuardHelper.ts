// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export interface BuiltinRuleCustomConfig {
  id: string;
  enabled: boolean;
  severity: "default" | "blocker" | "warning";
  max_file_size?: number;
  exclude_dirs: string[];
  exclude_files: string[];
  extensions: string[];
}

export interface ParsedGuardConfig {
  skip: string[];
  rules: Record<string, Partial<BuiltinRuleCustomConfig>>;
}

/**
 * Parses [guard] skip list and [guard.rules."<id>"] from TOML string without external parser
 */
export function parseGuardConfigFromToml(content: string): ParsedGuardConfig {
  const skip: string[] = [];
  const rules: Record<string, Partial<BuiltinRuleCustomConfig>> = {};

  if (!content) return { skip, rules };

  const lines = content.split("\n");
  let currentSection = "";
  let insideSkipArray = false;

  for (let i = 0; i < lines.length; i++) {
    const rawLine = lines[i];
    const line = rawLine.trim();

    // Skip comments
    if (line.startsWith("#")) continue;

    // Section header
    if (line.startsWith("[") && line.endsWith("]")) {
      insideSkipArray = false;
      currentSection = line.slice(1, -1).trim();
      continue;
    }

    // Inside [guard]
    if (currentSection === "guard") {
      if (line.startsWith("skip") && line.includes("=")) {
        const afterEq = line.substring(line.indexOf("=") + 1).trim();
        if (afterEq.startsWith("[")) {
          if (afterEq.endsWith("]")) {
            // single-line array
            const items = afterEq
              .slice(1, -1)
              .split(",")
              .map((s) => s.trim().replace(/^["']|["']$/g, ""))
              .filter(Boolean);
            skip.push(...items);
          } else {
            // multi-line array start
            insideSkipArray = true;
          }
        }
      } else if (insideSkipArray) {
        if (line.includes("]")) {
          const part = line.slice(0, line.indexOf("]")).trim();
          if (part) {
            const item = part.replace(/^["']|["'],?$/g, "").trim();
            if (item) skip.push(item);
          }
          insideSkipArray = false;
        } else {
          const item = line.replace(/^["']|["'],?$/g, "").trim();
          if (item) skip.push(item);
        }
      }
    }

    // Inside [guard.rules."<id>"] or [guard.rules.<id>]
    if (currentSection.startsWith("guard.rules.")) {
      const ruleID = currentSection.slice("guard.rules.".length).replace(/^["']|["']$/g, "");
      if (!rules[ruleID]) {
        rules[ruleID] = { id: ruleID };
      }
      const r = rules[ruleID];

      if (line.includes("=")) {
        const [kRaw, vRaw] = line.split("=").map((s) => s.trim());
        const key = kRaw;
        const val = vRaw.replace(/#.*$/, "").trim();

        if (key === "enabled") {
          r.enabled = val === "true";
        } else if (key === "severity") {
          const s = val.replace(/^["']|["']$/g, "");
          if (s === "blocker" || s === "warning") r.severity = s;
        } else if (key === "max_file_size") {
          const num = parseInt(val, 10);
          if (!isNaN(num)) r.max_file_size = num;
        } else if (key === "exclude_dirs" || key === "exclude_files" || key === "extensions") {
          if (val.startsWith("[") && val.endsWith("]")) {
            const list = val
              .slice(1, -1)
              .split(",")
              .map((s) => s.trim().replace(/^["']|["']$/g, ""))
              .filter(Boolean);
            if (key === "exclude_dirs") r.exclude_dirs = list;
            if (key === "exclude_files") r.exclude_files = list;
            if (key === "extensions") r.extensions = list;
          }
        }
      }
    }
  }

  return { skip, rules };
}

/**
 * Safely updates or removes a built-in rule configuration in TOML content
 */
export function updateBuiltinRuleInToml(content: string, config: BuiltinRuleCustomConfig): string {
  let lines = content ? content.split("\n") : [];
  const ruleID = config.id;

  // 1. Ensure [guard] section exists
  let guardHeaderIdx = lines.findIndex((l) => l.trim() === "[guard]");
  if (guardHeaderIdx === -1) {
    lines.push("");
    lines.push("[guard]");
    lines.push("enabled = true");
    guardHeaderIdx = lines.length - 2;
  }

  // Find end of [guard] section (next [section] or EOF)
  let guardEndIdx = lines.length;
  for (let i = guardHeaderIdx + 1; i < lines.length; i++) {
    const trimmed = lines[i].trim();
    if (trimmed.startsWith("[") && trimmed.endsWith("]")) {
      guardEndIdx = i;
      break;
    }
  }

  // 2. Handle skip array
  const guardLines = lines.slice(guardHeaderIdx, guardEndIdx);
  let skipLineIdx = -1;
  let skipMultiLineEnd = -1;

  for (let i = 0; i < guardLines.length; i++) {
    const l = guardLines[i].trim();
    if (l.startsWith("skip") && l.includes("=")) {
      skipLineIdx = guardHeaderIdx + i;
      const after = l.substring(l.indexOf("=") + 1).trim();
      if (after.startsWith("[") && !after.endsWith("]")) {
        for (let j = i + 1; j < guardLines.length; j++) {
          if (guardLines[j].includes("]")) {
            skipMultiLineEnd = guardHeaderIdx + j;
            break;
          }
        }
      }
      break;
    }
  }

  // Parse current skip items
  const parsed = parseGuardConfigFromToml(lines.join("\n"));
  let currentSkip = new Set<string>(parsed.skip);

  if (config.enabled) {
    currentSkip.delete(ruleID);
  } else {
    currentSkip.add(ruleID);
  }

  const updatedSkipArray = Array.from(currentSkip);

  // Reformat skip line
  if (skipLineIdx !== -1) {
    const end = skipMultiLineEnd !== -1 ? skipMultiLineEnd : skipLineIdx;
    if (updatedSkipArray.length === 0) {
      // Remove skip setting if empty
      lines.splice(skipLineIdx, end - skipLineIdx + 1);
    } else {
      const skipBlock =
        updatedSkipArray.length <= 2
          ? `skip = [${updatedSkipArray.map((s) => `"${s}"`).join(", ")}]`
          : `skip = [\n  ${updatedSkipArray.map((s) => `"${s}"`).join(",\n  ")},\n]`;
      lines.splice(skipLineIdx, end - skipLineIdx + 1, ...skipBlock.split("\n"));
    }
  } else if (updatedSkipArray.length > 0) {
    // Insert skip under [guard]
    const insertAt = guardHeaderIdx + 1;
    const skipBlock =
      updatedSkipArray.length <= 2
        ? `skip = [${updatedSkipArray.map((s) => `"${s}"`).join(", ")}]`
        : `skip = [\n  ${updatedSkipArray.map((s) => `"${s}"`).join(",\n  ")},\n]`;
    lines.splice(insertAt, 0, ...skipBlock.split("\n"));
  }

  // 3. Handle [guard.rules."<id>"] section
  const targetHeader = `[guard.rules."${ruleID}"]`;
  let ruleSectionStart = -1;
  let ruleSectionEnd = -1;

  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();
    if (trimmed === targetHeader || trimmed === `[guard.rules.'${ruleID}']` || trimmed === `[guard.rules.${ruleID}]`) {
      ruleSectionStart = i;
      ruleSectionEnd = lines.length;
      for (let j = i + 1; j < lines.length; j++) {
        if (lines[j].trim().startsWith("[") && lines[j].trim().endsWith("]")) {
          ruleSectionEnd = j;
          break;
        }
      }
      break;
    }
  }

  // Determine if rule has custom properties other than defaults
  const hasCustomProps =
    (config.severity && config.severity !== "default") ||
    (config.max_file_size && config.max_file_size > 0) ||
    (config.exclude_dirs && config.exclude_dirs.length > 0) ||
    (config.exclude_files && config.exclude_files.length > 0) ||
    (config.extensions && config.extensions.length > 0);

  if (hasCustomProps) {
    const newSectionLines: string[] = ["", targetHeader];
    if (config.severity && config.severity !== "default") {
      newSectionLines.push(`severity = "${config.severity}"`);
    }
    if (config.max_file_size && config.max_file_size > 0) {
      newSectionLines.push(`max_file_size = ${config.max_file_size}`);
    }
    if (config.exclude_dirs && config.exclude_dirs.length > 0) {
      newSectionLines.push(`exclude_dirs = [${config.exclude_dirs.map((d: string) => `"${d}"`).join(", ")}]`);
    }
    if (config.exclude_files && config.exclude_files.length > 0) {
      newSectionLines.push(`exclude_files = [${config.exclude_files.map((f: string) => `"${f}"`).join(", ")}]`);
    }
    if (config.extensions && config.extensions.length > 0) {
      newSectionLines.push(`extensions = [${config.extensions.map((e: string) => `"${e}"`).join(", ")}]`);
    }

    if (ruleSectionStart !== -1) {
      lines.splice(ruleSectionStart, ruleSectionEnd - ruleSectionStart, ...newSectionLines.slice(1));
    } else {
      lines.push(...newSectionLines);
    }
  } else if (ruleSectionStart !== -1) {
    // Clean up empty section if no custom props
    // Also remove preceding empty line if any
    let startDel = ruleSectionStart;
    if (startDel > 0 && lines[startDel - 1].trim() === "") {
      startDel--;
    }
    lines.splice(startDel, ruleSectionEnd - startDel);
  }

  return lines.join("\n").replace(/\n{3,}/g, "\n\n");
}
