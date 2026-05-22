package wizard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

// keyAuthTestFn 用于密钥连接测试（单测可注入 mock）。
type keyAuthTestFn func(ctx context.Context, in KeyModeInput) error

// collectKeyModeInput 用堆叠表单逐项收集；draft 非空时预填（确认页返回修改时用）。
func collectKeyModeInput(ctx context.Context, testAuth keyAuthTestFn, draft *KeyModeInput) (KeyModeInput, error) {
	if testAuth == nil {
		testAuth = defaultKeyAuthTest
	}
	var in KeyModeInput
	if draft != nil {
		in = *draft
	}
	var keyFeedback connFeedback
	reveal := &revealState{n: 1}
	seedRevealKey(reveal, in)
	emptyMsg := i18n.T(i18n.KeyWizardErrEmpty)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(stepTitle(2, i18n.KeyWizardHostIP)).
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
				Validate(func(string) error {
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
			huh.NewInput().
				Title(stepTitle(5, i18n.KeyWizardIdentityFile)).
				Description(i18n.T(i18n.KeyWizardIdentityDesc)).
				Value(&in.IdentityFile).
				Validate(func(path string) error {
					if err := keyIdentityValidate(ctx, &in, testAuth, &keyFeedback)(path); err != nil {
						return err
					}
					reveal.showThrough(4)
					return nil
				}),
		).WithHideFunc(hideUntilRevealed(3, reveal)),
		huh.NewGroup(
			huh.NewInput().
				Title(stepTitle(6, i18n.KeyWizardAlias)).
				DescriptionFunc(func() string { return aliasDescription(&in.HostName) }, &in.HostName).
				Value(&in.Alias),
		).WithHideFunc(hideUntilRevealed(4, reveal)),
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

func keyIdentityValidate(ctx context.Context, in *KeyModeInput, testAuth keyAuthTestFn, feedback *connFeedback) func(string) error {
	return func(path string) error {
		if strings.TrimSpace(path) == "" {
			return errors.New(i18n.T(i18n.KeyWizardErrEmpty))
		}
		in.IdentityFile = strings.TrimSpace(path)

		expanded, err := platform.ExpandPath(in.IdentityFile)
		if err != nil {
			return err
		}
		if _, err := os.Stat(expanded); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%w: %s", ErrInvalidInput, expanded)
			}
			return err
		}
		in.IdentityFile = expanded

		reportProgress(i18n.T(i18n.KeyWizardTestingConn))
		feedback.setTesting()
		err = testAuth(ctx, *in)
		if err != nil {
			feedback.message = ""
			return errors.New(keyConnectionTestFailureMessage(err))
		}
		feedback.setOK()
		fmt.Fprintf(progressOut, "%s\n", i18n.T(i18n.KeyWizardConnOK))
		return nil
	}
}

func keyConnectionTestFailureMessage(err error) string {
	if errors.Is(err, keys.ErrPassphraseNotSupported) {
		return i18n.T(i18n.KeyWizardPassphraseNA)
	}
	if errors.Is(err, sshclient.ErrDeployAuthFailed) {
		return i18n.T(i18n.KeyWizardKeyAuthFailed)
	}
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "sshclient: deploy failed: ")
	return msg
}
