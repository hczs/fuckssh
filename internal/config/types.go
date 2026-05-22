package config

// HostEntry 表示 ssh config 中一个 Host 块的结构化字段（MVP 受限解析）。
type HostEntry struct {
	// Alias 为 Host 行上的第一个别名，供表格主键展示。
	Alias string
	// Aliases 为 Host 行上的全部别名（含 Alias）。
	Aliases []string
	HostName     string
	User         string
	Port         string
	IdentityFile string
	// LineStart 为该 Host 指令所在行号（1-based），便于报错定位。
	LineStart int
}
