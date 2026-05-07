// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package property contains property-based tests for UniRTM.
//
// Property-based tests verify universal properties that should hold for all inputs,
// complementing example-based unit tests with comprehensive input coverage.
package property

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/transaction"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// Feature: unirtm, Property 8: Database Persistence Round-Trip
//
// **Validates: Requirements 2.2, 2.3, 2.4, 2.5**
//
// For any valid data object (Installation, CacheEntry, AuditEntry, IndexEntry),
// storing it in the database and retrieving it SHALL produce an equivalent object.
//
// This property ensures that:
// 1. Database serialization is lossless
// 2. All data types are correctly persisted and retrieved
// 3. Edge cases (empty strings, special characters, binary data) are handled
// 4. Round-trip operations preserve data integrity
func TestProperty_DatabasePersistenceRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Test Installation round-trip
	properties.Property("Installation round-trip preserves data", prop.ForAll(
		func(original repository.Installation) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create repository
			repo, err := sqlite.NewInstallationRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create repository: %v", err)
				return false
			}
			defer repo.Close()

			// Store the installation
			err = repo.Create(ctx, &original)
			if err != nil {
				t.Logf("Failed to create installation: %v", err)
				return false
			}

			// Retrieve the installation
			retrieved, err := repo.FindByToolAndVersion(ctx, original.Tool, original.Version)
			if err != nil {
				t.Logf("Failed to retrieve installation: %v", err)
				return false
			}

			// Verify deep equality (excluding auto-generated fields)
			if !installationsEqual(original, *retrieved) {
				t.Logf("Installations not equal after round-trip")
				t.Logf("Original: %+v", original)
				t.Logf("Retrieved: %+v", retrieved)
				return false
			}

			return true
		},
		genInstallation(),
	))

	// Test CacheEntry round-trip
	properties.Property("CacheEntry round-trip preserves data", prop.ForAll(
		func(key string, value []byte, ttlSeconds int) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create repository
			repo, err := sqlite.NewCacheRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create repository: %v", err)
				return false
			}
			defer repo.Close()

			// Store the cache entry
			ttl := time.Duration(ttlSeconds) * time.Second
			err = repo.Set(ctx, key, value, ttl)
			if err != nil {
				t.Logf("Failed to set cache entry: %v", err)
				return false
			}

			// Retrieve the cache entry
			retrieved, err := repo.Get(ctx, key)
			if err != nil {
				t.Logf("Failed to get cache entry: %v", err)
				return false
			}

			// Verify data equality
			if !bytesEqual(value, retrieved) {
				t.Logf("Cache values not equal after round-trip")
				t.Logf("Original length: %d", len(value))
				t.Logf("Retrieved length: %d", len(retrieved))
				return false
			}

			return true
		},
		genCacheKey(),
		genCacheValue(),
		gen.IntRange(60, 86400), // TTL between 1 minute and 1 day
	))

	// Test AuditEntry round-trip
	properties.Property("AuditEntry round-trip preserves data", prop.ForAll(
		func(original repository.AuditEntry) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create repository
			repo, err := sqlite.NewAuditRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create repository: %v", err)
				return false
			}
			defer repo.Close()

			// Store the audit entry
			err = repo.Log(ctx, &original)
			if err != nil {
				t.Logf("Failed to log audit entry: %v", err)
				return false
			}

			// Retrieve the audit entry
			entries, err := repo.Query(ctx, repository.AuditFilter{
				Operation: original.Operation,
				Tool:      original.Tool,
				Status:    original.Status,
			})
			if err != nil {
				t.Logf("Failed to query audit entries: %v", err)
				return false
			}

			if len(entries) == 0 {
				t.Logf("No audit entries found after insert")
				return false
			}

			// Find the matching entry (should be the first one)
			retrieved := entries[0]

			// Verify deep equality (excluding auto-generated fields)
			if !auditEntriesEqual(original, *retrieved) {
				t.Logf("Audit entries not equal after round-trip")
				t.Logf("Original: %+v", original)
				t.Logf("Retrieved: %+v", retrieved)
				return false
			}

			return true
		},
		genAuditEntry(),
	))

	// Test IndexEntry round-trip
	properties.Property("IndexEntry round-trip preserves data", prop.ForAll(
		func(original repository.IndexEntry) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create repository
			repo, err := sqlite.NewIndexRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create repository: %v", err)
				return false
			}
			defer repo.Close()

			// Store the index entry
			err = repo.Upsert(ctx, &original)
			if err != nil {
				t.Logf("Failed to upsert index entry: %v", err)
				return false
			}

			// Retrieve the index entry
			retrieved, err := repo.FindByTool(ctx, original.Tool)
			if err != nil {
				t.Logf("Failed to retrieve index entry: %v", err)
				return false
			}

			// Verify deep equality (excluding auto-generated fields)
			if !indexEntriesEqual(original, *retrieved) {
				t.Logf("Index entries not equal after round-trip")
				t.Logf("Original: %+v", original)
				t.Logf("Retrieved: %+v", retrieved)
				return false
			}

			return true
		},
		genIndexEntry(),
	))

	properties.TestingRun(t)
}

