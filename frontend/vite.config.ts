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

          if (req.url === "/api/runtimes/migrate-legacy" && req.method === "POST") {
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

          if (req.url === "/api/runtimes/status") {
            const wsPath = path.join(rootDir, ".unirtm.toml");
            const altWsPath = path.join(rootDir, "unirtm.toml");
            const userPath = path.join(os.homedir(), ".config", "unirtm", "unirtm.toml");
            const altUserPath = path.join(os.homedir(), ".config", "unirtm", "config.toml");

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
                    const m = tr.match(/^"?([^"=\s]+)"?\s*=\s*"([^"]+)"/);
                    if (m) res[m[1]] = m[2];
                  }
                }
              } catch {}
              return res;
            };

            const getInstalled = (toolId: string, fallback: string[]) => {
              const dir = path.join(os.homedir(), ".local", "share", "unirtm", "installs", toolId);
              if (!fs.existsSync(dir)) return fallback;
              try {
                const list = fs.readdirSync(dir).filter(f => !f.startsWith(".") && fs.statSync(path.join(dir, f)).isDirectory());
                if (list.length === 0) return fallback;
                return list.sort((a, b) => b.localeCompare(a, undefined, { numeric: true }));
              } catch {
                return fallback;
              }
            };

            const wsTools = parseTools(fs.existsSync(wsPath) ? wsPath : altWsPath);
            const userTools = parseTools(fs.existsSync(userPath) ? userPath : altUserPath);

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
              };
            };

            const knownLanguageRuntimes: Record<string, { name: string; desc: string; hp: string; defaultVer: string }> = {
              node: { name: "Node.js", desc: "JavaScript / TypeScript runtime (NVM superset)", hp: "https://nodejs.org", defaultVer: "20.18.3" },
              go: { name: "Go", desc: "Go programming language (GVM superset)", hp: "https://go.dev", defaultVer: "1.27.2" },
              python: { name: "Python", desc: "Python language interpreter (Pyenv superset)", hp: "https://python.org", defaultVer: "3.14.7" },
              rust: { name: "Rust", desc: "Rust toolchain & cargo build system", hp: "https://rust-lang.org", defaultVer: "1.85.0" },
              bun: { name: "Bun", desc: "Incredibly fast all-in-one JavaScript & TypeScript toolkit", hp: "https://bun.sh", defaultVer: "1.2.4" },
              deno: { name: "Deno", desc: "Next-generation JavaScript and TypeScript runtime", hp: "https://deno.com", defaultVer: "2.1.9" },
              java: { name: "Java", desc: "OpenJDK / Java runtime environment", hp: "https://openjdk.org", defaultVer: "21.0.2" },
              zig: { name: "Zig", desc: "General-purpose programming language and toolchain", hp: "https://ziglang.org", defaultVer: "0.13.0" },
              ruby: { name: "Ruby", desc: "Dynamic open source programming language with a focus on simplicity", hp: "https://ruby-lang.org", defaultVer: "3.4.1" },
              php: { name: "PHP", desc: "Popular general-purpose scripting language suited for web development", hp: "https://php.net", defaultVer: "8.4.3" },
              dotnet: { name: "Dotnet", desc: ".NET Developer Platform and SDK", hp: "https://dotnet.microsoft.com", defaultVer: "9.0.100" },
              elixir: { name: "Elixir", desc: "Dynamic functional language designed for building scalable applications", hp: "https://elixir-lang.org", defaultVer: "1.18.1" },
              erlang: { name: "Erlang", desc: "Programming language used to build massively scalable soft real-time systems", hp: "https://www.erlang.org", defaultVer: "27.2.1" },
              flutter: { name: "Flutter", desc: "UI software development kit by Google", hp: "https://flutter.dev", defaultVer: "3.27.3" },
              lua: { name: "Lua", desc: "Powerful, efficient, lightweight, embeddable scripting language", hp: "https://www.lua.org", defaultVer: "5.4.7" },
              swift: { name: "Swift", desc: "General-purpose programming language developed by Apple", hp: "https://swift.org", defaultVer: "6.0.3" },
            };

            // Only recognize programming language environments — filter out ecosystem/dev tools like github:*, npm:*, pipx:*, go:*, tools with slash, etc.
            const isLanguageRuntime = (id: string) => {
              if (id.includes(":") || id.includes("/")) return false;
              return id in knownLanguageRuntimes;
            };

            const coreLangIds = ["node", "go", "python", "rust"];
            const dynamicLangIds = [...Object.keys(wsTools), ...Object.keys(userTools)].filter(isLanguageRuntime);
            const allToolIds = new Set([...coreLangIds, ...dynamicLangIds]);
            const runtimes: any[] = [];
            for (const id of allToolIds) {
              const meta = knownLanguageRuntimes[id] || {
                name: id.charAt(0).toUpperCase() + id.slice(1),
                desc: `${id} runtime environment`,
                hp: "",
                defaultVer: "",
              };
              runtimes.push(getToolItem(id, meta.name, meta.desc, meta.hp, meta.defaultVer, []));
            }

            const totalInstalled = runtimes.reduce((acc, r) => acc + r.installed_versions.length, 0);

            res.setHeader("Content-Type", "application/json");
            res.end(
              JSON.stringify({
                runtimes,
                total_installed: totalInstalled,
                detected_legacies,
              })
            );
            return;
          }

          if (req.url === "/api/runtimes/use" && req.method === "POST") {
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

          if (req.url === "/api/runtimes/remove" && req.method === "POST") {
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

          if (req.url?.startsWith("/api/runtimes/remote")) {
            const urlObj = new URL(req.url, "http://localhost");
            const tool = urlObj.searchParams.get("tool");
            let versions: string[] = [];
            if (tool === "node") {
              versions = ["26.8.1", "22.14.0", "22.13.1", "20.18.3", "20.18.2", "18.20.7", "18.20.6"];
            } else if (tool === "go") {
              versions = ["1.27.2", "1.27.1", "1.24.0", "1.23.6", "1.23.5", "1.22.12"];
            } else if (tool === "python") {
              versions = ["3.14.7", "3.13.2", "3.12.9", "3.12.8", "3.12.2", "3.11.11", "3.10.16"];
            } else if (tool === "rust") {
              versions = ["1.85.0", "1.84.1", "1.83.0", "1.82.0", "1.81.0"];
            } else if (tool === "java") {
              versions = ["21.0.2", "21.0.1", "17.0.10", "11.0.22", "8.0.402"];
            } else if (tool === "bun") {
              versions = ["1.2.4", "1.2.2", "1.1.45", "1.0.35"];
            } else if (tool === "deno") {
              versions = ["2.2.2", "2.1.9", "2.0.6", "1.46.3"];
            } else if (tool === "zig") {
              versions = ["0.13.0", "0.12.1", "0.11.0"];
            } else {
              versions = ["2.0.0", "1.5.0", "1.2.0", "1.0.0"];
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
