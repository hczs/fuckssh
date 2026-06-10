package config

import (
	"os"
	"path/filepath"
	"testing"
)

func fixturePath(name string) string {
	return filepath.Join("..", "..", "testdata", "config", name)
}

func TestParse_singleHost_minimal(t *testing.T) {
	entries, err := ParseFile(fixturePath("single.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Alias != "myserver" {
		t.Errorf("Alias = %q, want myserver", e.Alias)
	}
	if e.HostName != "192.168.1.10" {
		t.Errorf("HostName = %q, want 192.168.1.10", e.HostName)
	}
	if e.User != "root" {
		t.Errorf("User = %q, want root", e.User)
	}
	if e.Port != "22" {
		t.Errorf("Port = %q, want default 22", e.Port)
	}
}

func TestParse_multipleHosts(t *testing.T) {
	entries, err := ParseFile(fixturePath("multiple.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}

	if entries[0].Alias != "srv1" || entries[0].Port != "2222" {
		t.Errorf("first host: %+v", entries[0])
	}
	if entries[1].Alias != "srv2" {
		t.Errorf("second Alias = %q, want srv2", entries[1].Alias)
	}
	if entries[1].HostName != "example.com" {
		t.Errorf("second HostName = %q, want example.com", entries[1].HostName)
	}
}

func TestParse_defaultPort22_whenOmitted(t *testing.T) {
	entries, err := ParseFile(fixturePath("default_port.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Port != "22" {
		t.Errorf("Port = %q, want 22", entries[0].Port)
	}
}

func TestParse_invalidLine_returnsErrorWithLineNumber(t *testing.T) {
	_, err := ParseFile(fixturePath("invalid.conf"))
	if err == nil {
		t.Fatal("expected parse error")
	}

	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("error type = %T, want *ParseError", err)
	}
	if pe.Line != 2 {
		t.Errorf("Line = %d, want 2", pe.Line)
	}
	if pe.File == "" {
		t.Error("File should be set")
	}
	if pe.Snippet == "" {
		t.Error("Snippet should include problem line")
	}
}

func TestParse_ignoresCommentAndBlankLines(t *testing.T) {
	entries, err := ParseFile(fixturePath("comments.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].HostName != "1.2.3.4" {
		t.Errorf("HostName = %q, want 1.2.3.4", entries[0].HostName)
	}
	if entries[0].Remark != "top comment" {
		t.Errorf("Remark = %q, want top comment", entries[0].Remark)
	}
}

func TestParse_associatesCommentAboveHost(t *testing.T) {
	entries, err := ParseFile(fixturePath("with_remark.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Remark != "生产环境主站" {
		t.Errorf("first Remark = %q, want 生产环境主站", entries[0].Remark)
	}
	if entries[1].Remark != "测试机" {
		t.Errorf("second Remark = %q, want 测试机", entries[1].Remark)
	}
}

func TestParse_hostWithMultipleAliases(t *testing.T) {
	entries, err := ParseFile(fixturePath("multiple.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	var multi *HostEntry
	for i := range entries {
		if entries[i].Alias == "srv2" {
			multi = &entries[i]
			break
		}
	}
	if multi == nil {
		t.Fatal("srv2 host not found")
		return
	}

	want := []string{"srv2", "srv2-alt"}
	if len(multi.Aliases) != len(want) {
		t.Fatalf("Aliases = %v, want %v", multi.Aliases, want)
	}
	for i, a := range want {
		if multi.Aliases[i] != a {
			t.Errorf("Aliases[%d] = %q, want %q", i, multi.Aliases[i], a)
		}
	}
}

func TestParse_emptyLineInsideHostBlock(t *testing.T) {
	entries, err := ParseFile(fixturePath("empty_line_in_block.conf"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Alias != "web" || entries[0].Port != "2222" {
		t.Errorf("first host: %+v", entries[0])
	}
	if entries[1].Alias != "db" || entries[1].User != "root" {
		t.Errorf("second host: %+v", entries[1])
	}
}

func TestParseFile_missingFile(t *testing.T) {
	_, err := ParseFile(filepath.Join(t.TempDir(), "missing.conf"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("err = %v, want IsNotExist", err)
	}
}
