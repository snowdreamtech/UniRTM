// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

// OperationPhase represents a timed phase of a tool operation.
type OperationPhase string

const (
	PhaseDownload  OperationPhase = "download"
	PhaseExtract   OperationPhase = "extract"
	PhaseInstall   OperationPhase = "install"
	PhaseActivate  OperationPhase = "activate"
	PhaseCacheHit  OperationPhase = "cache_hit"
	PhaseCacheMiss OperationPhase = "cache_miss"
)

// OperationMetric records the duration of a single timed operation.
//
// Validates Requirement: 17.1 (track operation durations)
type OperationMetric struct {
	Tool      string
	Version   string
	Phase     OperationPhase
	StartedAt time.Time
	Duration  time.Duration
	Success   bool
}

// PerformanceReport summarizes performance metrics.
//
// Validates Requirements: 17.5 (performance data querying), 17.6 (reports)
type PerformanceReport struct {
	Phase   OperationPhase
	Count   int
	Average time.Duration
	P50     time.Duration
	P95     time.Duration
	P99     time.Duration
	Min     time.Duration
	Max     time.Duration
}

// PerformanceMonitor tracks operation durations and cache hit rates.
//
// Validates Requirements: 17.1, 17.2, 17.3, 17.4, 17.5, 17.6, 17.7
type PerformanceMonitor struct {
	mu        sync.Mutex
	metrics   []OperationMetric
	auditRepo repository.AuditRepository

	// Cache statistics
	cacheHits   int64
	cacheMisses int64

	// Baseline for regression detection (Req 17.7)
	baselines map[OperationPhase]time.Duration
}

// NewPerformanceMonitor creates a new PerformanceMonitor.
func NewPerformanceMonitor(auditRepo repository.AuditRepository) *PerformanceMonitor {
	return &PerformanceMonitor{
		auditRepo: auditRepo,
		baselines: make(map[OperationPhase]time.Duration),
	}
}

// Record records an operation metric.
//
// Validates Requirements: 17.1, 17.2, 17.3, 17.4
func (pm *PerformanceMonitor) Record(ctx context.Context, metric OperationMetric) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.metrics = append(pm.metrics, metric)

	// Update cache hit/miss counters
	switch metric.Phase {
	case PhaseCacheHit:
		pm.cacheHits++
	case PhaseCacheMiss:
		pm.cacheMisses++
	}

	logger.Debug("Performance metric recorded", map[string]interface{}{
		"tool":     metric.Tool,
		"phase":    string(metric.Phase),
		"duration": metric.Duration.String(),
	})

	// Persist to audit log for long-term storage (Req 17.4)
	if pm.auditRepo != nil {
		_ = pm.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: metric.StartedAt,
			Operation: "perf_" + string(metric.Phase),
			Tool:      metric.Tool,
			Version:   metric.Version,
			Status: func() string {
				if metric.Success {
					return "success"
				}
				return "failure"
			}(),
			Duration: metric.Duration.Milliseconds(),
		})
	}
}

// Track is a convenience helper that times a function call and records the result.
//
// Usage:
//
//	done := pm.Track(ctx, tool, version, PhaseDownload)
//	err := doDownload(...)
//	done(err == nil)
func (pm *PerformanceMonitor) Track(ctx context.Context, tool, version string, phase OperationPhase) func(success bool) {
	start := time.Now()
	return func(success bool) {
		pm.Record(ctx, OperationMetric{
			Tool:      tool,
			Version:   version,
			Phase:     phase,
			StartedAt: start,
			Duration:  time.Since(start),
			Success:   success,
		})
	}
}

// CacheHitRate returns the current cache hit rate as a percentage (0-100).
//
// Validates Requirement: 17.2 (track cache hit rates)
func (pm *PerformanceMonitor) CacheHitRate() float64 {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	total := pm.cacheHits + pm.cacheMisses
	if total == 0 {
		return 0.0
	}
	return float64(pm.cacheHits) / float64(total) * 100.0
}

// Report generates a performance report for the given phase.
//
// Validates Requirements: 17.5, 17.6 (percentile reports)
func (pm *PerformanceMonitor) Report(phase OperationPhase) *PerformanceReport {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var durations []time.Duration
	for _, m := range pm.metrics {
		if m.Phase == phase && m.Success {
			durations = append(durations, m.Duration)
		}
	}

	if len(durations) == 0 {
		return &PerformanceReport{Phase: phase}
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	n := len(durations)
	total := time.Duration(0)
	for _, d := range durations {
		total += d
	}

	return &PerformanceReport{
		Phase:   phase,
		Count:   n,
		Average: total / time.Duration(n),
		P50:     percentile(durations, 50),
		P95:     percentile(durations, 95),
		P99:     percentile(durations, 99),
		Min:     durations[0],
		Max:     durations[n-1],
	}
}

// AllReports generates a performance report for every phase that has data.
func (pm *PerformanceMonitor) AllReports() []*PerformanceReport {
	phases := []OperationPhase{
		PhaseDownload, PhaseExtract, PhaseInstall, PhaseActivate,
	}
	var reports []*PerformanceReport
	for _, p := range phases {
		r := pm.Report(p)
		if r.Count > 0 {
			reports = append(reports, r)
		}
	}
	return reports
}

// SetBaseline sets the baseline duration for regression detection.
//
// Validates Requirement: 17.7 (detect performance regressions)
func (pm *PerformanceMonitor) SetBaseline(phase OperationPhase, d time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.baselines[phase] = d
}

// DetectRegressions compares current p95 against baselines.
//
// Returns a map of phase → regression percentage (positive = slower than baseline).
//
// Validates Requirement: 17.7
func (pm *PerformanceMonitor) DetectRegressions(threshold float64) map[OperationPhase]float64 {
	regressions := make(map[OperationPhase]float64)

	pm.mu.Lock()
	baselines := make(map[OperationPhase]time.Duration, len(pm.baselines))
	for k, v := range pm.baselines {
		baselines[k] = v
	}
	pm.mu.Unlock()

	for phase, baseline := range baselines {
		report := pm.Report(phase)
		if report.Count == 0 || baseline <= 0 {
			continue
		}
		regression := (float64(report.P95) - float64(baseline)) / float64(baseline) * 100.0
		if regression > threshold {
			regressions[phase] = regression
			logger.Warn("Performance regression detected", map[string]interface{}{
				"phase":      string(phase),
				"baseline":   baseline.String(),
				"p95":        report.P95.String(),
				"regression": fmt.Sprintf("%.1f%%", regression),
			})
		}
	}

	return regressions
}

// FormatReport formats a PerformanceReport as a human-readable string.
func (pm *PerformanceMonitor) FormatReport(r *PerformanceReport) string {
	if r.Count == 0 {
		return fmt.Sprintf("%-10s  no data\n", string(r.Phase))
	}
	return fmt.Sprintf(
		"%-10s  n=%-4d  avg=%-8s  p50=%-8s  p95=%-8s  p99=%-8s  min=%-8s  max=%s\n",
		string(r.Phase), r.Count,
		r.Average.Round(time.Millisecond),
		r.P50.Round(time.Millisecond),
		r.P95.Round(time.Millisecond),
		r.P99.Round(time.Millisecond),
		r.Min.Round(time.Millisecond),
		r.Max.Round(time.Millisecond),
	)
}

// percentile returns the value at the given percentile (0-100) of a sorted slice.
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100.0*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
