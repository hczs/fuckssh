package keys

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteKeyPair_createsTwoFiles(t *testing.T) {
	dir := t.TempDir()
	kp, err := GenerateEd25519()
	if err != nil {
		t.Fatalf("GenerateEd25519: %v", err)
	}

	if err := WriteKeyPair(dir, "id_ed25519_fuckssh_test", kp); err != nil {
		t.Fatalf("WriteKeyPair: %v", err)
	}

	privPath := filepath.Join(dir, "id_ed25519_fuckssh_test")
	pubPath := privPath + ".pub"

	privData, err := os.ReadFile(privPath)
	if err != nil {
		t.Fatalf("read private: %v", err)
	}
	if string(privData) != string(kp.PrivatePEM) {
		t.Fatal("private file content mismatch")
	}

	pubData, err := os.ReadFile(pubPath)
	if err != nil {
		t.Fatalf("read public: %v", err)
	}
	if string(pubData) != kp.PublicLine {
		t.Fatal("public file content mismatch")
	}
}

func TestWriteKeyPair_unixMode0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix private key mode only")
	}

	dir := t.TempDir()
	kp, err := GenerateEd25519()
	if err != nil {
		t.Fatalf("GenerateEd25519: %v", err)
	}

	if err := WriteKeyPair(dir, "id_ed25519_fuckssh_mode", kp); err != nil {
		t.Fatalf("WriteKeyPair: %v", err)
	}

	privPath := filepath.Join(dir, "id_ed25519_fuckssh_mode")
	info, err := os.Stat(privPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	got := info.Mode().Perm()
	if got != 0o600 {
		t.Errorf("private key mode = %o, want 0600", got)
	}
}

func TestWriteKeyPair_refusesExistingFile(t *testing.T) {
	dir := t.TempDir()
	kp, err := GenerateEd25519()
	if err != nil {
		t.Fatalf("GenerateEd25519: %v", err)
	}

	base := "id_ed25519_fuckssh_exists"
	privPath := filepath.Join(dir, base)
	if err := os.WriteFile(privPath, []byte("existing"), 0o600); err != nil {
		t.Fatalf("setup existing file: %v", err)
	}

	err = WriteKeyPair(dir, base, kp)
	if err == nil {
		t.Fatal("WriteKeyPair: want error when file exists")
	}
	if !errors.Is(err, ErrKeyFileExists) {
		t.Fatalf("WriteKeyPair err = %v, want ErrKeyFileExists", err)
	}
}
