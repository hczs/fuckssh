package wizard

import (
	"fmt"
	"io"

	"github.com/fuckssh/fuckssh/internal/i18n"
)

// WriteAddSuccessSummary 向 stderr 输出本次 add 已完成的操作摘要。
func WriteAddSuccessSummary(stderr io.Writer, result *WizardResult, configPath string) {
	if result == nil {
		return
	}
	write := func(format string, args ...any) {
		_, _ = fmt.Fprintf(stderr, format, args...)
	}
	write("%s", i18n.T(i18n.KeySummaryTitle))
	if result.PasswordFlowComplete {
		if result.BackupPath != "" {
			write(i18n.T(i18n.KeySummaryBackup), result.BackupPath)
		}
		write(i18n.T(i18n.KeySummaryKeygen), result.IdentityFile)
		write(i18n.T(i18n.KeySummaryHostWritten), result.Alias, configPath)
		write("%s", i18n.T(i18n.KeySummaryDeployed))
	} else {
		if result.BackupPath != "" {
			write(i18n.T(i18n.KeySummaryBackup), result.BackupPath)
		}
		write(i18n.T(i18n.KeySummaryHostWritten), result.Alias, configPath)
		write(i18n.T(i18n.KeySummaryExistingKey), result.IdentityFile)
	}
	write("%s", i18n.T(i18n.KeySummaryHostKey))
	write("%s", i18n.T(i18n.KeySummaryNextStep, configPath, result.Alias))
}
