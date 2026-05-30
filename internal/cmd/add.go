package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/sshclient"
	"github.com/fuckssh/fuckssh/internal/wizard"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// 非交互模式 flag 变量
var (
	addHost     string
	addUser     string
	addPort     string
	addAlias    string
	addPassword string
	addKeyFile  string
	addRemark   string
)

// checkSSHFn 可在测试中注入，默认调用 sshclient.CheckSSH。
var checkSSHFn = sshclient.CheckSSH

// runWizardFn 可在测试中注入，默认调用交互式向导（传入 config 路径）。
var runWizardFn = wizard.Run

// skipElapsedOutput 为 true 时，executeWithArgs 跳过默认耗时输出（由子命令自行处理）。
var skipElapsedOutput bool

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a VPS host via interactive wizard",
	Long: "Run the interactive wizard to generate keys, update ssh config, and optionally deploy a public key.\n\n" +
		"Non-interactive mode (when --host is provided):\n" +
		"  fuckssh add --host 1.2.3.4 --user root --password pass --alias myserver\n" +
		"  fuckssh add --host 1.2.3.4 --user root --identity-file ~/.ssh/id_ed25519",
	// 向导内 Ctrl+C 由 executeWithArgs 输出单行取消提示，不附带 usage。
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

// isNonInteractive 判断是否为非交互模式（提供了 --host flag）。
func isNonInteractive() bool {
	return addHost != ""
}

func runAdd(stdout, stderr io.Writer) error {
	if err := requireSSH(stderr); err != nil {
		return err
	}

	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if dir == "" || dir == "." {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return &os.PathError{Op: "mkdir", Path: dir, Err: err}
	}

	// 非交互模式：直接用 flag 构造输入，跳过 TUI，耗时从这里开始
	if isNonInteractive() {
		start := time.Now()
		err := runAddNonInteractive(stdout, stderr, configPath)
		printCmdElapsed(stderr, time.Since(start))
		return err
	}

	// 交互模式：耗时从用户确认执行后开始计算，不包含填表时间
	skipElapsedOutput = true
	result, err := runWizardFn(configPath)
	if err != nil {
		return err
	}

	// 用户已确认，开始计时
	start := time.Now()

	if result.PasswordFlowComplete {
		printAddSuccess(stdout, stderr, configPath, result)
		printCmdElapsed(stderr, time.Since(start))
		return nil
	}

	if err := wizard.RunKeyFlow(configPath, result); err != nil {
		return err
	}

	printAddSuccess(stdout, stderr, configPath, result)
	printCmdElapsed(stderr, time.Since(start))
	return nil
}

// runAddNonInteractive 解析 flag 并直接执行，不启动 TUI。
func runAddNonInteractive(stdout, stderr io.Writer, configPath string) error {
	in := wizard.AddInput{
		Alias:        strings.TrimSpace(addAlias),
		HostName:     strings.TrimSpace(addHost),
		User:         strings.TrimSpace(addUser),
		Port:         strings.TrimSpace(addPort),
		Password:     strings.TrimSpace(addPassword),
		IdentityFile: strings.TrimSpace(addKeyFile),
		Remark:       strings.TrimSpace(addRemark),
	}

	// 推断认证模式
	switch {
	case in.Password != "" && in.IdentityFile != "":
		return fmt.Errorf("%w: --password and --identity-file are mutually exclusive", wizard.ErrInvalidInput)
	case in.Password != "":
		in.Mode = wizard.ModePassword
	case in.IdentityFile != "":
		in.Mode = wizard.ModeKey
	default:
		return fmt.Errorf("%w: --password or --identity-file is required in non-interactive mode", wizard.ErrInvalidInput)
	}

	result, err := wizard.RunNonInteractive(configPath, in)
	if err != nil {
		return err
	}

	// 密钥模式需要额外执行 RunKeyFlow 写盘（进度从第 2 步开始，总计 4 步）
	if !result.PasswordFlowComplete {
		if err := wizard.RunKeyFlowWithProgress(configPath, result, 1, 4); err != nil {
			return err
		}
	}

	printAddSuccess(stdout, stderr, configPath, result)
	return nil
}

func printAddSuccess(stdout, stderr io.Writer, configPath string, result *wizard.WizardResult) {
	wizard.WriteAddSuccessSummary(stderr, result, configPath)
	_, _ = fmt.Fprintf(stdout, "ssh %s\n", result.Alias)
	if result.PasswordFlowComplete && isTerminalWriter(stdout) {
		_, _ = fmt.Fprintf(stderr, "%s\n", i18n.T(i18n.KeySummaryReadyHint))
	}
}

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd())
}

// requireSSH 检测系统 ssh；缺失时输出警告与安装指引并终止。
func requireSSH(stderr io.Writer) error {
	_, err := checkSSHFn()
	if err == nil {
		return nil
	}
	if errors.Is(err, sshclient.ErrSSHNotFound) {
		_, _ = fmt.Fprintf(stderr, "%s\n%s\n", i18n.T(i18n.KeySSHMissingWarning), i18n.InstallOpenSSHGuide())
		return err
	}
	return err
}

func init() {
	addCmd.Flags().StringVarP(&addHost, "host", "H", "", "host address (enables non-interactive mode)")
	addCmd.Flags().StringVarP(&addUser, "user", "u", "", "login user (default: root)")
	addCmd.Flags().StringVarP(&addPort, "port", "p", "", "SSH port (default: 22)")
	addCmd.Flags().StringVarP(&addAlias, "alias", "a", "", "host alias in ssh config (auto-generated from host if empty)")
	addCmd.Flags().StringVarP(&addPassword, "password", "P", "", "SSH password (triggers password mode)")
	addCmd.Flags().StringVarP(&addKeyFile, "identity-file", "i", "", "path to private key (triggers key mode)")
	addCmd.Flags().StringVarP(&addRemark, "remark", "r", "", "optional remark")
}
