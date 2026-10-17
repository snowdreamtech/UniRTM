// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

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
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/frontend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/hook"
	"github.com/snowdreamtech/unirtm/internal/network"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/theme"
	"github.com/spf13/cobra"
)

var (
	uiPort    int
	uiHost    string
	uiNoOpen  bool
	uiBrowser string
)

func init() {
	uiCmd.Flags().IntVarP(&uiPort, "port", "p", 5173, "port to listen on")
	uiCmd.Flags().StringVarP(&uiHost, "host", "H", "127.0.0.1", "host address to bind")
	uiCmd.Flags().BoolVar(&uiNoOpen, "no-open", false, "do not open browser automatically")
	uiCmd.Flags().StringVarP(&uiBrowser, "browser", "b", "", "browser to open (e.g. chrome, firefox, edge, or executable path)")

	if rootCmd != nil {
		rootCmd.AddCommand(uiCmd)
	}
}

// uiCmd starts the embedded interactive graphical workbench.
var uiCmd = &cobra.Command{
	Use:     "ui",
	Aliases: []string{"dashboard", "web"},
	Short:   "Launch the interactive graphical workbench",
	Long: `Launch the interactive graphical workbench in your default browser.

Starts an embedded local web server serving the UniRTM visual workspace to
manage tools, run tasks, inspect doctor health, and manage extensions.

Examples:
  # Start UI and open browser automatically
  unirtm ui

  # Start UI on a custom port without opening browser
  unirtm ui --port 8080 --no-open

  # Use alias
  unirtm dashboard
  unirtm web`,
	Args: cobra.NoArgs,
	RunE: runUI,
}

func runUI(cmd *cobra.Command, args []string) error {
	distFS, err := frontend.GetFS()
	if err != nil {
		return fmt.Errorf("failed to load embedded frontend assets: %w", err)
	}

	// Verify that index.html exists in the embedded filesystem
	if _, err := fs.Stat(distFS, "index.html"); err != nil {
		pterm.Warning.Println("Embedded frontend dist assets not found or empty.")
		pterm.Info.Println("Please run 'npm run build' in the frontend directory before compiling.")
	}

	listener, actualPort, err := listenOnAvailablePort(uiHost, uiPort)
	if err != nil {
		return fmt.Errorf("failed to bind server port: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d", uiHost, actualPort)

	mux := setupUIMux(distFS)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Banner output
	pterm.DefaultSection.Println("UniRTM Workbench")
	pterm.Info.Printf("Local Server : %s\n", pterm.Cyan(url))
	// Determine target browser:
	// Priority 1: CLI flag --browser / -b (temporary override for current invocation)
	// Priority 2: Global environment variable $BROWSER
	// Priority 3: User global/machine configuration only (LoadGlobal).
	// NOTE: Project-level .unirtm.toml configuration is explicitly ignored for browser settings
	// to prevent team repository pollution and ensure cross-machine compatibility.
	targetBrowser := strings.TrimSpace(uiBrowser)
	if targetBrowser == "" {
		if envBrowser := os.Getenv("BROWSER"); envBrowser != "" {
			targetBrowser = strings.TrimSpace(envBrowser)
		} else if globalCfg, err := config.LoadGlobal(); err == nil && globalCfg != nil && globalCfg.Settings.Browser != "" {
			targetBrowser = strings.TrimSpace(globalCfg.Settings.Browser)
		}
	}

	// Warn if project config attempted to define browser
	if localCfg, err := loadConfig(cmd.Context()); err == nil && localCfg != nil && localCfg.Settings.Browser != "" {
		pterm.Warning.Println("Notice: 'settings.browser' in project configuration is ignored (browser preferences are user-global only).")
	}

	pterm.Info.Printf("Process PID  : %d\n", os.Getpid())
	if targetBrowser != "" {
		pterm.Info.Printf("Browser      : %s\n", pterm.Green(targetBrowser))
	}
	pterm.Info.Println("Press Ctrl+C to stop the workbench server.")

	// Auto-open browser in background
	if !uiNoOpen {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := openBrowser(url, targetBrowser); err != nil {
				pterm.Warning.Printf("Could not open browser automatically: %v\n", err)
				pterm.Info.Printf("Please open %s manually in your browser.\n", url)
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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			pterm.Error.Printf("HTTP shutdown error: %v\n", err)
		}
		close(idleConnsClosed)
	}()

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("workbench server error: %w", err)
	}

	<-idleConnsClosed
	pterm.Success.Println("UniRTM Workbench stopped cleanly.")
	return nil
}

