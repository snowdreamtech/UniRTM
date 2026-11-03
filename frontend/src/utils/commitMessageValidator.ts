// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export interface CommitMessageValidation {
  isValid: boolean;
  type?: string;
  scope?: string;
  isBreaking?: boolean;
  description?: string;
  warning?: string;
  canAutoFix?: boolean;
  fixedMessage?: string;
}

const CONVENTIONAL_REGEX =
  /^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(([a-zA-Z0-9_.\-/]+)\))?(!)?: (.+)$/;
const SPECIAL_PREFIXES = ["Merge ", "Revert ", "fixup! ", "squash! ", "amend! "];
const VALID_TYPES = ["feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"];

/**
 * Validates a commit message against Conventional Commits specification.
 * Returns validation status, warnings, and optional auto-fix.
 */
export function validateCommitMessage(message: string): CommitMessageValidation {
  const trimmed = message.trim();
  if (!trimmed) {
    return { isValid: false };
  }

  // 1. Check for special Git automated commits
  for (const prefix of SPECIAL_PREFIXES) {
    if (trimmed.startsWith(prefix)) {
      return { isValid: true };
    }
  }

  const lines = trimmed.split(/\r?\n/);
  const firstLine = lines[0];

  // 2. Multi-line blank line check
  if (lines.length > 1 && lines[1].trim() !== "") {
    const fixedLines = [lines[0], "", ...lines.slice(1)];
    return {
      isValid: false,
      warning: "Second line must be empty in multi-line commit messages",
      canAutoFix: true,
      fixedMessage: fixedLines.join("\n"),
    };
  }

  // 3. Check for missing space after colon (e.g., feat:something)
  const missingSpaceMatch = firstLine.match(/^([a-z]+(\([a-zA-Z0-9_.\-/]+\))?(!)?):([^\s].*)$/);
  if (missingSpaceMatch) {
    const prefix = missingSpaceMatch[1];
    const rest = missingSpaceMatch[4];
    const fixedFirstLine = `${prefix}: ${rest}`;
    const fixed = [fixedFirstLine, ...lines.slice(1)].join("\n");
    return {
      isValid: false,
      warning: "Missing space after colon",
      canAutoFix: true,
      fixedMessage: fixed,
    };
  }

  // 4. Conventional format match
  const match = firstLine.match(CONVENTIONAL_REGEX);
  if (!match) {
    // Check if starts with a known type but wrong separator
    const startsWithType = VALID_TYPES.find((t) => firstLine.startsWith(t));
    if (startsWithType) {
      return {
        isValid: false,
        warning: `Format error, recommend following 'type(scope): description', e.g. ${startsWithType}: description`,
        canAutoFix: false,
      };
    }
    return {
      isValid: false,
      warning: "Recommend following conventional format (e.g. 'feat: new feature' or 'fix(scope): fix bug')",
      canAutoFix: false,
    };
  }

  const [, type, , scope, , description] = match;
  const isBreaking = match[0].includes("!:");

  // 5. Ending period check on first line
  if (description.endsWith(".") || description.endsWith("。")) {
    const cleanDesc = description.slice(0, -1);
    const prefix = scope ? `${type}(${scope})${isBreaking ? "!" : ""}: ` : `${type}${isBreaking ? "!" : ""}: `;
    const fixedFirstLine = prefix + cleanDesc;
    const fixed = [fixedFirstLine, ...lines.slice(1)].join("\n");
    return {
      isValid: false,
      warning: "Subject should not end with a period",
      canAutoFix: true,
      fixedMessage: fixed,
    };
  }

  // 6. Header max length enforcement (max 120)
  if (firstLine.length > 120) {
    return {
      isValid: false,
      type,
      scope,
      isBreaking,
      description,
      warning: "Header exceeds 120 characters limit",
      canAutoFix: false,
    };
  }

  // 7. First line length advisory (soft recommendation: 72 chars)
  if (firstLine.length > 72) {
    return {
      isValid: true,
      type,
      scope,
      isBreaking,
      description,
      warning: "Header exceeds 72 characters, recommend concise summary",
    };
  }

  return {
    isValid: true,
    type,
    scope,
    isBreaking,
    description,
  };
}
