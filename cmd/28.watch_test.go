// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatchStructure(t *testing.T) {
	assert.Contains(t, watchCmd.Use, "watch", "watchCmd command use should contain 'watch'")
	assert.NotEmpty(t, watchCmd.Short, "watchCmd command short description should not be empty")
	assert.True(t, watchCmd.Run != nil || watchCmd.RunE != nil, "Run or RunE function should be set for watchCmd")
}

func TestWatchIsMatched(t *testing.T) {
	// 1. empty globs matches anything unless ignored
	assert.True(t, isMatched("foo.txt", []string{}, []string{}))
	
	// 2. ignored
	assert.False(t, isMatched("foo.txt", []string{}, []string{"*.txt"}))
	
	// 3. glob matched
	assert.True(t, isMatched("foo.go", []string{"*.go"}, []string{}))
	
	// 4. glob not matched
	assert.False(t, isMatched("foo.txt", []string{"*.go"}, []string{}))
}

func TestWatchClearScreen(t *testing.T) {
	// Just executing for coverage
	clearScreen()
}

func TestWatchKillCurrentCmd(t *testing.T) {
	// Should not panic if cmd is nil
	killCurrentCmd()
}

