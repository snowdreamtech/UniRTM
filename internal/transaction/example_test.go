// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package transaction_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/transaction"
)

// Example_basicTransaction demonstrates basic transaction usage
func Example_basicTransaction() {
	dbPath := filepath.Join(os.TempDir(), fmt.Sprintf("unirtm_tx_basic_%d.db", time.Now().UnixNano()))
	defer os.Remove(dbPath)

	// Open database
	db, err := database.Open(context.Background(), database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		fmt.Printf("Open database error: %v\n", err)
		return
	}
	defer db.Close()

	// Create transaction manager
	tm := transaction.NewSQLiteTransactionManager(db.Conn())

	// Begin transaction
	ctx := context.Background()
	tx, err := tm.Begin(ctx)
	if err != nil {
		fmt.Printf("Begin error: %v\n", err)
		return
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Create installation
	installation := &repository.Installation{
		Tool:        "node",
		Version:     "20.0.0",
		Backend:     "github",
		Provider:    "node",
		InstallPath: "/opt/unirtm/node/20.0.0",
		Checksum:    "abc123",
		Metadata:    "{}",
	}

	err = tx.InstallationRepo().Create(ctx, installation)
	if err != nil {
		fmt.Printf("Create error: %v\n", err)
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		fmt.Printf("Commit error: %v\n", err)
		return
	}

	fmt.Println("Installation created successfully")
}

// Example_multiRepositoryTransaction demonstrates atomic operations across multiple repositories
func Example_multiRepositoryTransaction() {
	dbPath := filepath.Join(os.TempDir(), fmt.Sprintf("unirtm_tx_multi_%d.db", time.Now().UnixNano()))
	defer os.Remove(dbPath)

	db, err := database.Open(context.Background(), database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		fmt.Printf("Open database error: %v\n", err)
		return
	}
	defer db.Close()

	tm := transaction.NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Begin transaction
	tx, err := tm.Begin(ctx)
	if err != nil {
		fmt.Printf("Begin error: %v\n", err)
		return
	}

	// Automatic rollback on error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Create installation
	installation := &repository.Installation{
		Tool:        "python",
		Version:     "3.11.0",
		Backend:     "github",
		Provider:    "python",
		InstallPath: "/opt/unirtm/python/3.11.0",
		Checksum:    "def456",
		Metadata:    "{}",
	}
	err = tx.InstallationRepo().Create(ctx, installation)
	if err != nil {
		fmt.Printf("Create error: %v\n", err)
		return
	}

	// 2. Log audit entry
	auditEntry := &repository.AuditEntry{
		Operation: "install",
		Tool:      installation.Tool,
		Version:   installation.Version,
		Status:    "success",
		Duration:  5000,
		Metadata:  "{}",
	}
	err = tx.AuditRepo().Log(ctx, auditEntry)
	if err != nil {
		fmt.Printf("Log audit error: %v\n", err)
		return
	}

	// 3. Update tool index
	indexEntry := &repository.IndexEntry{
		Tool:        installation.Tool,
		Description: "Python programming language",
		Homepage:    "https://python.org",
		License:     "PSF",
		Backend:     installation.Backend,
		Metadata:    "{}",
	}
	err = tx.IndexRepo().Upsert(ctx, indexEntry)
	if err != nil {
		fmt.Printf("Upsert index error: %v\n", err)
		return
	}

	// 4. Cache installation metadata
	err = tx.CacheRepo().Set(ctx, "python:3.11.0:metadata", []byte("cached metadata"), 24*time.Hour)
	if err != nil {
		fmt.Printf("Set cache error: %v\n", err)
		return
	}

	// Commit all operations atomically
	err = tx.Commit()
	if err != nil {
		fmt.Printf("Commit error: %v\n", err)
		return
	}

	fmt.Println("Multi-repository transaction completed successfully")
}

// Example_errorHandlingWithRollback demonstrates automatic rollback on error
func Example_errorHandlingWithRollback() {
	dbPath := filepath.Join(os.TempDir(), fmt.Sprintf("unirtm_tx_err_%d.db", time.Now().UnixNano()))
	defer os.Remove(dbPath)

	db, err := database.Open(context.Background(), database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		fmt.Printf("Open database error: %v\n", err)
		return
	}
	defer db.Close()

	tm := transaction.NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Function that performs operations in a transaction
	performInstallation := func() error {
		tx, err := tm.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}

		// Automatic rollback on any error
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()

		// Create installation
		installation := &repository.Installation{
			Tool:        "go",
			Version:     "1.21.0",
			Backend:     "github",
			Provider:    "go",
			InstallPath: "/opt/unirtm/go/1.21.0",
			Checksum:    "ghi789",
			Metadata:    "{}",
		}

		err = tx.InstallationRepo().Create(ctx, installation)
		if err != nil {
			return fmt.Errorf("create installation: %w", err)
		}

		// Simulate an error condition
		if installation.Version == "1.21.0" {
			return fmt.Errorf("simulated error: version validation failed")
		}

		// This commit will never be reached due to the error above
		return tx.Commit()
	}

	// Call the function
	err = performInstallation()
	if err != nil {
		fmt.Printf("Installation failed (transaction rolled back): %v\n", err)
	}
}

// Example_contextCancellation demonstrates handling context cancellation
func Example_contextCancellation() {
	dbPath := filepath.Join(os.TempDir(), fmt.Sprintf("unirtm_tx_ctx_%d.db", time.Now().UnixNano()))
	defer os.Remove(dbPath)

	db, err := database.Open(context.Background(), database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		fmt.Printf("Open database error: %v\n", err)
		return
	}
	defer db.Close()

	tm := transaction.NewSQLiteTransactionManager(db.Conn())

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := tm.Begin(ctx)
	if err != nil {
		fmt.Printf("Begin error: %v\n", err)
		return
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Perform operations with the context
	installation := &repository.Installation{
		Tool:        "rust",
		Version:     "1.70.0",
		Backend:     "github",
		Provider:    "rust",
		InstallPath: "/opt/unirtm/rust/1.70.0",
		Checksum:    "jkl012",
		Metadata:    "{}",
	}

	err = tx.InstallationRepo().Create(ctx, installation)
	if err != nil {
		fmt.Printf("Operation failed: %v\n", err)
		return
	}

	// Commit before context timeout
	err = tx.Commit()
	if err != nil {
		fmt.Printf("Commit error: %v\n", err)
		return
	}

	fmt.Println("Transaction completed before timeout")
}
