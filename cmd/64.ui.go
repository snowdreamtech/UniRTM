// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
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

	"github.com/pelletier/go-toml/v2"
	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/frontend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/hook"
	"github.com/snowdreamtech/unirtm/internal/network"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/theme"
	"github.com/spf13/cobra"
	"golang.org/x/net/websocket"
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

// DetectedTerminal represents an installed shell/terminal found on the host system.
type DetectedTerminal struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDefault bool   `json:"is_default"`
}

// isRestrictedOrStubShell returns true if the shell is a restricted or non-interactive stub shell.
func isRestrictedOrStubShell(base string) bool {
	base = strings.ToLower(strings.TrimSuffix(base, ".exe"))
	switch base {
	case "rbash", "rksh", "rzsh", "rsh", "nologin", "false", "sync", "halt", "shutdown":
		return true
	default:
		return false
	}
}

// detectInstalledTerminals dynamically scans the host system for known installed shells/terminals,
// canonicalizing symlinks and hardlinks to eliminate duplicate aliases.
func detectInstalledTerminals() []DetectedTerminal {
	type termEntry struct {
		term     DetectedTerminal
		realPath string
	}
	var list []termEntry

	defaultShell := os.Getenv("SHELL")
	if defaultShell == "" && runtime.GOOS == "windows" {
		defaultShell = os.Getenv("COMSPEC")
	}
	realDefaultShell := ""
	if defaultShell != "" {
		if rp, err := filepath.EvalSymlinks(defaultShell); err == nil {
			realDefaultShell = rp
		} else {
			realDefaultShell = filepath.Clean(defaultShell)
		}
	}

	addTerminal := func(id, name, path string) {
		if path == "" {
			return
		}
		base := strings.ToLower(strings.TrimSuffix(filepath.Base(path), ".exe"))
		if isRestrictedOrStubShell(base) {
			return
		}

		fi, err := os.Stat(path)
		if err != nil {
			return
		}

		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			realPath = filepath.Clean(path)
		}

		// Check against already registered entries to eliminate symlink/hardlink duplicates
		for i, existing := range list {
			if existing.term.Path == path {
				return
			}

			sameUnderlying := (realPath != "" && existing.realPath == realPath)
			if !sameUnderlying {
				if exFi, err := os.Stat(existing.term.Path); err == nil {
					sameUnderlying = os.SameFile(fi, exFi)
				}
			}

			if sameUnderlying {
				// Same physical binary:
				// If both share the same ID, or if one is 'sh' and one is 'dash' (POSIX sh alias):
				if existing.term.ID == id || (id == "dash" && existing.term.ID == "sh") || (id == "sh" && existing.term.ID == "dash") {
					// If the newly scanned path matches the user's $SHELL exactly, prioritize this path
					if defaultShell != "" && path == defaultShell {
						list[i].term.Path = path
						list[i].term.ID = id
						list[i].term.Name = name
					}
					return
				}
			}

			// Same shell ID (e.g. "bash" found at another location)
			if existing.term.ID == id {
				if defaultShell != "" && path == defaultShell {
					list[i].term.Path = path
					list[i].term.Name = name
				}
				return
			}
		}

		list = append(list, termEntry{
			term: DetectedTerminal{
				ID:        id,
				Name:      name,
				Path:      path,
				IsDefault: false,
			},
			realPath: realPath,
		})
	}

	friendlyName := func(base string) string {
		base = strings.ToLower(strings.TrimSuffix(base, ".exe"))
		switch base {
		case "bash":
			return "Bash"
		case "zsh":
			return "Zsh"
		case "fish":
			return "Fish"
		case "sh":
			return "POSIX sh"
		case "pwsh":
			return "PowerShell Core"
		case "powershell":
			return "Windows PowerShell"
		case "cmd":
			return "Command Prompt"
		case "dash":
			return "Dash"
		case "tcsh":
			return "Tcsh"
		case "csh":
			return "Csh"
		case "tmux":
			return "Tmux"
		case "wsl":
			return "WSL"
		default:
			if len(base) > 0 {
				return strings.ToUpper(base[:1]) + base[1:]
			}
			return base
		}
	}

	// 1. On Unix-like systems, check /etc/shells
	if runtime.GOOS != "windows" {
		if content, err := os.ReadFile("/etc/shells"); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if _, err := os.Stat(line); err == nil {
					base := filepath.Base(line)
					addTerminal(strings.ToLower(base), friendlyName(base), line)
				}
			}
		}
	}

	// 2. Scan common shells via PATH
	commonShells := []string{"bash", "zsh", "fish", "sh", "pwsh", "dash"}
	if runtime.GOOS == "windows" {
		commonShells = []string{"powershell.exe", "pwsh.exe", "cmd.exe", "bash.exe", "wsl.exe"}
	}
	for _, sh := range commonShells {
		if p, err := exec.LookPath(sh); err == nil {
			base := filepath.Base(p)
			id := strings.ToLower(strings.TrimSuffix(base, ".exe"))
			addTerminal(id, friendlyName(base), p)
		}
	}

	// 3. Special checks on Windows (e.g. Git Bash, PowerShell Core, WSL, System32 fallbacks)
	if runtime.GOOS == "windows" {
		winDir := os.Getenv("SystemRoot")
		if winDir == "" {
			winDir = `C:\Windows`
		}
		systemFallbacks := []struct {
			id   string
			name string
			path string
		}{
			{"powershell", "Windows PowerShell", filepath.Join(winDir, `System32\WindowsPowerShell\v1.0\powershell.exe`)},
			{"pwsh", "PowerShell Core", `C:\Program Files\PowerShell\7\pwsh.exe`},
			{"pwsh", "PowerShell Core", `C:\Program Files\PowerShell\6\pwsh.exe`},
			{"cmd", "Command Prompt", filepath.Join(winDir, `System32\cmd.exe`)},
			{"wsl", "WSL", filepath.Join(winDir, `System32\wsl.exe`)},
		}
		for _, sf := range systemFallbacks {
			if _, err := os.Stat(sf.path); err == nil {
				addTerminal(sf.id, sf.name, sf.path)
			}
		}

		gitBashPaths := []string{
			`C:\Program Files\Git\bin\bash.exe`,
			`C:\Program Files (x86)\Git\bin\bash.exe`,
			filepath.Join(os.Getenv("LocalAppData"), `Programs\Git\bin\bash.exe`),
			filepath.Join(os.Getenv("UserProfile"), `AppData\Local\Programs\Git\bin\bash.exe`),
		}
		for _, gbp := range gitBashPaths {
			if gbp != "" {
				if _, err := os.Stat(gbp); err == nil {
					addTerminal("git-bash", "Git Bash", gbp)
					break
				}
			}
		}
	}

	// 4. Ensure at least defaultShell is present if specified and exists
	if defaultShell != "" {
		if _, err := os.Stat(defaultShell); err == nil {
			base := filepath.Base(defaultShell)
			addTerminal(strings.ToLower(strings.TrimSuffix(base, ".exe")), friendlyName(base), defaultShell)
		}
	}

	// 5. Resolve is_default: Exactly ONE terminal must be marked as default
	defaultIndex := -1

	// 5a. Prioritize exact path match with $SHELL
	if defaultShell != "" {
		for i, item := range list {
			if item.term.Path == defaultShell {
				defaultIndex = i
				break
			}
		}
	}

	// 5b. Prioritize realPath match with realDefaultShell
	if defaultIndex == -1 && realDefaultShell != "" {
		for i, item := range list {
			if item.realPath == realDefaultShell {
				defaultIndex = i
				break
			}
		}
	}

	// 5c. Match base name of default shell
	if defaultIndex == -1 && defaultShell != "" {
		defBase := strings.ToLower(strings.TrimSuffix(filepath.Base(defaultShell), ".exe"))
		for i, item := range list {
			if item.term.ID == defBase {
				defaultIndex = i
				break
			}
		}
	}

	// 5d. Fallback to first available terminal
	if defaultIndex == -1 && len(list) > 0 {
		defaultIndex = 0
	}

	var results []DetectedTerminal
	for i, item := range list {
		t := item.term
		t.IsDefault = (i == defaultIndex)
		results = append(results, t)
	}

	// Ensure system default terminal is strictly sorted to the top
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].IsDefault && !results[j].IsDefault {
			return true
		}
		if !results[i].IsDefault && results[j].IsDefault {
			return false
		}
		return i < j
	})

	return results
}

