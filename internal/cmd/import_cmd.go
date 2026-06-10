package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/vault"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <备份文件>",
	Short: "从加密备份文件导入 SSH 配置和密钥",
	Long: `解密并导入 fuckssh 导出的加密备份文件。

如果目标机器上已存在 SSH config，会自动检测冲突并询问处理方式：
  - 覆盖：用备份中的配置替换现有的
  - 跳过：保留现有的配置不动
  - 重命名：给备份中的 Host 起一个新别名再导入

导入前会自动备份现有 config。`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("请指定备份文件路径，例如：\n  fuckssh import fuckssh-backup-20260609-143022.tar.enc")
		}
		return runImportCmd(cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func runImportCmd(stdout, stderr io.Writer, filePath string) error {
	// 解密并解包（密码错误时允许重试）
	const maxRetries = 3
	var files []vault.ExtractedFile

	for attempt := 1; attempt <= maxRetries; attempt++ {
		password, err := readPasswordMasked(stderr, "请输入主密码: ")
		if err != nil {
			return err
		}
		if password == "" {
			return fmt.Errorf("密码不能为空")
		}

		_, _ = fmt.Fprintln(stderr, "正在解密...")
		files, err = vault.DecryptAndExtract(filePath, password)
		if err == nil {
			break // 解密成功
		}
		if attempt < maxRetries && errors.Is(err, vault.ErrWrongPassword) {
			_, _ = fmt.Fprintf(stderr, "密码错误，还剩 %d 次机会\n", maxRetries-attempt)
			continue
		}
		return err
	}

	// 提取备份中的 config 内容
	backupConfig := vault.GetConfigContent(files)
	if backupConfig == nil {
		return fmt.Errorf("备份中没有找到 SSH config")
	}

	// 解析备份中的 Host 列表
	incoming, err := config.Parse(strings.NewReader(string(backupConfig)), "backup")
	if err != nil {
		return fmt.Errorf("解析备份 config 失败: %w", err)
	}

	// 读取目标机器现有的 config
	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	var existing []config.HostEntry
	if _, statErr := os.Stat(configPath); statErr == nil {
		existing, err = config.ParseFile(configPath)
		if err != nil {
			return fmt.Errorf("解析现有 config 失败: %w", err)
		}
	}

	// 检测冲突
	conflicts := config.FindConflicts(existing, incoming)

	if len(conflicts) == 0 {
		// 无冲突，直接导入
		_, _ = fmt.Fprintf(stdout, "✓ 发现 %d 个 Host 配置，无冲突\n", len(incoming))
		result, err := vault.ImportFiles(files)
		if err != nil {
			return err
		}
		printImportResult(stdout, result, len(incoming))
		return nil
	}

	// 有冲突，逐个询问
	_, _ = fmt.Fprintf(stdout, "✓ 发现 %d 个 Host 配置，其中 %d 个与现有配置冲突：\n\n", len(incoming), len(conflicts))

	reader := bufio.NewReader(os.Stdin)
	conflictMap := make(map[string]config.ConflictInfo)

	for i, ci := range conflicts {
		_, _ = fmt.Fprint(stdout, config.FormatConflictSummary(ci))
		_, _ = fmt.Fprintf(stdout, "      [1] 覆盖  [2] 跳过  [3] 重命名\n")

		action := askConflictAction(reader, stdout, stderr, i+1, len(conflicts))
		ci.Action = action

		if action == config.ConflictRename {
			newAlias := askNewAlias(reader, stdout, ci.Alias)
			ci.NewAlias = newAlias
		}

		conflictMap[ci.Alias] = ci
		_, _ = fmt.Fprintln(stdout)
	}

	// 执行合并
	merged, mergeResult := config.MergeHosts(existing, incoming, conflictMap)

	// 重命名的 Host 需要同步更新 IdentityFile 和 archive 中的密钥文件
	if len(mergeResult.Renames) > 0 {
		renameArchiveKeys(files, mergeResult.Renames)
		updateIdentityFiles(merged, mergeResult.Renames)
	}

	// 序列化合并后的 config
	mergedContent := serializeHostEntries(merged)

	// 写入合并后的 config + 导入密钥（复用已解密的 files，不再重复解密）
	result, err := vault.ImportFilesWithConfig(files, []byte(mergedContent))
	if err != nil {
		return err
	}

	// 输出合并结果
	printMergeResult(stdout, result, mergeResult)
	return nil
}

// askConflictAction 交互式询问冲突处理方式。
func askConflictAction(reader *bufio.Reader, stdout, stderr io.Writer, current, total int) config.ConflictAction {
	for {
		_, _ = fmt.Fprintf(stderr, "  请选择 [%d/%d]: ", current, total)
		input, err := reader.ReadString('\n')
		if err != nil {
			_, _ = fmt.Fprintln(stderr, "  读取输入失败，默认跳过")
			return config.ConflictSkip
		}
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			return config.ConflictOverwrite
		case "2":
			return config.ConflictSkip
		case "3":
			return config.ConflictRename
		default:
			_, _ = fmt.Fprintln(stderr, "  请输入 1（覆盖）、2（跳过）或 3（重命名）")
		}
	}
}

