import type { PortProbeInfo, WorkspaceServiceHint, PortRiskAssessment } from "../types/ports";

export interface ServiceRecognitionResult {
  label: string;
  icon: string;
  defaultProtocol: "http" | "https";
  serviceType: string;
  sourceOrigin?: "workspace" | "probe" | "process" | "well-known";
}

export function recognizePortService(
  port: number,
  processName?: string,
  cmdline?: string,
  probe?: PortProbeInfo,
  workspaceHint?: WorkspaceServiceHint
): ServiceRecognitionResult {
  const pName = (processName || "").toLowerCase();
  const cmd = (cmdline || "").toLowerCase();
  const isHttps = probe?.isHttps || port === 443 || port === 8443;

  // 1. Method C: Workspace Context Hint (Highest priority for business semantic naming)
  if (workspaceHint && workspaceHint.name) {
    let icon = "terminal";
    const fw = (workspaceHint.framework || "").toLowerCase();
    const sType = workspaceHint.serviceType || "custom";

    if (fw.includes("next") || fw.includes("react")) icon = "react";
    else if (fw.includes("vue") || fw.includes("nuxt")) icon = "vue";
    else if (fw.includes("vite")) icon = "vite";
    else if (fw.includes("postgres") || fw.includes("mysql") || fw.includes("redis") || fw.includes("mongo"))
      icon = "database";
    else if (fw.includes("python") || fw.includes("fastapi") || fw.includes("flask") || fw.includes("django"))
      icon = "python";
    else if (fw.includes("node") || fw.includes("express")) icon = "node";

    return {
      label: workspaceHint.name,
      icon,
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: sType === "compose" ? "database" : "frontend",
      sourceOrigin: "workspace",
    };
  }

  // 2. Method B: Network Application Protocol Probe (Highest accuracy for framework & protocol)
  if (probe && (probe.detectedService || probe.isHttps)) {
    const detected = probe.detectedService || "";
    const detLower = detected.toLowerCase();

    let icon = "terminal";
    let serviceType = "backend";

    if (
      detLower.includes("fastapi") ||
      detLower.includes("uvicorn") ||
      detLower.includes("flask") ||
      detLower.includes("gunicorn")
    ) {
      icon = "python";
      serviceType = "backend";
    } else if (detLower.includes("next")) {
      icon = "nextjs";
      serviceType = "frontend";
    } else if (detLower.includes("express") || detLower.includes("node")) {
      icon = "node";
      serviceType = "backend";
    } else if (detLower.includes("vite")) {
      icon = "vite";
      serviceType = "frontend";
    } else if (detLower.includes("redis") || detLower.includes("mysql") || detLower.includes("mariadb")) {
      icon = "database";
      serviceType = "database";
    } else if (detLower.includes("nginx") || detLower.includes("caddy") || detLower.includes("apache")) {
      icon = "terminal";
      serviceType = "gateway";
    }

    if (detected) {
      return {
        label: detected,
        icon,
        defaultProtocol: isHttps ? "https" : "http",
        serviceType,
        sourceOrigin: "probe",
      };
    }
  }

  // 3. Method A: Process and Command Line Inspection
  if (cmd.includes("vite") || pName.includes("vite")) {
    return {
      label: "Vite Dev Server",
      icon: "vite",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "frontend",
      sourceOrigin: "process",
    };
  }

  if (cmd.includes("next") || cmd.includes("next-server")) {
    return {
      label: "Next.js App",
      icon: "nextjs",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "frontend",
      sourceOrigin: "process",
    };
  }

  if (cmd.includes("nuxt") || cmd.includes("nuxt-server")) {
    return {
      label: "Nuxt App",
      icon: "vue",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "frontend",
      sourceOrigin: "process",
    };
  }

  if (cmd.includes("jupyter") || pName.includes("jupyter")) {
    return {
      label: "Jupyter Notebook",
      icon: "python",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "datascience",
      sourceOrigin: "process",
    };
  }

  if (cmd.includes("ollama") || pName.includes("ollama")) {
    return {
      label: "Ollama LLM Server",
      icon: "terminal",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "ai",
      sourceOrigin: "process",
    };
  }

  if (cmd.includes("uvicorn") || cmd.includes("fastapi")) {
    return {
      label: "FastAPI Service",
      icon: "python",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "backend",
      sourceOrigin: "process",
    };
  }

  if (cmd.includes("gunicorn") || cmd.includes("django")) {
    return {
      label: "Django App",
      icon: "python",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "backend",
      sourceOrigin: "process",
    };
  }

  if (cmd.includes("flask")) {
    return {
      label: "Flask App",
      icon: "python",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "backend",
      sourceOrigin: "process",
    };
  }

  if (pName.includes("postgres") || cmd.includes("postgres")) {
    return {
      label: "PostgreSQL Database",
      icon: "database",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "database",
      sourceOrigin: "process",
    };
  }

  if (pName.includes("mysql") || pName.includes("mariadb") || cmd.includes("mysqld")) {
    return {
      label: "MySQL Database",
      icon: "database",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "database",
      sourceOrigin: "process",
    };
  }

  if (pName.includes("redis") || cmd.includes("redis-server")) {
    return {
      label: "Redis Cache Server",
      icon: "database",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "database",
      sourceOrigin: "process",
    };
  }

  if (pName.includes("mongod") || cmd.includes("mongod")) {
    return {
      label: "MongoDB Database",
      icon: "database",
      defaultProtocol: isHttps ? "https" : "http",
      serviceType: "database",
      sourceOrigin: "process",
    };
  }

  // 4. Well-Known Port Numbers Fallback
  switch (port) {
    case 5173:
    case 5174:
    case 5175:
      return {
        label: `Vite App (Port ${port})`,
        icon: "vite",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "frontend",
        sourceOrigin: "well-known",
      };
    case 3000:
    case 3001:
      return {
        label: `Web App (Port ${port})`,
        icon: "react",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "frontend",
        sourceOrigin: "well-known",
      };
    case 8080:
      return {
        label: "HTTP Server (Port 8080)",
        icon: "terminal",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "backend",
        sourceOrigin: "well-known",
      };
    case 8000:
    case 5000:
      return {
        label: `API Server (Port ${port})`,
        icon: cmd.includes("python") ? "python" : "terminal",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "backend",
        sourceOrigin: "well-known",
      };
    case 3306:
      return {
        label: "MySQL (3306)",
        icon: "database",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "database",
        sourceOrigin: "well-known",
      };
    case 5432:
      return {
        label: "PostgreSQL (5432)",
        icon: "database",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "database",
        sourceOrigin: "well-known",
      };
    case 6379:
      return {
        label: "Redis (6379)",
        icon: "database",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "database",
        sourceOrigin: "well-known",
      };
    case 27017:
      return {
        label: "MongoDB (27017)",
        icon: "database",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "database",
        sourceOrigin: "well-known",
      };
    case 8888:
    case 8889:
      return {
        label: `Jupyter / Proxy (Port ${port})`,
        icon: "python",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "datascience",
        sourceOrigin: "well-known",
      };
    case 11434:
      return {
        label: "Ollama (11434)",
        icon: "terminal",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "ai",
        sourceOrigin: "well-known",
      };
    case 443:
    case 8443:
      return {
        label: `HTTPS Server (Port ${port})`,
        icon: "terminal",
        defaultProtocol: "https",
        serviceType: "web",
        sourceOrigin: "well-known",
      };
    case 22:
      return {
        label: "SSH Server (22)",
        icon: "terminal",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "system",
        sourceOrigin: "well-known",
      };
    default:
      if (cmd.includes("node")) {
        return {
          label: `Node Service (${port})`,
          icon: "node",
          defaultProtocol: isHttps ? "https" : "http",
          serviceType: "backend",
          sourceOrigin: "process",
        };
      }
      if (cmd.includes("python")) {
        return {
          label: `Python Service (${port})`,
          icon: "python",
          defaultProtocol: isHttps ? "https" : "http",
          serviceType: "backend",
          sourceOrigin: "process",
        };
      }
      if (cmd.includes("go") || pName.includes("go")) {
        return {
          label: `Go Service (${port})`,
          icon: "go",
          defaultProtocol: isHttps ? "https" : "http",
          serviceType: "backend",
          sourceOrigin: "process",
        };
      }
      return {
        label: `Service (${port})`,
        icon: "terminal",
        defaultProtocol: isHttps ? "https" : "http",
        serviceType: "custom",
        sourceOrigin: "well-known",
      };
  }
}

