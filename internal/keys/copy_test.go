package keys

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyKeyPair_copiesPrivateAndPublic(t *testing.T) {
	dir := t.TempDir()
	srcPriv := filepath.Join(dir, "src_key")
	srcPub := srcPriv + ".pub"
	if err := os.WriteFile(srcPriv, []byte("private"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcPub, []byte("public\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(dir, "keys")
	got, err := CopyKeyPair(srcPriv, destDir, "id_ed25519_fuckssh_test")
	if err != nil {
		t.Fatalf("CopyKeyPair: %v", err)
	}

	wantPriv := filepath.Join(destDir, "id_ed25519_fuckssh_test")
	if got != wantPriv {
		t.Errorf("dest = %q, want %q", got, wantPriv)
	}
	for _, p := range []string{wantPriv, wantPriv + ".pub"} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %q: %v", p, err)
		}
	}
	// 源文件仍在
	if _, err := os.Stat(srcPriv); err != nil {
		t.Errorf("source removed: %v", err)
	}
}

func TestCopyKeyPair_noPubWhenMissing(t *testing.T) {
	dir := t.TempDir()
	srcPriv := filepath.Join(dir, "solo")
	if err := os.WriteFile(srcPriv, []byte("private"), 0o600); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(dir, "keys")
	if _, err := CopyKeyPair(srcPriv, destDir, "solo_dest"); err != nil {
		t.Fatalf("CopyKeyPair: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "solo_dest.pub")); !os.IsNotExist(err) {
		t.Errorf("unexpected .pub: %v", err)
	}
}

func TestCopyKeyPair_refusesExisting(t *testing.T) {
	dir := t.TempDir()
	srcPriv := filepath.Join(dir, "src")
	destDir := filepath.Join(dir, "keys")
	existing := filepath.Join(destDir, "id_ed25519_fuckssh_dup")
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcPriv, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := CopyKeyPair(srcPriv, destDir, "id_ed25519_fuckssh_dup")
	if !errors.Is(err, ErrKeyFileExists) {
		t.Fatalf("err = %v, want ErrKeyFileExists", err)
	}
}

func TestCopyKeyPair_unixPrivateMode0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix only")
	}
	dir := t.TempDir()
	srcPriv := filepath.Join(dir, "src")
	if err := os.WriteFile(srcPriv, []byte("private"), 0o600); err != nil {
		t.Fatal(err)
	}
	destDir := filepath.Join(dir, "keys")
	dest, err := CopyKeyPair(srcPriv, destDir, "k")
	if err != nil {
		t.Fatalf("CopyKeyPair: %v", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %o, want 0600", info.Mode().Perm())
	}
}
