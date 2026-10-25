import { describe, it, expect } from "vitest";
import type { GitProfile } from "../../types/credential";

function formatProfileTooltip(profile?: GitProfile): string {
  if (!profile) {
    return "Git Identity: Default (Click to switch)";
  }
  const email = profile.userEmail ? ` <${profile.userEmail}>` : "";
  const sign = profile.signingKey ? " • GPG Signed" : "";
  return `Git Identity: ${profile.name} (${profile.userName}${email})${sign} • Click to switch`;
}

function getDisplayProfileName(profile?: GitProfile): string {
  if (profile && profile.name) {
    return profile.name;
  }
  return "Default Profile";
}

describe("GitProfileStatus formatting and status logic", () => {
  it("formats fallback tooltip when no profile is active", () => {
    const tooltip = formatProfileTooltip(undefined);
    expect(tooltip).toBe("Git Identity: Default (Click to switch)");
    expect(getDisplayProfileName(undefined)).toBe("Default Profile");
  });

  it("formats active profile tooltip with GPG signed indicator", () => {
    const profile: GitProfile = {
      id: "prof-work",
      name: "Corporate Work",
      userName: "Alice Doe",
      userEmail: "alice@company.com",
      signingKey: "3AA5C34371567BD2",
      gpgSign: true,
      isDefault: true,
      updatedAt: 1000,
    };

    const tooltip = formatProfileTooltip(profile);
    expect(tooltip).toContain("Git Identity: Corporate Work (Alice Doe <alice@company.com>)");
    expect(tooltip).toContain("• GPG Signed");
    expect(getDisplayProfileName(profile)).toBe("Corporate Work");
  });

  it("formats profile tooltip without GPG when not signed", () => {
    const profile: GitProfile = {
      id: "prof-oss",
      name: "Open Source",
      userName: "Contributor",
      userEmail: "contributor@oss.org",
      gpgSign: false,
      isDefault: false,
      updatedAt: 2000,
    };

    const tooltip = formatProfileTooltip(profile);
    expect(tooltip).toContain("Git Identity: Open Source (Contributor <contributor@oss.org>)");
    expect(tooltip).not.toContain("GPG Signed");
  });
});