func handleGetTerminals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	terms := detectInstalledTerminals()

	defPath := ""
	for _, t := range terms {
		if t.IsDefault {
			defPath = t.Path
			break
		}
	}
	if defPath == "" && len(terms) > 0 {
		defPath = terms[0].Path
	}

	curPath := defPath
	cfg, err := config.LoadGlobal()
	if err == nil && cfg != nil && cfg.Settings.Terminal != "" {
		configured := cfg.Settings.Terminal
		for _, t := range terms {
			if t.Path == configured || t.ID == configured {
				curPath = t.Path
				break
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"default":   defPath,
		"current":   curPath,
		"terminals": terms,
	})
}

type TerminalExecRequest struct {
	Command string `json:"command"`
	Shell   string `json:"shell"`
	Cwd     string `json:"cwd"`
}

type TerminalExecResponse struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

func handleTerminalExec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TerminalExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmdStr := strings.TrimSpace(req.Command)
	if cmdStr == "" {
		_ = json.NewEncoder(w).Encode(TerminalExecResponse{
			Output:   "",
			ExitCode: 0,
		})
		return
	}

	shellPath := strings.TrimSpace(req.Shell)
	if shellPath == "" {
		if s := os.Getenv("SHELL"); s != "" {
			shellPath = s
		} else if runtime.GOOS == "windows" {
			shellPath = "cmd.exe"
		} else {
			shellPath = "/bin/sh"
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" && strings.Contains(strings.ToLower(shellPath), "cmd") {
		cmd = exec.Command(shellPath, "/c", cmdStr)
	} else if runtime.GOOS == "windows" && strings.Contains(strings.ToLower(shellPath), "powershell") {
		cmd = exec.Command(shellPath, "-NoProfile", "-Command", cmdStr)
	} else {
		cmd = exec.Command(shellPath, "-c", cmdStr)
	}

	if req.Cwd != "" {
		cmd.Dir = req.Cwd
	} else {
		if pwd, err := os.Getwd(); err == nil {
			cmd.Dir = pwd
		}
	}

	// Environment: inherit system environment and prepend UniRTM shims and bin
	env := os.Environ()
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		unirtmBin := filepath.Join(homeDir, ".unirtm", "bin")
		unirtmShims := filepath.Join(homeDir, ".unirtm", "shims")
		pathKeyFound := false
		for i, e := range env {
			if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
				origPath := e[5:]
				env[i] = fmt.Sprintf("PATH=%s%c%s%c%s", unirtmShims, os.PathListSeparator, unirtmBin, os.PathListSeparator, origPath)
				pathKeyFound = true
				break
			}
		}
		if !pathKeyFound {
			env = append(env, fmt.Sprintf("PATH=%s%c%s", unirtmShims, os.PathListSeparator, unirtmBin))
		}
	}
	cmd.Env = env

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	exitCode := 0
	errMsg := ""
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
			errMsg = err.Error()
		}
	}

	_ = json.NewEncoder(w).Encode(TerminalExecResponse{
		Output:   outBuf.String(),
		ExitCode: exitCode,
		Error:    errMsg,
	})
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
	mux.HandleFunc("/api/terminals", handleGetTerminals)
	mux.HandleFunc("/api/terminal/exec", handleTerminalExec)
	mux.Handle("/api/terminal/ws", websocket.Handler(handleTerminalWebSocket))

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

	// ─── File System API Handlers ─────────────────────────────────────────────
	mux.HandleFunc("/api/fs/tree", handleFSTree)
	mux.HandleFunc("/api/fs/create", handleFSCreate)
	mux.HandleFunc("/api/fs/rename", handleFSRename)
	mux.HandleFunc("/api/fs/delete", handleFSDelete)
	mux.HandleFunc("/api/fs/read", handleFSRead)
	mux.HandleFunc("/api/fs/write", handleFSWrite)

	// ─── Git SCM API Handlers ─────────────────────────────────────────────────
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
	if wsCfg.MinVersion == "" && wsRaw != "" {
		var doc struct {
			MinVersion string `toml:"min_version"`
		}
		if err := toml.Unmarshal([]byte(wsRaw), &doc); err == nil && doc.MinVersion != "" {
			wsCfg.MinVersion = doc.MinVersion
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"path":        userPath,
			"exists":      userExists,
			"raw":         userRaw,
			"min_version": userCfg.MinVersion,
			"settings":    userCfg.Settings,
			"plugins":     userCfg.Plugins,
			"env":         userCfg.Env,
			"hooks":       userCfg.Hooks,
			"aliases":     userCfg.Aliases,
		},
		"workspace": map[string]interface{}{
			"path":        wsPath,
			"exists":      wsExists,
			"raw":         wsRaw,
			"min_version": wsCfg.MinVersion,
			"settings":    wsCfg.Settings,
			"plugins":     wsCfg.Plugins,
			"env":         wsCfg.Env,
			"tools":       wsCfg.Tools,
			"tasks":       wsCfg.Tasks,
			"hooks":       wsCfg.Hooks,
			"aliases":     wsCfg.Aliases,
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
		Task  string `json:"task"`
		Scope string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request body"})
		return
	}
	taskName := strings.TrimSpace(body.Name)
	if taskName == "" {
		taskName = strings.TrimSpace(body.Task)
	}
	if taskName == "" || taskName == "undefined" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "task name is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "run", taskName)
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
	profile := strings.TrimSpace(r.URL.Query().Get("profile"))
	if scope == "" {
		scope = "workspace"
	}

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)
	baseEnvMap := config.GetEnvVarsFromContent(content)
	profiles := config.GetEnvironmentProfiles(content)

	type EnvVarItem struct {
		Name          string `json:"name"`
		Value         string `json:"value"`
		Scope         string `json:"scope"`
		Profile       string `json:"profile,omitempty"`
		IsSensitive   bool   `json:"is_sensitive"`
		IsOverride    bool   `json:"is_override"`
		BaseValue     string `json:"base_value,omitempty"`
		IsProfileOnly bool   `json:"is_profile_only"`
		IsInherited   bool   `json:"is_inherited"`
	}

	var res []EnvVarItem

	if profile == "" || scope == "user" {
		var names []string
		for n := range baseEnvMap {
			names = append(names, n)
		}
		sort.Strings(names)

		for _, n := range names {
			res = append(res, EnvVarItem{
				Name:        n,
				Value:       baseEnvMap[n],
				Scope:       scope,
				Profile:     "",
				IsSensitive: isSensitiveEnvKey(n),
			})
		}
	} else {
		// Specific workspace profile
		envs := config.GetEnvironmentsFromContent(content)
		var profEnvRaw map[string]interface{}
		if cfg, exists := envs[profile]; exists {
			profEnvRaw = cfg.Env
		}
		profEnvMap := make(map[string]string)
		for k, v := range profEnvRaw {
			profEnvMap[strings.ToUpper(k)] = fmt.Sprintf("%v", v)
		}

		// Union of profile keys and base keys
		allKeysMap := make(map[string]bool)
		for k := range profEnvMap {
			allKeysMap[k] = true
		}
		for k := range baseEnvMap {
			allKeysMap[k] = true
		}
		var names []string
		for k := range allKeysMap {
			names = append(names, k)
		}
		sort.Strings(names)

		for _, n := range names {
			profVal, inProf := profEnvMap[n]
			baseVal, inBase := baseEnvMap[n]

			if inProf {
				if inBase {
					res = append(res, EnvVarItem{
						Name:        n,
						Value:       profVal,
						Scope:       scope,
						Profile:     profile,
						IsSensitive: isSensitiveEnvKey(n),
						IsOverride:  true,
						BaseValue:   baseVal,
					})
				} else {
					res = append(res, EnvVarItem{
						Name:          n,
						Value:         profVal,
						Scope:         scope,
						Profile:       profile,
						IsSensitive:   isSensitiveEnvKey(n),
						IsProfileOnly: true,
					})
				}
			} else {
				res = append(res, EnvVarItem{
					Name:        n,
					Value:       baseVal,
					Scope:       scope,
					Profile:     profile,
					IsSensitive: isSensitiveEnvKey(n),
					IsInherited: true,
					BaseValue:   baseVal,
				})
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"env":      res,
		"total":    len(res),
		"scope":    scope,
		"profile":  profile,
		"profiles": profiles,
	})
}

