package config

import (
	"strings"
	"testing"
)

func TestFindConflicts(t *testing.T) {
	existing := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "192.168.1.100", User: "root", Port: "22"},
		{Alias: "prod-db", Aliases: []string{"prod-db"}, HostName: "10.0.0.5", User: "admin", Port: "2222"},
	}

	incoming := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "192.168.1.200", User: "ubuntu", Port: "22"},
		{Alias: "staging", Aliases: []string{"staging"}, HostName: "172.16.0.1", User: "dev", Port: "22"},
	}

	conflicts := FindConflicts(existing, incoming)

	if len(conflicts) != 1 {
		t.Fatalf("期望 1 个冲突，got %d", len(conflicts))
	}

	if conflicts[0].Alias != "myserver" {
		t.Errorf("冲突别名不匹配: got %q", conflicts[0].Alias)
	}
	if conflicts[0].Existing.HostName != "192.168.1.100" {
		t.Errorf("现有 HostName 不匹配: got %q", conflicts[0].Existing.HostName)
	}
	if conflicts[0].Incoming.HostName != "192.168.1.200" {
		t.Errorf("导入 HostName 不匹配: got %q", conflicts[0].Incoming.HostName)
	}
}

func TestFindConflictsCaseSensitive(t *testing.T) {
	// OpenSSH Host 别名大小写敏感，MyServer 和 myserver 是不同的 host
	existing := []HostEntry{
		{Alias: "MyServer", Aliases: []string{"MyServer"}, HostName: "192.168.1.100", User: "root"},
	}

	incoming := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "10.0.0.1", User: "user"},
	}

	conflicts := FindConflicts(existing, incoming)
	if len(conflicts) != 0 {
		t.Fatalf("大小写不同的别名不应视为冲突，got %d", len(conflicts))
	}
}

func TestFindConflictsNone(t *testing.T) {
	existing := []HostEntry{
		{Alias: "server1", Aliases: []string{"server1"}, HostName: "1.1.1.1", User: "root"},
	}

	incoming := []HostEntry{
		{Alias: "server2", Aliases: []string{"server2"}, HostName: "2.2.2.2", User: "user"},
	}

	conflicts := FindConflicts(existing, incoming)
	if len(conflicts) != 0 {
		t.Fatalf("无冲突时应返回空，got %d", len(conflicts))
	}
}

func TestMergeHostsNoConflict(t *testing.T) {
	existing := []HostEntry{
		{Alias: "server1", Aliases: []string{"server1"}, HostName: "1.1.1.1", User: "root"},
	}

	incoming := []HostEntry{
		{Alias: "server2", Aliases: []string{"server2"}, HostName: "2.2.2.2", User: "user"},
	}

	merged, result := MergeHosts(existing, incoming, nil)

	if len(merged) != 2 {
		t.Fatalf("合并后期望 2 个 Host，got %d", len(merged))
	}
	if len(result.Imported) != 1 || result.Imported[0] != "server2" {
		t.Errorf("导入结果不匹配: %v", result.Imported)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("不应有跳过: %v", result.Skipped)
	}
}

func TestMergeHostsOverwrite(t *testing.T) {
	existing := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "1.1.1.1", User: "root"},
	}

	incoming := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "2.2.2.2", User: "user"},
	}

	conflicts := map[string]ConflictInfo{
		"myserver": {Alias: "myserver", Action: ConflictOverwrite},
	}

	merged, result := MergeHosts(existing, incoming, conflicts)

	if len(merged) != 1 {
		t.Fatalf("覆盖后期望 1 个 Host，got %d", len(merged))
	}
	if merged[0].HostName != "2.2.2.2" {
		t.Errorf("覆盖后 HostName 应为 2.2.2.2，got %q", merged[0].HostName)
	}
	if len(result.Overwrite) != 1 {
		t.Errorf("覆盖结果不匹配: %v", result.Overwrite)
	}
}

func TestMergeHostsSkip(t *testing.T) {
	existing := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "1.1.1.1", User: "root"},
	}

	incoming := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "2.2.2.2", User: "user"},
	}

	conflicts := map[string]ConflictInfo{
		"myserver": {Alias: "myserver", Action: ConflictSkip},
	}

	merged, result := MergeHosts(existing, incoming, conflicts)

	if len(merged) != 1 {
		t.Fatalf("跳过后期望 1 个 Host，got %d", len(merged))
	}
	if merged[0].HostName != "1.1.1.1" {
		t.Errorf("跳过后 HostName 应保持 1.1.1.1，got %q", merged[0].HostName)
	}
	if len(result.Skipped) != 1 {
		t.Errorf("跳过结果不匹配: %v", result.Skipped)
	}
}

