# Makefile for Snowdream Tech AI IDE Template
# Purpose: Unified entry point for cross-platform project orchestration and developer governance.
# Design:
#   - POSIX-compliant shell delegation with Windows-specific Batch abstractions.
#   - Standardized lifecycle targets: init -> unirtm install -> verify -> audit.
#   - "World Class" AI documentation style (English-only technical metadata).

# =============================================================================
# Global Options
# =============================================================================
# Verbosity level: 0 (quiet), 1 (normal), 2 (verbose)
# Supports: unirtm install --verbose or V=2
V ?= 1
VERBOSE ?= $(V)
export VERBOSE

# Pass flags to sub-scripts
SCRIPT_ARGS :=
ifeq ($(shell [ $(VERBOSE) -ge 2 ] && echo 1),1)
	SCRIPT_ARGS += --verbose
endif
ifeq ($(shell [ $(VERBOSE) -eq 0 ] && echo 1),1)
	SCRIPT_ARGS += --quiet
endif

# =============================================================================
# OS Detection
# =============================================================================
ifeq ($(OS),Windows_NT)
	# Detect if we are running in a POSIX-like environment (Git Bash, WSL, etc.)
	# We check if 'sh' works and returns expected output.
	IS_POSIX := $(shell sh -c 'echo 1' 2>/dev/null)
	ifeq ($(IS_POSIX),1)
		OS_NAME := POSIX_WINDOWS
		ARCH_NAME := $(shell uname -m)
		SHELL_NAME := $(shell basename $$SHELL)
		# Colors for POSIX
		BLUE   := $(shell printf '\033[0;34m')
		GREEN  := $(shell printf '\033[0;32m')
		YELLOW := $(shell printf '\033[1;33m')
		RED    := $(shell printf '\033[0;31m')
		NC     := $(shell printf '\033[0m')
	else
		OS_NAME := Windows
		ARCH_NAME := $(PROCESSOR_ARCHITECTURE)
		SHELL_NAME := powershell.exe
		SHELL   := powershell.exe
		.SHELLFLAGS := -NoProfile -Command
		# Colors for native Windows (PowerShell handles this, but for make echo)
		BLUE   :=
		GREEN  :=
		YELLOW :=
		RED    :=
		NC     :=
	endif
else
	OS_NAME := $(shell uname -s)
	ARCH_NAME := $(shell uname -m)
	SHELL_NAME := $(shell basename $$SHELL)
	# Colors for POSIX
	BLUE   := $(shell printf '\033[0;34m')
	GREEN  := $(shell printf '\033[0;32m')
	YELLOW := $(shell printf '\033[1;33m')
	RED    := $(shell printf '\033[0;31m')
	NC     := $(shell printf '\033[0m')
endif

# SSoT Project Metadata
PROJECT_VERSION := $(shell cat VERSION 2>/dev/null || echo "unknown")
GIT_BRANCH      := $(shell git branch --show-current 2>/dev/null || echo "not a git repo")

# =============================================================================
# Targets
# =============================================================================
.PHONY: all help init lint format test build clean commit verify release env update audit bench docs archive-changelog check-env sync-docs precommit license-add license-check gen-dependabot sync-harden-runner update-tools

# Default target: display help
all: help

help: ## Show this help message
	@printf "$(BLUE)Snowdream Tech AI IDE Template$(NC) ($(GREEN)v$(PROJECT_VERSION)$(NC))\n"
	@printf "%-16s $(GREEN)$(OS_NAME)$(NC) ($(ARCH_NAME))\n" "Detected OS:"
	@printf "%-16s $(GREEN)$(SHELL_NAME)$(NC)\n" "Current Shell:"
	@printf "%-16s $(GREEN)$(GIT_BRANCH)$(NC)\n\n" "Git Branch:"
	@printf "$(YELLOW)Usage:$(NC)\n"
	@printf "  make $(GREEN)<target>$(NC) [ARGS=\"...\"] [V=1|2] [VARIABLE=value]\n\n"
	@printf "$(YELLOW)Main Lifecycle Targets:$(NC)\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-18s$(NC) %s\n", $$1, $$2}'

# Lifecycle Targets
init: ## Hydrate project from template (rename placeholders)
ifeq ($(OS_NAME),Windows)
	@scripts/init-project.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/init-project.sh $(SCRIPT_ARGS) $(ARGS)
endif



lint: ## Run standardized linter (pre-commit)
ifeq ($(OS_NAME),Windows)
	@scripts/lint.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/lint.sh $(SCRIPT_ARGS) $(ARGS)
endif

precommit: lint ## Alias for lint (Run pre-commit hooks)

format: ## Auto-format code (ruff, prettier, shfmt, etc.)
ifeq ($(OS_NAME),Windows)
	@scripts/format.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/format.sh $(SCRIPT_ARGS) $(ARGS)
endif

test: ## Run unified test suite
ifeq ($(OS_NAME),Windows)
	@scripts/test.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/test.sh $(SCRIPT_ARGS) $(ARGS)
endif

build: ## Build project artifacts
ifeq ($(OS_NAME),Windows)
	@scripts/build.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/build.sh $(SCRIPT_ARGS) $(ARGS)
endif

commit: ## Start the interactive Commitizen CLI
ifeq ($(OS_NAME),Windows)
	@scripts/commit.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/commit.sh $(SCRIPT_ARGS) $(ARGS)
endif

.NOTPARALLEL: verify
verify: ## Run full local verification (env, lint, test, audit)
ifeq ($(OS_NAME),Windows)
	@scripts/verify.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/verify.sh $(SCRIPT_ARGS) $(ARGS)
endif

release: ## Standardized release manager (versioning & tagging)
ifeq ($(OS_NAME),Windows)
	@scripts/release.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/release.sh $(SCRIPT_ARGS) $(ARGS)
