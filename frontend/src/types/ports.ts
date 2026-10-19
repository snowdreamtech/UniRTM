// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export type PortSource = "auto" | "user";
export type PortStatus = "active" | "inactive";
export type PortVisibility = "private" | "public";
export type PortProtocol = "http" | "https";

export interface PortProbeInfo {
  port: number;
  isHttps: boolean;
  httpStatus?: number;
  serverHeader?: string;
  xPoweredBy?: string;
  location?: string;
  detectedService?: string;
  probeDurationMs?: number;
}

export interface WorkspaceServiceHint {
  port: number;
  name: string;
  sourceFile: string;
  serviceType?: "compose" | "npm-script" | "env-var" | "custom";
  framework?: string;
}

export interface SystemListeningPort {
  port: number;
  protocol: string;
  localAddress: string;
  pid?: number;
  process?: string;
  commandLine?: string;
  probe?: PortProbeInfo;
}

export interface PortForwardItem {
  id: string; // e.g., "port-5173"
  port: number;
  localPort: number;
  label: string;
  protocol: PortProtocol;
  source: PortSource;
  status: PortStatus;
  visibility: PortVisibility;
  process?: string;
  commandLine?: string;
  pid?: number;
  forwardAddress: string;
  previewUrl: string;
  icon?: string;
  probe?: PortProbeInfo;
  workspaceHint?: WorkspaceServiceHint;
}

export interface PortForwardPayload {
  port: number;
  localPort?: number;
  label?: string;
  protocol?: PortProtocol;
  visibility?: PortVisibility;
}

export type PortRiskLevel = "high" | "standard";

export interface PortRiskAssessment {
  level: PortRiskLevel;
  serviceCategory: string;
  reason: string;
  recommendation: string;
}

