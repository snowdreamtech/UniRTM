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
	"strings"
	"syscall"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/frontend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
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

func openWithSpecificBrowser(url string, targetBrowser string) error {
	switch runtime.GOOS {
	case "darwin":
		if strings.Contains(targetBrowser, "/") {
			return exec.Command(targetBrowser, url).Start()
		}
		return exec.Command("open", "-a", targetBrowser, url).Start()

	case "windows":
		return exec.Command("cmd", "/c", "start", "", targetBrowser, url).Start()

	default:
		// Linux / Unix / WSL
		if isWSL() && strings.HasSuffix(strings.ToLower(targetBrowser), ".exe") {
			return exec.Command("cmd.exe", "/c", "start", "", targetBrowser, url).Start()
		}
		return exec.Command(targetBrowser, url).Start()
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
		},
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
	if !strings.HasPrefix(storeKey, "settings.") && (storeKey == "browser" || storeKey == "cache_ttl") {
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

