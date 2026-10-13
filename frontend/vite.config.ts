import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { execFileSync } from "node:child_process";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    {
      name: "preserve-dist-gitkeep",
      closeBundle() {
        const gitkeepPath = fileURLToPath(new URL("./dist/.gitkeep", import.meta.url));
        fs.writeFileSync(gitkeepPath, "");
      },
    },
    {
      name: "dev-api-middleware",
      configureServer(server) {
        server.middlewares.use((req: any, res: any, next: any) => {
          const rootDir = fileURLToPath(new URL("..", import.meta.url));
          const unirtmBin = path.join(rootDir, "bin", "unirtm");

          if (req.url?.startsWith("/api/network/status")) {
            const urlObj = new URL(req.url, "http://localhost");
            const scope = urlObj.searchParams.get("scope");
            const args = ["network", "status", "--json"];
            if (scope === "user") args.push("--global");
            try {
              const out = execFileSync(unirtmBin, args, { cwd: rootDir, encoding: "utf-8" });
              res.setHeader("Content-Type", "application/json");
              res.end(out);
            } catch (err: any) {
              res.statusCode = 500;
              res.setHeader("Content-Type", "application/json");
              res.end(JSON.stringify({ error: err.message }));
            }
            return;
          }

          if (req.url === "/api/network/profile/use" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { profile_id, scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "profile", "use", profile_id];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, active_profile: profile_id }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/network/profile/save" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { profile_id, name, scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "profile", "save-as", profile_id];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, profile_id, name }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/network/profile/reset" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { profile_id, scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "profile", "reset", profile_id];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, reset: profile_id }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/network/profile/delete" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { profile_id, scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "profile", "delete", profile_id];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, deleted: profile_id }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/network/registry/override" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { ecosystem, url, scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "registry", "set", ecosystem, url];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, ecosystem, url }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/network/registry/add" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { ecosystem, url, scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "registry", "add", ecosystem, url];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, ecosystem, url }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/network/registry/delete" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { ecosystem, scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "registry", "delete", ecosystem];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, ecosystem }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/network/registry/reset" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { ecosystem, scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "registry", "reset", ecosystem];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, ecosystem }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/network/proxy" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const body = JSON.parse(bodyStr || "{}");
                const proxy = body.proxy || body;
                const scope = body.scope;
                if (!proxy || !proxy.enabled) {
                  const args = ["network", "proxy", "off"];
                  if (scope === "user") args.push("--global");
                  execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                } else {
                  const rawUrl = proxy.http || proxy.socks5 || "";
                  if (rawUrl) {
                    const args = ["network", "proxy", "set", rawUrl];
                    if (scope === "user") args.push("--global");
                    execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                  } else {
                    const args = ["network", "proxy", "on"];
                    if (scope === "user") args.push("--global");
                    execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                  }
                }
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true, proxy }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/network/ping")) {
            res.setHeader("Content-Type", "application/json");
            res.end(
              JSON.stringify({
                results: [
                  { ecosystem: "npm", name: "npmmirror", endpoint: "https://registry.npmmirror.com", latency_ms: 28, status: "fast", ok: true },
                  { ecosystem: "go", name: "goproxy.cn", endpoint: "https://goproxy.cn", latency_ms: 35, status: "fast", ok: true },
                  { ecosystem: "python", name: "清华大学 TUNA", endpoint: "https://pypi.tuna.tsinghua.edu.cn", latency_ms: 42, status: "fast", ok: true },
                  { ecosystem: "rust", name: "rsproxy.cn", endpoint: "https://rsproxy.cn", latency_ms: 30, status: "fast", ok: true },
                  { ecosystem: "ruby", name: "Ruby China", endpoint: "https://gems.ruby-china.com", latency_ms: 55, status: "fast", ok: true },
                ],
              })
            );
            return;
          }

          if (req.url === "/api/system/browsers") {
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ browsers: [] }));
            return;
          }
          if (req.url === "/api/config") {
            const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
            const wsPath = path.join(rootDir, ".unirtm.toml");

            const parseSettings = (raw: string) => {
              const res: Record<string, any> = {};
              let inSettings = false;
              for (const line of raw.split("\n")) {
                const tr = line.trim();
                if (tr.startsWith("[") && tr.endsWith("]")) {
                  inSettings = tr === "[settings]";
                  continue;
                }
                if (inSettings && tr && !tr.startsWith("#")) {
                  const eq = tr.indexOf("=");
                  if (eq !== -1) {
                    const k = tr.slice(0, eq).trim();
                    let v = tr.slice(eq + 1).trim();
                    if ((v.startsWith('"') && v.endsWith('"')) || (v.startsWith("'") && v.endsWith("'"))) {
                      v = v.slice(1, -1);
                    }
                    res[k] = v;
                  }
                }
              }
              return res;
            };

            if (req.method === "GET") {
              const userExists = fs.existsSync(userPath);
              const wsExists = fs.existsSync(wsPath);
              const userRaw = userExists ? fs.readFileSync(userPath, "utf-8") : "";
              const wsRaw = wsExists ? fs.readFileSync(wsPath, "utf-8") : "";

              res.setHeader("Content-Type", "application/json");
              res.end(
                JSON.stringify({
                  user: {
                    path: userPath,
                    exists: userExists,
                    raw: userRaw,
                    settings: parseSettings(userRaw),
                  },
                  workspace: {
                    path: wsPath,
                    exists: wsExists,
                    raw: wsRaw,
                    settings: parseSettings(wsRaw),
                  },
                  detected_browsers: [],
                })
              );
              return;
            } else if (req.method === "POST") {
              let bodyStr = "";
              req.on("data", (c: any) => {
                bodyStr += c;
              });
              req.on("end", () => {
                try {
                  const { scope, key, value } = JSON.parse(bodyStr || "{}");
                  const unirtmBin = path.join(rootDir, "bin", "unirtm");
                  const args = ["config", "set"];
                  if (scope === "user") {
                    args.push("--global");
                  }
                  args.push(key, String(value ?? ""));

                  if (fs.existsSync(unirtmBin)) {
                    execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                  } else {
                    // Fallback to direct directory and file creation
                    const targetFile = scope === "user" ? userPath : wsPath;
                    fs.mkdirSync(path.dirname(targetFile), { recursive: true });
                    const cleanKey = key.replace(/^settings\./, "");
                    const entry = `\n[settings]\n${cleanKey} = "${value}"\n`;
                    fs.appendFileSync(targetFile, entry);
                  }

                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ ok: true, scope, key, value }));
                } catch (err: any) {
                  res.statusCode = 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }
          }
          next();
        });
      },
    },
  ],
  root: fileURLToPath(new URL(".", import.meta.url)),
  build: {
    outDir: fileURLToPath(new URL("./dist", import.meta.url)),
    emptyOutDir: true,
  },
  server: {
    port: process.env.PORT ? parseInt(process.env.PORT, 10) : 5173,
    strictPort: false,
  },
});
