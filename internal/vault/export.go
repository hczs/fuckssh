package vault

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fuckssh/fuckssh/internal/platform"
)

// ExportResult 导出结果。
type ExportResult struct {
	FilePath string // 导出文件的完整路径
	FileSize int64  // 文件大小（字节）
	Hosts    int    // config 中的 Host 条目数
	Keys     int    // 导出的私钥文件数
}

// Export 将本地 SSH 配置和密钥打包加密导出。
// outDir 为导出目录，password 为主密码。
func Export(outDir string, password string) (*ExportResult, error) {
	// 校验密码
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	// 收集需要打包的文件
	files, hostCount, keyCount, err := collectFiles()
	if err != nil {
		return nil, fmt.Errorf("vault: 收集文件失败: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("vault: 没有找到需要导出的文件")
	}

	// 打包成 tar.gz
	tarData, err := createTar(files)
	if err != nil {
		return nil, fmt.Errorf("vault: 打包失败: %w", err)
	}

	// 加密
	encrypted, err := Encrypt(tarData, password)
	if err != nil {
		return nil, fmt.Errorf("vault: 加密失败: %w", err)
	}

	// 生成文件名并写出
	filename := fmt.Sprintf("fuckssh-backup-%s.tar.enc", time.Now().Format("20060102-150405"))
	outPath := filepath.Join(outDir, filename)

	if err := os.WriteFile(outPath, encrypted, 0o600); err != nil {
		return nil, fmt.Errorf("vault: 写入文件失败: %w", err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		return nil, fmt.Errorf("vault: 获取导出文件信息失败: %w", err)
	}
	return &ExportResult{
		FilePath: outPath,
		FileSize: info.Size(),
		Hosts:    hostCount,
		Keys:     keyCount,
	}, nil
}

// backupFile 表示一个需要打包的文件。
type backupFile struct {
	// ArchivePath 是在 tar 包内的相对路径（如 ssh/config、ssh/keys/id_ed25519_xxx）。
	ArchivePath string
	// Content 是文件内容。
	Content []byte
	// Mode 是文件权限。
	Mode os.FileMode
}

// collectFiles 收集 ~/.ssh/config 和 ~/.ssh/keys/ 下的私钥文件。
func collectFiles() ([]backupFile, int, int, error) {
	var files []backupFile
	var hostCount, keyCount int

	// 1. 收集 ssh config
	configPath, err := defaultConfigPath()
	if err != nil {
		return nil, 0, 0, err
	}
	if data, err := os.ReadFile(configPath); err == nil {
		files = append(files, backupFile{
			ArchivePath: "ssh/config",
			Content:     data,
			Mode:        0o600,
		})
		// 粗略统计 Host 条目数
		hostCount = countHosts(data)
	} else if !os.IsNotExist(err) {
		return nil, 0, 0, fmt.Errorf("读取 config 失败: %w", err)
	}

	// 2. 收集 ~/.ssh/keys/ 下的私钥文件
	keysDir, err := defaultKeysDir()
	if err != nil {
		return nil, 0, 0, err
	}

	entries, err := os.ReadDir(keysDir)
	if err != nil {
		if os.IsNotExist(err) {
			// keys 目录不存在，只导出 config
			return files, hostCount, 0, nil
		}
		return nil, 0, 0, fmt.Errorf("读取 keys 目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// 跳过公钥文件（.pub），只导出私钥
		if filepath.Ext(name) == ".pub" {
			continue
		}

		keyPath := filepath.Join(keysDir, name)
		data, err := os.ReadFile(keyPath)
		if err != nil {
			continue // 跳过读不出来的文件
		}

		files = append(files, backupFile{
			ArchivePath: "ssh/keys/" + name,
			Content:     data,
			Mode:        0o600,
		})
		keyCount++
	}

	return files, hostCount, keyCount, nil
}

// createTar 将文件列表打包成 gzip 压缩的 tar。
func createTar(files []backupFile) ([]byte, error) {
	var buf bytes.Buffer

	// gzip 压缩层
	gz := gzip.NewWriter(&buf)

	// tar 写入层
	tw := tar.NewWriter(gz)

	for _, f := range files {
		header := &tar.Header{
			Name: f.ArchivePath,
			Mode: int64(f.Mode),
			Size: int64(len(f.Content)),
		}
		if err := tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("写入 tar header 失败: %w", err)
		}
		if _, err := tw.Write(f.Content); err != nil {
			return nil, fmt.Errorf("写入 tar 内容失败: %w", err)
		}
	}

	// 按顺序关闭，确保数据完整刷入
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// defaultConfigPath 返回 ~/.ssh/config 的绝对路径。
func defaultConfigPath() (string, error) {
	return platform.DefaultConfigPath()
}

// defaultKeysDir 返回 ~/.ssh/keys/ 的绝对路径。
func defaultKeysDir() (string, error) {
	return platform.KeysDir()
}

// countHosts 粗略统计 config 内容中的 Host 条目数。
func countHosts(data []byte) int {
	count := 0
	for _, line := range bytes.Split(data, []byte("\n")) {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 && (bytes.HasPrefix(trimmed, []byte("Host ")) || bytes.HasPrefix(trimmed, []byte("host "))) {
			count++
		}
	}
	return count
}
