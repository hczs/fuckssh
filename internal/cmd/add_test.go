package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/sshclient"
	"github.com/fuckssh/fuckssh/internal/wizard"
)

func TestAddCmd_warnsWhenSSHMissing(t *testing.T) {
	restoreSSH := stubCheckSSH(func() (string, error) {
		return "", sshclient.ErrSSHNotFound
	})
	defer restoreSSH()
	restoreWizard := stubRunWizard(func() (*wizard.WizardResult, error) {
		return nil, wizard.ErrInvalidInput
	})
	defer restoreWizard()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"add"})

	_ = rootCmd.Execute()
	if !strings.Contains(stderr.String(), "警告") {
		t.Errorf("stderr = %q, want warning", stderr.String())
	}
}

func TestAddCmd_noWarningWhenSSHPresent(t *testing.T) {
	restoreSSH := stubCheckSSH(func() (string, error) {
		return "/usr/bin/ssh", nil
	})
	defer restoreSSH()
	restoreWizard := stubRunWizard(func() (*wizard.WizardResult, error) {
		return nil, wizard.ErrInvalidInput
	})
	defer restoreWizard()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"add"})

	_ = rootCmd.Execute()
	if strings.Contains(stderr.String(), "警告") {
		t.Errorf("stderr should be quiet when ssh found, got: %q", stderr.String())
	}
}

func TestAdd_keyMode_integrationWithTempConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	key := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(key, []byte("fake-key"), 0o600); err != nil {
		t.Fatal(err)
	}

	restoreSSH := stubCheckSSH(func() (string, error) {
		return "/usr/bin/ssh", nil
	})
	defer restoreSSH()
	restoreWizard := stubRunWizard(func() (*wizard.WizardResult, error) {
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
	rootCmd.SetArgs([]string{"add", "--config", cfg})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "ssh test-vps") {
		t.Errorf("stdout = %q, want ssh hint", out)
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
	if e.IdentityFile != key {
		t.Errorf("IdentityFile = %q, want %q", e.IdentityFile, key)
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

func stubRunWizard(fn func() (*wizard.WizardResult, error)) func() {
	prev := runWizardFn
	runWizardFn = fn
	return func() { runWizardFn = prev }
}
