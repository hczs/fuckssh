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
	if msg == "" {
		t.Error("expected non-empty message for auth failure")
	}
}

func TestConnectionTestFailureMessage_refused(t *testing.T) {
	err := fmt.Errorf("%w: dial tcp 10.12.1.220:22: connectex: No connection could be made because the target machine actively refused it",
		sshclient.ErrDeployFailed)
	msg := connectionTestFailureMessage(err)
	if msg == "" {
		t.Error("expected non-empty message for connection refused")
	}
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "dial tcp") || strings.Contains(lower, "connectex:") {
		t.Errorf("message should not contain raw dial error: %q", msg)
	}
}
