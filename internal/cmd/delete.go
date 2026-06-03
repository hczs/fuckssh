package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/keys"
	"github.com/fuckssh/fuckssh/internal/platform"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var deleteForce bool

// confirmDeleteFn 可在测试中注入，默认从 stdin 读取 y/N 确认。
var confirmDeleteFn = confirmDelete

var deleteCmd = &cobra.Command{
	Use:          "delete <alias>",
	Aliases:      []string{"d"},
	Short:        "Delete a Host entry by alias",
	Long:         "Remove a Host entry from ssh config by its alias. A backup is created before deletion.",
	Args:         deleteArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete(args[0], cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func deleteArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return fmt.Errorf("请指定要删除的 Host 别名，例如: fuckssh delete myserver")
	}
	return nil
}

func runDelete(alias string, stdout, stderr io.Writer) error {
	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	maybeWarnInclude(stderr, configPath)
	entries, err := config.ParseFile(configPath)
	if err != nil {
		return err
	}

	// 查找目标条目。
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

	// 确认删除。
	ok, err := confirmDeleteFn(stderr, target.Alias, target.User, target.HostName)
	if err != nil {
		return err
	}
	if !ok {
		_, _ = fmt.Fprint(stderr, i18n.T(i18n.KeyDeleteCancelled))
		return nil
	}

	// 执行删除。
	if err := config.DeleteHost(configPath, alias); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, i18n.T(i18n.KeyDeleteSuccess), target.Alias, "")

	// 如果 IdentityFile 是 fuckssh 管理的密钥，一并删除。
	removeManagedKey(target.IdentityFile, stdout)

	return nil
}

// removeManagedKey 检查 IdentityFile 是否为 fuckssh 管理的密钥，若是则删除。
func removeManagedKey(identityFile string, stdout io.Writer) {
	if identityFile == "" {
		return
	}
	expanded, err := platform.ExpandPath(identityFile)
	if err != nil || !keys.IsManagedKeyPath(expanded) {
		return
	}
	if err := keys.RemoveKeyPair(expanded); err != nil {
		return
	}
	_, _ = fmt.Fprintf(stdout, i18n.T(i18n.KeyDeleteKeyRemoved), expanded)
}

// confirmDelete 提示用户确认删除操作。
// 返回 true 表示确认，false 表示取消。
func confirmDelete(stderr io.Writer, alias, user, host string) (bool, error) {
	if deleteForce {
		return true, nil
	}

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return false, errors.New(strings.TrimRight(i18n.T(i18n.KeyDeleteForceHint), "\n"))
	}

	_, _ = fmt.Fprintf(stderr, i18n.T(i18n.KeyDeleteConfirm), alias, user, host)

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation prompt")
}
