# 安装 UniRTM

UniRTM 是一个使用 Go 语言编写且不依赖于任何外部运行时（Runtime）的独立二进制程序。为了方便开发者在任何主流操作系统上无缝接入，我们通过持续集成提供了对极多包管理器渠道的一等公民支持。

> [!TIP]
> 无论使用哪种方式安装，UniRTM 安装完成后都可以通过 `unirtm --version` 验证是否成功。

## 独立脚本安装 (MacOS / Linux)

如果你不想使用任何包管理器，这是最通用的安装方式。脚本会自动检测你的操作系统和 CPU 架构（amd64 / arm64 等），下载最新的二进制包并进行配置。

```bash
curl -fsSL https://github.com/snowdreamtech/unirtm/raw/main/install.sh | sh
```

安装完成后，记得将输出中提示的路径添加到你的 `~/.bashrc` 或 `~/.zshrc` 中：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Homebrew (MacOS / Linux)

对于 Mac 和 Linux 用户，使用 Homebrew 是非常推荐的方式：

```bash
brew install snowdreamtech/tap/unirtm
```

## Windows 安装渠道

UniRTM 深度支持 Windows 生态系统。

### Winget

Windows 官方推荐的包管理器：

```powershell
winget install SnowdreamTech.unirtm
```

### Scoop

针对开发者优化的无权限要求包管理器：

```powershell
scoop bucket add snowdreamtech https://github.com/snowdreamtech/scoop-bucket.git
scoop install unirtm
```

## Linux 发行版原生包

我们为绝大部分主流 Linux 发行版构建了对应的底层包。你可以从 [GitHub Releases](https://github.com/snowdreamtech/unirtm/releases) 下载最新版本的安装包。

### DEB (Debian / Ubuntu)

```bash
sudo dpkg -i unirtm_*_linux_amd64.deb
```

### RPM (CentOS / RHEL / Fedora)

```bash
sudo rpm -ivh unirtm_*_linux_amd64.rpm
```

### APK (Alpine Linux)

完美支持体积极小的 Alpine 容器镜像：

```bash
apk add --allow-untrusted unirtm_*_linux_amd64.apk
```

### AUR (Arch Linux)

Arch Linux 用户可以直接从 AUR 源进行安装：

```bash
yay -S unirtm-bin
```

## Nix / NixOS

针对函数式包管理器的拥趸，可以通过我们官方的 NUR (Nix User Repository) 进行无状态安装：

```bash
nix profile install github:snowdreamtech/nur#unirtm
```

## 源码安装 (Go)

如果你是 Go 语言开发者或想从源码直接编译安装，可以使用标准的 `go install` 命令。这要求本地已安装 Go 1.20+ 环境：

```bash
go install github.com/snowdreamtech/unirtm@latest
```
