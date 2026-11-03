// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(syncCmd)
	}
}

// syncCmd represents the sync command which synchronizes tools from other version managers.
var syncCmd = &cobra.Command{
	Use:   "sync [tool]",
	Short: "Synchronize tools from other version managers with UniRTM",
	Long: `Synchronize tools from other version managers with UniRTM.

Autodetects existing versions installed via nvm, pyenv, rbenv, or asdf,
symlinks them into UniRTM's installs directory, and automatically registers
them in UniRTM's database with generated shims.

Examples:
  # Scan and sync all detected tools (node, python, ruby)
  unirtm sync

  # Scan and sync node versions only
  unirtm sync node

  # Scan and sync python versions only
  unirtm sync python

  # Scan and sync ruby versions only
  unirtm sync ruby`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSync,
}

type DetectedVersion struct {
	Tool         string
	Version      string
	Source       string // nvm, pyenv, rbenv, asdf
	ExternalPath string
}

// runSync executes the sync command.
func runSync(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	targetTool := ""
	if len(args) > 0 {
		targetTool = strings.ToLower(strings.TrimSpace(args[0]))
		if targetTool != "node" && targetTool != "python" && targetTool != "ruby" {
			return fmt.Errorf("unsupported tool: %q. Supported tools are: node, python, ruby", targetTool)
		}
	}

	pterm.Println()
	spinner, _ := output.StartSpinner("Scanning for external tool installations...")

	detected := scanExternalTools(targetTool)
	spinner.Stop()

	if len(detected) == 0 {
		formatter.Info("No external tool versions detected.", nil)
		return nil
	}

	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to open database: %v", err))
		return err
	}
	defer db.Close()

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to create repository: %v", err))
		return err
	}

	// Load existing installations to skip duplicates
	existingList, err := repo.List(ctx)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to list installations: %v", err))
		return err
	}

	existingMap := make(map[string]bool)
	for _, inst := range existingList {
		key := fmt.Sprintf("%s:%s", inst.Tool, inst.Version)
		existingMap[key] = true
	}

	// Prepare pterm table for visual output
	tableData := pterm.TableData{
		{"TOOL", "VERSION", "SOURCE", "STATUS"},
	}

	var toSync []DetectedVersion

	for _, d := range detected {
		key := fmt.Sprintf("%s:%s", d.Tool, d.Version)
		statusStr := ""
		if existingMap[key] {
			statusStr = pterm.FgGray.Sprint("Already synchronized ✓")
		} else {
			statusStr = pterm.FgYellow.Sprint("Ready to sync")
			toSync = append(toSync, d)
		}

		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(d.Tool),
			pterm.FgGreen.Sprint(d.Version),
			pterm.FgMagenta.Sprint(d.Source),
			statusStr,
		})
	}

	pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("   ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
		WithData(tableData).
		Render()

	if len(toSync) == 0 {
		pterm.Println()
		output.Success("All detected versions are already synchronized!")
		return nil
	}

	pterm.Println()
	output.Infof("Synchronizing %d new versions...", len(toSync))
	pterm.Println()

	installsDir := env.GetInstallsDir()
	shimsDir := env.GetShimsDir()
	dataDir := env.GetDataDir()
	generator := service.NewGenerator(shimsDir, dataDir+"/installs")
	providerRegistry := provider.NewRegistry()

	for _, d := range toSync {
		syncSpinner, _ := output.StartSpinner(fmt.Sprintf("Syncing %s@%s from %s...", d.Tool, d.Version, d.Source))

		targetDir := filepath.Join(installsDir, d.Tool, d.Version)

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
			syncSpinner.Fail(fmt.Sprintf("Failed to create installs dir: %v", err))
			continue
		}

		// Remove existing target path if it exists (e.g. dead symlink or duplicate folder)
		if _, err := os.Lstat(targetDir); err == nil {
			_ = os.Remove(targetDir)
		}

		// Create symlink
		if err := os.Symlink(d.ExternalPath, targetDir); err != nil {
			syncSpinner.Fail(fmt.Sprintf("Failed to symlink %s → %s: %v", d.ExternalPath, targetDir, err))
			continue
		}

		// Register in SQLite database
		inst := &repository.Installation{
			Tool:        d.Tool,
			Version:     d.Version,
			Backend:     d.Source,
			Provider:    "custom",
			InstallPath: targetDir,
			InstalledAt: time.Now(),
		}

		if err := repo.Create(ctx, inst); err != nil {
			syncSpinner.Fail(fmt.Sprintf("Failed to record installation: %v", err))
			_ = os.Remove(targetDir) // Clean up symlink on failure
			continue
		}

		// Automatically generate shims
		p := providerRegistry.GetWithBackend(d.Tool, d.Source)
		executables, err := p.ListExecutables(d.Tool, targetDir, d.Version)
		if err != nil || len(executables) == 0 {
			executables = []string{d.Tool}
		}

		if err := generator.GenerateShim(ctx, d.Tool, executables...); err != nil {
			syncSpinner.Warning(fmt.Sprintf("Registered %s@%s but failed to generate shims: %v", d.Tool, d.Version, err))
		} else {
			syncSpinner.Success(fmt.Sprintf("Successfully synced %s@%s!", pterm.FgCyan.Sprint(d.Tool), pterm.FgGreen.Sprint(d.Version)))
		}
	}

	return nil
}

