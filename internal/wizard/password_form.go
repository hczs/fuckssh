package wizard

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

// passwordAuthTestFn 用于测试连接（单测可注入 mock）。
type passwordAuthTestFn func(ctx context.Context, in PasswordModeInput) error

// collectPasswordModeInput 用单个堆叠表单逐项收集；draft 非空时预填并恢复可见步骤（确认页返回修改时用）。
func collectPasswordModeInput(ctx context.Context, configPath string, testAuth passwordAuthTestFn, draft *PasswordModeInput) (PasswordModeInput, error) {
	if testAuth == nil {
		testAuth = defaultPasswordAuthTest
	}
	var in PasswordModeInput
	if draft != nil {
		in = *draft
	}
	reveal := &revealState{n: 1}
	seedReveal(reveal, in)
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
			NewPasswordTestField(ctx, &in, testAuth,
				func() { reveal.showThrough(4) },
				func() { reveal.lockThrough(3) },
			).
				Title(stepTitle(5, i18n.KeyWizardPassword)).
				Key("password").
				Value(&in.Password),
		).WithHideFunc(hideUntilRevealed(3, reveal)),
		huh.NewGroup(
			NewAliasField(configPath, &in.HostName).
				Title(stepTitle(6, i18n.KeyWizardAlias)).
				Key("alias").
				Value(&in.Alias),
		).WithHideFunc(hideUntilRevealed(4, reveal)),
	).WithLayout(huh.LayoutStack).WithShowErrors(false)

	if err := form.Run(); err != nil {
		return PasswordModeInput{}, err
	}
	return in, nil
}

func defaultPasswordAuthTest(ctx context.Context, in PasswordModeInput) error {
	return sshclient.TestPasswordAuth(ctx, sshclient.DeployOpts{
		Host:     strings.TrimSpace(in.HostName),
		Port:     effectivePort(in.Port),
		User:     strings.TrimSpace(in.User),
		Password: in.Password,
	})
}

// testPasswordConnection 执行密码测连并返回耗时（供单测与自定义字段共用）。
func testPasswordConnection(ctx context.Context, in *PasswordModeInput, password string, testAuth passwordAuthTestFn) (time.Duration, error) {
	if strings.TrimSpace(password) == "" {
		return 0, errors.New(i18n.T(i18n.KeyWizardErrEmpty))
	}
	in.Password = strings.TrimSpace(password)
	start := time.Now()
	err := testAuth(ctx, *in)
	return time.Since(start), err
}
