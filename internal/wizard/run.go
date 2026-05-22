package wizard

import (
	"context"
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// ConnectionMode 表示 add 向导的连接方式。
type ConnectionMode string

const (
	ModeKey      ConnectionMode = "key"
	ModePassword ConnectionMode = "password"
)

// Run 编排 add 向导：选择模式后进入对应流程。
// configPath 写入密码模式使用的 ssh config（密钥模式由 cmd 层追加）。
func Run(configPath string) (*WizardResult, error) {
	var mode ConnectionMode = ModePassword
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[ConnectionMode]().
				Title(i18n.T(i18n.KeyWizardConnModeTitle)).
				Options(
					huh.NewOption(i18n.T(i18n.KeyWizardModePassword), ModePassword),
					huh.NewOption(i18n.T(i18n.KeyWizardModeKey), ModeKey),
				).
				Value(&mode),
		),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}

	switch mode {
	case ModeKey:
		return RunKeyMode()
	case ModePassword:
		result, bakPath, err := RunPasswordMode(context.Background(), configPath)
		if err != nil {
			return nil, err
		}
		if result != nil {
			result.PasswordFlowComplete = true
			result.BackupPath = bakPath
		}
		return result, nil
	default:
		return nil, errors.New("wizard: unknown connection mode")
	}
}
