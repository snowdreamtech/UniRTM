export type CredentialType = "ssh" | "gpg" | "git-auth";

export interface CredentialRequest {
  id: string; // Unique credential transaction ID (e.g. "cred-1729312345-abc")
  type: CredentialType; // "ssh" | "gpg" | "git-auth"
  prompt: string; // Raw prompt text (e.g. "Enter passphrase for key '/root/.ssh/id_ed25519':")
  targetKey?: string; // Parsed key identifier or path (e.g. "/home/user/.ssh/id_ed25519")
  timestamp: number;
}

export interface CredentialResponse {
  id: string;
  passphrase: string;
  rememberSession: boolean; // Whether to remember this passphrase in the current session
  canceled?: boolean; // Whether user canceled the prompt
}

export interface GitProfile {
  id: string;
  name: string;
  userName: string;
  userEmail: string;
  sshKeyPath?: string;
  signingKey?: string;
  gpgSign: boolean;
  isDefault: boolean;
  updatedAt: number;
}

export interface SystemSSHKey {
  name: string;
  path: string;
  algorithm: string;
  comment: string;
}

export interface SystemGPGKey {
  keyId: string;
  shortKeyId?: string;
  userId: string;
  algo?: string;
  created?: string;
}

export interface SystemKeysSummary {
  sshKeys: SystemSSHKey[];
  gpgKeys: SystemGPGKey[];
}

export interface NativeAgentStatus {
  sshAgentActive: boolean;
  sshAgentSock?: string;
  sshAgentKeys: number;
  gpgAgentActive: boolean;
}

export interface CredentialAuditRecord {
  id: string;
  timestamp: number;
  action: string;
  type: string;
  targetKey: string;
  result: string;
  clientIp: string;
}