// setupTempDB creates a temporary database for testing
func setupTempDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "unirtm-property-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := database.Open(context.Background(), database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

// genInstallation generates random Installation objects for property-based testing
func genInstallation() gopter.Gen {
	return gopter.CombineGens(
		genToolNameDB(),
		genVersionDB(),
		genBackend(),
		genProvider(),
		genInstallPath(),
		genChecksum(),
		genMetadata(),
	).Map(func(values []interface{}) repository.Installation {
		return repository.Installation{
			Tool:        values[0].(string),
			Version:     values[1].(string),
			Backend:     values[2].(string),
			Provider:    values[3].(string),
			InstallPath: values[4].(string),
			Checksum:    values[5].(string),
			Metadata:    values[6].(string),
		}
	})
}

// genAuditEntry generates random AuditEntry objects for property-based testing
func genAuditEntry() gopter.Gen {
	return gopter.CombineGens(
		genOperation(),
		genToolNameDB(),
		genVersionDB(),
		genStatus(),
		genErrorMessage(),
		gen.Int64Range(0, 60000), // Duration 0-60 seconds in milliseconds
		genMetadata(),
	).Map(func(values []interface{}) repository.AuditEntry {
		return repository.AuditEntry{
			Operation: values[0].(string),
			Tool:      values[1].(string),
			Version:   values[2].(string),
			Status:    values[3].(string),
			Error:     values[4].(string),
			Duration:  values[5].(int64),
			Metadata:  values[6].(string),
		}
	})
}

// genIndexEntry generates random IndexEntry objects for property-based testing
func genIndexEntry() gopter.Gen {
	return gopter.CombineGens(
		genToolNameDB(),
		genDescription(),
		genHomepage(),
		genLicense(),
		genBackend(),
		genMetadata(),
	).Map(func(values []interface{}) repository.IndexEntry {
		return repository.IndexEntry{
			Tool:        values[0].(string),
			Description: values[1].(string),
			Homepage:    values[2].(string),
			License:     values[3].(string),
			Backend:     values[4].(string),
			Metadata:    values[5].(string),
		}
	})
}

// Generator helpers

func genToolNameDB() gopter.Gen {
	return gen.OneConstOf(
		"node",
		"python",
		"go",
		"ruby",
		"rust",
		"java",
		"terraform",
		"kubectl",
		"docker",
		"helm",
	)
}

func genVersionDB() gopter.Gen {
	return gen.OneConstOf(
		"1.0.0",
		"2.3.4",
		"20.0.0",
		"3.11.5",
		"latest",
		"lts",
		"stable",
		"1.20.0",
		"3.11",
		"2.7.0",
	)
}

