package main

import (
	"context"
	"fmt"
	"log"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/snowdreamtech/unirtm/internal/transaction"
)

func main() {
	ctx := context.Background()

	// Initialize registries
	backendRegistry := backend.NewRegistry()
	providerRegistry := provider.NewRegistry()
	downloadManager := download.NewManager()

	// Setup database
	db, err := database.Open(ctx, database.Config{
		Path: env.GetDatabasePath(),
	})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create repository
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		log.Fatalf("failed to create installation repository: %v", err)
	}

	// Create transaction manager
	txManager := transaction.NewSQLiteTransactionManager(db.Conn())

	im := service.NewInstallationManagerWithLock(
		backendRegistry,
		providerRegistry,
		downloadManager,
		installRepo,
		txManager,
		nil,
		nil,
	)

	// Test IsInstalled for node
	installed, inst := im.IsInstalled(ctx, "node", "26.1.0", "native")
	fmt.Printf("IsInstalled(node, 26.1.0, native) = %v\n", installed)
	if inst != nil {
		fmt.Printf("  Existing installation: %+v\n", inst)
	} else {
		fmt.Println("  No existing installation found.")
	}

	// Let's debug inside IsInstalled logic manually:
	version := "26.1.0"
	fsToolName := env.GetFSToolName("node", "native")
	fmt.Printf("fsToolName: %s\n", fsToolName)

	existing, err := installRepo.FindByToolAndVersion(ctx, "node", version)
	fmt.Printf("FindByToolAndVersion(node, %s) error: %v\n", version, err)
	if existing != nil {
		fmt.Printf("  Found in DB: %+v\n", existing)
	}
}
