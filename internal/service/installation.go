// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/gpg"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/provider/native"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/transaction"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
)

// ProgressReporter is a callback for concurrent download progress reporting.
type ProgressReporter func(tool string, downloaded, total int64)

const (
	ContextKeyQuietProgress    = "quietProgress"
	ContextKeyProgressReporter = "concurrentProgressReporter"
)

// ErrAlreadyInstalled is returned when a tool version is already installed.
var ErrAlreadyInstalled = fmt.Errorf("already installed")

// InstallationManager manages tool installation workflow.
type InstallationManager struct {
	backendRegistry  *backend.Registry
	providerRegistry *provider.Registry
	downloadManager  *download.Manager
	installRepo      repository.InstallationRepository
	txManager        transaction.TransactionManager
	lockService      *LockService // optional; nil = lockfile disabled
	settings         *config.Settings
	aliases          map[string]map[string]string
	toolConfigs      map[string]config.ToolConfig
	gpgVerifier      gpg.Verifier
	shimGenerator    *Generator
}

// NewInstallationManager creates a new installation manager without lockfile support.
func NewInstallationManager(
	backendRegistry *backend.Registry,
	providerRegistry *provider.Registry,
	downloadManager *download.Manager,
	installRepo repository.InstallationRepository,
	txManager transaction.TransactionManager,
	settings *config.Settings,
) *InstallationManager {
	// Initialize global no-proxy list from settings
	if settings != nil {
		provider.GlobalNoProxy = settings.NoProxy
	}

	return &InstallationManager{
		backendRegistry:  backendRegistry,
		providerRegistry: providerRegistry,
		downloadManager:  downloadManager,
		installRepo:      installRepo,
		txManager:        txManager,
		settings:         settings,
		gpgVerifier:      gpg.NewVerifier(),
		shimGenerator:    NewGenerator(env.GetShimsDir(), env.GetInstallsDir()),
	}
}

// NewInstallationManagerWithLock creates an InstallationManager that reads and
// writes unirtm.lock for reproducible, API-call-free installations.
func NewInstallationManagerWithLock(
	backendRegistry *backend.Registry,
	providerRegistry *provider.Registry,
	downloadManager *download.Manager,
	installRepo repository.InstallationRepository,
	txManager transaction.TransactionManager,
	lockService *LockService,
	settings *config.Settings,
) *InstallationManager {
	im := NewInstallationManager(backendRegistry, providerRegistry, downloadManager, installRepo, txManager, settings)
	im.lockService = lockService
	return im
}

// SetAliases sets the version aliases for tools.
func (im *InstallationManager) SetAliases(aliases map[string]map[string]string) {
	im.aliases = aliases
}

// SetToolConfigs sets the tool configurations for hooks.
func (im *InstallationManager) SetToolConfigs(toolConfigs map[string]config.ToolConfig) {
	im.toolConfigs = toolConfigs
}

// resolveAlias resolves a version alias for a tool.
func (im *InstallationManager) resolveAlias(tool, version string) string {
	if im.aliases == nil {
		return version
	}
	if toolAliases, ok := im.aliases[tool]; ok {
		if resolved, ok := toolAliases[version]; ok {
			return resolved
		}
	}
	return version
}

// SelectVersionInteractive opens an interactive menu to select a tool version.
func (im *InstallationManager) SelectVersionInteractive(ctx context.Context, tool, backendName string) (string, error) {
	// 1. Get backend
	b, err := im.backendRegistry.Get(backendName)
	if err != nil {
		return "", fmt.Errorf("get backend %s: %w", backendName, err)
	}

	// 2. List remote versions
	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Fetching versions for %s...", tool))
	platform := backend.CurrentPlatform()
	versionInfos, err := b.ListVersions(ctx, tool, platform)
	if err != nil {
		spinner.Fail(err.Error())
		return "", fmt.Errorf("list versions: %w", err)
	}
	spinner.Success()

	if len(versionInfos) == 0 {
		return "", fmt.Errorf("no versions found for %s", tool)
	}

	// 3. Sort versions in descending order (newest first)
	sort.SliceStable(versionInfos, func(i, j int) bool {
		vI, errI := parseSemVer(versionInfos[i].Version)
		vJ, errJ := parseSemVer(versionInfos[j].Version)

		if errI == nil && errJ == nil {
			// Both are valid SemVer, sort descending
			return vI.Compare(vJ) > 0
		}

		// Fallback to alphabetical comparison for date-based versions and tags, sort descending
		return versionInfos[i].Version > versionInfos[j].Version
	})

	// 4. Convert to string slice
	versions := make([]string, len(versionInfos))
	for i, info := range versionInfos {
		versions[i] = info.Version
	}

	// 4. Show interactive menu
	selected, err := pterm.DefaultInteractiveSelect.
		WithOptions(versions).
		WithDefaultText(fmt.Sprintf("Select version for %s", tool)).
		Show()

	if err != nil {
		return "", fmt.Errorf("interactive select: %w", err)
	}

	return selected, nil
}

// executeHook executes a command as a tool hook.
func (im *InstallationManager) executeHook(ctx context.Context, cmdStr, tool, version string) error {
	if cmdStr == "" {
		return nil
	}

	quietProgress := false
	if val, ok := ctx.Value(ContextKeyQuietProgress).(bool); ok && val {
		quietProgress = true
	}

	if !quietProgress {
		fmt.Printf("➜ executing hook for %s@%s: %s\n", tool, version, cmdStr)
	}

	// Create command
	var shell, shellArg string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		shellArg = "/c"
	} else {
		shell = "sh"
		shellArg = "-c"
	}

	execCmd := exec.CommandContext(ctx, shell, shellArg, cmdStr)
	if quietProgress {
		execCmd.Stdout = nil
		execCmd.Stderr = nil
	} else {
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
	}

	// Set required environment variables
	execCmd.Env = os.Environ()

	// Add context env vars
	execCmd.Env = append(execCmd.Env, "UNIRTM_TOOL="+tool)
	execCmd.Env = append(execCmd.Env, "UNIRTM_VERSION="+version)

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("hook failed: %w", err)
	}

	return nil
}

