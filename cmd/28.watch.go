// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch <task>",
	Short: "Watch files and run task on changes",
	Long:  `Watch files in the current directory and automatically run the specified task when changes occur.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]
		
		// Run task immediately once
		runWatchTask(taskName)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			pterm.Error.Printf("Failed to create watcher: %v\n", err)
			os.Exit(1)
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
			// Skip hidden directories and common vendor dirs
			base := filepath.Base(path)
			// Ensure we don't skip the current directory "."
			if path != "." && (strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor") {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		})
		
		if err != nil {
			pterm.Error.Printf("Failed to scan directories: %v\n", err)
			os.Exit(1)
		}

		pterm.Info.Printf("Watching for file changes. Press Ctrl+C to stop.\n")

		// Debounce logic
		var timer *time.Timer
		debounceDuration := 500 * time.Millisecond

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Ignore chmod events
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}

				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounceDuration, func() {
					pterm.Info.Printf("File changed: %s. Restarting task %s...\n", event.Name, taskName)
					runWatchTask(taskName)
				})

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				pterm.Error.Printf("Watch error: %v\n", err)
			}
		}
	},
}

func runWatchTask(taskName string) {
	// Call unirtm run using os.Executable to ensure proper isolation and prevent os.Exit from terminating the watcher
	exe, err := os.Executable()
	if err != nil {
		pterm.Error.Printf("Failed to find executable: %v\n", err)
		return
	}

	cmd := exec.Command(exe, "run", taskName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		pterm.Warning.Printf("Task %s failed: %v\n", taskName, err)
	} else {
		pterm.Success.Printf("Task %s completed.\n", taskName)
	}
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
