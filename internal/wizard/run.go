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
// configPath 为 ssh config 路径（两种模式均用于别名检测与确认摘要）。
func Run(configPath string) (*WizardResult, error) {
	var mode = ModePassword
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[ConnectionMode]().
				Title(stepTitle(1, i18n.KeyWizardConnModeTitle)).
				Description(modeSelectDescription()).
				Options(
					huh.NewOption(modePasswordLabel(), ModePassword),
					huh.NewOption(modeKeyLabel(), ModeKey),
				).
				Value(&mode),
		),
	)
	if err := form.Run(); err != nil {
		return nil, mapWizardAbort(err)
	}

	switch mode {
	case ModeKey:
		return RunKeyMode(configPath)
	case ModePassword:
		result, bakPath, err := RunPasswordMode(context.Background(), configPath)
		if err != nil {
			return nil, mapWizardAbort(err)
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

func modeSelectDescription() string {
	return i18n.T(i18n.KeyWizardWelcome) + "\n" +
		i18n.T(i18n.KeyWizardWelcomeETA) + "\n\n" +
		i18n.T(i18n.KeyWizardConnModeDesc)
}

// modePasswordLabel 主选项标题 + 副标题（换行展示）。
func modePasswordLabel() string {
	return i18n.T(i18n.KeyWizardModePassword) + "\n  " + i18n.T(i18n.KeyWizardModePasswordSub)
}

func modeKeyLabel() string {
	return i18n.T(i18n.KeyWizardModeKey) + "\n  " + i18n.T(i18n.KeyWizardModeKeySub)
}

