export type CredentialType = "ssh" | "gpg" | "git-auth";

export interface CredentialRequest {
  id: string; // 唯一凭据事务 ID (如 "cred-1729312345-abc")
  type: CredentialType; // "ssh" | "gpg" | "git-auth"
  prompt: string; // 原始提示文本 (例如 "Enter passphrase for key '/root/.ssh/id_ed25519':")
  targetKey?: string; // 解析出的具体密钥标识或路径 (例如 "/home/user/.ssh/id_ed25519")
  timestamp: number;
}

export interface CredentialResponse {
  id: string;
  passphrase: string;
  rememberSession: boolean; // 是否在当前会话中记住此密码
  canceled?: boolean; // 用户是否取消
}
