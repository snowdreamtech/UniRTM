import { t } from "../i18n";

export interface UpdateProgressPayload {
  percentage?: number;
  status?: string;
  stage?: string;
  detail?: string;
  loadedBytes?: number;
  totalBytes?: number;
}

/**
 * Formats structured or legacy update progress into a localized status string.
 * @param p The update progress payload from backend events.
 * @param fallbackPercentage Optional percentage to display if payload percentage is absent.
 * @returns Localized progress description text.
 */
export function formatUpdateProgress(p: UpdateProgressPayload | null, fallbackPercentage: number = 0): string {
  if (!p) {
    return t("about.preparingDownload");
  }

  const stage = p.stage;
  if (stage === "checking") {
    return t("about.checkingForUpdate");
  }
  if (stage === "found_asset") {
    return t("about.foundReleaseAsset", { asset: p.detail || "" });
  }
  if (stage === "downloading") {
    if (p.totalBytes && p.totalBytes > 0 && p.loadedBytes != null) {
      const loadedMb = (p.loadedBytes / 1048576).toFixed(1);
      const totalMb = (p.totalBytes / 1048576).toFixed(1);
      const pct = p.percentage != null ? p.percentage : Math.round((p.loadedBytes / p.totalBytes) * 100);
      return t("about.downloadingWithTotal", { loaded: loadedMb, total: totalMb, progress: pct });
    }
    if (p.loadedBytes != null && p.loadedBytes > 0) {
      const loadedMb = (p.loadedBytes / 1048576).toFixed(1);
      return t("about.downloadingSizeOnly", { loaded: loadedMb });
    }
    return t("about.downloadingGuiUpdate", { progress: p.percentage || fallbackPercentage || 0 });
  }
  if (stage === "download_completed") {
    return t("about.downloadCompleted");
  }
  if (stage === "mounting") {
    return t("about.mountingDiskImage");
  }
  if (stage === "extracting") {
    return t("about.extractingPackage", { name: p.detail || "..." });
  }
  if (stage === "staging") {
    return t("about.stagingPackage");
  }
  if (stage === "preparing_script") {
    return t("about.preparingApplyScript");
  }
  if (stage === "verifying_checksum") {
    return t("about.verifyingChecksum");
  }
  if (stage === "ready") {
    return t("about.updateCompleteRestart");
  }

  // Fallback regex parsing for legacy status strings
  const raw = p.status || "";
  if (!raw) {
    return t("about.downloading", { progress: p.percentage || fallbackPercentage || 0 });
  }

  const matchBoth = raw.match(/Downloading update:\s*([\d.]+)\s*MB\s*\/\s*([\d.]+)\s*MB\s*\((\d+)%\)/i);
  if (matchBoth) {
    return t("about.downloadingWithTotal", { loaded: matchBoth[1], total: matchBoth[2], progress: matchBoth[3] });
  }

  const matchLoaded = raw.match(/Downloading update:\s*([\d.]+)\s*MB/i);
  if (matchLoaded) {
    return t("about.downloadingSizeOnly", { loaded: matchLoaded[1] });
  }

  if (/Download completed/i.test(raw)) {
    return t("about.downloadCompleted");
  }
  if (/Mounting disk image/i.test(raw)) {
    return t("about.mountingDiskImage");
  }
  const matchExtract = raw.match(/Extracting\s+(.+?)\.\.\./i);
  if (matchExtract) {
    return t("about.extractingPackage", { name: matchExtract[1] });
  }
  if (/Extracting Windows update package/i.test(raw) || /Staging Linux update package/i.test(raw)) {
    return t("about.stagingPackage");
  }
  if (/Preparing update apply script/i.test(raw)) {
    return t("about.preparingApplyScript");
  }
  if (/Verifying download checksum/i.test(raw)) {
    return t("about.verifyingChecksum");
  }
  if (/Checking for latest release/i.test(raw)) {
    return t("about.checkingForUpdate");
  }
  const matchAsset = raw.match(/Found release asset:\s*(.+)/i);
  if (matchAsset) {
    return t("about.foundReleaseAsset", { asset: matchAsset[1] });
  }
  if (/Update ready! Restart application to apply/i.test(raw)) {
    return t("about.updateCompleteRestart");
  }

  return raw;
}
