// Package bench provides performance benchmarks for UniRTM core operations.
//
// Run with:
//
//	go test ./tests/bench/... -bench=. -benchmem -benchtime=3s
//
// Validates Requirements: 17.x (performance monitoring), Task 33.1
package bench

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/service"
)

// BenchmarkVersionParse measures how fast version strings are parsed.
func BenchmarkVersionParse(b *testing.B) {
	inputs := []string{
		"20.0.0", "latest", "18.12.1", "1.21.0", "3.11.5",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		_, _ = service.ParseVersion(input)
	}
}

// BenchmarkConfigValidation measures configuration validation speed.
func BenchmarkConfigValidation(b *testing.B) {
	validator := service.NewConfigValidator(nil, []string{"github", "aqua", "http"})
	ctx := context.Background()

	// Build a config with 50 tools
	cfg := &config.Config{
		Tools: make(map[string]config.ToolConfig, 50),
	}
	for i := 0; i < 50; i++ {
		name := "tool" + string(rune('a'+i%26))
		cfg.Tools[name] = config.ToolConfig{
			Version: "1.0.0",
			Backend: "github",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(ctx, cfg)
	}
}

// BenchmarkShimGeneration measures shim script generation speed.
func BenchmarkShimGeneration(b *testing.B) {
	shimsDir := b.TempDir()
	gen := service.NewGenerator(shimsDir, b.TempDir())
	ctx := context.Background()
	tools := []string{"node", "python", "go", "ruby", "rust", "java", "deno", "bun"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool := tools[i%len(tools)]
		_ = gen.GenerateShim(ctx, tool)
	}
}

// BenchmarkPerformanceMonitorRecord measures metric recording throughput.
func BenchmarkPerformanceMonitorRecord(b *testing.B) {
	pm := service.NewPerformanceMonitor(nil)
	ctx := context.Background()
	metric := service.OperationMetric{
		Tool:    "node",
		Version: "20.0.0",
		Phase:   service.PhaseDownload,
		Success: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.Record(ctx, metric)
	}
}

// BenchmarkPerformanceReport measures report generation (percentile calculation).
func BenchmarkPerformanceReport(b *testing.B) {
	pm := service.NewPerformanceMonitor(nil)
	ctx := context.Background()

	// Pre-populate with 1000 metrics
	for i := 0; i < 1000; i++ {
		pm.Record(ctx, service.OperationMetric{
			Phase:   service.PhaseDownload,
			Success: true,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pm.Report(service.PhaseDownload)
	}
}

