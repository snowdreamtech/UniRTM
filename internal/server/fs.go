// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type fsTreeCacheEntry struct {
	response FSTreeResponse
	cachedAt time.Time
}

var (
	fsTreeCacheMu sync.RWMutex
	fsTreeCache   = make(map[string]fsTreeCacheEntry)
)

const fsTreeCacheTTL = 1 * time.Second

// InvalidateFSCache clears the cached filesystem tree.
func InvalidateFSCache() {
	fsTreeCacheMu.Lock()
	fsTreeCache = make(map[string]fsTreeCacheEntry)
	fsTreeCacheMu.Unlock()
}

// RegisterFSRoutes registers filesystem and workspace roots management routes.
func RegisterFSRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/workspace/roots", handleGetWorkspaceRoots)
	mux.HandleFunc("/api/workspace/roots/add", handleAddWorkspaceRoot)
	mux.HandleFunc("/api/workspace/roots/remove", handleRemoveWorkspaceRoot)

	mux.HandleFunc("/api/fs/tree", handleFSTree)
	mux.HandleFunc("/api/fs/create", handleFSCreate)
	mux.HandleFunc("/api/fs/rename", handleFSRename)
	mux.HandleFunc("/api/fs/delete", handleFSDelete)
	mux.HandleFunc("/api/fs/read", handleFSRead)
	mux.HandleFunc("/api/fs/write", handleFSWrite)
}

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

func getGitIgnoredPaths(wd string) map[string]bool {
	ignored := make(map[string]bool)
	cmd := exec.Command("git", "status", "--ignored=matching", "--porcelain=v1")
	cmd.Dir = wd
	out, err := cmd.Output()
	if err != nil {
		return ignored
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "!! ") {
			p := strings.TrimPrefix(line, "!! ")
			p = strings.TrimSpace(p)
			p = strings.TrimSuffix(p, "/")
			if p != "" {
				ignored[filepath.ToSlash(p)] = true
			}
		}
	}
	return ignored
}

func isPathIgnored(relPath, name string, ignoredMap map[string]bool, parentIgnored bool) bool {
	if parentIgnored || isCommonIgnored(name) {
		return true
	}
	cleanRel := filepath.ToSlash(relPath)
	if ignoredMap[cleanRel] {
		return true
	}
	// Check if any ancestor path is in ignoredMap
	parts := strings.Split(cleanRel, "/")
	current := ""
	for i := 0; i < len(parts)-1; i++ {
		if current == "" {
			current = parts[i]
		} else {
			current = current + "/" + parts[i]
		}
		if ignoredMap[current] {
			return true
		}
	}
	return false
}

func buildFSTree(wd, dirPath, currentRelPath string, currentDepth, maxDepth int, showHidden bool, ignoredMap map[string]bool, parentIgnored bool) ([]FSEntry, error) {
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
		isDir := d.IsDir()
		isIgnored := isPathIgnored(entryRelPath, name, ignoredMap, parentIgnored)

		size := int64(0)
		modTime := ""
		if !isDir {
			if fi, err := d.Info(); err == nil {
				size = fi.Size()
				modTime = fi.ModTime().UTC().Format(time.RFC3339)
			}
		}

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
				subChildren, _ := buildFSTree(wd, fullEntryPath, entryRelPath, currentDepth+1, maxDepth, showHidden, ignoredMap, isIgnored)
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

type WorkspaceRootItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func getWorkspaceConfigFile(rootDir string) string {
	return filepath.Join(rootDir, ".unirtm", "workspaces.json")
}

// discoverSubRepositories scans immediate subdirectories of rootDir for nested git repositories or unirtm projects.
func discoverSubRepositories(rootDir string) []string {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil
	}

	var subRepos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || isCommonIgnored(name) {
			continue
		}
		subPath := filepath.Join(rootDir, name)
		hasGit := false
		if _, err := os.Stat(filepath.Join(subPath, ".git")); err == nil {
			hasGit = true
		}
		hasUniRTM := false
		if _, err := os.Stat(filepath.Join(subPath, ".unirtm.toml")); err == nil {
			hasUniRTM = true
		}

		if hasGit || hasUniRTM {
			subRepos = append(subRepos, name)
		}
	}

	sort.Slice(subRepos, func(i, j int) bool {
		return strings.ToLower(subRepos[i]) < strings.ToLower(subRepos[j])
	})

	return subRepos
}

