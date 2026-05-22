package wizard

import (
	"errors"
	"strings"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

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
