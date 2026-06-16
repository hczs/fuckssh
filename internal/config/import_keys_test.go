package config

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/vault"
)

func TestIncomingHostsToImport_skipAndRename(t *testing.T) {
	incoming := []HostEntry{
		{Alias: "skip-me"},
		{Alias: "rename-me"},
		{Alias: "fresh"},
	}
	mergeResult := &MergeResult{
		Skipped: []string{"skip-me"},
		Renames: []RenameInfo{
			{OldAlias: "rename-me", NewAlias: "rename-me-new", OldIdentityFile: "~/.ssh/keys/old"},
		},
	}

	got := IncomingHostsToImport(incoming, mergeResult)
	if len(got) != 2 {
		t.Fatalf("期望 2 个待导入 Host，got %d", len(got))
	}
	if got[0].Alias != "rename-me-new" {
		t.Errorf("重命名 Host 别名未更新: got %q", got[0].Alias)
	}
	if got[1].Alias != "fresh" {
		t.Errorf("第二个 Host 应为 fresh: got %q", got[1].Alias)
	}
}

func TestPrepareImportKeys_skipHostRemovesKey(t *testing.T) {
	sharedPriv := "id_ed25519_fuckssh_shared"
	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/config", Content: []byte("config"), Mode: 0o600},
		{ArchivePath: "ssh/keys/" + sharedPriv, Content: []byte("skipped-key"), Mode: 0o600},
	}

	incoming := []HostEntry{
		{Alias: "skipped", IdentityFile: "~/.ssh/keys/" + sharedPriv},
	}
	merged := []HostEntry{{Alias: "local", HostName: "1.1.1.1"}}
	mergeResult := &MergeResult{Skipped: []string{"skipped"}}

	_, err := PrepareImportKeys(&files, &merged, incoming, mergeResult)
	if err != nil {
		t.Fatalf("PrepareImportKeys: %v", err)
	}

	for _, f := range files {
		if f.ArchivePath == "ssh/keys/"+sharedPriv {
			t.Fatal("Skip 的 Host 其密钥不应保留在 archive 写盘列表中")
		}
	}
}

func TestPrepareImportKeys_renamesKeyForDifferentAlias(t *testing.T) {
	setTestHome(t, t.TempDir())

	sharedPriv := "id_ed25519_fuckssh_shared"
	stagingPriv, _ := keys.KeyPaths("staging")

	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/config", Content: []byte("config"), Mode: 0o600},
		{ArchivePath: "ssh/keys/" + sharedPriv, Content: []byte("staging-key"), Mode: 0o600},
	}

	incoming := []HostEntry{
		{Alias: "staging", IdentityFile: "~/.ssh/keys/" + sharedPriv},
	}
	merged := []HostEntry{
		{Alias: "prod", IdentityFile: "~/.ssh/keys/" + sharedPriv},
		{Alias: "staging", IdentityFile: "~/.ssh/keys/" + sharedPriv},
	}

	plan, err := PrepareImportKeys(&files, &merged, incoming, nil)
	if err != nil {
		t.Fatalf("PrepareImportKeys: %v", err)
	}

	expectedArchive := "ssh/keys/" + stagingPriv
	found := false
	for _, f := range files {
		if f.ArchivePath == expectedArchive {
			found = true
		}
		if f.ArchivePath == "ssh/keys/"+sharedPriv {
			t.Fatal("共享密钥文件名不应保留在写盘列表中")
		}
	}
	if !found {
		t.Fatalf("archive 密钥应重命名为 %q", expectedArchive)
	}

	stagingEntry := findHost(merged, "staging")
	if stagingEntry == nil {
		t.Fatal("merged 中应有 staging")
		return
	}
	if filepath.Base(stagingEntry.IdentityFile) != stagingPriv {
		t.Errorf("IdentityFile 应指向 %q，got %q", stagingPriv, stagingEntry.IdentityFile)
	}

	if len(plan.KeysRenamed) != 1 {
		t.Fatalf("期望 1 条密钥重命名记录，got %d", len(plan.KeysRenamed))
	}
}

