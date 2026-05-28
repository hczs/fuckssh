package config

import "strings"

// FilterHosts 按关键词过滤：匹配任一别名、HostName（大小写不敏感子串）。
func FilterHosts(entries []HostEntry, query string) []HostEntry {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}

	var matched []HostEntry
	for _, e := range entries {
		if hostMatchesQuery(e, q) {
			matched = append(matched, e)
		}
	}
	return matched
}

// hostMatchesQuery 判断 entry 是否匹配 query（query 已由调用方 ToLower）。
// 每个字段只调一次 ToLower，避免重复转换。
func hostMatchesQuery(e HostEntry, query string) bool {
	for _, alias := range e.Aliases {
		if strings.Contains(strings.ToLower(alias), query) {
			return true
		}
	}
	host := strings.ToLower(e.HostName)
	if strings.Contains(host, query) {
		return true
	}
	remark := strings.ToLower(e.Remark)
	return strings.Contains(remark, query)
}
