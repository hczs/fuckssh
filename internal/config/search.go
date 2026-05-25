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

func hostMatchesQuery(e HostEntry, query string) bool {
	for _, alias := range e.Aliases {
		if strings.Contains(strings.ToLower(alias), query) {
			return true
		}
	}
	if strings.Contains(strings.ToLower(e.HostName), query) {
		return true
	}
	return strings.Contains(strings.ToLower(e.Remark), query)
}
