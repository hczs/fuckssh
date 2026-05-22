package keys

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSignerFromFile_ed25519(t *testing.T) {
	kp, err := GenerateEd25519()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "id_ed25519")
	if err := os.WriteFile(path, kp.PrivatePEM, 0o600); err != nil {
		t.Fatal(err)
	}

	signer, err := LoadSignerFromFile(path)
	if err != nil {
		t.Fatalf("LoadSignerFromFile: %v", err)
	}
	if signer == nil {
		t.Fatal("signer is nil")
	}
}

func TestLoadSignerFromFile_invalidKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad")
	if err := os.WriteFile(path, []byte("not-a-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSignerFromFile(path); err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestErrPassphraseNotSupported_isSentinel(t *testing.T) {
	var err error = ErrPassphraseNotSupported
	if !errors.Is(err, ErrPassphraseNotSupported) {
		t.Fatal("sentinel mismatch")
	}
}