func TestPrepareImportKeys_conflictRenameUpdatesHostAndKey(t *testing.T) {
	setTestHome(t, t.TempDir())

	oldPriv, _ := keys.KeyPaths("myserver")
	newPriv, _ := keys.KeyPaths("myserver-new")

	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/keys/" + oldPriv, Content: []byte("incoming-key"), Mode: 0o600},
	}

	incoming := []HostEntry{
		{Alias: "myserver", IdentityFile: "~/.ssh/keys/" + oldPriv, HostName: "10.0.0.2"},
	}
	merged := []HostEntry{
		{Alias: "myserver", IdentityFile: "~/.ssh/keys/" + oldPriv, HostName: "10.0.0.1"},
		{Alias: "myserver-new", IdentityFile: "~/.ssh/keys/" + oldPriv, HostName: "10.0.0.2"},
	}
	mergeResult := &MergeResult{
		Renamed: []string{"myserver→myserver-new"},
		Renames: []RenameInfo{
			{
				OldAlias:        "myserver",
				NewAlias:        "myserver-new",
				OldIdentityFile: "~/.ssh/keys/" + oldPriv,
			},
		},
	}

	plan, err := PrepareImportKeys(&files, &merged, incoming, mergeResult)
	if err != nil {
		t.Fatalf("PrepareImportKeys: %v", err)
	}

	if len(files) != 1 || files[0].ArchivePath != "ssh/keys/"+newPriv {
		t.Fatalf("archive 密钥应重命名为新别名对应文件，got %v", files[0].ArchivePath)
	}

	local := findHost(merged, "myserver")
	if local == nil || filepath.Base(local.IdentityFile) != oldPriv {
		t.Fatal("本地原有 myserver 及其密钥路径应保持不变")
	}

	renamed := findHost(merged, "myserver-new")
	if renamed == nil || filepath.Base(renamed.IdentityFile) != newPriv {
		t.Fatalf("重命名 Host 的 IdentityFile 应更新为 %q", newPriv)
	}

	if len(plan.KeysRenamed) != 1 {
		t.Fatalf("期望记录密钥重命名，got %v", plan.KeysRenamed)
	}
}

func TestPrepareImportKeys_customKeyRenamedOnHostRename(t *testing.T) {
	setTestHome(t, t.TempDir())

	newPriv, _ := keys.KeyPaths("myserver-new")
	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/keys/my_rsa", Content: []byte("key"), Mode: 0o600},
	}

	incoming := []HostEntry{
		{Alias: "myserver", IdentityFile: "~/.ssh/my_rsa"},
	}
	merged := []HostEntry{
		{Alias: "myserver", IdentityFile: "~/.ssh/keys/id_ed25519_fuckssh_myserver"},
		{Alias: "myserver-new", IdentityFile: "~/.ssh/my_rsa"},
	}
	mergeResult := &MergeResult{
		Renames: []RenameInfo{
			{OldAlias: "myserver", NewAlias: "myserver-new", OldIdentityFile: "~/.ssh/my_rsa"},
		},
	}

	_, err := PrepareImportKeys(&files, &merged, incoming, mergeResult)
	if err != nil {
		t.Fatalf("PrepareImportKeys: %v", err)
	}

	if files[0].ArchivePath != "ssh/keys/"+newPriv {
		t.Errorf("自定义密钥应规范化为 %q，got %q", newPriv, files[0].ArchivePath)
	}
}

func TestPrepareImportKeys_filtersOrphanKeys(t *testing.T) {
	setTestHome(t, t.TempDir())

	stagingPriv, _ := keys.KeyPaths("staging")
	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/config", Content: []byte("config"), Mode: 0o600},
		{ArchivePath: "ssh/keys/" + stagingPriv, Content: []byte("used"), Mode: 0o600},
		{ArchivePath: "ssh/keys/id_ed25519_fuckssh_orphan", Content: []byte("orphan"), Mode: 0o600},
	}

	incoming := []HostEntry{
		{Alias: "staging", IdentityFile: "~/.ssh/keys/" + stagingPriv},
	}
	merged := append([]HostEntry(nil), incoming...)

	_, err := PrepareImportKeys(&files, &merged, incoming, nil)
	if err != nil {
		t.Fatalf("PrepareImportKeys: %v", err)
	}

	for _, f := range files {
		if f.ArchivePath == "ssh/keys/id_ed25519_fuckssh_orphan" {
			t.Fatal("orphan 密钥不应进入写盘列表")
		}
	}
}