// listenOnAvailablePort attempts to bind to preferredPort, or finds an available port if preferredPort is busy.
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
func detectInstalledBrowsers() []DetectedBrowser {
	var list []DetectedBrowser
	seen := make(map[string]bool)

	addBrowser := func(id, name, path string, isDefault bool) {
		if seen[id] || path == "" {
			return
		}
		seen[id] = true
		list = append(list, DetectedBrowser{
			ID:        id,
			Name:      name,
			Path:      path,
			IsDefault: isDefault,
		})
	}

	switch runtime.GOOS {
	case "darwin":
		knownMac := []struct {
			id   string
			name string
			app  string
		}{
			{"chrome", "Google Chrome", "/Applications/Google Chrome.app"},
			{"safari", "Safari", "/Applications/Safari.app"},
			{"firefox", "Mozilla Firefox", "/Applications/Firefox.app"},
			{"msedge", "Microsoft Edge", "/Applications/Microsoft Edge.app"},
			{"brave", "Brave Browser", "/Applications/Brave Browser.app"},
			{"arc", "Arc", "/Applications/Arc.app"},
			{"chromium", "Chromium", "/Applications/Chromium.app"},
			{"opera", "Opera", "/Applications/Opera.app"},
			{"vivaldi", "Vivaldi", "/Applications/Vivaldi.app"},
		}
		home, _ := os.UserHomeDir()
		for _, b := range knownMac {
			if _, err := os.Stat(b.app); err == nil {
				addBrowser(b.id, b.name, b.app, false)
			} else if home != "" {
				userApp := filepath.Join(home, "Applications", filepath.Base(b.app))
				if _, err := os.Stat(userApp); err == nil {
					addBrowser(b.id, b.name, userApp, false)
				}
			}
		}
		if !seen["safari"] {
			if _, err := os.Stat("/System/Applications/Safari.app"); err == nil {
				addBrowser("safari", "Safari", "/System/Applications/Safari.app", false)
			}
		}

	case "windows":
		knownWin := []struct {
			id      string
			name    string
			subPath string
		}{
			{"chrome", "Google Chrome", `Google\Chrome\Application\chrome.exe`},
			{"msedge", "Microsoft Edge", `Microsoft\Edge\Application\msedge.exe`},
			{"firefox", "Mozilla Firefox", `Mozilla Firefox\firefox.exe`},
			{"brave", "Brave Browser", `BraveSoftware\Brave-Browser\Application\brave.exe`},
			{"opera", "Opera", `Opera\launcher.exe`},
			{"vivaldi", "Vivaldi", `Vivaldi\Application\vivaldi.exe`},
		}

		prefixes := []string{
			os.Getenv("ProgramFiles"),
			os.Getenv("ProgramFiles(x86)"),
			os.Getenv("LocalAppData"),
		}

		for _, b := range knownWin {
			found := false
			for _, p := range prefixes {
				if p == "" {
					continue
				}
				full := filepath.Join(p, b.subPath)
				if _, err := os.Stat(full); err == nil {
					addBrowser(b.id, b.name, full, false)
					found = true
					break
				}
			}
			if !found {
				if p, err := exec.LookPath(b.id + ".exe"); err == nil {
					addBrowser(b.id, b.name, p, false)
				}
			}
		}

	default:
		// Linux / Unix / WSL
		knownLinux := []struct {
			id   string
			name string
			bins []string
		}{
			{"chrome", "Google Chrome", []string{"google-chrome-stable", "google-chrome"}},
			{"firefox", "Mozilla Firefox", []string{"firefox", "firefox-esr"}},
			{"msedge", "Microsoft Edge", []string{"microsoft-edge-stable", "microsoft-edge"}},
			{"chromium", "Chromium", []string{"chromium-browser", "chromium"}},
			{"brave", "Brave Browser", []string{"brave-browser", "brave"}},
			{"opera", "Opera", []string{"opera"}},
			{"vivaldi", "Vivaldi", []string{"vivaldi-stable", "vivaldi"}},
			{"epiphany", "GNOME Web", []string{"epiphany"}},
		}

		defaultHandler := ""
		if out, err := exec.Command("xdg-mime", "query", "default", "x-scheme-handler/http").Output(); err == nil {
			defaultHandler = strings.ToLower(strings.TrimSpace(string(out)))
		}

		for _, b := range knownLinux {
			for _, bin := range b.bins {
				if p, err := exec.LookPath(bin); err == nil {
					isDef := false
					if defaultHandler != "" && strings.Contains(defaultHandler, b.id) {
						isDef = true
					}
					addBrowser(b.id, b.name, p, isDef)
					break
				}
			}
		}
	}

	return list
}

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

