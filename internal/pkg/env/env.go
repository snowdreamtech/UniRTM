// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"sync"
)

// EnvManager provides smart cached environment variable operations,
// featuring automatic change detection, case-insensitivity, UNIRTM_/MISE_ prefixes,
// and real-time cache updating when variables are added, updated, or removed.
type EnvManager struct {
	mu            sync.RWMutex
	fingerprint   uint64
	unirtmIndex   map[string]string // normalizedKey -> value
	miseIndex     map[string]string // normalizedKey -> value
	rawIndex      map[string]string // normalizedKey -> value
	resolvedCache map[string]string // queryKey -> value
}

var defaultEnvManager = &EnvManager{
	unirtmIndex:   make(map[string]string),
	miseIndex:     make(map[string]string),
	rawIndex:      make(map[string]string),
	resolvedCache: make(map[string]string),
}

// normalizeKey strips underscores and dashes and returns lowercase string.
func normalizeKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
}

// computeEnvironFingerprint computes a 64-bit FNV-1a hash over os.Environ()
// to detect when any environment variable is added, modified, or removed.
func computeEnvironFingerprint() uint64 {
	envList := os.Environ()
	var h uint64 = 14695981039346656037
	for _, s := range envList {
		for i := 0; i < len(s); i++ {
			h ^= uint64(s[i])
			h *= 1099511628211
		}
	}
	return h
}

func (m *EnvManager) rebuildIndexLocked(currentFp uint64) {
	m.fingerprint = currentFp
	m.unirtmIndex = make(map[string]string)
	m.miseIndex = make(map[string]string)
	m.rawIndex = make(map[string]string)
	m.resolvedCache = make(map[string]string)

	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			continue
		}
		rawKey := parts[0]
		val := strings.TrimSpace(parts[1])
		if val == "" {
			continue
		}

		clean := normalizeKey(rawKey)
		if strings.HasPrefix(clean, "unirtm") {
			stripped := strings.TrimPrefix(clean, "unirtm")
			if _, exists := m.unirtmIndex[stripped]; !exists {
				m.unirtmIndex[stripped] = val
			}
		} else if strings.HasPrefix(clean, "mise") {
			stripped := strings.TrimPrefix(clean, "mise")
			if _, exists := m.miseIndex[stripped]; !exists {
				m.miseIndex[stripped] = val
			}
		} else {
			if _, exists := m.rawIndex[clean]; !exists {
				m.rawIndex[clean] = val
			}
		}
	}
}

