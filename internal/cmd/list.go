package cmd

import (
	"io"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runListCmd(cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func runListCmd(stdout, stderr io.Writer) error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	return runList(path, stdout, stderr)
}

func runList(configPath string, stdout, stderr io.Writer) error {
	maybeWarnInclude(stderr, configPath)
	entries, err := config.ParseFile(configPath)
	if err != nil {
		return err
	}
	return WriteHostsReport(stdout, stderr, configPath, entries, "", false, nil)
}
