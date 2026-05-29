package wizard

import (
	"errors"
	"testing"
)

func TestValidatePort(t *testing.T) {
	if err := validatePort("22"); err != nil {
		t.Errorf("22: %v", err)
	}
	if err := validatePort("65535"); err != nil {
		t.Errorf("65535: %v", err)
	}
	for _, bad := range []string{"0", "65536", "abc", "-1"} {
		if err := validatePort(bad); err == nil {
			t.Errorf("validatePort(%q): want error", bad)
		} else if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("validatePort(%q): %v", bad, err)
		}
	}
}

func TestPasswordMode_rejectsInvalidPort(t *testing.T) {
	_, err := finalizePasswordModeInput(PasswordModeInput{
		HostName: "1.2.3.4",
		User:     "root",
		Password: "pw",
		Port:     "99999",
	})
	if err == nil || !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}
