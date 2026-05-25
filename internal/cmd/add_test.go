package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/sshclient"
	"github.com/fuckssh/fuckssh/internal/wizard"
)

func TestExitCode_deployFailed(t *testing.T) {
	err := fmt.Errorf("deploy: %w", sshclient.ErrDeployFailed)
	if got := ExitCode(err); got != 4 {
		t.Errorf("deploy failed = %d, want 4", got)
	}
}

func TestExitCode_deployAuthFailed(t *testing.T) {
	err := fmt.Errorf("SSH 密码认证失败: %w",
		fmt.Errorf("%w: %w", sshclient.ErrDeployFailed, sshclient.ErrDeployAuthFailed))
	if got := ExitCode(err); got != 4 {
		t.Errorf("auth deploy failed = %d, want 4", got)
	}
}

func TestAddCmd_abortsWhenSSHMissing(t *testing.T) {
	restoreSSH := stubCheckSSH(func() (string, error) {
		return "", sshclient.ErrSSHNotFound
	})
	defer restoreSSH()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	err := ExecuteWithArgs([]string{"add"})
	if err == nil {
		t.Fatal("Execute: want error when ssh missing")
	}
	if !errors.Is(err, sshclient.ErrSSHNotFound) {
		t.Errorf("err = %v, want ErrSSHNotFound", err)
	}
	if got := ExitCode(err); got != 5 {
		t.Errorf("exit code = %d, want 5", got)
	}
	if !strings.Contains(stderr.String(), "PATH") || !strings.Contains(stderr.String(), "ssh") {
		t.Errorf("stderr = %q, want warning and guide", stderr.String())
	}
}

func TestAddCmd_userCancelNoUsage(t *testing.T) {
	restoreSSH := stubCheckSSH(func() (string, error) {
		return "/usr/bin/ssh", nil
	})
	defer restoreSSH()
	restoreWizard := stubRunWizard(func(string) (*wizard.WizardResult, error) {
		return nil, wizard.UserCancelled()
	})
	defer restoreWizard()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	err := ExecuteWithArgs([]string{"add"})
	if err == nil {
		t.Fatal("want cancel error")
	}
	if !wizard.IsCancelled(err) {
		t.Fatalf("err = %v, want cancelled", err)
	}
	out := stderr.String()
	if !strings.Contains(out, "已取消") {
		t.Errorf("stderr = %q, want cancel message", out)
	}
	for _, forbidden := range []string{"Usage:", "add [flags]", "Flags:"} {
		if strings.Contains(out, forbidden) {
			t.Errorf("stderr should not contain help %q, got: %q", forbidden, out)
		}
	}
}

func TestAddCmd_noWarningWhenSSHPresent(t *testing.T) {
	restoreSSH := stubCheckSSH(func() (string, error) {
		return "/usr/bin/ssh", nil
	})
	defer restoreSSH()
	restoreWizard := stubRunWizard(func(string) (*wizard.WizardResult, error) {
		return nil, wizard.ErrInvalidInput
	})
	defer restoreWizard()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	_ = ExecuteWithArgs([]string{"add"})
	if strings.Contains(stderr.String(), "PATH") && strings.Contains(stderr.String(), "未在") {
		t.Errorf("stderr should not contain ssh missing warning, got: %q", stderr.String())
	}
}

func setTestHome(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
	} else {
		t.Setenv("HOME", dir)
	}
}

