package vault

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportRoundTrip(t *testing.T) {
	// 创建临时目录模拟导出
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "export")
	if err := os.MkdirAll(outDir, 0o700); err != nil {
		t.Fatalf("创建导出目录失败: %v", err)
	}

	// 创建模拟的 ~/.ssh 结构
	sshDir := filepath.Join(tmpDir, ".ssh")
	keysDir := filepath.Join(sshDir, "keys")
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		t.Fatalf("创建 ssh 目录失败: %v", err)
	}

	configContent := []byte("Host test\n    HostName 1.2.3.4\n    User root\n")
	if err := os.WriteFile(filepath.Join(sshDir, "config"), configContent, 0o600); err != nil {
		t.Fatalf("写入 config 失败: %v", err)
	}

	keyContent := []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nfake\n-----END OPENSSH PRIVATE KEY-----\n")
	if err := os.WriteFile(filepath.Join(keysDir, "id_ed25519_test"), keyContent, 0o600); err != nil {
		t.Fatalf("写入私钥失败: %v", err)
	}

	// 手动打包加密（因为 Export 依赖 platform 路径，这里直接构造）
	files := []backupFile{
		{ArchivePath: "ssh/config", Content: configContent, Mode: 0o600},
		{ArchivePath: "ssh/keys/id_ed25519_test", Content: keyContent, Mode: 0o600},
	}
	tarData, err := createTar(files)
	if err != nil {
		t.Fatalf("createTar 失败: %v", err)
	}

	encrypted, err := Encrypt(tarData, "testpass123")
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}

	// 写入加密文件
	encPath := filepath.Join(outDir, "backup.tar.enc")
	if err := os.WriteFile(encPath, encrypted, 0o600); err != nil {
		t.Fatalf("写入加密文件失败: %v", err)
	}

	// 测试 DecryptAndExtract
	extracted, err := DecryptAndExtract(encPath, "testpass123")
	if err != nil {
		t.Fatalf("DecryptAndExtract 失败: %v", err)
	}

	if len(extracted) != 2 {
		t.Fatalf("期望 2 个文件，got %d", len(extracted))
	}

	// 验证文件内容
	configFound := false
	keyFound := false
	for _, f := range extracted {
		switch f.ArchivePath {
		case "ssh/config":
			configFound = true
			if string(f.Content) != string(configContent) {
				t.Errorf("config 内容不匹配")
			}
		case "ssh/keys/id_ed25519_test":
			keyFound = true
			if string(f.Content) != string(keyContent) {
				t.Errorf("私钥内容不匹配")
			}
		}
	}
	if !configFound {
		t.Error("未找到 config 文件")
	}
	if !keyFound {
		t.Error("未找到私钥文件")
	}
}

func TestDecryptAndExtractWrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	encPath := filepath.Join(tmpDir, "test.tar.enc")

	// 创建一个加密文件
	encrypted, err := Encrypt([]byte("test data"), "correctpass1")
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}
	if err := os.WriteFile(encPath, encrypted, 0o600); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 用错误密码解密
	_, err = DecryptAndExtract(encPath, "wrongpass123")
	if !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("期望 ErrWrongPassword，got: %v", err)
	}
}

func TestDecryptAndExtractFileNotFound(t *testing.T) {
	_, err := DecryptAndExtract("/nonexistent/file.tar.enc", "password123")
	if err == nil {
		t.Fatal("文件不存在应该报错")
	}
}

func TestGetConfigContent(t *testing.T) {
	// 测试 GetConfigContent 从文件列表中提取 config
	configContent := []byte("Host demo\n    HostName 5.6.7.8\n")
	keyContent := []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nfakekey\n-----END OPENSSH PRIVATE KEY-----\n")

	extracted := []ExtractedFile{
		{ArchivePath: "ssh/config", Content: configContent, Mode: 0o600},
		{ArchivePath: "ssh/keys/id_ed25519_demo", Content: keyContent, Mode: 0o600},
	}

	// writeFiles 依赖 defaultSSHDir()，这里测试 GetConfigContent
	configOut := GetConfigContent(extracted)
	if configOut == nil {
		t.Fatal("GetConfigContent 返回 nil")
	}
	if string(configOut) != string(configContent) {
		t.Errorf("config 内容不匹配")
	}
}

func TestGetConfigContentNotFound(t *testing.T) {
	extracted := []ExtractedFile{
		{ArchivePath: "ssh/keys/id_ed25519_test", Content: []byte("key"), Mode: 0o600},
	}

	configOut := GetConfigContent(extracted)
	if configOut != nil {
		t.Errorf("没有 config 时应返回 nil，got %d bytes", len(configOut))
	}
}

func TestExtractTarRejectsDotDot(t *testing.T) {
	// 测试路径穿越（..）被拒绝
	files := []backupFile{
		{ArchivePath: "../../../etc/passwd", Content: []byte("evil"), Mode: 0o600},
	}

	tarData, err := createTar(files)
	if err != nil {
		t.Fatalf("createTar 失败: %v", err)
	}

	_, err = extractTar(tarData)
	if err == nil {
		t.Fatal("路径穿越应该被拒绝")
	}
	if !strings.Contains(err.Error(), "不安全") {
		t.Errorf("错误信息应包含'不安全'，got: %v", err)
	}
}

func TestExtractTarEmpty(t *testing.T) {
	// 空 tar
	var buf []byte
	// 创建一个空的 tar.gz
	files := []backupFile{}
	tarData, err := createTar(files)
	if err != nil {
		t.Fatalf("createTar 失败: %v", err)
	}

	_ = buf
	extracted, err := extractTar(tarData)
	if err != nil {
		t.Fatalf("extractTar 失败: %v", err)
	}

	if len(extracted) != 0 {
		t.Errorf("空 tar 应返回空列表，got %d", len(extracted))
	}
}
