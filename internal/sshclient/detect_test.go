package sshclient

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckSSH_found(t *testing.T) {
	restore := stubLookPath(func(name string) (string, error) {
		if name != "ssh" {
			t.Fatalf("LookPath name = %q, want ssh", name)
		}
		return "/usr/bin/ssh", nil
	})
	defer restore()

	path, err := CheckSSH()
	if err != nil {
		t.Fatalf("CheckSSH: %v", err)
	}
	if path != "/usr/bin/ssh" {
		t.Errorf("path = %q, want /usr/bin/ssh", path)
	}
}

func TestCheckSSH_notFound_wrapsGuide(t *testing.T) {
	restore := stubLookPath(func(string) (string, error) {
		return "", errors.New("executable file not found in %PATH%")
	})
	defer restore()

	_, err := CheckSSH()
	if err == nil {
		t.Fatal("CheckSSH: want error")
	}
	if !errors.Is(err, ErrSSHNotFound) {
		t.Fatalf("errors.Is(err, ErrSSHNotFound) = false, err = %v", err)
	}
	msg := err.Error()
	for _, kw := range []string{"PATH", "ssh"} {
		if !strings.Contains(msg, kw) {
			t.Errorf("error missing %q: %s", kw, msg)
		}
	}
	// 应附带当前平台的安装指引（表驱动在 platform 包测文案；此处只断言非空）
	if len(msg) < 40 {
		t.Errorf("error too short, expected install guide: %s", msg)
	}
}

func stubLookPath(fn func(string) (string, error)) func() {
	prev := lookPath
	lookPath = fn
	return func() { lookPath = prev }
}
