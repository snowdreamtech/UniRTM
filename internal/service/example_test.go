// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service_test

import (
	"context"
	"fmt"
	"time"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
)

// ExampleAuditService_LogInstall demonstrates how to use the audit service to log tool installations
func ExampleAuditService_LogInstall() {
	// Initialize database (in-memory for example)
	db, err := database.Open(context.Background(), database.Config{
		Path:    ":memory:",
		WALMode: false,
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create audit repository
	auditRepo, err := sqlite.NewAuditRepository(db.Conn())
	if err != nil {
		panic(err)
	}
	defer auditRepo.Close()

	// Create audit service
	auditService := service.NewAuditService(auditRepo)

	// Log a successful installation
	ctx := context.Background()
	err = auditService.LogInstall(ctx, "node", "20.0.0", 2*time.Second, nil)
	if err != nil {
		panic(err)
	}

	// Log a failed installation
	installErr := fmt.Errorf("download failed: connection timeout")
	err = auditService.LogInstall(ctx, "python", "3.11.0", 5*time.Second, installErr)
	if err != nil {
		panic(err)
	}

	// Query recent logs
	logs, err := auditService.GetRecentLogs(ctx, 10)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d audit log entries\n", len(logs))
	// Output: Found 2 audit log entries
}

// ExampleAuditService_LogOperation demonstrates how to use the audit service with custom metadata
func ExampleAuditService_LogOperation() {
	// Initialize database (in-memory for example)
	db, err := database.Open(context.Background(), database.Config{
		Path:    ":memory:",
		WALMode: false,
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create audit repository
	auditRepo, err := sqlite.NewAuditRepository(db.Conn())
	if err != nil {
		panic(err)
	}
	defer auditRepo.Close()

	// Create audit service
	auditService := service.NewAuditService(auditRepo)

	// Log an operation with custom metadata
	ctx := context.Background()
	entry := &service.AuditLogEntry{
		Operation: service.OperationInstall,
		Tool:      "node",
		Version:   "20.0.0",
		Status:    service.StatusSuccess,
		Duration:  2500,
		Metadata: map[string]interface{}{
			"backend":      "github",
			"download_url": "https://github.com/nodejs/node/releases/download/v20.0.0/node-v20.0.0.tar.gz",
			"size_bytes":   45678901,
			"checksum":     "abc123def456",
		},
	}

	err = auditService.LogOperation(ctx, entry)
	if err != nil {
		panic(err)
	}

	fmt.Println("Operation logged successfully")
	// Output: Operation logged successfully
}

// ExampleAuditService_QueryAuditLogs demonstrates how to query audit logs with filters
func ExampleAuditService_QueryAuditLogs() {
	// Initialize database (in-memory for example)
	db, err := database.Open(context.Background(), database.Config{
		Path:    ":memory:",
		WALMode: false,
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create audit repository
	auditRepo, err := sqlite.NewAuditRepository(db.Conn())
	if err != nil {
		panic(err)
	}
	defer auditRepo.Close()

	// Create audit service
	auditService := service.NewAuditService(auditRepo)

	ctx := context.Background()

	// Log some operations
	_ = auditService.LogInstall(ctx, "node", "20.0.0", 2*time.Second, nil)
	_ = auditService.LogInstall(ctx, "python", "3.11.0", 3*time.Second, nil)
	_ = auditService.LogUninstall(ctx, "go", "1.20.0", 1*time.Second, nil)

	// Query logs by operation type
	installLogs, err := auditService.GetLogsByOperation(ctx, service.OperationInstall, 10)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d install operations\n", len(installLogs))
	// Output: Found 2 install operations
}

// ExampleAuditService_GetLogsByTool demonstrates how to query audit logs for a specific tool
func ExampleAuditService_GetLogsByTool() {
	// Initialize database (in-memory for example)
	db, err := database.Open(context.Background(), database.Config{
		Path:    ":memory:",
		WALMode: false,
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create audit repository
	auditRepo, err := sqlite.NewAuditRepository(db.Conn())
	if err != nil {
		panic(err)
	}
	defer auditRepo.Close()

	// Create audit service
	auditService := service.NewAuditService(auditRepo)

	ctx := context.Background()

	// Log operations for different tools
	_ = auditService.LogInstall(ctx, "node", "20.0.0", 2*time.Second, nil)
	_ = auditService.LogActivate(ctx, "node", "20.0.0", 100*time.Millisecond, nil)
	_ = auditService.LogInstall(ctx, "python", "3.11.0", 3*time.Second, nil)

	// Query logs for a specific tool
	nodeLogs, err := auditService.GetLogsByTool(ctx, "node", 10)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d operations for node\n", len(nodeLogs))
	// Output: Found 2 operations for node
}

// ExampleAuditService_GetLogsByStatus demonstrates how to query failed operations
func ExampleAuditService_GetLogsByStatus() {
	// Initialize database (in-memory for example)
	db, err := database.Open(context.Background(), database.Config{
		Path:    ":memory:",
		WALMode: false,
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create audit repository
	auditRepo, err := sqlite.NewAuditRepository(db.Conn())
	if err != nil {
		panic(err)
	}
	defer auditRepo.Close()

	// Create audit service
	auditService := service.NewAuditService(auditRepo)

	ctx := context.Background()

	// Log successful and failed operations
	_ = auditService.LogInstall(ctx, "node", "20.0.0", 2*time.Second, nil)
	_ = auditService.LogInstall(ctx, "python", "3.11.0", 5*time.Second, fmt.Errorf("download failed"))
	_ = auditService.LogInstall(ctx, "go", "1.21.0", 3*time.Second, fmt.Errorf("checksum mismatch"))

	// Query failed operations
	failedLogs, err := auditService.GetLogsByStatus(ctx, service.StatusFailure, 10)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d failed operations\n", len(failedLogs))
	// Output: Found 2 failed operations
}
