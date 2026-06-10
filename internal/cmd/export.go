package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fuckssh/fuckssh/internal/vault"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [目录]",
	Short: "导出 SSH 配置和密钥到加密备份文件",
	Long: `将 ~/.ssh/config 和 ~/.ssh/keys/ 下的私钥打包加密导出。

不指定目录时导出到当前目录。文件名自动生成，格式如：
  fuckssh-backup-20260609-143022.tar.enc

导出时需要设置主密码（至少 6 位，不能纯数字），
导入时需要输入相同的主密码才能解密。`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outDir := "."
		if len(args) > 0 {
			outDir = args[0]
		}
		return runExport(cmd.OutOrStdout(), cmd.ErrOrStderr(), outDir)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func runExport(stdout, stderr io.Writer, outDir string) error {
	// 耗时由本函数自行管理，跳过 root.go 的默认输出
	skipElapsedOutput = true

	// 检查目录是否存在
	info, err := os.Stat(outDir)
	if err != nil {
		return fmt.Errorf("目录 %s 不存在: %w", outDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是一个目录", outDir)
	}

	// 读取主密码（两次确认），验证失败时重新输入
	_, _ = fmt.Fprintln(stderr, "请设置主密码（至少 6 位，不能纯数字）：")

	var password string
	for {
		password, err = readPasswordMasked(stderr, "主密码: ")
		if err != nil {
			return err
		}

		if err := vault.ValidatePassword(password); err != nil {
			_, _ = fmt.Fprintf(stderr, "密码不符合要求: %s，请重新输入\n", err)
			continue
		}

		confirm, err := readPasswordMasked(stderr, "确认密码: ")
		if err != nil {
			return err
		}

		if password != confirm {
			_, _ = fmt.Fprintln(stderr, "两次输入的密码不一致，请重新输入")
			continue
		}

		break
	}

	// 密码输入完成，开始计时（不包含用户输入时间）
	start := time.Now()

	// 执行导出
	_, _ = fmt.Fprintln(stderr, "正在导出...")
	result, err := vault.Export(outDir, password)
	if err != nil {
		return err
	}

	// 输出结果
	_, _ = fmt.Fprintf(stdout, "✓ 导出成功\n")
	_, _ = fmt.Fprintf(stdout, "  文件: %s\n", result.FilePath)
	_, _ = fmt.Fprintf(stdout, "  大小: %d 字节\n", result.FileSize)
	_, _ = fmt.Fprintf(stdout, "  包含: %d 个 Host 配置, %d 个私钥\n", result.Hosts, result.Keys)
	_, _ = fmt.Fprintf(stdout, "\n请将此文件拷贝到目标机器，然后运行:\n")
	_, _ = fmt.Fprintf(stdout, "  fuckssh import %s\n", filepath.Base(result.FilePath))

	printCmdElapsed(stderr, time.Since(start))
	return nil
}
