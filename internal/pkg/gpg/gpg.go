// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package gpg

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Verifier defines the interface for GPG signature verification.
type Verifier interface {
	// Verify checks if the signature file corresponds to the data file using the provided fingerprint.
	Verify(ctx context.Context, sigPath, dataPath string, fingerprints []string) error
	// ImportKey imports a public key from a keyserver.
	ImportKey(ctx context.Context, fingerprint string) error
	// IsAvailable checks if the GPG tool is installed and usable.
	IsAvailable(ctx context.Context) bool
}

// NewVerifier returns the best available GPG verifier.
// It prioritizes the native implementation to avoid system dependencies.
func NewVerifier() Verifier {
	return NewNativeGPGVerifier()
}

// SystemGPGVerifier implements Verifier using the system 'gpg' command.
type SystemGPGVerifier struct{}

func NewSystemGPGVerifier() *SystemGPGVerifier {
	return &SystemGPGVerifier{}
}

func (v *SystemGPGVerifier) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("gpg")
	return err == nil
}

func (v *SystemGPGVerifier) Verify(ctx context.Context, sigPath, dataPath string, fingerprints []string) error {
	if !v.IsAvailable(ctx) {
		return fmt.Errorf("gpg command not found in PATH")
	}

	// Basic verification command
	// gpg --verify --status-fd 1 <sig> <data>
	// We use --status-fd 1 to get machine-readable status updates
	args := []string{"--verify", "--status-fd", "1"}
	args = append(args, sigPath, dataPath)

	cmd := exec.CommandContext(ctx, "gpg", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		if strings.Contains(outputStr, "NO_PUBKEY") || strings.Contains(outputStr, "Can't check signature") {
			return fmt.Errorf("gpg: missing public key for verification")
		}
		return fmt.Errorf("gpg verification failed: %v\nOutput: %s", err, outputStr)
	}

	// If fingerprints are provided, we MUST ensure the signing key matches one of them
	if len(fingerprints) > 0 {
		matched := false
		for _, fp := range fingerprints {
			// GPG status output for a good signature looks like:
			// [GNUPG:] VALIDSIG <fingerprint> ...
			if strings.Contains(outputStr, "VALIDSIG "+strings.ToUpper(fp)) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("gpg security violation: signature is valid but key does not match any trusted fingerprints")
		}
	}

	// Check for GOODSIG in output to be sure
	if !strings.Contains(outputStr, "GOODSIG") && !strings.Contains(outputStr, "VALIDSIG") {
		return fmt.Errorf("gpg verification failed: no valid signature found in output\nOutput: %s", outputStr)
	}

	return nil
}

// ImportKey imports a public key from a keyserver or file.
func (v *SystemGPGVerifier) ImportKey(ctx context.Context, keyIDOrFingerprint string) error {
	// Try fetching from common keyservers
	keyservers := []string{"keys.openpgp.org", "keyserver.ubuntu.com", "pgp.mit.edu"}

	var lastErr error
	for _, server := range keyservers {
		cmd := exec.CommandContext(ctx, "gpg", "--keyserver", server, "--recv-keys", keyIDOrFingerprint)
		if output, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			lastErr = fmt.Errorf("failed to fetch from %s: %v (output: %s)", server, err, string(output))
		}
	}

	return lastErr
}
