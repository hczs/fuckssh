package cmd

import (
	"fmt"
	"io"

	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// WriteHostsReport 向 stderr 写元信息，向 stdout 写表格或空状态文案。
func WriteHostsReport(stdout, stderr io.Writer, configPath string, entries []config.HostEntry, query string) error {
	if query != "" {
		_, _ = fmt.Fprintf(stderr, i18n.T(i18n.KeySearchMeta), query, len(entries))
	} else {
		_, _ = fmt.Fprintf(stderr, i18n.T(i18n.KeyListReading), configPath)
		_, _ = fmt.Fprintf(stderr, i18n.T(i18n.KeyListTotal), len(entries))
	}

	if len(entries) == 0 {
		if query != "" {
			_, _ = fmt.Fprintf(stdout, i18n.T(i18n.KeySearchNoMatch), query)
			_, _ = fmt.Fprint(stdout, i18n.T(i18n.KeySearchHint))
		} else {
			_, _ = fmt.Fprint(stdout, i18n.T(i18n.KeyListEmpty))
			_, _ = fmt.Fprint(stdout, i18n.T(i18n.KeyListEmptyCTA))
		}
		return nil
	}

	_, err := fmt.Fprint(stdout, formatHostsTable(entries))
	return err
}

func formatHostsTable(entries []config.HostEntry) string {
	headers := []string{
		i18n.T(i18n.KeyTableAlias),
		i18n.T(i18n.KeyTableHostname),
		i18n.T(i18n.KeyTablePort),
		i18n.T(i18n.KeyTableUser),
	}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		rows[i] = []string{e.Alias, e.HostName, e.Port, e.User}
	}
	return formatBorderedTable(headers, rows)
}
