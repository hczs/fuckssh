package wizard

import (
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
)

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
	).WithShowErrors(false)

	if err := form.Run(); err != nil {
		return err
	}
	if !ok {
		SetFormRetryHint(i18n.T(i18n.KeyWizardRetryHint))
		return ErrWizardRetryForm
	}
	return nil
}
