// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestSearchStructure(t *testing.T) {
	assert.Contains(t, searchCmd.Use, "search", "searchCmd command use should contain 'search'")
	assert.NotEmpty(t, searchCmd.Short, "searchCmd command short description should not be empty")
	assert.True(t, searchCmd.Run != nil || searchCmd.RunE != nil, "Run or RunE function should be set for searchCmd")
}

func setupSearchDB(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{Path: env.GetDatabasePath(), WALMode: true})
	assert.NoError(t, err)
	defer db.Close()
	repo, err := sqlite.NewIndexRepository(db.Conn())
	assert.NoError(t, err)
	err = repo.Upsert(ctx, &repository.IndexEntry{
		Tool: "dummy-tool", Description: "A long description that will be truncated",
		Homepage: "https://example.com", License: "MIT", Backend: "github",
	})
	assert.NoError(t, err)
	err = repo.Upsert(ctx, &repository.IndexEntry{
		Tool: "dummy-tool-aqua", Description: "Aqua tool",
		Homepage: "https://aqua.com", License: "", Backend: "aqua",
	})
	assert.NoError(t, err)
	err = repo.Upsert(ctx, &repository.IndexEntry{
		Tool: "dummy-tool-native", Backend: "native",
	})
	assert.NoError(t, err)
}

func TestRunSearch_NoResults(t *testing.T) {
	tmpData := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpData)

	// Suppress standard output
	quiet = true
	defer func() { quiet = false }()

	err := runSearch(searchCmd, []string{"nonexistent_tool"})
	assert.NoError(t, err)
}

func TestRunSearch_Results(t *testing.T) {
	tmpData := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpData)
	setupSearchDB(t)

	err := runSearch(searchCmd, []string{"dummy"})
	assert.NoError(t, err)

	searchBackend = "github"
	err = runSearch(searchCmd, []string{"dummy"})
	assert.NoError(t, err)
	searchBackend = "" // reset
}

func TestRunSearch_JsonOutput(t *testing.T) {
	tmpData := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpData)

	jsonOutput = true
	defer func() { jsonOutput = false }()

	err := runSearch(searchCmd, []string{"nonexistent_tool"})
	assert.NoError(t, err)
	
	setupSearchDB(t)
	err = runSearch(searchCmd, []string{"dummy"})
	assert.NoError(t, err)
}

