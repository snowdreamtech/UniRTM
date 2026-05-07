package service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/service"
)

// ExampleAutoActivationManager_HandleDirectoryChange demonstrates how to use the
// Auto-Activation Manager to handle directory changes and automatically activate
// project toolchains.
func ExampleAutoActivationManager_HandleDirectoryChange() {
	// Create temporary project directory for demonstration
	tmpDir, _ := os.MkdirTemp("", "unirtm-example-*")
	defer os.RemoveAll(tmpDir)

	projectDir := filepath.Join(tmpDir, "myproject")
	_ = os.MkdirAll(projectDir, 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "unirtm.toml"), []byte("# project config"), 0644)

	// Create managers
	activationMgr := service.NewActivationManager("/tmp/shims", "/tmp/data")
	autoMgr := service.NewAutoActivationManager(activationMgr)

	// Initial state (no project active)
	currentState := &service.EnvironmentState{
		ProjectDir:   "",
		ToolVersions: make(map[string]string),
		EnvVars:      make(map[string]string),
		PreviousPath: "/usr/bin:/bin",
	}

	// Simulate entering the project directory
	event := service.DirectoryChangeEvent{
		OldDir: tmpDir,
		NewDir: projectDir,
		Shell:  service.ShellBash,
	}

	ctx := context.Background()
	change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Action: %s\n", change.Action)
	fmt.Printf("Project sub-directory: %s\n", filepath.Base(change.NewState.ProjectDir))
	fmt.Printf("Script generated: %t\n", len(change.Script) > 0)

	// Output:
	// Action: activate
	// Project sub-directory: myproject
	// Script generated: true
}

// ExampleAutoActivationManager_GenerateHookEnvScript demonstrates how to generate
// a shell hook script for automatic activation.
func ExampleAutoActivationManager_GenerateHookEnvScript() {
	activationMgr := service.NewActivationManager("/tmp/shims", "/tmp/data")
	autoMgr := service.NewAutoActivationManager(activationMgr)

	// Generate hook script for bash
	hookScript, err := autoMgr.GenerateHookEnvScript(service.ShellBash)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// The hook script can be added to ~/.bashrc
	fmt.Println("# Add this to your ~/.bashrc:")
	fmt.Println(hookScript)

	// Output will include:
	// # Add this to your ~/.bashrc:
	// # UniRTM auto-activation hook
	// _unirtm_hook() {
	//   ...
	// }
}

// ExampleAutoActivationManager_projectSwitch demonstrates switching between projects.
func ExampleAutoActivationManager_projectSwitch() {
	// Create temporary project directories
	tmpDir, _ := os.MkdirTemp("", "unirtm-example-*")
	defer os.RemoveAll(tmpDir)

	project1 := filepath.Join(tmpDir, "project1")
	project2 := filepath.Join(tmpDir, "project2")
	_ = os.MkdirAll(project1, 0755)
	_ = os.MkdirAll(project2, 0755)
	_ = os.WriteFile(filepath.Join(project1, "unirtm.toml"), []byte("# project1"), 0644)
	_ = os.WriteFile(filepath.Join(project2, "unirtm.toml"), []byte("# project2"), 0644)

	// Create managers
	activationMgr := service.NewActivationManager("/tmp/shims", "/tmp/data")
	autoMgr := service.NewAutoActivationManager(activationMgr)

	// Start in project1
	currentState := &service.EnvironmentState{
		ProjectDir: project1,
		ToolVersions: map[string]string{
			"node": "18.0.0",
		},
		EnvVars:      make(map[string]string),
		PreviousPath: "/usr/bin:/bin",
	}

	// Switch to project2
	event := service.DirectoryChangeEvent{
		OldDir: project1,
		NewDir: project2,
		Shell:  service.ShellBash,
	}

	ctx := context.Background()
	change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Action: %s\n", change.Action)
	fmt.Printf("Old project: %s\n", filepath.Base(currentState.ProjectDir))
	fmt.Printf("New project: %s\n", filepath.Base(change.NewState.ProjectDir))

	// Output:
	// Action: switch
	// Old project: project1
	// New project: project2
}

// ExampleAutoActivationManager_deactivation demonstrates leaving a project.
func ExampleAutoActivationManager_deactivation() {
	// Create temporary project directory
	tmpDir, _ := os.MkdirTemp("", "unirtm-example-*")
	defer os.RemoveAll(tmpDir)

	projectDir := filepath.Join(tmpDir, "myproject")
	_ = os.MkdirAll(projectDir, 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "unirtm.toml"), []byte("# project"), 0644)

	// Create managers
	activationMgr := service.NewActivationManager("/tmp/shims", "/tmp/data")
	autoMgr := service.NewAutoActivationManager(activationMgr)

	// Start in project
	currentState := &service.EnvironmentState{
		ProjectDir: projectDir,
		ToolVersions: map[string]string{
			"python": "3.11.0",
		},
		EnvVars:      make(map[string]string),
		PreviousPath: "/usr/bin:/bin",
	}

	// Leave project
	event := service.DirectoryChangeEvent{
		OldDir: projectDir,
		NewDir: tmpDir,
		Shell:  service.ShellBash,
	}

	ctx := context.Background()
	change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Action: %s\n", change.Action)
	fmt.Printf("Project deactivated: %t\n", change.NewState.ProjectDir == "")
	fmt.Printf("Tools cleared: %t\n", len(change.NewState.ToolVersions) == 0)

	// Output:
	// Action: deactivate
	// Project deactivated: true
	// Tools cleared: true
}
