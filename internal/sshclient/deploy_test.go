package sshclient

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDeployPublicKey_buildsAuthorizedKeysLine(t *testing.T) {
	existing := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB existing-key\n"
	pub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAInewkey comment\n"

	got := appendAuthorizedKey(existing, pub)
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("result should end with newline, got %q", got)
	}
	if !strings.Contains(got, strings.TrimSpace(pub)) {
		t.Errorf("result missing new key:\n%s", got)
	}
	if !strings.Contains(got, "existing-key") {
		t.Errorf("result should preserve existing lines:\n%s", got)
	}

	again := appendAuthorizedKey(got, pub)
	if again != got {
		t.Errorf("duplicate append changed content:\n%s", again)
	}
}

func TestDeployPublicKey_buildsFromEmpty(t *testing.T) {
	pub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAInewkey\n"
	got := appendAuthorizedKey("", pub)
	if strings.TrimSpace(got) != strings.TrimSpace(pub) {
		t.Errorf("got %q, want trimmed pub line", got)
	}
}

func TestDeployPublicKey_authFailed_returnsTypedError(t *testing.T) {
	prev := dialSSH
	defer func() { dialSSH = prev }()

	dialSSH = func(ctx context.Context, opts DeployOpts) (sshClient, error) {
		return nil, errors.New("ssh: unable to authenticate, attempted methods [none password]")
	}

	err := DeployPublicKey(context.Background(), DeployOpts{
		Host:       "203.0.113.1",
		Port:       "22",
		User:       "root",
		Password:   "secret",
		PublicLine: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test\n",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrDeployFailed) {
		t.Errorf("error = %v, want ErrDeployFailed", err)
	}
	if !errors.Is(err, ErrDeployAuthFailed) {
		t.Errorf("error = %v, want ErrDeployAuthFailed", err)
	}
}

func TestIsAuthError(t *testing.T) {
	if !isAuthError(errors.New("ssh: unable to authenticate")) {
		t.Error("want auth error")
	}
	if isAuthError(errors.New("connection refused")) {
		t.Error("connection refused is not auth error")
	}
}