type UpsertEnvVarReq struct {
	Key         string `json:"key"`
	OriginalKey string `json:"original_key,omitempty"`
	Value       string `json:"value"`
	Scope       string `json:"scope,omitempty"`
	Profile     string `json:"profile,omitempty"`
	Action      string `json:"action,omitempty"`
}

func handleUpsertEnvVar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req UpsertEnvVarReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body: " + err.Error()})
		return
	}

	scope := strings.ToLower(strings.TrimSpace(req.Scope))
	if scope == "" {
		scope = "workspace"
	}
	profile := strings.TrimSpace(req.Profile)

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)

	// Special action: create profile without adding variables yet
	if req.Action == "create_profile" {
		if scope == "user" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "profiles are only supported in workspace scope"})
			return
		}
		if profile == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "profile name cannot be empty"})
			return
		}
		updated, err := config.CreateEnvironmentProfile(content, profile)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"profile": profile,
		})
		return
	}

	if strings.TrimSpace(req.Key) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "variable key is required"})
		return
	}

	if profile != "" && scope == "workspace" {
		if req.OriginalKey != "" && !strings.EqualFold(req.OriginalKey, req.Key) {
			content, _ = config.UnsetEnvironmentEnvVar(content, profile, req.OriginalKey)
		}
		updated, err := config.UpsertEnvironmentEnvVar(content, profile, req.Key, req.Value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
			return
		}
	} else {
		if req.OriginalKey != "" && !strings.EqualFold(req.OriginalKey, req.Key) {
			content, _ = config.UnsetEnvVar(content, req.OriginalKey)
		}
		updated := config.UpsertEnvVar(content, req.Key, req.Value)
		if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
			return
		}
	}

	upperKey := strings.ToUpper(strings.TrimSpace(req.Key))
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"env": map[string]interface{}{
			"name":         upperKey,
			"value":        req.Value,
			"scope":        scope,
			"profile":      profile,
			"is_sensitive": isSensitiveEnvKey(upperKey),
		},
	})
}