endif

env: ## Environment configuration manager (.env)
ifeq ($(OS_NAME),Windows)
	@scripts/env.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/env.sh $(SCRIPT_ARGS) $(ARGS)
endif

update-tools: ## Upgrade Mise-managed tool versions (Tier 1 & 2)
ifeq ($(OS_NAME),Windows)
	@scripts/update-tools.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/update-tools.sh $(SCRIPT_ARGS) $(ARGS)
endif

update: ## Update all tools, dependencies, and hooks (Inc. Mise tool versions)
ifeq ($(OS_NAME),Windows)
	@scripts/update.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/update.sh $(SCRIPT_ARGS) $(ARGS)
endif

audit: ## Run security audit and vulnerability scans
ifeq ($(OS_NAME),Windows)
	@scripts/audit.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/audit.sh $(SCRIPT_ARGS) $(ARGS)
endif



bench: ## Run performance benchmarks
ifeq ($(OS_NAME),Windows)
	@scripts/bench.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/bench.sh $(SCRIPT_ARGS) $(ARGS)
endif

sync-docs: ## Synchronize documentation between versions
ifeq ($(OS_NAME),Windows)
	@scripts/sync-docs.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/sync-docs.sh $(SCRIPT_ARGS) $(ARGS)
endif

license-add: ## Add license headers to core source files (Safe Mode)
	@mise x -- addlicense -v -f .github/license-header.txt \
		$$(find src pkg internal cmd app lib include scripts tests \
		\( -path "*/vendor/*" -o -path "*/node_modules/*" -o -path "*/dist/*" -o -path "*/build/*" \) -prune -o \
		-type f \( -name "*.go" -o -name "*.sh" -o -name "*.py" -o -name "*.js" -o -name "*.mjs" -o -name "*.ts" -o -name "*.tsx" -o -name "*.jsx" -o -name "*.vue" -o -name "*.svelte" -o -name "*.astro" -o -name "*.java" -o -name "*.kt" -o -name "*.swift" -o -name "*.m" -o -name "*.mm" -o -name "*.c" -o -name "*.cpp" -o -name "*.cc" -o -name "*.h" -o -name "*.hpp" -o -name "*.rs" -o -name "*.rb" -o -name "*.php" -o -name "*.cs" -o -name "*.fs" -o -name "*.dart" -o -name "*.rego" -o -name "*.proto" -o -name "*.tf" -o -name "*.tfvars" -o -name "*.kcl" -o -name "*.pkl" -o -name "*.cue" \) -print 2>/dev/null)

license-check: ## Check for missing license headers
	@mise x -- addlicense -check -v -f .github/license-header.txt \
		$$(find src pkg internal cmd app lib include scripts tests \
		\( -path "*/vendor/*" -o -path "*/node_modules/*" -o -path "*/dist/*" -o -path "*/build/*" \) -prune -o \
		-type f \( -name "*.go" -o -name "*.sh" -o -name "*.py" -o -name "*.js" -o -name "*.mjs" -o -name "*.ts" -o -name "*.tsx" -o -name "*.jsx" -o -name "*.vue" -o -name "*.svelte" -o -name "*.astro" -o -name "*.java" -o -name "*.kt" -o -name "*.swift" -o -name "*.m" -o -name "*.mm" -o -name "*.c" -o -name "*.cpp" -o -name "*.cc" -o -name "*.h" -o -name "*.hpp" -o -name "*.rs" -o -name "*.rb" -o -name "*.php" -o -name "*.cs" -o -name "*.fs" -o -name "*.dart" -o -name "*.rego" -o -name "*.proto" -o -name "*.tf" -o -name "*.tfvars" -o -name "*.kcl" -o -name "*.pkl" -o -name "*.cue" \) -print 2>/dev/null)

docs: ## Documentation site manager (dev/build/preview)
ifeq ($(OS_NAME),Windows)
	@scripts/docs.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/docs.sh $(SCRIPT_ARGS) $(ARGS)
endif

gen-dependabot: ## Auto-generate dependabot.yml from detected ecosystems
ifeq ($(OS_NAME),Windows)
	@scripts/gen-dependabot.bat $(SCRIPT_ARGS) $(ARGS)
else
	@mise x -- sh scripts/gen-dependabot.sh $(SCRIPT_ARGS) $(ARGS)
endif

sync-harden-runner: ## Synchronize Harden Runner endpoints to all workflow files
ifeq ($(OS_NAME),Windows)
	@scripts/sync-harden-runner.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/sync-harden-runner.sh $(SCRIPT_ARGS) $(ARGS)
endif



archive-changelog: ## Archive major-version changelog entries
ifeq ($(OS_NAME),Windows)
	@scripts/archive-changelog.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/archive-changelog.sh $(SCRIPT_ARGS) $(ARGS)
endif

check-env: ## Onboarding environment health check
ifeq ($(OS_NAME),Windows)
	@scripts/check-env.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/check-env.sh $(SCRIPT_ARGS) $(ARGS)
endif

clean: ## Remove temporary and generated files
ifeq ($(OS_NAME),Windows)
	@scripts/cleanup.bat $(SCRIPT_ARGS) $(ARGS)
else
	@sh scripts/cleanup.sh $(SCRIPT_ARGS) $(ARGS)
endif

fix: ## Automatically fix common issues (line endings, typos)
	@printf "$(BLUE)Applying automated fixes...$(NC)\n"
	@find scripts -name "*.bat" -o -name "*.ps1" | xargs perl -pi -e 's/\r\n/\n/g; s/\n/\r\n/g'
	@git add --renormalize .
	@printf "$(GREEN)Fixes applied!$(NC)\n"
