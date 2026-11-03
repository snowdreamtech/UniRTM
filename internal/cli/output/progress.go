// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressIndicator represents a progress indicator for long-running operations
type ProgressIndicator interface {
	// Start starts the progress indicator
	Start()

	// Update updates the progress with current and total values
	Update(current, total int64)

	// SetMessage sets the progress message
	SetMessage(message string)

	// Finish completes the progress indicator
	Finish()

	// Fail marks the progress as failed
	Fail(err error)
}

// ProgressOptions contains options for creating a progress indicator
type ProgressOptions struct {
	// Writer is the output writer (defaults to os.Stderr)
	Writer io.Writer

	// Message is the initial progress message
	Message string

	// ShowPercentage shows percentage completion
	ShowPercentage bool

	// ShowBytes shows byte counts (for downloads)
	ShowBytes bool

	// ShowSpeed shows transfer speed (for downloads)
	ShowSpeed bool

	// Width is the progress bar width in characters
	Width int

	// NoColor disables color output
	NoColor bool

	// Quiet suppresses progress output
	Quiet bool
}

// progressIndicator implements ProgressIndicator
type progressIndicator struct {
	writer         io.Writer
	message        string
	current        int64
	total          int64
	startTime      time.Time
	lastUpdate     time.Time
	showPercentage bool
	showBytes      bool
	showSpeed      bool
	width          int
	noColor        bool
	quiet          bool
	finished       bool
	mu             sync.Mutex
	ticker         *time.Ticker
	done           chan bool
}

// NewProgressIndicator creates a new progress indicator
func NewProgressIndicator(opts ProgressOptions) ProgressIndicator {
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}

	if opts.Width == 0 {
		opts.Width = 40
	}

	return &progressIndicator{
		writer:         opts.Writer,
		message:        opts.Message,
		showPercentage: opts.ShowPercentage,
		showBytes:      opts.ShowBytes,
		showSpeed:      opts.ShowSpeed,
		width:          opts.Width,
		noColor:        opts.NoColor,
		quiet:          opts.Quiet,
		done:           make(chan bool),
	}
}

// Start starts the progress indicator
func (p *progressIndicator) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.quiet {
		return
	}

	p.startTime = time.Now()
	p.lastUpdate = p.startTime

	// Start a ticker to update the progress display
	p.ticker = time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-p.ticker.C:
				p.render()
			case <-p.done:
				return
			}
		}
	}()
}

// Update updates the progress with current and total values
func (p *progressIndicator) Update(current, total int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current
	p.total = total
	p.lastUpdate = time.Now()
}

// SetMessage sets the progress message
func (p *progressIndicator) SetMessage(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.message = message
}

// Finish completes the progress indicator
func (p *progressIndicator) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.quiet || p.finished {
		return
	}

	p.finished = true

	if p.ticker != nil {
		p.ticker.Stop()
		close(p.done)
	}

	// Render final state
	p.renderFinal(true)
}

// Fail marks the progress as failed
func (p *progressIndicator) Fail(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.quiet || p.finished {
		return
	}

	p.finished = true

	if p.ticker != nil {
		p.ticker.Stop()
		close(p.done)
	}

	// Render final state with error
	p.renderFinal(false)
	if err != nil {
		color := ""
		reset := ""
		if !p.noColor {
			color = colorRed
			reset = colorReset
		}
		fmt.Fprintf(p.writer, "%sError: %v%s\n", color, err, reset)
	}
}

// render renders the current progress state
func (p *progressIndicator) render() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.quiet || p.finished {
		return
	}

	// Clear the line
	fmt.Fprint(p.writer, "\r\033[K")

	// Build progress string
	var parts []string

	// Add message
	if p.message != "" {
		parts = append(parts, p.message)
	}

	// Add progress bar
	if p.total > 0 {
		bar := p.buildProgressBar()
		parts = append(parts, bar)
	} else {
		// Indeterminate progress (spinner)
		spinner := p.buildSpinner()
		parts = append(parts, spinner)
	}

	// Add percentage
	if p.showPercentage && p.total > 0 {
		percentage := float64(p.current) / float64(p.total) * 100
		parts = append(parts, fmt.Sprintf("%.1f%%", percentage))
	}

	// Add byte counts
	if p.showBytes && p.total > 0 {
		parts = append(parts, fmt.Sprintf("%s / %s", formatBytes(p.current), formatBytes(p.total)))
	}

	// Add speed
	if p.showSpeed && p.current > 0 {
		elapsed := time.Since(p.startTime).Seconds()
		if elapsed > 0 {
			speed := float64(p.current) / elapsed
			parts = append(parts, fmt.Sprintf("%s/s", formatBytes(int64(speed))))
		}
	}

	fmt.Fprint(p.writer, strings.Join(parts, " "))
}

// renderFinal renders the final progress state
func (p *progressIndicator) renderFinal(success bool) {
	if p.quiet {
		return
	}

	// Clear the line
	fmt.Fprint(p.writer, "\r\033[K")

	// Build final message
	symbol := "✓"
	color := colorGreen
	if !success {
		symbol = "✗"
		color = colorRed
	}

	if p.noColor {
		fmt.Fprintf(p.writer, "%s %s", symbol, p.message)
	} else {
		fmt.Fprintf(p.writer, "%s%s%s %s", color, symbol, colorReset, p.message)
	}

	if p.showBytes && p.total > 0 {
		fmt.Fprintf(p.writer, " (%s)", formatBytes(p.total))
	}

	fmt.Fprintln(p.writer)
}

// buildProgressBar builds a progress bar string
func (p *progressIndicator) buildProgressBar() string {
	if p.total == 0 {
		return ""
	}

	percentage := float64(p.current) / float64(p.total)
	filled := int(float64(p.width) * percentage)
	if filled > p.width {
		filled = p.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	if p.noColor {
		return fmt.Sprintf("[%s]", bar)
	}

	return fmt.Sprintf("[%s%s%s]", colorCyan, bar, colorReset)
}

// buildSpinner builds a spinner string for indeterminate progress
func (p *progressIndicator) buildSpinner() string {
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	elapsed := time.Since(p.startTime).Milliseconds()
	index := (elapsed / 100) % int64(len(spinnerChars))

	spinner := spinnerChars[index]

	if p.noColor {
		return spinner
	}

	return fmt.Sprintf("%s%s%s", colorCyan, spinner, colorReset)
}

// formatBytes formats byte count in human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// NoOpProgressIndicator is a progress indicator that does nothing
type NoOpProgressIndicator struct{}

// NewNoOpProgressIndicator creates a no-op progress indicator
func NewNoOpProgressIndicator() ProgressIndicator {
	return &NoOpProgressIndicator{}
}

// Start does nothing
func (n *NoOpProgressIndicator) Start() {}

// Update does nothing
func (n *NoOpProgressIndicator) Update(current, total int64) {}

// SetMessage does nothing
func (n *NoOpProgressIndicator) SetMessage(message string) {}

// Finish does nothing
func (n *NoOpProgressIndicator) Finish() {}

// Fail does nothing
func (n *NoOpProgressIndicator) Fail(err error) {}
