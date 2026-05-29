package wizard

// auth_connect.go 合并连接测试函数与连接错误消息转换。
//
// 设计说明：
// - passwordAuthTestFn / keyAuthTestFn 是测连策略的函数类型
// - connectionTestFailureMessage 将底层 SSH 错误转为用户可读文案
// - 两者都属于"连接测试"领域，合并后减少 1 个碎片文件

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
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

// --- 连接错误消息转换 ---

// connectionTestFailureMessage 将底层 SSH/网络错误转为用户可读的行内提示（不含原始 dial 文案）。
func connectionTestFailureMessage(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, sshclient.ErrDeployAuthFailed) {
		return i18n.T(i18n.KeyWizardConnFailInline)
	}
	if errors.Is(err, keys.ErrPassphraseNotSupported) {
		return i18n.T(i18n.KeyWizardPassphraseNA)
	}

	msg := strings.ToLower(err.Error())
	// 去掉包装前缀，避免露出技术栈信息。
	for _, prefix := range []string{
		"sshclient: deploy failed: ",
		"dial tcp ",
		"connectex: ",
	} {
		if idx := strings.Index(msg, prefix); idx >= 0 {
			msg = msg[idx+len(prefix):]
		}
	}

	switch {
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "actively refused"),
		strings.Contains(msg, "no connection could be made"):
		return i18n.T(i18n.KeyWizardConnRefused)
	case strings.Contains(msg, "i/o timeout"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "deadline exceeded"):
		return i18n.T(i18n.KeyWizardConnTimeout)
	case strings.Contains(msg, "no route to host"),
		strings.Contains(msg, "network is unreachable"),
		strings.Contains(msg, "no such host"),
		strings.Contains(msg, "knownhosts"):
		return i18n.T(i18n.KeyWizardConnUnreachable)
	case strings.Contains(msg, "unable to authenticate"),
		strings.Contains(msg, "authentication failed"),
		strings.Contains(msg, "permission denied"):
		return i18n.T(i18n.KeyWizardConnFailInline)
	default:
		return i18n.T(i18n.KeyWizardConnFailGeneric)
	}
}
