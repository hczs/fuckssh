package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search hosts by alias, hostname, or IP",
	Args:  searchArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearchCmd(args, cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func searchArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
		return err
	}
	if strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("search: 需要非空关键词")
	}
	return nil
}

func runSearchCmd(args []string, stdout, _ io.Writer) error {
	query := strings.TrimSpace(args[0])
	if query == "" {
		return fmt.Errorf("search: 需要非空关键词")
	}

	path, err := ConfigFilePath()
	if err != nil {
		return err
	}

	entries, err := config.ParseFile(path)
	if err != nil {
		return err
	}

	matched := config.FilterHosts(entries, query)
	if len(matched) == 0 {
		_, err = fmt.Fprintf(stdout, "未找到匹配 %q 的 Host 条目\n", query)
		return err
	}

	_, err = fmt.Fprint(stdout, FormatHosts(matched))
	return err
}
