# UniRTM activation script
# Shell: zsh
# Scope: global

# UniRTM PATH mode activation
export UNIRTM_PATH="/Users/snowdream/.local/share/unirtm/installs/npm-stylelint/17.10.0:/Users/snowdream/.local/share/unirtm/installs/npm-stylelint/17.10.0/bin:/Users/snowdream/.local/share/unirtm/installs/npm-bats/1.13.0:/Users/snowdream/.local/share/unirtm/installs/npm-bats/1.13.0/bin:/Users/snowdream/.local/share/unirtm/installs/npm-sort-package-json/3.6.1:/Users/snowdream/.local/share/unirtm/installs/npm-sort-package-json/3.6.1/bin:/Users/snowdream/.local/share/unirtm/installs/github-google-osv-scanner/2.3.5:/Users/snowdream/.local/share/unirtm/installs/github-google-osv-scanner/2.3.5/bin:/Users/snowdream/.local/share/unirtm/installs/github-dotenv-linter-dotenv-linter/4.0.0:/Users/snowdream/.local/share/unirtm/installs/github-dotenv-linter-dotenv-linter/4.0.0/bin:/Users/snowdream/.local/share/unirtm/installs/npm-eslint/10.3.0:/Users/snowdream/.local/share/unirtm/installs/npm-eslint/10.3.0/bin:/Users/snowdream/.local/share/unirtm/installs/npm-prettier/3.8.3:/Users/snowdream/.local/share/unirtm/installs/npm-prettier/3.8.3/bin:/Users/snowdream/.local/share/unirtm/installs/pipx-yamllint/1.38.0:/Users/snowdream/.local/share/unirtm/installs/pipx-yamllint/1.38.0/bin:/Users/snowdream/.local/share/unirtm/installs/npm-cz-conventional-changelog/3.3.0:/Users/snowdream/.local/share/unirtm/installs/github-astral-sh-ruff/0.15.12:/Users/snowdream/.local/share/unirtm/installs/github-astral-sh-ruff/0.15.12/bin:/Users/snowdream/.local/share/unirtm/installs/npm-markdownlint-cli2/0.22.1:/Users/snowdream/.local/share/unirtm/installs/npm-markdownlint-cli2/0.22.1/bin:/Users/snowdream/.local/share/unirtm/installs/node/26.1.0/bin:/Users/snowdream/.local/share/unirtm/installs/node/26.1.0/npm-global/bin:/Users/snowdream/.local/share/unirtm/installs/pipx-clang-format/22.1.4:/Users/snowdream/.local/share/unirtm/installs/pipx-clang-format/22.1.4/bin:/Users/snowdream/.local/share/unirtm/installs/npm-commitlint-config-conventional/20.5.3:/Users/snowdream/.local/share/unirtm/installs/python/3.14.4/bin:/Users/snowdream/.local/share/unirtm/installs/python/3.14.4/venv/bin:/Users/snowdream/.local/share/unirtm/installs/github-zizmorcore-zizmor/1.24.1:/Users/snowdream/.local/share/unirtm/installs/github-zizmorcore-zizmor/1.24.1/bin:/Users/snowdream/.local/share/unirtm/installs/npm-commitlint-cli/20.5.3:/Users/snowdream/.local/share/unirtm/installs/npm-commitlint-cli/20.5.3/bin:/Users/snowdream/.local/share/unirtm/installs/github-koalaman-shellcheck/0.11.0:/Users/snowdream/.local/share/unirtm/installs/github-koalaman-shellcheck/0.11.0/bin:/Users/snowdream/.local/share/unirtm/installs/github-editorconfig-checker-editorconfig-checker/3.6.1:/Users/snowdream/.local/share/unirtm/installs/github-editorconfig-checker-editorconfig-checker/3.6.1/bin:/Users/snowdream/.local/share/unirtm/installs/github-cli-cli/2.92.0:/Users/snowdream/.local/share/unirtm/installs/github-cli-cli/2.92.0/bin:/Users/snowdream/.local/share/unirtm/installs/pipx-pre-commit/4.6.0:/Users/snowdream/.local/share/unirtm/installs/pipx-pre-commit/4.6.0/bin:/Users/snowdream/.local/share/unirtm/installs/github-hadolint-hadolint/2.14.0:/Users/snowdream/.local/share/unirtm/installs/github-hadolint-hadolint/2.14.0/bin:/Users/snowdream/.local/share/unirtm/installs/npm-commitizen/4.3.1:/Users/snowdream/.local/share/unirtm/installs/npm-commitizen/4.3.1/bin:/Users/snowdream/.local/share/unirtm/installs/github-mvdan-sh/3.13.1:/Users/snowdream/.local/share/unirtm/installs/github-mvdan-sh/3.13.1/bin:/Users/snowdream/.local/share/unirtm/installs/github-anchore-syft/1.44.0:/Users/snowdream/.local/share/unirtm/installs/github-anchore-syft/1.44.0/bin:/Users/snowdream/.local/share/unirtm/installs/go/1.26.3/bin:/Users/snowdream/.local/share/unirtm/installs/npm-taplo-cli/0.7.0:/Users/snowdream/.local/share/unirtm/installs/npm-taplo-cli/0.7.0/bin:/Users/snowdream/.local/share/unirtm/installs/github-gitleaks-gitleaks/8.30.1:/Users/snowdream/.local/share/unirtm/installs/github-gitleaks-gitleaks/8.30.1/bin:/Users/snowdream/.local/share/unirtm/installs/npm-dockerfile-utils/0.16.3:/Users/snowdream/.local/share/unirtm/installs/npm-dockerfile-utils/0.16.3/bin:/Users/snowdream/.local/share/unirtm/installs/github-rhysd-actionlint/1.7.12:/Users/snowdream/.local/share/unirtm/installs/github-rhysd-actionlint/1.7.12/bin"
export PATH="$UNIRTM_PATH:$(echo "$PATH" | sed -E "s|/Users/snowdream/.local/share/unirtm/installs/npm-stylelint/17.10.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-stylelint/17.10.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-bats/1.13.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-bats/1.13.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-sort-package-json/3.6.1:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-sort-package-json/3.6.1/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-google-osv-scanner/2.3.5:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-google-osv-scanner/2.3.5/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-dotenv-linter-dotenv-linter/4.0.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-dotenv-linter-dotenv-linter/4.0.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-eslint/10.3.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-eslint/10.3.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-prettier/3.8.3:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-prettier/3.8.3/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/pipx-yamllint/1.38.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/pipx-yamllint/1.38.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-cz-conventional-changelog/3.3.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-astral-sh-ruff/0.15.12:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-astral-sh-ruff/0.15.12/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-markdownlint-cli2/0.22.1:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-markdownlint-cli2/0.22.1/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/node/26.1.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/node/26.1.0/npm-global/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/pipx-clang-format/22.1.4:?||g; s|/Users/snowdream/.local/share/unirtm/installs/pipx-clang-format/22.1.4/bin:?||g; s|?||g; s|?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-commitlint-config-conventional/20.5.3:?||g; s|/Users/snowdream/.local/share/unirtm/installs/python/3.14.4/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/python/3.14.4/venv/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-zizmorcore-zizmor/1.24.1:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-zizmorcore-zizmor/1.24.1/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-commitlint-cli/20.5.3:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-commitlint-cli/20.5.3/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-koalaman-shellcheck/0.11.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-koalaman-shellcheck/0.11.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-editorconfig-checker-editorconfig-checker/3.6.1:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-editorconfig-checker-editorconfig-checker/3.6.1/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-cli-cli/2.92.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-cli-cli/2.92.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/pipx-pre-commit/4.6.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/pipx-pre-commit/4.6.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-hadolint-hadolint/2.14.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-hadolint-hadolint/2.14.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-commitizen/4.3.1:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-commitizen/4.3.1/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-mvdan-sh/3.13.1:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-mvdan-sh/3.13.1/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-anchore-syft/1.44.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-anchore-syft/1.44.0/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/go/1.26.3/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-taplo-cli/0.7.0:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-taplo-cli/0.7.0/bin:?||g; s|?||g; s|?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-gitleaks-gitleaks/8.30.1:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-gitleaks-gitleaks/8.30.1/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-dockerfile-utils/0.16.3:?||g; s|/Users/snowdream/.local/share/unirtm/installs/npm-dockerfile-utils/0.16.3/bin:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-rhysd-actionlint/1.7.12:?||g; s|/Users/snowdream/.local/share/unirtm/installs/github-rhysd-actionlint/1.7.12/bin:?||g" | sed 's|:$||')"

