// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"golang.org/x/sync/errgroup"
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
	// Use a visited map (protected by a mutex) for cycle detection across goroutines.
	visited := &sync.Map{}
	return r.runTaskWithGraph(ctx, dir, taskName, args, env, visited)
}

// maxJobs returns the effective concurrency ceiling for parallel dependency execution.
func (r *NativeRunner) maxJobs() int {
	if r.settings.Jobs > 0 {
		return r.settings.Jobs
	}
	return runtime.NumCPU()
}

// runTaskWithGraph executes a task and its dependencies via a parallel DAG walk.
func (r *NativeRunner) runTaskWithGraph(ctx context.Context, dir string, taskName string, args []string, env []string, visited *sync.Map) error {
	// Cycle detection: mark before recursing, clear after.
	if _, alreadyVisiting := visited.LoadOrStore(taskName, true); alreadyVisiting {
		return fmt.Errorf("circular dependency detected involving task %q", taskName)
	}
	defer visited.Delete(taskName)

	taskDef, exists := r.tasks[taskName]
	if !exists {
		return fmt.Errorf("task %q not found in UniRTM configuration", taskName)
	}

	// ── Parallel dependency execution ──────────────────────────────────────────
	if len(taskDef.Depends) > 0 {
		sem := make(chan struct{}, r.maxJobs())
		g, gCtx := errgroup.WithContext(ctx)
		for _, dep := range taskDef.Depends {
			dep := dep // capture
			sem <- struct{}{}
			g.Go(func() error {
				defer func() { <-sem }()
				return r.runTaskWithGraph(gCtx, dir, dep, nil, env, visited)
			})
		}
		if err := g.Wait(); err != nil {
			return fmt.Errorf("dependency failed: %w", err)
		}
	}

	// ── Prepare the script ─────────────────────────────────────────────────────
	script := taskDef.Run.Script()
	if len(args) > 0 {
		if script != "" {
			script = script + " " + strings.Join(args, " ")
		} else {
			script = strings.Join(args, " ")
		}
	}

	// No script body – task is purely a dependency aggregator.
	if strings.TrimSpace(script) == "" {
		return nil
	}

	// ── Native retry loop ──────────────────────────────────────────────────────
	maxAttempts := taskDef.Retry + 1 // Retry=0 → one attempt; Retry=N → N+1 attempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			output.Infof("Retrying task %s (attempt %d/%d)…", taskName, attempt, maxAttempts)
		}
		lastErr = r.execScript(ctx, dir, taskName, script, taskDef, env)
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

// execScript compiles and runs the shell script for a single attempt.
func (r *NativeRunner) execScript(ctx context.Context, dir string, taskName string, script string, taskDef config.Task, env []string) error {
	// ── Timeout ────────────────────────────────────────────────────────────────
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

	// ── Environment ────────────────────────────────────────────────────────────
	fullEnv := append(os.Environ(), env...)
	for k, v := range taskDef.Env {
		fullEnv = append(fullEnv, fmt.Sprintf("%s=%s", k, v))
	}

	// ── Parse script ───────────────────────────────────────────────────────────
	parser := syntax.NewParser()
	file, err := parser.Parse(strings.NewReader(script), "")
	if err != nil {
		return fmt.Errorf("failed to parse task script: %w", err)
	}

	// ── Output mode ────────────────────────────────────────────────────────────
	outputStyle := r.settings.TaskOutput
	if taskDef.Output != "" {
		outputStyle = taskDef.Output
	}
	if envOutput := os.Getenv("UNIRTM_TASK_OUTPUT"); envOutput != "" {
		outputStyle = envOutput
	}
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
		spinner, _ = output.StartSpinner(fmt.Sprintf("Running task: %s", taskName))
		stdout = &buf
		stderr = &buf
	} else if outputStyle == "prefix" {
		prefix := fmt.Sprintf("[%s] ", pterm.FgCyan.Sprint(taskName))
		stdout = &prefixWriter{w: os.Stdout, prefix: prefix, atStart: true}
		stderr = &prefixWriter{w: os.Stderr, prefix: prefix, atStart: true}
	} else {
		// "interleaved" or other
		output.Infof("Running task: %s", taskName)
	}

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

	if spinner != nil {
		if err != nil {
			spinner.Fail(fmt.Sprintf("Task %s failed: %v", taskName, err))
			if buf.Len() > 0 {
				fmt.Fprintln(os.Stderr, buf.String())
			}
		} else {
			spinner.Success(fmt.Sprintf("Task %s completed", taskName))
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
