package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTempConfig 创建临时 config 文件并写入内容，返回路径。
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return path
}

func Test_EditHost_updatesHostName(t *testing.T) {
	config := `Host myserver
    HostName 1.2.3.4
    User root
    Port 22
    IdentityFile ~/.ssh/id_ed25519
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "myserver",
		HostName: "5.6.7.8",
		User:     "root",
		Port:     "22",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	entries, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].HostName != "5.6.7.8" {
		t.Errorf("HostName = %q, want %q", entries[0].HostName, "5.6.7.8")
	}
}

func Test_EditHost_updatesPort(t *testing.T) {
	config := `Host myserver
    HostName 1.2.3.4
    User root
    Port 22
    IdentityFile ~/.ssh/id_ed25519
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "myserver",
		HostName: "1.2.3.4",
		User:     "root",
		Port:     "2222",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if !strings.Contains(content, "Port 2222") {
		t.Errorf("expected Port 2222 in config:\n%s", content)
	}
}

func Test_EditHost_updatesRemark(t *testing.T) {
	config := `# 旧备注
Host myserver
    HostName 1.2.3.4
    User root
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "myserver",
		HostName: "1.2.3.4",
		User:     "root",
		Remark:   "新备注",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if strings.Contains(content, "旧备注") {
		t.Errorf("old remark should be removed:\n%s", content)
	}
	if !strings.Contains(content, "# 新备注") {
		t.Errorf("expected new remark in config:\n%s", content)
	}
}

func Test_EditHost_insertsRemarkWhenEmpty(t *testing.T) {
	config := `Host myserver
    HostName 1.2.3.4
    User root
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "myserver",
		HostName: "1.2.3.4",
		User:     "root",
		Remark:   "新增备注",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if !strings.Contains(content, "# 新增备注") {
		t.Errorf("expected remark in config:\n%s", content)
	}
}

func Test_EditHost_preservesUnknownDirectives(t *testing.T) {
	config := `Host myserver
    HostName 1.2.3.4
    User root
    Port 22
    ProxyJump jumphost
    ForwardAgent yes
    IdentityFile ~/.ssh/id_ed25519
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "myserver",
		HostName: "5.6.7.8",
		User:     "admin",
		Port:     "2222",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if !strings.Contains(content, "ProxyJump jumphost") {
		t.Errorf("ProxyJump should be preserved:\n%s", content)
	}
	if !strings.Contains(content, "ForwardAgent yes") {
		t.Errorf("ForwardAgent should be preserved:\n%s", content)
	}
	if !strings.Contains(content, "HostName 5.6.7.8") {
		t.Errorf("HostName should be updated:\n%s", content)
	}
	if !strings.Contains(content, "User admin") {
		t.Errorf("User should be updated:\n%s", content)
	}
	if !strings.Contains(content, "Port 2222") {
		t.Errorf("Port should be updated:\n%s", content)
	}
}

func Test_EditHost_renamesAlias(t *testing.T) {
	config := `Host myserver
    HostName 1.2.3.4
    User root
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "prod-server",
		HostName: "1.2.3.4",
		User:     "root",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	entries, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Alias != "prod-server" {
		t.Errorf("Alias = %q, want %q", entries[0].Alias, "prod-server")
	}
}

func Test_EditHost_aliasConflict(t *testing.T) {
	config := `Host server1
    HostName 1.2.3.4
    User root

Host server2
    HostName 5.6.7.8
    User admin
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "server2", // 与已有别名冲突
		HostName: "1.2.3.4",
		User:     "root",
	}
	err := EditHost(path, "server1", newEntry)
	if err == nil {
		t.Fatal("expected error for alias conflict, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func Test_EditHost_notFound(t *testing.T) {
	config := `Host myserver
    HostName 1.2.3.4
    User root
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "myserver",
		HostName: "5.6.7.8",
		User:     "root",
	}
	err := EditHost(path, "nonexistent", newEntry)
	if err == nil {
		t.Fatal("expected error for host not found, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func Test_EditHost_multipleHosts(t *testing.T) {
	config := `Host server1
    HostName 1.2.3.4
    User root

Host server2
    HostName 5.6.7.8
    User admin
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "server1",
		HostName: "10.0.0.1",
		User:     "root",
	}
	if err := EditHost(path, "server1", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	entries, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].HostName != "10.0.0.1" {
		t.Errorf("server1 HostName = %q, want %q", entries[0].HostName, "10.0.0.1")
	}
	if entries[1].HostName != "5.6.7.8" {
		t.Errorf("server2 HostName = %q, want %q", entries[1].HostName, "5.6.7.8")
	}
}

func Test_EditHost_insertsPortWhenMissing(t *testing.T) {
	// 原来没有 Port 行（默认 22），用户改成非 22 端口。
	config := `Host myserver
    HostName 1.2.3.4
    User root
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "myserver",
		HostName: "1.2.3.4",
		User:     "root",
		Port:     "2222",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if !strings.Contains(content, "Port 2222") {
		t.Errorf("expected Port 2222 in config:\n%s", content)
	}
}

func Test_EditHost_removesPortWhenSetToDefault(t *testing.T) {
	// 原来有 Port 2222，用户改成 22（默认值），应删除 Port 行。
	config := `Host myserver
    HostName 1.2.3.4
    User root
    Port 2222
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "myserver",
		HostName: "1.2.3.4",
		User:     "root",
		Port:     "22",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	raw, _ := os.ReadFile(path)
	content := string(raw)
	if strings.Contains(content, "Port") {
		t.Errorf("Port line should be removed for default value:\n%s", content)
	}
}

func Test_EditHost_caseInsensitiveAlias(t *testing.T) {
	config := `Host MyServer
    HostName 1.2.3.4
    User root
`
	path := writeTempConfig(t, config)

	newEntry := HostEntry{
		Alias:    "MyServer",
		HostName: "5.6.7.8",
		User:     "root",
	}
	if err := EditHost(path, "myserver", newEntry); err != nil {
		t.Fatalf("EditHost: %v", err)
	}

	entries, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if entries[0].HostName != "5.6.7.8" {
		t.Errorf("HostName = %q, want %q", entries[0].HostName, "5.6.7.8")
	}
}