// askNewAlias 交互式询问新别名。
func askNewAlias(reader *bufio.Reader, stdout io.Writer, oldAlias string) string {
	for {
		_, _ = fmt.Fprintf(stdout, "      请输入新别名（原名: %s）: ", oldAlias)
		input, err := reader.ReadString('\n')
		if err != nil {
			return ""
		}
		input = strings.TrimSpace(input)
		if input == "" {
			_, _ = fmt.Fprintln(stdout, "      别名不能为空，请重新输入")
			continue
		}
		if strings.EqualFold(input, oldAlias) {
			_, _ = fmt.Fprintln(stdout, "      新别名不能与原名相同")
			continue
		}
		return input
	}
}

// printImportResult 输出简单导入的结果。
func printImportResult(stdout io.Writer, result *vault.ImportResult, hostCount int) {
	_, _ = fmt.Fprintf(stdout, "✓ 导入成功\n")
	if result.ConfigImported {
		_, _ = fmt.Fprintf(stdout, "  SSH config: %d 个 Host 已导入\n", hostCount)
	}
	if result.KeysImported > 0 {
		_, _ = fmt.Fprintf(stdout, "  私钥文件: %d 个\n", result.KeysImported)
	}
	if result.BackupPath != "" {
		_, _ = fmt.Fprintf(stdout, "  原 config 已备份: %s\n", result.BackupPath)
	}
}

// printMergeResult 输出合并导入的结果。
func printMergeResult(stdout io.Writer, result *vault.ImportResult, merge *config.MergeResult) {
	_, _ = fmt.Fprintf(stdout, "✓ 合并导入完成\n")

	if len(merge.Imported) > 0 {
		_, _ = fmt.Fprintf(stdout, "  新增导入: %s\n", strings.Join(merge.Imported, ", "))
	}
	if len(merge.Overwrite) > 0 {
		_, _ = fmt.Fprintf(stdout, "  已覆盖: %s\n", strings.Join(merge.Overwrite, ", "))
	}
	if len(merge.Renamed) > 0 {
		_, _ = fmt.Fprintf(stdout, "  已重命名: %s\n", strings.Join(merge.Renamed, ", "))
	}
	if len(merge.Skipped) > 0 {
		_, _ = fmt.Fprintf(stdout, "  已跳过: %s\n", strings.Join(merge.Skipped, ", "))
	}

	if result.KeysImported > 0 {
		_, _ = fmt.Fprintf(stdout, "  私钥文件: %d 个\n", result.KeysImported)
	}
	if result.BackupPath != "" {
		_, _ = fmt.Fprintf(stdout, "  原 config 已备份: %s\n", result.BackupPath)
	}
}

// serializeHostEntries 将 HostEntry 列表序列化为 SSH config 格式。
func serializeHostEntries(entries []config.HostEntry) string {
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(config.FormatHostBlock(e.Alias, e.HostName, e.User, e.Port, e.IdentityFile, e.Remark))
	}
	return b.String()
}

// renameArchiveKeys 将 archive 中与旧别名关联的密钥文件重命名为新别名对应的文件名。
func renameArchiveKeys(files []vault.ExtractedFile, renames []config.RenameInfo) {
	for _, r := range renames {
		oldPriv, _ := keys.KeyPaths(r.OldAlias)
		newPriv, _ := keys.KeyPaths(r.NewAlias)

		oldArchivePath := "ssh/keys/" + oldPriv
		newArchivePath := "ssh/keys/" + newPriv

		for i, f := range files {
			if f.ArchivePath == oldArchivePath {
				files[i].ArchivePath = newArchivePath
				break
			}
		}
	}
}

// updateIdentityFiles 更新合并后 Host 条目的 IdentityFile，使其指向新别名对应的密钥路径。
func updateIdentityFiles(merged []config.HostEntry, renames []config.RenameInfo) {
	// 合并后的 entry.Alias 已经是新名字，所以用 newAlias 作为 key
	renameMap := make(map[string]string) // newAlias → oldAlias
	for _, r := range renames {
		renameMap[r.NewAlias] = r.OldAlias
	}

	for i, entry := range merged {
		oldAlias, ok := renameMap[entry.Alias]
		if !ok {
			continue
		}
		oldPriv, _ := keys.KeyPaths(oldAlias)
		newPriv, _ := keys.KeyPaths(entry.Alias)

		// 如果当前 IdentityFile 指向旧密钥，更新为新密钥路径
		if filepath.Base(entry.IdentityFile) == oldPriv {
			dir := filepath.Dir(entry.IdentityFile)
			merged[i].IdentityFile = filepath.Join(dir, newPriv)
		}
	}
}
