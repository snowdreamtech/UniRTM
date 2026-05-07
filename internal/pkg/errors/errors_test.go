// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrAlreadyExists", ErrAlreadyExists, "already exists"},
		{"ErrInvalidConfig", ErrInvalidConfig, "invalid configuration"},
		{"ErrNetworkFailure", ErrNetworkFailure, "network failure"},
		{"ErrChecksumMismatch", ErrChecksumMismatch, "checksum mismatch"},
		{"ErrTransactionFailed", ErrTransactionFailed, "transaction failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestErrorCategoryString(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		want     string
	}{
		{"CategoryUnknown", CategoryUnknown, "unknown"},
		{"CategoryUser", CategoryUser, "user"},
		{"CategorySystem", CategorySystem, "system"},
		{"CategoryExternal", CategoryExternal, "external"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.category.String())
		})
	}
}

func TestNewUserError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wrappedErr  error
		wantMessage string
		wantIs      error
	}{
		{
			name:        "without wrapped error",
			message:     "invalid email format",
			wrappedErr:  nil,
			wantMessage: "[user] invalid email format",
			wantIs:      nil,
		},
		{
			name:        "with wrapped error",
			message:     "validation failed",
			wrappedErr:  ErrInvalidConfig,
			wantMessage: "[user] validation failed: invalid configuration",
			wantIs:      ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewUserError(tt.message, tt.wrappedErr)
			require.Error(t, err)
			assert.Equal(t, tt.wantMessage, err.Error())
			assert.True(t, IsUserError(err))

			if tt.wantIs != nil {
				assert.True(t, errors.Is(err, tt.wantIs))
			}
		})
	}
}

func TestNewSystemError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wrappedErr  error
		wantMessage string
		wantIs      error
	}{
		{
			name:        "without wrapped error",
			message:     "disk full",
			wrappedErr:  nil,
			wantMessage: "[system] disk full",
			wantIs:      nil,
		},
		{
			name:        "with wrapped error",
			message:     "database operation failed",
			wrappedErr:  ErrTransactionFailed,
			wantMessage: "[system] database operation failed: transaction failed",
			wantIs:      ErrTransactionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSystemError(tt.message, tt.wrappedErr)
			require.Error(t, err)
			assert.Equal(t, tt.wantMessage, err.Error())
			assert.True(t, IsSystemError(err))

			if tt.wantIs != nil {
				assert.True(t, errors.Is(err, tt.wantIs))
			}
		})
	}
}

