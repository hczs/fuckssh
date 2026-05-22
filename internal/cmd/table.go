package cmd

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// formatBorderedTable 把表头与数据行格式化成带框线的 ASCII 表格。
// 使用 Unicode 框线字符；列宽按终端显示宽度计算（中文占 2 格）。
func formatBorderedTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}
	ncol := len(headers)
	widths := columnWidths(headers, rows)

	var b strings.Builder
	writeBorder(&b, "┌", "┬", "┐", widths)
	writeCells(&b, headers, widths)
	writeBorder(&b, "├", "┼", "┤", widths)
	for _, row := range rows {
		writeCells(&b, padRow(row, ncol), widths)
	}
	writeBorder(&b, "└", "┴", "┘", widths)
	return b.String()
}

func columnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = runewidth.StringWidth(h)
	}
	for _, row := range rows {
		for i := 0; i < len(headers) && i < len(row); i++ {
			if w := runewidth.StringWidth(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

func padRow(row []string, ncol int) []string {
	if len(row) >= ncol {
		return row[:ncol]
	}
	out := make([]string, ncol)
	copy(out, row)
	return out
}

func writeBorder(b *strings.Builder, left, mid, right string, widths []int) {
	b.WriteString(left)
	for i, w := range widths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(widths)-1 {
			b.WriteString(mid)
		}
	}
	b.WriteString(right)
	b.WriteByte('\n')
}

func writeCells(b *strings.Builder, cells []string, widths []int) {
	b.WriteString("│")
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteByte(' ')
		b.WriteString(padCell(cell, w))
		b.WriteByte(' ')
		b.WriteString("│")
	}
	b.WriteByte('\n')
}

func padCell(s string, width int) string {
	pad := width - runewidth.StringWidth(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}
