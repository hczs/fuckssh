package config

import "fmt"

// ErrHostNotFound 表示 config 中未找到指定别名的 Host 条目。
var ErrHostNotFound = fmt.Errorf("config: host alias not found")

// HostEntry 表示 ssh config 中一个 Host 块的结构化字段（MVP 受限解析）。
type HostEntry struct {
	// Alias 为 Host 行上的第一个别名，供表格主键展示。
	Alias string
	// Aliases 为 Host 行上的全部别名（含 Alias）。
	Aliases      []string
	HostName     string
	User         string
	Port         string
	IdentityFile string
	// Remark 为紧邻该 Host 块上方的 # 注释行（多行以换行连接）。
	Remark string
	// LineStart 为该 Host 指令所在行号（1-based），便于报错定位。
	LineStart int
}
