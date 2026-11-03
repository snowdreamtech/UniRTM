// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package errors_test

import (
	"fmt"
	"os"

	"github.com/snowdreamtech/unirtm/internal/pkg/errors"
)

// ExampleNewUserError demonstrates creating a user error.
func ExampleNewUserError() {
	err := errors.NewUserError("invalid version format", errors.ErrInvalidConfig)
	fmt.Println(err)
	// Output: [user] invalid version format: invalid configuration
}

// ExampleNewSystemError demonstrates creating a system error.
func ExampleNewSystemError() {
	err := errors.NewSystemError("database connection failed", errors.ErrTransactionFailed)
	fmt.Println(err)
	// Output: [system] database connection failed: transaction failed
}

// ExampleNewExternalError demonstrates creating an external error.
func ExampleNewExternalError() {
	err := errors.NewExternalError("API request failed", errors.ErrNetworkFailure)
	fmt.Println(err)
	// Output: [external] API request failed: network failure
}

// ExampleWrap demonstrates wrapping an error with context.
func ExampleWrap() {
	baseErr := errors.ErrNotFound
	wrappedErr := errors.Wrap(baseErr, "find user %d", 123)
	fmt.Println(wrappedErr)
	// Output: find user 123: not found
}

// ExampleGetCategory demonstrates getting the category of an error.
func ExampleGetCategory() {
	err := errors.NewUserError("test", nil)
	category := errors.GetCategory(err)
	fmt.Println(category)
	// Output: user
}

// ExampleIsUserError demonstrates checking if an error is a user error.
func ExampleIsUserError() {
	err := errors.NewUserError("test", nil)
	fmt.Println(errors.IsUserError(err))
	// Output: true
}

// ExampleExitCode demonstrates getting the exit code for an error.
func ExampleExitCode() {
	userErr := errors.NewUserError("test", nil)
	systemErr := errors.NewSystemError("test", nil)
	externalErr := errors.NewExternalError("test", nil)

	fmt.Println("User error exit code:", errors.ExitCode(userErr))
	fmt.Println("System error exit code:", errors.ExitCode(systemErr))
	fmt.Println("External error exit code:", errors.ExitCode(externalErr))
	// Output:
	// User error exit code: 1
	// System error exit code: 2
	// External error exit code: 3
}

// Example_errorHandlingPattern demonstrates a complete error handling pattern.
func Example_errorHandlingPattern() {
	// Simulate a function that might fail
	findUser := func(id int) error {
		// Simulate not found
		return errors.ErrNotFound
	}

	// Call the function
	err := findUser(123)
	if err != nil {
		// Wrap with context
		err = errors.Wrap(err, "find user %d", 123)

		// Categorize as user error
		err = errors.NewUserError("user lookup failed", err)

		// Check category and handle appropriately
		if errors.IsUserError(err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			// In real code: os.Exit(errors.ExitCode(err))
		}
	}
	// Output:
}

// Example_multipleWrapping demonstrates wrapping errors multiple times.
func Example_multipleWrapping() {
	// Start with a base error
	baseErr := errors.ErrNetworkFailure

	// Wrap at different layers
	layer1 := errors.Wrap(baseErr, "download artifact")
	layer2 := errors.Wrap(layer1, "install tool node")
	layer3 := errors.NewExternalError("installation failed", layer2)

	fmt.Println(layer3)
	// Output: [external] installation failed: install tool node: download artifact: network failure
}
