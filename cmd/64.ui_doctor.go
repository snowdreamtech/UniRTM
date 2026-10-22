// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/pkg/version"
	"github.com/snowdreamtech/unirtm/internal/utils"
)

// DoctorDiagnosticReport represents the structured report returned by /api/doctor
type DoctorDiagnosticReport struct {
	Timestamp   int64               `json:"timestamp"`
	Score       int                 `json:"score"`
	Status      string              `json:"status"`
	Core        DoctorCoreStatus    `json:"core"`
	Build       DoctorBuildInfo     `json:"build"`
	Context     DoctorContextInfo   `json:"context"`
	Directories []DoctorDirInfo     `json:"directories"`
	Configs     []DoctorConfigInfo  `json:"configs"`
	Settings    []DoctorSettingInfo `json:"settings"`
	Backends    []string            `json:"backends"`
	Tools       []DoctorToolInfo    `json:"tools"`
	EnvVars     []DoctorEnvVar      `json:"env_vars"`
	Tasks       []DoctorTaskInfo    `json:"tasks"`
	Health      DoctorHealthInfo    `json:"health"`
	Suggestions []DoctorSuggestion  `json:"suggestions"`
}

type DoctorCoreStatus struct {
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	Activated   bool   `json:"activated"`
	ShimsOnPath bool   `json:"shims_on_path"`
	ShimsDir    string `json:"shims_dir"`
	TrustStatus string `json:"trust_status"`
}

type DoctorBuildInfo struct {
	Target    string   `json:"target"`
	GoVersion string   `json:"go_version"`
	Built     string   `json:"built"`
	Features  []string `json:"features"`
}

type DoctorContextInfo struct {
	Shell        string `json:"shell"`
	ShellVersion string `json:"shell_version"`
	Editor       string `json:"editor"`
	EditorSource string `json:"editor_source"`
	WorkDir      string `json:"work_dir"`
	ActiveEnv    string `json:"active_env"`
}

type DoctorDirInfo struct {
	Type      string `json:"type"`
	Path      string `json:"path"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

type DoctorConfigInfo struct {
	Path   string `json:"path"`
	Loaded bool   `json:"loaded"`
	Scope  string `json:"scope"`
}

type DoctorSettingInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DoctorToolInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Backend     string `json:"backend"`
	Installed   bool   `json:"installed"`
	InstallPath string `json:"install_path,omitempty"`
}

type DoctorEnvVar struct {
	Key    string `json:"key"`
	Source string `json:"source"`
	Value  string `json:"value"`
}

type DoctorTaskInfo struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description"`
}

type DoctorHealthInfo struct {
	DatabasePath        string `json:"database_path"`
	DatabaseSize        string `json:"database_size"`
	DatabaseOk          bool   `json:"database_ok"`
	DatabaseError       string `json:"database_error,omitempty"`
	GitHubConnected     bool   `json:"github_connected"`
	GitHubStatusCode    int    `json:"github_status_code"`
	GitHubRateLimit     int64  `json:"github_rate_limit"`
	GitHubRateRemaining int64  `json:"github_rate_remaining"`
	GitHubRateReset     int64  `json:"github_rate_reset"`
	GitHubRateResetText string `json:"github_rate_reset_text"`
}

