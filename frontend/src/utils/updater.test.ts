import { describe, it, expect, beforeEach } from "vitest";
import { formatUpdateProgress } from "./updater";
import { setLanguage } from "../i18n";

describe("updater utility", () => {
  beforeEach(async () => {
    await setLanguage("en-US");
  });

  it("handles null payload with preparingDownload text", () => {
    expect(formatUpdateProgress(null)).toBe("Preparing download...");
  });

  it("formats stage: checking", () => {
    expect(formatUpdateProgress({ stage: "checking" })).toBe("Checking for latest release...");
  });

  it("formats stage: found_asset with detail", () => {
    expect(formatUpdateProgress({ stage: "found_asset", detail: "unigodesktop.dmg" })).toBe(
      "Found release asset: unigodesktop.dmg"
    );
  });

  it("formats stage: downloading with total bytes", () => {
    const res = formatUpdateProgress({
      stage: "downloading",
      loadedBytes: 10 * 1048576,
      totalBytes: 20 * 1048576,
      percentage: 50,
    });
    expect(res).toBe("Downloading update: 10.0 MB / 20.0 MB (50%)");
  });

  it("formats stage: verifying_checksum", () => {
    expect(formatUpdateProgress({ stage: "verifying_checksum" })).toBe("Verifying download checksum...");
  });

  it("formats stage: download_completed", () => {
    expect(formatUpdateProgress({ stage: "download_completed" })).toBe("Download completed");
  });

  it("formats stage: mounting", () => {
    expect(formatUpdateProgress({ stage: "mounting" })).toBe("Mounting disk image...");
  });

  it("formats stage: extracting with name", () => {
    expect(formatUpdateProgress({ stage: "extracting", detail: "core.tar.gz" })).toBe(
      "Extracting update package: core.tar.gz..."
    );
  });

  it("formats stage: staging", () => {
    expect(formatUpdateProgress({ stage: "staging" })).toBe("Staging update package...");
  });

  it("formats stage: preparing_script", () => {
    expect(formatUpdateProgress({ stage: "preparing_script" })).toBe("Preparing update apply script...");
  });

  it("formats stage: ready", () => {
    expect(formatUpdateProgress({ stage: "ready" })).toBe("New version ready. Restart the application to apply.");
  });

  it("localizes correctly into zh-CN", async () => {
    await setLanguage("zh-CN");
    expect(formatUpdateProgress({ stage: "verifying_checksum" })).toBe("正在校验更新包完整性...");
    expect(formatUpdateProgress({ stage: "mounting" })).toBe("正在挂载磁盘镜像...");
  });
});
