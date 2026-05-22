package wizard

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

// ConnectionMode 表示 add 向导的连接方式。
type ConnectionMode string

const (
	ModeKey      ConnectionMode = "key"
	ModePassword ConnectionMode = "password"
)

// Run 编排 add 向导：选择模式后进入对应流程。
func Run() (*WizardResult, error) {
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
		return nil, fmt.Errorf("%w: 密码连接模式将在后续版本提供", ErrInvalidInput)
	default:
		return nil, errors.New("wizard: 未知连接方式")
	}
}
