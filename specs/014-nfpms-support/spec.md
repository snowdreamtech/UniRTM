# Feature Specification: nfpms-support

**Feature Branch**: `[014-nfpms-support]`

**Created**: 2026-06-07

**Status**: Draft

**Input**: User description: "实施第一阶段，进行nfpms支持"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Linux Packages (Priority: P1)

As a Linux user or system administrator, I want to install UniRTM using standard Linux package managers (apt, yum/dnf, apk, pacman) so that I can easily deploy and manage it natively without manually downloading and extracting archives.

**Why this priority**: Essential for native Linux distribution support, heavily requested by enterprise users and sysadmins for automated deployments.

**Independent Test**: Can be independently tested by running GoReleaser locally or in CI and verifying that `.deb`, `.rpm`, `.apk`, and Arch Linux packages are generated correctly.

**Acceptance Scenarios**:

1. **Given** a new release is triggered, **When** GoReleaser runs, **Then** it generates `.deb`, `.rpm`, `.apk`, and `archlinux` packages.
2. **Given** a generated `.deb` package, **When** installed via `dpkg -i`, **Then** the `unirtm` binary is placed in `/usr/bin/unirtm`.

### Edge Cases

- What happens when a user installs via package manager and then tries to self-update? (Should be disabled or warn the user).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST configure GoReleaser NFPM integration.
- **FR-002**: System MUST generate `.deb` packages for Debian/Ubuntu environments.
- **FR-003**: System MUST generate `.rpm` packages for RHEL/CentOS/Fedora environments.
- **FR-004**: System MUST generate `.apk` packages for Alpine Linux environments.
- **FR-005**: System MUST generate packages for Arch Linux.
- **FR-006**: Generated packages MUST include proper metadata (description, maintainer, license, homepage).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: CI pipeline successfully generates 4 types of Linux packages (`deb`, `rpm`, `apk`, `archlinux`) upon release.
- **SC-002**: Packages can be successfully installed and uninstalled on their respective target systems without errors.

## Assumptions

- We assume standard GoReleaser NFPM defaults are sufficient for initial support without needing complex pre/post-install scripts.
