package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIdentityFilesFromConfig(t *testing.T) {
	data := []byte(`Host a
    IdentityFile ~/.ssh/keys/id_ed25519_fuckssh_a

Host b
    IdentityFile ~/dev/ssh_keys/mac.pem
`)
	got := identityFilesFromConfig(data)
	if len(got) != 2 {
		t.Fatalf("identityFilesFromConfig = %v", got)
	}
	if got[1] != "~/dev/ssh_keys/mac.pem" {
		t.Errorf("got %q", got[1])
	}
}

func TestCollectFilesIncludesCustomIdentityKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sshDir := filepath.Join(home, ".ssh")
	keysDir := filepath.Join(sshDir, "keys")
	customDir := filepath.Join(home, "dev", "ssh_keys")
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(customDir, 0o700); err != nil {
		t.Fatal(err)
	}

	configContent := `Host tencentrb
    HostName 124.156.223.9
    User ubuntu
    IdentityFile ~/dev/ssh_keys/mac.pem
`
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte(configContent), 0o600); err != nil {
		t.Fatal(err)
	}

	macKey := []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nmac\n-----END OPENSSH PRIVATE KEY-----\n")
	if err := os.WriteFile(filepath.Join(customDir, "mac.pem"), macKey, 0o600); err != nil {
		t.Fatal(err)
	}

	files, _, keyCount, err := collectFiles()
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}
	if keyCount != 1 {
		t.Fatalf("keyCount = %d, want 1", keyCount)
	}

	var found bool
	for _, f := range files {
		if f.ArchivePath == "ssh/keys/mac.pem" {
			found = true
			if string(f.Content) != string(macKey) {
				t.Fatal("mac.pem 内容不匹配")
			}
		}
	}
	if !found {
		t.Fatal("archive 应包含 ssh/keys/mac.pem")
	}
}

func TestExportAndImportRoundTrip(t *testing.T) {
	// 创建临时目录模拟 ~/.ssh
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	keysDir := filepath.Join(sshDir, "keys")

	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}

	// 写入模拟的 config
	configContent := `Host myserver
    HostName 192.168.1.100
    User root
    Port 22
    IdentityFile ~/.ssh/keys/id_ed25519_fuckssh_my

Host prod-db
    HostName 10.0.0.5
    User admin
    Port 2222
`
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte(configContent), 0o600); err != nil {
		t.Fatalf("写入 config 失败: %v", err)
	}

	// 写入模拟的私钥
	keyContent := []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nfake key content\n-----END OPENSSH PRIVATE KEY-----\n")
	if err := os.WriteFile(filepath.Join(keysDir, "id_ed25519_fuckssh_my"), keyContent, 0o600); err != nil {
		t.Fatalf("写入私钥失败: %v", err)
	}

	// 写入公钥（不应被导出）
	pubContent := []byte("ssh-ed25519 AAAAC3Nza fakekey comment\n")
	if err := os.WriteFile(filepath.Join(keysDir, "id_ed25519_fuckssh_my.pub"), pubContent, 0o644); err != nil {
		t.Fatalf("写入公钥失败: %v", err)
	}

	// 测试 collectFiles（需要 monkey-patch 路径，这里直接测试 tar 打包解包）
	files := []backupFile{
		{ArchivePath: "ssh/config", Content: []byte(configContent), Mode: 0o600},
		{ArchivePath: "ssh/keys/id_ed25519_fuckssh_my", Content: keyContent, Mode: 0o600},
	}

	// 测试 tar 打包
	tarData, err := createTar(files)
	if err != nil {
		t.Fatalf("createTar 失败: %v", err)
	}

	// 测试 tar 解包
	extracted, err := extractTar(tarData)
	if err != nil {
		t.Fatalf("extractTar 失败: %v", err)
	}

	if len(extracted) != 2 {
		t.Fatalf("解包文件数不匹配: got %d, want 2", len(extracted))
	}

	// 验证 config 内容
	if extracted[0].ArchivePath != "ssh/config" {
		t.Errorf("第一个文件路径不匹配: got %q", extracted[0].ArchivePath)
	}
	if string(extracted[0].Content) != configContent {
		t.Errorf("config 内容不匹配")
	}

	// 验证 key 内容
	if extracted[1].ArchivePath != "ssh/keys/id_ed25519_fuckssh_my" {
		t.Errorf("第二个文件路径不匹配: got %q", extracted[1].ArchivePath)
	}
	if string(extracted[1].Content) != string(keyContent) {
		t.Errorf("私钥内容不匹配")
	}
}

func TestCreateTarAndExtract(t *testing.T) {
	files := []backupFile{
		{ArchivePath: "ssh/config", Content: []byte("Host test\n"), Mode: 0o600},
	}

	tarData, err := createTar(files)
	if err != nil {
		t.Fatalf("createTar 失败: %v", err)
	}

	if len(tarData) == 0 {
		t.Fatal("tar 数据为空")
	}

	extracted, err := extractTar(tarData)
	if err != nil {
		t.Fatalf("extractTar 失败: %v", err)
	}

	if len(extracted) != 1 {
		t.Fatalf("期望 1 个文件，got %d", len(extracted))
	}

	if string(extracted[0].Content) != "Host test\n" {
		t.Errorf("内容不匹配")
	}
}

func TestExtractTarPathTraversal(t *testing.T) {
	// 构造一个包含路径穿越的 tar（模拟攻击）
	// 这个测试验证 extractTar 能拒绝恶意路径
	files := []backupFile{
		{ArchivePath: "../../../etc/passwd", Content: []byte("evil"), Mode: 0o600},
	}

	tarData, err := createTar(files)
	if err != nil {
		t.Fatalf("createTar 失败: %v", err)
	}

	_, err = extractTar(tarData)
	if err == nil {
		t.Fatal("应该拒绝路径穿越，但没有报错")
	}
	if !strings.Contains(err.Error(), "不安全") {
		t.Errorf("错误信息应包含'不安全'，got: %v", err)
	}
}

func TestCountHosts(t *testing.T) {
	data := []byte(`Host myserver
    HostName 192.168.1.100

Host prod-db
    HostName 10.0.0.5

host staging
    HostName 172.16.0.1

# Not a host
SomeGlobalOption value
`)
	count := countHosts(data)
	if count != 3 {
		t.Errorf("countHosts = %d, want 3", count)
	}
}

func TestResolveTargetPath(t *testing.T) {
	sshDir := "/home/user/.ssh"

	tests := []struct {
		archive string
		want    string
	}{
		{"ssh/config", filepath.Join(sshDir, "config")},
		{"ssh/keys/id_ed25519_fuckssh_my", filepath.Join(sshDir, "keys", "id_ed25519_fuckssh_my")},
	}

	for _, tt := range tests {
		t.Run(tt.archive, func(t *testing.T) {
			got := resolveTargetPath(sshDir, tt.archive)
			if got != tt.want {
				t.Errorf("resolveTargetPath(%q) = %q, want %q", tt.archive, got, tt.want)
			}
		})
	}
}
