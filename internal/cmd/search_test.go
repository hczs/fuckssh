package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func Test_Search_matchesAliasViaCmd(t *testing.T) {
	configFileFlag = fixtureConfig("multiple.conf")
	t.Cleanup(func() { configFileFlag = "" })

	var buf bytes.Buffer
	if err := runSearchCmd([]string{"srv1"}, &buf, &buf); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	if !strings.Contains(buf.String(), "srv1") || strings.Contains(buf.String(), "srv2") {
		t.Fatalf("output should only contain srv1:\n%s", buf.String())
	}
}

func Test_Search_noMatch_returnsEmptyWithHint(t *testing.T) {
	configFileFlag = fixtureConfig("multiple.conf")
	t.Cleanup(func() { configFileFlag = "" })

	var buf bytes.Buffer
	if err := runSearchCmd([]string{"nomatch-xyz"}, &buf, &buf); err != nil {
		t.Fatalf("runSearchCmd: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "未找到匹配") {
		t.Errorf("output = %q, want no-match hint", out)
	}
}

func Test_Search_emptyQuery_returnsUsageError(t *testing.T) {
	err := runSearchCmd([]string{""}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected usage error for empty query")
	}
	if !strings.Contains(err.Error(), "非空") {
		t.Errorf("err = %v, want empty-query message", err)
	}
}