# Set active tool versions
export UNIRTM_PIPX_PRE_COMMIT_VERSION="4.6.0"
export UNIRTM_GITHUB_HADOLINT_HADOLINT_VERSION="2.14.0"
export UNIRTM_NPM_COMMITIZEN_VERSION="4.3.1"
export UNIRTM_GITHUB_MVDAN_SH_VERSION="3.13.1"
export UNIRTM_GITHUB_ANCHORE_SYFT_VERSION="1.44.0"
export UNIRTM_GO_VERSION="1.26.3"
export UNIRTM_NPM__TAPLO_CLI_VERSION="0.7.0"
export UNIRTM_GITHUB_GITLEAKS_GITLEAKS_VERSION="8.30.1"
export UNIRTM_NPM_DOCKERFILE_UTILS_VERSION="0.16.3"
export UNIRTM_GITHUB_RHYSD_ACTIONLINT_VERSION="1.7.12"
export UNIRTM_NPM_STYLELINT_VERSION="17.10.0"
export UNIRTM_NPM_BATS_VERSION="1.13.0"
export UNIRTM_NPM_SORT_PACKAGE_JSON_VERSION="3.6.1"
export UNIRTM_GITHUB_GOOGLE_OSV_SCANNER_VERSION="2.3.5"
export UNIRTM_GITHUB_DOTENV_LINTER_DOTENV_LINTER_VERSION="4.0.0"
export UNIRTM_NPM_ESLINT_VERSION="10.3.0"
export UNIRTM_NPM_PRETTIER_VERSION="3.8.3"
export UNIRTM_PIPX_YAMLLINT_VERSION="1.38.0"
export UNIRTM_NPM_CZ_CONVENTIONAL_CHANGELOG_VERSION="3.3.0"
export UNIRTM_GITHUB_ASTRAL_SH_RUFF_VERSION="0.15.12"
export UNIRTM_NPM_MARKDOWNLINT_CLI2_VERSION="0.22.1"
export UNIRTM_NODE_VERSION="26.1.0"
export UNIRTM_PIPX_CLANG_FORMAT_VERSION="22.1.4"
export UNIRTM_GITHUB_CHECKMAKE_CHECKMAKE_VERSION="v0.3.2"
export UNIRTM_NPM__COMMITLINT_CONFIG_CONVENTIONAL_VERSION="20.5.3"
export UNIRTM_PYTHON_VERSION="3.14.4"
export UNIRTM_GITHUB_ZIZMORCORE_ZIZMOR_VERSION="1.24.1"
export UNIRTM_NPM__COMMITLINT_CLI_VERSION="20.5.3"
export UNIRTM_GITHUB_KOALAMAN_SHELLCHECK_VERSION="0.11.0"
export UNIRTM_GITHUB_EDITORCONFIG_CHECKER_EDITORCONFIG_CHECKER_VERSION="3.6.1"
export UNIRTM_GITHUB_CLI_CLI_VERSION="2.92.0"

