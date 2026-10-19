// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { recognizePortService } from "./portServiceRecognizer";

describe("recognizePortService", () => {
  it("recognizes Vite dev server by command line", () => {
    const res = recognizePortService(5173, "node", "node /path/to/vite");
    expect(res.label).toBe("Vite Dev Server");
    expect(res.icon).toBe("vite");
    expect(res.defaultProtocol).toBe("http");
  });

  it("recognizes Next.js app by command line", () => {
    const res = recognizePortService(3000, "node", "next-server");
    expect(res.label).toBe("Next.js App");
    expect(res.icon).toBe("nextjs");
  });

  it("recognizes Jupyter Notebook", () => {
    const res = recognizePortService(8888, "python3", "jupyter-lab");
    expect(res.label).toBe("Jupyter Notebook");
    expect(res.icon).toBe("python");
  });

  it("recognizes databases by port and process name", () => {
    const pg = recognizePortService(5432, "postgres", "/usr/lib/postgresql/15/bin/postgres");
    expect(pg.label).toBe("PostgreSQL Database");
    expect(pg.icon).toBe("database");

    const my = recognizePortService(3306, "mysqld", "/usr/sbin/mysqld");
    expect(my.label).toBe("MySQL Database");
    expect(my.icon).toBe("database");

    const redis = recognizePortService(6379, "redis-server", "redis-server *:6379");
    expect(redis.label).toBe("Redis Cache Server");
    expect(redis.icon).toBe("database");
  });

  it("recognizes Ollama LLM server", () => {
    const ollama = recognizePortService(11434, "ollama", "ollama serve");
    expect(ollama.label).toBe("Ollama LLM Server");
    expect(ollama.serviceType).toBe("ai");
  });

  it("falls back gracefully for generic custom ports", () => {
    const res = recognizePortService(4242, "custom-daemon", "/bin/custom");
    expect(res.label).toBe("Service (4242)");
    expect(res.icon).toBe("terminal");
  });
});
