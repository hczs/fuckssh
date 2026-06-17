package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/fuckssh/fuckssh/internal/i18n"
	"github.com/fuckssh/fuckssh/internal/update"
	"github.com/spf13/cobra"
)

var (
	updateVersion string
	updateCheck   bool
)

// runUpdateFn 可在测试中注入。
var runUpdateFn = runUpdate

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade"},
	Short:   "Update fuckssh to the latest release",
	Long:    "Download the latest fuckssh release from GitHub and replace the current binary.",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdateFn(cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "target version tag, e.g. v0.6.0 (default: latest)")
	updateCmd.Flags().BoolVarP(&updateCheck, "check", "c", false, "check for updates without installing")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(stdout, stderr io.Writer) error {
	dest, err := update.ResolveExecutable()
	if err != nil {
		return err
	}

	result, err := update.Run(update.Options{
		Version:    updateVersion,
		CheckOnly:  updateCheck,
		CurrentVer: version,
		DestPath:   dest,
		Out:        stdout,
		ErrOut:     stderr,
	})
	if err != nil {
		if errors.Is(err, update.ErrAlreadyLatest) {
			_, _ = fmt.Fprintf(stdout, i18n.T(i18n.KeyUpdateAlreadyLatest), result.CurrentVersion)
			return nil
		}
		return err
	}

	if result.AlreadyLatest {
		_, _ = fmt.Fprintf(stdout, i18n.T(i18n.KeyUpdateAlreadyLatest), result.CurrentVersion)
		return nil
	}

	if updateCheck {
		_, _ = fmt.Fprintf(stdout, i18n.T(i18n.KeyUpdateAvailable), result.CurrentVersion, result.TargetVersion)
		return nil
	}

	_, _ = fmt.Fprintf(stdout, i18n.T(i18n.KeyUpdateSuccess), result.TargetVersion, dest)
	return nil
}
