import { describe, it, expect } from "vitest";
import { parseMergeConflicts, getConflictResolutionText } from "./mergeConflict";

describe("mergeConflict utility", () => {
  it("returns empty array when no conflict markers exist", () => {
    const text = "const a = 1;\nconst b = 2;\n";
    expect(parseMergeConflicts(text)).toEqual([]);
    expect(parseMergeConflicts("")).toEqual([]);
  });

  it("parses single standard Git merge conflict correctly", () => {
    const sample = [
      "function test() {",
      "<<<<<<< HEAD",
      "  console.log('version 1');",
      "=======",
      "  console.log('version 2');",
      ">>>>>>> feature-branch",
      "}",
    ].join("\n");

    const conflicts = parseMergeConflicts(sample);
    expect(conflicts).toHaveLength(1);

    const c = conflicts[0];
    expect(c.startLine).toBe(2);
    expect(c.separatorLine).toBe(4);
    expect(c.endLine).toBe(6);
    expect(c.currentLabel).toBe("HEAD");
    expect(c.incomingLabel).toBe("feature-branch");
    expect(c.currentContent).toBe("  console.log('version 1');");
    expect(c.incomingContent).toBe("  console.log('version 2');");
  });

  it("parses multiple conflict blocks in a single file", () => {
    const sample = [
      "<<<<<<< main",
      "let x = 10;",
      "=======",
      "let x = 20;",
      ">>>>>>> dev",
      "// some middle code",
      "<<<<<<< main",
      "let y = 1;",
      "=======",
      "let y = 2;",
      ">>>>>>> dev",
    ].join("\n");

    const conflicts = parseMergeConflicts(sample);
    expect(conflicts).toHaveLength(2);
    expect(conflicts[0].startLine).toBe(1);
    expect(conflicts[1].startLine).toBe(7);
  });

  it("computes resolution replacement text accurately", () => {
    const sample = [
      "<<<<<<< HEAD",
      "first",
      "=======",
      "second",
      ">>>>>>> remote",
    ].join("\n");

    const conflicts = parseMergeConflicts(sample);
    const c = conflicts[0];

    expect(getConflictResolutionText(c, "current")).toBe("first");
    expect(getConflictResolutionText(c, "incoming")).toBe("second");
    expect(getConflictResolutionText(c, "both")).toBe("first\nsecond");
  });
});
