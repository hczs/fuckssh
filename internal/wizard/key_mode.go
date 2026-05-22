package wizard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/platform"
)

// ErrInvalidInput 表示向导收集的字段不合法（CLI 映射退出码 1）。
var ErrInvalidInput = errors.New("wizard: invalid input")

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

// RunKeyMode 通过堆叠表单收集密钥连接所需字段。
func RunKeyMode(configPath string) (*WizardResult, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, fmt.Errorf("%w: config 路径不能为空", ErrInvalidInput)
	}

	var draft KeyModeInput
	var out KeyModeInput
	for {
		in, err := collectKeyModeInput(context.Background(), nil, &draft)
		if err != nil {
			return nil, err
		}
		draft = in

		out, err = finalizeKeyModeInput(in, os.Stat)
		if err != nil {
			return nil, err
		}

		out.Alias, err = ensureAvailableAlias(configPath, out.Alias)
		if err != nil {
			return nil, err
		}

		if err := confirmKeyRun(out, configPath); err != nil {
			if errors.Is(err, ErrWizardRetryForm) {
				draft = out
				continue
			}
			return nil, err
		}
		break
	}

	return &WizardResult{
		Alias:        out.Alias,
		HostName:     out.HostName,
		User:         out.User,
		Port:         out.Port,
		IdentityFile: out.IdentityFile,
	}, nil
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
		return KeyModeInput{}, fmt.Errorf("%w: %s", ErrInvalidInput, i18n.T(i18n.KeyWizardErrFillBasic))
	}

	expanded, err := platform.ExpandPath(in.IdentityFile)
	if err != nil {
		return KeyModeInput{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	in.IdentityFile = expanded

	if _, err := stat(in.IdentityFile); err != nil {
		if os.IsNotExist(err) {
			return KeyModeInput{}, fmt.Errorf("%w: %s", ErrInvalidInput, i18n.T(i18n.KeyWizardErrKeyMissing, in.IdentityFile))
		}
		return KeyModeInput{}, fmt.Errorf("%w: %s", ErrInvalidInput, i18n.T(i18n.KeyWizardErrKeyRead, err))
	}

	if in.Port == "" {
		in.Port = "22"
	}
	if err := validatePort(in.Port); err != nil {
		return KeyModeInput{}, err
	}
	if in.Alias == "" {
		in.Alias = keys.SanitizeAlias(in.HostName)
		if in.Alias == "" {
			return KeyModeInput{}, fmt.Errorf("%w: %s", ErrInvalidInput, i18n.T(i18n.KeyWizardErrAliasGen))
		}
	} else {
		in.Alias = keys.SanitizeAlias(in.Alias)
	}

	return in, nil
}
