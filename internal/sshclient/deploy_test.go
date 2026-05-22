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

func TestTestPasswordAuth_authFailed_returnsTypedError(t *testing.T) {
	prev := dialSSH
	defer func() { dialSSH = prev }()

	dialSSH = func(ctx context.Context, opts DeployOpts) (sshClient, error) {
		return nil, errors.New("ssh: unable to authenticate, attempted methods [none password]")
	}

	err := TestPasswordAuth(context.Background(), DeployOpts{
		Host:     "203.0.113.1",
		Port:     "22",
		User:     "root",
		Password: "wrong",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrDeployAuthFailed) {
		t.Errorf("error = %v, want ErrDeployAuthFailed", err)
	}
}

func TestTestPasswordAuth_success_closesClient(t *testing.T) {
	prev := dialSSH
	defer func() { dialSSH = prev }()

	dialSSH = func(ctx context.Context, opts DeployOpts) (sshClient, error) {
		return &fakeSSHClient{}, nil
	}

	if err := TestPasswordAuth(context.Background(), DeployOpts{
		Host: "203.0.113.1", Port: "22", User: "root", Password: "ok",
	}); err != nil {
		t.Fatalf("TestPasswordAuth: %v", err)
	}
}

type fakeSSHClient struct {
	written map[string][]byte
}

func (f *fakeSSHClient) RunSession(cmd string) (string, string, int, error) {
	return "", "", 0, nil
}

func (f *fakeSSHClient) WriteAuthorizedKeys(content []byte) error {
	if f.written == nil {
		f.written = make(map[string][]byte)
	}
	f.written["authorized_keys"] = append([]byte(nil), content...)
	return nil
}

func (f *fakeSSHClient) Close() error { return nil }

func TestDeployPublicKey_usesWriteFileNotBase64(t *testing.T) {
	prev := dialSSH
	defer func() { dialSSH = prev }()

	fake := &fakeSSHClient{}
	dialSSH = func(ctx context.Context, opts DeployOpts) (sshClient, error) {
		return fake, nil
	}

	pub := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAInewkey comment\n"
	err := DeployPublicKey(context.Background(), DeployOpts{
		Host: "203.0.113.1", Port: "22", User: "root", Password: "pw",
		PublicLine: pub,
	})
	if err != nil {
		t.Fatalf("DeployPublicKey: %v", err)
	}
	got := fake.written["authorized_keys"]
	if !strings.Contains(string(got), strings.TrimSpace(pub)) {
		t.Errorf("authorized_keys = %q, want pub line", got)
	}
}

func TestDeployPublicKey_notWritable_returnsGuidance(t *testing.T) {
	prev := dialSSH
	defer func() { dialSSH = prev }()

	const absPath = "/home/boco/.ssh/authorized_keys"
	dialSSH = func(ctx context.Context, opts DeployOpts) (sshClient, error) {
		return &notWritableFake{path: absPath}, nil
	}

	err := DeployPublicKey(context.Background(), DeployOpts{
		Host: "10.12.2.220", Port: "22", User: "boco", Password: "pw",
		PublicLine: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAInewkey comment\n",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var detail *AuthorizedKeysNotWritableError
	if !errors.As(err, &detail) {
		t.Fatalf("error = %v, want *AuthorizedKeysNotWritableError", err)
	}
	if detail.User != "boco" || detail.AbsPath != absPath {
		t.Fatalf("detail = %+v", detail)
	}
	if !strings.Contains(detail.SSHCommand, "boco@10.12.2.220") {
		t.Errorf("SSHCommand = %q", detail.SSHCommand)
	}
	if !strings.Contains(detail.FixCommand, "sudo chown boco:boco") {
		t.Errorf("FixCommand = %q", detail.FixCommand)
	}
	if !strings.Contains(detail.FixCommand, "sudo chmod 600") {
		t.Errorf("FixCommand = %q", detail.FixCommand)
	}
}

type notWritableFake struct {
	path string
}

func (n *notWritableFake) RunSession(cmd string) (string, string, int, error) {
	switch {
	case strings.Contains(cmd, "readlink -f"), strings.Contains(cmd, `[ ! -e "$f" ]`):
		return n.path + "\n", "", 0, nil
	case strings.Contains(cmd, "test -w"):
		return "no\n", "", 0, nil
	case strings.Contains(cmd, "mkdir -p"), strings.Contains(cmd, "cat ~/.ssh/authorized_keys"):
		return "", "", 0, nil
	default:
		return "", "", 0, nil
	}
}

func (n *notWritableFake) WriteAuthorizedKeys([]byte) error {
	return errors.New("write authorized_keys: Process exited with status 2")
}

func (n *notWritableFake) Close() error { return nil }

func TestIsAuthError(t *testing.T) {
	if !isAuthError(errors.New("ssh: unable to authenticate")) {
		t.Error("want auth error")
	}
	if isAuthError(errors.New("connection refused")) {
		t.Error("connection refused is not auth error")
	}
}
