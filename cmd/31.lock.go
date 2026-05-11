// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/lockfile"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	// lockAllPlatforms generates lockfile entries for all standard platforms.
	lockAllPlatforms bool

	// lockPlatforms is a comma-separated list of platform keys to generate.
	lockPlatforms string

	// lockCheck validates the lockfile without regenerating.
	lockCheck bool
)

// init registers the lock command.
func init() {
	lockCmd.Flags().BoolVar(&lockAllPlatforms, "all-platforms", false,
		"generate lockfile entries for all standard platforms")
	lockCmd.Flags().StringVar(&lockPlatforms, "platform", "",
		"comma-separated list of platform keys (e.g. linux-amd64,macos-arm64)")
	lockCmd.Flags().BoolVar(&lockCheck, "check", false,
		"validate the lockfile without regenerating (exits non-zero on problems)")

	if rootCmd != nil {
		rootCmd.AddCommand(lockCmd)
	}
}

// lockCmd is the `unirtm lock` command.
var lockCmd = &cobra.Command{
	Use:   "lock [tool...]",
	Short: "Generate or update the unirtm.lock lockfile",
	Long: `Generate or update the unirtm.lock lockfile.

unirtm.lock pins exact tool versions, download URLs, and checksums for
reproducible installs across environments. It is UniRTM's equivalent of
package-lock.json (npm) or Cargo.lock (Rust).

Key benefits:
  • Reproducible builds: everyone uses exactly the same version
  • Avoids API rate limits: URLs are cached, so GitHub API is not called on install
  • Security: checksums are verified against the lockfile on each install
  • Offline installs: combine with UNIRTM_LOCKED=1 for fully offline CI

Once generated, unirtm install automatically uses the lockfile and keeps it
up to date after each successful installation.

Examples:
  # Generate / refresh the lockfile for the current platform
  unirtm lock

  # Generate entries for all standard platforms (for CI reproducibility)
  unirtm lock --all-platforms

  # Generate only for specific platforms
  unirtm lock --platform linux-amd64,macos-arm64

  # Refresh only selected tools
  unirtm lock cli/cli astral-sh/ruff

  # Validate the lockfile without regenerating (CI gate)
  unirtm lock --check

Environment variables:
  UNIRTM_LOCK_FILE   Override lockfile path (default: ./unirtm.lock)
  UNIRTM_LOCKED=1    Strict mode: installs fail unless URL is in lockfile`,
	RunE: runLock,
}

