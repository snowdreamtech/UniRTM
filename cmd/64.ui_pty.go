// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"golang.org/x/net/websocket"
)

// TerminalControlMessage defines the JSON protocol exchanged between xterm.js and the backend.
type TerminalControlMessage struct {
	Type string `json:"type"`           // "input", "resize", "kill", "ping"
	Data string `json:"data,omitempty"` // keystrokes or terminal output
	Cols uint16 `json:"cols,omitempty"` // terminal columns
	Rows uint16 `json:"rows,omitempty"` // terminal rows
}

// buildTerminalEnv prepares a complete, enriched environment for the PTY session.
// It ensures that UniRTM's shims and binary directories take precedence in PATH,
// and instructs tools to output rich 256/truecolor ANSI sequences.
func buildTerminalEnv() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}

	unirtmBin := filepath.Join(home, ".unirtm", "bin")
	unirtmShims := filepath.Join(home, ".unirtm", "shims")

	currentPath := os.Getenv("PATH")
	enhancedPath := fmt.Sprintf("%s%c%s%c%s", unirtmShims, os.PathListSeparator, unirtmBin, os.PathListSeparator, currentPath)

	envMap := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	envMap["PATH"] = enhancedPath
	envMap["TERM"] = "xterm-256color"
	envMap["COLORTERM"] = "truecolor"
	envMap["UNIRTM_TERMINAL"] = "1"
	if _, ok := envMap["LANG"]; !ok {
		envMap["LANG"] = "en_US.UTF-8"
	}

	var envList []string
	for k, v := range envMap {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}
	return envList
}

// handleTerminalWebSocket manages the full-duplex PTY streaming connection.
func handleTerminalWebSocket(ws *websocket.Conn) {
	defer ws.Close()

	query := ws.Request().URL.Query()
	shellPath := strings.TrimSpace(query.Get("shell"))
	cwd := strings.TrimSpace(query.Get("cwd"))

	cols := uint16(80)
	rows := uint16(24)
	if c, err := strconv.ParseUint(query.Get("cols"), 10, 16); err == nil && c > 0 {
		cols = uint16(c)
	}
	if r, err := strconv.ParseUint(query.Get("rows"), 10, 16); err == nil && r > 0 {
		rows = uint16(r)
	}

	currentCols := cols
	currentRows := rows

	session, err := startPTYSession(shellPath, cwd, cols, rows)
	if err != nil {
		pterm.Error.Printf("Failed to spawn PTY session: %v\n", err)
		errMsg, _ := json.Marshal(TerminalControlMessage{
			Type: "output",
			Data: fmt.Sprintf("\r\n\x1b[31mFailed to start terminal: %v\x1b[0m\r\n", err),
		})
		_, _ = ws.Write(errMsg)
		return
	}
	defer session.Close()

	// Goroutine 1: PTY stdout -> WebSocket
	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		buf := make([]byte, 4096)
		for {
			n, readErr := session.Read(buf)
			if n > 0 {
				data := string(buf[:n])
				// Send as JSON message to keep consistent framing
				payload, _ := json.Marshal(TerminalControlMessage{
					Type: "output",
					Data: data,
				})
				if _, writeErr := ws.Write(payload); writeErr != nil {
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()

	// Goroutine 2: WebSocket -> PTY stdin & control signals
	buf := make([]byte, 4096)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			break
		}
		if n == 0 {
			continue
		}

		raw := buf[:n]
		var ctrl TerminalControlMessage
		if jsonErr := json.Unmarshal(raw, &ctrl); jsonErr == nil && ctrl.Type != "" {
			switch ctrl.Type {
			case "input":
				if len(ctrl.Data) > 0 {
					_, _ = session.Write([]byte(ctrl.Data))
				}
			case "resize":
				if ctrl.Cols > 0 && ctrl.Rows > 0 && (ctrl.Cols != currentCols || ctrl.Rows != currentRows) {
					currentCols = ctrl.Cols
					currentRows = ctrl.Rows
					session.Resize(ctrl.Cols, ctrl.Rows)
				}
			case "kill":
				return
			case "ping":
				pong, _ := json.Marshal(TerminalControlMessage{Type: "pong"})
				_, _ = ws.Write(pong)
			}
		} else {
			// Fallback: raw keystroke bytes
			_, _ = session.Write(raw)
		}
	}

	<-outputDone
}

// PTYSession encapsulates an active PTY process and its control hooks.
type PTYSession interface {
	io.ReadWriteCloser
	Resize(cols, rows uint16)
}
