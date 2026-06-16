package vault

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fuckssh/fuckssh/internal/platform"
)

// ErrKeyWouldOverwrite 表示导入密钥会覆盖仍被其他 Host 引用的已有私钥。
var ErrKeyWouldOverwrite = errors.New("vault: import would overwrite key file used by another host")

// KeyWriteContext 导入写盘时的密钥归属与引用信息（由 config 层构建）。
type KeyWriteContext struct {
	// KeyOwner 记录本次导入写入的私钥文件名 → 归属 Host 别名。
	KeyOwner map[string]string
	// KeyRefs 记录 merged config 中私钥文件名 → 引用它的 Host 别名列表。
	KeyRefs map[string][]string
}

// ImportResult 导入结果。
type ImportResult struct {
	ConfigImported bool   // 是否导入了 config
	KeysImported   int    // 导入的私钥文件数
	KeysSkipped    int    // 内容相同、未重复写入的私钥数
	BackupPath     string // 原有 config 的备份路径（如果有）
}

// ExtractedFile 表示从 tar 包中解出的文件。
type ExtractedFile struct {
	ArchivePath string      // tar 内的相对路径（如 ssh/config、ssh/keys/xxx）
	Content     []byte      // 文件内容
	Mode        os.FileMode // 文件权限
}

// DecryptAndExtract 解密 vault 文件并解包，返回文件列表。
// 供 cmd 层在合并场景下使用（需要先解析 config 内容再决定如何写入）。
func DecryptAndExtract(filePath string, password string) ([]ExtractedFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("vault: 读取文件失败: %w", err)
	}

	tarData, err := Decrypt(data, password)
	if err != nil {
		return nil, err
	}

	files, err := extractTar(tarData)
	if err != nil {
		return nil, fmt.Errorf("vault: 解包失败: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("vault: 备份文件为空")
	}

	return files, nil
}

// Import 解密并导入 vault 文件（简单模式，直接覆盖写入）。
func Import(filePath string, password string) (*ImportResult, error) {
	files, err := DecryptAndExtract(filePath, password)
	if err != nil {
		return nil, err
	}

	return ImportFiles(files)
}

// ImportWithConfig 解密并导入 vault 文件，config 内容由调用方提供（合并后的版本）。
// keys 仍然从 vault 文件中提取并写入。
func ImportWithConfig(filePath string, password string, mergedConfig []byte) (*ImportResult, error) {
	files, err := DecryptAndExtract(filePath, password)
	if err != nil {
		return nil, err
	}

	return ImportFilesWithConfig(files, mergedConfig, KeyWriteContext{})
}

// ImportFiles 将已解密的文件列表写入本机（简单模式，直接覆盖写入）。
func ImportFiles(files []ExtractedFile) (*ImportResult, error) {
	return writeFiles(files, nil)
}

// ImportFilesWithConfig 将已解密的文件列表写入本机，config 内容由调用方提供。
func ImportFilesWithConfig(files []ExtractedFile, mergedConfig []byte, keyCtx KeyWriteContext) (*ImportResult, error) {
	// 用合并后的 config 替换备份中的 config
	for i, f := range files {
		if f.ArchivePath == "ssh/config" {
			files[i].Content = mergedConfig
			break
		}
	}

	return writeFiles(files, &keyCtx)
}

// isArchiveKeyPath 判断 tar 内路径是否为私钥文件（archive 内统一用 / 分隔）。
func isArchiveKeyPath(archivePath string) bool {
	return archivePath != "ssh/config" && path.Dir(archivePath) == "ssh/keys"
}

