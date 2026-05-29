package cmd

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fuckssh/fuckssh/internal/config"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

// highlightStyle 用于搜索结果中匹配文本的高亮样式（粗体黄色）。
var highlightStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))

// WriteHostsReport 向 stderr 写元信息，向 stdout 写表格或空状态文案。
// highlight 为 true 时对匹配关键词启用高亮（仅 TTY 场景传 true）。
func WriteHostsReport(stdout, stderr io.Writer, configPath string, entries []config.HostEntry, query string, highlight bool, keywords []string) error {
	if err := writeHostsReportMeta(stderr, configPath, len(entries), query); err != nil {
		return err
	}

	if len(entries) == 0 {
		return writeHostsEmpty(stdout, query)
	}

	var kw []string
	if highlight {
		kw = keywords
	}
	_, err := fmt.Fprint(stdout, formatHostsTable(entries, kw))
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
	if e.Remark == "" {
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

// formatHostsTable 生成表格字符串。keywords 非空时对匹配文本启用高亮。
func formatHostsTable(entries []config.HostEntry, keywords []string) string {
	headers := []string{
		i18n.T(i18n.KeyTableAlias),
		i18n.T(i18n.KeyTableHostname),
		i18n.T(i18n.KeyTablePort),
		i18n.T(i18n.KeyTableUser),
		i18n.T(i18n.KeyTableRemark),
	}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		rows[i] = []string{
			highlightText(formatHostAliases(e), keywords),
			highlightText(e.HostName, keywords),
			highlightText(e.Port, keywords),
			highlightText(e.User, keywords),
			highlightText(formatHostRemark(e), keywords),
		}
	}
	return formatBorderedTable(headers, rows)
}

// highlightText 对 text 中匹配 keywords 的子串应用高亮样式。
// keywords 应为小写形式；空列表时原样返回。
func highlightText(text string, keywords []string) string {
	if len(keywords) == 0 || text == "" {
		return text
	}
	return highlightMatches(text, keywords)
}

// highlightMatches 对 text 中所有匹配 keywords 的子串应用 lipgloss 高亮。
// 采用单次遍历：先收集所有匹配区间，合并重叠后一次性构建结果。
func highlightMatches(text string, keywords []string) string {
	// 为每个关键词编译正则，收集所有匹配区间。
	type interval struct{ start, end int }
	var intervals []interval

	for _, kw := range keywords {
		re, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(kw))
		if err != nil {
			continue
		}
		for _, loc := range re.FindAllStringIndex(text, -1) {
			intervals = append(intervals, interval{loc[0], loc[1]})
		}
	}

	if len(intervals) == 0 {
		return text
	}

	// 按起始位置排序，合并重叠区间。
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].start < intervals[j].start
	})
	merged := []interval{intervals[0]}
	for _, iv := range intervals[1:] {
		last := &merged[len(merged)-1]
		if iv.start <= last.end {
			if iv.end > last.end {
				last.end = iv.end
			}
		} else {
			merged = append(merged, iv)
		}
	}

	// 按区间构建结果：普通文本直接追加，匹配文本用 lipgloss 高亮。
	var b strings.Builder
	b.Grow(len(text) + len(merged)*20)
	pos := 0
	for _, iv := range merged {
		if iv.start > pos {
			b.WriteString(text[pos:iv.start])
		}
		b.WriteString(highlightStyle.Render(text[iv.start:iv.end]))
		pos = iv.end
	}
	if pos < len(text) {
		b.WriteString(text[pos:])
	}
	return b.String()
}
