// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

// AuditService provides high-level audit logging functionality
// Validates Requirements: 7.8 (Audit logging), 8.1 (Log all operations), 8.5 (Audit log recording)
type AuditService struct {
	repo repository.AuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(repo repository.AuditRepository) *AuditService {
	return &AuditService{
		repo: repo,
	}
}

// OperationType represents the type of operation being audited
type OperationType string

const (
	// OperationInstall represents a tool installation operation
	OperationInstall OperationType = "install"

	// OperationUninstall represents a tool uninstallation operation
	OperationUninstall OperationType = "uninstall"

	// OperationActivate represents a tool activation operation
	OperationActivate OperationType = "activate"

	// OperationDeactivate represents a tool deactivation operation
	OperationDeactivate OperationType = "deactivate"

	// OperationUpdate represents a tool update operation
	OperationUpdate OperationType = "update"

	// OperationCachePurge represents a cache purge operation
	OperationCachePurge OperationType = "cache_purge"

	// OperationConfigLoad represents a configuration load operation
	OperationConfigLoad OperationType = "config_load"

	// OperationConfigUpdate represents a configuration update operation
	OperationConfigUpdate OperationType = "config_update"

	// OperationIndexUpdate represents an index update operation
	OperationIndexUpdate OperationType = "index_update"

	// OperationVersionResolve represents a version resolution operation
	OperationVersionResolve OperationType = "version_resolve"
)

// OperationStatus represents the status of an operation
type OperationStatus string

const (
	// StatusSuccess indicates the operation completed successfully
	StatusSuccess OperationStatus = "success"

	// StatusFailure indicates the operation failed
	StatusFailure OperationStatus = "failure"
)

// AuditLogEntry represents a high-level audit log entry with convenience fields
type AuditLogEntry struct {
	// Operation is the type of operation being performed
	Operation OperationType

	// Tool is the name of the tool (optional, may be empty for non-tool operations)
	Tool string

	// Version is the version of the tool (optional)
	Version string

	// Status is the operation status (success or failure)
	Status OperationStatus

	// Error is the error message if the operation failed (optional)
	Error string

	// Duration is the operation duration in milliseconds
	Duration int64

	// Metadata contains additional operation-specific data (optional)
	Metadata map[string]interface{}
}

// LogOperation logs an operation to the audit log
// This is the primary method for recording audit entries
func (s *AuditService) LogOperation(ctx context.Context, entry *AuditLogEntry) error {
	// Convert metadata to JSON
	var metadataJSON string
	if entry.Metadata != nil && len(entry.Metadata) > 0 {
		data, err := json.Marshal(entry.Metadata)
		if err != nil {
			logger.Warn("Failed to marshal audit metadata", map[string]interface{}{
				"operation": entry.Operation,
				"error":     err.Error(),
			})
			metadataJSON = "{}"
		} else {
			metadataJSON = string(data)
		}
	}

	// Create repository audit entry
	auditEntry := &repository.AuditEntry{
		Operation: string(entry.Operation),
		Tool:      entry.Tool,
		Version:   entry.Version,
		Status:    string(entry.Status),
		Error:     entry.Error,
		Duration:  entry.Duration,
		Metadata:  metadataJSON,
	}

	// Log to database
	if err := s.repo.Log(ctx, auditEntry); err != nil {
		logger.ErrorWithErr(err, "Failed to write audit log", map[string]interface{}{
			"operation": entry.Operation,
			"tool":      entry.Tool,
			"version":   entry.Version,
		})
		return fmt.Errorf("log audit entry: %w", err)
	}

	// Also log to application logger for immediate visibility
	logFields := map[string]interface{}{
		"operation": entry.Operation,
		"status":    entry.Status,
		"duration":  entry.Duration,
	}
	if entry.Tool != "" {
		logFields["tool"] = entry.Tool
	}
	if entry.Version != "" {
		logFields["version"] = entry.Version
	}
	if entry.Error != "" {
		logFields["error"] = entry.Error
	}

	if entry.Status == StatusSuccess {
		logger.Info("Operation completed", logFields)
	} else {
		logger.Error("Operation failed", logFields)
	}

	return nil
}

// LogInstall logs a tool installation operation
func (s *AuditService) LogInstall(ctx context.Context, tool, version string, duration time.Duration, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationInstall,
		Tool:      tool,
		Version:   version,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogUninstall logs a tool uninstallation operation
func (s *AuditService) LogUninstall(ctx context.Context, tool, version string, duration time.Duration, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationUninstall,
		Tool:      tool,
		Version:   version,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogActivate logs a tool activation operation
func (s *AuditService) LogActivate(ctx context.Context, tool, version string, duration time.Duration, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationActivate,
		Tool:      tool,
		Version:   version,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogDeactivate logs a tool deactivation operation
func (s *AuditService) LogDeactivate(ctx context.Context, tool, version string, duration time.Duration, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationDeactivate,
		Tool:      tool,
		Version:   version,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogUpdate logs a tool update operation
func (s *AuditService) LogUpdate(ctx context.Context, tool, oldVersion, newVersion string, duration time.Duration, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationUpdate,
		Tool:      tool,
		Version:   newVersion,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
		Metadata: map[string]interface{}{
			"old_version": oldVersion,
			"new_version": newVersion,
		},
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogCachePurge logs a cache purge operation
func (s *AuditService) LogCachePurge(ctx context.Context, duration time.Duration, purgedCount int, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationCachePurge,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
		Metadata: map[string]interface{}{
			"purged_count": purgedCount,
		},
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogConfigLoad logs a configuration load operation
func (s *AuditService) LogConfigLoad(ctx context.Context, configPath string, duration time.Duration, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationConfigLoad,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
		Metadata: map[string]interface{}{
			"config_path": configPath,
		},
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogConfigUpdate logs a configuration update operation
func (s *AuditService) LogConfigUpdate(ctx context.Context, configPath string, duration time.Duration, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationConfigUpdate,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
		Metadata: map[string]interface{}{
			"config_path": configPath,
		},
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogIndexUpdate logs an index update operation
func (s *AuditService) LogIndexUpdate(ctx context.Context, backend string, duration time.Duration, toolCount int, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationIndexUpdate,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
		Metadata: map[string]interface{}{
			"backend":    backend,
			"tool_count": toolCount,
		},
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// LogVersionResolve logs a version resolution operation
func (s *AuditService) LogVersionResolve(ctx context.Context, tool, versionSpec, resolvedVersion string, duration time.Duration, err error) error {
	entry := &AuditLogEntry{
		Operation: OperationVersionResolve,
		Tool:      tool,
		Version:   resolvedVersion,
		Duration:  duration.Milliseconds(),
		Status:    StatusSuccess,
		Metadata: map[string]interface{}{
			"version_spec":     versionSpec,
			"resolved_version": resolvedVersion,
		},
	}

	if err != nil {
		entry.Status = StatusFailure
		entry.Error = err.Error()
	}

	return s.LogOperation(ctx, entry)
}

// QueryAuditLogs queries audit logs with filters
func (s *AuditService) QueryAuditLogs(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
	entries, err := s.repo.Query(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	return entries, nil
}

// GetRecentLogs returns the most recent audit logs
func (s *AuditService) GetRecentLogs(ctx context.Context, limit int) ([]*repository.AuditEntry, error) {
	filter := repository.AuditFilter{
		Limit: limit,
	}
	return s.QueryAuditLogs(ctx, filter)
}

// GetLogsByOperation returns audit logs filtered by operation type
func (s *AuditService) GetLogsByOperation(ctx context.Context, operation OperationType, limit int) ([]*repository.AuditEntry, error) {
	filter := repository.AuditFilter{
		Operation: string(operation),
		Limit:     limit,
	}
	return s.QueryAuditLogs(ctx, filter)
}

// GetLogsByTool returns audit logs filtered by tool name
func (s *AuditService) GetLogsByTool(ctx context.Context, tool string, limit int) ([]*repository.AuditEntry, error) {
	filter := repository.AuditFilter{
		Tool:  tool,
		Limit: limit,
	}
	return s.QueryAuditLogs(ctx, filter)
}

// GetLogsByStatus returns audit logs filtered by status
func (s *AuditService) GetLogsByStatus(ctx context.Context, status OperationStatus, limit int) ([]*repository.AuditEntry, error) {
	filter := repository.AuditFilter{
		Status: string(status),
		Limit:  limit,
	}
	return s.QueryAuditLogs(ctx, filter)
}

// GetLogsByTimeRange returns audit logs within a time range
func (s *AuditService) GetLogsByTimeRange(ctx context.Context, startTime, endTime time.Time, limit int) ([]*repository.AuditEntry, error) {
	filter := repository.AuditFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     limit,
	}
	return s.QueryAuditLogs(ctx, filter)
}