func TestNewExternalError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wrappedErr  error
		wantMessage string
		wantIs      error
	}{
		{
			name:        "without wrapped error",
			message:     "API request timeout",
			wrappedErr:  nil,
			wantMessage: "[external] API request timeout",
			wantIs:      nil,
		},
		{
			name:        "with wrapped error",
			message:     "download failed",
			wrappedErr:  ErrNetworkFailure,
			wantMessage: "[external] download failed: network failure",
			wantIs:      ErrNetworkFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewExternalError(tt.message, tt.wrappedErr)
			require.Error(t, err)
			assert.Equal(t, tt.wantMessage, err.Error())
			assert.True(t, IsExternalError(err))

			if tt.wantIs != nil {
				assert.True(t, errors.Is(err, tt.wantIs))
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		format      string
		args        []interface{}
		wantMessage string
		wantIs      error
	}{
		{
			name:        "wrap nil error",
			err:         nil,
			format:      "context",
			args:        nil,
			wantMessage: "",
			wantIs:      nil,
		},
		{
			name:        "wrap simple error",
			err:         ErrNotFound,
			format:      "find user %d",
			args:        []interface{}{123},
			wantMessage: "find user 123: not found",
			wantIs:      ErrNotFound,
		},
		{
			name:        "wrap categorized error",
			err:         NewUserError("invalid input", nil),
			format:      "process request",
			args:        nil,
			wantMessage: "process request: [user] invalid input",
			wantIs:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.err, tt.format, tt.args...)

			if tt.err == nil {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Equal(t, tt.wantMessage, err.Error())

			if tt.wantIs != nil {
				assert.True(t, errors.Is(err, tt.wantIs))
			}
		})
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorCategory
	}{
		{
			name: "user error",
			err:  NewUserError("test", nil),
			want: CategoryUser,
		},
		{
			name: "system error",
			err:  NewSystemError("test", nil),
			want: CategorySystem,
		},
		{
			name: "external error",
			err:  NewExternalError("test", nil),
			want: CategoryExternal,
		},
		{
			name: "uncategorized error",
			err:  errors.New("test"),
			want: CategoryUnknown,
		},
		{
			name: "wrapped categorized error",
			err:  fmt.Errorf("wrapped: %w", NewUserError("test", nil)),
			want: CategoryUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCategory(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsCategoryFunctions(t *testing.T) {
	userErr := NewUserError("test", nil)
	systemErr := NewSystemError("test", nil)
	externalErr := NewExternalError("test", nil)
	plainErr := errors.New("test")

	tests := []struct {
		name     string
		err      error
		isUser   bool
		isSystem bool
		isExt    bool
	}{
		{
			name:     "user error",
			err:      userErr,
			isUser:   true,
			isSystem: false,
			isExt:    false,
		},
		{
			name:     "system error",
			err:      systemErr,
			isUser:   false,
			isSystem: true,
			isExt:    false,
		},
		{
			name:     "external error",
			err:      externalErr,
			isUser:   false,
			isSystem: false,
			isExt:    true,
		},
		{
			name:     "plain error",
			err:      plainErr,
			isUser:   false,
			isSystem: false,
			isExt:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isUser, IsUserError(tt.err))
			assert.Equal(t, tt.isSystem, IsSystemError(tt.err))
			assert.Equal(t, tt.isExt, IsExternalError(tt.err))
		})
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "nil error",
			err:  nil,
			want: 0,
		},
		{
			name: "user error",
			err:  NewUserError("test", nil),
			want: 1,
		},
		{
			name: "system error",
			err:  NewSystemError("test", nil),
			want: 2,
		},
		{
			name: "external error",
			err:  NewExternalError("test", nil),
			want: 3,
		},
		{
			name: "uncategorized error",
			err:  errors.New("test"),
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCode(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorChaining(t *testing.T) {
	// Test that error wrapping preserves the error chain
	baseErr := ErrNotFound
	wrappedOnce := Wrap(baseErr, "layer 1")
	wrappedTwice := Wrap(wrappedOnce, "layer 2")
	wrappedThrice := Wrap(wrappedTwice, "layer 3")

	// Should be able to unwrap to the original error
	assert.True(t, errors.Is(wrappedThrice, ErrNotFound))
	assert.Equal(t, "layer 3: layer 2: layer 1: not found", wrappedThrice.Error())
}

func TestCategorizedErrorUnwrap(t *testing.T) {
	// Test that CategorizedError properly unwraps
	baseErr := ErrInvalidConfig
	catErr := NewUserError("validation failed", baseErr)

	// Should be able to check for the wrapped error
	assert.True(t, errors.Is(catErr, ErrInvalidConfig))

	// Should be able to extract the categorized error
	var extracted *CategorizedError
	assert.True(t, errors.As(catErr, &extracted))
	assert.Equal(t, CategoryUser, extracted.Category())
}

func TestCategorizedErrorContext(t *testing.T) {
	err := &CategorizedError{
		err:      errors.New("test error"),
		category: CategoryUser,
		context:  "test context",
	}

	assert.Equal(t, "test context", err.Context())
	assert.Equal(t, CategoryUser, err.Category())
	assert.Equal(t, "[user] test context: test error", err.Error())
}

func TestMultipleWrapping(t *testing.T) {
	// Test wrapping a categorized error multiple times
	baseErr := NewExternalError("API failed", ErrNetworkFailure)
	wrapped1 := Wrap(baseErr, "retry attempt 1")
	wrapped2 := Wrap(wrapped1, "retry attempt 2")
	wrapped3 := Wrap(wrapped2, "final failure")

	// Should preserve category through wrapping
	assert.True(t, IsExternalError(wrapped3))
	assert.True(t, errors.Is(wrapped3, ErrNetworkFailure))
	assert.Contains(t, wrapped3.Error(), "final failure")
	assert.Contains(t, wrapped3.Error(), "retry attempt 2")
	assert.Contains(t, wrapped3.Error(), "retry attempt 1")
	assert.Contains(t, wrapped3.Error(), "API failed")
}
