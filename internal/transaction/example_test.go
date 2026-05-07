package transaction_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/transaction"
)

// Example_basicTransaction demonstrates basic transaction usage
func Example_basicTransaction() {
	// Open database
	db, err := database.Open(context.Background(), database.Config{
		Path:    "/tmp/unirtm.db",
		WALMode: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create transaction manager
	tm := transaction.NewSQLiteTransactionManager(db.Conn())

	// Begin transaction
	ctx := context.Background()
	tx, err := tm.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			tx.Rollback()
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
		log.Fatal(err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Installation created successfully")
}

// Example_multiRepositoryTransaction demonstrates atomic operations across multiple repositories
func Example_multiRepositoryTransaction() {
	db, err := database.Open(context.Background(), database.Config{
		Path:    "/tmp/unirtm.db",
		WALMode: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tm := transaction.NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Begin transaction
	tx, err := tm.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Automatic rollback on error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("rollback failed: %v", rbErr)
			}
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
		log.Fatal(err)
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
		log.Fatal(err)
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
		log.Fatal(err)
	}

	// 4. Cache installation metadata
	err = tx.CacheRepo().Set(ctx, "python:3.11.0:metadata", []byte("cached metadata"), 24*time.Hour)
	if err != nil {
		log.Fatal(err)
	}

	// Commit all operations atomically
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Multi-repository transaction completed successfully")
}

// Example_errorHandlingWithRollback demonstrates automatic rollback on error
func Example_errorHandlingWithRollback() {
	db, err := database.Open(context.Background(), database.Config{
		Path:    "/tmp/unirtm.db",
		WALMode: true,
	})
	if err != nil {
		log.Fatal(err)
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
				if rbErr := tx.Rollback(); rbErr != nil {
					log.Printf("rollback failed: %v", rbErr)
				}
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
	db, err := database.Open(context.Background(), database.Config{
		Path:    "/tmp/unirtm.db",
		WALMode: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tm := transaction.NewSQLiteTransactionManager(db.Conn())

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := tm.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
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
		log.Printf("Operation failed: %v", err)
		return
	}

	// Commit before context timeout
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction completed before timeout")
}
