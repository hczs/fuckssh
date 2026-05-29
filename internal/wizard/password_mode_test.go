package wizard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
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
		"wizard.progress_backup",
		"wizard.progress_keygen",
		"wizard.progress_write_cfg",
		"wizard.progress_deploy_pubkey",
	}
	if len(steps) != len(want) {
		t.Fatalf("progress steps = %v, want %v", steps, want)
	}
	for i, key := range want {
		expected := i18n.T(key)
		if steps[i] != expected {
			t.Errorf("step %d = %q, want %q (key %q)", i, steps[i], expected, key)
		}
	}
}

func TestFormatPasswordDeployError_deployAbortedRolledBack(t *testing.T) {
	err := formatPasswordDeployError(ErrDeployAborted, "/tmp/config.bak", true)
	if err == nil {
		t.Fatal("expected error")
	}
	// formatPasswordDeployError 用 i18n 格式化后不再包裹 ErrDeployAborted，
	// 因此只验证错误消息非空（文案由 i18n 决定，不测具体文本）。
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestPasswordMode_retryDeployAfterPermissionFix(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}

	deployCalls := 0
	perm := &sshclient.AuthorizedKeysNotWritableError{
		User: "boco", AbsPath: "/home/boco/.ssh/authorized_keys",
		SSHCommand: "boco@10.12.2.220",
		FixCommand: "sudo chown boco:boco /home/boco/.ssh/authorized_keys",
	}
	deps := passwordFlowDeps{
		backup: func(string) (string, error) { return "", nil },
		writeKeys: func(dir, alias string) (string, string, error) {
			priv := filepath.Join(dir, "k")
			if err := os.WriteFile(priv, []byte("priv"), 0o600); err != nil {
				return "", "", err
			}
			return priv, "ssh-ed25519 AAAATEST\n", nil
		},
		appendHost: func(string, config.HostEntry) error { return nil },
		deploy: func(context.Context, sshclient.DeployOpts) error {
			deployCalls++
			if deployCalls == 1 {
				return fmt.Errorf("%w: %w", sshclient.ErrDeployFailed, perm)
			}
			return nil
		},
		promptPermissionFix: func(context.Context, *sshclient.AuthorizedKeysNotWritableError) error {
			return nil
		},
	}

	in, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "10.12.2.220", User: "boco", Password: "pw", Alias: "vps1",
	})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}

	_, _, err = executePasswordFlow(context.Background(), in, cfg, deps)
	if err != nil {
		t.Fatalf("executePasswordFlow: %v", err)
	}
	if deployCalls != 2 {
		t.Errorf("deployCalls = %d, want 2", deployCalls)
	}
}

func TestFormatPasswordDeployError_authMessageRolledBack(t *testing.T) {
	err := formatPasswordDeployError(
		fmt.Errorf("%w: %w", sshclient.ErrDeployFailed, sshclient.ErrDeployAuthFailed),
		"/tmp/config.bak",
		true,
	)
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	if !errors.Is(err, sshclient.ErrDeployAuthFailed) {
		t.Errorf("error should wrap ErrDeployAuthFailed: %v", err)
	}
}

func TestPasswordMode_duplicateAliasPreservesConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	orig := "Host dup\n    HostName 1.2.3.4\n    User root\n    IdentityFile /tmp/k\n"
	if err := os.WriteFile(cfg, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}

	deps := passwordFlowDeps{
		backup:     func(string) (string, error) { t.Fatal("backup should not run"); return "", nil },
		writeKeys:  func(string, string) (string, string, error) { return "", "", nil },
		appendHost: func(string, config.HostEntry) error { return nil },
		deploy:     func(context.Context, sshclient.DeployOpts) error { return nil },
	}
	in, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "203.0.113.50",
		User:     "ubuntu",
		Password: "pw",
		Alias:    "dup",
	})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}

	_, _, err = executePasswordFlow(context.Background(), in, cfg, deps)
	if err == nil {
		t.Fatal("expected duplicate alias error")
	}
	if !errors.Is(err, config.ErrHostExists) {
		t.Errorf("error = %v, want ErrHostExists", err)
	}

	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != orig {
		t.Errorf("config was modified on duplicate error:\n%s", got)
	}
}