func genBackend() gopter.Gen {
	return gen.OneConstOf("github", "aqua", "http", "custom")
}

func genProvider() gopter.Gen {
	return gen.OneConstOf("generic", "node", "python", "go", "custom")
}

func genInstallPath() gopter.Gen {
	return gen.OneConstOf(
		"/usr/local/unirtm/installs/node/20.0.0",
		"/home/user/.unirtm/installs/python/3.11.5",
		"/opt/unirtm/go/1.20.0",
		"C:\\Program Files\\unirtm\\ruby\\3.2.0",
		"/tmp/unirtm/test/1.0.0",
	)
}

func genChecksum() gopter.Gen {
	return gen.OneConstOf(
		"abc123def456",
		"sha256:1234567890abcdef",
		"",
		"0000000000000000",
		"ffffffffffffffff",
	)
}

func genMetadata() gopter.Gen {
	return gen.OneConstOf(
		`{}`,
		`{"arch":"x64","os":"linux"}`,
		`{"version":"1.0.0","build":"release"}`,
		`{"tags":["stable","lts"]}`,
		``,
	)
}

func genCacheKey() gopter.Gen {
	return gen.OneConstOf(
		"node:20.0.0:metadata",
		"python:3.11:versions",
		"github:releases:terraform",
		"cache:test:key",
		"",
		"key-with-special-chars:@#$%",
	)
}

func genCacheValue() gopter.Gen {
	return gen.OneConstOf(
		[]byte(`{"size":50000000}`),
		[]byte(`{"versions":["1.0.0","2.0.0"]}`),
		[]byte{},
		[]byte{0x00, 0x01, 0x02, 0xFF}, // Binary data
		[]byte("plain text value"),
	)
}

func genOperation() gopter.Gen {
	return gen.OneConstOf("install", "uninstall", "activate", "update", "configure")
}

func genStatus() gopter.Gen {
	return gen.OneConstOf("success", "failure", "started", "pending")
}

func genErrorMessage() gopter.Gen {
	return gen.OneConstOf(
		"",
		"network timeout",
		"checksum mismatch",
		"file not found",
		"permission denied",
	)
}

func genDescription() gopter.Gen {
	return gen.OneConstOf(
		"Node.js JavaScript runtime",
		"Python programming language",
		"Go programming language",
		"",
		"Tool with special chars: @#$%",
	)
}

func genHomepage() gopter.Gen {
	return gen.OneConstOf(
		"https://nodejs.org",
		"https://python.org",
		"https://golang.org",
		"",
		"https://example.com/tool",
	)
}

func genLicense() gopter.Gen {
	return gen.OneConstOf("MIT", "Apache-2.0", "GPL-3.0", "BSD-3-Clause", "")
}

// Equality comparison helpers

func installationsEqual(a, b repository.Installation) bool {
	// Compare all fields except ID and InstalledAt (auto-generated)
	return a.Tool == b.Tool &&
		a.Version == b.Version &&
		a.Backend == b.Backend &&
		a.Provider == b.Provider &&
		a.InstallPath == b.InstallPath &&
		a.Checksum == b.Checksum &&
		a.Metadata == b.Metadata
}

func auditEntriesEqual(a, b repository.AuditEntry) bool {
	// Compare all fields except ID and Timestamp (auto-generated)
	return a.Operation == b.Operation &&
		a.Tool == b.Tool &&
		a.Version == b.Version &&
		a.Status == b.Status &&
		a.Error == b.Error &&
		a.Duration == b.Duration &&
		a.Metadata == b.Metadata
}