func setupUIMux(distFS fs.FS) *http.ServeMux {
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

	mux.HandleFunc("/api/system/browsers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"browsers": detectInstalledBrowsers(),
		})
	})

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

	// Tools Management APIs (Canonical)
	mux.HandleFunc("/api/tools/status", handleGetToolsStatus)
	mux.HandleFunc("/api/tools/remote", handleGetToolsRemote)
	mux.HandleFunc("/api/tools/use", handleUseTool)
	mux.HandleFunc("/api/tools/remove", handleRemoveTool)
	mux.HandleFunc("/api/tools/pin", handlePinTool)
	mux.HandleFunc("/api/tools/check-updates", handleCheckToolsUpdates)
	mux.HandleFunc("/api/tools/batch-update", handleBatchUpdateTools)
	mux.HandleFunc("/api/tools/migrate-legacy", handleMigrateLegacyTool)

	// Backward Compatibility Aliases (/api/runtimes/*)
	mux.HandleFunc("/api/runtimes/status", handleGetToolsStatus)
	mux.HandleFunc("/api/runtimes/remote", handleGetToolsRemote)
	mux.HandleFunc("/api/runtimes/use", handleUseTool)
	mux.HandleFunc("/api/runtimes/remove", handleRemoveTool)
	mux.HandleFunc("/api/runtimes/pin", handlePinRuntime)
	mux.HandleFunc("/api/runtimes/check-updates", handleCheckRuntimesUpdates)
	mux.HandleFunc("/api/runtimes/batch-update", handleBatchUpdateRuntimes)
	mux.HandleFunc("/api/runtimes/migrate-legacy", handleMigrateLegacyTool)

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

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	userPath := config.GetGlobalConfigPath()
	userExists := false
	var userRaw string
	var userCfg *config.Config
	if data, err := os.ReadFile(userPath); err == nil {
		userExists = true
		userRaw = string(data)
	}
	userCfg, _ = config.LoadGlobal()
	if userCfg == nil {
		userCfg = &config.Config{}
	}

	wsPath := ".unirtm.toml"
	wsExists := false
	var wsRaw string
	if data, err := os.ReadFile(wsPath); err == nil {
		wsExists = true
		wsRaw = string(data)
	}
	cm := config.NewConfigManager()
	wsCfg, _ := cm.LoadHierarchy(r.Context())
	if wsCfg == nil {
		wsCfg = &config.Config{}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"path":     userPath,
			"exists":   userExists,
			"raw":      userRaw,
			"settings": userCfg.Settings,
			"plugins":  userCfg.Plugins,
			"env":      userCfg.Env,
			"hooks":    userCfg.Hooks,
			"aliases":  userCfg.Aliases,
		},
		"workspace": map[string]interface{}{
			"path":     wsPath,
			"exists":   wsExists,
			"raw":      wsRaw,
			"settings": wsCfg.Settings,
			"plugins":  wsCfg.Plugins,
			"env":      wsCfg.Env,
			"tools":    wsCfg.Tools,
			"tasks":    wsCfg.Tasks,
			"hooks":    wsCfg.Hooks,
			"aliases":  wsCfg.Aliases,
		},
		"detected_browsers": detectInstalledBrowsers(),
	})
}

func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	var tasksMap map[string]config.Task
	if scope == "user" {
		userCfg, _ := config.LoadGlobal()
		if userCfg != nil {
			tasksMap = userCfg.Tasks
		}
	} else {
		cm := config.NewConfigManager()
		cfg, _ := cm.LoadHierarchy(r.Context())
		if cfg != nil {
			tasksMap = cfg.Tasks
		}
	}
	if tasksMap == nil {
		tasksMap = make(map[string]config.Task)
	}

	type TaskResp struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Command     string   `json:"command"`
		Depends     []string `json:"depends,omitempty"`
		Env         string   `json:"env,omitempty"`
		Output      string   `json:"output,omitempty"`
		Timeout     int      `json:"timeout,omitempty"`
	}

	var names []string
	for n := range tasksMap {
		names = append(names, n)
	}
	sort.Strings(names)

	var res []TaskResp
	for _, n := range names {
		t := tasksMap[n]
		cmdStr := strings.TrimSpace(strings.Join(t.Run, " && "))
		if cmdStr == "" && len(t.Depends) > 0 {
			cmdStr = fmt.Sprintf("unirtm run %s", strings.Join(t.Depends, " "))
		}
		var envStr string
		if len(t.Env) > 0 {
			var parts []string
			for k, v := range t.Env {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
			sort.Strings(parts)
			envStr = strings.Join(parts, " ")
		}
		res = append(res, TaskResp{
			Name:        n,
			Description: t.Description,
			Command:     cmdStr,
			Depends:     t.Depends,
			Env:         envStr,
			Output:      t.Output,
			Timeout:     t.Timeout,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks": res,
		"total": len(res),
		"scope": scope,
	})
}

type UpsertTaskReq struct {
	Name         string   `json:"name"`
	OriginalName string   `json:"original_name,omitempty"`
	Description  string   `json:"description,omitempty"`
	Command      string   `json:"command"`
	Depends      []string `json:"depends,omitempty"`
	Scope        string   `json:"scope,omitempty"`
}

func handleUpsertTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req UpsertTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "task name is required"})
		return
	}

	targetFile := ".unirtm.toml"
	if strings.ToLower(req.Scope) == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)

	if req.OriginalName != "" && req.OriginalName != req.Name {
		content, _ = config.RemoveTask(content, req.OriginalName)
	}

	updated := config.UpsertTask(content, req.Name, req.Description, req.Command, req.Depends)
	if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"task":    req,
	})
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))

	if name == "" {
		var body struct {
			Name  string `json:"name"`
			Scope string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			name = strings.TrimSpace(body.Name)
			if scope == "" {
				scope = strings.ToLower(strings.TrimSpace(body.Scope))
			}
		}
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "task name is required"})
		return
	}

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, err := config.ReadFileOrEmpty(targetFile)
	if err != nil || content == "" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}

	updated, removed := config.RemoveTask(content, name)
	if removed {
		if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func handleRunTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name  string `json:"name"`
		Scope string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "task name is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "run", body.Name)
	if body.Scope == "user" {
		cmd.Args = append(cmd.Args, "--global")
	}
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   exitCode == 0,
		"exit_code": exitCode,
		"output":    string(out),
		"name":      body.Name,
	})
}

