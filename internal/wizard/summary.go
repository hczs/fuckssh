package wizard

import (
	"fmt"
	"io"
)

// HostKeySecurityNotice 为 MVP 未校验 Host Key 时的固定安全提示。
const HostKeySecurityNotice = "安全提示：当前未校验服务器 Host Key（存在中间人攻击风险），后续版本将支持 known_hosts 校验。"

// WriteAddSuccessSummary 向 stderr 输出本次 add 已完成的操作摘要。
func WriteAddSuccessSummary(stderr io.Writer, result *WizardResult, configPath string) {
	if result == nil {
		return
	}
	write := func(format string, args ...any) {
		_, _ = fmt.Fprintf(stderr, format, args...)
	}
	write("本次已完成：\n")
	if result.PasswordFlowComplete {
		if result.BackupPath != "" {
			write("  · 已备份 SSH config 至 %s\n", result.BackupPath)
		}
		write("  · 已生成 Ed25519 密钥：%s\n", result.IdentityFile)
		write("  · 已写入 Host %s 到 %s\n", result.Alias, configPath)
		write("  · 公钥已部署到远端 ~/.ssh/authorized_keys\n")
	} else if result.BackupPath != "" {
		write("  · 已备份 SSH config 至 %s\n", result.BackupPath)
		write("  · 已写入 Host %s 到 %s\n", result.Alias, configPath)
		write("  · 使用已有私钥：%s\n", result.IdentityFile)
	}
	if result.PasswordFlowComplete {
		write("  · %s\n", HostKeySecurityNotice)
	}
}
