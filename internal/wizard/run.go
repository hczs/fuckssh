package wizard

import (
	"context"
	"errors"

	"github.com/charmbracelet/huh"
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
	var mode ConnectionMode
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[ConnectionMode]().
				Title("连接方式").
				Options(
					huh.NewOption("密钥连接（已有私钥，仅写 config）", ModeKey),
					huh.NewOption("密码连接（生成密钥并部署公钥）", ModePassword),
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
		result, _, err := RunPasswordMode(context.Background(), configPath)
		if err != nil {
			return nil, err
		}
		if result != nil {
			result.PasswordFlowComplete = true
		}
		return result, nil
	default:
		return nil, errors.New("wizard: 未知连接方式")
	}
}