func TestPasswordMode_rollbackAfterKeysFailure(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	orig := "Host old\n    HostName 10.0.0.1\n    User root\n    IdentityFile /tmp/old\n"
	if err := os.WriteFile(cfg, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}

	deps := passwordFlowDeps{
		backup: config.Backup,
		writeKeys: func(string, string) (string, string, error) {
			return "", "", errors.New("key write failed")
		},
		appendHost: func(string, config.HostEntry) error { return nil },
		deploy:     func(context.Context, sshclient.DeployOpts) error { return nil },
	}
	in, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "203.0.113.50",
		User:     "ubuntu",
		Password: "pw",
		Alias:    "new-vps",
	})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}

	_, _, err = executePasswordFlow(context.Background(), in, cfg, deps)
	if err == nil {
		t.Fatal("expected key error")
	}

	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != orig {
		t.Errorf("config should be restored from backup, got:\n%s", got)
	}
}

func TestPasswordMode_rejectsDuplicateAliasBeforeSideEffects(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	content := "Host dup\n    HostName 1.2.3.4\n    User root\n    IdentityFile /tmp/k\n"
	if err := os.WriteFile(cfg, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var wroteKeys bool
	deps := passwordFlowDeps{
		backup: func(string) (string, error) {
			t.Fatal("backup should not run when alias exists")
			return "", nil
		},
		writeKeys: func(string, string) (string, string, error) {
			wroteKeys = true
			return "", "", nil
		},
		appendHost: func(string, config.HostEntry) error { return nil },
		deploy:     func(context.Context, sshclient.DeployOpts) error { return nil },
	}

	in, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "203.0.113.50",
		User:     "ubuntu",
		Password: "pw",
		Alias:    "dup",
	})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}

	_, _, err = executePasswordFlow(context.Background(), in, cfg, deps)
	if err == nil {
		t.Fatal("expected duplicate alias error")
	}
	if !errors.Is(err, config.ErrHostExists) {
		t.Errorf("error = %v, want ErrHostExists", err)
	}
	if wroteKeys {
		t.Error("keys should not be written")
	}
}

func TestPasswordMode_rollbackOnDeployFailure(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	orig := "Host old\n    HostName 10.0.0.1\n    User root\n    IdentityFile /tmp/old\n"
	if err := os.WriteFile(cfg, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}

	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}

	deps := passwordFlowDeps{
		backup: config.Backup,
		writeKeys: func(dir, alias string) (string, string, error) {
			priv := filepath.Join(dir, "id_ed25519_fuckssh_test")
			if err := os.WriteFile(priv, []byte("priv"), 0o600); err != nil {
				return "", "", err
			}
			if err := os.WriteFile(priv+".pub", []byte("pub\n"), 0o644); err != nil {
				return "", "", err
			}
			return priv, "ssh-ed25519 AAAATEST\n", nil
		},
		appendHost: config.AppendHost,
		deploy: func(context.Context, sshclient.DeployOpts) error {
			return sshclient.ErrDeployFailed
		},
	}

	in, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "203.0.113.50",
		User:     "ubuntu",
		Password: "pw",
		Alias:    "new-vps",
	})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}

	_, _, err = executePasswordFlow(context.Background(), in, cfg, deps)
	if err == nil {
		t.Fatal("expected deploy error")
	}

	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != orig {
		t.Errorf("config changed after rollback:\n%s", got)
	}
	priv := filepath.Join(sshDir, "id_ed25519_fuckssh_new_vps")
	if _, err := os.Stat(priv); !os.IsNotExist(err) {
		t.Errorf("priv key should be removed, err=%v", err)
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
	testAuth := func(ctx context.Context, got PasswordModeInput) error {
		attempts++
		if attempts < 2 {
			return authErr
		}
		return nil
	}

	_, err := testPasswordConnection(context.Background(), &in, "wrong", testAuth)
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if msg := connectionTestFailureMessage(err); msg == "" {
		t.Error("expected non-empty friendly hint for auth failure")
	}
	if _, err := testPasswordConnection(context.Background(), &in, "correct", testAuth); err != nil {
		t.Fatalf("correct password: %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
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
