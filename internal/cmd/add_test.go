package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/sshclient"
)

func TestAddCmd_warnsWhenSSHMissing(t *testing.T) {
	restore := stubCheckSSH(func() (string, error) {
		return "", sshclient.ErrSSHNotFound
	})
	defer restore()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"add"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Execute: want error (wizard not implemented)")
	}
	if !strings.Contains(stderr.String(), "警告") {
		t.Errorf("stderr = %q, want warning", stderr.String())
	}
	if !strings.Contains(stderr.String(), "ssh") {
		t.Errorf("stderr = %q, want ssh mention", stderr.String())
	}
}

func TestAddCmd_noWarningWhenSSHPresent(t *testing.T) {
	restore := stubCheckSSH(func() (string, error) {
		return "/usr/bin/ssh", nil
	})
	defer restore()

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"add"})

	_ = rootCmd.Execute()
	if strings.Contains(stderr.String(), "警告") {
		t.Errorf("stderr should be quiet when ssh found, got: %q", stderr.String())
	}
}

func stubCheckSSH(fn func() (string, error)) func() {
	prev := checkSSHFn
	checkSSHFn = fn
	return func() { checkSSHFn = prev }
}
