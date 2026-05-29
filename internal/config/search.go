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

// SearchOptions 定义搜索条件，支持多关键词 OR 匹配与字段精确过滤。
type SearchOptions struct {
	// Keywords 为已去空的关键词列表（OR 逻辑，任一命中即匹配）。
	Keywords []string
	// User 按 SSH 用户名精确过滤（大小写不敏感）。
	User string
	// Host 按 HostName 子串过滤（大小写不敏感）。
	Host string
	// Port 按端口号精确过滤。
	Port string
}

// SearchHosts 根据 opts 对 entries 进行过滤。
// 先用 Keywords 做 OR 子串匹配，再用字段过滤器收窄结果。
func SearchHosts(entries []HostEntry, opts SearchOptions) []HostEntry {
	var matched []HostEntry
	for _, e := range entries {
		if !matchesKeywords(e, opts.Keywords) {
			continue
		}
		if !matchesFieldFilters(e, opts) {
			continue
		}
		matched = append(matched, e)
	}
	return matched
}

// matchesKeywords 判断 entry 是否命中 keywords 中的任意一个（OR 逻辑）。
// keywords 应为小写形式。空列表视为全部命中。
func matchesKeywords(e HostEntry, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	for _, kw := range keywords {
		if hostMatchesQuery(e, kw) {
			return true
		}
	}
	return false
}

// matchesFieldFilters 检查 entry 是否满足所有字段级过滤条件。
func matchesFieldFilters(e HostEntry, opts SearchOptions) bool {
	if opts.User != "" && !strings.EqualFold(e.User, opts.User) {
		return false
	}
	if opts.Host != "" && !strings.Contains(strings.ToLower(e.HostName), strings.ToLower(opts.Host)) {
		return false
	}
	if opts.Port != "" && e.Port != opts.Port {
		return false
	}
	return true
}
