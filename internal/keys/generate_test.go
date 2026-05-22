package keys

import (
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGenerateEd25519_producesParseablePrivateKey(t *testing.T) {
	kp, err := GenerateEd25519()
	if err != nil {
		t.Fatalf("GenerateEd25519: %v", err)
	}

	if !strings.Contains(string(kp.PrivatePEM), "OPENSSH PRIVATE KEY") {
		t.Fatalf("private PEM missing OPENSSH header:\n%s", kp.PrivatePEM)
	}

	_, err = ssh.ParsePrivateKey(kp.PrivatePEM)
	if err != nil {
		t.Fatalf("ssh.ParsePrivateKey: %v", err)
	}
}

func TestGenerateEd25519_publicLineStartsWithSshEd25519(t *testing.T) {
	kp, err := GenerateEd25519()
	if err != nil {
		t.Fatalf("GenerateEd25519: %v", err)
	}

	line := strings.TrimSpace(kp.PublicLine)
	if !strings.HasPrefix(line, "ssh-ed25519 ") {
		t.Fatalf("PublicLine = %q, want ssh-ed25519 prefix", kp.PublicLine)
	}
}

func TestGenerateEd25519_eachCallUnique(t *testing.T) {
	a, err := GenerateEd25519()
	if err != nil {
		t.Fatalf("first GenerateEd25519: %v", err)
	}
	b, err := GenerateEd25519()
	if err != nil {
		t.Fatalf("second GenerateEd25519: %v", err)
	}

	if string(a.PrivatePEM) == string(b.PrivatePEM) {
		t.Fatal("two calls produced identical private PEM")
	}
	if a.PublicLine == b.PublicLine {
		t.Fatal("two calls produced identical public line")
	}
}
