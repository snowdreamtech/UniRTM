// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/config"
)

// NativeRunner is the fallback task runner that executes tasks defined
// directly in the UniRTM configuration file (e.g., [tasks] block).
type NativeRunner struct {
	tasks    map[string]config.Task
	settings config.Settings
}

// NewNativeRunner creates a new NativeRunner with the parsed configuration tasks and settings.
func NewNativeRunner(tasks map[string]config.Task, settings config.Settings) *NativeRunner {
	return &NativeRunner{tasks: tasks, settings: settings}
}

// Name returns the name of this runner.
func (r *NativeRunner) Name() string {
	return "native"
}

// CanExecute returns true if the task is defined in the UniRTM configuration.
func (r *NativeRunner) CanExecute(dir string, taskName string) bool {
	_, exists := r.tasks[taskName]
	return exists
}

// ListTasks returns all tasks defined in the configuration.
func (r *NativeRunner) ListTasks(dir string) ([]string, error) {
	tasks := make([]string, 0, len(r.tasks))
	for name := range r.tasks {
		tasks = append(tasks, name)
	}
	return tasks, nil
}

// Run executes a task defined in the unirtm.toml configuration.
func (r *NativeRunner) Run(ctx context.Context, dir string, taskName string, args []string, env []string) error {
	taskDef, exists := r.tasks[taskName]
	if !exists {
		return fmt.Errorf("task %q not found in UniRTM configuration", taskName)
	}

	// Recursively execute dependencies sequentially (MVP)
	for _, dep := range taskDef.Depends {
		if err := r.Run(ctx, dir, dep, nil, env); err != nil {
			return fmt.Errorf("dependency %q failed: %w", dep, err)
		}
	}

	// Prepare the script. If there are args, append them directly.
	script := taskDef.Run
	if len(args) > 0 {
		script = script + " " + strings.Join(args, " ")
	}

	// Apply shell-style expansion to the script
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	// Overlay UniRTM resolved env
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	script = os.Expand(script, func(k string) string {
		return envMap[k]
	})

	// Resolve timeout: task override > global setting
	timeout := r.settings.TaskTimeout
	if taskDef.Timeout > 0 {
		timeout = taskDef.Timeout
	}

	runCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(runCtx, "sh", "-c", script)
	cmd.Dir = dir
	
	// Inject process env + UniRTM env
	cmd.Env = append(os.Environ(), env...)
	
	// Inject task-specific env defined in TOML
	for k, v := range taskDef.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Bind IO streams based on output style
	outputStyle := r.settings.TaskOutput
	if taskDef.Output != "" {
		outputStyle = taskDef.Output
	}

	if outputStyle == "prefix" {
		prefix := fmt.Sprintf("[%s] ", taskName)
		cmd.Stdout = &prefixWriter{w: os.Stdout, prefix: prefix, atStart: true}
		cmd.Stderr = &prefixWriter{w: os.Stderr, prefix: prefix, atStart: true}
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

type prefixWriter struct {
	w       io.Writer
	prefix  string
	atStart bool
}

func (pw *prefixWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && len(line) == 0 {
			break
		}
		if pw.atStart || i > 0 {
			_, err = fmt.Fprint(pw.w, pw.prefix)
			if err != nil {
				return n, err
			}
		}
		_, err = fmt.Fprint(pw.w, line)
		if err != nil {
			return n, err
		}
		if i < len(lines)-1 {
			_, err = fmt.Fprint(pw.w, "\n")
			if err != nil {
				return n, err
			}
		}
	}
	pw.atStart = strings.HasSuffix(string(p), "\n")
	return len(p), nil
}
