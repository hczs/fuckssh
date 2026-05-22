package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSetPrivateKeyPerm_unix0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix only")
	}

	path := filepath.Join(t.TempDir(), "test_key")
	if err := os.WriteFile(path, []byte("secret"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := SetPrivateKeyPerm(path); err != nil {
		t.Fatalf("SetPrivateKeyPerm: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %o, want 0600", info.Mode().Perm())
	}
}
