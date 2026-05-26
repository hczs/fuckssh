package wizard

import (
	"os"
	"testing"
)

func TestAddInput_ToPasswordModeInput(t *testing.T) {
	in := AddInput{
		Alias:    "my-vps",
		HostName: "203.0.113.1",
		User:     "ubuntu",
		Port:     "2222",
		Mode:     ModePassword,
		Password: "secret",
		Remark:   "prod",
	}
	pw := in.ToPasswordModeInput()
	if pw.Alias != "my-vps" || pw.HostName != "203.0.113.1" || pw.Password != "secret" || pw.Port != "2222" {
		t.Fatalf("ToPasswordModeInput = %+v", pw)
	}
}

func TestClearCredentialsOnModeChange(t *testing.T) {
	in := AddInput{
		Mode:         ModePassword,
		Password:     "pw",
		IdentityFile: "/tmp/k",
		AuthTestOK:   true,
	}
	clearCredentialsOnModeChange(&in, ModePassword)
	if in.Password != "pw" || !in.AuthTestOK {
		t.Fatal("same mode should not clear credentials")
	}

	clearCredentialsOnModeChange(&in, ModeKey)
	if in.Password != "" || in.IdentityFile != "" || in.AuthTestOK {
		t.Fatalf("mode switch should clear credentials and AuthTestOK, got %+v", in)
	}
}

func TestFinalizePasswordModeInput_defaultUserRoot(t *testing.T) {
	out, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "1.2.3.4",
		Password: "pw",
	})
	if err != nil {
		t.Fatalf("finalizePasswordModeInput: %v", err)
	}
	if out.User != "root" {
		t.Errorf("User = %q, want root", out.User)
	}
}

func TestFinalizeKeyModeInput_defaultUserRoot(t *testing.T) {
	keyPath := t.TempDir() + "/id_ed25519"
	if err := writeEmptyFile(keyPath); err != nil {
		t.Fatal(err)
	}
	out, err := finalizeKeyModeInput(KeyModeInput{
		HostName:     "1.2.3.4",
		IdentityFile: keyPath,
	}, os.Stat)
	if err != nil {
		t.Fatalf("finalizeKeyModeInput: %v", err)
	}
	if out.User != "root" {
		t.Errorf("User = %q, want root", out.User)
	}
}

func writeEmptyFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}
