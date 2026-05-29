package wizard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fuckssh/fuckssh/internal/platform"
)

// ConnectionMode 表示 add 向导的认证方式。
type ConnectionMode string

const (
	ModeKey      ConnectionMode = "key"
	ModePassword ConnectionMode = "password"
)

// Run 编排 add 向导：全量表单收集 → 确认 → 按认证方式执行。
// configPath 为 ssh config 路径（别名检测与确认摘要共用）。
func Run(configPath string) (*WizardResult, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, fmt.Errorf("%w: config path must not be empty", ErrInvalidInput)
	}

	ctx := context.Background()
	var draft *AddInput

	for {
		in, err := collectAddInput(ctx, configPath, nil, nil, draft)
		if err != nil {
			return nil, mapWizardAbort(err)
		}

		switch in.Mode {
		case ModePassword:
			result, bakPath, retryDraft, err := runPasswordFromAddInput(ctx, configPath, in)
			if errors.Is(err, ErrWizardRetryForm) {
				draft = retryDraft
				continue
			}
			if err != nil {
				return nil, mapWizardAbort(err)
			}
			if result != nil {
				result.PasswordFlowComplete = true
				result.BackupPath = bakPath
			}
			return result, nil

		case ModeKey:
			result, retryDraft, err := runKeyFromAddInput(configPath, in)
			if errors.Is(err, ErrWizardRetryForm) {
				draft = retryDraft
				continue
			}
			if err != nil {
				return nil, mapWizardAbort(err)
			}
			return result, nil

		default:
			return nil, fmt.Errorf("%w: unknown connection mode", ErrInvalidInput)
		}
	}
}

// runPasswordFromAddInput 校验、确认并执行密码模式；retryDraft 非空时表示需重新填表。
func runPasswordFromAddInput(ctx context.Context, configPath string, in AddInput) (*WizardResult, string, *AddInput, error) {
	final, err := finalizePasswordModeInput(in.ToPasswordModeInput())
	if err != nil {
		return nil, "", nil, err
	}

	if err := confirmPasswordRun(final, configPath); err != nil {
		if errors.Is(err, ErrWizardRetryForm) {
			return nil, "", &in, ErrWizardRetryForm
		}
		return nil, "", nil, err
	}

	sshDir, err := platform.SSHDir()
	if err != nil {
		return nil, "", nil, err
	}
	deps := defaultPasswordFlowDeps(sshDir)
	result, bakPath, err := executePasswordFlow(ctx, final, configPath, deps)
	return result, bakPath, nil, err
}

// runKeyFromAddInput 校验、确认并返回密钥模式结果（写盘由 cmd 侧 RunKeyFlow 完成）。
func runKeyFromAddInput(configPath string, in AddInput) (*WizardResult, *AddInput, error) {
	final, err := finalizeKeyModeInput(in.ToKeyModeInput(), os.Stat)
	if err != nil {
		return nil, nil, err
	}

	if err := confirmKeyRun(final, configPath); err != nil {
		if errors.Is(err, ErrWizardRetryForm) {
			return nil, &in, ErrWizardRetryForm
		}
		return nil, nil, err
	}

	return &WizardResult{
		Alias:        final.Alias,
		HostName:     final.HostName,
		User:         final.User,
		Port:         final.Port,
		IdentityFile: final.IdentityFile,
		Remark:       final.Remark,
	}, nil, nil
}
