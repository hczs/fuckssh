package vault

import (
	"path"
	"testing"
)

// 回归：tar 内路径始终为 ssh/keys/...（正斜杠），在 Windows 上须用 path 而非 filepath 判断目录。
func TestArchiveKeyPathUsesForwardSlash(t *testing.T) {
	const archivePath = "ssh/keys/id_ed25519_fuckssh_demo"
	if path.Dir(archivePath) != "ssh/keys" {
		t.Fatalf("path.Dir = %q, want ssh/keys", path.Dir(archivePath))
	}
	if !isArchiveKeyPath(archivePath) {
		t.Fatal("isArchiveKeyPath should be true for ssh/keys/...")
	}
	if isArchiveKeyPath("ssh/config") {
		t.Fatal("config path should not be treated as key file")
	}
}
