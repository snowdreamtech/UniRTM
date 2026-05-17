// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package gpg

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// NativeGPGVerifier implements Verifier using the pure-Go gopenpgp library.
type NativeGPGVerifier struct {
	client *http.Client
}

func NewNativeGPGVerifier() *NativeGPGVerifier {
	return &NativeGPGVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}
}

func (v *NativeGPGVerifier) IsAvailable(ctx context.Context) bool {
	// Always available as it's built-in
	return true
}

func (v *NativeGPGVerifier) Verify(ctx context.Context, sigPath, dataPath string, fingerprints []string) error {
	// Read signature
	sigData, err := os.ReadFile(sigPath)
	if err != nil {
		return fmt.Errorf("failed to read signature file: %v", err)
	}

	// Read data
	dataPipe, err := os.Open(dataPath)
	if err != nil {
		return fmt.Errorf("failed to open data file: %v", err)
	}
	defer dataPipe.Close()

	// Read all data once for verification
	data, err := io.ReadAll(dataPipe)
	if err != nil {
		return fmt.Errorf("failed to read data file: %v", err)
	}

	// We need the public key to verify.

	var lastErr error
	for _, fp := range fingerprints {
		keyObj, err := v.fetchKey(ctx, fp)
		if err != nil {
			lastErr = err
			continue
		}

		keyRing, err := crypto.NewKeyRing(keyObj)
		if err != nil {
			lastErr = err
			continue
		}

		// Create signature object
		signature := crypto.NewPGPSignature(sigData)
		if signature == nil {
			// Try armored
			var err error
			signature, err = crypto.NewPGPSignatureFromArmored(string(sigData))
			if err != nil {
				return fmt.Errorf("invalid signature format: %v", err)
			}
		}

		// Verify
		err = keyRing.VerifyDetached(crypto.NewPlainMessage(data), signature, crypto.GetUnixTime())
		if err == nil {
			// Success!
			return nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return fmt.Errorf("gpg verification failed: %v", lastErr)
	}

	return fmt.Errorf("gpg verification failed: no trusted keys matched")
}

func (v *NativeGPGVerifier) ImportKey(ctx context.Context, fingerprint string) error {
	_, err := v.fetchKey(ctx, fingerprint)
	return err
}

func (v *NativeGPGVerifier) fetchKey(ctx context.Context, fingerprint string) (*crypto.Key, error) {
	// Try multiple keyservers
	keyservers := []string{
		"https://keys.openpgp.org/vks/v1/by-fingerprint/%s",
		"https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x%s",
	}

	for _, template := range keyservers {
		url := fmt.Sprintf(template, strings.ToUpper(fingerprint))
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := v.client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		keyData, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		key, err := crypto.NewKeyFromArmored(string(keyData))
		if err == nil {
			return key, nil
		}
	}

	return nil, fmt.Errorf("failed to fetch public key for fingerprint %s from any keyserver", fingerprint)
}