// IsInstalled checks if a tool version is installed, considering version variants (v-prefix).
func (im *InstallationManager) IsInstalled(ctx context.Context, tool, version, backendName string) (bool, *repository.Installation) {
	// 1. Resolve aliases
	version = im.resolveAlias(tool, version)

	// 2. Standardize tool name for filesystem check
	fsToolName := env.GetFSToolName(tool, backendName)

	// 3. Prepare variants to check (original, and normalized if it's a semver)
	variants := []string{version}
	if v, err := ParseVersion(version); err == nil && v.Type == VersionTypeExact {
		normalized := v.String()
		if normalized != version {
			variants = append(variants, normalized)
		}
		// If input didn't have 'v' but parse was successful, try adding 'v' just in case
		// though our standard is usually no 'v' in the directory name.
		if !strings.HasPrefix(version, "v") && !strings.HasPrefix(version, "V") {
			variants = append(variants, "v"+normalized)
		}
	} else {
		// Not a standard semver, try basic v-stripping as fallback
		if strings.HasPrefix(version, "v") || strings.HasPrefix(version, "V") {
			variants = append(variants, version[1:])
		} else {
			variants = append(variants, "v"+version)
		}
	}

	for _, v := range variants {
		existing, err := im.installRepo.FindByToolAndVersion(ctx, tool, v)
		if err == nil && existing != nil {
			// Verify physical existence on disk
			checkPath := filepath.Join(env.GetInstallsDir(), fsToolName, v)
			if _, statErr := os.Stat(checkPath); statErr == nil {
				return true, existing
			}
		}
	}

	return false, nil
}

