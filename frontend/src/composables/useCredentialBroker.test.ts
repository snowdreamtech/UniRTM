import { describe, it, expect, beforeEach, vi } from "vitest";
import { useCredentialBroker } from "./useCredentialBroker";
import type { CredentialRequest } from "../types/credential";

describe("useCredentialBroker composable", () => {
  beforeEach(() => {
    const broker = useCredentialBroker();
    broker.isPromptOpen.value = false;
    broker.pendingRequest.value = null;
    broker.clearSessionKeyring();
    vi.restoreAllMocks();
  });

  it("prompts modal when new credential request arrives", () => {
    const broker = useCredentialBroker();
    const req: CredentialRequest = {
      id: "req-1",
      type: "ssh",
      prompt: "Enter passphrase for key '/home/user/.ssh/id_rsa':",
      targetKey: "/home/user/.ssh/id_rsa",
      timestamp: Date.now(),
    };

    broker.promptCredential(req);
    expect(broker.isPromptOpen.value).toBe(true);
    expect(broker.pendingRequest.value?.id).toBe("req-1");
    expect(broker.pendingRequest.value?.type).toBe("ssh");
  });

  it("remembers passphrase in session keyring when requested", async () => {
    const broker = useCredentialBroker();
    const req: CredentialRequest = {
      id: "req-2",
      type: "gpg",
      prompt: "Enter GPG Passphrase:",
      targetKey: "gpg-key-0x1234",
      timestamp: Date.now(),
    };

    // mock fetch
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true }),
    } as any);

    broker.promptCredential(req);
    const success = await broker.respondCredential("req-2", "secret123", true);

    expect(success).toBe(true);
    expect(broker.isPromptOpen.value).toBe(false);
    expect(broker.sessionKeyring.value.get("gpg-key-0x1234")).toBe("secret123");
  });

  it("auto-responds when passphrase already exists in session keyring", () => {
    const broker = useCredentialBroker();
    broker.sessionKeyring.value.set("/home/user/.ssh/id_ed25519", "cached_pass");

    // mock fetch
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true }),
    } as any);

    const req: CredentialRequest = {
      id: "req-3",
      type: "ssh",
      prompt: "Enter passphrase for key '/home/user/.ssh/id_ed25519':",
      targetKey: "/home/user/.ssh/id_ed25519",
      timestamp: Date.now(),
    };

    broker.promptCredential(req);
    // 命中缓存时不打开弹窗
    expect(broker.isPromptOpen.value).toBe(false);
    expect(global.fetch).toHaveBeenCalledTimes(1);
  });

  it("handles user cancellation cleanly", async () => {
    const broker = useCredentialBroker();
    const req: CredentialRequest = {
      id: "req-4",
      type: "ssh",
      prompt: "Enter passphrase:",
      timestamp: Date.now(),
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true }),
    } as any);

    broker.promptCredential(req);
    expect(broker.isPromptOpen.value).toBe(true);

    await broker.cancelCredential();
    expect(broker.isPromptOpen.value).toBe(false);
    expect(broker.pendingRequest.value).toBeNull();
  });
});
