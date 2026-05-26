package wizard

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// collectAddInput 全量堆叠表单收集 7 项；draft 非空时预填（确认返回修改时用）。
// 顺序：主机 → 别名 → 用户 → 端口 → 认证+凭证 → 备注。
func collectAddInput(ctx context.Context, configPath string, pwTest passwordAuthTestFn, keyTest keyAuthTestFn, draft *AddInput) (AddInput, error) {
	if pwTest == nil {
		pwTest = defaultPasswordAuthTest
	}
	if keyTest == nil {
		keyTest = defaultKeyAuthTest
	}

	var in AddInput
	if draft != nil {
		in = *draft
	}
	if in.Mode == "" {
		in.Mode = ModePassword
	}

	var pwScratch PasswordModeInput
	var keyScratch KeyModeInput
	syncScratch := func() {
		pwScratch = in.ToPasswordModeInput()
		keyScratch = in.ToKeyModeInput()
	}
	syncScratch()

	pwField := NewPasswordTestField(ctx, &pwScratch, pwTest,
		func() {
			in.SyncFromPassword(pwScratch)
			in.AuthTestOK = true
		},
		func() {
			in.SyncFromPassword(pwScratch)
			in.AuthTestOK = false
		},
	).Title(fieldLabel(6, i18n.KeyWizardPassword)).
		Key("password").
		Value(&pwScratch.Password)

	keyField := NewKeyIdentityField(ctx, &keyScratch, keyTest,
		func() {
			in.SyncFromKey(keyScratch)
			in.AuthTestOK = true
		},
		func() {
			in.SyncFromKey(keyScratch)
			in.AuthTestOK = false
		},
	).Title(fieldLabel(6, i18n.KeyWizardIdentityFile)).
		Key("identity").
		Value(&keyScratch.IdentityFile)

	credField := NewCredentialSwitchField(&in.Mode, pwField, keyField)

	form := huh.NewForm(
		huh.NewGroup(
			NewHostField(syncScratch).
				Title(fieldLabel(1, i18n.KeyWizardHostIP)).
				Key("hostname").
				Value(&in.HostName),
		),
		huh.NewGroup(
			NewAliasField(configPath, &in.HostName).
				Title(fieldLabel(2, i18n.KeyWizardAlias)).
				Key("alias").
				Value(&in.Alias),
		),
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(3, i18n.KeyWizardUser)).
				Inline(true).
				Placeholder("root").
				Value(&in.User).
				Validate(func(s string) error {
					syncScratch()
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(4, i18n.KeyWizardPort)).
				Inline(true).
				Placeholder("22").
				Value(&in.Port).
				Validate(func(s string) error {
					if err := validatePort(s); err != nil {
						return err
					}
					syncScratch()
					return nil
				}),
		),
		huh.NewGroup(
			NewAuthModeField(&in.Mode, func(prev, next ConnectionMode) {
				clearCredentialsOnModeChange(&in, prev)
				syncScratch()
			}).
				Title(fieldLabel(5, i18n.KeyWizardConnModeTitle)).
				Key("auth_mode"),
			credField.Key("credential"),
		),
		huh.NewGroup(
			huh.NewInput().
				Title(fieldLabel(7, i18n.KeyWizardRemark)).
				Inline(true).
				Placeholder(i18n.T(i18n.KeyWizardRemarkEmptyHint)).
				Value(&in.Remark),
		),
	).WithKeyMap(wizardFormKeyMap()).
		WithLayout(huh.LayoutStack).
		WithShowErrors(false)

	if err := form.Run(); err != nil {
		return AddInput{}, err
	}
	syncScratch()
	in.SyncFromPassword(pwScratch)
	in.SyncFromKey(keyScratch)

	if in.Alias == "" {
		in.Alias = resolveAliasCandidate("", in.HostName)
	}

	if !in.AuthTestOK {
		return AddInput{}, fmt.Errorf("%w: %s", ErrInvalidInput, i18n.T(i18n.KeyWizardErrAuthTestRequired))
	}
	return in, nil
}
