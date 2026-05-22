package wizard

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

// passwordAuthTestFn 用于测试连接（单测可注入 mock）。
type passwordAuthTestFn func(ctx context.Context, in PasswordModeInput) error

// revealState 控制堆叠表单中已「解锁」展示的分组数量，避免未填项提前出现、已填项消失。
type revealState struct {
	n int
}

func (r *revealState) showThrough(index int) {
	if index+1 > r.n {
		r.n = index + 1
	}
}

func hideUntilRevealed(index int, r *revealState) func() bool {
	return func() bool { return index >= r.n }
}

// collectPasswordModeInput 用单个堆叠表单逐项收集；已填项保留在屏幕上，密码回车后测试连接。
func collectPasswordModeInput(ctx context.Context, testAuth passwordAuthTestFn) (PasswordModeInput, error) {
	if testAuth == nil {
		testAuth = defaultPasswordAuthTest
	}
	var in PasswordModeInput
	reveal := &revealState{n: 1}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("服务器 IP 或域名").
				Value(&in.HostName).
				Validate(func(s string) error {
					if err := nonEmpty("不能为空")(s); err != nil {
						return err
					}
					reveal.showThrough(1)
					return nil
				}),
		).WithHideFunc(hideUntilRevealed(0, reveal)),
		huh.NewGroup(
			huh.NewInput().
				Title("SSH 端口").
				Description("回车默认使用 22 端口").
				Placeholder("22").
				Value(&in.Port).
				Validate(func(string) error {
					reveal.showThrough(2)
					return nil
				}),
		).WithHideFunc(hideUntilRevealed(1, reveal)),
		huh.NewGroup(
			huh.NewInput().
				Title("SSH 用户名").
				Value(&in.User).
				Validate(func(s string) error {
					if err := nonEmpty("不能为空")(s); err != nil {
						return err
					}
					reveal.showThrough(3)
					return nil
				}),
		).WithHideFunc(hideUntilRevealed(2, reveal)),
		huh.NewGroup(
			huh.NewInput().
				Title("SSH 密码").
				EchoMode(huh.EchoModePassword).
				Value(&in.Password).
				Validate(func(password string) error {
					if err := passwordConnectionValidate(ctx, &in, testAuth)(password); err != nil {
						return err
					}
					reveal.showThrough(4)
					return nil
				}),
		).WithHideFunc(hideUntilRevealed(3, reveal)),
		huh.NewGroup(
			huh.NewInput().
				Title("Host 别名").
				Description("回车则根据 IP/域名自动生成").
				Value(&in.Alias),
		).WithHideFunc(hideUntilRevealed(4, reveal)),
	).WithLayout(huh.LayoutStack)

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

// passwordConnectionValidate 在密码项回车时测试 SSH 连接；失败则停留在该项并保留上方已填内容。
func passwordConnectionValidate(ctx context.Context, in *PasswordModeInput, testAuth passwordAuthTestFn) func(string) error {
	return func(password string) error {
		if strings.TrimSpace(password) == "" {
			return errors.New("不能为空")
		}
		in.Password = strings.TrimSpace(password)

		reportProgress("正在测试连接…")
		err := testAuth(ctx, *in)
		if err != nil {
			return errors.New(connectionTestFailureMessage(err))
		}
		fmt.Fprintf(progressOut, "连接成功\n")
		return nil
	}
}

func connectionTestFailureMessage(err error) string {
	if errors.Is(err, sshclient.ErrDeployAuthFailed) {
		return "用户名或密码不正确，请重新输入"
	}
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "sshclient: deploy failed: ")
	return msg
}

func effectivePort(port string) string {
	if strings.TrimSpace(port) == "" {
		return "22"
	}
	return strings.TrimSpace(port)
}
