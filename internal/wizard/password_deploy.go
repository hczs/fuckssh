package wizard

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

// ErrDeployAborted 表示用户在公钥部署阶段选择取消（将触发本地回滚）。
var ErrDeployAborted = errors.New("wizard: deploy aborted")

// deployPublicKeyWithRetry 部署公钥；若因 authorized_keys 无写权限失败，提示用户修复后等待确认再重试。
func deployPublicKeyWithRetry(ctx context.Context, in PasswordModeInput, pubLine string, deps passwordFlowDeps) error {
	prompt := deps.promptPermissionFix
	if prompt == nil {
		prompt = defaultPromptPermissionFix
	}
	for {
		err := deployPublicKey(ctx, in, pubLine, deps)
		if err == nil {
			return nil
		}
		var perm *sshclient.AuthorizedKeysNotWritableError
		if !errors.As(err, &perm) {
			return err
		}
		if err := prompt(ctx, perm); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return fmt.Errorf("%w", ErrDeployAborted)
			}
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}

func defaultPromptPermissionFix(ctx context.Context, perm *sshclient.AuthorizedKeysNotWritableError) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	retry := false
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(i18n.T(i18n.KeyWizardPermFixPrompt)).
				Description(perm.Error()),
			huh.NewConfirm().
				Title(i18n.T(i18n.KeyWizardPermFixPrompt)).
				Affirmative(i18n.T(i18n.KeyWizardPermFixYes)).
				Negative(i18n.T(i18n.KeyWizardPermFixNo)).
				Value(&retry),
		),
	).WithShowErrors(false)
	if err := form.Run(); err != nil {
		return err
	}
	if !retry {
		return huh.ErrUserAborted
	}
	return nil
}