func TestMergeHostsRename(t *testing.T) {
	existing := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "1.1.1.1", User: "root"},
	}

	incoming := []HostEntry{
		{Alias: "myserver", Aliases: []string{"myserver"}, HostName: "2.2.2.2", User: "user"},
	}

	conflicts := map[string]ConflictInfo{
		"myserver": {Alias: "myserver", Action: ConflictRename, NewAlias: "myserver-new"},
	}

	merged, result := MergeHosts(existing, incoming, conflicts)

	if len(merged) != 2 {
		t.Fatalf("重命名后期望 2 个 Host，got %d", len(merged))
	}
	if merged[1].Alias != "myserver-new" {
		t.Errorf("重命名后别名应为 myserver-new，got %q", merged[1].Alias)
	}
	if merged[1].HostName != "2.2.2.2" {
		t.Errorf("重命名后 HostName 应为 2.2.2.2，got %q", merged[1].HostName)
	}
	if len(result.Renamed) != 1 {
		t.Errorf("重命名结果不匹配: %v", result.Renamed)
	}
	// 验证 Renames 字段包含结构化的旧/新别名信息
	if len(result.Renames) != 1 {
		t.Fatalf("Renames 期望 1 条，got %d", len(result.Renames))
	}
	if result.Renames[0].OldAlias != "myserver" {
		t.Errorf("Renames[0].OldAlias 期望 myserver，got %q", result.Renames[0].OldAlias)
	}
	if result.Renames[0].NewAlias != "myserver-new" {
		t.Errorf("Renames[0].NewAlias 期望 myserver-new，got %q", result.Renames[0].NewAlias)
	}
}

func TestMergeHostsMixed(t *testing.T) {
	existing := []HostEntry{
		{Alias: "server1", Aliases: []string{"server1"}, HostName: "1.1.1.1", User: "root"},
		{Alias: "server2", Aliases: []string{"server2"}, HostName: "2.2.2.2", User: "root"},
	}

	incoming := []HostEntry{
		{Alias: "server1", Aliases: []string{"server1"}, HostName: "10.0.0.1", User: "user"}, // 冲突，覆盖
		{Alias: "server2", Aliases: []string{"server2"}, HostName: "10.0.0.2", User: "user"}, // 冲突，跳过
		{Alias: "server3", Aliases: []string{"server3"}, HostName: "10.0.0.3", User: "user"}, // 无冲突
	}

	conflicts := map[string]ConflictInfo{
		"server1": {Alias: "server1", Action: ConflictOverwrite},
		"server2": {Alias: "server2", Action: ConflictSkip},
	}

	merged, result := MergeHosts(existing, incoming, conflicts)

	if len(merged) != 3 {
		t.Fatalf("混合合并后期望 3 个 Host，got %d", len(merged))
	}
	if len(result.Overwrite) != 1 {
		t.Errorf("覆盖数不匹配: %v", result.Overwrite)
	}
	if len(result.Skipped) != 1 {
		t.Errorf("跳过数不匹配: %v", result.Skipped)
	}
	if len(result.Imported) != 1 {
		t.Errorf("导入数不匹配: %v", result.Imported)
	}
}

func TestFormatConflictSummary(t *testing.T) {
	ci := ConflictInfo{
		Alias: "myserver",
		Existing: HostSummary{
			HostName: "1.1.1.1", User: "root", Port: "22", IdentityFile: "~/.ssh/keys/id_ed25519_a",
		},
		Incoming: HostSummary{
			HostName: "2.2.2.2", User: "ubuntu", Port: "2222", IdentityFile: "~/.ssh/keys/id_ed25519_b",
		},
	}

	summary := FormatConflictSummary(ci)

	if summary == "" {
		t.Fatal("格式化结果不应为空")
	}
	// 简单检查包含关键信息
	if !strings.Contains(summary, "myserver") {
		t.Error("应包含别名")
	}
	if !strings.Contains(summary, "1.1.1.1") {
		t.Error("应包含现有 HostName")
	}
	if !strings.Contains(summary, "2.2.2.2") {
		t.Error("应包含导入 HostName")
	}
}
