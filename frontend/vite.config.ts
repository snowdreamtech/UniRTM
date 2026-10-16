import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { execFileSync, spawnSync } from "node:child_process";

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
                  inTools = (tr === "[tools]");
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
                  inPins = (tr === "[pins]");
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

          if ((req.url === "/api/tools/migrate-legacy" || req.url === "/api/runtimes/migrate-legacy") && req.method === "POST") {
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
                const out = execFileSync(unirtmBin, ["list", "--json"], { cwd: rootDir, encoding: "utf-8", timeout: 5000 });
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
                    const list = fs.readdirSync(dir).filter(f => !f.startsWith(".") && fs.statSync(path.join(dir, f)).isDirectory());
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

            const getToolItem = (id: string, name: string, desc: string, hp: string, defaultVer: string, fallbackInstalled: string[]) => {
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

            const knownLanguageRuntimes: Record<string, { name: string; desc: string; hp: string; defaultVer: string }> = {
              node: { name: "Node.js", desc: "JavaScript / TypeScript runtime (NVM superset)", hp: "https://nodejs.org", defaultVer: "20.18.3" },
              go: { name: "Go", desc: "Go programming language (GVM superset)", hp: "https://go.dev", defaultVer: "1.27.2" },
              python: { name: "Python", desc: "Python language interpreter (Pyenv superset)", hp: "https://python.org", defaultVer: "3.14.7" },
              rust: { name: "Rust", desc: "Rust toolchain & cargo build system", hp: "https://rust-lang.org", defaultVer: "1.85.0" },
              java: { name: "Java", desc: "OpenJDK / Java runtime environment", hp: "https://openjdk.org", defaultVer: "21.0.2" },
              ruby: { name: "Ruby", desc: "Dynamic open source programming language with a focus on simplicity", hp: "https://ruby-lang.org", defaultVer: "3.4.1" },
              bun: { name: "Bun", desc: "Incredibly fast all-in-one JavaScript & TypeScript toolkit", hp: "https://bun.sh", defaultVer: "1.2.4" },
              deno: { name: "Deno", desc: "Next-generation JavaScript and TypeScript runtime", hp: "https://deno.com", defaultVer: "2.1.9" },
              zig: { name: "Zig", desc: "General-purpose programming language and toolchain", hp: "https://ziglang.org", defaultVer: "0.13.0" },
              php: { name: "PHP", desc: "Popular general-purpose scripting language suited for web development", hp: "https://php.net", defaultVer: "8.4.3" },
              dotnet: { name: "Dotnet", desc: ".NET Developer Platform and SDK", hp: "https://dotnet.microsoft.com", defaultVer: "9.0.100" },
              elixir: { name: "Elixir", desc: "Dynamic functional language designed for building scalable applications", hp: "https://elixir-lang.org", defaultVer: "1.18.1" },
              erlang: { name: "Erlang", desc: "Programming language used to build massively scalable soft real-time systems", hp: "https://www.erlang.org", defaultVer: "27.2.1" },
              flutter: { name: "Flutter", desc: "UI software development kit by Google", hp: "https://flutter.dev", defaultVer: "3.27.3" },
              lua: { name: "Lua", desc: "Powerful, efficient, lightweight, embeddable scripting language", hp: "https://www.lua.org", defaultVer: "5.4.7" },
              swift: { name: "Swift", desc: "General-purpose programming language developed by Apple", hp: "https://swift.org", defaultVer: "6.0.3" },
            };

            // 1. Collect all tool IDs that are actually configured (workspace/user) or active via legacy files
            const activeLegacyIds: string[] = [];
            if (fs.existsSync(nvmrc) || fs.existsSync(nodeVer)) activeLegacyIds.push("node");
            if (fs.existsSync(goVer)) activeLegacyIds.push("go");
            if (fs.existsSync(pyVer)) activeLegacyIds.push("python");

            const configuredIds = Array.from(
              new Set([
                ...Object.keys(wsTools),
                ...Object.keys(userTools),
                ...activeLegacyIds,
              ])
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
            req.on("data", (c: any) => { bodyStr += c; });
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
                        const toolRegex = new RegExp(`(\\[\\[tools\\."${tool}"\\]\\][\\s\\S]*?version\\s*=\\s*)"[^"]+"`);
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
            req.on("data", (c: any) => { bodyStr += c; });
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
            req.on("data", (c: any) => { bodyStr += c; });
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

            const allTools = Array.from(new Set(["node", "go", "python", "rust", ...Object.keys(wsTools), ...Object.keys(userTools)]));
            const urlObj = new URL(req.url, "http://localhost");
            const filterTool = urlObj.searchParams.get("tool");
            const targetTools = filterTool ? [filterTool] : allTools;

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
            };

            const semverGt = (a: string, b: string) => {
              const pa = a.replace(/^[vV]+/, "").split(".").map(n => parseInt(n, 10));
              const pb = b.replace(/^[vV]+/, "").split(".").map(n => parseInt(n, 10));
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

              const candKeys = [toolKey, bare, `github:${toolKey}`, `github:${bare}`];
              for (const k of candKeys) {
                if (lockTools[k]) {
                  for (const v of lockTools[k]) {
                    if (!candidates.includes(v)) candidates.push(v);
                  }
                }
              }

              if (candidates.length === 0) {
                return curVer;
              }

              candidates.sort((a, b) => (semverGt(a, b) ? -1 : semverGt(b, a) ? 1 : 0));
              return candidates[0] || curVer;
            };

            const updates: any[] = [];
            let upgradableCount = 0;
            let pinnedCount = 0;

            for (const t of targetTools) {
              const cur = wsTools[t] || userTools[t] || "";
              const pin = wsPins[t] || userPins[t] || { pinned: false, reason: "" };
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
                latest_version: hasUpdate ? latest : (cur || latest || ""),
                has_update: hasUpdate,
                is_pinned: pin.pinned,
                pin_reason: pin.reason,
              });
            }

            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({
              updates,
              total_checked: updates.length,
              upgradable_count: upgradableCount,
              pinned_count: pinnedCount,
            }));
            return;
          }

          if ((req.url === "/api/tools/batch-update" || req.url === "/api/runtimes/batch-update") && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
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
                res.end(JSON.stringify({
                  success: true,
                  updated,
                  skipped,
                  updated_count: updated.length,
                  skipped_count: skipped.length,
                  file: targetPath,
                }));
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
              try {
                const res = spawnSync(unirtmBin, ["ls-remote", tool, "--json"], {
                  cwd: rootDir,
                  encoding: "utf-8",
                  maxBuffer: 50 * 1024 * 1024,
                  timeout: 10000,
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
                    versions = Array.from(new Set(extracted));
                  }
                }
              } catch (e) {
                console.warn(`[UniRTM DevServer] Dynamic ls-remote fallback for ${tool}:`, e);
              }
            }

            // 2. Static fallback only if CLI or upstream network fails
            if (versions.length === 0) {
              if (tool === "node" || tool === "nodejs") {
              versions = [
                "23.9.0", "23.8.0", "23.7.0", "23.6.0", "23.0.0",
                "22.14.0", "22.13.1", "22.13.0", "22.12.0", "22.11.0", "22.9.0", "22.0.0",
                "21.7.3", "21.6.2", "21.1.0",
                "20.18.3", "20.18.2", "20.18.0", "20.17.0", "20.16.0", "20.15.1", "20.14.0", "20.12.2", "20.11.1", "20.10.0", "20.9.0",
                "19.9.0", "19.8.1",
                "18.20.7", "18.20.6", "18.20.5", "18.20.4", "18.20.0", "18.19.1", "18.18.2", "18.17.1", "18.16.0", "18.14.2", "18.12.1",
                "16.20.2", "16.20.0", "16.19.1", "16.18.0", "16.14.2",
                "14.21.3", "14.20.1", "14.19.3",
              ];
            } else if (tool === "python" || tool === "python3") {
              versions = [
                "3.14.0a4",
                "3.13.2", "3.13.1", "3.13.0",
                "3.12.9", "3.12.8", "3.12.7", "3.12.6", "3.12.5", "3.12.4", "3.12.3", "3.12.2", "3.12.1", "3.12.0",
                "3.11.11", "3.11.10", "3.11.9", "3.11.8", "3.11.7", "3.11.6", "3.11.5", "3.11.4", "3.11.0",
                "3.10.16", "3.10.15", "3.10.14", "3.10.13", "3.10.12", "3.10.11", "3.10.8", "3.10.0",
                "3.9.21", "3.9.20", "3.9.19", "3.9.18", "3.9.16", "3.9.13", "3.9.0",
                "3.8.20", "3.8.18", "3.8.16", "3.8.12", "3.8.0",
                "3.7.17", "3.7.12",
              ];
            } else if (tool === "go" || tool === "golang") {
              versions = [
                "1.27.2", "1.27.1", "1.27.0",
                "1.26.5", "1.26.4", "1.26.3", "1.26.2", "1.26.1", "1.26.0",
                "1.25.8", "1.25.7", "1.25.6", "1.25.5", "1.25.4", "1.25.3", "1.25.2", "1.25.1", "1.25.0",
                "1.24.5", "1.24.4", "1.24.3", "1.24.2", "1.24.1", "1.24.0",
                "1.23.6", "1.23.5", "1.23.4", "1.23.3", "1.23.2", "1.23.1", "1.23.0",
                "1.22.12", "1.22.11", "1.22.10", "1.22.9", "1.22.8", "1.22.6", "1.22.5", "1.22.2", "1.22.0",
                "1.21.13", "1.21.12", "1.21.11", "1.21.8", "1.21.6", "1.21.0",
                "1.20.14", "1.20.12", "1.20.7", "1.20.0",
                "1.19.13", "1.19.8", "1.19.0",
                "1.18.10", "1.18.5",
              ];
            } else if (tool === "rust") {
              versions = [
                "1.85.0", "1.84.1", "1.84.0", "1.83.0", "1.82.0", "1.81.0",
                "1.80.1", "1.80.0", "1.79.0", "1.78.0", "1.77.2", "1.77.0",
                "1.76.0", "1.75.0", "1.74.1", "1.73.0", "1.72.1", "1.71.1", "1.70.0",
                "1.69.0", "1.68.2", "1.67.1", "1.66.1", "1.65.0",
              ];
            } else if (tool === "java" || tool === "openjdk") {
              versions = [
                "23.0.2", "23.0.1",
                "22.0.2", "22.0.1",
                "21.0.6", "21.0.5", "21.0.4", "21.0.2", "21.0.1", "21.0.0",
                "17.0.14", "17.0.12", "17.0.10", "17.0.9", "17.0.8", "17.0.2",
                "11.0.26", "11.0.22", "11.0.20", "11.0.18",
                "8.0.442", "8.0.402", "8.0.392", "8.0.352",
              ];
            } else if (tool === "bun") {
              versions = [
                "1.2.4", "1.2.3", "1.2.2", "1.2.1", "1.2.0",
                "1.1.45", "1.1.42", "1.1.38", "1.1.34", "1.1.30", "1.1.25", "1.1.20", "1.1.10", "1.1.0",
                "1.0.36", "1.0.35", "1.0.30", "1.0.25", "1.0.20", "1.0.18", "1.0.0",
              ];
            } else if (tool === "deno") {
              versions = [
                "2.2.3", "2.2.2", "2.2.1", "2.2.0",
                "2.1.9", "2.1.4", "2.1.0",
                "2.0.6", "2.0.4", "2.0.0",
                "1.46.3", "1.45.5", "1.44.4", "1.43.6", "1.42.4", "1.41.2", "1.40.5", "1.39.4", "1.38.2",
              ];
            } else if (tool === "zig") {
              versions = [
                "0.14.0", "0.13.0", "0.12.1", "0.12.0", "0.11.0", "0.10.1", "0.9.1", "0.8.1",
              ];
            } else if (tool === "ruby") {
              versions = [
                "3.4.2", "3.4.1",
                "3.3.7", "3.3.6", "3.3.5", "3.3.0",
                "3.2.7", "3.2.6", "3.2.3", "3.2.0",
                "3.1.6", "3.1.4", "3.1.2",
                "3.0.7", "3.0.6",
              ];
            } else if (tool === "php") {
              versions = [
                "8.4.4", "8.4.3", "8.4.1",
                "8.3.17", "8.3.16", "8.3.14", "8.3.10", "8.3.0",
                "8.2.27", "8.2.26", "8.2.20", "8.2.15",
                "8.1.31", "8.1.28", "8.1.20",
                "8.0.30", "8.0.28",
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
                        const list = fs.readdirSync(dir).filter(f => !f.startsWith(".") && fs.statSync(path.join(dir, f)).isDirectory());
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
                    message: "代理服务器连通正常"
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

          // --- Themes & Marketplace API for Dev Server ---
          if (req.url?.startsWith("/api/themes/marketplace/search")) {
            const urlObj = new URL(req.url, "http://localhost");
            const q = urlObj.searchParams.get("q") || "";
            const size = urlObj.searchParams.get("size") || "24";
            const offset = urlObj.searchParams.get("offset") || "0";
            const targetURL = `https://open-vsx.org/api/-/search?category=Themes&query=${encodeURIComponent(q)}&size=${size}&offset=${offset}`;

            fetch(targetURL, {
              headers: { "User-Agent": "UniRTM-Workbench-Dev/0.1.0" }
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
                res.end(JSON.stringify({
                  offset: data.offset || 0,
                  totalSize: data.totalSize || 0,
                  extensions,
                }));
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
              headers: { "User-Agent": "UniRTM-Workbench-Dev/0.1.0" }
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
                      headers: { "User-Agent": "UniRTM-Workbench-Dev/0.1.0" }
                    });
                    if (rRes.ok) readme = await rRes.text();
                  } catch {}
                }

                const contributes: { themes: any[]; iconThemes: any[] } = { themes: [], iconThemes: [] };
                if (data.files?.manifest) {
                  try {
                    const mRes = await fetch(data.files.manifest, {
                      headers: { "User-Agent": "UniRTM-Workbench-Dev/0.1.0" }
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
                res.end(JSON.stringify({
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
                }));
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
            req.on("data", (c: any) => { bodyStr += c; });
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
print(json.dumps({'name': name, 'version': pkg.get('version', '0.0.1'), 'themes': pkg.get('contributes', {}).get('themes', []), 'iconThemes': pkg.get('contributes', {}).get('iconThemes', [])}))
`;
                const outStr = execFileSync("python3", ["-c", script, tmpZip], { encoding: "utf-8" });
                try { fs.unlinkSync(tmpZip); } catch {}
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
            if (parts.length >= 3) {
              const packageId = decodeURIComponent(parts[1]);
              const subpath = parts.slice(2).map(decodeURIComponent).join("/");
              const themesDir = path.join(os.homedir(), ".config", "unirtm", "themes");
              const targetFile = path.join(themesDir, packageId, subpath);
              if (fs.existsSync(targetFile) && fs.statSync(targetFile).isFile()) {
                if (targetFile.endsWith(".json")) res.setHeader("Content-Type", "application/json");
                else if (targetFile.endsWith(".svg")) res.setHeader("Content-Type", "image/svg+xml");
                else if (targetFile.endsWith(".woff2")) res.setHeader("Content-Type", "font/woff2");
                else if (targetFile.endsWith(".woff")) res.setHeader("Content-Type", "font/woff");
                else if (targetFile.endsWith(".png")) res.setHeader("Content-Type", "image/png");
                res.end(fs.readFileSync(targetFile));
                return;
              }
            }
            res.statusCode = 404;
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ error: "Theme asset not found" }));
            return;
          }

          if ((req.url === "/api/themes" || req.url?.startsWith("/api/themes?")) && req.method === "GET") {
            const themesDir = path.join(os.homedir(), ".config", "unirtm", "themes");
            const colorThemes: any[] = [
              { id: "default", name: "UniRTM (默认)", type: "color", is_builtin: true, ui_theme: "vs-dark" },
              { id: "dark", name: "UniRTM Dark", type: "color", is_builtin: true, ui_theme: "vs-dark" },
              { id: "light", name: "UniRTM Light", type: "color", is_builtin: true, ui_theme: "vs" },
            ];
            const iconThemes: any[] = [
              { id: "seti", name: "Seti UI (VS Code Default)", type: "icon", is_builtin: true, description: "Official Visual Studio Code default file icon theme" },
            ];

            if (fs.existsSync(themesDir)) {
              try {
                const dirs = fs.readdirSync(themesDir, { withFileTypes: true });
                for (const d of dirs) {
                  if (d.isDirectory()) {
                    const pkgJson = path.join(themesDir, d.name, "package.json");
                    if (fs.existsSync(pkgJson)) {
                      try {
                        const p = JSON.parse(fs.readFileSync(pkgJson, "utf-8"));
                        const c = p.contributes || {};
                        for (const ct of c.themes || []) {
                          colorThemes.push({
                            id: ct.id || ct.label,
                            name: ct.label || ct.id,
                            type: "color",
                            is_builtin: false,
                            ui_theme: ct.uiTheme || "vs-dark",
                            package_id: p.name,
                            version: p.version,
                            relative_path: ct.path,
                          });
                        }
                        for (const it of c.iconThemes || []) {
                          iconThemes.push({
                            id: it.id || it.label,
                            name: it.label || it.id,
                            type: "icon",
                            is_builtin: false,
                            package_id: p.name,
                            version: p.version,
                            relative_path: it.path,
                          });
                        }
                      } catch {}
                    }
                  }
                }
              } catch {}
            }

            const userConfigDir = path.join(os.homedir(), ".config", "unirtm");
            const userPath = path.join(userConfigDir, "unirtm.toml");
            let activeColor = "default";
            let activeIcon = "seti";
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
            };
            readSetting(userPath);

            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({
              active: {
                color: activeColor,
                icon: activeIcon,
                appearance: activeAppearance,
              },
              active_color_theme: activeColor,
              active_icon_theme: activeIcon,
              colors: colorThemes,
              icons: iconThemes,
              color_themes: colorThemes,
              icon_themes: iconThemes,
            }));
            return;
          }

          if (req.url?.startsWith("/api/themes") && req.method === "DELETE") {
            const urlObj = new URL(req.url, "http://localhost");
            const id = urlObj.searchParams.get("id");
            const themesDir = path.join(os.homedir(), ".config", "unirtm", "themes");
            if (id && fs.existsSync(themesDir)) {
              try {
                for (const d of fs.readdirSync(themesDir)) {
                  const itemDir = path.join(themesDir, d);
                  const pkgJson = path.join(itemDir, "package.json");
                  if (fs.existsSync(pkgJson)) {
                    try {
                      const p = JSON.parse(fs.readFileSync(pkgJson, "utf-8"));
                      const c = p.contributes || {};
                      const hasTheme = (c.themes || []).some((t: any) => (t.id || t.label) === id) ||
                                       (c.iconThemes || []).some((t: any) => (t.id || t.label) === id) ||
                                       p.name === id;
                      if (hasTheme) {
                        fs.rmSync(itemDir, { recursive: true, force: true });
                        break;
                      }
                    } catch {}
                  }
                }
              } catch {}
            }
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ success: true }));
            return;
          }

          if (req.url === "/api/themes/apply" && req.method === "POST") {
            let bodyStr = "";
            req.on("data", (c: any) => { bodyStr += c; });
            req.on("end", () => {
              try {
                const body = JSON.parse(bodyStr || "{}");
                const color_theme_id = body.color || body.color_theme_id;
                const icon_theme_id = body.icon || body.icon_theme_id;
                const appearance = body.appearance;
                const userConfigDir = path.join(os.homedir(), ".config", "unirtm");
                if (!fs.existsSync(userConfigDir)) {
                  fs.mkdirSync(userConfigDir, { recursive: true });
                }
                const userPath = path.join(userConfigDir, "unirtm.toml");
                let content = fs.existsSync(userPath) ? fs.readFileSync(userPath, "utf-8") : "";

                if (color_theme_id) {
                  if (content.match(/(?:color_)?theme\s*=/)) {
                    content = content.replace(/(?:color_)?theme\s*=\s*"[^"]*"/, `theme = "${color_theme_id}"`);
                  } else {
                    content += `\n[settings]\ntheme = "${color_theme_id}"\n`;
                  }
                }
                if (icon_theme_id) {
                  if (content.match(/icon_theme\s*=/)) {
                    content = content.replace(/icon_theme\s*=\s*"[^"]*"/, `icon_theme = "${icon_theme_id}"`);
                  } else {
                    if (!content.includes("[settings]")) {
                      content += `\n[settings]\n`;
                    }
                    content += `icon_theme = "${icon_theme_id}"\n`;
                  }
                }
                if (appearance) {
                  if (content.match(/appearance\s*=/)) {
                    content = content.replace(/appearance\s*=\s*"[^"]*"/, `appearance = "${appearance}"`);
                  } else {
                    if (!content.includes("[settings]")) {
                      content += `\n[settings]\n`;
                    }
                    content += `appearance = "${appearance}"\n`;
                  }
                }
                fs.writeFileSync(userPath, content, "utf-8");

                res.setHeader("Content-Type", "application/json");
                res.end(JSON.stringify({
                  success: true,
                  color: color_theme_id,
                  icon: icon_theme_id,
                  appearance,
                }));
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