func handleDeleteEnvVar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := strings.TrimSpace(r.URL.Query().Get("key"))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	profile := strings.TrimSpace(r.URL.Query().Get("profile"))
	action := strings.TrimSpace(r.URL.Query().Get("action"))

	if key == "" && action == "" {
		var body struct {
			Key     string `json:"key"`
			Scope   string `json:"scope"`
			Profile string `json:"profile"`
			Action  string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			key = strings.TrimSpace(body.Key)
			if scope == "" {
				scope = strings.ToLower(strings.TrimSpace(body.Scope))
			}
			if profile == "" {
				profile = strings.TrimSpace(body.Profile)
			}
			if action == "" {
				action = strings.TrimSpace(body.Action)
			}
		}
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

	// Delete entire profile
	if action == "delete_profile" && profile != "" {
		updated, removed := config.RemoveEnvironmentProfile(content, profile)
		if removed {
			if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
				return
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":         true,
			"deleted_profile": profile,
		})
		return
	}

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "variable key is required"})
		return
	}

	var updated string
	var removed bool
	if profile != "" && scope == "workspace" {
		updated, removed = config.UnsetEnvironmentEnvVar(content, profile, key)
	} else {
		updated, removed = config.UnsetEnvVar(content, key)
	}

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
		"profile":     profile,
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

	// Machine-specific settings check: browser, editor and terminal are global-only
	cleanKey := strings.TrimPrefix(key, "settings.")
	if scope == "workspace" && (cleanKey == "browser" || cleanKey == "editor" || cleanKey == "terminal") {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("%s configuration is only permitted in global config to ensure project portability across team machines.", cleanKey),
			"hint":  fmt.Sprintf("Please configure settings.%s in User (全局) scope instead.", cleanKey),
		})
		return
	}

	// Workspace-specific settings check: min_version is workspace-only
	if scope == "user" && key == "min_version" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "min_version configuration is only permitted in workspace config to ensure project compatibility.",
			"hint":  "Please configure min_version in Workspace (工作区) scope instead.",
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
	var updatedTOML string
	if key == "min_version" && strings.TrimSpace(valStr) == "" {
		var removed bool
		updatedTOML, removed = config.RemoveRootField(string(existingRaw), "min_version")
		if !removed {
			updatedTOML = string(existingRaw)
		}
	} else {
		var err error
		updatedTOML, err = config.UpsertTOMLSetting(string(existingRaw), config.SplitConfigKey(storeKey), valStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "update toml setting: " + err.Error()})
			return
		}
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

// ─── File System API Handlers ────────────────────────────────────────────────

type FSEntry struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"is_dir"`
	Size      int64     `json:"size"`
	ModTime   string    `json:"mod_time"`
	IsIgnored bool      `json:"is_ignored"`
	Children  []FSEntry `json:"children,omitempty"`
}

type FSTreeResponse struct {
	Success  bool      `json:"success"`
	Root     string    `json:"root"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Entries  []FSEntry `json:"entries"`
	Children []FSEntry `json:"children"`
}

// getSafeFSPath resolves a relative path within the current working directory,
// preventing path traversal attacks.
func getSafeFSPath(relPath string) (rootDir string, safePath string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	cleanRel := filepath.Clean(filepath.FromSlash(strings.TrimSpace(relPath)))
	if cleanRel == "." || cleanRel == "/" || cleanRel == "\\" {
		cleanRel = ""
	}
	target := filepath.Join(wd, cleanRel)
	rel, err := filepath.Rel(wd, target)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return "", "", fmt.Errorf("access denied: path outside workspace")
	}
	return wd, target, nil
}

func isCommonIgnored(name string) bool {
	switch name {
	case ".git", "node_modules", ".DS_Store", "dist", "target", ".turbo", ".idea", ".vscode", ".unirtm":
		return true
	default:
		return false
	}
}

