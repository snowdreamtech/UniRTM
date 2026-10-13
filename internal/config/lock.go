// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"time"
)

// FileLock represents an acquired lock on a configuration file.
type FileLock struct {
	lockPath string
	file     *os.File
}

// AcquireFileLock acquires an exclusive lock on targetPath + ".lock".
// It retries until timeout (defaults to 5s if timeout <= 0).
func AcquireFileLock(targetPath string, timeout time.Duration) (*FileLock, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	lockPath := targetPath + ".lock"
	deadline := time.Now().Add(timeout)

	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
		if err == nil {
			lockErr := lockFile(f)
			if lockErr == nil {
				return &FileLock{
					lockPath: lockPath,
					file:     f,
				}, nil
			}
			f.Close()
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout (%v) waiting to acquire configuration file lock on %s", timeout, lockPath)
		}

		time.Sleep(25 * time.Millisecond)
	}
}

// Release unlocks and removes the lock file.
func (l *FileLock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}
	_ = unlockFile(l.file)
	err := l.file.Close()
	_ = os.Remove(l.lockPath)
	l.file = nil
	return err
}
