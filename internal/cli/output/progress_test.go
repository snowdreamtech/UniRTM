// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProgressIndicator(t *testing.T) {
	tests := []struct {
		name string
		opts ProgressOptions
	}{
		{
			name: "default options",
			opts: ProgressOptions{},
		},
		{
			name: "with message",
			opts: ProgressOptions{
				Message: "Downloading...",
			},
		},
		{
			name: "with custom width",
			opts: ProgressOptions{
				Width: 60,
			},
		},
		{
			name: "with all options",
			opts: ProgressOptions{
				Message:        "Processing...",
				ShowPercentage: true,
				ShowBytes:      true,
				ShowSpeed:      true,
				Width:          50,
				NoColor:        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := NewProgressIndicator(tt.opts)
			require.NotNil(t, progress)
		})
	}
}

func TestProgressIndicator_StartFinish(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:  buf,
		Message: "Test operation",
		NoColor: true,
	})

	progress.Start()
	time.Sleep(200 * time.Millisecond)
	progress.Finish()

	output := buf.String()
	assert.Contains(t, output, "Test operation")
	assert.Contains(t, output, "✓")
}

func TestProgressIndicator_Update(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:         buf,
		Message:        "Downloading",
		ShowPercentage: true,
		NoColor:        true,
	})

	progress.Start()
	progress.Update(50, 100)
	time.Sleep(200 * time.Millisecond)
	progress.Finish()

	// Progress should have been rendered
	assert.NotEmpty(t, buf.String())
}

func TestProgressIndicator_SetMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:  buf,
		Message: "Initial message",
		NoColor: true,
	})

	progress.Start()
	progress.SetMessage("Updated message")
	time.Sleep(200 * time.Millisecond)
	progress.Finish()

	output := buf.String()
	assert.Contains(t, output, "Updated message")
}

func TestProgressIndicator_Fail(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:  buf,
		Message: "Test operation",
		NoColor: true,
	})

	progress.Start()
	time.Sleep(100 * time.Millisecond)
	progress.Fail(errors.New("operation failed"))

	output := buf.String()
	assert.Contains(t, output, "Test operation")
	assert.Contains(t, output, "✗")
	assert.Contains(t, output, "operation failed")
}

func TestProgressIndicator_Quiet(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:  buf,
		Message: "Test operation",
		Quiet:   true,
	})

	progress.Start()
	progress.Update(50, 100)
	time.Sleep(100 * time.Millisecond)
	progress.Finish()

	// Quiet mode should produce no output
	assert.Empty(t, buf.String())
}

func TestProgressIndicator_ShowBytes(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:    buf,
		Message:   "Downloading",
		ShowBytes: true,
		NoColor:   true,
	})

	progress.Start()
	progress.Update(1024*1024, 10*1024*1024) // 1 MB of 10 MB
	time.Sleep(200 * time.Millisecond)
	progress.Finish()

	output := buf.String()
	// Should show byte counts in human-readable format
	assert.Contains(t, output, "MB")
}

func TestProgressIndicator_ShowSpeed(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:    buf,
		Message:   "Downloading",
		ShowSpeed: true,
		NoColor:   true,
	})

	progress.Start()
	time.Sleep(100 * time.Millisecond)
	progress.Update(1024*1024, 10*1024*1024) // 1 MB downloaded
	time.Sleep(200 * time.Millisecond)
	progress.Finish()

	output := buf.String()
	// Should show speed
	assert.Contains(t, output, "/s")
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "bytes",
			bytes: 512,
			want:  "512 B",
		},
		{
			name:  "kilobytes",
			bytes: 1024,
			want:  "1.0 KB",
		},
		{
			name:  "megabytes",
			bytes: 1024 * 1024,
			want:  "1.0 MB",
		},
		{
			name:  "gigabytes",
			bytes: 1024 * 1024 * 1024,
			want:  "1.0 GB",
		},
		{
			name:  "terabytes",
			bytes: 1024 * 1024 * 1024 * 1024,
			want:  "1.0 TB",
		},
		{
			name:  "fractional megabytes",
			bytes: 1536 * 1024, // 1.5 MB
			want:  "1.5 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProgressIndicator_IndeterminateProgress(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:  buf,
		Message: "Processing",
		NoColor: true,
	})

	progress.Start()
	// Don't set total - should show spinner
	progress.Update(0, 0)
	time.Sleep(200 * time.Millisecond)
	progress.Finish()

	output := buf.String()
	assert.Contains(t, output, "Processing")
}

