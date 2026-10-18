import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { execFileSync, spawnSync, exec, spawn } from "node:child_process";
import { Buffer } from "node:buffer";
import { WebSocketServer, WebSocket } from "ws";

function findUnirtmBinary(rootDir: string): string {
  const isWindows = process.platform === "win32";
  const exeSuffix = isWindows ? ".exe" : "";
  const candidates = [
    path.join(rootDir, "bin", `unirtm${exeSuffix}`),
    path.join(rootDir, `unirtm${exeSuffix}`),
    path.join(os.homedir(), ".local", "bin", `unirtm${exeSuffix}`),
    path.join(os.homedir(), "bin", `unirtm${exeSuffix}`),
  ];

  let bestBin = "";
  let latestMtime = -1;

  for (const p of candidates) {
    if (fs.existsSync(p)) {
      try {
        const stat = fs.statSync(p);
        if (stat.mtimeMs > latestMtime) {
          latestMtime = stat.mtimeMs;
          bestBin = p;
        }
      } catch {}
    }
  }

  return bestBin || candidates[0];
}

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
        const wss = new WebSocketServer({ noServer: true });

        server.httpServer?.on("upgrade", (req, socket, head) => {
          if (req.url?.startsWith("/api/terminal/ws")) {
            wss.handleUpgrade(req, socket, head, (ws) => {
              wss.emit("connection", ws, req);
            });
          }
        });

        wss.on("connection", (ws: WebSocket, req: any) => {
          const urlObj = new URL(req.url, "http://localhost");
          const shellPath = urlObj.searchParams.get("shell") || "/bin/bash";
          const cwd = urlObj.searchParams.get("cwd") || process.cwd();

          const home = os.homedir();
          const unirtmBin = path.join(home, ".unirtm", "bin");
          const unirtmShims = path.join(home, ".unirtm", "shims");
          const customPath = `${unirtmShims}:${unirtmBin}:${process.env.PATH || ""}`;

          const isWindows = process.platform === "win32";
          const cols = urlObj.searchParams.get("cols") || "120";
          const rows = urlObj.searchParams.get("rows") || "30";
          const bridgeScript = fileURLToPath(new URL("./scripts/pty_bridge.py", import.meta.url));

          let child: any;
          if (!isWindows && fs.existsSync(bridgeScript)) {
            child = spawn("python3", [bridgeScript, cols, rows, shellPath, cwd], {
              cwd: fs.existsSync(cwd) ? cwd : process.cwd(),
              env: {
                ...process.env,
                PATH: customPath,
                TERM: "xterm-256color",
                COLORTERM: "truecolor",
                UNIRTM_TERMINAL: "1",
                LANG: process.env.LANG || "en_US.UTF-8",
              },
              stdio: ["pipe", "pipe", "pipe"],
            });
          } else {
            const shellArgs = isWindows ? [] : ["-i"];
            child = spawn(shellPath, shellArgs, {
              cwd: fs.existsSync(cwd) ? cwd : process.cwd(),
              env: {
                ...process.env,
                PATH: customPath,
                TERM: "xterm-256color",
                COLORTERM: "truecolor",
                UNIRTM_TERMINAL: "1",
                LANG: process.env.LANG || "en_US.UTF-8",
              },
              stdio: ["pipe", "pipe", "pipe"],
            });
          }

          child.stdout?.on("data", (chunk: Buffer) => {
            if (ws.readyState === WebSocket.OPEN) {
              ws.send(JSON.stringify({ type: "output", data: chunk.toString("utf-8") }));
            }
          });

          child.stderr?.on("data", (chunk: Buffer) => {
            if (ws.readyState === WebSocket.OPEN) {
              ws.send(JSON.stringify({ type: "output", data: chunk.toString("utf-8") }));
            }
          });

          ws.on("message", (raw: any) => {
            try {
              const msg = JSON.parse(raw.toString());
              if (msg.type === "input" && typeof msg.data === "string") {
                child.stdin?.write(msg.data);
              } else if (msg.type === "resize") {
                if (!isWindows && fs.existsSync(bridgeScript)) {
                  child.stdin?.write(JSON.stringify(msg) + "\n");
                }
              } else if (msg.type === "kill") {
                child.kill();
              } else if (msg.type === "ping") {
                ws.send(JSON.stringify({ type: "pong" }));
              }
            } catch {
              child.stdin?.write(raw);
            }
          });

          const cleanup = () => {
            try {
              child.kill();
            } catch {}
          };

          ws.on("close", cleanup);
          ws.on("error", cleanup);
          child.on("close", () => {
            if (ws.readyState === WebSocket.OPEN) {
              ws.close();
            }
          });
        });

        const remoteVersionCache = new Map<string, { versions: string[]; timestamp: number }>();
        const queryRemoteVersions = (tool: string, rootDir: string, unirtmBin: string): string[] => {
          if (!tool || !fs.existsSync(unirtmBin)) return [];
          const cached = remoteVersionCache.get(tool);
          if (cached && Date.now() - cached.timestamp < 10 * 60 * 1000) {
            return cached.versions;
          }
          try {
            const res = spawnSync(unirtmBin, ["ls-remote", tool, "--json"], {
              cwd: rootDir,
              encoding: "utf-8",
              maxBuffer: 50 * 1024 * 1024,
              timeout: 3000,
            });
            const text = (res.stdout || "").trim() || (res.stderr || "").trim();
            const start = text.indexOf("{");
            const end = text.lastIndexOf("}");
            if (start !== -1 && end !== -1 && end > start) {
              const data = JSON.parse(text.slice(start, end + 1));
              const rawList = data.fields?.versions || data.versions || (Array.isArray(data) ? data : []);
              if (Array.isArray(rawList) && rawList.length > 0) {
                const extracted = rawList
                  .map((v: any) => v.Version || v.version || (typeof v === "string" ? v : ""))
                  .filter((v: string) => v && !v.includes("rc") && !v.includes("beta") && !v.includes("alpha"));
                const clean = Array.from(new Set(extracted));
                if (clean.length > 0) {
                  remoteVersionCache.set(tool, { versions: clean, timestamp: Date.now() });
                  return clean;
                }
              }
            }
          } catch (e) {
            console.warn(`[UniRTM DevServer] Remote lookup error for ${tool}:`, e);
          }
          return [];
        };

        server.middlewares.use((req: any, res: any, next: any) => {
          const rootDir = fileURLToPath(new URL("..", import.meta.url));
          const unirtmBin = findUnirtmBinary(rootDir);

          const parseTools = (filePath: string) => {
            const res: Record<string, string> = {};
            if (!fs.existsSync(filePath)) return res;
            try {
              const c = fs.readFileSync(filePath, "utf-8");
              let inTools = false;
              for (const line of c.split("\n")) {
                const tr = line.trim();
                if (tr.startsWith("[") && tr.endsWith("]")) {
                  inTools = tr === "[tools]";
                  continue;
                }
                if (inTools && tr && !tr.startsWith("#")) {
                  const eqIdx = tr.indexOf("=");
                  if (eqIdx !== -1) {
                    const rawKey = tr.slice(0, eqIdx).trim();
                    const rawVal = tr.slice(eqIdx + 1).trim();
                    const key = rawKey.replace(/^["']|["']$/g, "").trim();
                    const val = rawVal.replace(/^["']|["']$/g, "").trim();
                    if (key && val) {
                      res[key] = val;
                    }
                  }
                }
              }
            } catch {}
            return res;
          };

          const parsePins = (filePath: string) => {
            const res: Record<string, { pinned: boolean; reason: string }> = {};
            if (!fs.existsSync(filePath)) return res;
            try {
              const c = fs.readFileSync(filePath, "utf-8");
              let inPins = false;
              for (const line of c.split("\n")) {
                const tr = line.trim();
                if (tr.startsWith("[") && tr.endsWith("]")) {
                  inPins = tr === "[pins]";
                  continue;
                }
                if (inPins && tr && !tr.startsWith("#")) {
                  const eqIdx = tr.indexOf("=");
                  if (eqIdx !== -1) {
                    const rawKey = tr.slice(0, eqIdx).trim();
                    const rawVal = tr.slice(eqIdx + 1).trim();
                    const key = rawKey.replace(/^["']|["']$/g, "").trim();
                    const val = rawVal.replace(/^["']|["']$/g, "").trim();
                    if (key) {
                      if (val === "false") {
                        res[key] = { pinned: false, reason: "" };
                      } else if (val === "true") {
                        res[key] = { pinned: true, reason: "" };
                      } else {
                        res[key] = { pinned: true, reason: val };
                      }
                    }
                  }
                }
              }
            } catch {}
            return res;
          };

          const parseLockTools = (lockPath: string): Record<string, string[]> => {
            const res: Record<string, string[]> = {};
            if (!fs.existsSync(lockPath)) return res;
            try {
              const c = fs.readFileSync(lockPath, "utf-8");
              let currentTool = "";
              for (const line of c.split("\n")) {
                const tr = line.trim();
                const m = tr.match(/^\[\[tools\."([^"]+)"\]\]/);
                if (m) {
                  currentTool = m[1];
                  if (!res[currentTool]) res[currentTool] = [];
                  continue;
                }
                if (currentTool && tr.startsWith("version =")) {
                  const vm = tr.match(/version\s*=\s*"([^"]+)"/);
                  if (vm && vm[1]) {
                    if (!res[currentTool].includes(vm[1])) {
                      res[currentTool].push(vm[1]);
                    }
                  }
                }
              }
            } catch {}
            return res;
          };

          if (req.url === "/api/status" || req.url?.startsWith("/api/status?")) {
            let version = "0.35.0";
            let gitTag = "v0.35.0";
            let commit = "dev";
            let buildTime = new Date().toISOString().slice(0, 10);

            try {
              const versionFile = path.join(rootDir, "VERSION");
              if (fs.existsSync(versionFile)) {
                const lines = fs.readFileSync(versionFile, "utf-8").split("\n");
                for (const l of lines) {
                  const tr = l.trim();
                  if (tr && !tr.startsWith("#")) {
                    version = tr.replace(/^v/, "");
                    gitTag = `v${version}`;
                    break;
                  }
                }
              }
            } catch {}

            try {
              const gitRev = execFileSync("git", ["rev-parse", "--short", "HEAD"], {
                cwd: rootDir,
                encoding: "utf-8",
                timeout: 2000,
              }).trim();
              if (gitRev) commit = gitRev;
            } catch {}

            try {
              const gitDate = execFileSync("git", ["log", "-1", "--format=%cs"], {
                cwd: rootDir,
                encoding: "utf-8",
                timeout: 2000,
              }).trim();
              if (gitDate) buildTime = gitDate;
            } catch {}

            res.writeHead(200, { "Content-Type": "application/json" });
            res.end(
              JSON.stringify({
                status: "ready",
                version: gitTag,
                git_tag: gitTag,
                commit: commit,
                build_time: buildTime,
                platform: `${process.platform}/${process.arch}`,
                go_version: "go1.27+",
                project_name: "unirtm",
                author: "Snowdream Tech <snowdreamtech@qq.com>",
                license: "MIT License",
                copyright: "Copyright (c) 2026 SnowdreamTech. All rights reserved.",
                tools_count: 5,
                tasks_count: 4,
                plugins_count: 0,
                timestamp: Math.floor(Date.now() / 1000),
              })
            );
            return;
          }

          if (
            (req.url === "/api/tools/migrate-legacy" || req.url === "/api/runtimes/migrate-legacy") &&
            req.method === "POST"
          ) {
            try {
              const wsPath = path.join(rootDir, ".unirtm.toml");
              let content = fs.existsSync(wsPath) ? fs.readFileSync(wsPath, "utf-8") : "";
              const nvmrc = path.join(rootDir, ".nvmrc");
              const nodeVer = path.join(rootDir, ".node-version");
              const goVer = path.join(rootDir, ".go-version");
              const pyVer = path.join(rootDir, ".python-version");
              let count = 0;

              const upsert = (raw: string, tool: string, ver: string) => {
                const lines = raw.split("\n");
                let toolsIdx = -1;
                let nextIdx = -1;
                for (let i = 0; i < lines.length; i++) {
                  const tr = lines[i].trim();
                  if (tr === "[tools]") toolsIdx = i;
                  else if (toolsIdx !== -1 && tr.startsWith("[") && tr.endsWith("]")) {
                    nextIdx = i;
                    break;
                  }
                }
                const reg = new RegExp(`^\\s*${tool}\\s*=`);
                if (toolsIdx !== -1) {
                  const end = nextIdx !== -1 ? nextIdx : lines.length;
                  let found = false;
                  for (let i = toolsIdx + 1; i < end; i++) {
                    if (reg.test(lines[i])) {
                      lines[i] = `${tool} = "${ver}"`;
                      found = true;
                      break;
                    }
                  }
                  if (!found) lines.splice(toolsIdx + 1, 0, `${tool} = "${ver}"`);
                  return lines.join("\n");
                }
                return raw + `\n[tools]\n${tool} = "${ver}"\n`;
              };

              if (fs.existsSync(nvmrc)) {
                content = upsert(content, "node", fs.readFileSync(nvmrc, "utf-8").trim().replace(/^v/, ""));
                count++;
              } else if (fs.existsSync(nodeVer)) {
                content = upsert(content, "node", fs.readFileSync(nodeVer, "utf-8").trim().replace(/^v/, ""));
                count++;
              }
              if (fs.existsSync(goVer)) {
                content = upsert(content, "go", fs.readFileSync(goVer, "utf-8").trim().replace(/^v/, ""));
                count++;
              }
              if (fs.existsSync(pyVer)) {
                content = upsert(content, "python", fs.readFileSync(pyVer, "utf-8").trim().replace(/^v/, ""));
                count++;
              }
              if (count === 0) {
                content = upsert(content, "node", "20.18.3");
                content = upsert(content, "go", "1.27.2");
                count = 2;
              }
              fs.writeFileSync(wsPath, content, "utf-8");
              res.setHeader("Content-Type", "application/json");
              res.end(JSON.stringify({ success: true, migrated_count: count, file: wsPath }));
            } catch (err: any) {
              res.statusCode = 500;
              res.setHeader("Content-Type", "application/json");
              res.end(JSON.stringify({ error: err.message }));
            }
            return;
          }

          if (req.url === "/api/tools/status" || req.url === "/api/runtimes/status") {
            const wsPath = path.join(rootDir, ".unirtm.toml");
            const altWsPath = path.join(rootDir, "unirtm.toml");
            const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
            const altUserPath = path.join(os.homedir(), ".config", "unirtm", "config.toml");

            const installedMap: Record<string, string[]> = {};
            if (fs.existsSync(unirtmBin)) {
              try {
                const out = execFileSync(unirtmBin, ["list", "--json"], {
                  cwd: rootDir,
                  encoding: "utf-8",
                  timeout: 5000,
                });
                const data = JSON.parse(out);
                const list = data.fields?.tools || data.tools || [];
                for (const item of list) {
                  const v = item.version;
                  if (!v) continue;
                  const add = (k: string) => {
                    if (!installedMap[k]) installedMap[k] = [];
                    if (!installedMap[k].includes(v)) installedMap[k].push(v);
                  };
                  add(item.tool);
                  if (item.backend === "go-pkg" || item.backend === "go") {
                    add(`go:${item.tool}`);
                  } else if (item.backend && item.backend !== "native") {
                    add(`${item.backend}:${item.tool}`);
                  }
                }
              } catch (e) {
                console.warn("[Vite Middleware] unirtm list fallback:", e);
              }
            }

            const getInstalled = (toolId: string, fallback: string[]) => {
              if (installedMap[toolId] && installedMap[toolId].length > 0) {
                return installedMap[toolId];
              }
              const bareTool = toolId.includes(":") ? toolId.split(":").slice(1).join(":") : toolId;
              if (installedMap[bareTool] && installedMap[bareTool].length > 0) {
                return installedMap[bareTool];
              }

              const installsRoot = path.join(os.homedir(), ".local", "share", "unirtm", "installs");
              if (!fs.existsSync(installsRoot)) return fallback;

              const candDirs = [
                toolId,
                bareTool,
                toolId.replace(/[:/@]/g, "-"),
                toolId.replace(/^go:/, "go-pkg-").replace(/[:/@]/g, "-"),
                toolId.replace(/^npm:/, "npm-").replace(/[:/@]/g, "-"),
                toolId.replace(/^github:/, "github-").replace(/[:/@]/g, "-"),
                toolId.replace(/^pipx:/, "pipx-").replace(/[:/@]/g, "-"),
              ];

              for (const cand of candDirs) {
                const dir = path.join(installsRoot, cand);
                if (fs.existsSync(dir)) {
                  try {
                    const list = fs
                      .readdirSync(dir)
                      .filter((f) => !f.startsWith(".") && fs.statSync(path.join(dir, f)).isDirectory());
                    if (list.length > 0) {
                      return list.sort((a, b) => b.localeCompare(a, undefined, { numeric: true }));
                    }
                  } catch {}
                }
              }
              return fallback;
            };

            const wsTools = parseTools(fs.existsSync(wsPath) ? wsPath : altWsPath);
            const userTools = parseTools(fs.existsSync(userPath) ? userPath : altUserPath);
            const wsPins = parsePins(fs.existsSync(wsPath) ? wsPath : altWsPath);
            const userPins = parsePins(fs.existsSync(userPath) ? userPath : altUserPath);

            const nvmrc = path.join(rootDir, ".nvmrc");
            const nodeVer = path.join(rootDir, ".node-version");
            const goVer = path.join(rootDir, ".go-version");
            const pyVer = path.join(rootDir, ".python-version");

            const detected_legacies: string[] = [];
            if (fs.existsSync(nvmrc)) detected_legacies.push(".nvmrc");
            if (fs.existsSync(nodeVer)) detected_legacies.push(".node-version");
            if (fs.existsSync(goVer)) detected_legacies.push(".go-version");
            if (fs.existsSync(pyVer)) detected_legacies.push(".python-version");

            const getToolItem = (
              id: string,
              name: string,
              desc: string,
              hp: string,
              defaultVer: string,
              fallbackInstalled: string[]
            ) => {
              const installed = getInstalled(id, fallbackInstalled);
              let active = wsTools[id] || "";
              let source = active ? ".unirtm.toml" : "";
              let isLegacy = false;
              let legacyFile = "";

              if (!active) {
                if (id === "node" && fs.existsSync(nvmrc)) {
                  active = fs.readFileSync(nvmrc, "utf-8").trim().replace(/^v/, "");
                  source = ".nvmrc";
                  isLegacy = true;
                  legacyFile = ".nvmrc";
                } else if (id === "node" && fs.existsSync(nodeVer)) {
                  active = fs.readFileSync(nodeVer, "utf-8").trim().replace(/^v/, "");
                  source = ".node-version";
                  isLegacy = true;
                  legacyFile = ".node-version";
                } else if (id === "go" && fs.existsSync(goVer)) {
                  active = fs.readFileSync(goVer, "utf-8").trim().replace(/^v/, "");
                  source = ".go-version";
                  isLegacy = true;
                  legacyFile = ".go-version";
                } else if (id === "python" && fs.existsSync(pyVer)) {
                  active = fs.readFileSync(pyVer, "utf-8").trim().replace(/^v/, "");
                  source = ".python-version";
                  isLegacy = true;
                  legacyFile = ".python-version";
                }
              }

              const globalVer = userTools[id] || "";
              const pinInfo = wsPins[id] || userPins[id] || { pinned: false, reason: "" };

              return {
                id,
                name,
                description: desc,
                homepage: hp,
                installed_versions: installed,
                active_version: active,
                global_version: globalVer,
                source: source || (active ? ".unirtm.toml" : ""),
                is_legacy: isLegacy,
                legacy_file: legacyFile,
                pinned: pinInfo.pinned,
                pin_reason: pinInfo.reason,
              };
            };

            const knownLanguageRuntimes: Record<
              string,
              { name: string; desc: string; hp: string; defaultVer: string }
            > = {
              node: {
                name: "Node.js",
                desc: "JavaScript / TypeScript runtime (NVM superset)",
                hp: "https://nodejs.org",
                defaultVer: "20.18.3",
              },
              go: {
                name: "Go",
                desc: "Go programming language (GVM superset)",
                hp: "https://go.dev",
                defaultVer: "1.27.2",
              },
              python: {
                name: "Python",
                desc: "Python language interpreter (Pyenv superset)",
                hp: "https://python.org",
                defaultVer: "3.14.7",
              },
              rust: {
                name: "Rust",
                desc: "Rust toolchain & cargo build system",
                hp: "https://rust-lang.org",
                defaultVer: "1.85.0",
              },
              java: {
                name: "Java",
                desc: "OpenJDK / Java runtime environment",
                hp: "https://openjdk.org",
                defaultVer: "21.0.2",
              },
              ruby: {
                name: "Ruby",
                desc: "Dynamic open source programming language with a focus on simplicity",
                hp: "https://ruby-lang.org",
                defaultVer: "3.4.1",
              },
              bun: {
                name: "Bun",
                desc: "Incredibly fast all-in-one JavaScript & TypeScript toolkit",
                hp: "https://bun.sh",
                defaultVer: "1.2.4",
              },
              deno: {
                name: "Deno",
                desc: "Next-generation JavaScript and TypeScript runtime",
                hp: "https://deno.com",
                defaultVer: "2.1.9",
              },
              zig: {
                name: "Zig",
                desc: "General-purpose programming language and toolchain",
                hp: "https://ziglang.org",
                defaultVer: "0.13.0",
              },
              php: {
                name: "PHP",
                desc: "Popular general-purpose scripting language suited for web development",
                hp: "https://php.net",
                defaultVer: "8.4.3",
              },
              dotnet: {
                name: "Dotnet",
                desc: ".NET Developer Platform and SDK",
                hp: "https://dotnet.microsoft.com",
                defaultVer: "9.0.100",
              },
              elixir: {
                name: "Elixir",
                desc: "Dynamic functional language designed for building scalable applications",
                hp: "https://elixir-lang.org",
                defaultVer: "1.18.1",
              },
              erlang: {
                name: "Erlang",
                desc: "Programming language used to build massively scalable soft real-time systems",
                hp: "https://www.erlang.org",
                defaultVer: "27.2.1",
              },
              flutter: {
                name: "Flutter",
                desc: "UI software development kit by Google",
                hp: "https://flutter.dev",
                defaultVer: "3.27.3",
              },
              lua: {
                name: "Lua",
                desc: "Powerful, efficient, lightweight, embeddable scripting language",
                hp: "https://www.lua.org",
                defaultVer: "5.4.7",
              },
              swift: {
                name: "Swift",
                desc: "General-purpose programming language developed by Apple",
                hp: "https://swift.org",
                defaultVer: "6.0.3",
              },
            };

            // 1. Collect all tool IDs that are actually configured (workspace/user) or active via legacy files
            const activeLegacyIds: string[] = [];
            if (fs.existsSync(nvmrc) || fs.existsSync(nodeVer)) activeLegacyIds.push("node");
            if (fs.existsSync(goVer)) activeLegacyIds.push("go");
            if (fs.existsSync(pyVer)) activeLegacyIds.push("python");

            const configuredIds = Array.from(
              new Set([...Object.keys(wsTools), ...Object.keys(userTools), ...activeLegacyIds])
            );

            // Priority order for standard language runtimes if they are configured
            const priorityLangs = ["node", "go", "python", "rust", "bun", "deno", "java", "ruby"];
            const sortedToolIds = configuredIds.sort((a, b) => {
              const aIdx = priorityLangs.indexOf(a);
              const bIdx = priorityLangs.indexOf(b);
              if (aIdx !== -1 && bIdx !== -1) return aIdx - bIdx;
              if (aIdx !== -1) return -1;
              if (bIdx !== -1) return 1;
              return a.localeCompare(b);
            });

            const toolsList: any[] = [];
            for (const id of sortedToolIds) {
              const meta = knownLanguageRuntimes[id] || {
                name: id,
                desc: "Unified toolchain managed via UniRTM",
                hp: "",
                defaultVer: "",
              };
              toolsList.push(getToolItem(id, meta.name, meta.desc, meta.hp, meta.defaultVer, []));
            }

            const totalInstalled = toolsList.reduce((acc, r) => acc + r.installed_versions.length, 0);

            res.setHeader("Content-Type", "application/json");
            res.end(
              JSON.stringify({
                tools: toolsList,
                runtimes: toolsList,
                total_installed: totalInstalled,
                detected_legacies,
              })
            );
            return;
          }

          if ((req.url === "/api/tools/use" || req.url === "/api/runtimes/use") && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const { tool, version, global: isGlobal } = JSON.parse(bodyStr || "{}");
                let executedCli = false;
                if (fs.existsSync(unirtmBin)) {
                  try {
                    const args = ["use", "--no-install", `${tool}@${version}`];
                    if (isGlobal) args.push("--global");
                    execFileSync(unirtmBin, args, { cwd: rootDir, encoding: "utf-8" });
                    executedCli = true;
                  } catch (e) {
                    console.warn("unirtm CLI execution fallback:", e);
                  }
                }

                // Fallback direct update only if CLI was skipped or failed
                if (!executedCli) {
                  const targetPath = isGlobal
                    ? path.join(os.homedir(), ".config", "unirtm", "unirtm.toml")
                    : path.join(rootDir, ".unirtm.toml");
                  fs.mkdirSync(path.dirname(targetPath), { recursive: true });
                  let content = fs.existsSync(targetPath) ? fs.readFileSync(targetPath, "utf-8") : "";

                  const lines = content.split("\n");
                  let toolsIdx = -1;
                  let nextIdx = -1;
                  for (let i = 0; i < lines.length; i++) {
                    const tr = lines[i].trim();
                    if (tr === "[tools]") toolsIdx = i;
                    else if (toolsIdx !== -1 && tr.startsWith("[") && tr.endsWith("]")) {
                      nextIdx = i;
                      break;
                    }
                  }
                  const reg = new RegExp(`^\\s*"?${tool}"?\\s*=`);
                  if (toolsIdx !== -1) {
                    const end = nextIdx !== -1 ? nextIdx : lines.length;
                    let found = false;
                    for (let i = toolsIdx + 1; i < end; i++) {
                      if (reg.test(lines[i])) {
                        lines[i] = `${tool} = "${version}"`;
                        found = true;
                        break;
                      }
                    }
                    if (!found) lines.splice(toolsIdx + 1, 0, `${tool} = "${version}"`);
                    content = lines.join("\n");
                  } else {
                    content += `\n[tools]\n${tool} = "${version}"\n`;
                  }

                  fs.writeFileSync(targetPath, content, "utf-8");

                  // Also synchronize unirtm.lock in fallback if in workspace mode
                  if (!isGlobal) {
                    const lockPath = path.join(rootDir, "unirtm.lock");
                    if (fs.existsSync(lockPath)) {
                      let lockContent = fs.readFileSync(lockPath, "utf-8");
                      const toolBlockHeader = `[[tools."${tool}"]]`;
                      if (lockContent.includes(toolBlockHeader)) {
                        const toolRegex = new RegExp(
                          `(\\[\\[tools\\."${tool}"\\]\\][\\s\\S]*?version\\s*=\\s*)"[^"]+"`
                        );
                        lockContent = lockContent.replace(toolRegex, `$1"${version}"`);
                      } else {
                        lockContent += `\n[[tools."${tool}"]]\nversion = "${version}"\nbackend = "native"\n`;
                      }
                      fs.writeFileSync(lockPath, lockContent, "utf-8");
                    }
                  }
                }
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: true, tool, version, global: isGlobal, cli: executedCli }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if ((req.url === "/api/tools/remove" || req.url === "/api/runtimes/remove") && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const { tool, global: isGlobal } = JSON.parse(bodyStr || "{}");
                let executedCli = false;
                if (fs.existsSync(unirtmBin)) {
                  try {
                    const args = ["unuse", tool];
                    if (isGlobal) args.push("--global");
                    execFileSync(unirtmBin, args, { cwd: rootDir, encoding: "utf-8" });
                    executedCli = true;
                  } catch (e) {
                    console.warn("unirtm unuse fallback:", e);
                  }
                }

                const targetPath = isGlobal
                  ? path.join(os.homedir(), ".config", "unirtm", "unirtm.toml")
                  : path.join(rootDir, ".unirtm.toml");

                if (fs.existsSync(targetPath)) {
                  let content = fs.readFileSync(targetPath, "utf-8");
                  const lines = content.split("\n");
                  let toolsIdx = -1;
                  let nextIdx = -1;
                  for (let i = 0; i < lines.length; i++) {
                    const tr = lines[i].trim();
                    if (tr === "[tools]") toolsIdx = i;
                    else if (toolsIdx !== -1 && tr.startsWith("[") && tr.endsWith("]")) {
                      nextIdx = i;
                      break;
                    }
                  }
                  if (toolsIdx !== -1) {
                    const end = nextIdx !== -1 ? nextIdx : lines.length;
                    const reg = new RegExp(`^\\s*"?${tool}"?\\s*=`);
                    for (let i = toolsIdx + 1; i < end; i++) {
                      if (reg.test(lines[i])) {
                        lines.splice(i, 1);
                        break;
                      }
                    }
                    fs.writeFileSync(targetPath, lines.join("\n"), "utf-8");
                  }
                }

                // If fallback and not global, remove from unirtm.lock as well
                if (!executedCli && !isGlobal) {
                  const lockPath = path.join(rootDir, "unirtm.lock");
                  if (fs.existsSync(lockPath)) {
                    let lockContent = fs.readFileSync(lockPath, "utf-8");
                    const blockRegex = new RegExp(`\n?\\[\\[tools\\."${tool}"\\]\\][\\s\\S]*?(?=(\n\\[\\[|\$))`, "g");
                    lockContent = lockContent.replace(blockRegex, "");
                    fs.writeFileSync(lockPath, lockContent, "utf-8");
                  }
                }

                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: true, tool, cli: executedCli }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if ((req.url === "/api/tools/pin" || req.url === "/api/runtimes/pin") && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const { tool, pinned, reason, global: isGlobal } = JSON.parse(bodyStr || "{}");
                const targetPath = isGlobal
                  ? path.join(os.homedir(), ".config", "unirtm", "unirtm.toml")
                  : path.join(rootDir, ".unirtm.toml");

                fs.mkdirSync(path.dirname(targetPath), { recursive: true });
                let content = fs.existsSync(targetPath) ? fs.readFileSync(targetPath, "utf-8") : "";

                const lines = content.split("\n");
                let pinsIdx = -1;
                let nextIdx = -1;
                for (let i = 0; i < lines.length; i++) {
                  const tr = lines[i].trim();
                  if (tr === "[pins]") pinsIdx = i;
                  else if (pinsIdx !== -1 && tr.startsWith("[") && tr.endsWith("]")) {
                    nextIdx = i;
                    break;
                  }
                }

                const quoteKey = (k: string) => (/^[a-zA-Z0-9_-]+$/.test(k) ? k : `"${k}"`);
                const keyPattern = new RegExp(`^\\s*"?${tool}"?\\s*=`);
                const valStr = reason ? `"${reason}"` : "true";

                if (pinned) {
                  if (pinsIdx !== -1) {
                    const end = nextIdx !== -1 ? nextIdx : lines.length;
                    let found = false;
                    for (let i = pinsIdx + 1; i < end; i++) {
                      if (keyPattern.test(lines[i])) {
                        lines[i] = `${quoteKey(tool)} = ${valStr}`;
                        found = true;
                        break;
                      }
                    }
                    if (!found) lines.splice(pinsIdx + 1, 0, `${quoteKey(tool)} = ${valStr}`);
                    content = lines.join("\n");
                  } else {
                    content += `\n[pins]\n${quoteKey(tool)} = ${valStr}\n`;
                  }
                } else {
                  if (pinsIdx !== -1) {
                    const end = nextIdx !== -1 ? nextIdx : lines.length;
                    for (let i = pinsIdx + 1; i < end; i++) {
                      if (keyPattern.test(lines[i])) {
                        lines.splice(i, 1);
                        break;
                      }
                    }
                    content = lines.join("\n");
                  }
                }

                fs.writeFileSync(targetPath, content, "utf-8");
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: true, tool, pinned, reason, global: isGlobal, file: targetPath }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/tools/check-updates") || req.url?.startsWith("/api/runtimes/check-updates")) {
            const wsPath = path.join(rootDir, ".unirtm.toml");
            const altWsPath = path.join(rootDir, "unirtm.toml");
            const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
            const altUserPath = path.join(os.homedir(), ".config", "unirtm", "config.toml");

            const wsTools = parseTools(fs.existsSync(wsPath) ? wsPath : altWsPath);
            const userTools = parseTools(fs.existsSync(userPath) ? userPath : altUserPath);
            const wsPins = parsePins(fs.existsSync(wsPath) ? wsPath : altWsPath);
            const userPins = parsePins(fs.existsSync(userPath) ? userPath : altUserPath);

            const allTools = Array.from(
              new Set(["node", "go", "python", "rust", ...Object.keys(wsTools), ...Object.keys(userTools)])
            );
            const urlObj = new URL(req.url, "http://localhost");
            const filterTool = urlObj.searchParams.get("tool");
            const filterScope = urlObj.searchParams.get("scope");
            const isGlobal = urlObj.searchParams.get("global") === "true" || filterScope === "user";
            const isWorkspace =
              filterScope === "workspace" || (urlObj.searchParams.get("global") === "false" && !filterScope);

            let targetTools: string[];
            if (filterTool) {
              targetTools = [filterTool];
            } else if (isGlobal) {
              targetTools = Array.from(new Set(Object.keys(userTools)));
            } else if (isWorkspace) {
              targetTools = Array.from(new Set(Object.keys(wsTools)));
            } else {
              targetTools = allTools;
            }

            const fallbackLatest: Record<string, string> = {
              node: "23.9.0",
              go: "1.27.2",
              python: "3.13.2",
              rust: "1.85.0",
              java: "23.0.2",
              bun: "1.2.4",
              deno: "2.2.3",
              ruby: "3.4.2",
              zig: "0.14.0",
              php: "8.4.4",
              trivy: "0.76.0",
              "aquasecurity/trivy": "0.76.0",
              "github:aquasecurity/trivy": "0.76.0",
              checkmake: "0.2.2",
              "checkmake/checkmake": "0.2.2",
              "github:checkmake/checkmake": "0.2.2",
              ruff: "0.16.10",
              "astral-sh/ruff": "0.16.10",
              "github:astral-sh/ruff": "0.16.10",
            };

            const semverGt = (a: string, b: string) => {
              const pa = a
                .replace(/^[vV]+/, "")
                .split(".")
                .map((n) => parseInt(n, 10));
              const pb = b
                .replace(/^[vV]+/, "")
                .split(".")
                .map((n) => parseInt(n, 10));
              for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
                const na = isNaN(pa[i]) ? 0 : pa[i];
                const nb = isNaN(pb[i]) ? 0 : pb[i];
                if (na !== nb) return na > nb;
              }
              return false;
            };

            const lockPath = path.join(rootDir, "unirtm.lock");
            const lockTools = parseLockTools(lockPath);

            const getLatestVersion = (toolKey: string, curVer: string): string => {
              const candidates: string[] = [];
              const bare = toolKey.includes(":") ? toolKey.split(":").slice(1).join(":") : toolKey;

              if (fallbackLatest[toolKey]) {
                candidates.push(fallbackLatest[toolKey]);
              }
              if (fallbackLatest[bare] && !candidates.includes(fallbackLatest[bare])) {
                candidates.push(fallbackLatest[bare]);
              }

              // Dynamic live remote check via unirtm ls-remote (cached)
              if ((filterTool && filterTool === toolKey) || remoteVersionCache.has(toolKey)) {
                const remotes = queryRemoteVersions(toolKey, rootDir, unirtmBin);
                for (const v of remotes) {
                  if (!candidates.includes(v)) candidates.push(v);
                }
              }

              const candKeys = [toolKey, bare, `github:${toolKey}`, `github:${bare}`];
              for (const k of candKeys) {
                if (lockTools[k]) {
                  for (const v of lockTools[k]) {
                    if (!candidates.includes(v)) candidates.push(v);
                  }
                }
              }

              // Scan locally installed versions from ~/.local/share/unirtm/installs
              const installsRoot = path.join(os.homedir(), ".local", "share", "unirtm", "installs");
              if (fs.existsSync(installsRoot)) {
                const dirCands = [
                  toolKey,
                  bare,
                  toolKey.replace(/[:/@]/g, "-"),
                  toolKey.replace(/^github:/, "github-").replace(/[:/@]/g, "-"),
                  toolKey.replace(/^npm:/, "npm-").replace(/[:/@]/g, "-"),
                  toolKey.replace(/^go:/, "go-pkg-").replace(/[:/@]/g, "-"),
                  toolKey.replace(/^pipx:/, "pipx-").replace(/[:/@]/g, "-"),
                ];
                for (const d of dirCands) {
                  const dir = path.join(installsRoot, d);
                  if (fs.existsSync(dir)) {
                    try {
                      const list = fs
                        .readdirSync(dir)
                        .filter((f: string) => !f.startsWith(".") && fs.statSync(path.join(dir, f)).isDirectory());
                      for (const v of list) {
                        if (!candidates.includes(v)) candidates.push(v);
                      }
                    } catch {}
                  }
                }
              }

              if (candidates.length === 0) {
                return curVer;
              }

              candidates.sort((a: string, b: string) => (semverGt(a, b) ? -1 : semverGt(b, a) ? 1 : 0));
              return candidates[0] || curVer;
            };

            const updates: any[] = [];
            let upgradableCount = 0;
            let pinnedCount = 0;

            for (const t of targetTools) {
              const cur = isGlobal
                ? userTools[t] || ""
                : isWorkspace
                  ? wsTools[t] || ""
                  : wsTools[t] || userTools[t] || "";
              const pin = isGlobal
                ? userPins[t] || { pinned: false, reason: "" }
                : isWorkspace
                  ? wsPins[t] || { pinned: false, reason: "" }
                  : wsPins[t] || userPins[t] || { pinned: false, reason: "" };
              const latest = getLatestVersion(t, cur);
              const cleanCur = cur.replace(/^[vV]+/, "").trim();
              const cleanLatest = latest.replace(/^[vV]+/, "").trim();
              const hasUpdate = Boolean(cur && latest && cleanCur !== cleanLatest && semverGt(latest, cur));

              if (pin.pinned) {
                pinnedCount++;
              } else if (hasUpdate) {
                upgradableCount++;
              }

              updates.push({
                tool: t,
                current_version: cur,
                latest_version: hasUpdate ? latest : cur || latest || "",
                has_update: hasUpdate,
                is_pinned: pin.pinned,
                pin_reason: pin.reason,
                in_workspace: Boolean(wsTools[t]),
                in_user: Boolean(userTools[t]),
                scope: userTools[t] && !wsTools[t] ? "user" : wsTools[t] && !userTools[t] ? "workspace" : "both",
              });
            }

            res.setHeader("Content-Type", "application/json");
            res.end(
              JSON.stringify({
                updates,
                total_checked: updates.length,
                upgradable_count: upgradableCount,
                pinned_count: pinnedCount,
              })
            );
            return;
          }

          if (
            (req.url === "/api/tools/batch-update" || req.url === "/api/runtimes/batch-update") &&
            req.method === "POST"
          ) {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const { tools, global: isGlobal } = JSON.parse(bodyStr || "{}");
                const targetPath = isGlobal
                  ? path.join(os.homedir(), ".config", "unirtm", "unirtm.toml")
                  : path.join(rootDir, ".unirtm.toml");

                const wsPins = parsePins(path.join(rootDir, ".unirtm.toml"));
                const userPins = parsePins(path.join(os.homedir(), ".config", "unirtm", "unirtm.toml"));
                const activePins = isGlobal ? userPins : wsPins;

                fs.mkdirSync(path.dirname(targetPath), { recursive: true });
                let content = fs.existsSync(targetPath) ? fs.readFileSync(targetPath, "utf-8") : "";

                const updated: string[] = [];
                const skipped: string[] = [];

                if (Array.isArray(tools)) {
                  for (const item of tools) {
                    const tool = item.tool;
                    const ver = item.version;
                    if (!tool || !ver) continue;

                    if (wsPins[tool]?.pinned || userPins[tool]?.pinned) {
                      skipped.push(tool);
                      continue;
                    }

                    const lines = content.split("\n");
                    let toolsIdx = -1;
                    let nextIdx = -1;
                    for (let i = 0; i < lines.length; i++) {
                      const tr = lines[i].trim();
                      if (tr === "[tools]") toolsIdx = i;
                      else if (toolsIdx !== -1 && tr.startsWith("[") && tr.endsWith("]")) {
                        nextIdx = i;
                        break;
                      }
                    }

                    const reg = new RegExp(`^\\s*"?${tool}"?\\s*=`);
                    const quoteKey = (k: string) => (/^[a-zA-Z0-9_-]+$/.test(k) ? k : `"${k}"`);

                    if (toolsIdx !== -1) {
                      const end = nextIdx !== -1 ? nextIdx : lines.length;
                      let found = false;
                      for (let i = toolsIdx + 1; i < end; i++) {
                        if (reg.test(lines[i])) {
                          lines[i] = `${quoteKey(tool)} = "${ver}"`;
                          found = true;
                          break;
                        }
                      }
                      if (!found) lines.splice(toolsIdx + 1, 0, `${quoteKey(tool)} = "${ver}"`);
                      content = lines.join("\n");
                    } else {
                      content += `\n[tools]\n${quoteKey(tool)} = "${ver}"\n`;
                    }
                    updated.push(tool);
                  }
                }

                fs.writeFileSync(targetPath, content, "utf-8");
                res.setHeader("Content-Type", "application/json");
                res.end(
                  JSON.stringify({
                    success: true,
                    updated,
                    skipped,
                    updated_count: updated.length,
                    skipped_count: skipped.length,
                    file: targetPath,
                  })
                );
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/tools/remote") || req.url?.startsWith("/api/runtimes/remote")) {
            const urlObj = new URL(req.url, "http://localhost");
            const tool = urlObj.searchParams.get("tool");
            let versions: string[] = [];

            // 1. Dynamic live query via UniRTM CLI: unirtm ls-remote <tool> --json
            if (tool && fs.existsSync(unirtmBin)) {
              versions = queryRemoteVersions(tool, rootDir, unirtmBin);
            }

            // 2. Static fallback only if CLI or upstream network fails
            if (versions.length === 0) {
              if (tool === "node" || tool === "nodejs") {
                versions = [
                  "23.9.0",
                  "23.8.0",
                  "23.7.0",
                  "23.6.0",
                  "23.0.0",
                  "22.14.0",
                  "22.13.1",
                  "22.13.0",
                  "22.12.0",
                  "22.11.0",
                  "22.9.0",
                  "22.0.0",
                  "21.7.3",
                  "21.6.2",
                  "21.1.0",
                  "20.18.3",
                  "20.18.2",
                  "20.18.0",
                  "20.17.0",
                  "20.16.0",
                  "20.15.1",
                  "20.14.0",
                  "20.12.2",
                  "20.11.1",
                  "20.10.0",
                  "20.9.0",
                  "19.9.0",
                  "19.8.1",
                  "18.20.7",
                  "18.20.6",
                  "18.20.5",
                  "18.20.4",
                  "18.20.0",
                  "18.19.1",
                  "18.18.2",
                  "18.17.1",
                  "18.16.0",
                  "18.14.2",
                  "18.12.1",
                  "16.20.2",
                  "16.20.0",
                  "16.19.1",
                  "16.18.0",
                  "16.14.2",
                  "14.21.3",
                  "14.20.1",
                  "14.19.3",
                ];
              } else if (tool === "python" || tool === "python3") {
                versions = [
                  "3.14.0a4",
                  "3.13.2",
                  "3.13.1",
                  "3.13.0",
                  "3.12.9",
                  "3.12.8",
                  "3.12.7",
                  "3.12.6",
                  "3.12.5",
                  "3.12.4",
                  "3.12.3",
                  "3.12.2",
                  "3.12.1",
                  "3.12.0",
                  "3.11.11",
                  "3.11.10",
                  "3.11.9",
                  "3.11.8",
                  "3.11.7",
                  "3.11.6",
                  "3.11.5",
                  "3.11.4",
                  "3.11.0",
                  "3.10.16",
                  "3.10.15",
                  "3.10.14",
                  "3.10.13",
                  "3.10.12",
                  "3.10.11",
                  "3.10.8",
                  "3.10.0",
                  "3.9.21",
                  "3.9.20",
                  "3.9.19",
                  "3.9.18",
                  "3.9.16",
                  "3.9.13",
                  "3.9.0",
                  "3.8.20",
                  "3.8.18",
                  "3.8.16",
                  "3.8.12",
                  "3.8.0",
                  "3.7.17",
                  "3.7.12",
                ];
              } else if (tool === "go" || tool === "golang") {
                versions = [
                  "1.27.2",
                  "1.27.1",
                  "1.27.0",
                  "1.26.5",
                  "1.26.4",
                  "1.26.3",
                  "1.26.2",
                  "1.26.1",
                  "1.26.0",
                  "1.25.8",
                  "1.25.7",
                  "1.25.6",
                  "1.25.5",
                  "1.25.4",
                  "1.25.3",
                  "1.25.2",
                  "1.25.1",
                  "1.25.0",
                  "1.24.5",
                  "1.24.4",
                  "1.24.3",
                  "1.24.2",
                  "1.24.1",
                  "1.24.0",
                  "1.23.6",
                  "1.23.5",
                  "1.23.4",
                  "1.23.3",
                  "1.23.2",
                  "1.23.1",
                  "1.23.0",
                  "1.22.12",
                  "1.22.11",
                  "1.22.10",
                  "1.22.9",
                  "1.22.8",
                  "1.22.6",
                  "1.22.5",
                  "1.22.2",
                  "1.22.0",
                  "1.21.13",
                  "1.21.12",
                  "1.21.11",
                  "1.21.8",
                  "1.21.6",
                  "1.21.0",
                  "1.20.14",
                  "1.20.12",
                  "1.20.7",
                  "1.20.0",
                  "1.19.13",
                  "1.19.8",
                  "1.19.0",
                  "1.18.10",
                  "1.18.5",
                ];
              } else if (tool === "rust") {
                versions = [
                  "1.85.0",
                  "1.84.1",
                  "1.84.0",
                  "1.83.0",
                  "1.82.0",
                  "1.81.0",
                  "1.80.1",
                  "1.80.0",
                  "1.79.0",
                  "1.78.0",
                  "1.77.2",
                  "1.77.0",
                  "1.76.0",
                  "1.75.0",
                  "1.74.1",
                  "1.73.0",
                  "1.72.1",
                  "1.71.1",
                  "1.70.0",
                  "1.69.0",
                  "1.68.2",
                  "1.67.1",
                  "1.66.1",
                  "1.65.0",
                ];
              } else if (tool === "java" || tool === "openjdk") {
                versions = [
                  "23.0.2",
                  "23.0.1",
                  "22.0.2",
                  "22.0.1",
                  "21.0.6",
                  "21.0.5",
                  "21.0.4",
                  "21.0.2",
                  "21.0.1",
                  "21.0.0",
                  "17.0.14",
                  "17.0.12",
                  "17.0.10",
                  "17.0.9",
                  "17.0.8",
                  "17.0.2",
                  "11.0.26",
                  "11.0.22",
                  "11.0.20",
                  "11.0.18",
                  "8.0.442",
                  "8.0.402",
                  "8.0.392",
                  "8.0.352",
                ];
              } else if (tool === "bun") {
                versions = [
                  "1.2.4",
                  "1.2.3",
                  "1.2.2",
                  "1.2.1",
                  "1.2.0",
                  "1.1.45",
                  "1.1.42",
                  "1.1.38",
                  "1.1.34",
                  "1.1.30",
                  "1.1.25",
                  "1.1.20",
                  "1.1.10",
                  "1.1.0",
                  "1.0.36",
                  "1.0.35",
                  "1.0.30",
                  "1.0.25",
                  "1.0.20",
                  "1.0.18",
                  "1.0.0",
                ];
              } else if (tool === "deno") {
                versions = [
                  "2.2.3",
                  "2.2.2",
                  "2.2.1",
                  "2.2.0",
                  "2.1.9",
                  "2.1.4",
                  "2.1.0",
                  "2.0.6",
                  "2.0.4",
                  "2.0.0",
                  "1.46.3",
                  "1.45.5",
                  "1.44.4",
                  "1.43.6",
                  "1.42.4",
                  "1.41.2",
                  "1.40.5",
                  "1.39.4",
                  "1.38.2",
                ];
              } else if (tool === "zig") {
                versions = ["0.14.0", "0.13.0", "0.12.1", "0.12.0", "0.11.0", "0.10.1", "0.9.1", "0.8.1"];
              } else if (tool === "ruby") {
                versions = [
                  "3.4.2",
                  "3.4.1",
                  "3.3.7",
                  "3.3.6",
                  "3.3.5",
                  "3.3.0",
                  "3.2.7",
                  "3.2.6",
                  "3.2.3",
                  "3.2.0",
                  "3.1.6",
                  "3.1.4",
                  "3.1.2",
                  "3.0.7",
                  "3.0.6",
                ];
              } else if (tool === "php") {
                versions = [
                  "8.4.4",
                  "8.4.3",
                  "8.4.1",
                  "8.3.17",
                  "8.3.16",
                  "8.3.14",
                  "8.3.10",
                  "8.3.0",
                  "8.2.27",
                  "8.2.26",
                  "8.2.20",
                  "8.2.15",
                  "8.1.31",
                  "8.1.28",
                  "8.1.20",
                  "8.0.30",
                  "8.0.28",
                ];
              } else {
                const lockPath = path.join(rootDir, "unirtm.lock");
                const lockTools = parseLockTools(lockPath);
                const bare = tool && tool.includes(":") ? tool.split(":").slice(1).join(":") : tool;
                const candKeys = [tool, bare, `github:${tool}`, `github:${bare}`];
                for (const k of candKeys) {
                  if (k && lockTools[k] && lockTools[k].length > 0) {
                    versions = [...lockTools[k]];
                    break;
                  }
                }
                if (versions.length === 0 && tool) {
                  const installsRoot = path.join(os.homedir(), ".local", "share", "unirtm", "installs");
                  if (fs.existsSync(installsRoot)) {
                    const candDirs = [
                      tool,
                      bare,
                      tool.replace(/[:/@]/g, "-"),
                      tool.replace(/^github:/, "github-").replace(/[:/@]/g, "-"),
                      tool.replace(/^npm:/, "npm-").replace(/[:/@]/g, "-"),
                      tool.replace(/^go:/, "go-pkg-").replace(/[:/@]/g, "-"),
                      tool.replace(/^pipx:/, "pipx-").replace(/[:/@]/g, "-"),
                    ];
                    for (const cand of candDirs) {
                      const dir = path.join(installsRoot, cand);
                      if (fs.existsSync(dir)) {
                        try {
                          const list = fs
                            .readdirSync(dir)
                            .filter((f) => !f.startsWith(".") && fs.statSync(path.join(dir, f)).isDirectory());
                          if (list.length > 0) {
                            versions = list.sort((a, b) => b.localeCompare(a, undefined, { numeric: true }));
                            break;
                          }
                        } catch {}
                      }
                    }
                  }
                }
              }
            }
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ tool, versions: versions.map((v) => ({ version: v })) }));
            return;
          }

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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const body = JSON.parse(bodyStr || "{}");
                const proxy = body.proxy || body;
                const scope = body.scope;
                const args = ["network", "proxy", "set"];
                if (scope === "user") args.push("--global");

                if (proxy.http !== undefined) {
                  args.push("--http", String(proxy.http));
                }
                if (proxy.https !== undefined) {
                  args.push("--https", String(proxy.https));
                }
                if (proxy.socks5 !== undefined) {
                  args.push("--socks5", String(proxy.socks5));
                }
                if (proxy.no_proxy !== undefined) {
                  args.push("--no-proxy", String(proxy.no_proxy));
                }
                if (!proxy || !proxy.enabled) {
                  args.push("--off");
                }
                execFileSync(unirtmBin, args, { cwd: rootDir, stdio: "pipe" });
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", async () => {
              try {
                let proxyCfg: any = {};
                if (bodyStr) {
                  try {
                    proxyCfg = JSON.parse(bodyStr);
                  } catch {}
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
                  res.end(
                    JSON.stringify({
                      success: true,
                      rtt_ms: rtt,
                      message: "代理服务器连通正常",
                    })
                  );
                });
                socket.on("error", (err: any) => {
                  res.setHeader("Content-Type", "application/json");
                  res.end(
                    JSON.stringify({
                      success: false,
                      error: `代理服务器端口无法连通 (${testHost}:${testPort}): ${err.message}`,
                    })
                  );
                });
                socket.on("timeout", () => {
                  socket.destroy();
                  res.setHeader("Content-Type", "application/json");
                  res.end(
                    JSON.stringify({
                      success: false,
                      error: `连接代理服务器端口超时 (${testHost}:${testPort})`,
                    })
                  );
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
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", async () => {
              try {
                let targets: Array<{ ecosystem: string; name?: string; url: string }> = [];
                if (bodyStr) {
                  try {
                    const parsed = JSON.parse(bodyStr);
                    if (Array.isArray(parsed.targets)) {
                      targets = parsed.targets;
                    } else if (parsed.url) {
                      targets = [
                        {
                          ecosystem: parsed.ecosystem || "custom",
                          name: parsed.name || parsed.ecosystem,
                          url: parsed.url,
                        },
                      ];
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
                    {
                      ecosystem: "nuget",
                      name: "华为云 NuGet",
                      url: "https://repo.huaweicloud.com/repository/nuget/v3/index.json",
                    },
                    { ecosystem: "docker", name: "阿里云 Docker", url: "https://registry.cn-hangzhou.aliyuncs.com" },
                    { ecosystem: "dart", name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/dart-pub" },
                    { ecosystem: "cocoapods", name: "CocoaPods CDN", url: "https://cdn.cocoapods.org" },
                    {
                      ecosystem: "conda",
                      name: "清华大学 Conda",
                      url: "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main",
                    },
                    {
                      ecosystem: "homebrew",
                      name: "清华大学 Homebrew",
                      url: "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles",
                    },
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

          // --- Themes & Marketplace API for Dev Server ---
          if (req.url?.startsWith("/api/themes/marketplace/search")) {
            const urlObj = new URL(req.url, "http://localhost");
            const q = urlObj.searchParams.get("q") || "";
            const size = urlObj.searchParams.get("size") || "24";
            const offset = urlObj.searchParams.get("offset") || "0";
            const targetURL = `https://open-vsx.org/api/-/search?category=Themes&query=${encodeURIComponent(q)}&size=${size}&offset=${offset}`;

            fetch(targetURL, {
              headers: { "User-Agent": "UniRTM-Workbench-Dev/0.1.0" },
            })
              .then(async (resp) => {
                if (!resp.ok) {
                  res.statusCode = resp.status;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: `Open VSX search error: HTTP ${resp.status}` }));
                  return;
                }
                const data = await resp.json();
                const themesDir = path.join(os.homedir(), ".config", "unirtm", "themes");
                const installedPkgIds = new Set<string>();
                if (fs.existsSync(themesDir)) {
                  try {
                    const entries = fs.readdirSync(themesDir, { withFileTypes: true });
                    for (const entry of entries) {
                      if (entry.isDirectory()) {
                        installedPkgIds.add(entry.name.toLowerCase());
                        const pkgJsonPath = path.join(themesDir, entry.name, "package.json");
                        if (fs.existsSync(pkgJsonPath)) {
                          try {
                            const p = JSON.parse(fs.readFileSync(pkgJsonPath, "utf-8"));
                            if (p.name) installedPkgIds.add(p.name.toLowerCase());
                          } catch {}
                        }
                      }
                    }
                  } catch {}
                }

                const extensions = (data.extensions || []).map((ext: any) => {
                  const pkgId = ext.name || "";
                  const nsPkgId = `${ext.namespace}.${ext.name}`.toLowerCase();
                  const isInstalled = installedPkgIds.has(pkgId.toLowerCase()) || installedPkgIds.has(nsPkgId);
                  return {
                    namespace: ext.namespace,
                    name: ext.name,
                    displayName: ext.displayName || ext.name,
                    version: ext.version,
                    description: ext.description || "",
                    downloadCount: ext.downloadCount || 0,
                    averageRating: ext.averageRating || 0,
                    files: {
                      download: ext.files?.download || "",
                      icon: ext.files?.icon || "",
                    },
                    url: ext.url || "",
                    is_installed: isInstalled,
                    package_id: ext.name,
                  };
                });

                res.setHeader("Content-Type", "application/json");
                res.end(
                  JSON.stringify({
                    offset: data.offset || 0,
                    totalSize: data.totalSize || 0,
                    extensions,
                  })
                );
              })
              .catch((err) => {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              });
            return;
          }

          if (req.url?.startsWith("/api/themes/marketplace/detail")) {
            const urlObj = new URL(req.url, "http://localhost");
            const namespace = urlObj.searchParams.get("namespace");
            const name = urlObj.searchParams.get("name");
            if (!namespace || !name) {
              res.statusCode = 400;
              res.setHeader("Content-Type", "application/json");
              res.end(JSON.stringify({ error: "namespace and name are required" }));
              return;
            }

            const targetURL = `https://open-vsx.org/api/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`;
            fetch(targetURL, {
              headers: { "User-Agent": "UniRTM-Workbench-Dev/0.1.0" },
            })
              .then(async (resp) => {
                if (!resp.ok) {
                  res.statusCode = resp.status;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: `Open VSX detail error: HTTP ${resp.status}` }));
                  return;
                }
                const data = await resp.json();

                let readme = "";
                if (data.files?.readme) {
                  try {
                    const rRes = await fetch(data.files.readme, {
                      headers: { "User-Agent": "UniRTM-Workbench-Dev/0.1.0" },
                    });
                    if (rRes.ok) readme = await rRes.text();
                  } catch {}
                }

                const contributes: { themes: any[]; iconThemes: any[]; productIconThemes: any[] } = {
                  themes: [],
                  iconThemes: [],
                  productIconThemes: [],
                };
                if (data.files?.manifest) {
                  try {
                    const mRes = await fetch(data.files.manifest, {
                      headers: { "User-Agent": "UniRTM-Workbench-Dev/0.1.0" },
                    });
                    if (mRes.ok) {
                      const pkg = await mRes.json();
                      if (pkg.contributes?.themes) {
                        contributes.themes = pkg.contributes.themes.map((t: any) => ({
                          id: t.id || "",
                          label: t.label || t.id || "Unnamed Theme",
                          uiTheme: t.uiTheme || "vs-dark",
                          path: t.path || "",
                        }));
                      }
                      if (pkg.contributes?.iconThemes) {
                        contributes.iconThemes = pkg.contributes.iconThemes.map((it: any) => ({
                          id: it.id || "",
                          label: it.label || it.id || "Unnamed Icon Theme",
                          path: it.path || "",
                        }));
                      }
                      if (pkg.contributes?.productIconThemes) {
                        contributes.productIconThemes = pkg.contributes.productIconThemes.map((pit: any) => ({
                          id: pit.id || "",
                          label: pit.label || pit.id || "Unnamed Product Icon Theme",
                          path: pit.path || "",
                        }));
                      }
                    }
                  } catch {}
                }

                const themesDir = path.join(os.homedir(), ".config", "unirtm", "themes");
                const installedPkgIds = new Set<string>();
                if (fs.existsSync(themesDir)) {
                  try {
                    const entries = fs.readdirSync(themesDir, { withFileTypes: true });
                    for (const entry of entries) {
                      if (entry.isDirectory()) {
                        installedPkgIds.add(entry.name.toLowerCase());
                        const pkgJsonPath = path.join(themesDir, entry.name, "package.json");
                        if (fs.existsSync(pkgJsonPath)) {
                          try {
                            const p = JSON.parse(fs.readFileSync(pkgJsonPath, "utf-8"));
                            if (p.name) installedPkgIds.add(p.name.toLowerCase());
                          } catch {}
                        }
                      }
                    }
                  } catch {}
                }

                const pkgId = data.name || "";
                const nsPkgId = `${data.namespace}.${data.name}`.toLowerCase();
                const isInstalled = installedPkgIds.has(pkgId.toLowerCase()) || installedPkgIds.has(nsPkgId);

                res.setHeader("Content-Type", "application/json");
                res.end(
                  JSON.stringify({
                    namespace: data.namespace,
                    name: data.name,
                    displayName: data.displayName || data.name,
                    version: data.version,
                    description: data.description || "",
                    downloadCount: data.downloadCount || 0,
                    averageRating: data.averageRating || 0,
                    reviewCount: data.reviewCount || 0,
                    repository: data.repository || "",
                    homepage: data.homepage || "",
                    bugs: data.bugs || "",
                    license: data.license || "",
                    timestamp: data.timestamp || "",
                    tags: data.tags || [],
                    categories: data.categories || [],
                    files: data.files || {},
                    url: data.url || "",
                    is_installed: isInstalled,
                    package_id: data.name,
                    readme,
                    contributes,
                  })
                );
              })
              .catch((err) => {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              });
            return;
          }

          if (req.url === "/api/themes/marketplace/install" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", async () => {
              try {
                const { download_url } = JSON.parse(bodyStr || "{}");
                if (!download_url) {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "missing download_url" }));
                  return;
                }
                const resp = await fetch(download_url);
                if (!resp.ok) {
                  res.statusCode = 502;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: `Failed to download VSIX: HTTP ${resp.status}` }));
                  return;
                }
                const buffer = Buffer.from(await resp.arrayBuffer());
                const tmpZip = path.join(os.tmpdir(), `unirtm_theme_${Date.now()}.vsix`);
                fs.writeFileSync(tmpZip, buffer);

                const script = `
import zipfile, json, os, sys
zf = zipfile.ZipFile(sys.argv[1], 'r')
pkg_raw = zf.read('extension/package.json').decode('utf-8', errors='ignore')
pkg = json.loads(pkg_raw)
name = pkg.get('name', 'custom-theme')
dest = os.path.expanduser(os.path.join('~/.config/unirtm/themes', name))
os.makedirs(dest, exist_ok=True)
for member in zf.namelist():
    if member.startswith('extension/') and not member.endswith('/'):
        rel = member[len('extension/'):]
        target = os.path.join(dest, rel)
        os.makedirs(os.path.dirname(target), exist_ok=True)
        with open(target, 'wb') as f:
            f.write(zf.read(member))
print(json.dumps({'name': name, 'version': pkg.get('version', '0.0.1'), 'themes': pkg.get('contributes', {}).get('themes', []), 'iconThemes': pkg.get('contributes', {}).get('iconThemes', []), 'productIconThemes': pkg.get('contributes', {}).get('productIconThemes', [])}))
`;
                const outStr = execFileSync("python3", ["-c", script, tmpZip], { encoding: "utf-8" });
                try {
                  fs.unlinkSync(tmpZip);
                } catch {}
                const parsed = JSON.parse(outStr.trim());
                const installed: any[] = [];
                for (const t of parsed.themes || []) {
                  installed.push({
                    id: t.id || t.label,
                    name: t.label || t.id,
                    type: "color",
                    is_builtin: false,
                    ui_theme: t.uiTheme || "vs-dark",
                    package_id: parsed.name,
                    version: parsed.version,
                  });
                }
                for (const t of parsed.iconThemes || []) {
                  installed.push({
                    id: t.id || t.label,
                    name: t.label || t.id,
                    type: "icon",
                    is_builtin: false,
                    package_id: parsed.name,
                    version: parsed.version,
                  });
                }
                for (const t of parsed.productIconThemes || []) {
                  installed.push({
                    id: t.id || t.label,
                    name: t.label || t.id,
                    type: "product",
                    is_builtin: false,
                    package_id: parsed.name,
                    version: parsed.version,
                  });
                }
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: true, installed }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url === "/api/themes/install" && req.method === "POST") {
            const chunks: any[] = [];
            req.on("data", (c: any) => {
              chunks.push(Buffer.isBuffer(c) ? c : Buffer.from(c));
            });
            req.on("end", async () => {
              try {
                const totalBuffer = Buffer.concat(chunks);
                const tmpZip = path.join(os.tmpdir(), `unirtm_theme_upload_${Date.now()}.vsix`);

                const script = `
import sys, io, zipfile, json, os

raw = open(sys.argv[1], 'rb').read()
idx = raw.find(b'PK\\x03\\x04')
if idx == -1:
    print(json.dumps({'error': 'invalid zip archive'}))
    sys.exit(1)
eocd = raw.rfind(b'PK\\x05\\x06')
zip_bytes = raw[idx:eocd + 22] if eocd != -1 else raw[idx:]
zf = zipfile.ZipFile(io.BytesIO(zip_bytes), 'r')

pkg_raw = zf.read('extension/package.json').decode('utf-8', errors='ignore')
pkg = json.loads(pkg_raw)
name = pkg.get('name', 'custom-theme')
dest = os.path.expanduser(os.path.join('~/.config/unirtm/themes', name))
os.makedirs(dest, exist_ok=True)
for member in zf.namelist():
    if member.startswith('extension/') and not member.endswith('/'):
        rel = member[len('extension/'):]
        target = os.path.join(dest, rel)
        os.makedirs(os.path.dirname(target), exist_ok=True)
        with open(target, 'wb') as f:
            f.write(zf.read(member))
print(json.dumps({
    'name': name,
    'version': pkg.get('version', '0.0.1'),
    'themes': pkg.get('contributes', {}).get('themes', []),
    'iconThemes': pkg.get('contributes', {}).get('iconThemes', []),
    'productIconThemes': pkg.get('contributes', {}).get('productIconThemes', [])
}))
`;
                fs.writeFileSync(tmpZip, totalBuffer);
                const outStr = execFileSync("python3", ["-c", script, tmpZip], { encoding: "utf-8" });
                try {
                  fs.unlinkSync(tmpZip);
                } catch {}
                const parsed = JSON.parse(outStr.trim());
                if (parsed.error) {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: parsed.error }));
                  return;
                }

                const installed: any[] = [];
                for (const t of parsed.themes || []) {
                  installed.push({
                    id: t.id || t.label,
                    name: t.label || t.id,
                    type: "color",
                    is_builtin: false,
                    ui_theme: t.uiTheme || "vs-dark",
                    package_id: parsed.name,
                    version: parsed.version,
                  });
                }
                for (const t of parsed.iconThemes || []) {
                  installed.push({
                    id: t.id || t.label,
                    name: t.label || t.id,
                    type: "icon",
                    is_builtin: false,
                    package_id: parsed.name,
                    version: parsed.version,
                  });
                }
                for (const t of parsed.productIconThemes || []) {
                  installed.push({
                    id: t.id || t.label,
                    name: t.label || t.id,
                    type: "product",
                    is_builtin: false,
                    package_id: parsed.name,
                    version: parsed.version,
                  });
                }
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: true, installed }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/themes/assets/") && req.method === "GET") {
            const rel = req.url.slice("/api/themes/assets/".length).split("?")[0];
            const parts = rel.split("/");
            if (parts.length >= 2) {
              let packageId = "";
              let subpathParts: string[] = [];
              if (["product", "products", "icon", "icons", "color", "colors"].includes(parts[0])) {
                packageId = decodeURIComponent(parts[1]);
                subpathParts = parts.slice(2);
              } else {
                packageId = decodeURIComponent(parts[0]);
                subpathParts = parts.slice(1);
              }

              const subpath = subpathParts.map(decodeURIComponent).join("/");
              const candidateDirs = [
                path.join(os.homedir(), ".config", "unirtm", "themes", packageId),
                path.join(os.homedir(), ".local", "share", "unirtm", "themes", "product_icons", packageId),
                path.join(os.homedir(), ".local", "share", "unirtm", "themes", "icons", packageId),
                path.join(os.homedir(), ".local", "share", "unirtm", "themes", "colors", packageId),
                path.join(os.homedir(), ".local", "share", "unirtm", "themes", packageId),
              ];

              for (const dir of candidateDirs) {
                const targetFile = path.join(dir, subpath);
                if (fs.existsSync(targetFile) && fs.statSync(targetFile).isFile()) {
                  if (targetFile.endsWith(".json")) res.setHeader("Content-Type", "application/json");
                  else if (targetFile.endsWith(".svg")) res.setHeader("Content-Type", "image/svg+xml");
                  else if (targetFile.endsWith(".woff2")) res.setHeader("Content-Type", "font/woff2");
                  else if (targetFile.endsWith(".woff")) res.setHeader("Content-Type", "font/woff");
                  else if (targetFile.endsWith(".ttf")) res.setHeader("Content-Type", "font/ttf");
                  else if (targetFile.endsWith(".eot")) res.setHeader("Content-Type", "application/vnd.ms-fontobject");
                  else if (targetFile.endsWith(".png")) res.setHeader("Content-Type", "image/png");
                  res.end(fs.readFileSync(targetFile));
                  return;
                }
              }
            }
            res.statusCode = 404;
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ error: "Theme asset not found" }));
            return;
          }

          if ((req.url === "/api/themes" || req.url?.startsWith("/api/themes?")) && req.method === "GET") {
            const colorThemes: any[] = [
              { id: "default", name: "UniRTM (默认)", type: "color", is_builtin: true, ui_theme: "vs-dark" },
              { id: "dark", name: "UniRTM Dark", type: "color", is_builtin: true, ui_theme: "vs-dark" },
              { id: "light", name: "UniRTM Light", type: "color", is_builtin: true, ui_theme: "vs" },
            ];
            const iconThemes: any[] = [
              {
                id: "seti",
                name: "Seti UI (VS Code Default)",
                type: "icon",
                is_builtin: true,
                description: "Official Visual Studio Code default file icon theme",
              },
            ];
            const productIconThemes: any[] = [
              {
                id: "codicons",
                name: "Codicons (默认)",
                type: "product",
                is_builtin: true,
                description: "Official Visual Studio Code product icon theme",
              },
            ];

            const scanDir = (themesDir: string) => {
              if (!fs.existsSync(themesDir)) return;
              try {
                const dirs = fs.readdirSync(themesDir, { withFileTypes: true });
                for (const d of dirs) {
                  if (d.isDirectory()) {
                    const pkgJson = path.join(themesDir, d.name, "package.json");
                    if (fs.existsSync(pkgJson)) {
                      try {
                        const p = JSON.parse(fs.readFileSync(pkgJson, "utf-8"));
                        const c = p.contributes || {};
                        const iconPath = p.icon ? p.icon.replace(/^\.\//, "") : "";
                        const iconUrl = iconPath ? `/api/themes/assets/icons/${d.name}/${iconPath}` : "";
                        for (const ct of c.themes || []) {
                          const id = ct.id || ct.label;
                          if (!colorThemes.some((item) => item.id === id)) {
                            colorThemes.push({
                              id,
                              name: ct.label || ct.id,
                              type: "color",
                              is_builtin: false,
                              ui_theme: ct.uiTheme || "vs-dark",
                              package_id: p.name,
                              version: p.version,
                              relative_path: ct.path,
                              icon_url: iconUrl,
                            });
                          }
                        }
                        for (const it of c.iconThemes || []) {
                          const id = it.id || it.label;
                          if (!iconThemes.some((item) => item.id === id)) {
                            iconThemes.push({
                              id,
                              name: it.label || it.id,
                              type: "icon",
                              is_builtin: false,
                              package_id: p.name,
                              version: p.version,
                              relative_path: it.path,
                              icon_url: iconUrl,
                            });
                          }
                        }
                        for (const pit of c.productIconThemes || []) {
                          const id = pit.id || pit.label;
                          if (!productIconThemes.some((item) => item.id === id)) {
                            productIconThemes.push({
                              id,
                              name: pit.label || pit.id,
                              type: "product",
                              is_builtin: false,
                              package_id: p.name,
                              version: p.version,
                              relative_path: pit.path,
                              icon_url: iconUrl,
                            });
                          }
                        }
                      } catch {}
                    }
                  }
                }
              } catch {}
            };

            scanDir(path.join(os.homedir(), ".config", "unirtm", "themes"));
            scanDir(path.join(os.homedir(), ".local", "share", "unirtm", "themes", "product_icons"));
            scanDir(path.join(os.homedir(), ".local", "share", "unirtm", "themes", "icons"));
            scanDir(path.join(os.homedir(), ".local", "share", "unirtm", "themes", "colors"));
            scanDir(path.join(os.homedir(), ".local", "share", "unirtm", "themes"));

            const userConfigDir = path.join(os.homedir(), ".config", "unirtm");
            const userPath = path.join(userConfigDir, "unirtm.toml");
            let activeColor = "default";
            let activeIcon = "seti";
            let activeProduct = "codicons";
            let activeAppearance = "system";

            const readSetting = (fp: string) => {
              if (!fs.existsSync(fp)) return;
              const text = fs.readFileSync(fp, "utf-8");
              const mApp = text.match(/appearance\s*=\s*"([^"]+)"/);
              if (mApp) {
                activeAppearance = mApp[1];
              }
              const mColor = text.match(/(?:color_)?theme\s*=\s*"([^"]+)"/);
              if (mColor) {
                const val = mColor[1];
                if (val === "dark" || val === "light" || val === "system") {
                  activeAppearance = val;
                  activeColor = "default";
                } else {
                  activeColor = val;
                }
              }
              const mIcon = text.match(/icon_theme\s*=\s*"([^"]+)"/);
              if (mIcon) {
                activeIcon = mIcon[1];
                if (activeIcon === "vs-seti") activeIcon = "seti";
              }
              const mProd = text.match(/product_icon_theme\s*=\s*"([^"]+)"/);
              if (mProd) {
                activeProduct = mProd[1];
              }
            };
            readSetting(userPath);

            res.setHeader("Content-Type", "application/json");
            res.end(
              JSON.stringify({
                active: {
                  color: activeColor,
                  icon: activeIcon,
                  product: activeProduct,
                  appearance: activeAppearance,
                },
                active_color_theme: activeColor,
                active_icon_theme: activeIcon,
                active_product_icon_theme: activeProduct,
                colors: colorThemes,
                icons: iconThemes,
                products: productIconThemes,
                color_themes: colorThemes,
                icon_themes: iconThemes,
                product_icon_themes: productIconThemes,
              })
            );
            return;
          }

          if (req.url?.startsWith("/api/themes") && req.method === "DELETE") {
            const urlObj = new URL(req.url, "http://localhost");
            const id = urlObj.searchParams.get("id");
            const candidateDirs = [
              path.join(os.homedir(), ".config", "unirtm", "themes"),
              path.join(os.homedir(), ".local", "share", "unirtm", "themes", "product_icons"),
              path.join(os.homedir(), ".local", "share", "unirtm", "themes", "icons"),
              path.join(os.homedir(), ".local", "share", "unirtm", "themes", "colors"),
              path.join(os.homedir(), ".local", "share", "unirtm", "themes"),
            ];
            if (id) {
              for (const themesDir of candidateDirs) {
                if (fs.existsSync(themesDir)) {
                  try {
                    for (const d of fs.readdirSync(themesDir)) {
                      const itemDir = path.join(themesDir, d);
                      const pkgJson = path.join(itemDir, "package.json");
                      if (fs.existsSync(pkgJson)) {
                        try {
                          const p = JSON.parse(fs.readFileSync(pkgJson, "utf-8"));
                          const c = p.contributes || {};
                          const hasTheme =
                            (c.themes || []).some((t: any) => (t.id || t.label) === id) ||
                            (c.iconThemes || []).some((t: any) => (t.id || t.label) === id) ||
                            (c.productIconThemes || []).some((t: any) => (t.id || t.label) === id) ||
                            p.name === id;
                          if (hasTheme) {
                            fs.rmSync(itemDir, { recursive: true, force: true });
                          }
                        } catch {}
                      }
                    }
                  } catch {}
                }
              }
            }
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ success: true }));
            return;
          }

          if (req.url === "/api/themes/apply" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const body = JSON.parse(bodyStr || "{}");
                const color_theme_id = body.color || body.color_theme_id;
                const icon_theme_id = body.icon || body.icon_theme_id;
                const product_theme_id = body.product || body.product_icon || body.product_icon_theme_id;
                const appearance = body.appearance;
                const unirtmBin = findUnirtmBinary(rootDir);
                if (fs.existsSync(unirtmBin)) {
                  try {
                    if (color_theme_id) {
                      execFileSync(unirtmBin, ["config", "set", "--global", "settings.theme", String(color_theme_id)], {
                        cwd: rootDir,
                        stdio: "pipe",
                      });
                    }
                    if (icon_theme_id) {
                      execFileSync(
                        unirtmBin,
                        ["config", "set", "--global", "settings.icon_theme", String(icon_theme_id)],
                        { cwd: rootDir, stdio: "pipe" }
                      );
                    }
                    if (product_theme_id) {
                      execFileSync(
                        unirtmBin,
                        ["config", "set", "--global", "settings.product_icon_theme", String(product_theme_id)],
                        { cwd: rootDir, stdio: "pipe" }
                      );
                    }
                    if (appearance) {
                      execFileSync(
                        unirtmBin,
                        ["config", "set", "--global", "settings.appearance", String(appearance)],
                        { cwd: rootDir, stdio: "pipe" }
                      );
                    }
                  } catch (e) {
                    console.warn("[Vite Middleware] unirtm config set fallback:", e);
                  }
                }

                // Fallback direct update to ~/.config/unirtm/unirtm.toml
                const userConfigDir = path.join(os.homedir(), ".config", "unirtm");
                const userPath = path.join(userConfigDir, "unirtm.toml");
                if (fs.existsSync(userPath)) {
                  let text = fs.readFileSync(userPath, "utf-8");
                  if (color_theme_id) {
                    text = text.match(/theme\s*=\s*"[^"]+"/)
                      ? text.replace(/theme\s*=\s*"[^"]+"/, `theme = "${color_theme_id}"`)
                      : text + `\n[settings]\ntheme = "${color_theme_id}"\n`;
                  }
                  if (icon_theme_id) {
                    text = text.match(/icon_theme\s*=\s*"[^"]+"/)
                      ? text.replace(/icon_theme\s*=\s*"[^"]+"/, `icon_theme = "${icon_theme_id}"`)
                      : text + `\nicon_theme = "${icon_theme_id}"\n`;
                  }
                  if (product_theme_id) {
                    text = text.match(/product_icon_theme\s*=\s*"[^"]+"/)
                      ? text.replace(/product_icon_theme\s*=\s*"[^"]+"/, `product_icon_theme = "${product_theme_id}"`)
                      : text + `\nproduct_icon_theme = "${product_theme_id}"\n`;
                  }
                  if (appearance) {
                    text = text.match(/appearance\s*=\s*"[^"]+"/)
                      ? text.replace(/appearance\s*=\s*"[^"]+"/, `appearance = "${appearance}"`)
                      : text + `\nappearance = "${appearance}"\n`;
                  }
                  fs.writeFileSync(userPath, text, "utf-8");
                }

                res.setHeader("Content-Type", "application/json");
                res.end(
                  JSON.stringify({
                    success: true,
                    color: color_theme_id,
                    icon: icon_theme_id,
                    product: product_theme_id,
                    appearance,
                  })
                );
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          // ─── File System API Handlers ─────────────────────────────────────
          if (req.url?.startsWith("/api/fs/")) {
            const getSafeFSPath = (relPath: string) => {
              const cleanRel = (relPath || "").trim().replace(/^\/+/, "");
              const target = path.resolve(rootDir, cleanRel);
              const rel = path.relative(rootDir, target);
              if (rel.startsWith("..") || path.isAbsolute(rel)) {
                throw new Error("Access denied: path outside workspace");
              }
              return { rootDir, safePath: target, relPath: rel === "" ? "" : rel.replace(/\\/g, "/") };
            };

            const isCommonIgnored = (name: string) => {
              return [".git", "node_modules", ".DS_Store", "dist", "target", ".turbo", ".idea", ".vscode", ".unirtm"].includes(name);
            };

            const buildFSTree = (dirPath: string, currentRel: string, currentDepth: number, maxDepth: number, showHidden: boolean): any[] => {
              if (!fs.existsSync(dirPath)) return [];
              const items = fs.readdirSync(dirPath, { withFileTypes: true });
              const dirs: any[] = [];
              const files: any[] = [];

              for (const item of items) {
                const name = item.name;
                if (!showHidden && name.startsWith(".") && name !== ".unirtm.toml" && name !== ".gitignore" && name !== ".env") {
                  continue;
                }
                const entryRel = currentRel ? `${currentRel}/${name}` : name;
                const fullEntryPath = path.join(dirPath, name);
                const isDir = item.isDirectory();
                let size = 0;
                let modTime = "";
                try {
                  const stat = fs.statSync(fullEntryPath);
                  size = stat.size;
                  modTime = stat.mtime.toISOString();
                } catch (_) {}

                const entry: any = {
                  name,
                  path: entryRel,
                  is_dir: isDir,
                  size,
                  mod_time: modTime,
                  is_ignored: isCommonIgnored(name),
                };

                if (isDir) {
                  if (currentDepth < maxDepth && name !== ".git" && name !== "node_modules") {
                    entry.children = buildFSTree(fullEntryPath, entryRel, currentDepth + 1, maxDepth, showHidden);
                  }
                  dirs.push(entry);
                } else {
                  files.push(entry);
                }
              }

              dirs.sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));
              files.sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));
              return [...dirs, ...files];
            };

            // GET /api/fs/tree
            if (req.url.startsWith("/api/fs/tree") && req.method === "GET") {
              try {
                const urlObj = new URL(req.url, "http://localhost");
                const p = urlObj.searchParams.get("path") || "";
                const depth = parseInt(urlObj.searchParams.get("depth") || "3", 10);
                const showHidden = urlObj.searchParams.get("showHidden") === "true";

                const { rootDir: rd, safePath, relPath } = getSafeFSPath(p);
                if (!fs.existsSync(safePath) || !fs.statSync(safePath).isDirectory()) {
                  res.statusCode = 404;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "Directory not found" }));
                  return;
                }

                const entries = buildFSTree(safePath, relPath, 1, Math.min(depth || 3, 10), showHidden);
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({
                  root: rd,
                  name: path.basename(rd),
                  path: relPath,
                  entries,
                }));
              } catch (err: any) {
                res.statusCode = err.message.includes("Access denied") ? 403 : 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
              return;
            }

            // POST /api/fs/create
            if (req.url.startsWith("/api/fs/create") && req.method === "POST") {
              let bodyStr = "";
              req.on("data", (c: any) => { bodyStr += c; });
              req.on("end", () => {
                try {
                  const body = JSON.parse(bodyStr || "{}");
                  if (!body.path) {
                    res.statusCode = 400;
                    res.setHeader("Content-Type", "application/json");
                    res.end(JSON.stringify({ error: "path is required" }));
                    return;
                  }
                  const { safePath, relPath } = getSafeFSPath(body.path);
                  if (body.is_dir) {
                    fs.mkdirSync(safePath, { recursive: true });
                  } else {
                    fs.mkdirSync(path.dirname(safePath), { recursive: true });
                    if (!fs.existsSync(safePath)) {
                      fs.writeFileSync(safePath, "");
                    }
                  }
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ success: true, path: relPath }));
                } catch (err: any) {
                  res.statusCode = err.message.includes("Access denied") ? 403 : 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }

            // POST /api/fs/rename
            if (req.url.startsWith("/api/fs/rename") && req.method === "POST") {
              let bodyStr = "";
              req.on("data", (c: any) => { bodyStr += c; });
              req.on("end", () => {
                try {
                  const body = JSON.parse(bodyStr || "{}");
                  if (!body.old_path || !body.new_path) {
                    res.statusCode = 400;
                    res.setHeader("Content-Type", "application/json");
                    res.end(JSON.stringify({ error: "old_path and new_path are required" }));
                    return;
                  }
                  const { safePath: oldTarget, relPath: oldRel } = getSafeFSPath(body.old_path);
                  const { safePath: newTarget, relPath: newRel } = getSafeFSPath(body.new_path);
                  fs.renameSync(oldTarget, newTarget);
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ success: true, old_path: oldRel, new_path: newRel }));
                } catch (err: any) {
                  res.statusCode = err.message.includes("Access denied") ? 403 : 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }

            // POST /api/fs/delete
            if ((req.url.startsWith("/api/fs/delete") || req.url === "/api/fs") && (req.method === "POST" || req.method === "DELETE")) {
              let bodyStr = "";
              req.on("data", (c: any) => { bodyStr += c; });
              req.on("end", () => {
                try {
                  const urlObj = new URL(req.url!, "http://localhost");
                  const body = JSON.parse(bodyStr || "{}");
                  const targetP = body.path || urlObj.searchParams.get("path") || "";
                  if (!targetP) {
                    res.statusCode = 400;
                    res.setHeader("Content-Type", "application/json");
                    res.end(JSON.stringify({ error: "path is required" }));
                    return;
                  }
                  const { safePath, relPath, rootDir: rd } = getSafeFSPath(targetP);
                  if (safePath === rd) {
                    res.statusCode = 400;
                    res.setHeader("Content-Type", "application/json");
                    res.end(JSON.stringify({ error: "Cannot delete workspace root" }));
                    return;
                  }
                  fs.rmSync(safePath, { recursive: true, force: true });
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ success: true, path: relPath }));
                } catch (err: any) {
                  res.statusCode = err.message.includes("Access denied") ? 403 : 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }

            // GET /api/fs/read
            if (req.url.startsWith("/api/fs/read") && req.method === "GET") {
              try {
                const urlObj = new URL(req.url, "http://localhost");
                const targetP = urlObj.searchParams.get("path") || "";
                if (!targetP) {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "path is required" }));
                  return;
                }
                const { safePath, relPath } = getSafeFSPath(targetP);
                if (!fs.existsSync(safePath) || fs.statSync(safePath).isDirectory()) {
                  res.statusCode = 404;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "File not found" }));
                  return;
                }
                const stat = fs.statSync(safePath);
                if (stat.size > 10 * 1024 * 1024) {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "File too large (exceeds 10MB)" }));
                  return;
                }
                const content = fs.readFileSync(safePath, "utf-8");
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ path: relPath, content, size: stat.size }));
              } catch (err: any) {
                res.statusCode = err.message.includes("Access denied") ? 403 : 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
              return;
            }

            // POST /api/fs/write
            if (req.url.startsWith("/api/fs/write") && (req.method === "POST" || req.method === "PUT")) {
              let bodyStr = "";
              req.on("data", (c: any) => { bodyStr += c; });
              req.on("end", () => {
                try {
                  const body = JSON.parse(bodyStr || "{}");
                  if (!body.path) {
                    res.statusCode = 400;
                    res.setHeader("Content-Type", "application/json");
                    res.end(JSON.stringify({ error: "path is required" }));
                    return;
                  }
                  const { safePath, relPath } = getSafeFSPath(body.path);
                  fs.mkdirSync(path.dirname(safePath), { recursive: true });
                  fs.writeFileSync(safePath, body.content || "", "utf-8");
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ success: true, path: relPath, size: (body.content || "").length }));
                } catch (err: any) {
                  res.statusCode = err.message.includes("Access denied") ? 403 : 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }
          }

          if (req.url === "/api/system/browsers") {
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ browsers: [] }));
            return;
          }

          if (req.url === "/api/terminals" || req.url?.startsWith("/api/terminals?")) {
            res.setHeader("Content-Type", "application/json");
            const defaultShell =
              process.env.SHELL || (process.platform === "win32" ? process.env.COMSPEC || "cmd.exe" : "/bin/bash");
            let realDefaultShell = "";
            try {
              if (defaultShell && fs.existsSync(defaultShell)) {
                realDefaultShell = fs.realpathSync(defaultShell);
              }
            } catch (_) {}

            const isRestrictedOrStub = (baseName: string) => {
              const b = baseName.toLowerCase().replace(/\.exe$/, "");
              return ["rbash", "rksh", "rzsh", "rsh", "nologin", "false", "sync", "halt", "shutdown"].includes(b);
            };

            interface InternalTerm {
              id: string;
              name: string;
              path: string;
              realPath: string;
              devIno: string;
            }

            const terms: InternalTerm[] = [];

            const addTerm = (id: string, name: string, shPath: string) => {
              if (!shPath) return;
              const base = path
                .basename(shPath)
                .toLowerCase()
                .replace(/\.exe$/, "");
              if (isRestrictedOrStub(base)) return;

              try {
                if (!fs.existsSync(shPath)) return;
                let realPath = shPath;
                try {
                  realPath = fs.realpathSync(shPath);
                } catch (_) {}

                let devIno = "";
                try {
                  const stat = fs.statSync(shPath);
                  devIno = `${stat.dev}:${stat.ino}`;
                } catch (_) {}

                for (let i = 0; i < terms.length; i++) {
                  const ex = terms[i];
                  if (ex.path === shPath) return;

                  const sameUnderlying = (realPath && ex.realPath === realPath) || (devIno && ex.devIno === devIno);
                  if (sameUnderlying) {
                    if (ex.id === id || (id === "dash" && ex.id === "sh") || (id === "sh" && ex.id === "dash")) {
                      if (defaultShell && shPath === defaultShell) {
                        ex.path = shPath;
                        ex.id = id;
                        ex.name = name;
                      }
                      return;
                    }
                  }

                  if (ex.id === id) {
                    if (defaultShell && shPath === defaultShell) {
                      ex.path = shPath;
                      ex.name = name;
                    }
                    return;
                  }
                }

                terms.push({ id, name, path: shPath, realPath, devIno });
              } catch (_) {}
            };

            if (process.platform !== "win32") {
              if (fs.existsSync("/etc/shells")) {
                try {
                  const lines = fs.readFileSync("/etc/shells", "utf-8").split("\n");
                  for (const line of lines) {
                    const clean = line.trim();
                    if (!clean || clean.startsWith("#")) continue;
                    const b = path.basename(clean);
                    let name = b.charAt(0).toUpperCase() + b.slice(1);
                    if (b === "sh") name = "POSIX sh";
                    if (b === "pwsh") name = "PowerShell Core";
                    addTerm(b.toLowerCase(), name, clean);
                  }
                } catch (_) {}
              }
              const candidates = [
                { id: "bash", name: "Bash", bin: "/bin/bash" },
                { id: "zsh", name: "Zsh", bin: "/usr/bin/zsh" },
                { id: "fish", name: "Fish", bin: "/usr/bin/fish" },
                { id: "sh", name: "POSIX sh", bin: "/bin/sh" },
                { id: "pwsh", name: "PowerShell Core", bin: "/usr/bin/pwsh" },
              ];
              for (const c of candidates) {
                addTerm(c.id, c.name, c.bin);
              }
            } else {
              const winDir = process.env.SystemRoot || "C:\\Windows";
              const winFallbacks = [
                {
                  id: "powershell",
                  name: "Windows PowerShell",
                  path: path.join(winDir, "System32\\WindowsPowerShell\\v1.0\\powershell.exe"),
                },
                {
                  id: "pwsh",
                  name: "PowerShell Core",
                  path: "C:\\Program Files\\PowerShell\\7\\pwsh.exe",
                },
                {
                  id: "pwsh",
                  name: "PowerShell Core",
                  path: "C:\\Program Files\\PowerShell\\6\\pwsh.exe",
                },
                {
                  id: "cmd",
                  name: "Command Prompt",
                  path: path.join(winDir, "System32\\cmd.exe"),
                },
                {
                  id: "wsl",
                  name: "WSL",
                  path: path.join(winDir, "System32\\wsl.exe"),
                },
              ];
              for (const fb of winFallbacks) {
                addTerm(fb.id, fb.name, fb.path);
              }

              const gitBashPaths = [
                "C:\\Program Files\\Git\\bin\\bash.exe",
                "C:\\Program Files (x86)\\Git\\bin\\bash.exe",
                process.env.LOCALAPPDATA
                  ? path.join(process.env.LOCALAPPDATA, "Programs\\Git\\bin\\bash.exe")
                  : "",
                process.env.USERPROFILE
                  ? path.join(process.env.USERPROFILE, "AppData\\Local\\Programs\\Git\\bin\\bash.exe")
                  : "",
              ];
              for (const gbp of gitBashPaths) {
                if (gbp && fs.existsSync(gbp)) {
                  addTerm("git-bash", "Git Bash", gbp);
                  break;
                }
              }
            }

            if (defaultShell && fs.existsSync(defaultShell)) {
              const b = path.basename(defaultShell);
              addTerm(b.toLowerCase(), b.charAt(0).toUpperCase() + b.slice(1), defaultShell);
            }

            // Resolve is_default: Exactly ONE terminal must be default
            let defaultIdx = -1;
            if (defaultShell) {
              defaultIdx = terms.findIndex((t) => t.path === defaultShell);
            }
            if (defaultIdx === -1 && realDefaultShell) {
              defaultIdx = terms.findIndex((t) => t.realPath === realDefaultShell);
            }
            if (defaultIdx === -1 && defaultShell) {
              const defBase = path
                .basename(defaultShell)
                .toLowerCase()
                .replace(/\.exe$/, "");
              defaultIdx = terms.findIndex((t) => t.id === defBase);
            }
            if (defaultIdx === -1 && terms.length > 0) {
              defaultIdx = 0;
            }

            const formattedTerms = terms.map((t, idx) => ({
              id: t.id,
              name: t.name,
              path: t.path,
              is_default: idx === defaultIdx,
            }));

            formattedTerms.sort((a, b) => (b.is_default ? 1 : 0) - (a.is_default ? 1 : 0));

            const defTerm = formattedTerms.find((t) => t.is_default)?.path || formattedTerms[0]?.path || "/bin/bash";
            let current = defTerm;
            try {
              const home = os.homedir();
              const cfgPath = path.join(home, ".unirtm.toml");
              if (fs.existsSync(cfgPath)) {
                const c = fs.readFileSync(cfgPath, "utf-8");
                const m = c.match(/terminal\s*=\s*"([^"]+)"/);
                if (m && m[1]) {
                  current = m[1];
                }
              }
            } catch (_) {}

            res.end(
              JSON.stringify({
                default: defTerm,
                current: current,
                terminals: formattedTerms,
              })
            );
            return;
          }

          if (req.url === "/api/terminal/exec" && req.method === "POST") {
            res.setHeader("Content-Type", "application/json");
            let body = "";
            req.on("data", (chunk) => {
              body += chunk;
            });
            req.on("end", () => {
              try {
                const parsed = JSON.parse(body || "{}");
                const command = (parsed.command || "").trim();
                const shell = parsed.shell || process.env.SHELL || "/bin/bash";
                const cwd = parsed.cwd || process.cwd();

                if (!command) {
                  res.end(JSON.stringify({ output: "", exit_code: 0 }));
                  return;
                }

                const home = os.homedir();
                const unirtmBin = path.join(home, ".unirtm", "bin");
                const unirtmShims = path.join(home, ".unirtm", "shims");
                const customPath = `${unirtmShims}:${unirtmBin}:${process.env.PATH || ""}`;

                exec(
                  command,
                  {
                    shell: shell,
                    cwd: cwd,
                    env: { ...process.env, PATH: customPath },
                    maxBuffer: 10 * 1024 * 1024,
                  },
                  (error: any, stdout: string, stderr: string) => {
                    const combined = (stdout || "") + (stderr || "");
                    const exitCode = error ? error.code || 1 : 0;
                    res.end(
                      JSON.stringify({
                        output: combined,
                        exit_code: exitCode,
                        error: error ? error.message : undefined,
                      })
                    );
                  }
                );
              } catch (err: any) {
                res.statusCode = 500;
                res.end(JSON.stringify({ error: err.message, exit_code: 1 }));
              }
            });
            return;
          }
          // Tasks API: Read tasks from .unirtm.toml or user global config
          if (req.url?.startsWith("/api/tasks/run") && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const payload = JSON.parse(bodyStr || "{}");
                const taskName = payload.name || payload.task || payload.taskName || payload.id;
                const scope = payload.scope;

                if (!taskName || typeof taskName !== "string" || taskName.trim() === "" || taskName === "undefined") {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(
                    JSON.stringify({
                      success: false,
                      exit_code: 1,
                      error: "Task name is required",
                    })
                  );
                  return;
                }

                const trimmedName = taskName.trim();
                const unirtmBin = findUnirtmBinary(rootDir);
                let output = "";
                let exitCode = 0;
                if (fs.existsSync(unirtmBin)) {
                  try {
                    const args = ["run", trimmedName];
                    if (scope === "user") args.push("--global");
                    output = execFileSync(unirtmBin, args, { cwd: rootDir, encoding: "utf-8", timeout: 60000 });
                  } catch (e: any) {
                    exitCode = e.status || 1;
                    output = (e.stdout || "") + "\n" + (e.stderr || "") + "\n" + (e.message || "");
                  }
                } else {
                  output = `[Mock] Executed task ${trimmedName} in ${scope} scope. (unirtm binary not found, simulated success)`;
                }
                res.setHeader("Content-Type", "application/json");
                res.end(
                  JSON.stringify({
                    success: exitCode === 0,
                    exit_code: exitCode,
                    output: output.trim(),
                    name: trimmedName,
                    task: trimmedName,
                  })
                );
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message, success: false, exit_code: 1 }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/tasks") && req.method === "GET") {
            const urlObj = new URL(req.url, "http://localhost");
            const scope = urlObj.searchParams.get("scope") || "workspace";
            const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
            const wsPath = path.join(rootDir, ".unirtm.toml");
            const targetPath = scope === "user" ? userPath : wsPath;

            const parseTasksFromToml = (filePath: string) => {
              if (!fs.existsSync(filePath)) return [];
              const content = fs.readFileSync(filePath, "utf-8");
              const lines = content.split("\n");
              const tasks: any[] = [];
              let currentTask: any = null;
              let inRun = false;
              let runBuf: string[] = [];
              let inDeps = false;
              let depsBuf: string[] = [];

              const save = () => {
                if (currentTask && currentTask.name) {
                  if (!currentTask.command && currentTask.depends?.length) {
                    currentTask.command = `unirtm run ${currentTask.depends.join(" ")}`;
                  }
                  tasks.push(currentTask);
                }
                currentTask = null;
                inRun = false;
                runBuf = [];
                inDeps = false;
                depsBuf = [];
              };

              for (const raw of lines) {
                const tr = raw.trim();
                const m = tr.match(/^\[tasks\.(?:"([^"]+)"|'([^']+)'|([a-zA-Z0-9_\-:]+))\]$/);
                if (m) {
                  save();
                  currentTask = { name: m[1] || m[2] || m[3], description: "", command: "", depends: [], env: "" };
                  continue;
                }
                if (tr.startsWith("[") && tr.endsWith("]") && !tr.startsWith("[tasks.")) {
                  save();
                  continue;
                }
                if (!currentTask) continue;

                if (inRun) {
                  if (tr.endsWith("'''") || tr.endsWith('"""')) {
                    const last = tr.slice(0, -3);
                    if (last) runBuf.push(last);
                    currentTask.command = runBuf
                      .map((l) => l.trim())
                      .filter(Boolean)
                      .join(" ; ");
                    inRun = false;
                    runBuf = [];
                  } else {
                    runBuf.push(raw);
                  }
                  continue;
                }

                if (inDeps) {
                  if (tr.includes("]")) {
                    const before = tr.slice(0, tr.indexOf("]"));
                    const items = before
                      .split(",")
                      .map((s) => s.trim().replace(/^["']|["']$/g, ""))
                      .filter(Boolean);
                    depsBuf.push(...items);
                    currentTask.depends = [...depsBuf];
                    inDeps = false;
                    depsBuf = [];
                  } else {
                    const items = tr
                      .split(",")
                      .map((s) => s.trim().replace(/^["']|["']$/g, ""))
                      .filter(Boolean);
                    depsBuf.push(...items);
                  }
                  continue;
                }

                if (tr.startsWith("description =")) {
                  currentTask.description = tr.replace(/^description\s*=\s*["']?/, "").replace(/["']?$/, "");
                } else if (tr.startsWith("run =")) {
                  const rest = tr.slice(tr.indexOf("=") + 1).trim();
                  if (rest.startsWith("'''") || rest.startsWith('"""')) {
                    if ((rest.endsWith("'''") || rest.endsWith('"""')) && rest.length > 6) {
                      currentTask.command = rest.slice(3, -3).trim();
                    } else {
                      inRun = true;
                      runBuf = [rest.slice(3)];
                    }
                  } else {
                    currentTask.command = rest.replace(/^["']|["']$/g, "");
                  }
                } else if (tr.startsWith("depends =")) {
                  const rest = tr.slice(tr.indexOf("=") + 1).trim();
                  if (rest.startsWith("[") && rest.endsWith("]")) {
                    currentTask.depends = rest
                      .slice(1, -1)
                      .split(",")
                      .map((s) => s.trim().replace(/^["']|["']$/g, ""))
                      .filter(Boolean);
                  } else if (rest.startsWith("[")) {
                    inDeps = true;
                    depsBuf = rest
                      .slice(1)
                      .split(",")
                      .map((s) => s.trim().replace(/^["']|["']$/g, ""))
                      .filter(Boolean);
                  }
                } else if (tr.startsWith("env =")) {
                  const rest = tr.slice(tr.indexOf("=") + 1).trim();
                  currentTask.env = rest.replace(/^\{|\}$/g, "").trim();
                }
              }
              save();
              return tasks;
            };

            const tasks = parseTasksFromToml(targetPath);
            res.setHeader("Content-Type", "application/json");
            res.end(
              JSON.stringify({
                tasks,
                total: tasks.length,
                scope,
                file: targetPath,
              })
            );
            return;
          }

          if (req.url?.startsWith("/api/tasks") && (req.method === "POST" || req.method === "PUT")) {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const body = JSON.parse(bodyStr || "{}");
                const name = (body.name || "").trim();
                if (!name) {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "Task name is required" }));
                  return;
                }
                const scope = (body.scope || "workspace").toLowerCase();
                const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
                const wsPath = path.join(rootDir, ".unirtm.toml");
                const targetPath = scope === "user" ? userPath : wsPath;

                let content = fs.existsSync(targetPath) ? fs.readFileSync(targetPath, "utf-8") : "";
                const lines = content.split("\n");

                const formatTaskLines = () => {
                  const res: string[] = [`[tasks.${name}]`];
                  if (body.description) {
                    res.push(`description = ${JSON.stringify(body.description)}`);
                  }
                  if (body.command) {
                    if (body.command.includes("\n")) {
                      res.push("run = \'\'\'");
                      res.push(body.command);
                      res.push("\'\'\'");
                    } else {
                      res.push(`run = ${JSON.stringify(body.command)}`);
                    }
                  }
                  if (Array.isArray(body.depends) && body.depends.length > 0) {
                    const depsStr = body.depends.map((d: string) => JSON.stringify(d)).join(", ");
                    res.push(`depends = [${depsStr}]`);
                  }
                  return res;
                };

                const findBounds = (tName: string) => {
                  let start = -1;
                  let end = -1;
                  for (let i = 0; i < lines.length; i++) {
                    const tr = lines[i].trim();
                    const m = tr.match(/^\[tasks\.(?:"([^"]+)"|'([^']+)'|([a-zA-Z0-9_\-:]+))\]$/);
                    if (m) {
                      const matched = m[1] || m[2] || m[3];
                      if (start !== -1) {
                        end = i;
                        break;
                      }
                      if (matched === tName) {
                        start = i;
                        continue;
                      }
                    } else if (start !== -1 && tr.startsWith("[") && tr.endsWith("]")) {
                      end = i;
                      break;
                    }
                  }
                  if (start !== -1 && end === -1) end = lines.length;
                  return { start, end };
                };

                const newLines = formatTaskLines();
                const targetMatchName = body.original_name || name;
                const bounds = findBounds(targetMatchName);

                if (bounds.start !== -1) {
                  // In-place replacement preserving position
                  lines.splice(bounds.start, bounds.end - bounds.start, ...newLines);
                } else {
                  // Find insertion point before settings/plugins or after last task
                  let insertAt = -1;
                  for (let i = lines.length - 1; i >= 0; i--) {
                    if (/^\[tasks\.[^\]]+\]$/.test(lines[i].trim())) {
                      for (let j = i + 1; j < lines.length; j++) {
                        const tr = lines[j].trim();
                        if (tr.startsWith("[") && tr.endsWith("]")) {
                          insertAt = j;
                          break;
                        }
                      }
                      if (insertAt === -1) insertAt = lines.length;
                      break;
                    }
                  }
                  if (insertAt === -1) {
                    for (let i = 0; i < lines.length; i++) {
                      const tr = lines[i].trim();
                      if (tr === "[settings]" || tr === "[plugins]" || tr === "[alias]") {
                        insertAt = i;
                        break;
                      }
                    }
                  }
                  if (insertAt !== -1 && insertAt < lines.length) {
                    lines.splice(insertAt, 0, "", ...newLines);
                  } else {
                    lines.push("", ...newLines);
                  }
                }

                fs.mkdirSync(path.dirname(targetPath), { recursive: true });
                fs.writeFileSync(targetPath, lines.join("\n"), "utf-8");

                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: true, task: body }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/tasks") && req.method === "DELETE") {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const urlObj = new URL(req.url!, "http://localhost");
                let name = urlObj.searchParams.get("name") || "";
                let scope = urlObj.searchParams.get("scope") || "";

                if (!name && bodyStr) {
                  try {
                    const parsed = JSON.parse(bodyStr);
                    name = parsed.name || "";
                    scope = scope || parsed.scope || "";
                  } catch {}
                }

                name = name.trim();
                scope = (scope || "workspace").toLowerCase();

                if (!name) {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "Task name is required" }));
                  return;
                }

                const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
                const wsPath = path.join(rootDir, ".unirtm.toml");
                const targetPath = scope === "user" ? userPath : wsPath;

                if (fs.existsSync(targetPath)) {
                  let content = fs.readFileSync(targetPath, "utf-8");
                  const lines = content.split("\n");
                  let start = -1;
                  let end = -1;
                  for (let i = 0; i < lines.length; i++) {
                    const tr = lines[i].trim();
                    const m = tr.match(/^\[tasks\.(?:"([^"]+)"|'([^']+)'|([a-zA-Z0-9_\-:]+))\]$/);
                    if (m) {
                      const matched = m[1] || m[2] || m[3];
                      if (start !== -1) {
                        end = i;
                        break;
                      }
                      if (matched === name) {
                        start = i;
                        continue;
                      }
                    } else if (start !== -1 && tr.startsWith("[") && tr.endsWith("]")) {
                      end = i;
                      break;
                    }
                  }
                  if (start !== -1) {
                    if (end === -1) end = lines.length;
                    lines.splice(start, end - start);
                    fs.writeFileSync(targetPath, lines.join("\n"), "utf-8");
                  }
                }

                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: true }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          // Environment Variables APIs (Mock Dev Server)
          if (req.url?.startsWith("/api/env") && req.method === "GET") {
            const urlObj = new URL(req.url, "http://localhost");
            const scope = (urlObj.searchParams.get("scope") || "workspace").toLowerCase();
            const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
            const wsPath = path.join(rootDir, ".unirtm.toml");
            const targetPath = scope === "user" ? userPath : wsPath;

            const parseEnvFromToml = (filePath: string) => {
              if (!fs.existsSync(filePath)) return [];
              const content = fs.readFileSync(filePath, "utf-8");
              const lines = content.split("\n");
              const list: any[] = [];
              let inEnv = false;

              for (const raw of lines) {
                const tr = raw.trim();
                if (tr === "[env]") {
                  inEnv = true;
                  continue;
                }
                if (inEnv) {
                  if (tr.startsWith("[") && tr !== "[env]") {
                    break;
                  }
                  if (tr.startsWith("#") || !tr.includes("=")) continue;
                  const eqIdx = tr.indexOf("=");
                  const k = tr
                    .slice(0, eqIdx)
                    .trim()
                    .replace(/^["']|["']$/g, "")
                    .toUpperCase();
                  const rawVal = tr.slice(eqIdx + 1).trim();
                  let val = rawVal;
                  if (
                    (rawVal.startsWith('"') && rawVal.endsWith('"')) ||
                    (rawVal.startsWith("'") && rawVal.endsWith("'"))
                  ) {
                    val = rawVal.slice(1, -1);
                  }
                  const isSensitive =
                    /(TOKEN|KEY|SECRET|PASSWORD|PASS|AUTH|PRIVATE|CREDENTIAL|APIKEY|ACCESS|CERT|SALT|HASH|SIGNATURE)/i.test(
                      k
                    );
                  if (k) {
                    list.push({
                      name: k,
                      value: val,
                      scope,
                      is_sensitive: isSensitive,
                    });
                  }
                }
              }
              return list.sort((a, b) => a.name.localeCompare(b.name));
            };

            const envItems = parseEnvFromToml(targetPath);
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ env: envItems, total: envItems.length, scope }));
            return;
          }

          if (req.url?.startsWith("/api/env") && (req.method === "POST" || req.method === "PUT")) {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const body = JSON.parse(bodyStr);
                const rawKey = (body.key || "").trim();
                const originalKey = (body.original_key || "").trim();
                const value = String(body.value ?? "");
                const scope = (body.scope || "workspace").toLowerCase();

                if (!rawKey) {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "Environment variable key is required" }));
                  return;
                }

                const upperKey = rawKey.toUpperCase();
                const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
                const wsPath = path.join(rootDir, ".unirtm.toml");
                const targetPath = scope === "user" ? userPath : wsPath;

                const targetDir = path.dirname(targetPath);
                if (!fs.existsSync(targetDir)) {
                  fs.mkdirSync(targetDir, { recursive: true });
                }

                let content = fs.existsSync(targetPath) ? fs.readFileSync(targetPath, "utf-8") : "";
                let lines = content.split("\n");

                const quotedVal = JSON.stringify(value);
                const newEntry = `${upperKey} = ${quotedVal}`;

                let inEnv = false;
                let envStart = -1;
                let envEnd = -1;
                let matchingIdx = -1;
                const toRemoveIndices: number[] = [];

                for (let i = 0; i < lines.length; i++) {
                  const tr = lines[i].trim();
                  if (tr === "[env]") {
                    inEnv = true;
                    envStart = i;
                    continue;
                  }
                  if (inEnv) {
                    if (tr.startsWith("[") && tr !== "[env]") {
                      envEnd = i;
                      inEnv = false;
                      break;
                    }
                    if (tr.startsWith("#") || !tr.includes("=")) continue;
                    const eqIdx = tr.indexOf("=");
                    const existingKey = tr
                      .slice(0, eqIdx)
                      .trim()
                      .replace(/^["']|["']$/g, "")
                      .toUpperCase();

                    if (originalKey && existingKey === originalKey.toUpperCase() && existingKey !== upperKey) {
                      toRemoveIndices.push(i);
                    } else if (existingKey === upperKey) {
                      if (matchingIdx === -1) {
                        matchingIdx = i;
                      } else {
                        toRemoveIndices.push(i);
                      }
                    }
                  }
                }
                if (inEnv && envEnd === -1) {
                  envEnd = lines.length;
                }

                if (toRemoveIndices.length > 0) {
                  lines = lines.filter((_, idx) => !toRemoveIndices.includes(idx));
                  envStart = -1;
                  envEnd = -1;
                  matchingIdx = -1;
                  inEnv = false;
                  for (let i = 0; i < lines.length; i++) {
                    const tr = lines[i].trim();
                    if (tr === "[env]") {
                      inEnv = true;
                      envStart = i;
                      continue;
                    }
                    if (inEnv) {
                      if (tr.startsWith("[") && tr !== "[env]") {
                        envEnd = i;
                        inEnv = false;
                        break;
                      }
                      if (tr.startsWith("#") || !tr.includes("=")) continue;
                      const eqIdx = tr.indexOf("=");
                      const existingKey = tr
                        .slice(0, eqIdx)
                        .trim()
                        .replace(/^["']|["']$/g, "")
                        .toUpperCase();
                      if (existingKey === upperKey && matchingIdx === -1) {
                        matchingIdx = i;
                      }
                    }
                  }
                  if (inEnv && envEnd === -1) {
                    envEnd = lines.length;
                  }
                }

                if (envStart === -1) {
                  if (content && !content.endsWith("\n")) {
                    content += "\n";
                  }
                  content += `\n[env]\n${newEntry}\n`;
                  fs.writeFileSync(targetPath, content, "utf-8");
                } else if (matchingIdx !== -1) {
                  lines[matchingIdx] = newEntry;
                  fs.writeFileSync(targetPath, lines.join("\n"), "utf-8");
                } else {
                  lines.splice(envEnd, 0, newEntry);
                  fs.writeFileSync(targetPath, lines.join("\n"), "utf-8");
                }

                const isSensitive =
                  /(TOKEN|KEY|SECRET|PASSWORD|PASS|AUTH|PRIVATE|CREDENTIAL|APIKEY|ACCESS|CERT|SALT|HASH|SIGNATURE)/i.test(
                    upperKey
                  );
                res.setHeader("Content-Type", "application/json");
                res.end(
                  JSON.stringify({
                    success: true,
                    env: {
                      name: upperKey,
                      value,
                      scope,
                      is_sensitive: isSensitive,
                    },
                  })
                );
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/env") && req.method === "DELETE") {
            let bodyStr = "";
            req.on("data", (c: any) => {
              bodyStr += c;
            });
            req.on("end", () => {
              try {
                const urlObj = new URL(req.url!, "http://localhost");
                let key = urlObj.searchParams.get("key") || "";
                let scope = urlObj.searchParams.get("scope") || "";

                if (!key && bodyStr) {
                  try {
                    const parsed = JSON.parse(bodyStr);
                    key = parsed.key || "";
                    scope = scope || parsed.scope || "";
                  } catch {}
                }

                key = key.trim().toUpperCase();
                scope = (scope || "workspace").toLowerCase();

                if (!key) {
                  res.statusCode = 400;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: "Environment variable key is required" }));
                  return;
                }

                const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
                const wsPath = path.join(rootDir, ".unirtm.toml");
                const targetPath = scope === "user" ? userPath : wsPath;

                if (fs.existsSync(targetPath)) {
                  const content = fs.readFileSync(targetPath, "utf-8");
                  const lines = content.split("\n");
                  let inEnv = false;
                  const newLines: string[] = [];

                  for (let i = 0; i < lines.length; i++) {
                    const tr = lines[i].trim();
                    if (tr === "[env]") {
                      inEnv = true;
                      newLines.push(lines[i]);
                      continue;
                    }
                    if (inEnv) {
                      if (tr.startsWith("[") && tr !== "[env]") {
                        inEnv = false;
                        newLines.push(lines[i]);
                        continue;
                      }
                      if (tr.startsWith("#") || !tr.includes("=")) {
                        newLines.push(lines[i]);
                        continue;
                      }
                      const eqIdx = tr.indexOf("=");
                      const existingKey = tr
                        .slice(0, eqIdx)
                        .trim()
                        .replace(/^["']|["']$/g, "")
                        .toUpperCase();
                      if (existingKey === key) {
                        continue;
                      }
                    }
                    newLines.push(lines[i]);
                  }
                  fs.writeFileSync(targetPath, newLines.join("\n"), "utf-8");
                }

                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ success: true, deleted_key: key }));
              } catch (err: any) {
                res.statusCode = 500;
                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({ error: err.message }));
              }
            });
            return;
          }

          if (req.url?.startsWith("/api/hooks")) {
            const url = new URL(req.url, "http://localhost");
            const pathname = url.pathname;
            const scope = (url.searchParams.get("scope") || "workspace").toLowerCase();
            const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
            const wsPath = path.join(rootDir, ".unirtm.toml");
            const targetPath = scope === "user" ? userPath : wsPath;

            const parseHooksTable = (raw: string) => {
              const res: Record<string, string> = {};
              let inTable = false;
              for (const line of raw.split("\n")) {
                const tr = line.trim();
                if (tr.startsWith("[") && tr.endsWith("]")) {
                  inTable = tr === "[hooks]";
                  continue;
                }
                if (inTable && tr && !tr.startsWith("#")) {
                  const eq = tr.indexOf("=");
                  if (eq !== -1) {
                    const k = tr
                      .slice(0, eq)
                      .trim()
                      .replace(/^["']|["']$/g, "");
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

            if (pathname === "/api/hooks" && req.method === "GET") {
              const raw = fs.existsSync(targetPath) ? fs.readFileSync(targetPath, "utf-8") : "";
              const hooksMap = parseHooksTable(raw);
              const gitDir = path.join(rootDir, ".git");
              const gitDirExists = fs.existsSync(gitDir) && fs.statSync(gitDir).isDirectory();

              const hooksList = Object.keys(hooksMap)
                .sort()
                .map((name) => {
                  let installed = false;
                  if (scope !== "user" && gitDirExists) {
                    const hookFile = path.join(gitDir, "hooks", name);
                    if (fs.existsSync(hookFile)) {
                      const content = fs.readFileSync(hookFile, "utf-8");
                      installed = content.includes("unirtm hook");
                    }
                  }
                  return {
                    name,
                    command: hooksMap[name],
                    installed,
                  };
                });

              res.setHeader("Content-Type", "application/json");
              res.end(
                JSON.stringify({
                  hooks: hooksList,
                  total: hooksList.length,
                  scope,
                  git_dir_exists: gitDirExists,
                })
              );
              return;
            }

            if (pathname === "/api/hooks" && (req.method === "POST" || req.method === "PUT")) {
              let bodyStr = "";
              req.on("data", (c) => {
                bodyStr += c;
              });
              req.on("end", () => {
                try {
                  const { name, command, scope: reqScope, install_bridge } = JSON.parse(bodyStr || "{}");
                  if (!name || !command) {
                    res.statusCode = 400;
                    res.setHeader("Content-Type", "application/json");
                    res.end(JSON.stringify({ error: "name and command are required" }));
                    return;
                  }
                  const curScope = (reqScope || scope || "workspace").toLowerCase();
                  const curTarget = curScope === "user" ? userPath : wsPath;
                  fs.mkdirSync(path.dirname(curTarget), { recursive: true });
                  const raw = fs.existsSync(curTarget) ? fs.readFileSync(curTarget, "utf-8") : "";

                  const lines = raw.split("\n");
                  let inHooks = false;
                  let hooksStart = -1;
                  let hooksEnd = -1;
                  let matchIdx = -1;

                  for (let i = 0; i < lines.length; i++) {
                    const tr = lines[i].trim();
                    if (tr === "[hooks]") {
                      inHooks = true;
                      hooksStart = i;
                      continue;
                    }
                    if (inHooks) {
                      if (tr.startsWith("[") && tr !== "[hooks]") {
                        hooksEnd = i;
                        inHooks = false;
                        break;
                      }
                      const eq = tr.indexOf("=");
                      if (eq !== -1) {
                        const k = tr
                          .slice(0, eq)
                          .trim()
                          .replace(/^["']|["']$/g, "");
                        if (k.toLowerCase() === name.toLowerCase()) {
                          matchIdx = i;
                        }
                      }
                    }
                  }
                  if (inHooks) hooksEnd = lines.length;

                  const newEntry = `${name} = ${JSON.stringify(command)}`;
                  if (hooksStart === -1) {
                    const appendStr = (raw && !raw.endsWith("\n") ? "\n" : "") + `\n[hooks]\n${newEntry}\n`;
                    fs.writeFileSync(curTarget, raw + appendStr, "utf-8");
                  } else if (matchIdx !== -1) {
                    lines[matchIdx] = newEntry;
                    fs.writeFileSync(curTarget, lines.join("\n"), "utf-8");
                  } else {
                    lines.splice(hooksEnd, 0, newEntry);
                    fs.writeFileSync(curTarget, lines.join("\n"), "utf-8");
                  }

                  let installed = false;
                  if (install_bridge && curScope !== "user") {
                    const unirtmBin = findUnirtmBinary(rootDir);
                    if (fs.existsSync(unirtmBin)) {
                      try {
                        execFileSync(unirtmBin, ["hook", "install", name], { cwd: rootDir, stdio: "pipe" });
                        installed = true;
                      } catch {}
                    }
                  }

                  res.setHeader("Content-Type", "application/json");
                  res.end(
                    JSON.stringify({
                      success: true,
                      name,
                      command,
                      scope: curScope,
                      installed,
                    })
                  );
                } catch (err: any) {
                  res.statusCode = 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }

            if (pathname === "/api/hooks" && req.method === "DELETE") {
              let bodyStr = "";
              req.on("data", (c) => {
                bodyStr += c;
              });
              req.on("end", () => {
                try {
                  let hookName = url.searchParams.get("name") || "";
                  let delScope = scope;
                  if (!hookName && bodyStr) {
                    try {
                      const parsed = JSON.parse(bodyStr);
                      hookName = parsed.name || "";
                      delScope = parsed.scope || delScope;
                    } catch {}
                  }
                  if (!hookName) {
                    res.statusCode = 400;
                    res.setHeader("Content-Type", "application/json");
                    res.end(JSON.stringify({ error: "hook name is required" }));
                    return;
                  }
                  const curTarget = delScope === "user" ? userPath : wsPath;
                  if (fs.existsSync(curTarget)) {
                    const content = fs.readFileSync(curTarget, "utf-8");
                    const lines = content.split("\n");
                    let inHooks = false;
                    const newLines: string[] = [];
                    for (let i = 0; i < lines.length; i++) {
                      const tr = lines[i].trim();
                      if (tr === "[hooks]") {
                        inHooks = true;
                        newLines.push(lines[i]);
                        continue;
                      }
                      if (inHooks) {
                        if (tr.startsWith("[") && tr !== "[hooks]") {
                          inHooks = false;
                          newLines.push(lines[i]);
                          continue;
                        }
                        const eq = tr.indexOf("=");
                        if (eq !== -1) {
                          const k = tr
                            .slice(0, eq)
                            .trim()
                            .replace(/^["']|["']$/g, "");
                          if (k.toLowerCase() === hookName.toLowerCase()) {
                            continue;
                          }
                        }
                      }
                      newLines.push(lines[i]);
                    }
                    fs.writeFileSync(curTarget, newLines.join("\n"), "utf-8");
                  }
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ success: true, deleted_name: hookName }));
                } catch (err: any) {
                  res.statusCode = 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }

            if (pathname === "/api/hooks/install" && req.method === "POST") {
              let bodyStr = "";
              req.on("data", (c) => {
                bodyStr += c;
              });
              req.on("end", () => {
                try {
                  const { name, all } = JSON.parse(bodyStr || "{}");
                  const unirtmBin = findUnirtmBinary(rootDir);
                  const args = ["hook", "install"];
                  if (all || !name) {
                    args.push("-a");
                  } else {
                    args.push(name);
                  }
                  let output = "";
                  if (fs.existsSync(unirtmBin)) {
                    output = execFileSync(unirtmBin, args, { cwd: rootDir, encoding: "utf-8" });
                  }
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ success: true, installed: name ? [name] : ["all"], output }));
                } catch (err: any) {
                  res.statusCode = 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }

            if (pathname === "/api/hooks/run" && req.method === "POST") {
              let bodyStr = "";
              req.on("data", (c) => {
                bodyStr += c;
              });
              req.on("end", () => {
                try {
                  const { name, args: hArgs } = JSON.parse(bodyStr || "{}");
                  const unirtmBin = findUnirtmBinary(rootDir);
                  const runArgs = ["hook", "run", name || "all"];
                  if (Array.isArray(hArgs)) {
                    runArgs.push(...hArgs);
                  }
                  let output = "";
                  let exitCode = 0;
                  try {
                    output = execFileSync(unirtmBin, runArgs, { cwd: rootDir, encoding: "utf-8" });
                  } catch (runErr: any) {
                    exitCode = runErr.status || 1;
                    output = runErr.stdout || runErr.stderr || runErr.message;
                  }
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ success: exitCode === 0, exit_code: exitCode, output, name }));
                } catch (err: any) {
                  res.statusCode = 500;
                  res.setHeader("Content-Type", "application/json");
                  res.end(JSON.stringify({ error: err.message }));
                }
              });
              return;
            }
          }

          if (req.url === "/api/config") {
            const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
            const wsPath = path.join(rootDir, ".unirtm.toml");

            const parseTable = (raw: string, tableName: string) => {
              const res: Record<string, any> = {};
              let inTable = false;
              for (const line of raw.split("\n")) {
                const tr = line.trim();
                if (tr.startsWith("[") && tr.endsWith("]")) {
                  inTable = tr === `[${tableName}]`;
                  continue;
                }
                if (inTable && tr && !tr.startsWith("#")) {
                  const eq = tr.indexOf("=");
                  if (eq !== -1) {
                    const k = tr
                      .slice(0, eq)
                      .trim()
                      .replace(/^["']|["']$/g, "");
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

            const parseSettings = (raw: string) => parseTable(raw, "settings");

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
                    hooks: parseTable(userRaw, "hooks"),
                  },
                  workspace: {
                    path: wsPath,
                    exists: wsExists,
                    raw: wsRaw,
                    settings: parseSettings(wsRaw),
                    hooks: parseTable(wsRaw, "hooks"),
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
                  const unirtmBin = findUnirtmBinary(rootDir);
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
