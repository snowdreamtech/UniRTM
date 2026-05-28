// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	// Add command to root
	if rootCmd != nil {
		rootCmd.AddCommand(testToolCmd)
	}
}

// testToolCmd represents the test-tool command which tests a tool installs and executes.
var testToolCmd = &cobra.Command{
	Use:   "test-tool [tool[@version]...]",
	Short: "Test a tool installs and executes",
	Long: `Test a tool installs and executes by downloading it and running a version/help check.

This command installs the specified tools (or all tools in the configuration if no arguments
are provided) and then attempts to execute their provided binaries to verify that they run
correctly without dynamic linking or architecture issues.

Examples:
  # Test all tools in the current unirtm.toml
  unirtm test-tool

  # Test GitHub CLI
  unirtm test-tool cli/cli

  # Test a specific version
  unirtm test-tool node@20.0.0
  unirtm test-tool node 20.0.0`,
	Args: cobra.ArbitraryArgs,
	RunE: runTestTool,
}

// runTestTool executes the test-tool command.
func runTestTool(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cfg, _ := config.Load()
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	if len(args) > 0 {
		formatter.Info("Testing requested tools...", map[string]interface{}{"args": args})
	} else {
		formatter.Info("Testing all tools from configuration...", nil)
	}

	// 1. Delegate to runInstall to ensure the tools are installed
	// runInstall will handle parsing, fallback to config, and actual installation
	err := runInstall(cmd, args)
	if err != nil {
		output.Errorf("Test failed: installation step failed: %v", err)
		return err
	}

	// 2. Parse exactly which tools we are meant to test (duplicating parsing logic from install.go)
	im, err := getInstallationManager(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get installation manager: %w", err)
	}

	var toolsToTest map[string]service.ToolSpec
	if len(args) == 0 {
		if cfg == nil || len(cfg.Tools) == 0 {
			formatter.Warning("No tools found in project configuration to test.")
			return nil
		}
		toolsToTest = make(map[string]service.ToolSpec, len(cfg.Tools))
		for name, tc := range cfg.Tools {
			backendName, toolName, version, _ := im.ParseToolSpec(name)
			if tc.Backend != "" {
				backendName = tc.Backend
			}
			if tc.Version != "" {
				version = tc.Version
			}
			toolsToTest[name] = service.ToolSpec{
				Name:        toolName,
				Version:     version,
				BackendName: backendName,
			}
		}
	} else {
		toolsToTest = make(map[string]service.ToolSpec)
		isLegacySingleTool := false
		if len(args) == 2 {
			if !strings.Contains(args[0], "@") && !strings.Contains(args[0], ":") && !strings.Contains(args[1], "@") {
				isLegacySingleTool = true
			}
		}

		if isLegacySingleTool {
			backendName, tool, _, _ := im.ParseToolSpec(args[0])
			version := args[1]
			if installBackend != "" {
				backendName = installBackend
			}
			toolsToTest[tool] = service.ToolSpec{
				Name:        tool,
				Version:     version,
				BackendName: backendName,
			}
		} else {
			for _, arg := range args {
				backendName, tool, version, explicitVersion := im.ParseToolSpec(arg)
				if installBackend != "" {
					backendName = installBackend
				}
				if !explicitVersion && cfg != nil {
					if tc, ok := cfg.Tools[tool]; ok && tc.Version != "" {
						version = tc.Version
						if tc.Backend != "" {
							backendName = tc.Backend
						}
					}
				}
				if version == "" {
					version = "latest"
				}
				key := tool
				if backendName != "" && backendName != "asdf" {
					key = backendName + ":" + tool
				}
				toolsToTest[key] = service.ToolSpec{
					Name:        tool,
					Version:     version,
					BackendName: backendName,
				}
			}
		}
	}

	// 3. Test each tool by finding and executing its binaries
	formatter.Info("\nExecuting tests for installed tools...", nil)

	hasError := false
	for _, spec := range toolsToTest {
		toolName := spec.Name
		version := spec.Version
		backendName := spec.BackendName

		fsName := env.GetFSToolName(toolName, backendName)

		// If the version is "latest", runInstall resolved it dynamically.
		// We can find the installed version by finding the newest directory.
		if version == "latest" {
			if entries, err := os.ReadDir(filepath.Join(env.GetInstallsDir(), fsName)); err == nil {
				var newestName string
				var newestTime int64
				for _, e := range entries {
					if e.IsDir() {
						if info, err := e.Info(); err == nil {
							if info.ModTime().UnixNano() > newestTime {
								newestTime = info.ModTime().UnixNano()
								newestName = e.Name()
							}
						}
					}
				}
				if newestName != "" {
					version = newestName
				}
			}
		}

		// Use IsInstalled to get the actual resolved version (e.g. adding 'v' prefix if needed)
		installed, resolvedInst := im.IsInstalled(ctx, toolName, version, backendName)
		if !installed {
			output.Errorf("Tool %s@%s is not properly installed, skipping execution test.", toolName, version)
			hasError = true
			continue
		}

		resolvedVersion := resolvedInst.Version

		p := provider.NewRegistry().GetWithBackend(toolName, backendName)
		if p == nil {
			output.Warningf("No provider found for %s, skipping execution test.", toolName)
			continue
		}

		installPath := filepath.Join(env.GetInstallsDir(), fsName, resolvedVersion)

		execs, err := p.ListExecutables(toolName, installPath, resolvedVersion)
		if err != nil || len(execs) == 0 {
			output.Warningf("No executables found for %s@%s to test.", toolName, resolvedVersion)
			continue
		}

		// Prepare environment for execution (e.g. JAVA_HOME)
		envVars, _ := p.GetEnvVars(toolName, installPath, resolvedVersion)
		cmdEnv := os.Environ()
		for k, v := range envVars {
			cmdEnv = append(cmdEnv, k+"="+v)
		}

		output.Infof("Testing %s@%s executables:", toolName, resolvedVersion)
		for _, exe := range execs {
			err := testExecutable(exe, cmdEnv)
			if err != nil {
				output.Errorf("  %s (failed: %v)", filepath.Base(exe), err)
				hasError = true
			} else {
				output.Successf("%s", filepath.Base(exe))
			}
		}
	}

	if hasError {
		return fmt.Errorf("one or more tools failed the execution test")
	}

	output.Successf("All tested tools executed successfully")
	return nil
}

// testExecutable attempts to run the given executable with standard "version" or "help" flags.
// Returns nil if the executable successfully runs and exits with status code 0.
func testExecutable(exePath string, cmdEnv []string) error {
	testFlags := [][]string{
		{"--version"},
		{"-V"},
		{"version"},
		{"--help"},
		{"-h"},
	}

	var lastErr error
	var lastOutput string

	for _, flags := range testFlags {
		cmd := exec.Command(exePath, flags...)
		cmd.Env = cmdEnv
		output, err := cmd.CombinedOutput()

		if err == nil {
			return nil // Success
		}

		lastErr = err
		lastOutput = strings.TrimSpace(string(output))
		// Truncate output if it's too long
		if len(lastOutput) > 100 {
			lastOutput = lastOutput[:97] + "..."
		}
	}

	if lastOutput != "" {
		return fmt.Errorf("%v (output: %s)", lastErr, lastOutput)
	}
	return lastErr
}