func scanExternalTools(targetTool string) []DetectedVersion {
	var detected []DetectedVersion

	home, err := os.UserHomeDir()
	if err != nil {
		return detected
	}

	// 1. Scan Node (nvm, asdf)
	if targetTool == "" || targetTool == "node" {
		// NVM
		nvmDir := ""
		if runtime.GOOS == "windows" {
			appData := os.Getenv("APPDATA")
			if appData != "" {
				nvmDir = filepath.Join(appData, "nvm")
			}
		} else {
			nvmDir = filepath.Join(home, ".nvm", "versions", "node")
		}

		if nvmDir != "" {
			if entries, err := os.ReadDir(nvmDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						name := entry.Name()
						version := strings.TrimPrefix(name, "v")
						extPath := filepath.Join(nvmDir, name)

						binName := "node"
						if runtime.GOOS == "windows" {
							binName = "node.exe"
						}

						binPath := filepath.Join(extPath, "bin", binName)
						if runtime.GOOS == "windows" {
							binPath = filepath.Join(extPath, binName)
						}

						if _, err := os.Stat(binPath); err == nil {
							detected = append(detected, DetectedVersion{
								Tool:         "node",
								Version:      version,
								Source:       "nvm",
								ExternalPath: extPath,
							})
						}
					}
				}
			}
		}

		// ASDF Node
		asdfNodeDir := filepath.Join(home, ".asdf", "installs", "nodejs")
		if entries, err := os.ReadDir(asdfNodeDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					version := entry.Name()
					extPath := filepath.Join(asdfNodeDir, version)
					binPath := filepath.Join(extPath, "bin", "node")
					if runtime.GOOS == "windows" {
						binPath = filepath.Join(extPath, "node.exe")
					}

					if _, err := os.Stat(binPath); err == nil {
						detected = append(detected, DetectedVersion{
							Tool:         "node",
							Version:      version,
							Source:       "asdf",
							ExternalPath: extPath,
						})
					}
				}
			}
		}
	}

	// 2. Scan Python (pyenv, asdf)
	if targetTool == "" || targetTool == "python" {
		// Pyenv
		pyenvDir := ""
		if runtime.GOOS == "windows" {
			pyenvDir = filepath.Join(home, ".pyenv", "pyenv-win", "versions")
		} else {
			pyenvDir = filepath.Join(home, ".pyenv", "versions")
		}

		if entries, err := os.ReadDir(pyenvDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					version := entry.Name()
					extPath := filepath.Join(pyenvDir, version)

					binName := "python"
					if runtime.GOOS == "windows" {
						binName = "python.exe"
					}

					binPath := filepath.Join(extPath, "bin", binName)
					if runtime.GOOS == "windows" {
						binPath = filepath.Join(extPath, binName)
					}

					if _, err := os.Stat(binPath); err == nil {
						detected = append(detected, DetectedVersion{
							Tool:         "python",
							Version:      version,
							Source:       "pyenv",
							ExternalPath: extPath,
						})
					}
				}
			}
		}

		// ASDF Python
		asdfPyDir := filepath.Join(home, ".asdf", "installs", "python")
		if entries, err := os.ReadDir(asdfPyDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					version := entry.Name()
					extPath := filepath.Join(asdfPyDir, version)
					binPath := filepath.Join(extPath, "bin", "python")
					if runtime.GOOS == "windows" {
						binPath = filepath.Join(extPath, "python.exe")
					}

					if _, err := os.Stat(binPath); err == nil {
						detected = append(detected, DetectedVersion{
							Tool:         "python",
							Version:      version,
							Source:       "asdf",
							ExternalPath: extPath,
						})
					}
				}
			}
		}
	}

	// 3. Scan Ruby (rbenv, asdf)
	if targetTool == "" || targetTool == "ruby" {
		// Rbenv
		rbenvDir := filepath.Join(home, ".rbenv", "versions")
		if entries, err := os.ReadDir(rbenvDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					version := entry.Name()
					extPath := filepath.Join(rbenvDir, version)

					binName := "ruby"
					if runtime.GOOS == "windows" {
						binName = "ruby.exe"
					}

					binPath := filepath.Join(extPath, "bin", binName)
					if runtime.GOOS == "windows" {
						binPath = filepath.Join(extPath, binName)
					}

					if _, err := os.Stat(binPath); err == nil {
						detected = append(detected, DetectedVersion{
							Tool:         "ruby",
							Version:      version,
							Source:       "rbenv",
							ExternalPath: extPath,
						})
					}
				}
			}
		}

		// ASDF Ruby
		asdfRubyDir := filepath.Join(home, ".asdf", "installs", "ruby")
		if entries, err := os.ReadDir(asdfRubyDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					version := entry.Name()
					extPath := filepath.Join(asdfRubyDir, version)
					binPath := filepath.Join(extPath, "bin", "ruby")
					if runtime.GOOS == "windows" {
						binPath = filepath.Join(extPath, "ruby.exe")
					}

					if _, err := os.Stat(binPath); err == nil {
						detected = append(detected, DetectedVersion{
							Tool:         "ruby",
							Version:      version,
							Source:       "asdf",
							ExternalPath: extPath,
						})
					}
				}
			}
		}
	}

	return detected
}
