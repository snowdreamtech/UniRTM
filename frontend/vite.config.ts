import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import fs from "node:fs";

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
        server.middlewares.use((req, res, next) => {
          if (req.url === "/api/system/browsers") {
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ browsers: [] }));
            return;
          }
          if (req.url === "/api/config") {
            if (req.method === "GET") {
              res.setHeader("Content-Type", "application/json");
              res.end(
                JSON.stringify({
                  user: {
                    path: "~/.config/unirtm/unirtm.toml",
                    exists: true,
                    settings: { browser: "", cache_ttl: "24h" },
                  },
                  workspace: {
                    path: "./.unirtm.toml",
                    exists: true,
                    settings: { cache_ttl: "24h" },
                  },
                  detected_browsers: [],
                })
              );
              return;
            } else if (req.method === "POST") {
              res.setHeader("Content-Type", "application/json");
              res.end(JSON.stringify({ ok: true }));
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
