package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fuckssh/fuckssh/internal/sshclient"
)

func Test_Search_matchesAliasViaCmd(t *testing.T) {
	configFileFlag = fixtureConfig("multiple.conf")
	t.Cleanup(func() { configFileFlag = "" })

	var stdout, stderr bytes.Buffer
	if err := runSearchCmd([]string{"srv1"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	if !strings.Contains(stderr.String(), "搜索") && !strings.Contains(stderr.String(), "Search") {
		t.Errorf("stderr = %q, want search meta", stderr.String())
	}
	if !strings.Contains(stdout.String(), "srv1") || strings.Contains(stdout.String(), "srv2") {
		t.Fatalf("stdout should only contain srv1:\n%s", stdout.String())
	}
}

func Test_Search_noMatch_returnsEmptyWithHint(t *testing.T) {
	configFileFlag = fixtureConfig("multiple.conf")
	t.Cleanup(func() { configFileFlag = "" })

	var stdout, stderr bytes.Buffer
	if err := runSearchCmd([]string{"nomatch-xyz"}, &stdout, &stderr); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "未找到匹配") && !strings.Contains(out, "No Host entries match") {
		t.Errorf("stdout = %q, want no-match hint", out)
	}
	if !strings.Contains(out, "fuckssh list") {
		t.Errorf("stdout = %q, want hint to list", out)
	}
}

func Test_Search_emptyQuery_returnsUsageError(t *testing.T) {
	err := runSearchCmd([]string{""}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected usage error for empty query")
	}
	if !strings.Contains(err.Error(), "非空") && !strings.Contains(err.Error(), "empty") {
		t.Errorf("err = %v, want empty-query message", err)
	}
}

func TestExitCode_sshNotFound(t *testing.T) {
	if got := ExitCode(sshclient.ErrSSHNotFound); got != 5 {
		t.Errorf("ssh not found = %d, want 5", got)
	}
}
