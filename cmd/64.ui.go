// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/frontend"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	uiPort   int
	uiHost   string
	uiNoOpen bool
)

func init() {
	uiCmd.Flags().IntVarP(&uiPort, "port", "p", 5173, "port to listen on")
	uiCmd.Flags().StringVarP(&uiHost, "host", "H", "127.0.0.1", "host address to bind")
	uiCmd.Flags().BoolVar(&uiNoOpen, "no-open", false, "do not open browser automatically")

	if rootCmd != nil {
		rootCmd.AddCommand(uiCmd)
	}
}

// uiCmd starts the embedded interactive graphical workbench.
var uiCmd = &cobra.Command{
	Use:     "ui",
	Aliases: []string{"dashboard", "web"},
	Short:   "Launch the interactive graphical workbench",
	Long: `Launch the interactive graphical workbench in your default browser.

Starts an embedded local web server serving the UniRTM visual workspace to
manage tools, run tasks, inspect doctor health, and manage extensions.

Examples:
  # Start UI and open browser automatically
  unirtm ui

  # Start UI on a custom port without opening browser
  unirtm ui --port 8080 --no-open

  # Use alias
  unirtm dashboard
  unirtm web`,
	Args: cobra.NoArgs,
	RunE: runUI,
}

func runUI(cmd *cobra.Command, args []string) error {
	distFS, err := frontend.GetFS()
	if err != nil {
		return fmt.Errorf("failed to load embedded frontend assets: %w", err)
	}

	// Verify that index.html exists in the embedded filesystem
	if _, err := fs.Stat(distFS, "index.html"); err != nil {
		pterm.Warning.Println("Embedded frontend dist assets not found or empty.")
		pterm.Info.Println("Please run 'npm run build' in the frontend directory before compiling.")
	}

	listener, actualPort, err := listenOnAvailablePort(uiHost, uiPort)
	if err != nil {
		return fmt.Errorf("failed to bind server port: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d", uiHost, actualPort)

	mux := http.NewServeMux()

	// ─── API Endpoints ────────────────────────────────────────────────────────
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "ready",
			"version":     env.GitTag,
			"commit":      env.CommitHash,
			"platform":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"go_version":  runtime.Version(),
			"tools_count": 5,
			"tasks_count": 4,
			"timestamp":   time.Now().Unix(),
		})
	})

	// ─── Static SPA Handler ───────────────────────────────────────────────────
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// If static file exists, serve it; otherwise fallback to index.html (SPA routing)
		if _, err := fs.Stat(distFS, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA client-side routes
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Banner output
	pterm.DefaultSection.Println("UniRTM Workbench")
	pterm.Info.Printf("Local Server : %s\n", pterm.Cyan(url))
	pterm.Info.Printf("Process PID  : %d\n", os.Getpid())
	pterm.Info.Println("Press Ctrl+C to stop the workbench server.")

	// Auto-open browser in background
	if !uiNoOpen {
		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = openBrowser(url)
		}()
	}

	// Graceful shutdown handling
	idleConnsClosed := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		pterm.Info.Println("\nShutting down UniRTM Workbench...")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			pterm.Error.Printf("HTTP shutdown error: %v\n", err)
		}
		close(idleConnsClosed)
	}()

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("workbench server error: %w", err)
	}

	<-idleConnsClosed
	pterm.Success.Println("UniRTM Workbench stopped cleanly.")
	return nil
}

// listenOnAvailablePort attempts to bind to preferredPort, or finds an available port if preferredPort is busy.
func listenOnAvailablePort(host string, preferredPort int) (net.Listener, int, error) {
	addr := fmt.Sprintf("%s:%d", host, preferredPort)
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		return listener, preferredPort, nil
	}

	// If preferred port is in use, fallback to any available port
	fallbackAddr := fmt.Sprintf("%s:0", host)
	fallbackListener, err := net.Listen("tcp", fallbackAddr)
	if err != nil {
		return nil, 0, err
	}

	actualPort := fallbackListener.Addr().(*net.TCPAddr).Port
	return fallbackListener, actualPort, nil
}

// openBrowser opens the specified URL in the system's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
