# Feature Specification: Linux libc Detection (musl vs glibc)

## 1. Overview

## 2. Motivation

By default, Go's `runtime.GOOS` returns `linux` for all Linux distributions. However, executing `glibc`-linked binaries on `musl`-based systems leads to runtime linker errors. Enhancing UniRTM with explicit libc detection allows for smarter binary distribution fetching and execution.

# . Requirements

* **Performance:** The detection must be lightweight (e.g., fast file system checks like `/etc/alpine-release` or `/lib/ld-musl*`).

* Cross-Platform Safety:** Must gracefully degrade or return `false` on non-Linux systems or standard `glibc` systems.

## 4. Proposed API

* `func IsMusl() bool` in the `env` package.
