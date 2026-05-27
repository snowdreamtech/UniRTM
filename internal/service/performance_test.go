// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/repository"
)

// performanceMockAuditRepo for testing PerformanceMonitor
type performanceMockAuditRepo struct {
	repository.AuditRepository
	entries []*repository.AuditEntry
}

func (m *performanceMockAuditRepo) Log(ctx context.Context, entry *repository.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func TestPerformanceMonitor_RecordAndReport(t *testing.T) {
	auditRepo := &performanceMockAuditRepo{}
	pm := NewPerformanceMonitor(auditRepo)

	ctx := context.Background()

	// Record some metrics
	pm.Record(ctx, OperationMetric{
		Tool:      "node",
		Version:   "20.0.0",
		Phase:     PhaseDownload,
		StartedAt: time.Now(),
		Duration:  100 * time.Millisecond,
		Success:   true,
	})

	pm.Record(ctx, OperationMetric{
		Tool:      "node",
		Version:   "20.0.0",
		Phase:     PhaseDownload,
		StartedAt: time.Now(),
		Duration:  200 * time.Millisecond,
		Success:   true,
	})

	pm.Record(ctx, OperationMetric{
		Tool:      "node",
		Version:   "20.0.0",
		Phase:     PhaseCacheHit,
		StartedAt: time.Now(),
		Duration:  10 * time.Millisecond,
		Success:   true,
	})

	pm.Record(ctx, OperationMetric{
		Tool:      "node",
		Version:   "20.0.0",
		Phase:     PhaseCacheMiss,
		StartedAt: time.Now(),
		Duration:  5 * time.Millisecond,
		Success:   true,
	})

	// Test CacheHitRate
	hitRate := pm.CacheHitRate()
	if hitRate != 50.0 {
		t.Errorf("expected 50.0%% cache hit rate, got %f", hitRate)
	}

	// Test Report
	report := pm.Report(PhaseDownload)
	if report.Count != 2 {
		t.Errorf("expected 2 downloads, got %d", report.Count)
	}
	if report.Average != 150*time.Millisecond {
		t.Errorf("expected avg 150ms, got %v", report.Average)
	}
	if report.Min != 100*time.Millisecond {
		t.Errorf("expected min 100ms, got %v", report.Min)
	}
	if report.Max != 200*time.Millisecond {
		t.Errorf("expected max 200ms, got %v", report.Max)
	}

	// Test AllReports
	allReports := pm.AllReports()
	// PhaseDownload has 2
	foundDownload := false
	for _, r := range allReports {
		if r.Phase == PhaseDownload {
			foundDownload = true
			break
		}
	}
	if !foundDownload {
		t.Error("expected PhaseDownload in all reports")
	}

	// Test FormatReport
	reportStr := pm.FormatReport(report)
	if reportStr == "" {
		t.Error("expected non-empty report string")
	}

	// Audit logs check
	if len(auditRepo.entries) != 4 {
		t.Errorf("expected 4 audit entries, got %d", len(auditRepo.entries))
	}
}

func TestPerformanceMonitor_Track(t *testing.T) {
	pm := NewPerformanceMonitor(nil)

	done := pm.Track(context.Background(), "go", "1.21.0", PhaseInstall)
	time.Sleep(10 * time.Millisecond) // Ensure duration is > 0
	done(true)

	report := pm.Report(PhaseInstall)
	if report.Count != 1 {
		t.Errorf("expected 1 install, got %d", report.Count)
	}
	if report.Average == 0 {
		t.Error("expected average > 0")
	}
}

func TestPerformanceMonitor_DetectRegressions(t *testing.T) {
	pm := NewPerformanceMonitor(nil)

	pm.SetBaseline(PhaseExtract, 100*time.Millisecond)

	pm.Record(context.Background(), OperationMetric{
		Tool:      "python",
		Version:   "3.11",
		Phase:     PhaseExtract,
		StartedAt: time.Now(),
		Duration:  200 * time.Millisecond,
		Success:   true,
	})

	regressions := pm.DetectRegressions(50.0) // 50% threshold
	if pct, ok := regressions[PhaseExtract]; !ok {
		t.Error("expected regression for PhaseExtract")
	} else if pct != 100.0 {
		t.Errorf("expected 100%% regression, got %f", pct)
	}
}

func TestPerformanceMonitor_percentile(t *testing.T) {
	durations := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}

	p50 := percentile(durations, 50)
	if p50 != 30*time.Millisecond {
		t.Errorf("expected p50=30ms, got %v", p50)
	}

	p95 := percentile(durations, 95)
	if p95 != 50*time.Millisecond {
		t.Errorf("expected p95=50ms, got %v", p95)
	}
}
