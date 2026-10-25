// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/config"
	"golang.org/x/net/websocket"
)

// RegisterSystemRoutes registers all system, browser, and terminal routes.
func RegisterSystemRoutes(mux *http.ServeMux) {
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
}

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

// SelfUpdateResult holds release and update status for the Workbench UI.
type SelfUpdateResult struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	ReleaseName    string `json:"release_name"`
	ReleaseNotes   string `json:"release_notes"`
	InstallMethod  string `json:"install_method"`
	UpgradeHint    string `json:"upgrade_hint,omitempty"`
	IsUnsupported  bool   `json:"is_unsupported"`
}

var (
	CheckSelfUpdateFn   func() (*SelfUpdateResult, error)
	ExecuteSelfUpdateFn func(version string) (string, error)
)

// handleSystemCheckUpdate handles GET /api/system/check-update by delegating to CLI self-update core.
func handleSystemCheckUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	if CheckSelfUpdateFn != nil {
		result, err := CheckSelfUpdateFn()
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":         true,
			"current_version": result.CurrentVersion,
			"latest_version":  result.LatestVersion,
			"has_update":      result.HasUpdate,
			"release_name":    result.ReleaseName,
			"release_notes":   result.ReleaseNotes,
			"install_method":  result.InstallMethod,
			"upgrade_hint":    result.UpgradeHint,
			"is_unsupported":  result.IsUnsupported,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"current_version": "dev",
		"latest_version":  "dev",
		"has_update":      false,
	})
}

// handleSystemSelfUpdate handles POST /api/system/self-update by invoking official self-update workflow.
func handleSystemSelfUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	if ExecuteSelfUpdateFn != nil {
		outputLog, err := ExecuteSelfUpdateFn("latest")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"output":  outputLog,
			})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "UniRTM updated successfully.",
			"output":  outputLog,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Self-update is handled via terminal.",
	})
}
