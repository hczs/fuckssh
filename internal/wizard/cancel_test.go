package wizard

import (
	"errors"
	"testing"

	"github.com/charmbracelet/huh"
)

func TestMapWizardAbort_userAborted(t *testing.T) {
	err := mapWizardAbort(huh.ErrUserAborted)
	if !IsCancelled(err) {
		t.Fatal("want cancelled")
	}
	if errors.Is(err, huh.ErrUserAborted) {
		// 仍可通过 Is 匹配原始中止，供退出码等使用
	} else {
		t.Fatal("want ErrUserAborted in chain")
	}
	if got := CancelMessage(err); got == "" || got == "user aborted" {
		t.Errorf("CancelMessage = %q, want localized cancel", got)
	}
}

func TestMapWizardAbort_retryFormUnchanged(t *testing.T) {
	err := mapWizardAbort(ErrWizardRetryForm)
	if !errors.Is(err, ErrWizardRetryForm) {
		t.Fatalf("err = %v, want ErrWizardRetryForm", err)
	}
	if IsCancelled(err) {
		t.Fatal("retry form should not be treated as cancel")
	}
}