// writeFiles 将解包后的文件写入本机对应位置。
func writeFiles(files []ExtractedFile, keyCtx *KeyWriteContext) (*ImportResult, error) {
	result := &ImportResult{}
	sshDir, err := defaultSSHDir()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		targetPath := resolveTargetPath(sshDir, f.ArchivePath)

		// 如果是 config 文件且已存在，先备份
		if f.ArchivePath == "ssh/config" {
			if _, statErr := os.Stat(targetPath); statErr == nil {
				bakPath := targetPath + ".bak." + time.Now().Format("20060102-150405")
				if bakErr := copyFile(targetPath, bakPath); bakErr == nil {
					result.BackupPath = bakPath
				}
			}
			result.ConfigImported = true
		}

		// tar 内路径始终用 /，须用 path 而非 filepath（Windows 上 filepath.Dir 会得到 ssh\keys）。
		isKeyFile := isArchiveKeyPath(f.ArchivePath)
		if isKeyFile {
			skip, skipErr := shouldSkipKeyWrite(targetPath, f.Content, path.Base(f.ArchivePath), keyCtx)
			if skipErr != nil {
				return nil, skipErr
			}
			if skip {
				result.KeysSkipped++
				continue
			}
		}

		// 创建目录（如果需要）
		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}

		mode := f.Mode
		if mode == 0 {
			mode = 0o600
		}
		if err := os.WriteFile(targetPath, f.Content, mode); err != nil {
			return nil, fmt.Errorf("写入文件 %s 失败: %w", targetPath, err)
		}

		// 设置权限（非 Windows 系统）
		if runtime.GOOS != "windows" {
			_ = os.Chmod(targetPath, mode)
		}

		if isKeyFile {
			result.KeysImported++
		}
	}

	return result, nil
}

// shouldSkipKeyWrite 检查是否应跳过密钥写入（内容相同）或拒绝覆盖（被其他 Host 引用）。
func shouldSkipKeyWrite(targetPath string, incoming []byte, keyBasename string, keyCtx *KeyWriteContext) (skip bool, err error) {
	existing, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return false, nil
		}
		return false, fmt.Errorf("读取已有密钥 %s 失败: %w", targetPath, readErr)
	}

	if bytes.Equal(existing, incoming) {
		return true, nil
	}

	if keyCtx == nil {
		return false, nil
	}

	owner := keyCtx.KeyOwner[keyBasename]
	for _, alias := range keyCtx.KeyRefs[keyBasename] {
		if alias != owner {
			return false, fmt.Errorf("%w: %s (referenced by %q)", ErrKeyWouldOverwrite, keyBasename, alias)
		}
	}
	return false, nil
}

// GetConfigContent 从解包文件列表中提取 config 内容。
func GetConfigContent(files []ExtractedFile) []byte {
	for _, f := range files {
		if f.ArchivePath == "ssh/config" {
			return f.Content
		}
	}
	return nil
}

// extractTar 从 gzip+tar 数据中提取所有文件。
func extractTar(data []byte) ([]ExtractedFile, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("创建 gzip reader 失败: %w", err)
	}
	defer func() { _ = gz.Close() }()

	var files []ExtractedFile
	tr := tar.NewReader(gz)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取 tar 条目失败: %w", err)
		}

		// 跳过目录条目
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// 安全检查：拒绝路径穿越
		if filepath.IsAbs(header.Name) || containsDotDot(header.Name) {
			return nil, fmt.Errorf("vault: 检测到不安全的路径 %q，拒绝解包", header.Name)
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("读取文件内容失败: %w", err)
		}

		files = append(files, ExtractedFile{
			ArchivePath: header.Name,
			Content:     content,
			Mode:        os.FileMode(header.Mode),
		})
	}

	return files, nil
}

// resolveTargetPath 将 tar 内的相对路径解析为本机绝对路径。
func resolveTargetPath(sshDir string, archivePath string) string {
	rel := archivePath
	if len(rel) > 4 && rel[:4] == "ssh/" {
		rel = rel[4:]
	}
	return filepath.Join(sshDir, rel)
}

// defaultSSHDir 返回 ~/.ssh 的绝对路径。
func defaultSSHDir() (string, error) {
	return platform.SSHDir()
}

// containsDotDot 检查路径中是否包含 ..（防止路径穿越攻击）。
func containsDotDot(path string) bool {
	for _, part := range filepath.SplitList(path) {
		if part == ".." {
			return true
		}
	}
	for _, part := range bytes.Split([]byte(path), []byte("/")) {
		if string(part) == ".." {
			return true
		}
	}
	return false
}

// copyFile 复制文件（用于备份）。
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}