func TestAdd_keyMode_integrationWithTempConfig(t *testing.T) {
	dir := t.TempDir()
	setTestHome(t, dir)
	cfg := filepath.Join(dir, "config")
	key := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(key, []byte("fake-key"), 0o600); err != nil {
		t.Fatal(err)
	}

	restoreSSH := stubCheckSSH(func() (string, error) {
		return "/usr/bin/ssh", nil
	})
	defer restoreSSH()
	restoreWizard := stubRunWizard(func(string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{
			Alias:        "test-vps",
			HostName:     "203.0.113.50",
			User:         "ubuntu",
			Port:         "22",
			IdentityFile: key,
		}, nil
	})
	defer restoreWizard()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	if err := ExecuteWithArgs([]string{"add", "--config", cfg}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	out := strings.TrimSpace(stdout.String())
	if out != "ssh test-vps" {
		t.Errorf("stdout = %q, want only ssh line", stdout.String())
	}
	if !strings.Contains(stderr.String(), "本次已完成") {
		t.Errorf("stderr = %q, want success summary", stderr.String())
	}
	if !strings.Contains(stderr.String(), "ssh test-vps") {
		t.Errorf("stderr = %q, want ssh command in summary", stderr.String())
	}

	entries, err := config.ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.Alias != "test-vps" || e.HostName != "203.0.113.50" || e.User != "ubuntu" {
		t.Errorf("entry = %+v", e)
	}
	privName, _ := keys.KeyPaths("test-vps")
	wantIdentity := "~/.ssh/keys/" + privName
	if e.IdentityFile != wantIdentity {
		t.Errorf("IdentityFile = %q, want %q", e.IdentityFile, wantIdentity)
	}
	copied := filepath.Join(dir, ".ssh", "keys", privName)
	if _, err := os.Stat(copied); err != nil {
		t.Errorf("copied key missing at %q: %v", copied, err)
	}
	if _, err := os.Stat(key); err != nil {
		t.Errorf("source key should remain: %v", err)
	}
}

func TestAdd_passwordMode_integrationSkipsSecondBackup(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	key := filepath.Join(dir, "id_ed25519_fuckssh_pw")

	restoreSSH := stubCheckSSH(func() (string, error) {
		return "/usr/bin/ssh", nil
	})
	defer restoreSSH()
	restoreWizard := stubRunWizard(func(string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{
			Alias:                "pw-vps",
			HostName:             "203.0.113.60",
			User:                 "root",
			Port:                 "22",
			IdentityFile:         key,
			PasswordFlowComplete: true,
			BackupPath:           cfg + ".fuckssh.bak.test",
		}, nil
	})
	defer restoreWizard()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	if err := ExecuteWithArgs([]string{"add", "--config", cfg}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if strings.TrimSpace(stdout.String()) != "ssh pw-vps" {
		t.Errorf("stdout = %q, want only ssh line", stdout.String())
	}
	if !strings.Contains(stderr.String(), "本次已完成") {
		t.Errorf("stderr = %q, want success summary", stderr.String())
	}
	if !strings.Contains(stderr.String(), "ssh pw-vps") {
		t.Errorf("stderr = %q, want ssh command in summary", stderr.String())
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "config.fuckssh.bak.*"))
	if len(matches) > 0 {
		t.Errorf("unexpected extra backup files: %v", matches)
	}
}

func TestExitCode_mapping(t *testing.T) {
	if got := ExitCode(wizard.ErrInvalidInput); got != 1 {
		t.Errorf("invalid input = %d, want 1", got)
	}
	if got := ExitCode(config.ErrHostExists); got != 1 {
		t.Errorf("host exists = %d, want 1", got)
	}
	pe := &config.ParseError{File: "c", Line: 1, Msg: "bad"}
	if got := ExitCode(pe); got != 2 {
		t.Errorf("parse = %d, want 2", got)
	}
	pathErr := &os.PathError{Op: "open", Path: "/x", Err: os.ErrPermission}
	if got := ExitCode(pathErr); got != 3 {
		t.Errorf("path = %d, want 3", got)
	}
}

func stubCheckSSH(fn func() (string, error)) func() {
	prev := checkSSHFn
	checkSSHFn = fn
	return func() { checkSSHFn = prev }
}

func stubRunWizard(fn func(string) (*wizard.WizardResult, error)) func() {
	prev := runWizardFn
	runWizardFn = fn
	return func() { runWizardFn = prev }
}
