package wizard

import (
	"strings"
	"testing"
)

func TestSafeTTYString_windowsPath(t *testing.T) {
	got := safeTTYString(`C:\Users\hczs8\.ssh\config`)
	if strings.Contains(got, `\`) {
		t.Errorf("safeTTYString should use forward slashes, got %q", got)
	}
	if !strings.Contains(got, "Users/hczs8") {
		t.Errorf("path should remain readable, got %q", got)
	}
}

func TestBuildPasswordConfirmSummary_noMangledPath(t *testing.T) {
	s := buildPasswordConfirmSummary(PasswordModeInput{
		HostName: "10.12.2.220",
		User:     "boco",
		Port:     "22",
		Alias:    "10_12_2_220",
	}, `C:\Users\hczs8\.ssh\config`)

	if strings.Contains(s, "Host 220 →") {
		t.Errorf("summary should not use ambiguous Host arrow format: %q", s)
	}
	if !strings.Contains(s, "10_12_2_220") {
		t.Errorf("summary should include full alias: %q", s)
	}
	if !strings.Contains(s, "C:/Users/hczs8/.ssh/config") {
		t.Errorf("summary should include safe config path: %q", s)
	}
	if strings.Contains(s, "Usershczs8") {
		t.Errorf("summary should not contain mangled path: %q", s)
	}
	if !strings.Contains(s, "Host Key") && !strings.Contains(s, "Host key") {
		t.Errorf("summary should include host key notice: %q", s)
	}
}