func buildFSTree(wd, dirPath, currentRelPath string, currentDepth, maxDepth int, showHidden bool) ([]FSEntry, error) {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var dirs []FSEntry
	var files []FSEntry

	for _, d := range dirEntries {
		name := d.Name()
		if !showHidden && strings.HasPrefix(name, ".") && name != ".unirtm.toml" && name != ".gitignore" && name != ".env" {
			continue
		}

		entryRelPath := filepath.ToSlash(filepath.Join(currentRelPath, name))
		fullEntryPath := filepath.Join(dirPath, name)
		fi, err := d.Info()
		size := int64(0)
		modTime := ""
		if err == nil {
			size = fi.Size()
			modTime = fi.ModTime().UTC().Format(time.RFC3339)
		}

		isDir := d.IsDir()
		isIgnored := isCommonIgnored(name)

		entry := FSEntry{
			Name:      name,
			Path:      entryRelPath,
			IsDir:     isDir,
			Size:      size,
			ModTime:   modTime,
			IsIgnored: isIgnored,
		}

		if isDir {
			if currentDepth < maxDepth && name != ".git" && name != "node_modules" {
				subChildren, _ := buildFSTree(wd, fullEntryPath, entryRelPath, currentDepth+1, maxDepth, showHidden)
				if subChildren == nil {
					subChildren = []FSEntry{}
				}
				entry.Children = subChildren
			}
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return append(dirs, files...), nil
}

func handleFSTree(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqPath := r.URL.Query().Get("path")
	depthStr := r.URL.Query().Get("depth")
	showHiddenStr := r.URL.Query().Get("showHidden")

	depth := 3
	if d, err := strconv.Atoi(depthStr); err == nil && d > 0 && d <= 10 {
		depth = d
	}
	showHidden := showHiddenStr == "true"

	wd, targetDir, err := getSafeFSPath(reqPath)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	st, err := os.Stat(targetDir)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "directory not found"})
		return
	}
	if !st.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "path is not a directory"})
		return
	}

	cleanRel, _ := filepath.Rel(wd, targetDir)
	if cleanRel == "." {
		cleanRel = ""
	}

	entries, err := buildFSTree(wd, targetDir, cleanRel, 1, depth, showHidden)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	resp := FSTreeResponse{
		Success:  true,
		Root:     wd,
		Name:     filepath.Base(wd),
		Path:     cleanRel,
		Entries:  entries,
		Children: entries,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleFSCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Path  string `json:"path"`
		IsDir bool   `json:"is_dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Path) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "path is required"})
		return
	}

	_, targetPath, err := getSafeFSPath(body.Path)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if body.IsDir {
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
	} else {
		parent := filepath.Dir(targetPath)
		if err := os.MkdirAll(parent, 0755); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		f, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			if os.IsExist(err) {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "file already exists"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		_ = f.Close()
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "path": body.Path})
}

func handleFSRename(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.OldPath) == "" || strings.TrimSpace(body.NewPath) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "old_path and new_path are required"})
		return
	}

	_, oldTarget, err := getSafeFSPath(body.OldPath)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_, newTarget, err := getSafeFSPath(body.NewPath)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if err := os.Rename(oldTarget, newTarget); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "old_path": body.OldPath, "new_path": body.NewPath})
}

func handleFSDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Path string `json:"path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Path == "" {
		body.Path = r.URL.Query().Get("path")
	}
	if strings.TrimSpace(body.Path) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "path is required"})
		return
	}

	wd, target, err := getSafeFSPath(body.Path)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if target == wd {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "cannot delete workspace root"})
		return
	}

	if err := os.RemoveAll(target); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "path": body.Path})
}

func handleFSRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqPath := r.URL.Query().Get("path")
	if strings.TrimSpace(reqPath) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "path is required"})
		return
	}

	_, target, err := getSafeFSPath(reqPath)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	st, err := os.Stat(target)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "file not found"})
		return
	}
	if st.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "path is a directory, not a file"})
		return
	}
	if st.Size() > 10*1024*1024 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":        "file too large (exceeds 10MB)",
			"size":         st.Size(),
			"is_too_large": true,
		})
		return
	}

	data, err := os.ReadFile(target)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"path":    reqPath,
		"content": string(data),
		"size":    st.Size(),
	})
}

func handleFSWrite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Path) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "path is required"})
		return
	}

	_, target, err := getSafeFSPath(body.Path)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if err := os.WriteFile(target, []byte(body.Content), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "path": body.Path, "size": len(body.Content)})
}

// ─── Git SCM Handlers Implementation ───────────────────────────────────────

type GitChangeItem struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type GitStatusResponse struct {
	Success      bool            `json:"success"`
	Branch       string          `json:"branch"`
	Upstream     string          `json:"upstream"`
	Ahead        int             `json:"ahead"`
	Behind       int             `json:"behind"`
	Staged       []GitChangeItem `json:"staged"`
	Unstaged     []GitChangeItem `json:"unstaged"`
	Untracked    []GitChangeItem `json:"untracked"`
	TotalChanges int             `json:"totalChanges"`
}