func TestProgressIndicator_MultipleUpdates(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:         buf,
		Message:        "Downloading",
		ShowPercentage: true,
		NoColor:        true,
	})

	progress.Start()

	// Simulate progressive updates
	for i := int64(0); i <= 100; i += 25 {
		progress.Update(i, 100)
		time.Sleep(50 * time.Millisecond)
	}

	progress.Finish()

	output := buf.String()
	assert.Contains(t, output, "Downloading")
	assert.Contains(t, output, "✓")
}

func TestProgressIndicator_DoubleFinish(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:  buf,
		Message: "Test operation",
		NoColor: true,
	})

	progress.Start()
	time.Sleep(100 * time.Millisecond)
	progress.Finish()

	// Second finish should be a no-op
	initialOutput := buf.String()
	progress.Finish()
	assert.Equal(t, initialOutput, buf.String())
}

func TestProgressIndicator_FinishAfterFail(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:  buf,
		Message: "Test operation",
		NoColor: true,
	})

	progress.Start()
	time.Sleep(100 * time.Millisecond)
	progress.Fail(errors.New("failed"))

	// Finish after fail should be a no-op
	initialOutput := buf.String()
	progress.Finish()
	assert.Equal(t, initialOutput, buf.String())
}

func TestNoOpProgressIndicator(t *testing.T) {
	progress := NewNoOpProgressIndicator()
	require.NotNil(t, progress)

	// All methods should be safe to call and do nothing
	progress.Start()
	progress.Update(50, 100)
	progress.SetMessage("test")
	progress.Finish()
	progress.Fail(errors.New("test error"))
}

func TestProgressIndicator_BuildProgressBar(t *testing.T) {
	progress := &progressIndicator{
		width:   10,
		noColor: true,
		current: 50,
		total:   100,
	}

	bar := progress.buildProgressBar()
	assert.Contains(t, bar, "[")
	assert.Contains(t, bar, "]")
	// Should have 5 filled and 5 empty characters (50% of 10)
	assert.Contains(t, bar, "█")
	assert.Contains(t, bar, "░")
}

func TestProgressIndicator_BuildSpinner(t *testing.T) {
	progress := &progressIndicator{
		noColor:   true,
		startTime: time.Now(),
	}

	spinner := progress.buildSpinner()
	assert.NotEmpty(t, spinner)
	// Should be one of the spinner characters
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	found := false
	for _, char := range spinnerChars {
		if strings.Contains(spinner, char) {
			found = true
			break
		}
	}
	assert.True(t, found, "spinner should contain one of the spinner characters")
}

func TestProgressIndicator_ColorOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:  buf,
		Message: "Test operation",
		NoColor: false, // Enable colors
	})

	progress.Start()
	time.Sleep(100 * time.Millisecond)
	progress.Finish()

	output := buf.String()
	// Should contain ANSI color codes
	assert.Contains(t, output, "\033[")
}

func TestProgressIndicator_ShowPercentage(t *testing.T) {
	buf := &bytes.Buffer{}
	progress := NewProgressIndicator(ProgressOptions{
		Writer:         buf,
		Message:        "Processing",
		ShowPercentage: true,
		NoColor:        true,
	})

	progress.Start()
	progress.Update(75, 100)
	time.Sleep(200 * time.Millisecond)
	progress.Finish()

	output := buf.String()
	// Should show percentage somewhere in the output
	assert.Regexp(t, `\d+\.\d+%`, output)
}
