package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search [query ...]",
	Aliases: []string{"s"},
	Short:   "Search hosts by alias, hostname, or IP",
	Long:    "Match hosts by alias, HostName, or IP substring. Multiple keywords are OR-ed.",
	Args:    searchArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearchCmd(args, cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

// searchUserFlag / searchHostFlag / searchPortFlag 为 search 子命令的本地过滤 flag。
var (
	searchUserFlag string
	searchHostFlag string
	searchPortFlag string
)

func init() {
	searchCmd.Flags().StringVar(&searchUserFlag, "user", "", "")
	searchCmd.Flags().StringVar(&searchHostFlag, "host", "", "")
	searchCmd.Flags().StringVar(&searchPortFlag, "port", "", "")
}

func searchArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
		return err
	}
	for _, a := range args {
		if strings.TrimSpace(a) != "" {
			return nil
		}
	}
	return fmt.Errorf("%s", i18n.T(i18n.KeySearchEmptyQ))
}

func runSearchCmd(args []string, stdout, stderr io.Writer) error {
	// 收集去重后的非空关键词（小写）。
	keywords := make([]string, 0, len(args))
	seen := make(map[string]bool, len(args))
	for _, a := range args {
		kw := strings.ToLower(strings.TrimSpace(a))
		if kw != "" && !seen[kw] {
			seen[kw] = true
			keywords = append(keywords, kw)
		}
	}

	path, err := ConfigFilePath()
	if err != nil {
		return err
	}

	maybeWarnInclude(stderr, path)
	entries, err := config.ParseFile(path)
	if err != nil {
		return err
	}

	opts := config.SearchOptions{
		Keywords: keywords,
		User:     searchUserFlag,
		Host:     searchHostFlag,
		Port:     searchPortFlag,
	}
	matched := config.SearchHosts(entries, opts)

	// 构造用于显示的查询字符串。
	query := strings.Join(keywords, " | ")
	highlight := isTerminalWriter(stdout) && query != ""
	return WriteHostsReport(stdout, stderr, path, matched, query, highlight, keywords)
}

// resetSearchFlags 清除 flag 值，避免测试间残留。
func resetSearchFlags() {
	searchUserFlag = ""
	searchHostFlag = ""
	searchPortFlag = ""
}
