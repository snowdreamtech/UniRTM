// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
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
	// Use a visited map for cycle detection in the dependency graph
	visited := make(map[string]bool)
	return r.runTaskWithGraph(ctx, dir, taskName, args, env, visited)
}

func (r *NativeRunner) runTaskWithGraph(ctx context.Context, dir string, taskName string, args []string, env []string, visited map[string]bool) error {
	if visited[taskName] {
		return fmt.Errorf("circular dependency detected involving task %q", taskName)
	}
	visited[taskName] = true
	defer func() { visited[taskName] = false }() // allow multiple paths to same task if needed, though DAG usually means we run it once.
	// Actually, for a proper DAG, we should only run a task once per session.
	// But for simplicity, we just do cycle detection.

	taskDef, exists := r.tasks[taskName]
	if !exists {
		return fmt.Errorf("task %q not found in UniRTM configuration", taskName)
	}

	// Recursively execute dependencies sequentially
	for _, dep := range taskDef.Depends {
		if err := r.runTaskWithGraph(ctx, dir, dep, nil, env, visited); err != nil {
			return fmt.Errorf("dependency %q failed: %w", dep, err)
		}
	}

	// Prepare the script. If there are args, append them directly.
	script := taskDef.Run
	if len(args) > 0 {
		if script != "" {
			script = script + " " + strings.Join(args, " ")
		} else {
			script = strings.Join(args, " ")
		}
	}

	// Resolve timeout: task override > global setting
	timeout := r.settings.TaskTimeout
	if taskDef.Timeout > 0 {
		timeout = config.DurationOrInt(taskDef.Timeout)
	}

	runCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	// Inject process env + UniRTM env
	fullEnv := append(os.Environ(), env...)
	// Inject task-specific env defined in TOML
	for k, v := range taskDef.Env {
		fullEnv = append(fullEnv, fmt.Sprintf("%s=%s", k, v))
	}

	var file *syntax.File
	var err error
	if strings.TrimSpace(script) != "" {
		parser := syntax.NewParser()
		file, err = parser.Parse(strings.NewReader(script), "")
		if err != nil {
			return fmt.Errorf("failed to parse task script: %w", err)
		}
	}

	// Bind IO streams based on output style
	outputStyle := r.settings.TaskOutput
	if taskDef.Output != "" {
		outputStyle = taskDef.Output
	}

	// Scheme 4: Check environment variable override (UNIRTM_TASK_OUTPUT)
	if envOutput := os.Getenv("UNIRTM_TASK_OUTPUT"); envOutput != "" {
		outputStyle = envOutput
	}

	// Scheme 5: Auto-detect CI environment and use interleaved mode
	if outputStyle == "spinner" || outputStyle == "" {
		isCIEnv := os.Getenv("CI") != "" ||
			os.Getenv("GITHUB_ACTIONS") != "" ||
			os.Getenv("GITLAB_CI") != "" ||
			os.Getenv("CIRCLECI") != "" ||
			os.Getenv("TRAVIS") != "" ||
			os.Getenv("JENKINS_URL") != "" ||
			os.Getenv("BUILDKITE") != "" ||
			os.Getenv("DRONE") != ""
		if isCIEnv {
			outputStyle = "interleaved"
		}
	}

	var spinner *pterm.SpinnerPrinter
	var buf bytes.Buffer
	var stdout, stderr io.Writer
	stdout = os.Stdout
	stderr = os.Stderr

	if outputStyle == "spinner" || outputStyle == "" {
		// Create a local copy to avoid data races when tasks run concurrently
		spinner, _ = output.StartSpinner(fmt.Sprintf("Running task: %s", taskName))
		// Capture output so we can show it if it fails, or just hide it
		if file != nil {
			stdout = &buf
			stderr = &buf
		}
	} else if outputStyle == "prefix" {
		prefix := fmt.Sprintf("[%s] ", pterm.FgCyan.Sprint(taskName))
		if file != nil {
			stdout = &prefixWriter{w: os.Stdout, prefix: prefix, atStart: true}
			stderr = &prefixWriter{w: os.Stderr, prefix: prefix, atStart: true}
		}
	} else {
		// "interleaved" or other
		output.Infof("Running task: %s", taskName)
	}

	if file != nil {
		runner, runnerErr := interp.New(
			interp.Env(expand.ListEnviron(fullEnv...)),
			interp.Dir(dir),
			interp.StdIO(os.Stdin, stdout, stderr),
			interp.Params("-e"),
		)
		if runnerErr != nil {
			err = fmt.Errorf("failed to create shell runner: %w", runnerErr)
		} else {
			err = runner.Run(runCtx, file)
		}
	}

	if spinner != nil {
		if err != nil {
			spinner.Fail(fmt.Sprintf("Task %s failed: %v", taskName, err))
			if buf.Len() > 0 {
				fmt.Fprintln(os.Stderr, buf.String())
			}
		} else {
			spinner.Success(fmt.Sprintf("Task %s completed", taskName))
			// Scheme 2: Show success output if captured
			if buf.Len() > 0 {
				fmt.Println(buf.String())
			}
		}
	} else {
		if err != nil {
			output.Errorf("Task %s failed: %v", taskName, err)
		} else {
			output.Successf("Task %s completed", taskName)
		}
	}

	return err
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
