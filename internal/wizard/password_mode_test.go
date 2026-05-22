package wizard

import (
	"context"
	"errors"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/sshclient"
)

func TestPasswordMode_validateRequiresPassword(t *testing.T) {
	_, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "203.0.113.1",
		User:     "root",
	})
	if err == nil {
		t.Fatal("expected error without password")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestPasswordMode_defaultAlgorithmEd25519(t *testing.T) {
	in := PasswordModeInput{
		HostName: "203.0.113.1",
		User:     "root",
		Password: "secret",
	}
	out, err := finalizePasswordModeInput(in)
	if err != nil {
		t.Fatalf("finalizePasswordModeInput: %v", err)
	}
	if out.Algorithm != AlgorithmEd25519 {
		t.Errorf("Algorithm = %q, want %q", out.Algorithm, AlgorithmEd25519)
	}
}

func TestPasswordMode_defaultPort22(t *testing.T) {
	out, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "1.2.3.4",
		User:     "ubuntu",
		Password: "pw",
	})
	if err != nil {
		t.Fatalf("finalizePasswordModeInput: %v", err)
	}
	if out.Port != "22" {
		t.Errorf("Port = %q, want 22", out.Port)
	}
}

func TestPasswordMode_order_backupBeforeWrite(t *testing.T) {
	var order []string
	cfg := t.TempDir() + "/config"

	deps := passwordFlowDeps{
		backup: func(path string) (string, error) {
			order = append(order, "backup")
			return cfg + ".bak", nil
		},
		writeKeys: func(sshDir, alias string) (privPath, pubLine string, err error) {
			order = append(order, "keys")
			return sshDir + "/id_ed25519_test", "ssh-ed25519 AAAATEST user@host\n", nil
		},
		appendHost: func(path string, entry config.HostEntry) error {
			order = append(order, "append")
			return nil
		},
		deploy: func(ctx context.Context, opts sshclient.DeployOpts) error {
			order = append(order, "deploy")
			return nil
		},
	}

	in, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "203.0.113.50",
		User:     "ubuntu",
		Password: "pw",
		Alias:    "my-vps",
	})
	if err != nil {
		t.Fatalf("finalizePasswordModeInput: %v", err)
	}

	_, _, err = executePasswordFlow(context.Background(), in, cfg, deps)
	if err != nil {
		t.Fatalf("executePasswordFlow: %v", err)
	}

	want := []string{"backup", "keys", "append", "deploy"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i, step := range want {
		if order[i] != step {
			t.Errorf("step %d = %q, want %q (full order %v)", i, order[i], step, order)
		}
	}
}
