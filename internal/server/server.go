// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/frontend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// Options defines startup parameters for the Workbench HTTP server.
type Options struct {
	Host    string
	Port    int
	NoOpen  bool
	Browser string
}

// Run boots the embedded interactive graphical workbench server.
func Run(ctx context.Context, opts Options) error {
	distFS, err := frontend.GetFS()
	if err != nil {
		return fmt.Errorf("failed to load embedded frontend assets: %w", err)
	}

	// Verify that index.html exists in the embedded filesystem
	if _, err := fs.Stat(distFS, "index.html"); err != nil {
		pterm.Warning.Println("Embedded frontend dist assets not found or empty.")
		pterm.Info.Println("Please run 'npm run build' in the frontend directory before compiling.")
	}

	listener, actualPort, err := listenOnAvailablePort(opts.Host, opts.Port)
	if err != nil {
		return fmt.Errorf("failed to bind server port: %w", err)
	}
	ActivePort = actualPort

	serverURL := fmt.Sprintf("http://%s:%d", opts.Host, actualPort)

	mux := SetupUIMux(distFS)

	baseCtx, cancelBaseCtx := context.WithCancel(context.Background())
	defer cancelBaseCtx()

	httpServer := &http.Server{
		Handler: localSecurityMiddleware(mux),
		BaseContext: func(net.Listener) context.Context {
			return baseCtx
		},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Banner output
	pterm.DefaultSection.Println("UniRTM Workbench")
	pterm.Info.Printf("Local Server : %s\n", pterm.Cyan(serverURL))
	pterm.Info.Printf("Process PID  : %d\n", os.Getpid())
	if opts.Browser != "" {
		pterm.Info.Printf("Browser      : %s\n", pterm.Green(opts.Browser))
	}
	pterm.Info.Println("Press Ctrl+C to stop the workbench server.")

	// Auto-open browser in background
	if !opts.NoOpen {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := openBrowser(serverURL, opts.Browser); err != nil {
				pterm.Warning.Printf("Could not open browser automatically: %v\n", err)
				pterm.Info.Printf("Please open %s manually in your browser.\n", serverURL)
			}
		}()
	}

	// Graceful shutdown handling
	idleConnsClosed := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sigChan:
		case <-ctx.Done():
		}

		pterm.Info.Println("\nShutting down UniRTM Workbench...")

		// Cancel base context to notify all active SSE streams and request handlers to exit immediately
		cancelBaseCtx()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			// If graceful shutdown timed out or failed (e.g. hijacked WebSocket connections),
			// force close the server immediately without printing an alarm.
			_ = httpServer.Close()
		}
		close(idleConnsClosed)
	}()

	if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("workbench server error: %w", err)
	}

	<-idleConnsClosed
	pterm.Success.Println("UniRTM Workbench stopped cleanly.")
	return nil
}

func listenOnAvailablePort(host string, preferredPort int) (net.Listener, int, error) {
	addr := fmt.Sprintf("%s:%d", host, preferredPort)
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		return listener, preferredPort, nil
	}

	// If preferred port is in use, fallback to any available port
	fallbackAddr := fmt.Sprintf("%s:0", host)
	fallbackListener, err := net.Listen("tcp", fallbackAddr)
	if err != nil {
		return nil, 0, err
	}

	actualPort := fallbackListener.Addr().(*net.TCPAddr).Port
	return fallbackListener, actualPort, nil
}

// openBrowser opens the specified URL in the preferred or system default browser.
// Supports Windows, macOS, Linux, and WSL (Windows Subsystem for Linux).
func openBrowser(url string, targetBrowser string) error {
	targetBrowser = strings.TrimSpace(targetBrowser)

	// If a specific browser was requested (e.g. "chrome", "firefox", "edge", or executable path)
	if targetBrowser != "" {
		if err := openWithSpecificBrowser(url, targetBrowser); err == nil {
			return nil
		}
		// Graceful degradation: custom browser not installed or failed to start
		pterm.Warning.Printf("Specified browser %q not found or failed to launch. Falling back to system default browser.\n", targetBrowser)
	}

	// Fallback to system default browser handler
	return openWithDefaultBrowser(url)
}

// DetectedBrowser represents an installed browser found on the host system.
type DetectedBrowser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDefault bool   `json:"is_default"`
}

// detectInstalledBrowsers dynamically scans the host system for known installed browsers.

func resolveBrowserExecutable(target string) string {
	switch strings.ToLower(target) {
	case "chrome":
		for _, b := range []string{"google-chrome-stable", "google-chrome", "chromium", "chrome"} {
			if p, err := exec.LookPath(b); err == nil {
				return p
			}
		}
	case "msedge", "edge":
		for _, b := range []string{"microsoft-edge-stable", "microsoft-edge", "msedge"} {
			if p, err := exec.LookPath(b); err == nil {
				return p
			}
		}
	case "brave":
		for _, b := range []string{"brave-browser", "brave"} {
			if p, err := exec.LookPath(b); err == nil {
				return p
			}
		}
	case "firefox":
		for _, b := range []string{"firefox", "firefox-esr"} {
			if p, err := exec.LookPath(b); err == nil {
				return p
			}
		}
	}
	return target
}