func indexEntriesEqual(a, b repository.IndexEntry) bool {
	// Compare all fields except UpdatedAt (auto-generated)
	return a.Tool == b.Tool &&
		a.Description == b.Description &&
		a.Homepage == b.Homepage &&
		a.License == b.License &&
		a.Backend == b.Backend &&
		a.Metadata == b.Metadata
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Feature: unirtm, Property 9: Concurrent Database Reads
//
// **Validates: Requirements 2.7**
//
// For any set of concurrent read operations on the database, all reads SHALL
// complete successfully without conflicts or data corruption.
//
// This property ensures that:
// 1. Multiple goroutines can read from the database simultaneously
// 2. No data races occur during concurrent reads
// 3. All reads return consistent data
// 4. The database remains stable under concurrent load
func TestProperty_ConcurrentDatabaseReads(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Concurrent reads complete successfully without conflicts", prop.ForAll(
		func(numReaders int, installations []repository.Installation) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create repository
			repo, err := sqlite.NewInstallationRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create repository: %v", err)
				return false
			}
			defer repo.Close()

			// Populate database with test data
			for i := range installations {
				err := repo.Create(ctx, &installations[i])
				if err != nil {
					t.Logf("Failed to create installation: %v", err)
					return false
				}
			}

			// Launch concurrent readers using errgroup
			g := new(errgroup.Group)
			results := make([][]repository.Installation, numReaders)

			for i := 0; i < numReaders; i++ {
				readerIndex := i
				g.Go(func() error {
					// Each reader lists all installations
					list, err := repo.List(ctx)
					if err != nil {
						return fmt.Errorf("reader %d failed: %w", readerIndex, err)
					}

					// Store results for consistency check
					results[readerIndex] = make([]repository.Installation, len(list))
					for j, inst := range list {
						results[readerIndex][j] = *inst
					}

					return nil
				})
			}

			// Wait for all readers to complete
			if err := g.Wait(); err != nil {
				t.Logf("Concurrent reads failed: %v", err)
				return false
			}

			// Verify all readers got consistent results
			if len(results) > 0 {
				expectedCount := len(results[0])
				for i := 1; i < len(results); i++ {
					if len(results[i]) != expectedCount {
						t.Logf("Inconsistent read results: reader 0 got %d items, reader %d got %d items",
							expectedCount, i, len(results[i]))
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(5, 20),        // 5-20 concurrent readers
		genInstallationList(1, 10), // 1-10 installations
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 10: Transaction Atomicity
//
// **Validates: Requirements 2.8, 2.9**
//
// For any database transaction that fails, all changes within that transaction
// SHALL be rolled back and the database SHALL contain no partial changes.
//
// This property ensures that:
// 1. Failed transactions leave no partial state
// 2. Database state is unchanged after rollback
// 3. Successful transactions commit all changes atomically
// 4. Transaction boundaries are properly enforced
func TestProperty_TransactionAtomicity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Failed transactions rollback completely", prop.ForAll(
		func(installations []repository.Installation, failurePoint int) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create transaction manager
			txMgr := transaction.NewSQLiteTransactionManager(db.Conn())

			// Record initial state
			initialRepo, err := sqlite.NewInstallationRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create initial repository: %v", err)
				return false
			}
			defer initialRepo.Close()

			initialList, err := initialRepo.List(ctx)
			if err != nil {
				t.Logf("Failed to list initial installations: %v", err)
				return false
			}
			initialCount := len(initialList)

			// Begin transaction
			tx, err := txMgr.Begin(ctx)
			if err != nil {
				t.Logf("Failed to begin transaction: %v", err)
				return false
			}

			// Perform operations within transaction
			txRepo := tx.InstallationRepo()
			var txErr error

			for i, inst := range installations {
				// Inject failure at the specified point
				if i == failurePoint {
					txErr = fmt.Errorf("injected failure at operation %d", i)
					break
				}

				err := txRepo.Create(ctx, &inst)
				if err != nil {
					txErr = err
					break
				}
			}

			// Rollback if there was an error
			if txErr != nil {
				if err := tx.Rollback(); err != nil {
					t.Logf("Failed to rollback transaction: %v", err)
					return false
				}
			} else {
				// Commit if all operations succeeded
				if err := tx.Commit(); err != nil {
					t.Logf("Failed to commit transaction: %v", err)
					return false
				}
			}

			// Verify final state
			finalRepo, err := sqlite.NewInstallationRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create final repository: %v", err)
				return false
			}
			defer finalRepo.Close()

			finalList, err := finalRepo.List(ctx)
			if err != nil {
				t.Logf("Failed to list final installations: %v", err)
				return false
			}
			finalCount := len(finalList)

			// If transaction failed, count should be unchanged
			if txErr != nil {
				if finalCount != initialCount {
					t.Logf("Transaction rollback failed: initial count=%d, final count=%d",
						initialCount, finalCount)
					return false
				}
			} else {
				// If transaction succeeded, all items should be present
				expectedCount := initialCount + len(installations)
				if finalCount != expectedCount {
					t.Logf("Transaction commit incomplete: expected count=%d, final count=%d",
						expectedCount, finalCount)
					return false
				}
			}

			return true
		},
		genInstallationList(1, 10), // 1-10 operations
		gen.IntRange(0, 10),        // Failure point (0-10)
	))

	properties.Property("Successful transactions commit all changes atomically", prop.ForAll(
		func(installations []repository.Installation) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create transaction manager
			txMgr := transaction.NewSQLiteTransactionManager(db.Conn())

			// Record initial state
			initialRepo, err := sqlite.NewInstallationRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create initial repository: %v", err)
				return false
			}
			defer initialRepo.Close()

			initialList, err := initialRepo.List(ctx)
			if err != nil {
				t.Logf("Failed to list initial installations: %v", err)
				return false
			}
			initialCount := len(initialList)

			// Begin transaction
			tx, err := txMgr.Begin(ctx)
			if err != nil {
				t.Logf("Failed to begin transaction: %v", err)
				return false
			}

			// Perform all operations within transaction
			txRepo := tx.InstallationRepo()
			for i := range installations {
				err := txRepo.Create(ctx, &installations[i])
				if err != nil {
					t.Logf("Failed to create installation in transaction: %v", err)
					tx.Rollback()
					return false
				}
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				t.Logf("Failed to commit transaction: %v", err)
				return false
			}

			// Verify all changes are present
			finalRepo, err := sqlite.NewInstallationRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create final repository: %v", err)
				return false
			}
			defer finalRepo.Close()

			finalList, err := finalRepo.List(ctx)
			if err != nil {
				t.Logf("Failed to list final installations: %v", err)
				return false
			}

			expectedCount := initialCount + len(installations)
			if len(finalList) != expectedCount {
				t.Logf("Atomic commit failed: expected %d items, got %d",
					expectedCount, len(finalList))
				return false
			}

			return true
		},
		genInstallationList(1, 10), // 1-10 operations
	))

	properties.TestingRun(t)
}

// genInstallationList generates a list of random Installation objects
// with unique tool+version combinations to avoid constraint violations
func genInstallationList(minSize, maxSize int) gopter.Gen {
	return gen.IntRange(minSize, maxSize).FlatMap(func(size interface{}) gopter.Gen {
		n := size.(int)
		return gopter.CombineGens(
			gen.SliceOfN(n, genInstallation()),
		).Map(func(values []interface{}) []repository.Installation {
			installations := values[0].([]repository.Installation)

			// Make tool+version combinations unique by appending index
			seen := make(map[string]bool)
			for i := range installations {
				key := installations[i].Tool + ":" + installations[i].Version
				counter := 0
				originalKey := key

				// If duplicate, append counter until unique
				for seen[key] {
					counter++
					key = fmt.Sprintf("%s-%d", originalKey, counter)
					// Update version to make it unique
					installations[i].Version = fmt.Sprintf("%s-%d", installations[i].Version, counter)
				}
				seen[key] = true
			}

			return installations
		})
	}, reflect.TypeOf([]repository.Installation{}))
}
