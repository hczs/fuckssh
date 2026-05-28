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
		return fmt.Errorf("%s", i18n.T(i18n.KeySearchEmptyQ))
	}
	return nil
}

func runSearchCmd(args []string, stdout, stderr io.Writer) error {
	query := strings.TrimSpace(args[0])

	path, err := ConfigFilePath()
	if err != nil {
		return err
	}

	maybeWarnInclude(stderr, path)
	entries, err := config.ParseFile(path)
	if err != nil {
		return err
	}

	matched := config.FilterHosts(entries, query)
	return WriteHostsReport(stdout, stderr, path, matched, query)
}
