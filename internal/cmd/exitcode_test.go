package cmd

import (
	"testing"

	"github.com/fuckssh/fuckssh/internal/sshclient"
)

func TestExitCode_sshNotFound(t *testing.T) {
	if got := ExitCode(sshclient.ErrSSHNotFound); got != 5 {
		t.Errorf("ssh not found = %d, want 5", got)
	}
}
