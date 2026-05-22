package platform

import "runtime"

// InstallOpenSSHGuide 返回当前操作系统安装 OpenSSH 客户端的中文指引。
func InstallOpenSSHGuide() string {
	return installOpenSSHGuideFor(runtime.GOOS)
}

// installOpenSSHGuideFor 按 GOOS 返回指引，便于表驱动测试。
func installOpenSSHGuideFor(goos string) string {
	switch goos {
	case "windows":
		return windowsOpenSSHGuide()
	case "darwin":
		return darwinOpenSSHGuide()
	default:
		return linuxOpenSSHGuide()
	}
}

func windowsOpenSSHGuide() string {
	return `Windows 安装 OpenSSH 客户端：
1. 打开「设置」→「应用」→「可选功能」
2. 点击「查看功能」或「添加可选功能」
3. 搜索并安装「OpenSSH 客户端」
4. 安装后重新打开终端，执行 ssh -V 确认可用`
}

func darwinOpenSSHGuide() string {
	return `macOS 通常已内置 OpenSSH 客户端：
1. 打开「终端」，执行 ssh -V 确认是否可用
2. 若提示找不到命令，可安装 Xcode Command Line Tools：
   xcode-select --install
3. 安装完成后重新打开终端再试`
}

func linuxOpenSSHGuide() string {
	return `Linux 安装 OpenSSH 客户端：
- Debian/Ubuntu: sudo apt install openssh-client
- Fedora/RHEL:   sudo dnf install openssh-clients
- Arch:          sudo pacman -S openssh
安装后执行 ssh -V 确认可用`
}