type DoctorSuggestion struct {
	ID         string `json:"id"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Command    string `json:"command,omitempty"`
	Actionable bool   `json:"actionable"`
}

func handleDoctorDiagnostic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	report, err := collectDoctorReport(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(report)
}

func collectDoctorReport(ctx context.Context) (*DoctorDiagnosticReport, error) {
	cfg, _ := config.LoadFull()

	cwd, _ := os.Getwd()
	isActivated := env.Get("ACTIVE") != ""
	shimsDir := env.GetShimsDir()
	pathItems := filepath.SplitList(env.Get("PATH"))
	shimsOnPath := false
	for _, p := range pathItems {
		if strings.EqualFold(p, shimsDir) {
			shimsOnPath = true
			break
		}
	}

	shellPath := env.Get("SHELL")
	shellVer := "unknown"
	if shellPath != "" {
		if out, err := exec.Command(shellPath, "-c", "echo $ZSH_VERSION $BASH_VERSION").CombinedOutput(); err == nil {
			shellVer = strings.TrimSpace(string(out))
		}
	} else if runtime.GOOS == "windows" {
		if msystem := os.Getenv("MSYSTEM"); msystem != "" {
			shellPath = fmt.Sprintf("Git Bash / MSYS2 (%s)", msystem)
		} else if os.Getenv("PSModulePath") != "" {
			shellPath = "PowerShell"
		} else if os.Getenv("PROMPT") != "" {
			shellPath = "Command Prompt (CMD)"
		} else {
			shellPath = "Windows Native"
		}
	}

	editor, editorSource := getBestEditorWithSource(cfg)

	// 1. Core
	core := DoctorCoreStatus{
		Version:     env.GitTag,
		Commit:      env.CommitHash,
		Activated:   isActivated,
		ShimsOnPath: shimsOnPath,
		ShimsDir:    shimsDir,
		TrustStatus: "trusted",
	}

	// 2. Build
	build := DoctorBuildInfo{
		Target:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion: runtime.Version(),
		Built:     env.BuildTime,
		Features:  []string{"SQLITE", "OPENSSL", "GPG", "VFOX_COMPAT"},
	}

	// 3. Context
	activeEnv := "base"
	if cfg != nil {
		if e := env.Get("ENV"); e != "" {
			activeEnv = e
		}
	}
	diagContext := DoctorContextInfo{
		Shell:        shellPath,
		ShellVersion: shellVer,
		Editor:       editor,
		EditorSource: editorSource,
		WorkDir:      cwd,
		ActiveEnv:    activeEnv,
	}

	// 4. Directories
	calcDirSize := func(p string) (string, int64) {
		s, err := utils.CalculateDirectorySize(p)
		if err != nil {
			return "-", 0
		}
		return utils.FormatBytes(s), s
	}

	cacheSize, cacheBytes := calcDirSize(env.GetCacheDir())
	dataSize, dataBytes := calcDirSize(env.GetDataDir())
	shimsSize, shimsBytes := calcDirSize(env.GetShimsDir())

	dirs := []DoctorDirInfo{
		{Type: "cache", Path: env.GetCacheDir(), Size: cacheSize, SizeBytes: cacheBytes},
		{Type: "config", Path: env.GetConfigDir(), Size: "-"},
		{Type: "data", Path: env.GetDataDir(), Size: dataSize, SizeBytes: dataBytes},
		{Type: "shims", Path: env.GetShimsDir(), Size: shimsSize, SizeBytes: shimsBytes},
	}

	// 5. Configs
	configsToCheck := []struct {
		path  string
		scope string
	}{
		{filepath.Join(cwd, ".unirtm.toml"), "workspace"},
		{filepath.Join(cwd, "unirtm.toml"), "workspace"},
		{filepath.Join(cwd, ".mise.toml"), "workspace"},
		{filepath.Join(cwd, "mise.toml"), "workspace"},
		{env.GetGlobalConfigPath(), "global"},
	}
	var configs []DoctorConfigInfo
	for _, c := range configsToCheck {
		if _, err := os.Stat(c.path); err == nil {
			configs = append(configs, DoctorConfigInfo{
				Path:   c.path,
				Loaded: true,
				Scope:  c.scope,
			})
		}
	}

	// 6. Settings
	var settings []DoctorSettingInfo
	if cfg != nil {
		v := reflect.ValueOf(cfg.Settings)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			val := v.Field(i).Interface()
			displayVal := fmt.Sprintf("%v", val)
			if field.Type.Kind() == reflect.Ptr && !v.Field(i).IsNil() {
				displayVal = fmt.Sprintf("%v", v.Field(i).Elem().Interface())
			} else if field.Type.Kind() == reflect.Ptr && v.Field(i).IsNil() {
				displayVal = "unset"
			}
			if strings.Contains(field.Name, "Token") && displayVal != "unset" && displayVal != "" {
				displayVal = "******** [REDACTED]"
			}
			tagName := field.Tag.Get("toml")
			if tagName == "" {
				tagName = field.Name
			}
			tagName = strings.Split(tagName, ",")[0]
			settings = append(settings, DoctorSettingInfo{
				Key:   tagName,
				Value: displayVal,
			})
		}
	}

	// 7. Backends
	backends := backend.List()
	sort.Strings(backends)

	// 8. Tools
	var tools []DoctorToolInfo
	missingToolsCount := 0
	if cfg != nil && len(cfg.Tools) > 0 {
		keys := make([]string, 0, len(cfg.Tools))
		for k := range cfg.Tools {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, name := range keys {
			t := cfg.Tools[name]
			backendName := t.Backend
			toolName := name
			if backendName == "" {
				if idx := strings.Index(name, ":"); idx != -1 {
					backendName = name[:idx]
					toolName = name[idx+1:]
					if backendName == "go" {
						backendName = "go-pkg"
					}
				}
			}
			slug := env.GetFSToolName(toolName, backendName)
			v := t.Version
			installed := false
			installPath := ""

			toolPathExact := filepath.Join(env.GetInstallsDir(), slug, v)
			if _, err := os.Stat(toolPathExact); err == nil {
				installed = true
				installPath = toolPathExact
			} else {
				if ver, err := version.ParseVersion(v); err == nil {
					toolPathNorm := filepath.Join(env.GetInstallsDir(), slug, ver.String())
					if _, err := os.Stat(toolPathNorm); err == nil {
						installed = true
						installPath = toolPathNorm
					}
				}
			}

			if !installed {
				missingToolsCount++
			}

			tools = append(tools, DoctorToolInfo{
				Name:        name,
				Version:     t.Version,
				Backend:     backendName,
				Installed:   installed,
				InstallPath: installPath,
			})
		}
	}

	// 9. Env Vars
	relevantVars := make(map[string]bool)
	configKeys := []string{"DATA_DIR", "CONFIG_DIR", "CACHE_DIR", "EXPERIMENTAL", "DEBUG", "GITHUB_TOKEN", "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "ACTIVATION_SCOPE"}
	for _, k := range configKeys {
		relevantVars[k] = true
	}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], "UNIRTM_") {
			relevantVars[strings.TrimPrefix(pair[0], "UNIRTM_")] = true
		} else if strings.HasPrefix(pair[0], "MISE_") {
			relevantVars[strings.TrimPrefix(pair[0], "MISE_")] = true
		}
	}
	sortedEnvKeys := make([]string, 0, len(relevantVars))
	for k := range relevantVars {
		sortedEnvKeys = append(sortedEnvKeys, k)
	}
	sort.Strings(sortedEnvKeys)

	var envVars []DoctorEnvVar
	for _, k := range sortedEnvKeys {
		unirtmVal := os.Getenv("UNIRTM_" + k)
		miseVal := os.Getenv("MISE_" + k)
		rawVal := os.Getenv(k)
		resolved := env.Get(k)

		source := "Raw"
		if unirtmVal != "" {
			source = "UNIRTM_"
		} else if miseVal != "" {
			source = "MISE_"
		} else if rawVal == "" {
			source = "Default"
			switch k {
			case "DATA_DIR":
				resolved = env.GetDataDir()
			case "CONFIG_DIR":
				resolved = env.GetConfigDir()
			case "CACHE_DIR":
				resolved = env.GetCacheDir()
			}
		}

		if resolved == "" && source == "Raw" {
			continue
		}

		displayVal := resolved
		if strings.Contains(k, "TOKEN") && displayVal != "" {
			displayVal = "******** [REDACTED]"
		}

		envVars = append(envVars, DoctorEnvVar{
			Key:    k,
			Source: source,
			Value:  displayVal,
		})
	}

	// 10. Tasks
	var tasks []DoctorTaskInfo
	if cfg != nil && len(cfg.Tasks) > 0 {
		for name, t := range cfg.Tasks {
			tasks = append(tasks, DoctorTaskInfo{
				Name:        name,
				Command:     strings.Join(t.Run, " && "),
				Description: t.Description,
			})
		}
	}

	// 11. Health (DB + GitHub)
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	dbOk := err == nil
	dbErr := ""
	if err != nil {
		dbErr = err.Error()
	} else {
		_ = db.Close()
	}

	client := pkgHttp.NewClientWithTimeout(5 * time.Second)
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/rate_limit", nil)
	if token := env.Get("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	githubConnected := false
	githubStatusCode := 0
	var githubLimit, githubRemaining, githubReset int64
	githubResetText := ""

	if resp, err := client.Do(req); err == nil {
		githubConnected = true
		githubStatusCode = resp.StatusCode
		limitStr := resp.Header.Get("X-RateLimit-Limit")
		remainingStr := resp.Header.Get("X-RateLimit-Remaining")
		resetStr := resp.Header.Get("X-RateLimit-Reset")

		fmt.Sscanf(limitStr, "%d", &githubLimit)
		fmt.Sscanf(remainingStr, "%d", &githubRemaining)
		fmt.Sscanf(resetStr, "%d", &githubReset)

		if githubReset > 0 {
			githubResetText = time.Until(time.Unix(githubReset, 0)).Round(time.Minute).String()
		}
		_ = resp.Body.Close()
	}

	health := DoctorHealthInfo{
		DatabasePath:        dbPath,
		DatabaseSize:        getFileSize(dbPath),
		DatabaseOk:          dbOk,
		DatabaseError:       dbErr,
		GitHubConnected:     githubConnected,
		GitHubStatusCode:    githubStatusCode,
		GitHubRateLimit:     githubLimit,
		GitHubRateRemaining: githubRemaining,
		GitHubRateReset:     githubReset,
		GitHubRateResetText: githubResetText,
	}

	// 12. Fix Suggestions & Score
	var suggestions []DoctorSuggestion
	score := 100

	if !shimsOnPath {
		suggestions = append(suggestions, DoctorSuggestion{
			ID:         "shims-missing",
			Severity:   "warning",
			Message:    "UniRTM shims 目录未添加到系统 PATH 环境变量中",
			Command:    "unirtm enable --shims",
			Actionable: true,
		})
		score -= 10
	}

	if !isActivated {
		suggestions = append(suggestions, DoctorSuggestion{
			ID:         "not-activated",
			Severity:   "warning",
			Message:    "当前 shell 环境未激活 UniRTM 自动代理路由",
			Command:    "unirtm enable --shims",
			Actionable: true,
		})
		score -= 10
	}

	if missingToolsCount > 0 {
		suggestions = append(suggestions, DoctorSuggestion{
			ID:         "missing-tools",
			Severity:   "warning",
			Message:    fmt.Sprintf("检测到当前配置有 %d 个工具未安装", missingToolsCount),
			Command:    "unirtm install",
			Actionable: true,
		})
		score -= 15
	}

	if !dbOk {
		suggestions = append(suggestions, DoctorSuggestion{
			ID:         "database-unhealthy",
			Severity:   "error",
			Message:    fmt.Sprintf("UniRTM 本地 SQLite 状态数据库访问异常: %s", dbErr),
			Actionable: false,
		})
		score -= 25
	}

	if score < 0 {
		score = 0
	}

	status := "healthy"
	if score < 80 {
		status = "warning"
	}
	if score < 50 || !dbOk {
		status = "critical"
	}

	return &DoctorDiagnosticReport{
		Timestamp:   time.Now().Unix(),
		Score:       score,
		Status:      status,
		Core:        core,
		Build:       build,
		Context:     diagContext,
		Directories: dirs,
		Configs:     configs,
		Settings:    settings,
		Backends:    backends,
		Tools:       tools,
		EnvVars:     envVars,
		Tasks:       tasks,
		Health:      health,
		Suggestions: suggestions,
	}, nil
}
