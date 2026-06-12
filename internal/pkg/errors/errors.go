// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
// These errors can be used with errors.Is() for error checking.
var (
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a resource already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidConfig indicates invalid configuration.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrNetworkFailure indicates a network operation failed.
	ErrNetworkFailure = errors.New("network failure")

	// ErrChecksumMismatch indicates checksum verification failed.
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrTransactionFailed indicates a transaction failed.
	ErrTransactionFailed = errors.New("transaction failed")
)

// ErrorCategory represents the classification of an error.
type ErrorCategory int

const (
	// CategoryUnknown represents an unclassified error.
	CategoryUnknown ErrorCategory = iota

	// CategoryUser represents user errors (invalid input, configuration errors, version not found).
	// These errors should return descriptive messages and suggest corrective actions.
	// Exit code: 1
	CategoryUser

	// CategorySystem represents system errors (disk full, permission denied, database corruption).
	// These errors should log full context and return generic user-safe messages.
	// Exit code: 2
	CategorySystem

	// CategoryExternal represents external errors (network failures, backend API errors, download failures).
	// These errors should implement retry logic and return wrapped errors with context.
	// Exit code: 3
	CategoryExternal
)

// String returns the string representation of the error category.
func (c ErrorCategory) String() string {
	switch c {
	case CategoryUser:
		return "user"
	case CategorySystem:
		return "system"
	case CategoryExternal:
		return "external"
	default:
		return "unknown"
	}
}

// CategorizedError wraps an error with a category classification.
type CategorizedError struct {
	err      error
	category ErrorCategory
	context  string
}

// Error implements the error interface.
func (e *CategorizedError) Error() string {
	if e.context != "" {
		return fmt.Sprintf("[%s] %s: %v", e.category, e.context, e.err)
	}
	return fmt.Sprintf("[%s] %v", e.category, e.err)
}

// Unwrap returns the wrapped error for errors.Is() and errors.As() support.
func (e *CategorizedError) Unwrap() error {
	return e.err
}

// Category returns the error category.
func (e *CategorizedError) Category() ErrorCategory {
	return e.category
}

// Context returns the error context.
func (e *CategorizedError) Context() string {
	return e.context
}

// NewUserError creates a new user error with the given message and optional wrapped error.
// User errors indicate invalid input, configuration errors, or version not found.
func NewUserError(message string, err error) error {
	if err == nil {
		return &CategorizedError{
			err:      errors.New(message),
			category: CategoryUser,
		}
	}
	return &CategorizedError{
		err:      fmt.Errorf("%s: %w", message, err),
		category: CategoryUser,
	}
}

// NewSystemError creates a new system error with the given message and optional wrapped error.
// System errors indicate disk full, permission denied, or database corruption.
func NewSystemError(message string, err error) error {
	if err == nil {
		return &CategorizedError{
			err:      errors.New(message),
			category: CategorySystem,
		}
	}
	return &CategorizedError{
		err:      fmt.Errorf("%s: %w", message, err),
		category: CategorySystem,
	}
}

// NewExternalError creates a new external error with the given message and optional wrapped error.
// External errors indicate network failures, backend API errors, or download failures.
func NewExternalError(message string, err error) error {
	if err == nil {
		return &CategorizedError{
			err:      errors.New(message),
			category: CategoryExternal,
		}
	}
	return &CategorizedError{
		err:      fmt.Errorf("%s: %w", message, err),
		category: CategoryExternal,
	}
}

// Wrap wraps an error with additional context using fmt.Errorf with %w.
// This preserves the error chain for errors.Is() and errors.As().
//
// Example:
//
//	user, err := repo.FindByID(ctx, id)
//	if err != nil {
//	    return nil, Wrap(err, "find user %d", id)
//	}
func Wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// GetCategory returns the category of an error.
// If the error is not a CategorizedError, it returns CategoryUnknown.
func GetCategory(err error) ErrorCategory {
	var catErr *CategorizedError
	if errors.As(err, &catErr) {
		return catErr.Category()
	}
	return CategoryUnknown
}

// IsUserError checks if an error is a user error.
func IsUserError(err error) bool {
	return GetCategory(err) == CategoryUser
}

// IsSystemError checks if an error is a system error.
func IsSystemError(err error) bool {
	return GetCategory(err) == CategorySystem
}

// IsExternalError checks if an error is an external error.
func IsExternalError(err error) bool {
	return GetCategory(err) == CategoryExternal
}

// ExitCode returns the appropriate exit code for an error based on its category.
// User errors: 1, System errors: 2, External errors: 3, Unknown: 1
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	category := GetCategory(err)
	switch category {
	case CategoryUser:
		return 1
	case CategorySystem:
		return 2
	case CategoryExternal:
		return 3
	default:
		return 1
	}
}

// Is reports whether any error in err's tree matches target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's tree that matches target, and if one is found, sets
// target to that error value and returns true.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}
