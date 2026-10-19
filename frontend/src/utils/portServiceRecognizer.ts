// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export interface ServiceRecognitionResult {
  label: string;
  icon: string;
  defaultProtocol: "http" | "https";
  serviceType: string;
}

export function recognizePortService(
  port: number,
  processName?: string,
  cmdline?: string
): ServiceRecognitionResult {
  const pName = (processName || "").toLowerCase();
  const cmd = (cmdline || "").toLowerCase();

  // 1. Check by command line or process signature first (High priority)
  if (cmd.includes("vite") || pName.includes("vite")) {
    return {
      label: "Vite Dev Server",
      icon: "vite",
      defaultProtocol: "http",
      serviceType: "frontend",
    };
  }

  if (cmd.includes("next") || cmd.includes("next-server")) {
    return {
      label: "Next.js App",
      icon: "nextjs",
      defaultProtocol: "http",
      serviceType: "frontend",
    };
  }

  if (cmd.includes("nuxt") || cmd.includes("nuxt-server")) {
    return {
      label: "Nuxt App",
      icon: "vue",
      defaultProtocol: "http",
      serviceType: "frontend",
    };
  }

  if (cmd.includes("jupyter") || pName.includes("jupyter")) {
    return {
      label: "Jupyter Notebook",
      icon: "python",
      defaultProtocol: "http",
      serviceType: "datascience",
    };
  }

  if (cmd.includes("ollama") || pName.includes("ollama")) {
    return {
      label: "Ollama LLM Server",
      icon: "terminal",
      defaultProtocol: "http",
      serviceType: "ai",
    };
  }

  if (cmd.includes("uvicorn") || cmd.includes("fastapi")) {
    return {
      label: "FastAPI Service",
      icon: "python",
      defaultProtocol: "http",
      serviceType: "backend",
    };
  }

  if (cmd.includes("gunicorn") || cmd.includes("django")) {
    return {
      label: "Django App",
      icon: "python",
      defaultProtocol: "http",
      serviceType: "backend",
    };
  }

  if (cmd.includes("flask")) {
    return {
      label: "Flask App",
      icon: "python",
      defaultProtocol: "http",
      serviceType: "backend",
    };
  }

  if (pName.includes("postgres") || cmd.includes("postgres")) {
    return {
      label: "PostgreSQL Database",
      icon: "database",
      defaultProtocol: "http",
      serviceType: "database",
    };
  }

  if (pName.includes("mysql") || pName.includes("mariadb") || cmd.includes("mysqld")) {
    return {
      label: "MySQL Database",
      icon: "database",
      defaultProtocol: "http",
      serviceType: "database",
    };
  }

  if (pName.includes("redis") || cmd.includes("redis-server")) {
    return {
      label: "Redis Cache Server",
      icon: "database",
      defaultProtocol: "http",
      serviceType: "database",
    };
  }

  if (pName.includes("mongod") || cmd.includes("mongod")) {
    return {
      label: "MongoDB Database",
      icon: "database",
      defaultProtocol: "http",
      serviceType: "database",
    };
  }

  // 2. Check by Well-Known Port Numbers
  switch (port) {
    case 5173:
    case 5174:
    case 5175:
      return {
        label: `Vite App (Port ${port})`,
        icon: "vite",
        defaultProtocol: "http",
        serviceType: "frontend",
      };
    case 3000:
    case 3001:
      return {
        label: `Web App (Port ${port})`,
        icon: "react",
        defaultProtocol: "http",
        serviceType: "frontend",
      };
    case 8080:
      return {
        label: "HTTP Server (Port 8080)",
        icon: "terminal",
        defaultProtocol: "http",
        serviceType: "backend",
      };
    case 8000:
    case 5000:
      return {
        label: `API Server (Port ${port})`,
        icon: cmd.includes("python") ? "python" : "terminal",
        defaultProtocol: "http",
        serviceType: "backend",
      };
    case 3306:
      return {
        label: "MySQL (3306)",
        icon: "database",
        defaultProtocol: "http",
        serviceType: "database",
      };
    case 5432:
      return {
        label: "PostgreSQL (5432)",
        icon: "database",
        defaultProtocol: "http",
        serviceType: "database",
      };
    case 6379:
      return {
        label: "Redis (6379)",
        icon: "database",
        defaultProtocol: "http",
        serviceType: "database",
      };
    case 27017:
      return {
        label: "MongoDB (27017)",
        icon: "database",
        defaultProtocol: "http",
        serviceType: "database",
      };
    case 8888:
    case 8889:
      return {
        label: `Jupyter / Proxy (Port ${port})`,
        icon: "python",
        defaultProtocol: "http",
        serviceType: "datascience",
      };
    case 11434:
      return {
        label: "Ollama (11434)",
        icon: "terminal",
        defaultProtocol: "http",
        serviceType: "ai",
      };
    case 443:
    case 8443:
      return {
        label: `HTTPS Server (Port ${port})`,
        icon: "terminal",
        defaultProtocol: "https",
        serviceType: "web",
      };
    case 22:
      return {
        label: "SSH Server (22)",
        icon: "terminal",
        defaultProtocol: "http",
        serviceType: "system",
      };
    default:
      if (cmd.includes("node")) {
        return {
          label: `Node Service (${port})`,
          icon: "node",
          defaultProtocol: "http",
          serviceType: "backend",
        };
      }
      if (cmd.includes("python")) {
        return {
          label: `Python Service (${port})`,
          icon: "python",
          defaultProtocol: "http",
          serviceType: "backend",
        };
      }
      if (cmd.includes("go") || pName.includes("go")) {
        return {
          label: `Go Service (${port})`,
          icon: "go",
          defaultProtocol: "http",
          serviceType: "backend",
        };
      }
      return {
        label: `Service (${port})`,
        icon: "terminal",
        defaultProtocol: "http",
        serviceType: "custom",
      };
  }
}
