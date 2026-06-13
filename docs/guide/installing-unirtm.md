# Installing UniRTM

UniRTM is written in Go and is distributed as a standalone binary with absolutely no external runtime dependencies. To ensure a seamless experience on any major operating system, we provide first-class support for a wide variety of package managers through our CI/CD pipelines.

> [!TIP]
> After installing, you can verify that UniRTM is correctly installed by running `unirtm --version`.

## Standalone Script (MacOS / Linux)

If you prefer not to use a package manager, this is the most universal method. The script will automatically detect your OS and CPU architecture (amd64 / arm64, etc.), download the latest binary, and set it up for you.

```bash
curl -fsSL https://github.com/snowdreamtech/unirtm/raw/main/install.sh | sh
```

After installation, make sure to add the path mentioned in the output to your `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Homebrew (MacOS / Linux)

For Mac and Linux users, Homebrew is a highly recommended method:

```bash
brew install snowdreamtech/tap/unirtm
```

## Windows Installation

UniRTM natively supports the Windows ecosystem.

### Winget

The official Windows package manager:

```powershell
winget install SnowdreamTech.unirtm
```

### Scoop

A command-line installer for Windows built for developers:

```powershell
scoop bucket add snowdreamtech https://github.com/snowdreamtech/scoop-bucket.git
scoop install unirtm
```

## Linux Native Packages

We build underlying native packages for almost all mainstream Linux distributions. You can download the latest installation packages directly from our [GitHub Releases](https://github.com/snowdreamtech/unirtm/releases).

### DEB (Debian / Ubuntu)

```bash
sudo dpkg -i unirtm_*_linux_amd64.deb
```

### RPM (CentOS / RHEL / Fedora)

```bash
sudo rpm -ivh unirtm_*_linux_amd64.rpm
```

### APK (Alpine Linux)

Perfectly supports ultra-small Alpine container images:

```bash
apk add --allow-untrusted unirtm_*_linux_amd64.apk
```

### AUR (Arch Linux)

Arch Linux users can install directly from the AUR repository:

```bash
yay -S unirtm-bin
```

## Nix / NixOS

For fans of functional package management, UniRTM can be installed statelessly via our official NUR (Nix User Repository):

```bash
nix profile install github:snowdreamtech/nur#unirtm
```

## Building from Source (Go)

If you are a Go developer or want to compile directly from source, you can use the standard `go install` command. This requires a local Go 1.20+ environment:

```bash
go install github.com/snowdreamtech/unirtm@latest
```
