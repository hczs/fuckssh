package cmd

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// ansiEscape 匹配 ANSI 转义序列，用于计算显示宽度时剥离。
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// formatBorderedTable 把表头与数据行格式化成带框线的 ASCII 表格。
// 使用 Unicode 框线字符；列宽按终端显示宽度计算（中文占 2 格）。
func formatBorderedTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}
	ncol := len(headers)
	asciiOnly := tableIsASCII(headers, rows)
	widths := columnWidths(headers, rows, asciiOnly)

	var b strings.Builder
	b.Grow(estimateTableSize(len(rows), widths))
	writeBorder(&b, "┌", "┬", "┐", widths)
	writeCells(&b, headers, widths, asciiOnly)
	writeBorder(&b, "├", "┼", "┤", widths)
	for _, row := range rows {
		writeCells(&b, padRow(row, ncol), widths, asciiOnly)
	}
	writeBorder(&b, "└", "┴", "┘", widths)
	return b.String()
}

func tableIsASCII(headers []string, rows [][]string) bool {
	for _, h := range headers {
		if !isASCII(h) {
			return false
		}
	}
	for _, row := range rows {
		for _, cell := range row {
			if !isASCII(cell) {
				return false
			}
		}
	}
	return true
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func displayWidth(s string, asciiOnly bool) int {
	// 剥离 ANSI 转义序列后再计算显示宽度。
	stripped := ansiEscape.ReplaceAllString(s, "")
	if asciiOnly || isASCII(stripped) {
		return len(stripped)
	}
	return runewidth.StringWidth(stripped)
}

func columnWidths(headers []string, rows [][]string, asciiOnly bool) []int {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = displayWidth(h, asciiOnly)
	}
	for _, row := range rows {
		for i := 0; i < len(headers) && i < len(row); i++ {
			if w := displayWidth(row[i], asciiOnly); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

func estimateTableSize(rowCount int, widths []int) int {
	sum := 0
	for _, w := range widths {
		sum += w + 3
	}
	// 顶框 + 表头 + 分隔 + 数据行 + 底框
	return (rowCount + 4) * (sum + 1)
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

func writeCells(b *strings.Builder, cells []string, widths []int, asciiOnly bool) {
	b.WriteString("│")
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteByte(' ')
		b.WriteString(padCell(cell, w, asciiOnly))
		b.WriteByte(' ')
		b.WriteString("│")
	}
	b.WriteByte('\n')
}

func padCell(s string, width int, asciiOnly bool) string {
	pad := width - displayWidth(s, asciiOnly)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}
