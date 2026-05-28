// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
)

var (
	watchGlobs    []string
	watchIgnores  []string
	watchClear    bool
	watchDebounce int
	watchShell    bool

	runningTaskCmd *exec.Cmd
	cmdMutex       sync.Mutex
)

// watchCmd represents the watch command which watches files and runs a task on changes.
var watchCmd = &cobra.Command{
	Use:     "watch <task>",
	Short:   "Watch files and run task on changes",
	Long:    `Watch files in the current directory and automatically run the specified task when changes occur.`,
	Aliases: []string{"w"},
	Args:    cobra.ExactArgs(1),
	RunE:    runWatch,
}

func init() {
	watchCmd.Flags().StringSliceVarP(&watchGlobs, "glob", "g", []string{}, "Watch only files matching these glob patterns (e.g. *.go)")
	watchCmd.Flags().StringSliceVarP(&watchIgnores, "ignore", "i", []string{}, "Ignore files matching these patterns")
	watchCmd.Flags().BoolVar(&watchClear, "clear", false, "Clear the screen before running the task")
	watchCmd.Flags().IntVarP(&watchDebounce, "debounce", "d", 500, "Debounce delay in milliseconds")
	watchCmd.Flags().BoolVarP(&watchShell, "shell", "s", false, "Run task inside a shell")

	if rootCmd != nil {
		rootCmd.AddCommand(watchCmd)
	}
}

func runWatch(cmd *cobra.Command, args []string) error {
	taskName := args[0]
	debounceDuration := time.Duration(watchDebounce) * time.Millisecond

	pterm.DefaultSection.Println("Watcher Settings")

	output.Infof("Task to execute:   %s", pterm.LightCyan(taskName))
	output.Infof("Debounce delay:    %s", pterm.LightYellow(debounceDuration.String()))
	if len(watchGlobs) > 0 {
		output.Infof("Watch globs:       %s", pterm.LightMagenta(strings.Join(watchGlobs, ", ")))
	} else {
		output.Infof("Watch directories: %s (recursive)\n", pterm.LightGreen("."))
	}
	if len(watchIgnores) > 0 {
		output.Infof("Ignore patterns:   %s", pterm.LightRed(strings.Join(watchIgnores, ", ")))
	}
	if watchClear {
		output.Info("Screen clearing:   Enabled")
	}
	if watchShell {
		output.Info("Shell execution:   Enabled")
	}

	pterm.Println(pterm.FgGray.Sprint(strings.Repeat("─", 60)))
	output.Info("Watching for file changes. Press Ctrl+C to stop.")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	// Add directories recursively
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		// Skip hidden directories and common vendor dirs
		if path != "." && (strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor") {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})

	if err != nil {
		return fmt.Errorf("scan directories: %w", err)
	}

	// Run task immediately once at startup
	if watchClear {
		clearScreen()
	}
	runWatchTask(taskName)

	// Debounce logic
	var timer *time.Timer
	var timerMutex sync.Mutex

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// Ignore chmod events as they are usually metadata adjustments
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}

			// Apply glob & ignore filtering
			if !isMatched(event.Name, watchGlobs, watchIgnores) {
				continue
			}

			timerMutex.Lock()
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounceDuration, func() {
				// Clear screen if requested before starting the task
				if watchClear {
					clearScreen()
				}
				pterm.Println(pterm.FgGray.Sprint(strings.Repeat("─", 60)))
				output.Infof("Change detected in: %s", pterm.LightYellow(event.Name))
				output.Infof("Restarting task %s...", pterm.LightCyan(taskName))

				// Kill currently running task if active to support hot-reloading (Surpassing mise!)
				killCurrentCmd()

				runWatchTask(taskName)
			})
			timerMutex.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			output.Errorf("Watch error: %v", err)
		}
	}
}

// runWatchTask executes the task while properly managing the current command pointer.
func runWatchTask(taskName string) {
	exe, err := os.Executable()
	if err != nil {
		output.Errorf("Failed to find executable: %v", err)
		return
	}

	var cmd *exec.Cmd
	if watchShell {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		cmd = exec.Command(shell, "-c", fmt.Sprintf("%q run %q", exe, taskName))
	} else {
		cmd = exec.Command(exe, "run", taskName)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmdMutex.Lock()
	runningTaskCmd = cmd
	cmdMutex.Unlock()

	startTime := time.Now()
	err = cmd.Run()
	duration := time.Since(startTime)

	cmdMutex.Lock()
	// Only set to nil if it is still the same running command (prevents overwriting newer runs)
	if runningTaskCmd == cmd {
		runningTaskCmd = nil
	}
	cmdMutex.Unlock()

	if err != nil {
		// If command was killed (exit status -1/killed), print a friendly reload indicator
		if strings.Contains(err.Error(), "killed") || strings.Contains(err.Error(), "exit status -1") || strings.Contains(err.Error(), "signal: killed") {
			output.Warningf("🔄 Task %s interrupted & reloaded.", taskName)
		} else {
			output.Warningf("Task %s failed: %v (took %v)\n", taskName, err, duration.Round(time.Millisecond))
		}
	} else {
		output.Successf("✅ Task %s completed successfully in %v.", taskName, duration.Round(time.Millisecond))
	}
}

// killCurrentCmd terminates the currently running task if any (Surpassing mise!).
func killCurrentCmd() {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()
	if runningTaskCmd != nil && runningTaskCmd.Process != nil {
		output.Info("Killing running task execution to restart...")
		_ = runningTaskCmd.Process.Kill()
	}
}

// clearScreen clears the terminal screen cleanly across platforms.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// isMatched performs robust glob and ignore checking.
func isMatched(path string, globs, ignores []string) bool {
	base := filepath.Base(path)

	// 1. Check ignore list
	for _, ignore := range ignores {
		if matched, _ := filepath.Match(ignore, base); matched {
			return false
		}
		if matched, _ := filepath.Match(ignore, path); matched {
			return false
		}
		if strings.Contains(path, ignore) {
			return false
		}
	}

	// 2. If no globs are specified, everything is accepted
	if len(globs) == 0 {
		return true
	}

	// 3. Check glob list
	for _, glob := range globs {
		if matched, _ := filepath.Match(glob, base); matched {
			return true
		}
		if matched, _ := filepath.Match(glob, path); matched {
			return true
		}
		if strings.HasSuffix(glob, "/*") {
			prefix := glob[:len(glob)-2]
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}
