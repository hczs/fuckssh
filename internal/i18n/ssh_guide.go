package i18n

import "runtime"

// InstallOpenSSHGuide 返回当前操作系统安装 OpenSSH 客户端的指引（随界面语言）。
func InstallOpenSSHGuide() string {
	return installOpenSSHGuideFor(runtime.GOOS, Current())
}

func installOpenSSHGuideFor(goos string, lang Lang) string {
	if lang == LangEN {
		return installOpenSSHGuideEN(goos)
	}
	return installOpenSSHGuideZH(goos)
}

func installOpenSSHGuideZH(goos string) string {
	switch goos {
	case "windows":
		return `Windows 安装 OpenSSH 客户端：
1. 打开「设置」→「应用」→「可选功能」
2. 点击「查看功能」或「添加可选功能」
3. 搜索并安装「OpenSSH 客户端」
4. 安装后重新打开终端，执行 ssh -V 确认可用`
	case "darwin":
		return `macOS 通常已内置 OpenSSH 客户端：
1. 打开「终端」，执行 ssh -V 确认是否可用
2. 若提示找不到命令，可安装 Xcode Command Line Tools：
   xcode-select --install
3. 安装完成后重新打开终端再试`
	default:
		return `Linux 安装 OpenSSH 客户端：
- Debian/Ubuntu: sudo apt install openssh-client
- Fedora/RHEL:   sudo dnf install openssh-clients
- Arch:          sudo pacman -S openssh
安装后执行 ssh -V 确认可用`
	}
}

func installOpenSSHGuideEN(goos string) string {
	switch goos {
	case "windows":
		return `Install OpenSSH Client on Windows:
1. Open Settings → Apps → Optional features
2. Click "View features" or "Add an optional feature"
3. Search and install "OpenSSH Client"
4. Restart the terminal and run: ssh -V`
	case "darwin":
		return `macOS usually includes OpenSSH:
1. Open Terminal and run: ssh -V
2. If missing, install Xcode Command Line Tools:
   xcode-select --install
3. Restart Terminal and try again`
	default:
		return `Install OpenSSH client on Linux:
- Debian/Ubuntu: sudo apt install openssh-client
- Fedora/RHEL:   sudo dnf install openssh-clients
- Arch:          sudo pacman -S openssh
Then run: ssh -V`
	}
}