func TestPrepareImportKeys_conflictRenameCustomExternalKey(t *testing.T) {
	setTestHome(t, t.TempDir())

	targetPriv, _ := keys.KeyPaths("t1")
	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/keys/mac.pem", Content: []byte("mac-key"), Mode: 0o600},
	}

	incoming := []HostEntry{
		{Alias: "tencentrb", IdentityFile: "~/dev/ssh_keys/mac.pem", HostName: "124.156.223.9"},
	}
	merged := []HostEntry{
		{Alias: "tencentrb", IdentityFile: "~/dev/ssh_keys/mac.pem", HostName: "124.156.223.9"},
		{Alias: "t1", IdentityFile: "~/dev/ssh_keys/mac.pem", HostName: "124.156.223.9"},
	}
	mergeResult := &MergeResult{
		Renames: []RenameInfo{
			{OldAlias: "tencentrb", NewAlias: "t1", OldIdentityFile: "~/dev/ssh_keys/mac.pem"},
		},
	}

	_, err := PrepareImportKeys(&files, &merged, incoming, mergeResult)
	if err != nil {
		t.Fatalf("PrepareImportKeys: %v", err)
	}

	if len(files) != 1 || files[0].ArchivePath != "ssh/keys/"+targetPriv {
		t.Fatalf("mac.pem 应规范化为 %q，got %q", targetPriv, files[0].ArchivePath)
	}
	if string(files[0].Content) != "mac-key" {
		t.Fatal("密钥内容应保留")
	}

	renamed := findHost(merged, "t1")
	if renamed == nil || filepath.Base(renamed.IdentityFile) != targetPriv {
		t.Fatalf("t1 的 IdentityFile 应指向 %q", targetPriv)
	}
}

func TestPrepareImportKeys_conflictRenameWhenIdentityAlreadyTargetName(t *testing.T) {
	setTestHome(t, t.TempDir())

	// 复现：备份里 Host 别名 foo，但 IdentityFile 已指向 t1 的密钥名；
	// archive 实际只有 foo 别名对应的私钥文件。冲突导入重命名为 t1 后，必须把 foo 密钥移到 t1。
	oldPriv, _ := keys.KeyPaths("foo")
	targetPriv, _ := keys.KeyPaths("t1")

	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/keys/" + oldPriv, Content: []byte("incoming-key"), Mode: 0o600},
	}

	incoming := []HostEntry{
		{Alias: "foo", IdentityFile: "~/.ssh/keys/" + targetPriv, HostName: "10.0.0.2"},
	}
	merged := []HostEntry{
		{Alias: "foo", IdentityFile: "~/.ssh/keys/" + oldPriv, HostName: "10.0.0.1"},
		{Alias: "t1", IdentityFile: "~/.ssh/keys/" + targetPriv, HostName: "10.0.0.2"},
	}
	mergeResult := &MergeResult{
		Renames: []RenameInfo{
			{OldAlias: "foo", NewAlias: "t1", OldIdentityFile: "~/.ssh/keys/" + targetPriv},
		},
	}

	_, err := PrepareImportKeys(&files, &merged, incoming, mergeResult)
	if err != nil {
		t.Fatalf("PrepareImportKeys: %v", err)
	}

	if len(files) != 1 || files[0].ArchivePath != "ssh/keys/"+targetPriv {
		t.Fatalf("archive 密钥应落到 %q，got %q", targetPriv, files[0].ArchivePath)
	}
	if string(files[0].Content) != "incoming-key" {
		t.Fatal("密钥内容应保留")
	}

	renamed := findHost(merged, "t1")
	if renamed == nil || filepath.Base(renamed.IdentityFile) != targetPriv {
		t.Fatalf("重命名 Host 的 IdentityFile 应指向 %q", targetPriv)
	}
}

func TestPrepareImportKeys_missingArchiveKeyReturnsError(t *testing.T) {
	setTestHome(t, t.TempDir())

	files := []vault.ExtractedFile{
		{ArchivePath: "ssh/config", Content: []byte("config"), Mode: 0o600},
	}
	incoming := []HostEntry{{Alias: "t1", HostName: "1.2.3.4"}}
	merged := append([]HostEntry(nil), incoming...)

	_, err := PrepareImportKeys(&files, &merged, incoming, nil)
	if err == nil {
		t.Fatal("archive 缺少私钥时应返回错误")
	}
	if !errors.Is(err, ErrImportKeyMissing) {
		t.Fatalf("期望 ErrImportKeyMissing，got %v", err)
	}
}

func TestHostsReferencingKey(t *testing.T) {
	priv, _ := keys.KeyPaths("prod")
	entries := []HostEntry{
		{Alias: "prod", IdentityFile: "~/.ssh/keys/" + priv},
		{Alias: "staging", IdentityFile: "~/.ssh/keys/other"},
	}

	refs := HostsReferencingKey(entries, priv)
	if len(refs) != 1 || refs[0] != "prod" {
		t.Fatalf("HostsReferencingKey = %v", refs)
	}
}

func findHost(entries []HostEntry, alias string) *HostEntry {
	for i := range entries {
		if entries[i].Alias == alias {
			return &entries[i]
		}
	}
	return nil
}
