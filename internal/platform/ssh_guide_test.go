package platform

import (
	"strings"
	"testing"
)

func TestInstallGuide_windows_containsOptionalFeature(t *testing.T) {
	got := installOpenSSHGuideFor("windows")
	for _, kw := range []string{"可选功能", "OpenSSH", "ssh"} {
		if !strings.Contains(got, kw) {
			t.Errorf("windows guide missing %q:\n%s", kw, got)
		}
	}
}

func TestInstallGuide_darwin_containsBuiltin(t *testing.T) {
	got := installOpenSSHGuideFor("darwin")
	for _, kw := range []string{"内置", "ssh"} {
		if !strings.Contains(got, kw) {
			t.Errorf("darwin guide missing %q:\n%s", kw, got)
		}
	}
}

func TestInstallGuide_linux_mentionsOpensshClient(t *testing.T) {
	got := installOpenSSHGuideFor("linux")
	for _, kw := range []string{"openssh-client", "ssh"} {
		if !strings.Contains(got, kw) {
			t.Errorf("linux guide missing %q:\n%s", kw, got)
		}
	}
}