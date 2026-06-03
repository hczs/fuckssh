package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/wizard"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     "edit <alias>",
	Aliases: []string{"e"},
	Short:   i18n.T(i18n.KeyEditShort),
	Long:    i18n.T(i18n.KeyEditLong),
	Args:    editArgs,
	// 向导内 Ctrl+C 由 executeWithArgs 输出单行取消提示，不附带 usage。
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runEdit(args[0], cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func editArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return fmt.Errorf("请指定要编辑的 Host 别名，例如: fuckssh edit myserver")
	}
	return nil
}

func runEdit(alias string, stdout, stderr io.Writer) error {
	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	maybeWarnInclude(stderr, configPath)

	// 解析 config，找到目标条目。
	entries, err := config.ParseFile(configPath)
	if err != nil {
		return err
	}

	var target *config.HostEntry
	for i := range entries {
		for _, a := range entries[i].Aliases {
			if strings.EqualFold(a, alias) {
				target = &entries[i]
				break
			}
		}
		if target != nil {
			break
		}
	}
	if target == nil {
		return fmt.Errorf("%w: %q", config.ErrHostNotFound, alias)
	}

	// 启动编辑向导（预填现有值），向导内部处理"返回修改"重试循环。
	skipElapsedOutput = true
	editIn, err := runEditWizardFn(configPath, *target)
	if err != nil {
		return err
	}

	// 执行行级编辑。
	if err := config.EditHost(configPath, alias, config.HostEntry{
		Alias:        editIn.Alias,
		HostName:     editIn.HostName,
		User:         editIn.User,
		Port:         editIn.Port,
		IdentityFile: editIn.IdentityFile,
		Remark:       editIn.Remark,
	}); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, i18n.T(i18n.KeyEditSuccess), editIn.Alias)
	_, _ = fmt.Fprintf(stdout, "ssh %s\n", editIn.Alias)
	return nil
}

// runEditWizardFn 可在测试中注入，默认调用编辑向导。
var runEditWizardFn = wizard.RunEditWizard
