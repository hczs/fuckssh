package wizard

import (
	"context"
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

// keyAuthTestFn 用于密钥连接测试（单测可注入 mock）。
type keyAuthTestFn func(ctx context.Context, in KeyModeInput) error

func keyConnectionTestFailureMessage(err error) string {
	if errors.Is(err, keys.ErrPassphraseNotSupported) {
		return i18n.T(i18n.KeyWizardPassphraseNA)
	}
	if errors.Is(err, sshclient.ErrDeployAuthFailed) {
		return i18n.T(i18n.KeyWizardKeyAuthFailed)
	}
	return connectionTestFailureMessage(err)
}

// collectKeyModeInput 用堆叠表单逐项收集；draft 非空时预填（确认页返回修改时用）。
func collectKeyModeInput(ctx context.Context, configPath string, testAuth keyAuthTestFn, draft *KeyModeInput) (KeyModeInput, error) {
	if testAuth == nil {
		testAuth = defaultKeyAuthTest
	}
	var in KeyModeInput
	if draft != nil {
		in = *draft
	}
	reveal := &revealState{n: 1}
	seedRevealKey(reveal, in)
	emptyMsg := i18n.T(i18n.KeyWizardErrEmpty)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(stepTitle(2, i18n.KeyWizardHostIP)).
				Description(firstFieldDescription("")).
				Value(&in.HostName).
				Validate(func(s string) error {
					if err := nonEmpty(emptyMsg)(s); err != nil {
						return err
					}
					reveal.showThrough(1)
					return nil
				}),
		).WithHideFunc(hideUntilRevealed(0, reveal)),
		huh.NewGroup(
			huh.NewInput().
				Title(stepTitle(3, i18n.KeyWizardPort)).
				Description(i18n.T(i18n.KeyWizardPortDesc)).
				Placeholder("22").
				Value(&in.Port).
				Validate(func(s string) error {
					if err := validatePort(s); err != nil {
						return err
					}
					reveal.showThrough(2)
					return nil
				}),
		).WithHideFunc(hideUntilRevealed(1, reveal)),
		huh.NewGroup(
			huh.NewInput().
				Title(stepTitle(4, i18n.KeyWizardUser)).
				Description(i18n.T(i18n.KeyWizardUserDesc)).
				Value(&in.User).
				Validate(func(s string) error {
					if err := nonEmpty(emptyMsg)(s); err != nil {
						return err
					}
					reveal.showThrough(3)
					return nil
				}),
		).WithHideFunc(hideUntilRevealed(2, reveal)),
		huh.NewGroup(
			NewKeyIdentityField(ctx, &in, testAuth,
				func() { reveal.showThrough(4) },
				func() { reveal.lockThrough(3) },
			).
				Title(stepTitle(5, i18n.KeyWizardIdentityFile)).
				Key("identity").
				Value(&in.IdentityFile),
		).WithHideFunc(hideUntilRevealed(3, reveal)),
		huh.NewGroup(
			NewAliasField(configPath, &in.HostName).
				Title(stepTitle(6, i18n.KeyWizardAlias)).
				Key("alias").
				Value(&in.Alias).
				OnAdvance(func() { reveal.showThrough(5) }),
		).WithHideFunc(hideUntilRevealed(4, reveal)),
		huh.NewGroup(
			huh.NewInput().
				Title(stepTitle(7, i18n.KeyWizardRemark)).
				Description(i18n.T(i18n.KeyWizardRemarkDesc)).
				Value(&in.Remark),
		).WithHideFunc(hideUntilRevealed(5, reveal)),
	).WithLayout(huh.LayoutStack).WithShowErrors(false)

	if err := form.Run(); err != nil {
		return KeyModeInput{}, err
	}
	return in, nil
}

func defaultKeyAuthTest(ctx context.Context, in KeyModeInput) error {
	expanded, err := platform.ExpandPath(strings.TrimSpace(in.IdentityFile))
	if err != nil {
		return err
	}
	return sshclient.TestKeyAuth(ctx, sshclient.KeyAuthOpts{
		Host:         strings.TrimSpace(in.HostName),
		Port:         effectivePort(in.Port),
		User:         strings.TrimSpace(in.User),
		IdentityFile: expanded,
	})
}
