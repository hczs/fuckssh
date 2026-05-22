package cmd

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List hosts from ssh config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runListCmd(cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func runListCmd(stdout, _ io.Writer) error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	return runList(path, stdout)
}

func runList(configPath string, w io.Writer) error {
	entries, err := config.ParseFile(configPath)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, FormatHosts(entries))
	return err
}

// FormatHosts 将 Host 条目格式化为 tabwriter 对齐的表格文本。
func FormatHosts(entries []config.HostEntry) string {
	if len(entries) == 0 {
		return "未找到 Host 条目\n"
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ALIAS\tHOSTNAME\tPORT\tUSER")
	for _, e := range entries {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", e.Alias, e.HostName, e.Port, e.User)
	}
	_ = tw.Flush()
	return buf.String()
}
