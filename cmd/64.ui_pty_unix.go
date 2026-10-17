// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
)

type unixPTYSession struct {
	cmd  *exec.Cmd
	ptmx *os.File
}

func (s *unixPTYSession) Read(p []byte) (int, error) {
	return s.ptmx.Read(p)
}

func (s *unixPTYSession) Write(p []byte) (int, error) {
	return s.ptmx.Write(p)
}

func (s *unixPTYSession) Close() error {
	_ = s.ptmx.Close()
	if s.cmd.Process != nil {
		// Send SIGHUP first, then kill if not exited
		_ = s.cmd.Process.Signal(syscall.SIGHUP)
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}
	return nil
}

func (s *unixPTYSession) Resize(cols, rows uint16) {
	if s.ptmx != nil {
		_ = pty.Setsize(s.ptmx, &pty.Winsize{
			Rows: rows,
			Cols: cols,
		})
	}
}

// startPTYSession creates a native POSIX pseudo-terminal and launches the interactive shell.
func startPTYSession(shellPath, cwd string, cols, rows uint16) (PTYSession, error) {
	if shellPath == "" {
		if bashPath, err := exec.LookPath("bash"); err == nil {
			shellPath = bashPath
		} else if shPath, err := exec.LookPath("sh"); err == nil {
			shellPath = shPath
		} else {
			shellPath = "/bin/sh"
		}
	}

	// Launch as interactive login shell to load prompt and environment
	cmd := exec.Command(shellPath, "-l")
	cmd.Env = buildTerminalEnv()

	if cwd != "" {
		if fi, err := os.Stat(cwd); err == nil && fi.IsDir() {
			cmd.Dir = cwd
		}
	}
	if cmd.Dir == "" {
		if wd, err := os.Getwd(); err == nil {
			cmd.Dir = wd
		}
	}

	if cols == 0 {
		cols = 80
	}
	if rows == 0 {
		rows = 24
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
	if err != nil {
		// Fallback without -l if login flag fails on certain constrained environments
		cmd = exec.Command(shellPath)
		cmd.Env = buildTerminalEnv()
		cmd.Dir = cwd
		ptmx, err = pty.StartWithSize(cmd, &pty.Winsize{
			Rows: rows,
			Cols: cols,
		})
		if err != nil {
			return nil, fmt.Errorf("pty spawn failed: %w", err)
		}
	}

	return &unixPTYSession{
		cmd:  cmd,
		ptmx: ptmx,
	}, nil
}
