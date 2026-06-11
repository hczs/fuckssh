package cmd

import (
	"path/filepath"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/vault"
)

func TestRenameArchiveKeys(t *testing.T) {
	oldPriv, _ := keys.KeyPaths("myserver")
	newPriv, _ := keys.KeyPaths("myserver-new")

	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/config", Content: []byte("config"), Mode: 0o600},
		{ArchivePath: "ssh/keys/" + oldPriv, Content: []byte("key-data"), Mode: 0o600},
	}

	renames := []config.RenameInfo{
		{OldAlias: "myserver", NewAlias: "myserver-new", OldIdentityFile: "~/.ssh/keys/" + oldPriv},
	}

	renameArchiveKeys(files, renames)

	expectedPath := "ssh/keys/" + newPriv
	if files[1].ArchivePath != expectedPath {
		t.Errorf("密钥 archive 路径未更新: 期望 %q，got %q", expectedPath, files[1].ArchivePath)
	}
	// config 不应被修改
	if files[0].ArchivePath != "ssh/config" {
		t.Errorf("config 路径不应被修改: got %q", files[0].ArchivePath)
	}
}

func TestRenameArchiveKeysCustomKey(t *testing.T) {
	// 测试自定义密钥名（非 fuckssh 命名规则）
	newPriv, _ := keys.KeyPaths("myserver-new")

	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/keys/my_rsa", Content: []byte("key-data"), Mode: 0o600},
	}

	renames := []config.RenameInfo{
		{OldAlias: "myserver", NewAlias: "myserver-new", OldIdentityFile: "~/.ssh/my_rsa"},
	}

	renameArchiveKeys(files, renames)

	expectedPath := "ssh/keys/" + newPriv
	if files[0].ArchivePath != expectedPath {
		t.Errorf("自定义密钥 archive 路径未更新: 期望 %q，got %q", expectedPath, files[0].ArchivePath)
	}
}

func TestRenameArchiveKeysNoMatch(t *testing.T) {
	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/keys/id_ed25519_fuckssh_other", Content: []byte("key"), Mode: 0o600},
	}

	renames := []config.RenameInfo{
		{OldAlias: "myserver", NewAlias: "myserver-new", OldIdentityFile: "~/.ssh/keys/id_ed25519_fuckssh_myserver"},
	}

	originalPath := files[0].ArchivePath
	renameArchiveKeys(files, renames)

	if files[0].ArchivePath != originalPath {
		t.Errorf("不匹配的密钥不应被修改: got %q", files[0].ArchivePath)
	}
}

func TestUpdateIdentityFiles(t *testing.T) {
	oldPriv, _ := keys.KeyPaths("myserver")
	newPriv, _ := keys.KeyPaths("myserver-new")

	merged := []config.HostEntry{
		// 原有 Host（未重命名，不应受影响）
		{Alias: "myserver", IdentityFile: "~/.ssh/keys/" + oldPriv},
		// 重命名后的 Host
		{Alias: "myserver-new", IdentityFile: "~/.ssh/keys/" + oldPriv},
	}

	renames := []config.RenameInfo{
		{OldAlias: "myserver", NewAlias: "myserver-new", OldIdentityFile: "~/.ssh/keys/" + oldPriv},
	}

	updateIdentityFiles(merged, renames)

	// 原有 Host 的 IdentityFile 不应改变
	expectedOld := "~/.ssh/keys/" + oldPriv
	if merged[0].IdentityFile != expectedOld {
		t.Errorf("原有 Host IdentityFile 不应改变: 期望 %q，got %q", expectedOld, merged[0].IdentityFile)
	}

	// 重命名后的 Host 的 IdentityFile 应更新
	expectedNew := filepath.Join("~/.ssh/keys", newPriv)
	if merged[1].IdentityFile != expectedNew {
		t.Errorf("重命名 Host IdentityFile 应更新: 期望 %q，got %q", expectedNew, merged[1].IdentityFile)
	}
}

func TestUpdateIdentityFilesCustomKey(t *testing.T) {
	// 测试自定义密钥名（非 fuckssh 命名规则）
	newPriv, _ := keys.KeyPaths("myserver-new")

	merged := []config.HostEntry{
		// 原有 Host（未重命名，不应受影响）
		{Alias: "myserver", IdentityFile: "~/.ssh/my_rsa"},
		// 重命名后的 Host
		{Alias: "myserver-new", IdentityFile: "~/.ssh/my_rsa"},
	}

	renames := []config.RenameInfo{
		{OldAlias: "myserver", NewAlias: "myserver-new", OldIdentityFile: "~/.ssh/my_rsa"},
	}

	updateIdentityFiles(merged, renames)

	// 原有 Host 的 IdentityFile 不应改变
	if merged[0].IdentityFile != "~/.ssh/my_rsa" {
		t.Errorf("原有 Host IdentityFile 不应改变: got %q", merged[0].IdentityFile)
	}

	// 重命名后的 Host 的 IdentityFile 应更新为新密钥名
	expectedNew := filepath.Join("~/.ssh", newPriv)
	if merged[1].IdentityFile != expectedNew {
		t.Errorf("重命名 Host IdentityFile 应更新: 期望 %q，got %q", expectedNew, merged[1].IdentityFile)
	}
}

func TestUpdateIdentityFilesMultipleRenames(t *testing.T) {
	privA, _ := keys.KeyPaths("server-a")
	privB, _ := keys.KeyPaths("server-b")
	privA2, _ := keys.KeyPaths("server-a2")
	privB2, _ := keys.KeyPaths("server-b2")

	merged := []config.HostEntry{
		{Alias: "server-a2", IdentityFile: "~/.ssh/keys/" + privA},
		{Alias: "server-b2", IdentityFile: "~/.ssh/keys/" + privB},
	}

	renames := []config.RenameInfo{
		{OldAlias: "server-a", NewAlias: "server-a2", OldIdentityFile: "~/.ssh/keys/" + privA},
		{OldAlias: "server-b", NewAlias: "server-b2", OldIdentityFile: "~/.ssh/keys/" + privB},
	}

	updateIdentityFiles(merged, renames)

	expectedA := filepath.Join("~/.ssh/keys", privA2)
	expectedB := filepath.Join("~/.ssh/keys", privB2)

	if merged[0].IdentityFile != expectedA {
		t.Errorf("server-a2 IdentityFile: 期望 %q，got %q", expectedA, merged[0].IdentityFile)
	}
	if merged[1].IdentityFile != expectedB {
		t.Errorf("server-b2 IdentityFile: 期望 %q，got %q", expectedB, merged[1].IdentityFile)
	}
}

func TestUpdateIdentityFilesSkipsNonMatching(t *testing.T) {
	// IdentityFile 指向的不是旧别名对应的密钥，不应被修改
	merged := []config.HostEntry{
		{Alias: "myserver-new", IdentityFile: "~/.ssh/custom_key"},
	}

	renames := []config.RenameInfo{
		{OldAlias: "myserver", NewAlias: "myserver-new", OldIdentityFile: "~/.ssh/other_key"},
	}

	updateIdentityFiles(merged, renames)

	if merged[0].IdentityFile != "~/.ssh/custom_key" {
		t.Errorf("非匹配的 IdentityFile 不应被修改: got %q", merged[0].IdentityFile)
	}
}
