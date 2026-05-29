package config

import (
	"path/filepath"
	"testing"
)

func Test_Search_matchesAlias(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	got := FilterHosts(entries, "srv2")
	if len(got) != 1 || got[0].Alias != "srv2" {
		t.Fatalf("FilterHosts = %+v, want srv2 only", got)
	}
}

func Test_Search_matchesHostName(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	got := FilterHosts(entries, "example")
	if len(got) != 1 || got[0].HostName != "example.com" {
		t.Fatalf("FilterHosts = %+v, want example.com host", got)
	}
}

func Test_Search_matchesIP(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	got := FilterHosts(entries, "10.0.0")
	if len(got) != 1 || got[0].HostName != "10.0.0.1" {
		t.Fatalf("FilterHosts = %+v, want 10.0.0.1 host", got)
	}
}

func Test_Search_caseInsensitiveAlias(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	got := FilterHosts(entries, "SRV2-ALT")
	if len(got) != 1 {
		t.Fatalf("FilterHosts = %+v, want case-insensitive match", got)
	}
}

func Test_Search_matchesRemark(t *testing.T) {
	entries := mustParseFixture(t, "with_remark.conf")
	got := FilterHosts(entries, "生产")
	if len(got) != 1 || got[0].Alias != "my-vps" {
		t.Fatalf("FilterHosts = %+v, want my-vps by remark", got)
	}
}

func Test_Search_noMatch_returnsEmpty(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	got := FilterHosts(entries, "nomatch-xyz")
	if len(got) != 0 {
		t.Fatalf("FilterHosts = %+v, want empty", got)
	}
}

// --- SearchHosts 测试 ---

func Test_SearchHosts_multiKeywordOR(t *testing.T) {
	// srv1 (10.0.0.1) 和 srv2 (example.com) 应通过 OR 匹配。
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{Keywords: []string{"srv1", "srv2"}}
	got := SearchHosts(entries, opts)
	if len(got) != 2 {
		t.Fatalf("SearchHosts OR = %d, want 2; got %+v", len(got), got)
	}
}

func Test_SearchHosts_singleKeyword(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{Keywords: []string{"example"}}
	got := SearchHosts(entries, opts)
	if len(got) != 1 || got[0].HostName != "example.com" {
		t.Fatalf("SearchHosts = %+v, want example.com only", got)
	}
}

func Test_SearchHosts_emptyKeywords_matchesAll(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{} // 无关键词
	got := SearchHosts(entries, opts)
	if len(got) != len(entries) {
		t.Fatalf("SearchHosts no-keywords = %d, want %d", len(got), len(entries))
	}
}

func Test_SearchHosts_filterByUser(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{User: "admin"}
	got := SearchHosts(entries, opts)
	if len(got) != 1 || got[0].Alias != "srv1" {
		t.Fatalf("SearchHosts --user admin = %+v, want srv1 only", got)
	}
}

func Test_SearchHosts_filterByUser_caseInsensitive(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{User: "ADMIN"}
	got := SearchHosts(entries, opts)
	if len(got) != 1 || got[0].Alias != "srv1" {
		t.Fatalf("SearchHosts --user ADMIN = %+v, want srv1 only", got)
	}
}

func Test_SearchHosts_filterByPort(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{Port: "2222"}
	got := SearchHosts(entries, opts)
	if len(got) != 1 || got[0].Alias != "srv1" {
		t.Fatalf("SearchHosts --port 2222 = %+v, want srv1 only", got)
	}
}

func Test_SearchHosts_filterByHost(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{Host: "example"}
	got := SearchHosts(entries, opts)
	if len(got) != 1 || got[0].HostName != "example.com" {
		t.Fatalf("SearchHosts --host example = %+v, want example.com only", got)
	}
}

func Test_SearchHosts_keywordsPlusFilter(t *testing.T) {
	// 关键词匹配 srv1 和 srv2，但 --user admin 只保留 srv1。
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{Keywords: []string{"srv1", "srv2"}, User: "admin"}
	got := SearchHosts(entries, opts)
	if len(got) != 1 || got[0].Alias != "srv1" {
		t.Fatalf("SearchHosts OR+filter = %+v, want srv1 only", got)
	}
}

func Test_SearchHosts_noMatch(t *testing.T) {
	entries := mustParseFixture(t, "multiple.conf")
	opts := SearchOptions{Keywords: []string{"nonexistent"}}
	got := SearchHosts(entries, opts)
	if len(got) != 0 {
		t.Fatalf("SearchHosts = %+v, want empty", got)
	}
}

func mustParseFixture(t *testing.T, name string) []HostEntry {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "config", name)
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	return entries
}