/**
 * Smart Port Security Risk Assessor
 * Evaluates whether a port hosts sensitive database/infrastructure management services
 * that require mandatory high-risk security interception before being exposed publicly.
 */
export function assessPortRisk(
  port: number,
  options?: {
    serviceType?: string;
    label?: string;
    process?: string;
  }
): PortRiskAssessment {
  const lbl = (options?.label || "").toLowerCase();
  const proc = (options?.process || "").toLowerCase();
  const sType = (options?.serviceType || "").toLowerCase();

  // High-Risk: Databases & System Admin Daemons
  if (
    port === 3306 || // MySQL / MariaDB
    port === 5432 || // PostgreSQL
    port === 6379 || // Redis
    port === 27017 || // MongoDB
    port === 1521 || // Oracle
    port === 1433 || // SQL Server
    port === 9200 || // Elasticsearch
    port === 9300 ||
    port === 9042 || // Cassandra
    port === 5984 || // CouchDB
    port === 8086 || // InfluxDB
    port === 2375 || // Docker Daemon (unencrypted)
    port === 2376 || // Docker Daemon (TLS)
    port === 22 || // SSH
    port === 2379 || // etcd
    sType === "database" ||
    lbl.includes("redis") ||
    lbl.includes("mysql") ||
    lbl.includes("postgres") ||
    lbl.includes("mongo") ||
    lbl.includes("database") ||
    proc.includes("redis") ||
    proc.includes("mysqld") ||
    proc.includes("postgres") ||
    proc.includes("mongod")
  ) {
    let name = "数据库或管理服务";
    if (port === 6379 || lbl.includes("redis") || proc.includes("redis")) name = "Redis 缓存数据库";
    else if (port === 3306 || lbl.includes("mysql") || proc.includes("mysql")) name = "MySQL / MariaDB 数据库";
    else if (port === 5432 || lbl.includes("postgres") || proc.includes("postgres")) name = "PostgreSQL 数据库";
    else if (port === 27017 || lbl.includes("mongo") || proc.includes("mongo")) name = "MongoDB 数据库";
    else if (port === 2375 || port === 2376 || proc.includes("dockerd")) name = "Docker Daemon 控制守护进程";
    else if (port === 22 || proc.includes("sshd")) name = "SSH 远程管理服务";

    return {
      level: "high",
      serviceCategory: name,
      reason: `该端口运行着敏感的 ${name}。将其公开给全网或外部局域网极易导致未授权读取、篡改或勒索攻击。`,
      recommendation: "强烈建议保持私有访问，通过安全的本地隧道连接使用。",
    };
  }

  // Standard Web / API Service
  return {
    level: "standard",
    serviceCategory: "Web 应用服务",
    reason: "公开端口后，任何拥有链接或局域网 IP 的人员均可直接访问此 Web 服务。",
    recommendation: "如仅需临时联调或演示，可在调试完成后一键重置为私有。",
  };
}
