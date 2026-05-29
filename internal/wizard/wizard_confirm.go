package wizard

// wizard_confirm.go 合并确认页与成功摘要输出。
//
// 设计说明：
// - 确认页（confirm）和摘要输出（summary）都属于向导"最后一步"展示逻辑
// - 合并后减少 2 个碎片文件

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
)

// --- 确认页 ---

// confirmPasswordRun 展示密码模式执行摘要，用户确认后才开始写盘与部署。
func confirmPasswordRun(in PasswordModeInput, configPath string) error {
	summary := buildPasswordConfirmSummary(in, configPath)
	return runConfirmForm(summary)
}

// confirmKeyRun 展示密钥模式执行摘要。
func confirmKeyRun(in KeyModeInput, configPath string) error {
	summary := buildKeyConfirmSummary(in, configPath)
	return runConfirmForm(summary)
}

func buildPasswordConfirmSummary(in PasswordModeInput, configPath string) string {
	alias := confirmAlias(in.Alias, in.HostName)
	summary := i18n.T(
		i18n.KeyWizardConfirmSummaryPW,
		safeTTYString(alias),
		in.User,
		strings.TrimSpace(in.HostName),
		effectivePort(in.Port),
		safeTTYString(configPath),
	) + "\n" + i18n.T(i18n.KeyWizardConfirmHostKey)
	return summary + confirmRemarkLine(in.Remark)
}

func buildKeyConfirmSummary(in KeyModeInput, configPath string) string {
	alias := confirmAlias(in.Alias, in.HostName)
	summary := i18n.T(
		i18n.KeyWizardConfirmSummaryKey,
		safeTTYString(alias),
		in.User,
		strings.TrimSpace(in.HostName),
		effectivePort(in.Port),
		safeTTYString(configPath),
		safeTTYString(in.IdentityFile),
	) + "\n" + i18n.T(i18n.KeyWizardConfirmHostKey)
	return summary + confirmRemarkLine(in.Remark)
}

func confirmRemarkLine(remark string) string {
	if strings.TrimSpace(remark) == "" {
		return ""
	}
	return "\n" + i18n.T(i18n.KeyWizardConfirmRemark, remark)
}

func confirmAlias(alias, hostName string) string {
	if a := strings.TrimSpace(alias); a != "" {
		return keys.NormalizeHostAlias(a)
	}
	return keys.NormalizeHostAlias(hostName)
}

func runConfirmForm(summary string) error {
	ok := true // 默认选中「确认执行」
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(confirmTitle()).
				Description(summary),
			huh.NewConfirm().
				Title(i18n.T(i18n.KeyWizardConfirmTitle)).
				Affirmative(i18n.T(i18n.KeyWizardConfirmYes)).
				Negative(i18n.T(i18n.KeyWizardConfirmNo)).
				Value(&ok),
		),
	).WithTheme(WizardTheme()).
		WithShowErrors(false)

	if err := form.Run(); err != nil {
		return err
	}
	if !ok {
		SetFormRetryHint(i18n.T(i18n.KeyWizardRetryHint))
		return ErrWizardRetryForm
	}
	return nil
}

// --- 成功摘要输出 ---

// WriteAddSuccessSummary 向 stderr 输出本次 add 已完成的操作摘要。
func WriteAddSuccessSummary(stderr io.Writer, result *WizardResult, configPath string) {
	if result == nil {
		return
	}
	write := func(format string, args ...any) {
		_, _ = fmt.Fprintf(stderr, format, args...)
	}
	write("%s\n", i18n.T(i18n.KeySummaryHeadline))
	write(i18n.T(i18n.KeySummarySSHCmd), result.Alias)
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
	write("%s\n", i18n.T(i18n.KeySummaryListHint))
	write("%s\n", i18n.T(i18n.KeySummaryNextStep, configPath))
}
