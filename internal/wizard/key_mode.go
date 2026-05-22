package wizard

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/platform"
)

// ErrInvalidInput 表示向导收集的字段不合法（CLI 映射退出码 1）。
var ErrInvalidInput = errors.New("wizard: 输入无效")

// KeyModeInput 为密钥连接模式的用户输入（可在测试中直接构造）。
type KeyModeInput struct {
	HostName     string
	User         string
	Port         string
	Alias        string
	IdentityFile string
}

// WizardResult 为向导完成后的连接参数。
type WizardResult struct {
	Alias        string
	HostName     string
	User         string
	Port         string
	IdentityFile string
	// PasswordFlowComplete 为 true 时表示密码模式已在向导内完成备份、写 config 与部署。
	PasswordFlowComplete bool
	// BackupPath 为密码模式备份的 config 路径（成功时供提示；失败时错误信息亦会包含）。
	BackupPath string
}

type fileStatFunc func(name string) (os.FileInfo, error)

// RunKeyMode 通过 huh 表单收集密钥连接所需字段。
func RunKeyMode() (*WizardResult, error) {
	var in KeyModeInput
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("服务器 IP 或域名").
				Value(&in.HostName).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("不能为空")
					}
					return nil
				}),
			huh.NewInput().
				Title("SSH 用户名").
				Value(&in.User).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("不能为空")
					}
					return nil
				}),
			huh.NewInput().
				Title("私钥路径").
				Description("已有私钥的完整路径，例如 ~/.ssh/id_ed25519").
				Value(&in.IdentityFile).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("不能为空")
					}
					return nil
				}),
			huh.NewInput().
				Title("SSH 端口").
				Placeholder("22").
				Value(&in.Port),
			huh.NewInput().
				Title("Host 别名").
				Description("回车则根据 IP/域名自动生成").
				Value(&in.Alias),
		),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}
	return keyModeResult(in, os.Stat)
}

func keyModeResult(in KeyModeInput, stat fileStatFunc) (*WizardResult, error) {
	out, err := finalizeKeyModeInput(in, stat)
	if err != nil {
		return nil, err
	}
	return &WizardResult{
		Alias:        out.Alias,
		HostName:     out.HostName,
		User:         out.User,
		Port:         out.Port,
		IdentityFile: out.IdentityFile,
	}, nil
}

// finalizeKeyModeInput 校验并补全默认值（供单测与 RunKeyMode 共用）。
func finalizeKeyModeInput(in KeyModeInput, stat fileStatFunc) (KeyModeInput, error) {
	in.HostName = strings.TrimSpace(in.HostName)
	in.User = strings.TrimSpace(in.User)
	in.Port = strings.TrimSpace(in.Port)
	in.Alias = strings.TrimSpace(in.Alias)
	in.IdentityFile = strings.TrimSpace(in.IdentityFile)

	if in.HostName == "" || in.User == "" || in.IdentityFile == "" {
		return KeyModeInput{}, fmt.Errorf("%w: 请填写 IP/域名、用户名与私钥路径", ErrInvalidInput)
	}

	expanded, err := platform.ExpandPath(in.IdentityFile)
	if err != nil {
		return KeyModeInput{}, fmt.Errorf("%w: 私钥路径: %v", ErrInvalidInput, err)
	}
	in.IdentityFile = expanded

	if _, err := stat(in.IdentityFile); err != nil {
		if os.IsNotExist(err) {
			return KeyModeInput{}, fmt.Errorf("%w: 私钥不存在: %s", ErrInvalidInput, in.IdentityFile)
		}
		return KeyModeInput{}, fmt.Errorf("%w: 无法读取私钥: %v", ErrInvalidInput, err)
	}

	if in.Port == "" {
		in.Port = "22"
	}
	if in.Alias == "" {
		in.Alias = keys.SanitizeAlias(in.HostName)
		if in.Alias == "" {
			return KeyModeInput{}, fmt.Errorf("%w: 无法根据 HostName 生成别名", ErrInvalidInput)
		}
	} else {
		in.Alias = keys.SanitizeAlias(in.Alias)
	}

	return in, nil
}