func runGitCommand(args ...string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = wd
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		if errStr == "" {
			errStr = err.Error()
		}
		return "", errors.New(strings.TrimSpace(errStr))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func handleGitStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	porcelain, err := runGitCommand("status", "--porcelain=v1", "-b")
	if err != nil {
		// Not a git repository or git error
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      false,
			"error":        err.Error(),
			"isGitRepo":    false,
			"branch":       "",
			"totalChanges": 0,
		})
		return
	}

	lines := strings.Split(porcelain, "\n")
	branch := "HEAD"
	upstream := ""
	ahead := 0
	behind := 0

	staged := []GitChangeItem{}
	unstaged := []GitChangeItem{}
	untracked := []GitChangeItem{}

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "## ") {
			header := strings.TrimPrefix(line, "## ")
			header = strings.TrimSpace(header)
			if strings.Contains(header, "...") {
				parts := strings.Split(header, "...")
				branch = parts[0]
				rest := parts[1]
				if idx := strings.Index(rest, " "); idx != -1 {
					upstream = rest[:idx]
					statusPart := rest[idx:]
					if strings.Contains(statusPart, "ahead") {
						var a int
						if n, _ := fmt.Sscanf(statusPart[strings.Index(statusPart, "ahead")+5:], "%d", &a); n > 0 {
							ahead = a
						}
					}
					if strings.Contains(statusPart, "behind") {
						var b int
						if n, _ := fmt.Sscanf(statusPart[strings.Index(statusPart, "behind")+6:], "%d", &b); n > 0 {
							behind = b
						}
					}
				} else {
					upstream = rest
				}
			} else {
				branch = header
			}
			continue
		}

		if len(line) < 3 {
			continue
		}

		x := line[0]
		y := line[1]
		fPath := strings.TrimSpace(line[3:])
		fPath = strings.Trim(fPath, "\"")

		if x == '?' && y == '?' {
			untracked = append(untracked, GitChangeItem{Path: fPath, Status: "U"})
		} else {
			if x != ' ' && x != '?' {
				staged = append(staged, GitChangeItem{Path: fPath, Status: string(x)})
			}
			if y != ' ' && y != '?' {
				unstaged = append(unstaged, GitChangeItem{Path: fPath, Status: string(y)})
			}
		}
	}

	resp := GitStatusResponse{
		Success:      true,
		Branch:       branch,
		Upstream:     upstream,
		Ahead:        ahead,
		Behind:       behind,
		Staged:       staged,
		Unstaged:     unstaged,
		Untracked:    untracked,
		TotalChanges: len(staged) + len(unstaged) + len(untracked),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleGitStage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Paths []string `json:"paths"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	args := []string{"add", "--"}
	if len(body.Paths) == 0 || (len(body.Paths) == 1 && (body.Paths[0] == "." || body.Paths[0] == "*")) {
		args = []string{"add", "-A"}
	} else {
		args = append(args, body.Paths...)
	}

	if _, err := runGitCommand(args...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleGitUnstage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Paths []string `json:"paths"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	args := []string{"reset", "HEAD", "--"}
	if len(body.Paths) > 0 && !(len(body.Paths) == 1 && (body.Paths[0] == "." || body.Paths[0] == "*")) {
		args = append(args, body.Paths...)
	}

	if _, err := runGitCommand(args...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleGitDiscard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Paths []string `json:"paths"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	for _, p := range body.Paths {
		if _, err := runGitCommand("checkout", "--", p); err != nil {
			_, _ = runGitCommand("clean", "-fd", "--", p)
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleGitCommit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Message string `json:"message"`
		Amend   bool   `json:"amend"`
		Sign    *bool  `json:"sign"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Message) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Commit message is required"})
		return
	}

	args := []string{"commit", "-m", strings.TrimSpace(body.Message)}
	if body.Amend {
		args = []string{"commit", "--amend", "-m", strings.TrimSpace(body.Message)}
	}
	if body.Sign != nil {
		if *body.Sign {
			args = append(args, "-S")
		} else {
			args = append(args, "--no-gpg-sign")
		}
	}

	out, err := runGitCommand(args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("path")
	isStaged := r.URL.Query().Get("staged") == "true"

	original := ""
	modified := ""

	if isStaged {
		orig, _ := runGitCommand("show", "HEAD:"+filePath)
		original = orig
		mod, _ := runGitCommand("show", ":"+filePath)
		modified = mod
	} else {
		orig, err := runGitCommand("show", ":"+filePath)
		if err != nil {
			orig, _ = runGitCommand("show", "HEAD:"+filePath)
		}
		original = orig

		wd, safePath, err := getSafeFSPath(filePath)
		if err == nil && wd != "" {
			data, _ := os.ReadFile(safePath)
			modified = string(data)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"path":     filePath,
		"original": original,
		"modified": modified,
	})
}

type GitBranchItem struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"isCurrent"`
	IsRemote  bool   `json:"isRemote"`
}

func handleGitBranches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	out, err := runGitCommand("branch", "-a", "--format=%(refname:short)|%(HEAD)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	lines := strings.Split(out, "\n")
	current := ""
	branches := []GitBranchItem{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		bName := parts[0]
		isHead := len(parts) > 1 && parts[1] == "*"
		if isHead {
			current = bName
		}
		branches = append(branches, GitBranchItem{
			Name:      bName,
			IsCurrent: isHead,
			IsRemote:  strings.HasPrefix(bName, "origin/"),
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"current":  current,
		"branches": branches,
	})
}

func handleGitCheckout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Branch string `json:"branch"`
		Create bool   `json:"create"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Branch) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch name is required"})
		return
	}

	args := []string{"checkout", body.Branch}
	if body.Create {
		args = []string{"checkout", "-b", body.Branch}
	}

	if _, err := runGitCommand(args...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "branch": body.Branch})
}

func handleGitSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pullOut, _ := runGitCommand("pull", "--ff-only")
	pushOut, _ := runGitCommand("push")

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"pull":    pullOut,
		"push":    pushOut,
	})
}

