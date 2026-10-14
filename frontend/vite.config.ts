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

          if (req.url === "/api/network/registry/reset-all" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const { scope } = JSON.parse(bodyStr || "{}");
                const args = ["network", "registry", "reset-all"];
                if (scope === "user") args.push("--global");
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ ok: true }));
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

          if (req.url === "/api/network/proxy/detect") {
            try {
              let detected = false;
              let port = 8889;
              let desc = "";
              const singBoxPath = "/etc/sing-box/config.json";
              if (fs.existsSync(singBoxPath)) {
                try {
                  const cfg = JSON.parse(fs.readFileSync(singBoxPath, "utf-8"));
                  for (const inb of cfg.inbounds || []) {
                    if (inb.type === "mixed" || inb.type === "http" || inb.type === "socks") {
                      port = inb.listen_port || 8889;
                      detected = true;
                      desc = `检测到本机正在运行的 sing-box 混合代理服务 (端口 ${port})`;
                      break;
                    }
                  }
                } catch {}
              }
              if (!detected) {
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ detected: false }));
                return;
              }
              res.setHeader("Content-Type", "application/json");
              res.end(
                JSON.stringify({
                  detected: true,
                  type: "sing-box",
                  http: `http://127.0.0.1:${port}`,
                  https: `http://127.0.0.1:${port}`,
                  socks5: `socks5://127.0.0.1:${port}`,
                  port,
                  description: desc,
                })
              );
            } catch (err: any) {
              res.statusCode = 500;
              res.setHeader("Content-Type", "application/json");
              res.end(JSON.stringify({ error: err.message }));
            }
            return;
          }

          if (req.url === "/api/network/proxy/test") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", async () => {
              try {
                let proxyCfg: any = {};
                if (bodyStr) {
                  try { proxyCfg = JSON.parse(bodyStr); } catch {}
                }
                const httpProxy = proxyCfg.http || proxyCfg.https;
                if (!httpProxy && !proxyCfg.socks5) {
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ success: false, error: "未填写任何代理服务器地址" }));
                  return;
                }
                const testURL = httpProxy || proxyCfg.socks5;
                const parsed = new URL(testURL);
                const testPort = parseInt(parsed.port, 10);
                const testHost = parsed.hostname || "127.0.0.1";

                const net = await import("node:net");
                const start = Date.now();
                const socket = net.createConnection({ host: testHost, port: testPort, timeout: 2500 }, () => {
                  const rtt = Date.now() - start;
                  socket.destroy();
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({
                    success: true,
                    rtt_ms: rtt,
                    message: `代理服务器连通正常 (响应延迟 ${rtt}ms)`
                  }));
                });
                socket.on("error", (err: any) => {
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({
                    success: false,
                    error: `代理服务器端口无法连通 (${testHost}:${testPort}): ${err.message}`
                  }));
                });
                socket.on("timeout", () => {
                  socket.destroy();
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({
                    success: false,
                    error: `连接代理服务器端口超时 (${testHost}:${testPort})`
                  }));
                });
              } catch (err: any) {
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: false, error: err.message }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/network/ping")) {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", async () => {
              try {
                let targets: Array<{ ecosystem: string; name?: string; url: string }> = [];
                if (bodyStr) {
                  try {
                    const parsed = JSON.parse(bodyStr);
                    if (Array.isArray(parsed.targets)) {
                      targets = parsed.targets;
                    } else if (parsed.url) {
                      targets = [{ ecosystem: parsed.ecosystem || "custom", name: parsed.name || parsed.ecosystem, url: parsed.url }];
                    }
                  } catch {}
                }

                if (targets.length === 0) {
                  targets = [
                    { ecosystem: "npm", name: "npmmirror", url: "https://registry.npmmirror.com" },
                    { ecosystem: "go", name: "goproxy.cn", url: "https://goproxy.cn" },
                    { ecosystem: "python", name: "清华大学 TUNA", url: "https://pypi.tuna.tsinghua.edu.cn/simple" },
                    { ecosystem: "rust", name: "rsproxy.cn", url: "https://rsproxy.cn/crates.io-index" },
                    { ecosystem: "ruby", name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/rubygems/" },
                    { ecosystem: "composer", name: "阿里云 Composer", url: "https://mirrors.aliyun.com/composer/" },
                    { ecosystem: "maven", name: "阿里云 Maven", url: "https://maven.aliyun.com/repository/public" },
                    { ecosystem: "nuget", name: "华为云 NuGet", url: "https://repo.huaweicloud.com/repository/nuget/v3/index.json" },
                    { ecosystem: "docker", name: "阿里云 Docker", url: "https://registry.cn-hangzhou.aliyuncs.com" },
                    { ecosystem: "dart", name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/dart-pub" },
                    { ecosystem: "cocoapods", name: "CocoaPods CDN", url: "https://cdn.cocoapods.org" },
                    { ecosystem: "conda", name: "清华大学 Conda", url: "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main" },
                    { ecosystem: "homebrew", name: "清华大学 Homebrew", url: "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles" },
                    { ecosystem: "git", name: "GitHub", url: "https://github.com" },
                  ];
                }

                const probeOne = async (t: { ecosystem: string; name?: string; url: string }) => {
                  const start = Date.now();
                  try {
                    let targetUrl = t.url;
                    if (targetUrl.includes(",")) {
                      const parts = targetUrl.split(",");
                      for (const p of parts) {
                        const trimmed = p.trim();
                        if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) {
                          targetUrl = trimmed;
                          break;
                        }
                      }
                    }
                    const controller = new AbortController();
                    const timer = setTimeout(() => controller.abort(), 3000);
                    const resp = await fetch(targetUrl, {
                      method: "HEAD",
                      signal: controller.signal,
                      headers: { "User-Agent": "UniRTM-Probe/0.1.0" },
                    });
                    clearTimeout(timer);
                    const rtt = Math.max(1, Date.now() - start);
                    let status = "normal";
                    if (resp.status >= 500) {
                      status = "error";
                    } else if (rtt < 100) {
                      status = "fast";
                    } else if (rtt < 300) {
                      status = "normal";
                    } else {
                      status = "slow";
                    }
                    return {
                      ecosystem: t.ecosystem,
                      name: t.name || t.ecosystem,
                      endpoint: t.url,
                      rtt_ms: rtt,
                      latency_ms: rtt,
                      status,
                      ok: resp.status < 500,
                    };
                  } catch (err: any) {
                    const rtt = Date.now() - start;
                    const isTimeout = err.name === "AbortError" || rtt >= 2800;
                    return {
                      ecosystem: t.ecosystem,
                      name: t.name || t.ecosystem,
                      endpoint: t.url,
                      rtt_ms: 0,
                      latency_ms: 0,
                      status: isTimeout ? "timeout" : "error",
                      error: err.message,
                      ok: false,
                    };
                  }
                };

                const results = await Promise.all(targets.map(probeOne));
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ results }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
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
