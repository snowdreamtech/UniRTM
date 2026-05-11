// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/lockfile"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// LockService manages the unirtm.lock file lifecycle:
//   - Resolving download info from the lockfile (bypassing remote API calls)
//   - Writing back resolved info after a successful install
//   - Generating / refreshing lockfile entries for multiple platforms
//   - Enforcing strict mode (UNIRTM_LOCKED=1): fail if URL not in lockfile
type LockService struct {
	mu             sync.RWMutex
	lf             *lockfile.LockFile
	dirty          bool
	strictMode     bool
	lockfilePath   string
	backendRegistry *backend.Registry
}

// LockServiceOptions configures a LockService.
type LockServiceOptions struct {
	// LockfilePath overrides the default lock file location.
	// When empty, GetLockFilePath() from env/paths is used.
	LockfilePath string

	// StrictMode mirrors the UNIRTM_LOCKED env var / settings.locked setting.
	// When true, Install() will fail if the tool+platform URL is absent from
	// the lockfile (preventing any outbound API call).
	StrictMode bool
}

// NewLockService creates a LockService, loading the lockfile from disk if it exists.
func NewLockService(opts LockServiceOptions) (*LockService, error) {
	path := opts.LockfilePath
	if path == "" {
		path = defaultLockFilePath()
	}

	lf, err := lockfile.Load(path)
	if err != nil {
		return nil, fmt.Errorf("lock service: %w", err)
	}

	strict := opts.StrictMode || os.Getenv("UNIRTM_LOCKED") == "1"

	return &LockService{
		lf:           lf,
		lockfilePath: path,
		strictMode:   strict,
	}, nil
}

// defaultLockFilePath returns "unirtm.lock" in the current working directory,
// consistent with how mise.lock sits next to mise.toml.
func defaultLockFilePath() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unirtm.lock"
	}
	return wd + "/unirtm.lock"
}

// ─── Read API ─────────────────────────────────────────────────────────────────

// Resolve returns a *backend.VersionInfo populated from the lockfile for the
// given lockKey, version and current platform.  Returns (nil, false) when the
// lockfile has no entry for this combination.
//
// Callers (InstallationManager) should use the returned VersionInfo directly
// and skip the remote backend API call when ok==true.
func (ls *LockService) Resolve(
	lockKey, version string,
	platform backend.Platform,
) (*backend.VersionInfo, bool) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	platKey := lockfile.PlatformKey(platform.OS, platform.Arch, false)

	pe := ls.lf.GetPlatform(lockKey, version, platKey)
	if pe == nil || pe.URL == "" {
		return nil, false
	}

	info := &backend.VersionInfo{
		Version:     version,
		DownloadURL: pe.URL,
		Checksum:    pe.Checksum,
		Platform:    platform,
		Metadata:    map[string]string{"lock_url_api": pe.URLAPI},
	}
	return info, true
}

// CheckStrict verifies that the lockfile contains a URL for the given lockKey on
// the current platform when strict mode is enabled.  Returns an error (which
// callers must propagate) when the URL is absent.
func (ls *LockService) CheckStrict(lockKey, version string, platform backend.Platform) error {
	if !ls.strictMode {
		return nil
	}
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	platKey := lockfile.PlatformKey(platform.OS, platform.Arch, false)

	req := lockfile.LockRequirement{
		ToolKey:     lockKey,
		Version:     version,
		PlatformKey: platKey,
	}
	return ls.lf.CheckStrict([]lockfile.LockRequirement{req})
}

// ─── Write API ────────────────────────────────────────────────────────────────

// RecordInstall writes the resolved VersionInfo back into the lockfile for
// the current platform.  The lockfile is saved to disk immediately so that
// subsequent runs can use the cached URL.
//
// This is called by InstallationManager after a successful download.
func (ls *LockService) RecordInstall(
	lockKey, backendName string,
	info *backend.VersionInfo,
) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	platKey := lockfile.PlatformKey(info.Platform.OS, info.Platform.Arch, false)

	// Ensure the top-level entry exists (create or update backend field).
	existing := ls.lf.GetEntry(lockKey, info.Version)
	if existing == nil {
		ls.lf.UpsertEntry(lockKey, &lockfile.ToolLockEntry{
			Version:   info.Version,
			Backend:   backendName,
			Platforms: make(map[string]*lockfile.PlatformEntry),
		})
	}

	// Extract api URL from metadata if the backend stored it there.
	urlAPI := ""
	if info.Metadata != nil {
		urlAPI = info.Metadata["url_api"]
	}

	pe := &lockfile.PlatformEntry{
		Checksum: info.Checksum,
		URL:      info.DownloadURL,
		URLAPI:   urlAPI,
	}
	ls.lf.UpsertPlatform(lockKey, info.Version, platKey, pe)
	ls.dirty = true

	return ls.save()
}

// RemoveTool removes all lockfile entries for a tool (called on uninstall).
func (ls *LockService) RemoveTool(lockKey string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.lf.RemoveEntry(lockKey)
	ls.dirty = true
	return ls.save()
}

// ─── Generation API ───────────────────────────────────────────────────────────

// GenerateOptions controls what `unirtm lock` generates.
type GenerateOptions struct {
	// Tools is the subset of tools to refresh.  Empty = all tools from config.
	Tools []string
	// Platforms is the list of platform keys to generate entries for.
	// Empty = only current platform.  Use lockfile.StandardPlatforms for all.
	Platforms []string
}

