import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import { useGitIdentity } from "./useGitIdentity";

describe("useGitIdentity", () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("loads profiles and computes active profile", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        profiles: [
          {
            id: "prof-work",
            name: "Work",
            userName: "Alice",
            userEmail: "alice@work.com",
            gpgSign: true,
            isDefault: true,
            updatedAt: 12345,
          },
        ],
        activeProfileId: "prof-work",
      }),
    } as unknown as Response);

    const { profiles, activeProfileId, activeProfile, loadProfiles } = useGitIdentity();
    await loadProfiles();

    expect(profiles.value).toHaveLength(1);
    expect(activeProfileId.value).toBe("prof-work");
    expect(activeProfile.value?.name).toBe("Work");
    expect(activeProfile.value?.userEmail).toBe("alice@work.com");
  });

  it("saves a profile and reloads", async () => {
    globalThis.fetch = vi.fn().mockImplementation((_url, options) => {
      if (options?.method === "POST") {
        return Promise.resolve({ ok: true });
      }
      return Promise.resolve({
        ok: true,
        json: async () => ({
          profiles: [
            {
              id: "prof-saved",
              name: "Saved",
              userName: "Bob",
              userEmail: "bob@oss.org",
              gpgSign: false,
              isDefault: false,
              updatedAt: 1,
            },
          ],
          activeProfileId: "prof-saved",
        }),
      });
    });

    const { saveProfile, profiles } = useGitIdentity();
    const success = await saveProfile({ name: "Saved", userName: "Bob" });

    expect(success).toBe(true);
    expect(profiles.value).toHaveLength(1);
    expect(profiles.value[0].userName).toBe("Bob");
  });

  it("applies a profile to the current workspace", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, profileId: "prof-applied" }),
    } as unknown as Response);

    const { applyProfile, activeProfileId } = useGitIdentity();
    const success = await applyProfile("prof-applied", "/path/to/repo");

    expect(success).toBe(true);
    expect(activeProfileId.value).toBe("prof-applied");
  });

  it("loads native agent status and audit records", async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes("agent-status")) {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            sshAgentActive: true,
            sshAgentSock: "ssh-agent.sock",
            sshAgentKeys: 2,
            gpgAgentActive: true,
          }),
        });
      }
      if (url.includes("credential-audit")) {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            records: [
              {
                id: "rec-1",
                timestamp: 1000,
                action: "ssh_push",
                type: "ssh",
                targetKey: "id_ed25519",
                result: "success",
                clientIp: "127.0.0.1",
              },
            ],
          }),
        });
      }
      return Promise.resolve({ ok: false });
    });

    const { agentStatus, auditRecords, loadAgentStatus, loadAuditRecords } = useGitIdentity();

    await loadAgentStatus();
    expect(agentStatus.value.sshAgentActive).toBe(true);
    expect(agentStatus.value.sshAgentKeys).toBe(2);

    await loadAuditRecords();
    expect(auditRecords.value).toHaveLength(1);
    expect(auditRecords.value[0].action).toBe("ssh_push");
  });
});