type GitGpgInfo struct {
	Status      string `json:"status"` // "verified" | "unverified" | "bad" | "expired" | "revoked" | "none"
	KeyID       string `json:"keyId,omitempty"`
	Signer      string `json:"signer,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

func parseGpgInfo(rawStatus, keyID, signer, fingerprint string) *GitGpgInfo {
	s := strings.ToUpper(strings.TrimSpace(rawStatus))
	status := "none"
	switch s {
	case "G":
		status = "verified"
	case "U":
		status = "unverified"
	case "B":
		status = "bad"
	case "X", "Y":
		status = "expired"
	case "R":
		status = "revoked"
	default:
		status = "none"
	}
	return &GitGpgInfo{
		Status:      status,
		KeyID:       strings.TrimSpace(keyID),
		Signer:      strings.TrimSpace(signer),
		Fingerprint: strings.TrimSpace(fingerprint),
	}
}

type GitCommitItem struct {
	Hash      string      `json:"hash"`
	ShortHash string      `json:"shortHash"`
	Author    string      `json:"author"`
	Email     string      `json:"email"`
	Timestamp int64       `json:"timestamp"`
	Message   string      `json:"message"`
	Gpg       *GitGpgInfo `json:"gpg,omitempty"`
}

func handleGitLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	out, err := runGitCommand("log", "-n", "30", "--pretty=format:%H%x09%h%x09%an%x09%ae%x09%at%x09%s%x09%G?%x09%GK%x09%GS%x09%GF")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	lines := strings.Split(out, "\n")
	commits := []GitCommitItem{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 6 {
			var ts int64
			_, _ = fmt.Sscanf(parts[4], "%d", &ts)
			var gpg *GitGpgInfo
			if len(parts) >= 10 {
				gpg = parseGpgInfo(parts[6], parts[7], parts[8], parts[9])
			} else {
				gpg = parseGpgInfo("", "", "", "")
			}
			commits = append(commits, GitCommitItem{
				Hash:      parts[0],
				ShortHash: parts[1],
				Author:    parts[2],
				Email:     parts[3],
				Timestamp: ts,
				Message:   parts[5],
				Gpg:       gpg,
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"commits": commits,
	})
}

type GitGraphRef struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // "head" | "branch" | "remote" | "tag"
	Current bool   `json:"current"`
}

type GitGraphCommit struct {
	Hash      string        `json:"hash"`
	ShortHash string        `json:"shortHash"`
	Parents   []string      `json:"parents"`
	Author    string        `json:"author"`
	Email     string        `json:"email"`
	Timestamp int64         `json:"timestamp"`
	Message   string        `json:"message"`
	Refs      []GitGraphRef `json:"refs"`
	Gpg       *GitGpgInfo   `json:"gpg,omitempty"`
}

func handleGitGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	branch := r.URL.Query().Get("branch")
	filePath := r.URL.Query().Get("path")
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "200"
	}

	args := []string{"log", "--date-order", "-n", limitStr, "--pretty=format:%H%x09%h%x09%P%x09%an%x09%ae%x09%at%x09%s%x09%D%x09%G?%x09%GK%x09%GS%x09%GF"}
	if branch != "" && branch != "all" {
		args = append(args, branch)
	} else {
		args = append(args, "--all")
	}
	if filePath != "" {
		args = append(args, "--", filePath)
	}

	out, err := runGitCommand(args...)
	if err != nil {
		out = ""
	}

	lines := strings.Split(out, "\n")
	commits := []GitGraphCommit{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 7 {
			var ts int64
			_, _ = fmt.Sscanf(parts[5], "%d", &ts)
			parentsStr := strings.TrimSpace(parts[2])
			parents := []string{}
			if parentsStr != "" {
				parents = strings.Fields(parentsStr)
			}

			refsStr := ""
			if len(parts) >= 8 {
				refsStr = parts[7]
			}
			var refs []GitGraphRef
			if refsStr != "" {
				for _, raw := range strings.Split(refsStr, ",") {
					raw = strings.TrimSpace(raw)
					if raw == "" {
						continue
					}
					if strings.HasPrefix(raw, "HEAD -> ") {
						refs = append(refs, GitGraphRef{Name: strings.TrimPrefix(raw, "HEAD -> "), Type: "branch", Current: true})
					} else if raw == "HEAD" {
						refs = append(refs, GitGraphRef{Name: "HEAD", Type: "head", Current: true})
					} else if strings.HasPrefix(raw, "tag: ") {
						refs = append(refs, GitGraphRef{Name: strings.TrimPrefix(raw, "tag: "), Type: "tag", Current: false})
					} else if strings.HasPrefix(raw, "refs/tags/") {
						refs = append(refs, GitGraphRef{Name: strings.TrimPrefix(raw, "refs/tags/"), Type: "tag", Current: false})
					} else if strings.Contains(raw, "/") {
						refs = append(refs, GitGraphRef{Name: raw, Type: "remote", Current: false})
					} else {
						refs = append(refs, GitGraphRef{Name: raw, Type: "branch", Current: false})
					}
				}
			}

			var gpg *GitGpgInfo
			if len(parts) >= 12 {
				gpg = parseGpgInfo(parts[8], parts[9], parts[10], parts[11])
			} else {
				gpg = parseGpgInfo("", "", "", "")
			}

			commits = append(commits, GitGraphCommit{
				Hash:      parts[0],
				ShortHash: parts[1],
				Parents:   parents,
				Author:    parts[3],
				Email:     parts[4],
				Timestamp: ts,
				Message:   parts[6],
				Refs:      refs,
				Gpg:       gpg,
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"commits": commits,
	})
}

type GitCommitFileItem struct {
	Path      string `json:"path"`
	OldPath   string `json:"oldPath,omitempty"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

func handleGitCommitDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	hash := r.URL.Query().Get("hash")
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Hash is required"})
		return
	}

	metaOut, err := runGitCommand("show", "-s", "--format=%H%x09%h%x09%P%x09%an%x09%ae%x09%at%x09%cn%x09%ce%x09%ct%x09%G?%x09%GK%x09%GS%x09%GF%x09%B", hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	metaParts := strings.Split(metaOut, "\t")
	var fullHash, shortHash, author, email, committer, committerEmail, fullBody string
	var ts, cTs int64
	var parents []string
	var gpg *GitGpgInfo

	if len(metaParts) >= 14 {
		fullHash = metaParts[0]
		shortHash = metaParts[1]
		if pStr := strings.TrimSpace(metaParts[2]); pStr != "" {
			parents = strings.Fields(pStr)
		}
		author = metaParts[3]
		email = metaParts[4]
		_, _ = fmt.Sscanf(metaParts[5], "%d", &ts)
		committer = metaParts[6]
		committerEmail = metaParts[7]
		_, _ = fmt.Sscanf(metaParts[8], "%d", &cTs)
		gpg = parseGpgInfo(metaParts[9], metaParts[10], metaParts[11], metaParts[12])
		fullBody = strings.Join(metaParts[13:], "\t")
	} else if len(metaParts) >= 10 {
		fullHash = metaParts[0]
		shortHash = metaParts[1]
		if pStr := strings.TrimSpace(metaParts[2]); pStr != "" {
			parents = strings.Fields(pStr)
		}
		author = metaParts[3]
		email = metaParts[4]
		_, _ = fmt.Sscanf(metaParts[5], "%d", &ts)
		committer = metaParts[6]
		committerEmail = metaParts[7]
		_, _ = fmt.Sscanf(metaParts[8], "%d", &cTs)
		fullBody = metaParts[9]
		gpg = parseGpgInfo("", "", "", "")
	}

	nameStatusOut, _ := runGitCommand("diff-tree", "--no-commit-id", "--name-status", "-r", "-m", "--first-parent", hash)
	numstatOut, _ := runGitCommand("show", "--numstat", "--pretty=", hash)

	numstatMap := make(map[string][2]int)
	for _, line := range strings.Split(numstatOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			var a, d int
			_, _ = fmt.Sscanf(parts[0], "%d", &a)
			_, _ = fmt.Sscanf(parts[1], "%d", &d)
			numstatMap[strings.TrimSpace(parts[2])] = [2]int{a, d}
		}
	}

	var files []GitCommitFileItem
	for _, line := range strings.Split(nameStatusOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			statusCode := parts[0]
			p1 := parts[1]
			p2 := ""
			if len(parts) >= 3 {
				p2 = parts[2]
			}
			finalPath := p1
			oldPath := ""
			if p2 != "" {
				oldPath = p1
				finalPath = p2
			}
			st := "M"
			if len(statusCode) > 0 {
				st = strings.ToUpper(string(statusCode[0]))
			}
			stat := numstatMap[finalPath]
			files = append(files, GitCommitFileItem{
				Path:      finalPath,
				OldPath:   oldPath,
				Status:    st,
				Additions: stat[0],
				Deletions: stat[1],
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"commit": map[string]interface{}{
			"hash":               fullHash,
			"shortHash":          shortHash,
			"parents":            parents,
			"author":             author,
			"email":              email,
			"timestamp":          ts,
			"committer":          committer,
			"committerEmail":     committerEmail,
			"committerTimestamp": cTs,
			"message":            strings.TrimSpace(fullBody),
			"gpg":                gpg,
			"files":              files,
		},
	})
}

func handleGitCommitDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	hash := r.URL.Query().Get("hash")
	filePath := r.URL.Query().Get("path")
	parent := r.URL.Query().Get("parent")

	if hash == "" || filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "hash and path are required"})
		return
	}

	if parent == "" {
		p, err := runGitCommand("rev-parse", hash+"^")
		if err == nil {
			parent = strings.TrimSpace(p)
		}
	}

	original := ""
	if parent != "" {
		orig, err := runGitCommand("show", parent+":"+filePath)
		if err == nil {
			original = orig
		}
	}

	modified := ""
	mod, err := runGitCommand("show", hash+":"+filePath)
	if err == nil {
		modified = mod
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"path":     filePath,
		"hash":     hash,
		"parent":   parent,
		"original": original,
		"modified": modified,
	})
}

func handleGitTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Name    string `json:"name"`
		Hash    string `json:"hash"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Tag name and commit hash are required"})
		return
	}

	var err error
	if body.Message != "" {
		_, err = runGitCommand("tag", "-a", body.Name, "-m", body.Message, body.Hash)
	} else {
		_, err = runGitCommand("tag", body.Name, body.Hash)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name})
}

func handleGitBranchCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Name     string `json:"name"`
		Hash     string `json:"hash"`
		Checkout bool   `json:"checkout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch name and commit hash are required"})
		return
	}

	if _, err := runGitCommand("branch", body.Name, body.Hash); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if body.Checkout {
		if _, err := runGitCommand("checkout", body.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name, "checkout": body.Checkout})
}

func handleGitCherryPick(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Commit hash is required"})
		return
	}

	out, err := runGitCommand("cherry-pick", body.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitRevert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Commit hash is required"})
		return
	}

	out, err := runGitCommand("revert", "--no-edit", body.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Hash string `json:"hash"`
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Commit hash is required"})
		return
	}

	mode := body.Mode
	if mode != "soft" && mode != "hard" {
		mode = "mixed"
	}

	out, err := runGitCommand("reset", "--"+mode, body.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "mode": mode, "output": out})
}

func handleGitBranchRename(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		OldName string `json:"oldName"`
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OldName == "" || body.NewName == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Both oldName and newName are required"})
		return
	}
	if _, err := runGitCommand("branch", "-m", body.OldName, body.NewName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "oldName": body.OldName, "newName": body.NewName})
}

func handleGitBranchDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name  string `json:"name"`
		Force bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch name is required"})
		return
	}
	flag := "-d"
	if body.Force {
		flag = "-D"
	}
	if _, err := runGitCommand("branch", flag, body.Name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name})
}

func handleGitBranchMerge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Branch string `json:"branch"`
		NoFf   bool   `json:"noFf"`
		Squash bool   `json:"squash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch is required"})
		return
	}
	args := []string{"merge"}
	if body.NoFf {
		args = append(args, "--no-ff")
	}
	if body.Squash {
		args = append(args, "--squash")
	}
	args = append(args, body.Branch)
	out, err := runGitCommand(args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "branch": body.Branch, "output": out})
}

func handleGitBranchRebase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Branch string `json:"branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch is required"})
		return
	}
	out, err := runGitCommand("rebase", body.Branch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "branch": body.Branch, "output": out})
}

func handleGitBranchPush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Branch      string `json:"branch"`
		SetUpstream bool   `json:"setUpstream"`
		Force       bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch is required"})
		return
	}
	args := []string{"push"}
	if body.Force {
		args = append(args, "--force")
	}
	if body.SetUpstream {
		args = append(args, "-u", "origin", body.Branch)
	} else {
		args = append(args, "origin", body.Branch)
	}
	out, err := runGitCommand(args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "branch": body.Branch, "output": out})
}

func handleGitTagDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Tag name is required"})
		return
	}
	if _, err := runGitCommand("tag", "-d", body.Name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name})
}

func handleGitTagPush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Tag name is required"})
		return
	}
	out, err := runGitCommand("push", "origin", body.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name, "output": out})
}

func handleGitArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		ref = "HEAD"
	}
	cleanRef := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' || r == '/' {
			return r
		}
		return -1
	}, ref)
	filename := strings.ReplaceAll(cleanRef, "/", "_") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	cmd := exec.Command("git", "archive", "--format=zip", cleanRef)
	cmd.Stdout = w
	_ = cmd.Run()
}



