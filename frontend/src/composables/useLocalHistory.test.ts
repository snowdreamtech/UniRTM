import { describe, it, expect, beforeEach } from "vitest";
import { useLocalHistory } from "./useLocalHistory";
import { useWorkspaceBus } from "./useWorkspaceBus";

describe("useLocalHistory composable", () => {
  beforeEach(() => {
    const { clearSnapshots, closeLocalHistory } = useLocalHistory();
    clearSnapshots();
    closeLocalHistory();
  });

  it("records snapshots and maintains chronological order", () => {
    const { recordSnapshot, getSnapshots } = useLocalHistory();

    const s1 = recordSnapshot("src/main.ts", "console.log('v1')");
    expect(s1).not.toBeNull();
    expect(s1?.content).toBe("console.log('v1')");

    const s2 = recordSnapshot("src/main.ts", "console.log('v2')");
    expect(s2).not.toBeNull();
    expect(s2?.content).toBe("console.log('v2')");

    const snapshots = getSnapshots("src/main.ts");
    expect(snapshots).toHaveLength(2);
    expect(snapshots[0].content).toBe("console.log('v2')");
    expect(snapshots[1].content).toBe("console.log('v1')");
  });

  it("deduplicates identical consecutive snapshots", () => {
    const { recordSnapshot, getSnapshots } = useLocalHistory();

    recordSnapshot("src/test.ts", "same content");
    recordSnapshot("src/test.ts", "same content");

    const snapshots = getSnapshots("src/test.ts");
    expect(snapshots).toHaveLength(1);
  });

  it("caps maximum history snapshots to 20 per file", () => {
    const { recordSnapshot, getSnapshots } = useLocalHistory();

    for (let i = 0; i < 25; i++) {
      recordSnapshot("src/large.ts", `version ${i}`);
    }

    const snapshots = getSnapshots("src/large.ts");
    expect(snapshots).toHaveLength(20);
    expect(snapshots[0].content).toBe("version 24");
  });

  it("delegates diff comparison to workspace bus", () => {
    const { recordSnapshot, compareWithCurrent } = useLocalHistory();
    const { onOpenDiff } = useWorkspaceBus();

    let capturedReq: any = null;
    const cleanup = onOpenDiff((req) => {
      capturedReq = req;
    });

    const snap = recordSnapshot("src/app.ts", "old content")!;
    compareWithCurrent(snap, "new content");

    expect(capturedReq).not.toBeNull();
    expect(capturedReq.path).toBe("src/app.ts");
    expect(capturedReq.originalContent).toBe("old content");
    expect(capturedReq.modifiedContent).toBe("new content");

    cleanup();
  });
});