// Get returns the value of the environment variable with the given key,
// searching with prefixes in order: UNIRTM_, MISE_, and then the raw key.
// It uses a smart fingerprint-backed cache that automatically refreshes
// when any environment variable is added, modified, or removed.
func (m *EnvManager) Get(key string) string {
	if key == "PATH" {
		return os.Getenv("PATH")
	}

	fp := computeEnvironFingerprint()

	m.mu.RLock()
	if m.fingerprint == fp && m.fingerprint != 0 {
		if val, ok := m.resolvedCache[key]; ok {
			m.mu.RUnlock()
			return validateCriticalKey(key, val)
		}
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Recheck under write lock
	if m.fingerprint != fp || m.fingerprint == 0 {
		m.rebuildIndexLocked(fp)
	} else if val, ok := m.resolvedCache[key]; ok {
		return validateCriticalKey(key, val)
	}

	norm := normalizeKey(key)
	lookupKey := norm
	if strings.HasPrefix(norm, "unirtm") {
		lookupKey = strings.TrimPrefix(norm, "unirtm")
	} else if strings.HasPrefix(norm, "mise") {
		lookupKey = strings.TrimPrefix(norm, "mise")
	}

	var res string
	// Priority 1: UNIRTM_ prefix
	if v, ok := m.unirtmIndex[lookupKey]; ok && v != "" {
		res = v
	} else if v, ok := m.miseIndex[lookupKey]; ok && v != "" { // Priority 2: MISE_ prefix
		res = v
	} else if v, ok := m.rawIndex[lookupKey]; ok && v != "" { // Priority 3: Raw key
		res = v
	}

	m.resolvedCache[key] = res
	return validateCriticalKey(key, res)
}

// Set sets the environment variable and immediately updates the smart cache.
func (m *EnvManager) Set(key, value string) error {
	if err := os.Setenv(key, value); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rebuildIndexLocked(computeEnvironFingerprint())
	return nil
}

// Unset unsets the environment variable (and related UNIRTM_/MISE_ variants) and updates the smart cache.
func (m *EnvManager) Unset(key string) error {
	norm := normalizeKey(key)

	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		clean := normalizeKey(parts[0])
		strippedUnirtm := strings.TrimPrefix(clean, "unirtm")
		strippedMise := strings.TrimPrefix(clean, "mise")
		if clean == norm || strippedUnirtm == norm || strippedMise == norm {
			_ = os.Unsetenv(parts[0])
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.rebuildIndexLocked(computeEnvironFingerprint())
	return nil
}

// Has returns true if the given environment variable exists in any valid prefix or case form.
func (m *EnvManager) Has(key string) bool {
	return m.Get(key) != ""
}

// ResetCache clears the smart cache and forces re-indexing on next access.
func (m *EnvManager) ResetCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fingerprint = 0
	m.rebuildIndexLocked(computeEnvironFingerprint())
}

// Get returns the value of the environment variable with the given key using the smart cache.
func Get(key string) string {
	return defaultEnvManager.Get(key)
}

// Set sets an environment variable and updates the smart cache in real time.
func Set(key, value string) error {
	return defaultEnvManager.Set(key, value)
}

// Unset unsets an environment variable and updates the smart cache in real time.
func Unset(key string) error {
	return defaultEnvManager.Unset(key)
}

// Has checks if an environment variable exists with smart case and prefix resolution.
func Has(key string) bool {
	return defaultEnvManager.Has(key)
}

// ResetCache explicitly invalidates the smart environment cache.
func ResetCache() {
	defaultEnvManager.ResetCache()
}

// InvalidateCache is an alias for ResetCache.
func InvalidateCache() {
	defaultEnvManager.ResetCache()
}

func validateCriticalKey(key, value string) string {
	switch key {
	case "GITHUB_PROXY":
		if value != "" && value != "direct" {
			if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
				return ""
			}
		}
	case "JOBS":
		if value != "" {
			var n int
			if _, err := fmt.Sscanf(value, "%d", &n); err != nil || n < 1 || n > 256 {
				return ""
			}
		}
	case "HTTP2":
		if value != "" && value != "0" && value != "1" {
			return ""
		}
	}
	return value
}

var (
	//ProjectName Project Name
	ProjectName string = "unirtm"

	//Author Author
	Author string = "Snowdream Tech <snowdreamtech@qq.com>"

	//BuildTime Build Time
	BuildTime string = "N/A"

	//GitTag Git Tag
	GitTag string = "N/A"

	//CommitHash Commit Hash
	CommitHash string = "N/A"

	//CommitHashFull Commit Hash
	CommitHashFull string = "N/A"

	//COPYRIGHT COPYRIGHT
	COPYRIGHT string = "Copyright (c) 2023-present SnowdreamTech Inc."

	//LICENSE LICENSE
	LICENSE string = "MIT <https://github.com/snowdreamtech/unirtm/blob/main/LICENSE>"

	//Config Config File Path
	Config string = "unirtm.toml"

	// Debug indicates whether the application should run in debug mode.
	Debug bool

	// Trace indicates whether the application should run in trace mode.
	Trace bool

	// Quiet indicates whether the application should run in quiet mode.
	Quiet bool

	// Cwd specifies the current working directory for the application.
	Cwd string

	// EnvName specifies the environment name for loading environment-specific configs.
	EnvName string

	// Jobs specifies the number of parallel jobs to run.
	Jobs int

	// Yes indicates whether to automatically answer yes to all confirmation prompts.
	Yes bool

	// Locked indicates whether to require lockfile URLs to be present during installation.
	Locked bool

	// Silent indicates whether to suppress all output and non-error messages.
	Silent bool

	CryptoRandRead = rand.Read
)

// RandomString returns a random string of the specified length.
func RandomString(n int) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := make([]byte, n)
	if _, err := CryptoRandRead(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
