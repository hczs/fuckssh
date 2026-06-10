package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/config"
)

func Test_DeleteCmd_deletesEntry(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeTestConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n\nHost other\n    HostName 10.0.0.1\n    User admin\n")
	testConfigPath = cfg
	t.Cleanup(func() { testConfigPath = "" })

	// 注入确认函数：总是返回 true。
	origConfirm := confirmDeleteFn
	confirmDeleteFn = func(_ io.Writer, _, _, _ string) (bool, error) { return true, nil }
	t.Cleanup(func() { confirmDeleteFn = origConfirm })

	var stdout, stderr bytes.Buffer
	err := runDelete("myserver", &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDelete: %v", err)
	}

	assertContainsAny(t, stdout.String(), "success message", "已删除", "deleted")

	entries, err := config.ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Alias != "other" {
		t.Errorf("remaining alias = %q, want other", entries[0].Alias)
	}
}

func Test_DeleteCmd_aliasNotFound(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeTestConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n")
	testConfigPath = cfg
	t.Cleanup(func() { testConfigPath = "" })

	var stdout, stderr bytes.Buffer
	err := runDelete("nonexistent", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent alias")
	}
	if !errors.Is(err, config.ErrHostNotFound) {
		t.Errorf("error wraps ErrHostNotFound: %v", err)
	}
}

func Test_DeleteCmd_cancelledByUser(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeTestConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n")
	testConfigPath = cfg
	t.Cleanup(func() { testConfigPath = "" })

	// 注入确认函数：返回 false（用户取消）。
	origConfirm := confirmDeleteFn
	confirmDeleteFn = func(_ io.Writer, _, _, _ string) (bool, error) { return false, nil }
	t.Cleanup(func() { confirmDeleteFn = origConfirm })

	var stdout, stderr bytes.Buffer
	err := runDelete("myserver", &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDelete: %v", err)
	}

	assertContainsAny(t, stderr.String(), "cancelled message", "已取消", "Cancelled")

	// 文件应未被修改。
	entries, err := config.ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("len(entries) = %d, want 1 (should be unchanged)", len(entries))
	}
}

func Test_DeleteCmd_caseInsensitive(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeTestConfig(t, cfg, "Host MyServer\n    HostName 1.2.3.4\n    User root\n")
	testConfigPath = cfg
	t.Cleanup(func() { testConfigPath = "" })

	origConfirm := confirmDeleteFn
	confirmDeleteFn = func(_ io.Writer, _, _, _ string) (bool, error) { return true, nil }
	t.Cleanup(func() { confirmDeleteFn = origConfirm })

	var stdout, stderr bytes.Buffer
	err := runDelete("myserver", &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDelete: %v", err)
	}

	entries, err := config.ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}

func Test_DeleteCmd_exitCode1_forNotFound(t *testing.T) {
	testConfigPath = filepath.Join(t.TempDir(), "empty")
	writeTestConfig(t, testConfigPath, "")
	t.Cleanup(func() { testConfigPath = "" })

	err := ExecuteWithArgs([]string{"delete", "ghost"})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := ExitCode(err); got != 1 {
		t.Errorf("ExitCode = %d, want 1", got)
	}
	if !errors.Is(err, config.ErrHostNotFound) {
		t.Errorf("error wraps ErrHostNotFound: %v", err)
	}
}

func Test_DeleteCmd_rejectsNoArgs(t *testing.T) {
	testConfigPath = filepath.Join(t.TempDir(), "empty")
	writeTestConfig(t, testConfigPath, "")
	t.Cleanup(func() { testConfigPath = "" })

	err := ExecuteWithArgs([]string{"delete"})
	if err == nil {
		t.Fatal("expected error when no args passed to delete")
	}
}

func Test_DeleteCmd_withForceFlag(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeTestConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n")
	testConfigPath = cfg
	t.Cleanup(func() { testConfigPath = "" })

	// 不注入确认函数 — --force 应跳过确认。
	origForce := deleteForce
	deleteForce = true
	t.Cleanup(func() { deleteForce = origForce })

	var stdout, stderr bytes.Buffer
	err := runDelete("myserver", &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDelete with --force: %v", err)
	}

	entries, err := config.ParseFile(cfg)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}

func Test_DeleteCmd_removesManagedKey(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		t.Fatal(err)
	}

	cfg := filepath.Join(dir, "config")
	keyPath := filepath.Join(keysDir, "id_ed25519_fuckssh_myserver")
	writeTestConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n    IdentityFile "+keyPath+"\n")
	writeTestConfig(t, keyPath, "fake-private-key")
	writeTestConfig(t, keyPath+".pub", "fake-public-key")
	testConfigPath = cfg
	t.Cleanup(func() { testConfigPath = "" })

	origConfirm := confirmDeleteFn
	confirmDeleteFn = func(_ io.Writer, _, _, _ string) (bool, error) { return true, nil }
	t.Cleanup(func() { confirmDeleteFn = origConfirm })

	var stdout, stderr bytes.Buffer
	err := runDelete("myserver", &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDelete: %v", err)
	}

	// 密钥文件应被删除。
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		t.Errorf("managed key should be deleted, stat err: %v", err)
	}
	if _, err := os.Stat(keyPath + ".pub"); !os.IsNotExist(err) {
		t.Errorf("managed public key should be deleted, stat err: %v", err)
	}

	assertContainsAny(t, stdout.String(), "key removed message", "已删除关联密钥", "Removed managed key")
}

func Test_DeleteCmd_preservesExternalKey(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	keyPath := filepath.Join(dir, "id_ed25519")
	writeTestConfig(t, cfg, "Host myserver\n    HostName 1.2.3.4\n    User root\n    IdentityFile "+keyPath+"\n")
	writeTestConfig(t, keyPath, "fake-external-key")
	testConfigPath = cfg
	t.Cleanup(func() { testConfigPath = "" })

	origConfirm := confirmDeleteFn
	confirmDeleteFn = func(_ io.Writer, _, _, _ string) (bool, error) { return true, nil }
	t.Cleanup(func() { confirmDeleteFn = origConfirm })

	var stdout, stderr bytes.Buffer
	err := runDelete("myserver", &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDelete: %v", err)
	}

	// 外部密钥不应被删除。
	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("external key should be preserved, stat err: %v", err)
	}

	out := stdout.String()
	if strings.Contains(out, "已删除关联密钥") || strings.Contains(out, "Removed managed key") {
		t.Errorf("should not report key removal for external key:\n%s", out)
	}
}

func writeTestConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
