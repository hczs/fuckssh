package wizard

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/sshclient"
)

func TestConnectionTestFailureMessage_auth(t *testing.T) {
	err := fmt.Errorf("%w: %w", sshclient.ErrDeployFailed, sshclient.ErrDeployAuthFailed)
	msg := connectionTestFailureMessage(err)
	if !strings.Contains(msg, "密码") {
		t.Errorf("message = %q", msg)
	}
}

func TestConnectionTestFailureMessage_refused(t *testing.T) {
	err := fmt.Errorf("%w: dial tcp 10.12.1.220:22: connectex: No connection could be made because the target machine actively refused it",
		sshclient.ErrDeployFailed)
	msg := connectionTestFailureMessage(err)
	if !strings.Contains(msg, "无法连接服务器") {
		t.Errorf("message = %q, want friendly refused hint", msg)
	}
	if strings.Contains(msg, "dial tcp") {
		t.Errorf("message should not contain raw dial error: %q", msg)
	}
}
