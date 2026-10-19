// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { recognizePortService } from "./portServiceRecognizer";

describe("portServiceRecognizer", () => {
  it("recognizes service by process and command line (Method A)", () => {
    const resVite = recognizePortService(5173, "node", "vite --port 5173");
    expect(resVite.label).toBe("Vite Dev Server");
    expect(resVite.icon).toBe("vite");
    expect(resVite.defaultProtocol).toBe("http");
    expect(resVite.sourceOrigin).toBe("process");

    const resPy = recognizePortService(8000, "python3", "uvicorn main:app --reload");
    expect(resPy.label).toBe("FastAPI Service");
    expect(resPy.icon).toBe("python");
    expect(resPy.sourceOrigin).toBe("process");
  });

  it("recognizes service and adapts to HTTPS by network probe (Method B)", () => {
    const resProbe = recognizePortService(
      3000,
      "docker-proxy", // opaque container proxy
      "/usr/bin/docker-proxy -proto tcp -host-ip 0.0.0.0 -host-port 3000",
      {
        port: 3000,
        isHttps: true,
        serverHeader: "uvicorn",
        detectedService: "FastAPI / Starlette (Uvicorn)",
      }
    );

    // Method B overrides opaque docker-proxy and detects actual service
    expect(resProbe.label).toBe("FastAPI / Starlette (Uvicorn)");
    expect(resProbe.defaultProtocol).toBe("https");
    expect(resProbe.icon).toBe("python");
    expect(resProbe.sourceOrigin).toBe("probe");
  });

  it("prioritizes workspace business context name (Method C)", () => {
    const resWorkspace = recognizePortService(
      5432,
      "postgres",
      "postgres -D /data",
      {
        port: 5432,
        isHttps: false,
        detectedService: "PostgreSQL Database",
      },
      {
        port: 5432,
        name: "主业务数据库 (Postgres)",
        sourceFile: "docker-compose.yml",
        serviceType: "compose",
        framework: "PostgreSQL",
      }
    );

    // Method C provides user's business semantic title
    expect(resWorkspace.label).toBe("主业务数据库 (Postgres)");
    expect(resWorkspace.icon).toBe("database");
    expect(resWorkspace.sourceOrigin).toBe("workspace");
  });
});
