package wizard

import (
	"testing"

	"github.com/charmbracelet/huh"
)

func TestWizardTheme_blurredTitleDiffersFromFocused(t *testing.T) {
	tm := WizardTheme()
	if tm == nil {
		t.Fatal("WizardTheme returned nil")
		return
	}
	fgFocused := tm.Focused.Title.GetForeground()
	fgBlurred := tm.Blurred.Title.GetForeground()
	if fgFocused == fgBlurred {
		t.Fatal("Blurred.Title should differ from Focused.Title for stack contrast")
	}
}

func TestWizardTheme_blurredTextInputDiffersFromFocused(t *testing.T) {
	tm := WizardTheme()
	fgFocused := tm.Focused.TextInput.Text.GetForeground()
	fgBlurred := tm.Blurred.TextInput.Text.GetForeground()
	if fgFocused == fgBlurred {
		t.Fatal("Blurred.TextInput.Text should differ from Focused for filled rows")
	}
}

func TestWizardInputKeyMap_nextIncludesEnter(t *testing.T) {
	km := wizardInputKeyMap()
	keys := km.Next.Keys()
	if len(keys) != 2 || keys[0] != "down" || keys[1] != "enter" {
		t.Fatalf("Next keys = %v, want [down enter]", keys)
	}
}

func TestApplyCredentialNavPosition_submitAlwaysEnabled(t *testing.T) {
	km := wizardCredentialKeyMap()
	pos := huh.FieldPosition{Field: 5, LastField: 6, Group: 4, LastGroup: 5}
	applyCredentialNavPosition(&km, pos)
	if !km.Submit.Enabled() {
		t.Fatal("credential Submit must stay enabled for enter-to-test")
	}
}