func openWithSpecificBrowser(url string, targetBrowser string) error {
	resolved := resolveBrowserExecutable(targetBrowser)
	switch runtime.GOOS {
	case "darwin":
		if strings.Contains(resolved, "/") {
			return exec.Command(resolved, url).Start()
		}
		return exec.Command("open", "-a", resolved, url).Start()

	case "windows":
		return exec.Command("cmd", "/c", "start", "", resolved, url).Start()

	default:
		// Linux / Unix / WSL
		if isWSL() && strings.HasSuffix(strings.ToLower(resolved), ".exe") {
			return exec.Command("cmd.exe", "/c", "start", "", resolved, url).Start()
		}
		return exec.Command(resolved, url).Start()
	}
}

func openWithDefaultBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		// Windows: "cmd /c start <url>" handles associations cleanly. Fallback to rundll32.
		cmd := exec.Command("cmd", "/c", "start", "", url)
		if err := cmd.Start(); err == nil {
			return nil
		}
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()

	case "darwin":
		return exec.Command("open", url).Start()

	default:
		// Linux / WSL
		if isWSL() {
			if _, err := exec.LookPath("wslview"); err == nil {
				if err := exec.Command("wslview", url).Start(); err == nil {
					return nil
				}
			}
			if err := exec.Command("cmd.exe", "/c", "start", "", url).Start(); err == nil {
				return nil
			}
		}

		if _, err := exec.LookPath("xdg-open"); err == nil {
			if err := exec.Command("xdg-open", url).Start(); err == nil {
				return nil
			}
		}
		if _, err := exec.LookPath("gio"); err == nil {
			if err := exec.Command("gio", "open", url).Start(); err == nil {
				return nil
			}
		}
		return exec.Command("x-www-browser", url).Start()
	}
}

// isWSL detects if the current Linux environment is running inside Windows Subsystem for Linux.
func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}
	data, err := os.ReadFile("/proc/version")
	if err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "microsoft") || strings.Contains(content, "wsl") {
			return true
		}
	}
	return false
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cm := config.NewConfigManager()
	cfg, _ := cm.LoadHierarchy(r.Context())
	pluginCount := 0
	if cfg != nil {
		pluginCount = len(cfg.Plugins)
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ready",
		"version":       env.GitTag,
		"commit":        env.CommitHash,
		"platform":      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"go_version":    runtime.Version(),
		"tools_count":   5,
		"tasks_count":   4,
		"plugins_count": pluginCount,
		"timestamp":     time.Now().Unix(),
	})
}

func handlePlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cm := config.NewConfigManager()
	cfg, _ := cm.LoadHierarchy(r.Context())
	if cfg == nil {
		cfg = &config.Config{}
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"plugins": cfg.Plugins,
	})
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return gz
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

func isCompressibleExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".mjs", ".css", ".html", ".svg", ".json", ".map", ".txt", ".xml", ".wasm":
		return true
	default:
		return false
	}
}

// RegisterStaticSPARoutes mounts the embedded frontend SPA filesystem with gzip compression and caching headers.
func RegisterStaticSPARoutes(mux *http.ServeMux, distFS fs.FS) {
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		targetPath := path
		if _, err := fs.Stat(distFS, path); err != nil {
			// Fallback to index.html for SPA client-side routes
			targetPath = "index.html"
			r.URL.Path = "/"
		}

		// Apply caching headers based on path pattern
		if strings.HasPrefix(targetPath, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		// Check if client supports gzip and content is compressible
		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if acceptsGzip && isCompressibleExtension(targetPath) {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")

			gz := gzipWriterPool.Get().(*gzip.Writer)
			defer gzipWriterPool.Put(gz)

			gz.Reset(w)
			defer gz.Close()

			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				writer:         gz,
			}
			fileServer.ServeHTTP(gzw, r)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}

// localSecurityMiddleware protects local developer workbench from cross-site DNS rebinding and malicious origins.
func localSecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if u, err := url.Parse(origin); err == nil {
				h := u.Hostname()
				if h != "localhost" && h != "127.0.0.1" && h != "::1" && h != "" {
					http.Error(w, "Forbidden: Untrusted Origin", http.StatusForbidden)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// SetupUIMux configures and returns the central HTTP servemux with all modular domain routes.
func SetupUIMux(distFS fs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	// ─── Status & Info ────────────────────────────────────────────────────────
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/plugins", handlePlugins)

	// ─── Domain Module Route Registrations ────────────────────────────────────
	RegisterSystemRoutes(mux)
	RegisterConfigRoutes(mux)
	RegisterNetworkRoutes(mux)
	RegisterToolsRoutes(mux)
	RegisterTasksRoutes(mux)
	RegisterGuardRoutes(mux)
	RegisterDoctorRoutes(mux)
	RegisterMarketplaceRoutes(mux)
	RegisterFSRoutes(mux)
	RegisterGitRoutes(mux)
	RegisterGitProfileRoutes(mux)
	RegisterCredentialRoutes(mux)
	RegisterPortsRoutes(mux)

	// ─── Static SPA Handler ───────────────────────────────────────────────────
	RegisterStaticSPARoutes(mux, distFS)

	return mux
}

// setupUIMux is an alias for SetupUIMux.
var setupUIMux = SetupUIMux
