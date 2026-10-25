// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/hook"
)

// RegisterTasksRoutes registers tasks, environment variables, and hook routes.
func RegisterTasksRoutes(mux *http.ServeMux) {
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
			"hint":  "Please configure hooks in Workspace scope instead.",
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
