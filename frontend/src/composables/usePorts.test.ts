// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, vi, beforeEach } from "vitest";
import { usePorts } from "./usePorts";

const store = new Map<string, string>();
const localStorageMock = {
  getItem: (key: string) => store.get(key) || null,
  setItem: (key: string, val: string) => store.set(key, val),
  removeItem: (key: string) => store.delete(key),
  clear: () => store.clear(),
};
vi.stubGlobal("localStorage", localStorageMock);

describe("usePorts composable", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    store.clear();
  });

  it("fetchPorts merges auto-detected ports correctly", async () => {
    const mockListening = [
      {
        port: 5173,
        protocol: "tcp",
        localAddress: "127.0.0.1:5173",
        process: "node",
        commandLine: "node /path/to/vite",
        pid: 1234,
      },
      {
        port: 8889,
        protocol: "tcp",
        localAddress: "127.0.0.1:8889",
        process: "python3",
        commandLine: "jupyter-lab",
        pid: 5678,
      },
    ];

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, listening: mockListening }),
    });

    const { ports, activePortsCount, fetchPorts } = usePorts();
    await fetchPorts();

    expect(ports.value).toHaveLength(2);
    expect(activePortsCount.value).toBe(2);

    const vitePort = ports.value.find((p) => p.port === 5173);
    expect(vitePort).toBeDefined();
    expect(vitePort?.label).toBe("Vite Dev Server");
    expect(vitePort?.source).toBe("auto");
    expect(vitePort?.status).toBe("active");
  });

  it("addPort, updatePort, stopPort, startPort and removePort manage lifecycle", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true }),
    });

    const { ports, listeningPorts, addPort, updatePort, stopPort, startPort, removePort } = usePorts();

    // Mark 3306 as listening on OS
    listeningPorts.value = [{ port: 3306, protocol: "tcp", localAddress: "127.0.0.1:3306", process: "mysqld" }];

    // 1. Add manual port 3306
    const added = await addPort({
      port: 3306,
      label: "MySQL Dev DB",
      protocol: "http",
      visibility: "private",
    });
    expect(added).toBe(true);

    const item = ports.value.find((p) => p.port === 3306);
    expect(item).toBeDefined();
    expect(item?.source).toBe("user");
    expect(item?.label).toBe("MySQL Dev DB");

    // 2. Update port 3306
    await updatePort(3306, { label: "Production DB Replica", visibility: "public" });
    const updated = ports.value.find((p) => p.port === 3306);
    expect(updated?.label).toBe("Production DB Replica");
    expect(updated?.visibility).toBe("public");

    // 3. Stop port
    await stopPort(3306);
    const stopped = ports.value.find((p) => p.port === 3306);
    expect(stopped?.status).toBe("inactive");

    // 4. Start port
    await startPort(3306);
    const started = ports.value.find((p) => p.port === 3306);
    expect(started?.status).toBe("active");

    // 5. Remove port
    await removePort(3306);
    const removed = ports.value.find((p) => p.port === 3306);
    expect(removed).toBeUndefined();
  });

  it("registerDetectedPort integrates terminal url sniffer", () => {
    const { ports, registerDetectedPort } = usePorts();
    registerDetectedPort(4000, { label: "Express API" });

    const p = ports.value.find((x) => x.port === 4000);
    expect(p).toBeDefined();
    expect(p?.status).toBe("active");
    expect(p?.forwardAddress).toBe("localhost:4000");
  });

  it("findAvailableLocalPort avoids port collision", async () => {
    const { addPort, findAvailableLocalPort } = usePorts();
    await addPort({ port: 8080, localPort: 8080 });

    const nextPort = findAvailableLocalPort(8080);
    expect(nextPort).toBe(8081);
  });

  it("fetchPorts adapts to HTTPS and detected framework banner from probe", async () => {
    const mockListening = [
      {
        port: 8443,
        protocol: "tcp",
        localAddress: "127.0.0.1:8443",
        process: "uvicorn",
        probe: {
          port: 8443,
          isHttps: true,
          httpStatus: 200,
          serverHeader: "uvicorn",
          detectedService: "FastAPI / Starlette (Uvicorn)",
          probeDurationMs: 5,
        },
      },
    ];

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, listening: mockListening }),
    });

    const { ports, fetchPorts } = usePorts();
    await fetchPorts();

    const uvicornPort = ports.value.find((p) => p.port === 8443);
    expect(uvicornPort).toBeDefined();
    expect(uvicornPort?.protocol).toBe("https");
    expect(uvicornPort?.label).toBe("FastAPI / Starlette (Uvicorn)");
    expect(uvicornPort?.previewUrl).toContain("https://localhost:8443/");
    expect(uvicornPort?.probe?.isHttps).toBe(true);
  });

  it("fetchPorts incorporates Method C workspace hints and pre-reserves unstarted expected ports", async () => {
    const mockListening = [
      {
        port: 5173,
        protocol: "tcp",
        localAddress: "127.0.0.1:5173",
        process: "node",
      },
    ];

    const mockWorkspaceHints = [
      {
        port: 5173,
        name: "UniRTM Portal Web",
        sourceFile: "package.json",
        serviceType: "npm-script",
        framework: "Vite",
      },
      {
        port: 5432,
        name: "PostgreSQL Database (compose: db)",
        sourceFile: "docker-compose.yml",
        serviceType: "compose",
        framework: "PostgreSQL",
      },
    ];

    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes("/api/ports/workspace-hints")) {
        return Promise.resolve({
          ok: true,
          json: async () => ({ success: true, hints: mockWorkspaceHints }),
        });
      }
      return Promise.resolve({
        ok: true,
        json: async () => ({ success: true, listening: mockListening }),
      });
    });

    const { ports, fetchPorts } = usePorts();
    await fetchPorts();

    // 1. Running port 5173 gets enriched with workspace business name
    const webPort = ports.value.find((p) => p.port === 5173);
    expect(webPort).toBeDefined();
    expect(webPort?.label).toBe("UniRTM Portal Web");
    expect(webPort?.status).toBe("active");
    expect(webPort?.workspaceHint?.framework).toBe("Vite");

    // 2. Expected but not yet started compose port 5432 is pre-reserved as inactive
    const dbPort = ports.value.find((p) => p.port === 5432);
    expect(dbPort).toBeDefined();
    expect(dbPort?.label).toBe("PostgreSQL Database (compose: db)");
    expect(dbPort?.status).toBe("inactive");
    expect(dbPort?.workspaceHint?.sourceFile).toBe("docker-compose.yml");
  });

  it("publicPortsCount tracks count of public ports and resetAllPortsToPrivate resets all to private", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true }),
    });

    const { ports, publicPortsCount, addPort, updatePort, resetAllPortsToPrivate } = usePorts();

    await addPort({ port: 3000, label: "Frontend", visibility: "private" });
    await addPort({ port: 8080, label: "Backend API", visibility: "private" });
    await addPort({ port: 3306, label: "Database", visibility: "private" });

    expect(publicPortsCount.value).toBe(0);

    // Make two ports public
    await updatePort(3000, { visibility: "public" });
    await updatePort(8080, { visibility: "public" });

    expect(publicPortsCount.value).toBe(2);

    // Call resetAllPortsToPrivate
    await resetAllPortsToPrivate();

    expect(publicPortsCount.value).toBe(0);
    const p3000 = ports.value.find((p) => p.port === 3000);
    const p8080 = ports.value.find((p) => p.port === 8080);
    const p3306 = ports.value.find((p) => p.port === 3306);
    expect(p3000?.visibility).toBe("private");
    expect(p8080?.visibility).toBe("private");
    expect(p3306?.visibility).toBe("private");
  });

  it("assessPortRisk correctly classifies high-risk vs standard web services", async () => {
    const { assessPortRisk } = await import("../utils/portServiceRecognizer");

    // High risk ports & processes
    const mysqlRisk = assessPortRisk(3306);
    expect(mysqlRisk.level).toBe("high");
    expect(mysqlRisk.serviceCategory).toContain("MySQL");

    const redisRisk = assessPortRisk(6379);
    expect(redisRisk.level).toBe("high");
    expect(redisRisk.serviceCategory).toContain("Redis");

    const mongoRisk = assessPortRisk(27017);
    expect(mongoRisk.level).toBe("high");
    expect(mongoRisk.serviceCategory).toContain("MongoDB");

    const dockerRisk = assessPortRisk(2375);
    expect(dockerRisk.level).toBe("high");
    expect(dockerRisk.serviceCategory).toContain("Docker");

    const sshRisk = assessPortRisk(22);
    expect(sshRisk.level).toBe("high");
    expect(sshRisk.serviceCategory).toContain("SSH");

    const customNamedDb = assessPortRisk(9999, { label: "My Postgres Node", serviceType: "database" });
    expect(customNamedDb.level).toBe("high");

    // Standard Web ports
    const viteRisk = assessPortRisk(5173, { label: "Vite App" });
    expect(viteRisk.level).toBe("standard");
    expect(viteRisk.serviceCategory).toBe("Web 应用服务");

    const nextRisk = assessPortRisk(3000, { label: "Next.js Web" });
    expect(nextRisk.level).toBe("standard");

    const apiRisk = assessPortRisk(8080, { label: "Spring Boot / Express API" });
    expect(apiRisk.level).toBe("standard");
  });
});
