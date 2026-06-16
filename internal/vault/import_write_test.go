package vault

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func setupTestSSHDir(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	sshDir := filepath.Join(home, ".ssh")
	keysDir := filepath.Join(sshDir, "keys")
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		t.Fatalf("mkdir keys: %v", err)
	}
	return keysDir
}

func TestWriteFiles_skipsIdenticalKey(t *testing.T) {
	keysDir := setupTestSSHDir(t)
	keyName := "id_ed25519_fuckssh_demo"
	keyPath := filepath.Join(keysDir, keyName)
	content := []byte("same-key-content")
	if err := os.WriteFile(keyPath, content, 0o600); err != nil {
		t.Fatalf("write existing key: %v", err)
	}

	files := []ExtractedFile{
		{ArchivePath: "ssh/config", Content: []byte("Host demo\n"), Mode: 0o600},
		{ArchivePath: "ssh/keys/" + keyName, Content: content, Mode: 0o600},
	}

	ctx := KeyWriteContext{
		KeyOwner: map[string]string{keyName: "demo"},
		KeyRefs:  map[string][]string{keyName: {"demo"}},
	}

	result, err := writeFiles(files, &ctx)
	if err != nil {
		t.Fatalf("writeFiles: %v", err)
	}
	if result.KeysImported != 0 {
		t.Errorf("相同内容不应重复写入，KeysImported=%d", result.KeysImported)
	}
	if result.KeysSkipped != 1 {
		t.Errorf("KeysSkipped=%d, want 1", result.KeysSkipped)
	}
}

func TestWriteFiles_refusesOverwriteUsedByOtherHost(t *testing.T) {
	keysDir := setupTestSSHDir(t)
	keyName := "id_ed25519_fuckssh_shared"
	keyPath := filepath.Join(keysDir, keyName)
	if err := os.WriteFile(keyPath, []byte("local-key"), 0o600); err != nil {
		t.Fatalf("write existing key: %v", err)
	}

	files := []ExtractedFile{
		{ArchivePath: "ssh/keys/" + keyName, Content: []byte("incoming-key"), Mode: 0o600},
	}

	ctx := KeyWriteContext{
		KeyOwner: map[string]string{keyName: "staging"},
		KeyRefs:  map[string][]string{keyName: {"prod", "staging"}},
	}

	_, err := writeFiles(files, &ctx)
	if !errors.Is(err, ErrKeyWouldOverwrite) {
		t.Fatalf("期望 ErrKeyWouldOverwrite，got %v", err)
	}

	after, readErr := os.ReadFile(keyPath)
	if readErr != nil {
		t.Fatalf("read key after failed import: %v", readErr)
	}
	if string(after) != "local-key" {
		t.Fatal("本地密钥内容不应被修改")
	}
}

func TestWriteFiles_allowsOverwriteForSameOwner(t *testing.T) {
	keysDir := setupTestSSHDir(t)
	keyName := "id_ed25519_fuckssh_myserver"
	keyPath := filepath.Join(keysDir, keyName)
	if err := os.WriteFile(keyPath, []byte("old-key"), 0o600); err != nil {
		t.Fatalf("write existing key: %v", err)
	}

	incoming := []byte("new-key")
	files := []ExtractedFile{
		{ArchivePath: "ssh/keys/" + keyName, Content: incoming, Mode: 0o600},
	}

	ctx := KeyWriteContext{
		KeyOwner: map[string]string{keyName: "myserver"},
		KeyRefs:  map[string][]string{keyName: {"myserver"}},
	}

	result, err := writeFiles(files, &ctx)
	if err != nil {
		t.Fatalf("writeFiles: %v", err)
	}
	if result.KeysImported != 1 {
		t.Errorf("KeysImported=%d, want 1", result.KeysImported)
	}

	after, readErr := os.ReadFile(keyPath)
	if readErr != nil {
		t.Fatalf("read key: %v", readErr)
	}
	if string(after) != string(incoming) {
		t.Fatal("同别名覆盖场景应写入新密钥内容")
	}
}