// Install performs the complete installation workflow for a tool.
// Workflow: check → download → verify → extract → activate → record
func (im *InstallationManager) Install(ctx context.Context, toolKey, tool, version, backendName string) error {
	quietProgress, _ := ctx.Value(ContextKeyQuietProgress).(bool)

	// Standardize tool name for filesystem check
	fsToolName := env.GetFSToolName(tool, backendName)

	// 1. Resolve aliases
	version = im.resolveAlias(tool, version)

	// 2. Fetch tool config for hooks
	var preInstall, postInstall string
	if im.toolConfigs != nil {
		if tc, ok := im.toolConfigs[tool]; ok {
			preInstall = tc.PreInstall
			postInstall = tc.PostInstall
		}
	}

	// 3. Run PreInstall hook
	if preInstall != "" {
		output.Warningf("⚠️  SECURITY: executing pre_install hook for %s: %s", tool, preInstall)
		if err := im.executeHook(ctx, preInstall, tool, version); err != nil {
			return fmt.Errorf("pre_install hook failed: %w", err)
		}
	}

	// Get backend
	b, err := im.backendRegistry.Get(backendName)
	if err != nil {
		return fmt.Errorf("backend not found: %w", err)
	}

	// 4. Optimization: Check if this CONCRETE version is already installed before any network/API calls
	// We only do this if it's not a symbolic version like "latest" or "3.12" (which needs resolution)
	// Simple heuristic: if it contains 2 dots, it's likely a concrete version.
	isConcrete := strings.Count(version, ".") >= 2 || (tool == "go" && strings.Count(version, ".") >= 1)
	if isConcrete {
		if installed, _ := im.IsInstalled(ctx, tool, version, backendName); installed {
			return ErrAlreadyInstalled
		}
	}

	// Get download info — check lockfile first to avoid remote API calls.
	platform := backend.CurrentPlatform()
	var versionInfo *backend.VersionInfo

	if im.lockService != nil {
		// Enforce strict mode before any API call.
		if err := im.lockService.CheckStrict(toolKey, version, platform); err != nil {
			return err
		}
		// Try to resolve directly from the lockfile.
		if info, ok := im.lockService.Resolve(toolKey, version, platform); ok {
			logger.Debug("lockfile hit: using cached URL", map[string]interface{}{
				"tool":     toolKey,
				"version":  version,
				"platform": platform.String(),
			})
			versionInfo = info
		}
	}

	if versionInfo == nil {
		// Lockfile miss — fall back to the remote backend.
		if !quietProgress {
			fmt.Printf("ℹ resolving download info for %s@%s...\n", tool, version)
		}
		info, err := b.ResolveVersion(ctx, tool, version, platform)
		if err != nil {
			return fmt.Errorf("failed to resolve version: %w", err)
		}
		version = info.Version // Update to the concrete resolved version
		if !quietProgress {
			output.Successf("✓ resolved %s to version %s", tool, version)
		}
		versionInfo = info
	}

	// 5. Check if already installed (AFTER resolving concrete version)
	// Note: We use im.installRepo (non-transactional) here to avoid holding a transaction during download.
	existing, err := im.installRepo.FindByToolAndVersion(ctx, tool, version)
	if err == nil && existing != nil {
		// Update path if needed (in case db has old path but we want to check new standard)
		checkPath := filepath.Join(env.GetInstallsDir(), fsToolName, version)
		if _, statErr := os.Stat(checkPath); statErr == nil {
			// Found a valid installation on the NEW standard disk path
			return ErrAlreadyInstalled
		}
	}
	var downloadPath string
	var gpgStatus string = "NotRequested"
	if versionInfo.DownloadURL != "" {
		// Extract extension from URL if possible
		ext := ""
		if u, err := url.Parse(versionInfo.DownloadURL); err == nil {
			ext = filepath.Ext(u.Path)
			// Handle some common double extensions like .tar.gz
			if strings.HasSuffix(u.Path, ".tar.gz") {
				ext = ".tar.gz"
			} else if strings.HasSuffix(u.Path, ".tar.xz") {
				ext = ".tar.xz"
			}
		}

		downloadPath = filepath.Join(env.GetDownloadsDir(), fmt.Sprintf("%s-%s%s", tool, version, ext))
		if !quietProgress {
			fmt.Printf("ℹ downloading %s@%s to %s...\n", tool, version, downloadPath)
		}
		if err := os.MkdirAll(filepath.Dir(downloadPath), 0755); err != nil {
			return fmt.Errorf("failed to create downloads directory: %w", err)
		}
		downloader, err := im.downloadManager.Get("https")
		if err != nil {
			return fmt.Errorf("failed to get downloader: %w", err)
		}

		opts := download.DefaultDownloadOptions()
		if im.settings != nil {
			opts.GitHubProxy = im.settings.GitHubProxy
			if opts.GitHubProxy == "" {
				opts.GitHubProxy = env.Get("GITHUB_PROXY")
			}
			if im.settings.HTTPTimeout > 0 {
				opts.Timeout = time.Duration(im.settings.HTTPTimeout) * time.Second
			}
		} else {
			opts.GitHubProxy = env.Get("GITHUB_PROXY")
		}
		if versionInfo.Checksum != "" {
			opts = opts.WithChecksum(versionInfo.Checksum)
		}

		// Cleanup any stale temporary files from previous interrupted attempts
		if tmpFiles, err := filepath.Glob(downloadPath + ".tmp.*"); err == nil {
			for _, tmpFile := range tmpFiles {
				os.Remove(tmpFile)
			}
		}

		// Use a randomized temporary path for downloading to ensure atomicity and concurrency safety
		// Similar to how homebrew and mise handle incomplete downloads.
		randSuffix, _ := env.RandomString(8)
		downloadTmpPath := fmt.Sprintf("%s.tmp.%s", downloadPath, randSuffix)

		// Start a spinner for the connection/download phase
		var spinner *pterm.SpinnerPrinter
		if !quietProgress {
			spinner, _ = pterm.DefaultSpinner.
				WithText(fmt.Sprintf("Connecting to %s...", tool)).
				Start()
		}

		// Initialize progress bar
		var progressbar *pterm.ProgressbarPrinter
		var lastDownloaded int64
		var lastUpdateTime time.Time
		var progressMutex sync.Mutex

		opts.ProgressCallback = func(downloaded, total int64) {
			if quietProgress {
				if reporter, ok := ctx.Value(ContextKeyProgressReporter).(ProgressReporter); ok {
					reporter(tool, downloaded, total)
				}
				return
			}
			progressMutex.Lock()
			defer progressMutex.Unlock()

			// Stop the connection spinner once we start receiving bytes,
			// BUT ONLY if we are going to start a progress bar.
			if spinner != nil && total > 0 {
				spinner.Stop()
				spinner = nil
			}

			// Initialize progress bar if not already done and total is known
			if progressbar == nil && total > 0 {
				progressbar, _ = pterm.DefaultProgressbar.
					WithTotal(int(total)).
					WithTitle(fmt.Sprintf("Downloading %s (%s)", tool, humanize.Bytes(uint64(total)))).
					WithShowCount(false).
					Start()
				lastUpdateTime = time.Now()
			}

			// Throttle updates to prevent terminal rendering bottlenecks (max 10 updates per second)
			now := time.Now()
			if now.Sub(lastUpdateTime) > 100*time.Millisecond || downloaded >= total {
				if progressbar != nil {
					diff := downloaded - lastDownloaded
					if diff != 0 {
						// MUST update title BEFORE Add().
						// If Add() reaches 100%, pterm internally calls Stop() and freezes the UI.
						// Any title update after Add() would be completely ignored on the final frame.
						progressbar.UpdateTitle(fmt.Sprintf("Downloading %s (%s/%s)",
							tool,
							humanize.Bytes(uint64(downloaded)),
							humanize.Bytes(uint64(total))))

						if diff < 0 {
							// Download was reset (e.g. fallback from concurrent to sequential)
							lastDownloaded = downloaded
						} else {
							progressbar.Add(int(diff))
							lastDownloaded = downloaded
						}

						lastUpdateTime = now
					}
					if total > 0 && downloaded >= total {
						progressbar.Stop()
					}
				} else if spinner != nil {
					// Update spinner text if we don't have a progress bar (unknown total)
					spinner.UpdateText(fmt.Sprintf("Downloading %s (%s)...", tool, humanize.Bytes(uint64(downloaded))))
					lastUpdateTime = now
				}
			}
		}

		if err := downloader.Download(ctx, versionInfo.DownloadURL, downloadTmpPath, opts); err != nil {
			if spinner != nil {
				spinner.Stop()
			}
			if progressbar != nil {
				progressbar.Stop()
			}
			os.Remove(downloadTmpPath)
			return fmt.Errorf("failed to download: %w", err)
		}
		if spinner != nil {
			spinner.Stop()
		}
		if progressbar != nil {
			progressbar.Stop()
		}

		// Atomic rename from temp download path to final download path
		if err := os.Rename(downloadTmpPath, downloadPath); err != nil {
			os.Remove(downloadTmpPath)
			return fmt.Errorf("failed to finalize download: %w", err)
		}

		if !quietProgress {
			output.Successf("✓ downloaded to %s", downloadPath)
		}
		defer func() {
			if im.settings != nil && im.settings.AlwaysKeepDownload {
				logger.Debug("AlwaysKeepDownload is enabled, keeping artifact", map[string]interface{}{"path": downloadPath})
				return
			}
			os.Remove(downloadPath)
			// Clean up empty parent directories up to the downloads root
			im.removeEmptyDirs(downloadPath, env.GetDownloadsDir())
		}()

		// Verify checksum
		if versionInfo.Checksum != "" && versionInfo.Metadata["skip_checksum"] != "1" {
			if err := downloader.VerifyChecksum(ctx, downloadPath, versionInfo.Checksum); err != nil {
				return fmt.Errorf("checksum verification failed: %w", err)
			}
		}
		// 5.5 GPG Signature Verification
		verifyMetadata := true
		if im.settings != nil && im.settings.VerifyMetadata != nil {
			verifyMetadata = *im.settings.VerifyMetadata
		}
		if v := env.Get("VERIFY_METADATA"); v != "" {
			verifyMetadata = (v == "1" || strings.ToLower(v) == "true" || strings.ToLower(v) == "yes")
		}

		if (versionInfo.SignatureURL != "" || versionInfo.GPGSignature != "") && verifyMetadata && (im.settings == nil || im.settings.GPGVerify != "off") {
			fmt.Printf("ℹ verifying GPG signature for %s@%s...\n", tool, version)

			sigPath := downloadPath + ".asc"
			var downloadErr error
			if versionInfo.GPGSignature != "" {
				// Use embedded signature
				if err := os.WriteFile(sigPath, []byte(versionInfo.GPGSignature), 0644); err != nil {
					downloadErr = fmt.Errorf("failed to save embedded GPG signature: %w", err)
				}
			} else {
				// Download signature file
				downloader, _ := im.downloadManager.Get("https")
				sigOpts := download.DefaultDownloadOptions()
				downloadErr = downloader.Download(ctx, versionInfo.SignatureURL, sigPath, sigOpts)
			}

			if downloadErr != nil {
				msg := fmt.Sprintf("failed to obtain GPG signature: %v", downloadErr)
				if im.settings != nil && im.settings.GPGVerify == "strict" {
					return fmt.Errorf("GPG signature required in strict mode: %w", downloadErr)
				}
				if !quietProgress {
					output.Warningf("⚠️  WARNING: %s. Continuing anyway (GPGVerify=%s)\n", msg, im.settings.GPGVerify)
				}
				gpgStatus = "Failed (Download)"
			} else {
				defer os.Remove(sigPath)

				// Collect all trusted keys (Explicit + Bundled/Lockfile)
				trustedKeys := make([]string, 0)
				if im.settings != nil {
					trustedKeys = append(trustedKeys, im.settings.GPGKeys...)
				}
				if tc, ok := im.toolConfigs[tool]; ok {
					trustedKeys = append(trustedKeys, tc.GPGKeys...)
				}
				trustedKeys = append(trustedKeys, versionInfo.GPGKeys...)

				// Verify signature
				err := im.gpgVerifier.Verify(ctx, sigPath, downloadPath, trustedKeys)
				if err != nil && strings.Contains(err.Error(), "missing public key") && len(trustedKeys) > 0 {
					// Handle missing public key: Ask user in TTY, or fail in CI
					if pterm.PrintColor && !pterm.RawOutput { // Check if we are likely in a TTY
						output.Warningf("⚠️  GPG signature found but public key is missing locally.")
						fp := trustedKeys[0] // Try first fingerprint
						confirm, _ := pterm.DefaultInteractiveConfirm.
							WithDefaultText(fmt.Sprintf("Do you want to trust and import GPG key %s from keyservers?", fp)).
							Show()

						if confirm {
							spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Importing GPG key %s...", fp))
							if importErr := im.gpgVerifier.ImportKey(ctx, fp); importErr == nil {
								spinner.Success("GPG key imported successfully")
								// Retry verification
								err = im.gpgVerifier.Verify(ctx, sigPath, downloadPath, trustedKeys)
							} else {
								spinner.Fail(fmt.Sprintf("Failed to import GPG key: %v", importErr))
							}
						}
					} else {
						if !quietProgress {
							output.Warningf("⚠️  GPG verification skipped: missing public key (Non-interactive mode)\n")
						}
						gpgStatus = "Failed (Missing Key)"
					}
				}

				if err != nil {
					msg := fmt.Sprintf("GPG verification failed: %v", err)
					if im.settings != nil && im.settings.GPGVerify == "strict" {
						return fmt.Errorf("SECURITY ERROR: %s", msg)
					}
					if !quietProgress {
						output.Warningf("⚠️  SECURITY WARNING: %s. Continuing anyway (GPGVerify=%s)\n", msg, im.settings.GPGVerify)
					}
					gpgStatus = "Failed (Invalid)"
				} else {
					if !quietProgress {
						output.Successf("✓ GPG signature verified successfully")
					}
					gpgStatus = "Verified"
				}
			}
		} else if versionInfo.SignatureURL == "" && im.settings != nil && im.settings.GPGVerify == "strict" {
			// Intelligent handling: If we have fingerprints but no signature URL, it's a violation.
			// If we have neither, the tool likely doesn't support GPG, so we warn and allow SHA256 fallback.
			if len(versionInfo.GPGKeys) > 0 {
				return fmt.Errorf("GPG security violation: trusted fingerprints exist for %s but no signature URL found in strict mode", tool)
			}
			if !quietProgress {
				fmt.Printf("ℹ %s does not appear to support GPG signatures. Falling back to strong SHA256 checksum verification.\n", tool)
			}
			gpgStatus = "NotSupported"
		}

		// Verify provenance (SLSA attestation) if the backend is github, ubi, or gitlab.
		// If the project does not publish attestations, this is a no-op.
		// If attestations exist, verification MUST pass to prevent supply chain attacks.
		if (backendName == "github" || backendName == "ubi" || backendName == "gitlab") && verifyMetadata {
			// Extract owner/repo from tool string (expected format: "owner/repo").
			provenanceStatus, provenanceErr := tryVerifyProvenance(ctx, backendName, tool, downloadPath)
			if provenanceErr != nil {
				return fmt.Errorf("%s provenance verification failed: %w", backendName, provenanceErr)
			}
			// Store result in versionInfo metadata for audit logging below.
			if versionInfo.Metadata == nil {
				versionInfo.Metadata = make(map[string]string)
			}
			versionInfo.Metadata["provenance_status"] = provenanceStatus
		}

		// Write resolved info back into the lockfile so future installs can
		// skip the remote API call (lock hit) and use the cached URL directly.
		if im.lockService != nil {
			if lockErr := im.lockService.RecordInstall(tool, backendName, versionInfo); lockErr != nil {
				// Non-fatal: log but don't abort the install.
				logger.Warn("lockfile: failed to record install", map[string]interface{}{
					"tool":  tool,
					"error": lockErr.Error(),
				})
			}
		}
	}

	// 6. Install using provider with atomic rename strategy
	// Standardize tool name for filesystem (Scheme B: provider-tool-name)
	fsToolName = env.GetFSToolName(tool, backendName)

	installPath := filepath.Join(env.GetInstallsDir(), fsToolName, version)
	tmpInstallPath := installPath + ".unirtm-tmp"

	// Clean up any stale directories from previous failed attempts
	os.RemoveAll(tmpInstallPath)
	// If final path exists but we reached here, it means it's not in the database.
	// We'll overwrite it to be safe.
	os.RemoveAll(installPath)

	if err := os.MkdirAll(filepath.Dir(tmpInstallPath), 0755); err != nil {
		return fmt.Errorf("failed to create installs directory: %w", err)
	}

	p := im.providerRegistry.GetWithBackend(tool, backendName)

	if !quietProgress {
		fmt.Printf("ℹ extracting %s@%s...\n", tool, version)
	}
	if err := p.Install(ctx, tool, tmpInstallPath, downloadPath, version); err != nil {
		os.RemoveAll(tmpInstallPath)
		return fmt.Errorf("installation failed: %w", err)
	}

	if !quietProgress {
		fmt.Printf("ℹ running post-install hooks for %s@%s...\n", tool, version)
	}
	if err := p.PostInstall(ctx, tool, tmpInstallPath, version); err != nil {
		os.RemoveAll(tmpInstallPath)
		return fmt.Errorf("post-install failed: %w", err)
	}

	// Atomic rename from temp to final path
	if err := os.Rename(tmpInstallPath, installPath); err != nil {
		os.RemoveAll(tmpInstallPath)
		return fmt.Errorf("failed to finalize installation: %w", err)
	}

	if !quietProgress {
		fmt.Printf("ℹ recording %s@%s to database...\n", tool, version)
	}
	// Record installation
	installation := &repository.Installation{
		Tool:        tool,
		Version:     version,
		Backend:     backendName,
		InstallPath: installPath,
		Checksum:    versionInfo.Checksum,
	}

	// Start transaction for recording
	tx, err := im.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := tx.InstallationRepo().Upsert(ctx, installation); err != nil {
		os.RemoveAll(installPath)
		return fmt.Errorf("failed to record installation: %w", err)
	}

	// Record audit entry
	auditMeta := map[string]string{"gpg": gpgStatus}
	if versionInfo.Metadata != nil {
		if ps, ok := versionInfo.Metadata["provenance_status"]; ok && ps != "" {
			auditMeta["provenance"] = ps
		}
	}
	auditMetaJSON, _ := json.Marshal(auditMeta)
	auditEntry := &repository.AuditEntry{
		Operation:       "install",
		Tool:            tool,
		Version:         version,
		Status:          "success",
		GpgVerification: gpgStatus,
		Metadata:        string(auditMetaJSON),
	}
	if err := tx.AuditRepo().Log(ctx, auditEntry); err != nil {
		// Log but don't fail
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		os.RemoveAll(installPath)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 11. Run PostInstall hook
	if postInstall != "" {
		output.Warningf("⚠️  SECURITY: executing post_install hook for %s: %s", tool, postInstall)
		if err := im.executeHook(ctx, postInstall, tool, version); err != nil {
			return fmt.Errorf("post_install hook failed: %w", err)
		}
	}

	// 12. Generate shims for the tool executables
	if !quietProgress {
		fmt.Printf("ℹ generating shims for %s...\n", tool)
	}
	execs, _ := p.ListExecutables(tool, installPath, version)
	if err := im.shimGenerator.GenerateShim(ctx, tool, execs...); err != nil {
		if !quietProgress {
			output.Warningf("⚠️  WARNING: failed to generate shims for %s: %v", tool, err)
		}
		// Non-fatal, don't return error
	}

	if !quietProgress {
		output.Successf("✓ %s@%s installed successfully to %s", tool, version, installPath)
	}
	return nil
}

// ToolToInstall represents a tool with its resolved backend information for sorting.
type ToolToInstall struct {
	OriginalName string
	ToolName     string
	BackendName  string
	Version      string
	Config       config.ToolConfig
}

// SortTools sorts tools such that dependencies are installed before dependent tools.
func (im *InstallationManager) SortTools(tools map[string]config.ToolConfig) []ToolToInstall {
	var result []ToolToInstall
	for name, tc := range tools {
		toolName := name
		backendName := tc.Backend

		// Resolve backend and tool name similarly to EnsureInstalled
		if backendName == "" {
			if idx := strings.Index(name, ":"); idx != -1 {
				backendName = name[:idx]
				toolName = name[idx+1:]
			} else if strings.Contains(name, "/") {
				backendName = "github"
			}
		}

		// Intercept go: prefix
		if backendName == "go" || strings.HasPrefix(name, "go:") {
			backendName = "go-pkg"
			if strings.HasPrefix(name, "go:") {
				toolName = strings.TrimPrefix(name, "go:")
			}
		}

		result = append(result, ToolToInstall{
			OriginalName: name,
			ToolName:     toolName,
			BackendName:  backendName,
			Version:      tc.Version,
			Config:       tc,
		})
	}

	// Simple topological sort or layered sorting.
	// Since dependencies are mostly backend-based (e.g., npm depends on node),
	// we can use a simple priority system or a more formal topological sort.
	// Let's use a simple topological sort logic here.

	// Map tool name to its info for easy lookup
	toolMap := make(map[string]*ToolToInstall)
	for i := range result {
		toolMap[result[i].ToolName] = &result[i]
	}

	// Resulting sorted slice
	sorted := make([]ToolToInstall, 0, len(result))
	visited := make(map[string]bool)
	tempVisited := make(map[string]bool)

	var visit func(t ToolToInstall) error
	visit = func(t ToolToInstall) error {
		if tempVisited[t.ToolName] {
			// Circular dependency detected, but we'll just ignore and proceed
			return nil
		}
		if !visited[t.ToolName] {
			tempVisited[t.ToolName] = true

			// Get backend dependencies
			if b, err := backend.Get(t.BackendName); err == nil {
				for _, dep := range b.Dependencies() {
					// Check if this dependency is also in our tools list
					if depTool, ok := toolMap[dep]; ok {
						visit(*depTool)
					}
				}
			}

			tempVisited[t.ToolName] = false
			visited[t.ToolName] = true
			sorted = append(sorted, t)
		}
		return nil
	}

	for _, t := range result {
		visit(t)
	}

	return sorted
}

// SortToolsFromSpecs sorts tools such that dependencies are installed before dependent tools.
func (im *InstallationManager) SortToolsFromSpecs(tools map[string]ToolSpec) []ToolToInstall {
	var result []ToolToInstall
	for name, ts := range tools {
		result = append(result, ToolToInstall{
			OriginalName: name,
			ToolName:     ts.Name,
			BackendName:  ts.BackendName,
			Version:      ts.Version,
		})
	}

	// Map tool name to its info for easy lookup
	toolMap := make(map[string]*ToolToInstall)
	for i := range result {
		toolMap[result[i].ToolName] = &result[i]
	}

	// Resulting sorted slice
	sorted := make([]ToolToInstall, 0, len(result))
	visited := make(map[string]bool)
	tempVisited := make(map[string]bool)

	var visit func(t ToolToInstall) error
	visit = func(t ToolToInstall) error {
		if tempVisited[t.ToolName] {
			return nil
		}
		if !visited[t.ToolName] {
			tempVisited[t.ToolName] = true

			if b, err := backend.Get(t.BackendName); err == nil {
				for _, dep := range b.Dependencies() {
					if depTool, ok := toolMap[dep]; ok {
						visit(*depTool)
					}
				}
			}

			tempVisited[t.ToolName] = false
			visited[t.ToolName] = true
			sorted = append(sorted, t)
		}
		return nil
	}

	for _, t := range result {
		visit(t)
	}

	return sorted
}

// EnsureInstalled checks if all tools in the configuration are installed,
// and installs any missing ones.
func (im *InstallationManager) EnsureInstalled(ctx context.Context, tools map[string]config.ToolConfig) error {
	specs := make(map[string]ToolSpec, len(tools))
	for name, tc := range tools {
		backendName, toolName, version, _ := im.ParseToolSpec(name)
		if tc.Backend != "" {
			backendName = tc.Backend
		}
		if tc.Version != "" {
			version = tc.Version
		}
		specs[name] = ToolSpec{
			Name:         toolName,
			Version:      version,
			BackendName:  backendName,
			OriginalName: name,
		}
	}
	return im.EnsureInstalledFromSpecs(ctx, specs)
}

// EnsureInstalledFromSpecs checks if all tools in the specs are installed,
// and installs any missing ones.
func (im *InstallationManager) EnsureInstalledFromSpecs(ctx context.Context, tools map[string]ToolSpec) error {
	sortedTools := im.SortToolsFromSpecs(tools)

	for _, t := range sortedTools {
		toolName := t.ToolName
		backendName := t.BackendName
		version := t.Version

		// Check if already installed using robust variant detection
		installed, _ := im.IsInstalled(ctx, toolName, version, backendName)
		if installed {
			continue
		}

		// Not installed, proceed with installation
		fmt.Printf("ℹ auto-installing missing tool: %s@%s\n", toolName, version)
		if err := im.Install(ctx, t.OriginalName, toolName, version, backendName); err != nil {
			if err == ErrAlreadyInstalled {
				continue
			}
			return fmt.Errorf("auto-install failed for %s: %w", toolName, err)
		}
	}
	return nil
}

// Uninstall removes a tool installation.
func (im *InstallationManager) Uninstall(ctx context.Context, tool, version string) error {
	installation, err := im.installRepo.FindByToolAndVersion(ctx, tool, version)
	if err != nil {
		return fmt.Errorf("installation not found: %w", err)
	}

	// Retrieve provider and list all executables before removing the directory
	p := im.providerRegistry.GetWithBackend(tool, installation.Backend)
	execs, _ := p.ListExecutables(tool, installation.InstallPath, version)
	if len(execs) == 0 {
		execs = []string{tool}
	}

	// Run provider uninstall
	if err := p.Uninstall(ctx, tool, installation.InstallPath, version); err != nil {
		return fmt.Errorf("provider uninstall failed: %w", err)
	}

	// Remove installation directory
	if err := os.RemoveAll(installation.InstallPath); err != nil {
		return fmt.Errorf("failed to remove installation directory: %w", err)
	}

	// Clean up parent directory if empty (ignoring system/hidden files like .DS_Store)
	parentDir := filepath.Dir(installation.InstallPath)
	if entries, err := os.ReadDir(parentDir); err == nil {
		isEmpty := true
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".") {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			_ = os.RemoveAll(parentDir)
		}
	}

	// Start transaction for removal
	tx, err := im.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove from database
	if err := tx.InstallationRepo().Delete(ctx, tool, version); err != nil {
		return fmt.Errorf("failed to remove installation record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Check if other versions of this tool are still installed
	installations, err := im.installRepo.List(ctx)
	otherVersionsExist := false
	if err == nil {
		for _, inst := range installations {
			if inst.Tool == tool && inst.Version != version {
				otherVersionsExist = true
				break
			}
		}
	}

	// If no other versions exist, remove the shim files for all executables of this tool
	if !otherVersionsExist {
		for _, exe := range execs {
			_ = im.shimGenerator.RemoveShim(ctx, exe)
		}
	}

	// Remove from lockfile
	if im.lockService != nil {
		_ = im.lockService.RemoveTool(tool)
	}

	return nil
}

// tryVerifyProvenance runs GitHub/GitLab provenance (SLSA attestation) verification
// for tools whose tool string is in "owner/repo" format.
//
// Returns a short status string for audit logging:
//   - "not_applicable" — tool format is not owner/repo (e.g. language runtimes)
//   - "not_supported"  — project published no attestations (safely skipped)
//   - "verified"       — attestation present and verified successfully
//
// An error is returned only when attestations exist but verification fails,
// which is treated as a hard security failure.
func tryVerifyProvenance(ctx context.Context, backendName, tool, artifactPath string) (string, error) {
	// Support skipping provenance verification via environment variables
	// (e.g. UNIRTM_VERIFY_PROVENANCE=0 or MISE_VERIFY_PROVENANCE=0).
	// This is highly useful for users in offline or restricted network environments (like China).
	if v := env.Get("VERIFY_PROVENANCE"); v == "0" || strings.ToLower(v) == "false" {
		logger.Warn("⚠️ provenance verification is disabled via environment, skipping check")
		return "skipped", nil
	}

	parts := strings.SplitN(tool, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "not_applicable", nil
	}
	owner, repo := parts[0], parts[1]

	var result *backend.ProvenanceResult
	var err error

	if backendName == "gitlab" {
		token := env.Get("GITLAB_TOKEN")
		result, err = backend.VerifyGitlabArtifactProvenance(ctx, token, owner, repo, artifactPath)
	} else {
		// Resolve a token (reuses the 6-tier resolver from the GitHub backend).
		token := backend.ResolveGitHubTokenPublic("github.com")
		result, err = backend.VerifyArtifactProvenance(ctx, token, owner, repo, artifactPath)
	}

	if err != nil {
		return "failed", err
	}

	if !result.Supported {
		logger.Debug("provenance not supported, skipping", map[string]interface{}{"tool": tool, "backend": backendName})
		return "not_supported", nil
	}

	logger.Debug("provenance verified", map[string]interface{}{
		"tool":          tool,
		"backend":       backendName,
		"repository":    result.Repository,
		"workflowRef":   result.WorkflowRef,
		"predicateType": result.PredicateType,
	})
	return "verified", nil
}

// ResolveToolEnvBySpec returns the environment variables exported by an installed
// tool identified by its name, version, and backend.  It is used by the exec
// sub-command to inject per-tool environment variables (e.g. GOROOT, JAVA_HOME,
// UNIRTM_<TOOL>_VERSION) and bin-directory PATH entries.
//
// If the tool is not installed or the provider does not export env vars the
// function returns an empty (non-nil) map and a nil error.
func (im *InstallationManager) ResolveToolEnvBySpec(
	toolName, version, backendName string,
) map[string]string {
	result := make(map[string]string)

	// Compute the expected installation path.
	fsName := env.GetFSToolName(toolName, backendName)
	installPath := filepath.Join(env.GetInstallsDir(), fsName, version)
	if _, err := os.Stat(installPath); err != nil {
		// Tool is not installed — nothing to inject.
		return result
	}

	// Ask the provider for its canonical environment exports.
	p := im.providerRegistry.GetWithBackend(toolName, backendName)
	if p != nil {
		if envVars, err := p.GetEnvVars(toolName, installPath, version); err == nil {
			for k, v := range envVars {
				result[k] = v
			}
		}
	}

	// Always export UNIRTM_<TOOL>_VERSION so shims and scripts can read it.
	envKey := "UNIRTM_" + strings.ToUpper(strings.ReplaceAll(toolName, "-", "_")) + "_VERSION"
	result[envKey] = version

	// Prepend the tool's bin directory to PATH.
	for _, binDir := range []string{
		filepath.Join(installPath, "bin"),
		installPath,
	} {
		if fi, err := os.Stat(binDir); err == nil && fi.IsDir() {
			if existing, ok := result["PATH"]; ok && existing != "" {
				result["PATH"] = binDir + string(os.PathListSeparator) + existing
			} else {
				result["PATH"] = binDir
			}
			break // Only prepend the first valid bin directory.
		}
	}

	return result
}

// archiveExtensions contains file extensions that are never directly executable.
// Files with these extensions must never be passed to exec/syscall.Exec.
var archiveExtensions = map[string]bool{
	".zst": true, ".gz": true, ".bz2": true, ".xz": true, ".lz4": true,
	".tar": true, ".tgz": true, ".tbz2": true, ".txz": true,
	".zip": true, ".7z": true, ".rar": true,
	".deb": true, ".rpm": true, ".apk": true,
	".sig": true, ".asc": true, ".sha256": true, ".sha512": true,
}

// isExecutableFile returns true if path is a regular file that can be executed.
// It rejects directories, archive/compressed files, and (on Unix) files
// without the execute permission bit.
func isExecutableFile(path string) bool {
	// Reject known archive extensions regardless of platform.
	ext := strings.ToLower(filepath.Ext(path))
	if archiveExtensions[ext] {
		return false
	}
	// Handle double-extensions like .tar.gz — check last two components.
	base := strings.ToLower(filepath.Base(path))
	for _, suffix := range []string{".tar.gz", ".tar.xz", ".tar.bz2", ".tar.zst", ".tar.lz4"} {
		if strings.HasSuffix(base, suffix) {
			return false
		}
	}

	fi, err := os.Stat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return false
	}

	// On Unix, verify the execute permission bit is set.
	if runtime.GOOS != "windows" {
		return fi.Mode()&0111 != 0
	}

	// On Windows, trust the extension (.exe, .cmd, .bat, .ps1).
	winExec := map[string]bool{".exe": true, ".cmd": true, ".bat": true, ".ps1": true}
	return winExec[ext] || ext == ""
}

// ResolveExecutable finds the absolute path and environment variables for a given executable name
// by searching through installed tools in the current context.
func (im *InstallationManager) ResolveExecutable(ctx context.Context, exeName string, platform backend.Platform) (string, map[string]string, error) {
	// 1. Get all installations from repository
	installations, err := im.installRepo.List(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("list installations: %w", err)
	}

	// 2. Filter candidates that provide the executable
	type candidate struct {
		inst    *repository.Installation
		exePath string
		// exactMatch is true when the base name equals exeName exactly.
		exactMatch bool
	}
	var candidates []candidate

	for _, inst := range installations {
		// If the tool is configured in the current context, only consider the active version.
		if tc, ok := im.toolConfigs[inst.Tool]; ok {
			expectedVersion := im.resolveAlias(inst.Tool, tc.Version)
			// Parse the expected version out in case it's a path@version format
			_, _, parsedVersion, explicit := im.ParseToolSpec(inst.Tool + "@" + expectedVersion)
			if explicit && inst.Version != parsedVersion {
				continue
			}
		}

		p := im.providerRegistry.GetWithBackend(inst.Tool, inst.Backend)
		if p == nil {
			continue
		}

		execs, err := p.ListExecutables(inst.Tool, inst.InstallPath, inst.Version)
		if err != nil {
			continue
		}

		for _, exec := range execs {
			baseName := filepath.Base(exec)

			// Resolve absolute path first so we can stat it.
			absPath := exec
			if !filepath.IsAbs(exec) {
				absPath = filepath.Join(inst.InstallPath, exec)
			}

			// Hard filter: skip files that are not actually executable binaries.
			if !isExecutableFile(absPath) {
				continue
			}

			exact := baseName == exeName
			prefix := !exact && strings.HasPrefix(baseName, exeName)
			if prefix {
				remainder := baseName[len(exeName):]
				// Only accept prefix match when remainder starts with a version separator.
				// This avoids matching "python-3.14.4.zst" for "python".
				if len(remainder) == 0 || (remainder[0] != '-' && remainder[0] != '_' && remainder[0] != '@' && remainder[0] != '.') {
					prefix = false
				}
				// Reject if the remainder still contains a file extension we consider
				// non-executable (e.g. the full name is "python-3.14.4.zst").
				if prefix {
					remExt := strings.ToLower(filepath.Ext(baseName))
					if archiveExtensions[remExt] {
						prefix = false
					}
				}
			}

			if exact || prefix {
				candidates = append(candidates, candidate{
					inst:       inst,
					exePath:    absPath,
					exactMatch: exact,
				})
				break
			}
		}
	}

	if len(candidates) == 0 {
		return "", nil, fmt.Errorf("executable %s not found", exeName)
	}

	// 3. Pick the best candidate:
	//    a) exact base-name match over prefix match
	//    b) among equals, prefer the tool whose name matches the executable name
	selected := candidates[0]
	for _, c := range candidates {
		// Prefer exact name match over prefix match.
		if c.exactMatch && !selected.exactMatch {
			selected = c
			break
		}
		// Among candidates of equal match quality, prefer the tool whose name
		// matches the executable (e.g. tool "python" for exe "python").
		if c.exactMatch == selected.exactMatch &&
			(c.inst.Tool == exeName || filepath.Base(c.inst.Tool) == exeName) {
			selected = c
			break
		}
	}

	// 4. Get environment variables for the selected tool
	p := im.providerRegistry.GetWithBackend(selected.inst.Tool, selected.inst.Backend)
	envVars, _ := p.GetEnvVars(selected.inst.Tool, selected.inst.InstallPath, selected.inst.Version)

	return selected.exePath, envVars, nil
}

// ParseToolSpec parses a tool specification string (e.g., "node@20", "github:cli/cli@v2.0.0")
// into its constituent parts: backend, tool name, version, and whether the version was explicit.
func (im *InstallationManager) ParseToolSpec(spec string) (backend, tool, version string, explicit bool) {
	version = "latest"
	tool = spec

	// 1. Handle backend:tool[@version]
	if idx := strings.Index(spec, ":"); idx != -1 {
		backend = spec[:idx]
		tool = spec[idx+1:]
		// Intercept go: prefix and route to the internal go-pkg provider
		if backend == "go" {
			backend = "go-pkg"
		}
	}

	// 2. Handle tool[@version]
	// The version separator is the LAST '@' that is NOT at index 0 of the tool part.
	if idx := strings.LastIndex(tool, "@"); idx > 0 {
		version = tool[idx+1:]
		tool = tool[:idx]
		explicit = true
		if version == "" {
			version = "latest"
		}
	}

	// 3. Auto-detect backend if not specified
	if backend == "" {
		backend = im.AutoDetectBackend(tool)
	}

	return backend, tool, version, explicit
}

// AutoDetectBackend attempts to identify the best backend for a given tool name.
func (im *InstallationManager) AutoDetectBackend(toolName string) string {
	if strings.Contains(toolName, "/") {
		return "github"
	}
	if native.IsNativeTool(toolName) {
		return "native"
	}
	// Default to asdf for compatibility with the broadest range of tools
	return "asdf"
}

// removeEmptyDirs recursively removes empty parent directories of path up to root.
func (im *InstallationManager) removeEmptyDirs(path string, root string) {
	dir := filepath.Dir(path)
	for {
		if dir == root || dir == "." || dir == filepath.Dir(root) {
			break
		}
		// Try to remove the directory. os.Remove only removes if empty.
		if err := os.Remove(dir); err != nil {
			break
		}
		dir = filepath.Dir(dir)
	}
}
