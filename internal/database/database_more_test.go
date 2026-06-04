// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabase_Open_MkdirError(t *testing.T) {
	ctx := context.Background()
	// An invalid path that cannot be created, e.g. /dev/null/test.db
	invalidPath := filepath.Join("/dev/null", "test.db")

	_, err := Open(ctx, Config{
		Path:    invalidPath,
		WALMode: true,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "create database directory")
}

func TestDatabase_Open_EnableWALModeError(t *testing.T) {
	// A canceled context will cause enableWALMode (ExecContext) to fail
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "enable WAL mode")
}

func TestDatabase_Close_NilConn(t *testing.T) {
	db := &DB{conn: nil}
	err := db.Close()
	require.NoError(t, err)
}

func TestDatabase_Path(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(ctx, Config{Path: dbPath, WALMode: false})
	require.NoError(t, err)
	defer db.Close()
	require.Equal(t, dbPath, db.Path())
}