// runLock executes the lock command.
func runLock(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	lockPath := env.GetLockFilePath()

	// ── --check mode ──────────────────────────────────────────────────────────
	if lockCheck {
		return runLockCheck(formatter, lockPath)
	}

	// ── Determine platforms ───────────────────────────────────────────────────
	var platforms []string
	var err error

	switch {
	case lockAllPlatforms:
		platforms = lockfile.StandardPlatforms
	case lockPlatforms != "":
		platforms, err = lockfile.ParsePlatformKeys(lockPlatforms)
		if err != nil {
			formatter.Error(err.Error())
			return err
		}
	default:
		// Default: all standard platforms (parity with mise).
		platforms = lockfile.StandardPlatforms
	}

	formatter.Info(fmt.Sprintf("Generating lockfile: %s", lockPath), map[string]interface{}{
		"platforms": strings.Join(platforms, ", "),
	})

	// Load project configuration
	cfg, err := config.LoadProjectConfig()
	if err != nil {
		formatter.Warning(fmt.Sprintf("Failed to load project config: %v", err))
	} else {
		// Apply [env] variables from config to current process
		cfg.ApplyEnvironment()
	}

	var tools map[string]service.ToolSpec
	if cfg != nil && len(cfg.Tools) > 0 {
		tools = make(map[string]service.ToolSpec, len(cfg.Tools))
		for name, tc := range cfg.Tools {
			backendName := tc.Backend
			toolName := name
			if backendName == "" {
				if idx := strings.Index(name, ":"); idx != -1 {
					backendName = name[:idx]
					toolName = name[idx+1:]
				} else if strings.Contains(name, "/") {
					backendName = "github"
				} else {
					backendName = "asdf"
				}
			}
			tools[name] = service.ToolSpec{
				Name:        toolName,
				Version:     tc.Version,
				BackendName: backendName,
			}
		}
	} else {
		tools = make(map[string]service.ToolSpec)
	}

	// CLI arguments override / filter the config tools.
	if len(args) > 0 {
		subset := make(map[string]service.ToolSpec)
		for _, arg := range args {
			toolName, version, backendName := parseLockToolArg(arg)
			if spec, ok := tools[toolName]; ok {
				if version != "" {
					spec.Version = version
				}
				if backendName != "" {
					spec.BackendName = backendName
				}
				subset[toolName] = spec
			} else {
				if version == "" {
					version = "latest"
				}
				if backendName == "" {
					backendName = "github"
				}
				subset[toolName] = service.ToolSpec{Version: version, BackendName: backendName}
			}
		}
		tools = subset
	}

	if len(tools) == 0 {
		formatter.Warning("No tools found. Add tools to unirtm.yaml or pass them as arguments.")
		return nil
	}

	// ── Create LockService ────────────────────────────────────────────────────
	lockSvc, err := service.NewLockService(service.LockServiceOptions{
		LockfilePath: lockPath,
	})
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to initialise lock service: %v", err))
		return err
	}

	backendRegistry := backend.NewRegistry()
	lockSvc.SetBackendRegistry(backendRegistry)

	// ── Progress display ──────────────────────────────────────────────────────
	spinner, _ := pterm.DefaultSpinner.Start(
		fmt.Sprintf("Resolving %d tool(s) for %d platform(s)...", len(tools), len(platforms)),
	)

	ctx := context.Background()
	genErr := lockSvc.Generate(ctx, tools, service.GenerateOptions{
		Platforms: platforms,
	})

	if genErr != nil {
		spinner.Fail("Failed to generate lockfile")
		formatter.Error(genErr.Error())
		return genErr
	}

	spinner.Success(fmt.Sprintf("Lockfile written: %s", lockPath))
	printLockSummaryTable(tools, platforms)
	return nil
}

// runLockCheck validates the lockfile without regenerating.
func runLockCheck(formatter output.Formatter, lockPath string) error {
	lf, err := lockfile.Load(lockPath)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load lockfile %q: %v", lockPath, err))
		return err
	}

	if lf.IsEmpty() {
		formatter.Warning(fmt.Sprintf(
			"Lockfile %q is empty. Run `unirtm lock` to generate it.", lockPath))
		return fmt.Errorf("lockfile is empty")
	}

	if err := lf.Validate(); err != nil {
		formatter.Error(fmt.Sprintf("Lockfile validation failed:\n%v", err))
		return err
	}

	formatter.Success(fmt.Sprintf("Lockfile %q is valid.", lockPath))
	return nil
}

// printLockSummaryTable renders a table of locked tool entries.
func printLockSummaryTable(tools map[string]service.ToolSpec, platforms []string) {
	tableData := pterm.TableData{
		{"Tool", "Version", "Backend", "Platforms"},
	}
	for name, spec := range tools {
		tableData = append(tableData, []string{
			name,
			spec.Version,
			spec.BackendName,
			strings.Join(platforms, ", "),
		})
	}
	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}

// parseLockToolArg parses a CLI tool argument:
//
//	"cli/cli"              → tool=cli/cli, version="", backend=""
//	"cli/cli@2.72.0"      → tool=cli/cli, version=2.72.0, backend=""
//	"github:cli/cli"      → tool=cli/cli, version="", backend=github
//	"github:cli/cli@2.72" → tool=cli/cli, version=2.72, backend=github
func parseLockToolArg(arg string) (tool, version, backendName string) {
	if idx := strings.Index(arg, ":"); idx != -1 {
		backendName = arg[:idx]
		arg = arg[idx+1:]
	}
	if idx := strings.LastIndex(arg, "@"); idx != -1 {
		version = arg[idx+1:]
		arg = arg[:idx]
	}
	tool = arg
	return
}