func loadWorkspaceRoots(rootDir string) []WorkspaceRootItem {
	conf := getWorkspaceConfigFile(rootDir)
	data, err := os.ReadFile(conf)
	if err == nil {
		var wrapper struct {
			Roots []WorkspaceRootItem `json:"roots"`
		}
		if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Roots) > 0 {
			return wrapper.Roots
		}
	}
	name := filepath.Base(rootDir)
	if name == "" || name == "." || name == "/" {
		name = "workspace"
	}

	// Auto-discover sub-repositories if no explicit workspaces.json is configured.
	subRepos := discoverSubRepositories(rootDir)
	if len(subRepos) > 0 {
		var roots []WorkspaceRootItem
		hasGitRoot := false
		if _, err := os.Stat(filepath.Join(rootDir, ".git")); err == nil {
			hasGitRoot = true
		}
		hasUniRTMRoot := false
		if _, err := os.Stat(filepath.Join(rootDir, ".unirtm.toml")); err == nil {
			hasUniRTMRoot = true
		}

		if hasGitRoot || hasUniRTMRoot {
			roots = append(roots, WorkspaceRootItem{
				ID:   "root-primary",
				Name: name,
				Path: "",
			})
		}

		for _, sub := range subRepos {
			roots = append(roots, WorkspaceRootItem{
				ID:   fmt.Sprintf("root-%s", sub),
				Name: sub,
				Path: sub,
			})
		}
		if len(roots) > 0 {
			return roots
		}
	}

	return []WorkspaceRootItem{
		{
			ID:   "root-primary",
			Name: name,
			Path: "",
		},
	}
}

func saveWorkspaceRoots(rootDir string, roots []WorkspaceRootItem) error {
	confDir := filepath.Join(rootDir, ".unirtm")
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return err
	}
	wrapper := struct {
		Roots []WorkspaceRootItem `json:"roots"`
	}{
		Roots: roots,
	}
	bytes, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getWorkspaceConfigFile(rootDir), bytes, 0644)
}

func handleGetWorkspaceRoots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	roots := loadWorkspaceRoots(wd)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"roots":   roots,
	})
}

func handleAddWorkspaceRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request body"})
		return
	}
	cleanPath := strings.TrimSpace(req.Path)
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	roots := loadWorkspaceRoots(wd)
	folderName := strings.TrimSpace(req.Name)
	if folderName == "" {
		folderName = filepath.Base(cleanPath)
		if folderName == "" || folderName == "." {
			folderName = "Folder"
		}
	}
	exists := false
	for _, rt := range roots {
		if rt.Path == cleanPath {
			exists = true
			break
		}
	}
	if !exists {
		roots = append(roots, WorkspaceRootItem{
			ID:   fmt.Sprintf("root-%d", time.Now().UnixNano()),
			Name: folderName,
			Path: cleanPath,
		})
		_ = saveWorkspaceRoots(wd, roots)
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"roots":   roots,
	})
}

func handleRemoveWorkspaceRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request body"})
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	roots := loadWorkspaceRoots(wd)
	if len(roots) <= 1 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "cannot remove the only root folder"})
		return
	}
	var filtered []WorkspaceRootItem
	for _, item := range roots {
		if item.Path != req.Path {
			filtered = append(filtered, item)
		}
	}
	_ = saveWorkspaceRoots(wd, filtered)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"roots":   filtered,
	})
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

	cacheKey := fmt.Sprintf("%s:%d:%t", targetDir, depth, showHidden)
	fsTreeCacheMu.RLock()
	if entry, found := fsTreeCache[cacheKey]; found && time.Since(entry.cachedAt) < fsTreeCacheTTL {
		fsTreeCacheMu.RUnlock()
		_ = json.NewEncoder(w).Encode(entry.response)
		return
	}
	fsTreeCacheMu.RUnlock()

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

	ignoredMap := getGitIgnoredPaths(wd)
	entries, err := buildFSTree(wd, targetDir, cleanRel, 1, depth, showHidden, ignoredMap, false)

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

	fsTreeCacheMu.Lock()
	fsTreeCache[cacheKey] = fsTreeCacheEntry{
		response: resp,
		cachedAt: time.Now(),
	}
	fsTreeCacheMu.Unlock()

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

	InvalidateFSCache()
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

	InvalidateFSCache()
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

	InvalidateFSCache()
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

	InvalidateFSCache()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "path": body.Path, "size": len(body.Content)})
}
