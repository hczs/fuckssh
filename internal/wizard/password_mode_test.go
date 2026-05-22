package wizard

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

func TestPasswordMode_reportsProgressSteps(t *testing.T) {
	var steps []string
	cfg := t.TempDir() + "/config"
	deps := passwordFlowDeps{
		backup: func(path string) (string, error) {
			return cfg + ".bak", nil
		},
		writeKeys: func(sshDir, alias string) (string, string, error) {
			return sshDir + "/id_ed25519_test", "ssh-ed25519 AAAATEST\n", nil
		},
		appendHost: func(path string, entry config.HostEntry) error { return nil },
		deploy:     func(ctx context.Context, opts sshclient.DeployOpts) error { return nil },
		onProgress: func(msg string) { steps = append(steps, msg) },
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
	if _, _, err := executePasswordFlow(context.Background(), in, cfg, deps); err != nil {
		t.Fatalf("executePasswordFlow: %v", err)
	}
	want := []string{
		"正在备份 SSH config…",
		"正在生成 Ed25519 密钥…",
		"正在写入 SSH config…",
		"正在连接服务器并部署公钥…",
	}
	if len(steps) != len(want) {
		t.Fatalf("progress steps = %v, want %v", steps, want)
	}
	for i, s := range want {
		if steps[i] != s {
			t.Errorf("step %d = %q, want %q", i, steps[i], s)
		}
	}
}

func TestFormatPasswordDeployError_authMessage(t *testing.T) {
	err := formatPasswordDeployError(
		fmt.Errorf("%w: %w", sshclient.ErrDeployFailed, sshclient.ErrDeployAuthFailed),
		"/tmp/config.bak",
	)
	msg := err.Error()
	if !strings.Contains(msg, "SSH 密码认证失败") {
		t.Errorf("message = %q, want auth hint", msg)
	}
	if !strings.Contains(msg, "/tmp/config.bak") {
		t.Errorf("message = %q, want backup path", msg)
	}
	if !errors.Is(err, sshclient.ErrDeployAuthFailed) {
		t.Errorf("error should wrap ErrDeployAuthFailed: %v", err)
	}
}

func TestPasswordConnectionValidate_retriesUntilSuccess(t *testing.T) {
	var attempts int
	authErr := fmt.Errorf("%w: %w", sshclient.ErrDeployFailed, sshclient.ErrDeployAuthFailed)
	in := PasswordModeInput{
		HostName: "203.0.113.50",
		User:     "ubuntu",
		Port:     "22",
	}
	validate := passwordConnectionValidate(context.Background(), &in, func(ctx context.Context, got PasswordModeInput) error {
		attempts++
		if attempts < 2 {
			return authErr
		}
		return nil
	})

	err := validate("wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !strings.Contains(err.Error(), "用户名或密码") {
		t.Errorf("error = %q, want auth hint", err.Error())
	}
	if err := validate("correct"); err != nil {
		t.Fatalf("validate correct password: %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestConnectionTestFailureMessage_auth(t *testing.T) {
	err := fmt.Errorf("%w: %w", sshclient.ErrDeployFailed, sshclient.ErrDeployAuthFailed)
	msg := connectionTestFailureMessage(err)
	if !strings.Contains(msg, "用户名或密码") {
		t.Errorf("message = %q", msg)
	}
}

func TestEffectivePort_defaultsTo22(t *testing.T) {
	if got := effectivePort(""); got != "22" {
		t.Errorf("effectivePort(\"\") = %q, want 22", got)
	}
	if got := effectivePort(" 2222 "); got != "2222" {
		t.Errorf("effectivePort = %q, want 2222", got)
	}
}