// ─── Lifecycle Hooks Management Handlers ──────────────────────────────────────

type HookResp struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	Installed bool   `json:"installed"`
}

func checkHookInstalledInGit(hookName string) bool {
	gitHooksDir := filepath.Join(".git", "hooks")
	hookPath := filepath.Join(gitHooksDir, hookName)
	data, err := os.ReadFile(hookPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "unirtm hook")
}

func handleGetHooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if scope == "" {
		scope = "workspace"
	}

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)
	hooksMap := config.GetHooksFromContent(content)

	var res []HookResp
	var names []string
	for k := range hooksMap {
		names = append(names, k)
	}
	sort.Strings(names)

	gitDirExists := false
	if fi, err := os.Stat(".git"); err == nil && fi.IsDir() {
		gitDirExists = true
	}

	for _, name := range names {
		installed := false
		if scope != "user" && gitDirExists {
			installed = checkHookInstalledInGit(name)
		}
		res = append(res, HookResp{
			Name:      name,
			Command:   hooksMap[name],
			Installed: installed,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"hooks":          res,
		"total":          len(res),
		"scope":          scope,
		"git_dir_exists": gitDirExists,
	})
}

func handleUpsertHook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Name          string `json:"name"`
		Command       string `json:"command"`
		Scope         string `json:"scope"`
		InstallBridge bool   `json:"install_bridge"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request body: " + err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Command = strings.TrimSpace(req.Command)
	if req.Scope == "" {
		req.Scope = "workspace"
	}
	req.Scope = strings.ToLower(strings.TrimSpace(req.Scope))

	if req.Scope == "user" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Git lifecycle hooks are only permitted in workspace scope as they bind to local repository .git/hooks.",
			"hint":  "Please configure hooks in Workspace (工作区) scope instead.",
		})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "hook name cannot be empty"})
		return
	}
	if req.Command == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "command cannot be empty"})
		return
	}

	targetFile := ".unirtm.toml"
	content, _ := config.ReadFileOrEmpty(targetFile)
	updated := config.UpsertHook(content, req.Name, req.Command)
	if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
		return
	}

	installed := false
	if req.InstallBridge {
		pwd, _ := os.Getwd()
		_ = hook.InstallBridgeScript(r.Context(), pwd, req.Name)
		installed = checkHookInstalledInGit(req.Name)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"name":      req.Name,
		"command":   req.Command,
		"scope":     "workspace",
		"installed": installed,
	})
}

func handleDeleteHook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))

	if name == "" {
		var body struct {
			Name  string `json:"name"`
			Scope string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			name = strings.TrimSpace(body.Name)
			if scope == "" {
				scope = strings.ToLower(strings.TrimSpace(body.Scope))
			}
		}
	}

	if scope == "user" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Git lifecycle hooks are only permitted in workspace scope.",
		})
		return
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "hook name is required"})
		return
	}

	targetFile := ".unirtm.toml"
	content, err := config.ReadFileOrEmpty(targetFile)
	if err != nil || content == "" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}

	updated, removed := config.UnsetHook(content, name)
	if removed {
		if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"deleted_name": name,
	})
}

func handleInstallHooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name string `json:"name"`
		All  bool   `json:"all"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	pwd, err := os.Getwd()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if body.All || body.Name == "" {
		updated, err := hook.InstallAllBridgeScripts(r.Context(), pwd)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"installed": updated,
		})
		return
	}

	if err := hook.ValidateHookName(body.Name); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if err := hook.InstallBridgeScript(r.Context(), pwd, body.Name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"installed": []string{body.Name},
	})
}

