// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/frontend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/theme"
	"golang.org/x/net/websocket"
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

	httpServer := &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
		<-sigChan

		pterm.Info.Println("\nShutting down UniRTM Workbench...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			pterm.Error.Printf("HTTP shutdown error: %v\n", err)
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

func SetupUIMux(distFS fs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	// ─── API Endpoints ────────────────────────────────────────────────────────
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("/api/plugins", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cm := config.NewConfigManager()
		cfg, _ := cm.LoadHierarchy(r.Context())
		if cfg == nil {
			cfg = &config.Config{}
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"plugins": cfg.Plugins,
		})
	})

	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			handleGetConfig(w, r)
			return
		} else if r.Method == http.MethodPost {
			handlePostConfig(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
	})

	mux.HandleFunc("/api/config/raw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			handleGetConfigRaw(w, r)
			return
		} else if r.Method == http.MethodPost || r.Method == http.MethodPut {
			handlePostConfigRaw(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
	})

	mux.HandleFunc("/api/keybindings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			handleGetKeybindings(w, r)
			return
		} else if r.Method == http.MethodPost || r.Method == http.MethodPut {
			handlePostKeybindings(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
	})

	mux.HandleFunc("/api/system/browsers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"browsers": detectInstalledBrowsers(),
		})
	})
	mux.HandleFunc("/api/system/check-update", handleSystemCheckUpdate)
	mux.HandleFunc("/api/system/self-update", handleSystemSelfUpdate)
	mux.HandleFunc("/api/terminals", handleGetTerminals)
	mux.HandleFunc("/api/terminal/exec", handleTerminalExec)
	mux.Handle("/api/terminal/ws", websocket.Handler(HandleTerminalWebSocket))

	// ─── Network & Mirror API Endpoints ──────────────────────────────────────
	mux.HandleFunc("/api/network/status", handleGetNetworkStatus)
	mux.HandleFunc("/api/network/profile/use", handleUseNetworkProfile)
	mux.HandleFunc("/api/network/profile/save", handleSaveNetworkProfile)
	mux.HandleFunc("/api/network/profile/reset", handleResetNetworkProfile)
	mux.HandleFunc("/api/network/profile/delete", handleDeleteNetworkProfile)
	mux.HandleFunc("/api/network/registry/add", handleAddNetworkRegistry)
	mux.HandleFunc("/api/network/registry/delete", handleDeleteNetworkRegistry)
	mux.HandleFunc("/api/network/registry/override", handleOverrideNetworkRegistry)
	mux.HandleFunc("/api/network/registry/reset", handleResetNetworkRegistry)
	mux.HandleFunc("/api/network/registry/reset-all", handleResetAllNetworkRegistries)
	mux.HandleFunc("/api/network/proxy", handleSetNetworkProxy)
	mux.HandleFunc("/api/network/proxy/detect", handleDetectNetworkProxy)
	mux.HandleFunc("/api/network/proxy/test", handleTestNetworkProxy)
	mux.HandleFunc("/api/network/ping", handlePingNetwork)

	// ─── Tools & Runtimes APIs ───────────────────────────────────────────────
	RegisterToolsRoutes(mux)

	// Tasks Management APIs
	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetTasks(w, r)
		case http.MethodPost, http.MethodPut:
			handleUpsertTask(w, r)
		case http.MethodDelete:
			handleDeleteTask(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		}
	})
	mux.HandleFunc("/api/tasks/run", handleRunTask)

	// Environment Variables Management APIs
	mux.HandleFunc("/api/env", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetEnvVars(w, r)
		case http.MethodPost, http.MethodPut:
			handleUpsertEnvVar(w, r)
		case http.MethodDelete:
			handleDeleteEnvVar(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		}
	})

	// Lifecycle Hooks Management APIs
	mux.HandleFunc("/api/hooks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetHooks(w, r)
		case http.MethodPost, http.MethodPut:
			handleUpsertHook(w, r)
		case http.MethodDelete:
			handleDeleteHook(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		}
	})
	mux.HandleFunc("/api/hooks/install", handleInstallHooks)
	mux.HandleFunc("/api/hooks/run", handleRunHook)

	// ─── Quality Guard & Doctor Diagnostics ──────────────────────────────────
	RegisterGuardRoutes(mux)
	RegisterDoctorRoutes(mux)

	// Version Aliases Management APIs
	mux.HandleFunc("/api/aliases", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetAliases(w, r)
		case http.MethodPost, http.MethodPut:
			handleUpsertAlias(w, r)
		case http.MethodDelete:
			handleDeleteAlias(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		}
	})

	// Theme Management APIs
	themeMgr := theme.NewManager()
	mux.HandleFunc("/api/themes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetThemes(w, r, themeMgr)
		case http.MethodDelete:
			handleDeleteTheme(w, r, themeMgr)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		}
	})
	mux.HandleFunc("/api/themes/install", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handleInstallTheme(w, r, themeMgr)
	})
	mux.HandleFunc("/api/themes/apply", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handleApplyTheme(w, r)
	})
	mux.HandleFunc("/api/themes/assets/", func(w http.ResponseWriter, r *http.Request) {
		handleThemeAssets(w, r, themeMgr)
	})
	mux.HandleFunc("/api/themes/marketplace/search", func(w http.ResponseWriter, r *http.Request) {
		handleMarketplaceSearch(w, r, themeMgr)
	})
	mux.HandleFunc("/api/themes/marketplace/detail", func(w http.ResponseWriter, r *http.Request) {
		handleMarketplaceDetail(w, r, themeMgr)
	})
	mux.HandleFunc("/api/themes/marketplace/install", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handleMarketplaceInstall(w, r, themeMgr)
	})

	// ─── Multi-root Workspace API Handlers ────────────────────────────────────
	mux.HandleFunc("/api/workspace/roots", handleGetWorkspaceRoots)
	mux.HandleFunc("/api/workspace/roots/add", handleAddWorkspaceRoot)
	mux.HandleFunc("/api/workspace/roots/remove", handleRemoveWorkspaceRoot)

	// ─── File System API Handlers ─────────────────────────────────────────────
	mux.HandleFunc("/api/fs/tree", handleFSTree)
	mux.HandleFunc("/api/fs/create", handleFSCreate)
	mux.HandleFunc("/api/fs/rename", handleFSRename)
	mux.HandleFunc("/api/fs/delete", handleFSDelete)
	mux.HandleFunc("/api/fs/read", handleFSRead)
	mux.HandleFunc("/api/fs/write", handleFSWrite)

	// ─── Git SCM API Handlers ─────────────────────────────────────────────────
	mux.HandleFunc("/api/git/repositories", handleGitRepositories)
	mux.HandleFunc("/api/git/status", handleGitStatus)
	mux.HandleFunc("/api/git/stage", handleGitStage)
	mux.HandleFunc("/api/git/unstage", handleGitUnstage)
	mux.HandleFunc("/api/git/discard", handleGitDiscard)
	mux.HandleFunc("/api/git/commit", handleGitCommit)
	mux.HandleFunc("/api/git/diff", handleGitDiff)
	mux.HandleFunc("/api/git/branches", handleGitBranches)
	mux.HandleFunc("/api/git/checkout", handleGitCheckout)
	mux.HandleFunc("/api/git/sync", handleGitSync)
	mux.HandleFunc("/api/git/log", handleGitLog)
	mux.HandleFunc("/api/git/graph", handleGitGraph)
	mux.HandleFunc("/api/git/commit-detail", handleGitCommitDetail)
	mux.HandleFunc("/api/git/commit-diff", handleGitCommitDiff)
	mux.HandleFunc("/api/git/compare", handleGitCompare)
	mux.HandleFunc("/api/git/tag", handleGitTag)
	mux.HandleFunc("/api/git/branch/create", handleGitBranchCreate)
	mux.HandleFunc("/api/git/cherry-pick", handleGitCherryPick)
	mux.HandleFunc("/api/git/revert", handleGitRevert)
	mux.HandleFunc("/api/git/reset", handleGitReset)
	mux.HandleFunc("/api/git/branch/rename", handleGitBranchRename)
	mux.HandleFunc("/api/git/branch/delete", handleGitBranchDelete)
	mux.HandleFunc("/api/git/branch/merge", handleGitBranchMerge)
	mux.HandleFunc("/api/git/branch/rebase", handleGitBranchRebase)
	mux.HandleFunc("/api/git/branch/push", handleGitBranchPush)
	mux.HandleFunc("/api/git/tag/delete", handleGitTagDelete)
	mux.HandleFunc("/api/git/tag/push", handleGitTagPush)
	mux.HandleFunc("/api/git/archive", handleGitArchive)
	mux.HandleFunc("/api/git/stash", handleGitStash)
	mux.HandleFunc("/api/git/stash/apply", handleGitStashApply)
	mux.HandleFunc("/api/git/stash/pop", handleGitStashPop)
	mux.HandleFunc("/api/git/stash/drop", handleGitStashDrop)
	mux.HandleFunc("/api/git/stash/branch", handleGitStashBranch)
	mux.HandleFunc("/api/git/reset-uncommitted", handleGitResetUncommitted)
	mux.HandleFunc("/api/git/clean", handleGitClean)
	mux.HandleFunc("/api/git/pull", handleGitPull)
	mux.HandleFunc("/api/git/fetch", handleGitFetch)
	RegisterCredentialRoutes(mux)

	// ─── Port Forwarding API Handlers ──────────────────────────────────────────
	mux.HandleFunc("/api/ports/list", handlePortsList)
	mux.HandleFunc("/api/ports/forward", handlePortForward)
	mux.HandleFunc("/api/ports/update", handlePortUpdate)
	mux.HandleFunc("/api/ports/stop", handlePortStop)
	mux.HandleFunc("/api/ports/unforward", handlePortUnforward)
	mux.HandleFunc("/api/ports/preview/", handlePortPreview)
	mux.HandleFunc("/api/ports/probe", handlePortProbe)
	mux.HandleFunc("/api/ports/workspace-hints", handlePortsWorkspaceHints)

	// ─── Static SPA Handler ───────────────────────────────────────────────────
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

		// If static file exists, serve it; otherwise fallback to index.html (SPA routing)
		if _, err := fs.Stat(distFS, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA client-side routes
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	return mux
}

// setupUIMux is an alias for SetupUIMux.
var setupUIMux = SetupUIMux
