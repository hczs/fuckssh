package wizard

import "testing"

func TestEffectiveUser_defaultsToRoot(t *testing.T) {
	if got := effectiveUser(""); got != "root" {
		t.Errorf("effectiveUser(\"\") = %q, want root", got)
	}
	if got := effectiveUser("  "); got != "root" {
		t.Errorf("effectiveUser(\"  \") = %q, want root", got)
	}
	if got := effectiveUser(" ubuntu "); got != "ubuntu" {
		t.Errorf("effectiveUser = %q, want ubuntu", got)
	}
}

func TestUserField_commitEmptyUsesRoot(t *testing.T) {
	var user string
	f := NewUserField(nil).Value(&user)
	f.commit("")
	if user != "root" {
		t.Errorf("user = %q, want root", user)
	}
}
