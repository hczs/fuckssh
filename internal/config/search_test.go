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

func mustParseFixture(t *testing.T, name string) []HostEntry {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "config", name)
	entries, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	return entries
}
