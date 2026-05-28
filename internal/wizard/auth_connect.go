package wizard

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

// passwordAuthTestFn 用于密码连接测试（单测可注入 mock）。
type passwordAuthTestFn func(ctx context.Context, in PasswordModeInput) error

// keyAuthTestFn 用于密钥连接测试（单测可注入 mock）。
type keyAuthTestFn func(ctx context.Context, in KeyModeInput) error

func defaultPasswordAuthTest(ctx context.Context, in PasswordModeInput) error {
	return sshclient.TestPasswordAuth(ctx, sshclient.DeployOpts{
		Host:     strings.TrimSpace(in.HostName),
		Port:     effectivePort(in.Port),
		User:     effectiveUser(in.User),
		Password: in.Password,
	})
}

func defaultKeyAuthTest(ctx context.Context, in KeyModeInput) error {
	expanded, err := platform.ExpandPath(strings.TrimSpace(in.IdentityFile))
	if err != nil {
		return err
	}
	return sshclient.TestKeyAuth(ctx, sshclient.KeyAuthOpts{
		Host:         strings.TrimSpace(in.HostName),
		Port:         effectivePort(in.Port),
		User:         effectiveUser(in.User),
		IdentityFile: expanded,
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