# Set additional environment variables
export PIP_INDEX_URL="https://mirrors.aliyun.com/pypi/simple"
export RUSTUP_UPDATE_ROOT="https://mirrors.aliyun.com/rustup/rustup"
export RUBY_BUILD_MIRROR_URL="https://mirrors.aliyun.com/ruby-build/"
export UNIRTM_DISABLE_AQUA="1"
export CONFIG_AUTO_UPDATE="1"
export UNIRTM_NODE_MIRROR_URL="https://npmmirror.com/mirrors/node/"
export PIP_TRUSTED_HOST="mirrors.aliyun.com"
export NPM_CONFIG_REGISTRY="https://registry.npmmirror.com"
export UNIRTM_NODE_FLAVOR=""
export UNIRTM_GO_DOWNLOAD_MIRROR="https://mirrors.aliyun.com/golang/"
export PYTHONUTF8="1"
export PYTHON_BUILD_MIRROR_URL="https://mirrors.aliyun.com/python/"
export GITHUB_PROXY="https://gh-proxy.sn0wdr1am.com/"
export NODEJS_ORG_MIRROR="https://npmmirror.com/mirrors/node/"
export GOPROXY="https://mirrors.aliyun.com/goproxy/,direct"
export RUSTUP_DIST_SERVER="https://mirrors.aliyun.com/rustup"
export ENABLE_GITHUB_PROXY="1"
export COREPACK_INTEGRITY_KEYS="0"
export UNIRTM_GO_SKIP_CHECKSUM="1"
export UNIRTM_NODE_COMPILE=""

export UNIRTM_ACTIVATION_SCOPE="global"

# UniRTM auto-activation hook
_unirtm_hook() {
  local old_pwd="${UNIRTM_OLD_PWD:-}"
  local new_pwd="$PWD"

  # Only run if directory changed
  if [ "$old_pwd" != "$new_pwd" ]; then
    export UNIRTM_OLD_PWD="$new_pwd"

    # Call unirtm hook-env to get activation changes
    eval "$(/Users/snowdream/Workspace/snowdreamtech/UniRTM/unirtm hook-env --shell zsh)"
  fi
}

# Install the hook in zsh
autoload -U add-zsh-hook
add-zsh-hook precmd _unirtm_hook
