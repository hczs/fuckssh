package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// WriteHostsReport 向 stderr 写元信息，向 stdout 写表格或空状态文案。
func WriteHostsReport(stdout, stderr io.Writer, configPath string, entries []config.HostEntry, query string) error {
	if err := writeHostsReportMeta(stderr, configPath, len(entries), query); err != nil {
		return err
	}

	if len(entries) == 0 {
		return writeHostsEmpty(stdout, query)
	}

	_, err := fmt.Fprint(stdout, formatHostsTable(entries))
	return err
}

func writeHostsReportMeta(stderr io.Writer, configPath string, count int, query string) error {
	if query != "" {
		_, err := fmt.Fprintf(stderr, i18n.T(i18n.KeySearchMeta), query, count)
		return err
	}
	if _, err := fmt.Fprintf(stderr, i18n.T(i18n.KeyListReading), configPath); err != nil {
		return err
	}
	_, err := fmt.Fprintf(stderr, i18n.T(i18n.KeyListTotal), count)
	return err
}

func writeHostsEmpty(stdout io.Writer, query string) error {
	if query != "" {
		if _, err := fmt.Fprintf(stdout, i18n.T(i18n.KeySearchNoMatch), query); err != nil {
			return err
		}
		_, err := fmt.Fprint(stdout, i18n.T(i18n.KeySearchHint))
		return err
	}
	if _, err := fmt.Fprint(stdout, i18n.T(i18n.KeyListEmpty)); err != nil {
		return err
	}
	_, err := fmt.Fprint(stdout, i18n.T(i18n.KeyListEmptyCTA))
	return err
}

// formatHostAliases 将 Host 行上的全部别名用逗号连接，供表格「别名」列展示。
func formatHostRemark(e config.HostEntry) string {
	if strings.TrimSpace(e.Remark) == "" {
		return "-"
	}
	return e.Remark
}

func formatHostAliases(e config.HostEntry) string {
	if len(e.Aliases) == 0 {
		return e.Alias
	}
	return strings.Join(e.Aliases, ", ")
}

func formatHostsTable(entries []config.HostEntry) string {
	headers := []string{
		i18n.T(i18n.KeyTableAlias),
		i18n.T(i18n.KeyTableHostname),
		i18n.T(i18n.KeyTablePort),
		i18n.T(i18n.KeyTableUser),
		i18n.T(i18n.KeyTableRemark),
	}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		rows[i] = []string{formatHostAliases(e), e.HostName, e.Port, e.User, formatHostRemark(e)}
	}
	return formatBorderedTable(headers, rows)
}