func handleRunHook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name  string   `json:"name"`
		Args  []string `json:"args"`
		Scope string   `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "hook name is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "hook", "run", body.Name)
	if len(body.Args) > 0 {
		cmd.Args = append(cmd.Args, body.Args...)
	}
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   exitCode == 0,
		"exit_code": exitCode,
		"output":    string(out),
		"name":      body.Name,
	})
}

// ─── Version Aliases Management Handlers ─────────────────────────────────────

func handleGetAliases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if scope == "" {
		scope = "workspace"
	}

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)
	aliases := config.GetAliasesFromContent(content)
	if aliases == nil {
		aliases = make(map[string]map[string]string)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"scope":   scope,
		"aliases": aliases,
	})
}

func handleUpsertAlias(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Tool    string `json:"tool"`
		Alias   string `json:"alias"`
		Version string `json:"version"`
		Scope   string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request body: " + err.Error()})
		return
	}

	req.Tool = strings.TrimSpace(req.Tool)
	req.Alias = strings.TrimSpace(req.Alias)
	req.Version = strings.TrimSpace(req.Version)
	if req.Scope == "" {
		req.Scope = "workspace"
	}
	req.Scope = strings.ToLower(strings.TrimSpace(req.Scope))

	if req.Tool == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "tool name cannot be empty"})
		return
	}
	if req.Alias == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "alias name cannot be empty"})
		return
	}
	if req.Version == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "version cannot be empty"})
		return
	}

	targetFile := ".unirtm.toml"
	if req.Scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)
	updated := config.UpsertAlias(content, req.Tool, req.Alias, req.Version)
	if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tool":    req.Tool,
		"alias":   req.Alias,
		"version": req.Version,
		"scope":   req.Scope,
	})
}

func handleDeleteAlias(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tool := strings.TrimSpace(r.URL.Query().Get("tool"))
	alias := strings.TrimSpace(r.URL.Query().Get("alias"))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))

	if tool == "" || alias == "" {
		var body struct {
			Tool  string `json:"tool"`
			Alias string `json:"alias"`
			Scope string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if tool == "" {
				tool = strings.TrimSpace(body.Tool)
			}
			if alias == "" {
				alias = strings.TrimSpace(body.Alias)
			}
			if scope == "" {
				scope = strings.ToLower(strings.TrimSpace(body.Scope))
			}
		}
	}

	if scope == "" {
		scope = "workspace"
	}

	if tool == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "tool name is required"})
		return
	}
	if alias == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "alias name is required"})
		return
	}

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, err := config.ReadFileOrEmpty(targetFile)
	if err != nil || content == "" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}

	updated, removed := config.RemoveAlias(content, tool, alias)
	if removed {
		if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tool":    tool,
		"alias":   alias,
		"scope":   scope,
	})
}

// ─── Environment Variables Management Handlers ────────────────────────────────

func isSensitiveEnvKey(k string) bool {
	upper := strings.ToUpper(k)
	sensitiveSubstrings := []string{
		"TOKEN", "KEY", "SECRET", "PASSWORD", "PASS", "AUTH",
		"PRIVATE", "CREDENTIAL", "APIKEY", "ACCESS", "CERT",
		"SALT", "HASH", "SIGNATURE",
	}
	for _, s := range sensitiveSubstrings {
		if strings.Contains(upper, s) {
			return true
		}
	}
	return false
}

func handleGetEnvVars(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if scope == "" {
		scope = "workspace"
	}

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)
	envMap := config.GetEnvVarsFromContent(content)

	var names []string
	for n := range envMap {
		names = append(names, n)
	}
	sort.Strings(names)

	type EnvVarItem struct {
		Name        string `json:"name"`
		Value       string `json:"value"`
		Scope       string `json:"scope"`
		IsSensitive bool   `json:"is_sensitive"`
	}

	var res []EnvVarItem
	for _, n := range names {
		res = append(res, EnvVarItem{
			Name:        n,
			Value:       envMap[n],
			Scope:       scope,
			IsSensitive: isSensitiveEnvKey(n),
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"env":   res,
		"total": len(res),
		"scope": scope,
	})
}

type UpsertEnvVarReq struct {
	Key         string `json:"key"`
	OriginalKey string `json:"original_key,omitempty"`
	Value       string `json:"value"`
	Scope       string `json:"scope,omitempty"`
}

func handleUpsertEnvVar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req UpsertEnvVarReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Key) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "variable key is required"})
		return
	}

	scope := strings.ToLower(strings.TrimSpace(req.Scope))
	if scope == "" {
		scope = "workspace"
	}
	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)

	if req.OriginalKey != "" && !strings.EqualFold(req.OriginalKey, req.Key) {
		content, _ = config.UnsetEnvVar(content, req.OriginalKey)
	}

	updated := config.UpsertEnvVar(content, req.Key, req.Value)
	if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
		return
	}

	upperKey := strings.ToUpper(strings.TrimSpace(req.Key))
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"env": map[string]interface{}{
			"name":         upperKey,
			"value":        req.Value,
			"scope":        scope,
			"is_sensitive": isSensitiveEnvKey(upperKey),
		},
	})
}

func handleDeleteEnvVar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := strings.TrimSpace(r.URL.Query().Get("key"))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))

	if key == "" {
		var body struct {
			Key   string `json:"key"`
			Scope string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			key = strings.TrimSpace(body.Key)
			if scope == "" {
				scope = strings.ToLower(strings.TrimSpace(body.Scope))
			}
		}
	}

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "variable key is required"})
		return
	}

	if scope == "" {
		scope = "workspace"
	}
	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, err := config.ReadFileOrEmpty(targetFile)
	if err != nil || content == "" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}

	updated, removed := config.UnsetEnvVar(content, key)
	if removed {
		if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"deleted_key": strings.ToUpper(key),
	})
}

func handlePostConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Scope string      `json:"scope"`
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body: " + err.Error()})
		return
	}

	scope := strings.ToLower(strings.TrimSpace(body.Scope))
	key := strings.TrimSpace(body.Key)
	if scope != "user" && scope != "workspace" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "scope must be 'user' or 'workspace'"})
		return
	}

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "key cannot be empty"})
		return
	}

	// Machine-specific settings check: browser is global-only
	cleanKey := strings.TrimPrefix(key, "settings.")
	if scope == "workspace" && cleanKey == "browser" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Browser configuration is only permitted in global config to ensure project portability across team machines.",
			"hint":  "Please configure settings.browser in User (全局) scope instead.",
		})
		return
	}

	// Validate namespace adhering to reserved core tables and plugin namespaces
	if err := config.ValidateKeyNamespace(key); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	valStr := fmt.Sprintf("%v", body.Value)

	// Validate plugin schema if applicable
	segments := config.SplitConfigKey(key)
	if len(segments) >= 4 && segments[0] == "plugins" && segments[2] == "config" {
		pluginID := segments[1]
		property := segments[3]
		cwdDir, _ := os.Getwd()
		manifest, _ := config.FindPluginManifest(pluginID, cwdDir)
		if manifest != nil && manifest.Contributes.Configuration != nil {
			if schema, exists := manifest.Contributes.Configuration[property]; exists {
				if err := config.ValidateConfigValue(property, valStr, schema); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"error": fmt.Sprintf("plugin schema validation failed: %v", err),
					})
					return
				}
			}
		}
	}

	var targetFile string
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
		if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "create global config dir: " + err.Error()})
			return
		}
	} else {
		targetFile = ".unirtm.toml"
	}

	// Normalize settings shorthand (e.g. "browser" -> "settings.browser")
	storeKey := key
	if !strings.HasPrefix(storeKey, "settings.") && (storeKey == "browser" || storeKey == "cache_ttl" || storeKey == "theme" || storeKey == "icon_theme" || storeKey == "language") {
		storeKey = "settings." + storeKey
	}

	fileLock, lockErr := config.AcquireFileLock(targetFile, 5*time.Second)
	if lockErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "acquire config file lock: " + lockErr.Error()})
		return
	}
	defer fileLock.Release()

	existingRaw, _ := os.ReadFile(targetFile)
	updatedTOML, err := config.UpsertTOMLSetting(string(existingRaw), config.SplitConfigKey(storeKey), valStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "update toml setting: " + err.Error()})
		return
	}

	if err := os.WriteFile(targetFile, []byte(updatedTOML), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "write config file: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"scope":   scope,
		"key":     storeKey,
		"value":   valStr,
		"file":    targetFile,
	})
}

func resolveNetworkPathFromRequest(r *http.Request, defaultScope string) string {
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = defaultScope
	}
	if scope == "user" || scope == "global" {
		return config.GetGlobalConfigPath()
	}
	if _, err := os.Stat(".unirtm.toml"); err == nil {
		return ".unirtm.toml"
	}
	return config.GetGlobalConfigPath()
}

func handleGetNetworkStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	targetPath := resolveNetworkPathFromRequest(r, "workspace")
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	status := mgr.GetStatus()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      status,
		"recommended": network.RecommendedMirrors(),
		"path":        targetPath,
	})
}

func handleUseNetworkProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ProfileID string `json:"profile_id"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.UseProfile(req.ProfileID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleSaveNetworkProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ProfileID   string `json:"profile_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.SaveAsProfile(req.ProfileID, req.Name, req.Description); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleResetNetworkProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ProfileID string `json:"profile_id"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.ResetProfile(req.ProfileID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleDeleteNetworkProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ProfileID string `json:"profile_id"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.DeleteProfile(req.ProfileID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleOverrideNetworkRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Ecosystem string `json:"ecosystem"`
		URL       string `json:"url"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.SetOverride(req.Ecosystem, req.URL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleResetNetworkRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Ecosystem string `json:"ecosystem"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.ResetOverride(req.Ecosystem); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleAddNetworkRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Ecosystem string `json:"ecosystem"`
		URL       string `json:"url"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.AddEcosystem(req.Ecosystem, req.URL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleDeleteNetworkRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Ecosystem string `json:"ecosystem"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.DeleteEcosystem(req.Ecosystem); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleResetAllNetworkRegistries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Scope string `json:"scope"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.ResetDefaultEcosystems(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleSetNetworkProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Enabled bool   `json:"enabled"`
		HTTP    string `json:"http"`
		HTTPS   string `json:"https"`
		SOCKS5  string `json:"socks5"`
		NoProxy string `json:"no_proxy"`
		Scope   string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	proxyCfg := network.ProxyConfig{
		Enabled: req.Enabled,
		HTTP:    req.HTTP,
		HTTPS:   req.HTTPS,
		SOCKS5:  req.SOCKS5,
		NoProxy: req.NoProxy,
	}

	if err := mgr.SetProxy(proxyCfg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleDetectNetworkProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	detected := network.DetectSystemProxy()
	_ = json.NewEncoder(w).Encode(detected)
}

func handleTestNetworkProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		HTTP   string `json:"http"`
		HTTPS  string `json:"https"`
		SOCKS5 string `json:"socks5"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	res := network.TestProxyConnectivity(network.ProxyConfig{
		Enabled: true,
		HTTP:    req.HTTP,
		HTTPS:   req.HTTPS,
		SOCKS5:  req.SOCKS5,
	}, 4*time.Second)

	_ = json.NewEncoder(w).Encode(res)
}

func handlePingNetwork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var targets []network.ProbeTarget

	if r.Method == http.MethodPost && r.Body != nil {
		var req struct {
			Targets []struct {
				Ecosystem string `json:"ecosystem"`
				Name      string `json:"name"`
				URL       string `json:"url"`
			} `json:"targets"`
			Ecosystem string `json:"ecosystem"`
			URL       string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if len(req.Targets) > 0 {
				for _, t := range req.Targets {
					targets = append(targets, network.ProbeTarget{
						Ecosystem: t.Ecosystem,
						Name:      t.Name,
						Endpoint:  t.URL,
					})
				}
			} else if req.URL != "" {
				targets = append(targets, network.ProbeTarget{
					Ecosystem: req.Ecosystem,
					Name:      req.Ecosystem,
					Endpoint:  req.URL,
				})
			}
		}
	}

	if len(targets) == 0 {
		targets = network.DefaultProbeTargets()
	}

	results := network.ProbeAll(targets, 3000*time.Millisecond)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}

// ─── Theme Management Handlers ────────────────────────────────────────────────

func handleGetThemes(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	resp, err := themeMgr.ListThemes()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	activeColor := "default"
	activeIcon := "seti"
	activeProduct := "codicons"
	activeAppearance := "system"

	checkCfg := func(c *config.Config) {
		if c == nil {
			return
		}
		if c.Settings.Appearance != "" {
			activeAppearance = c.Settings.Appearance
		}
		if c.Settings.Theme != "" {
			th := c.Settings.Theme
			if th == "light" || th == "dark" || th == "system" {
				activeAppearance = th
				activeColor = "default"
			} else {
				activeColor = th
			}
		}
		if c.Settings.IconTheme != "" {
			activeIcon = c.Settings.IconTheme
		}
		if c.Settings.ProductIconTheme != "" {
			activeProduct = c.Settings.ProductIconTheme
		}
	}

	if globalCfg, err := config.LoadGlobal(); err == nil && globalCfg != nil {
		checkCfg(globalCfg)
	}
	cm := config.NewConfigManager()
	if cfg, _ := cm.LoadHierarchy(r.Context()); cfg != nil {
		checkCfg(cfg)
	}
	resp.Active.Color = activeColor
	resp.Active.Icon = activeIcon
	resp.Active.Product = activeProduct
	resp.Active.Appearance = activeAppearance

	_ = json.NewEncoder(w).Encode(resp)
}

func handleInstallTheme(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to parse multipart form: " + err.Error()})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "missing 'file' in form data: " + err.Error()})
		return
	}
	defer file.Close()

	items, err := themeMgr.InstallVSIX(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "install failed: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"installed": items,
	})
}

func handleDeleteTheme(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	themeType := r.URL.Query().Get("type")
	packageID := r.URL.Query().Get("id")

	if themeType == "" || packageID == "" {
		// Fallback to JSON body if provided
		var body struct {
			Type      string `json:"type"`
			PackageID string `json:"package_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if themeType == "" {
				themeType = body.Type
			}
			if packageID == "" {
				packageID = body.PackageID
			}
		}
	}

	if themeType == "" || packageID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "type and id are required"})
		return
	}

	if err := themeMgr.DeleteTheme(theme.ThemeType(themeType), packageID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func handleApplyTheme(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var body struct {
		Color      string `json:"color"`
		Icon       string `json:"icon"`
		Product    string `json:"product"`
		Appearance string `json:"appearance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body: " + err.Error()})
		return
	}

	targetFile := config.GetGlobalConfigPath()
	if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "create config dir: " + err.Error()})
		return
	}

	fileLock, lockErr := config.AcquireFileLock(targetFile, 5*time.Second)
	if lockErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "acquire config file lock: " + lockErr.Error()})
		return
	}
	defer fileLock.Release()

	existingRaw, _ := os.ReadFile(targetFile)
	content := string(existingRaw)
	var err error

	if body.Color != "" {
		content, err = config.UpsertTOMLSetting(content, []string{"settings", "theme"}, body.Color)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "upsert theme: " + err.Error()})
			return
		}
	}

	if body.Icon != "" {
		content, err = config.UpsertTOMLSetting(content, []string{"settings", "icon_theme"}, body.Icon)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "upsert icon_theme: " + err.Error()})
			return
		}
	}

	if body.Product != "" {
		content, err = config.UpsertTOMLSetting(content, []string{"settings", "product_icon_theme"}, body.Product)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "upsert product_icon_theme: " + err.Error()})
			return
		}
	}

	if body.Appearance != "" {
		content, err = config.UpsertTOMLSetting(content, []string{"settings", "appearance"}, body.Appearance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "upsert appearance: " + err.Error()})
			return
		}
	}

	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "write config file: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"color":      body.Color,
		"icon":       body.Icon,
		"product":    body.Product,
		"appearance": body.Appearance,
	})
}

func handleThemeAssets(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	rel := strings.TrimPrefix(r.URL.Path, "/api/themes/assets/")
	parts := strings.SplitN(rel, "/", 3)
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	themeType := parts[0]
	packageID := parts[1]
	subpath := parts[2]

	absPath, err := themeMgr.ResolveThemeFile(theme.ThemeType(themeType), packageID, subpath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, absPath)
}

func handleMarketplaceSearch(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query().Get("q")
	sizeStr := r.URL.Query().Get("size")
	offsetStr := r.URL.Query().Get("offset")

	size := 20
	offset := 0
	if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
		size = s
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	installedIDs := make(map[string]bool)
	if themes, err := themeMgr.ListThemes(); err == nil && themes != nil {
		for _, c := range themes.Colors {
			if c.PackageID != "" {
				installedIDs[c.PackageID] = true
			}
		}
		for _, i := range themes.Icons {
			if i.PackageID != "" {
				installedIDs[i.PackageID] = true
			}
		}
		for _, p := range themes.Products {
			if p.PackageID != "" {
				installedIDs[p.PackageID] = true
			}
		}
	}

	result, err := theme.SearchOpenVSXThemes(q, size, offset, installedIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(result)
}

func handleMarketplaceInstall(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	var body struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DownloadURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "download_url is required"})
		return
	}

	items, err := themeMgr.InstallFromURL(body.DownloadURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "online install failed: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"installed": items,
	})
}

func handleMarketplaceDetail(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "namespace and name are required"})
		return
	}

	installedIDs := make(map[string]bool)
	if themes, err := themeMgr.ListThemes(); err == nil && themes != nil {
		for _, c := range themes.Colors {
			if c.PackageID != "" {
				installedIDs[c.PackageID] = true
			}
		}
		for _, i := range themes.Icons {
			if i.PackageID != "" {
				installedIDs[i.PackageID] = true
			}
		}
		for _, p := range themes.Products {
			if p.PackageID != "" {
				installedIDs[p.PackageID] = true
			}
		}
	}

	detail, err := theme.GetOpenVSXExtensionDetail(namespace, name, installedIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(detail)
}