// Generate resolves download info from the backend for each (tool, platform)
// pair and writes the result into the lockfile.
//
// ctx is used for backend API calls (cancellation, deadline).
// tools is a map of toolName → {version, backendName} from the project config.
func (ls *LockService) Generate(
	ctx context.Context,
	tools map[string]ToolSpec,
	opts GenerateOptions,
) error {
	platforms := opts.Platforms
	if len(platforms) == 0 {
		platforms = []string{lockfile.CurrentPlatformKey()}
	}

	subset := buildSubset(tools, opts.Tools)

	var wg sync.WaitGroup
	errs := make(chan error, len(subset)*len(platforms))

	// Limit concurrency to avoid hitting API rate limits (e.g. GitHub)
	// and to prevent excessive resource usage.
	sem := make(chan struct{}, 10)

	// Add timeout to context for the entire generation process
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Clear old entries for tools we are refreshing to avoid duplicates
	// (e.g. "asdf:go" vs "go")
	ls.mu.Lock()
	for uniqueKey := range subset {
		ls.lf.RemoveEntry(uniqueKey)
		// Also remove old prefixed versions if any
		for _, b := range ls.backendRegistry.List() {
			ls.lf.RemoveEntry(b + ":" + uniqueKey)
		}
	}
	ls.mu.Unlock()

	for uniqueKey, spec := range subset {
		uniqueKey := uniqueKey
		spec := spec

		toolName := spec.Name
		if toolName == "" {
			toolName = uniqueKey
		}

		b, err := ls.backendForSpec(toolName, spec.BackendName)
		if err != nil {
			logger.Warn("lockfile generate: skipping tool (no backend)", map[string]interface{}{
				"tool":  toolName,
				"error": err.Error(),
			})
			continue
		}

		for _, platKey := range platforms {
			platKey := platKey
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Acquire semaphore
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-ctx.Done():
					errs <- ctx.Err()
					return
				}

				goos, goarch, _, err := lockfile.ParsePlatformKey(platKey)
				if err != nil {
					errs <- fmt.Errorf("lockfile generate: %w", err)
					return
				}
				plat := backend.Platform{OS: goos, Arch: goarch}

				info, err := b.GetDownloadInfo(ctx, toolName, spec.Version, plat)
				if err != nil {
					logger.Warn("lockfile generate: could not resolve download info", map[string]interface{}{
						"tool":     toolName,
						"version":  spec.Version,
						"platform": platKey,
						"error":    err.Error(),
					})
					return
				}

				lockKey := uniqueKey
				urlAPI := ""
				if info.Metadata != nil {
					urlAPI = info.Metadata["url_api"]
				}

				ls.mu.Lock()
				defer ls.mu.Unlock()

				existing := ls.lf.GetEntry(lockKey, info.Version)
				if existing == nil {
					ls.lf.UpsertEntry(lockKey, &lockfile.ToolLockEntry{
						Version:   info.Version,
						Backend:   spec.BackendName,
						Platforms: make(map[string]*lockfile.PlatformEntry),
					})
				}
				ls.lf.UpsertPlatform(lockKey, info.Version, platKey, &lockfile.PlatformEntry{
					Checksum: info.Checksum,
					URL:      info.DownloadURL,
					URLAPI:   urlAPI,
				})
				ls.dirty = true

				logger.Debug("lockfile: resolved", map[string]interface{}{
					"tool":     toolName,
					"version":  info.Version,
					"platform": platKey,
				})
			}()
		}
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.save()
}

// ToolSpec describes a single tool entry from project config.
type ToolSpec struct {
	Name        string
	Version     string
	BackendName string
}

// ─── Inspection ───────────────────────────────────────────────────────────────

// Validate runs structural validation on the in-memory lockfile.
func (ls *LockService) Validate() error {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.lf.Validate()
}

// Path returns the filesystem path of the managed lockfile.
func (ls *LockService) Path() string { return ls.lockfilePath }

// IsEmpty reports whether the lockfile contains any tool entries.
func (ls *LockService) IsEmpty() bool {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.lf.IsEmpty()
}

// IsStrictMode reports whether strict mode is active.
func (ls *LockService) IsStrictMode() bool { return ls.strictMode }

// ─── Internal helpers ─────────────────────────────────────────────────────────

// save must be called with ls.mu held (write-lock).
func (ls *LockService) save() error {
	if !ls.dirty {
		return nil
	}
	if err := ls.lf.Save(); err != nil {
		return fmt.Errorf("lock service: save: %w", err)
	}
	ls.dirty = false
	return nil
}

// backendForSpec returns the backend Registry getter stub.
// The actual registry is injected via SetBackendRegistry.
func (ls *LockService) backendForSpec(toolName, backendName string) (backend.Backend, error) {
	if ls.backendRegistry == nil {
		return nil, fmt.Errorf("no backend registry configured")
	}
	return ls.backendRegistry.Get(backendName)
}

// SetBackendRegistry wires the backend registry used by Generate().
func (ls *LockService) SetBackendRegistry(r *backend.Registry) {
	ls.mu.Lock()
	ls.backendRegistry = r
	ls.mu.Unlock()
}

// backendRegistry is stored separately for injection.
func (ls *LockService) init() {} // placeholder for future init logic

var _ = (*LockService).init // suppress unused warning

// buildSubset filters a full tools map to only the requested names.
func buildSubset(all map[string]ToolSpec, filter []string) map[string]ToolSpec {
	if len(filter) == 0 {
		return all
	}
	want := make(map[string]bool, len(filter))
	for _, t := range filter {
		want[t] = true
	}
	out := make(map[string]ToolSpec)
	for name, spec := range all {
		if want[name] {
			out[name] = spec
		}
	}
	return out
}
