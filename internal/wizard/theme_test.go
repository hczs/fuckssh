package wizard

import "testing"

func TestWizardTheme_blurredTitleDiffersFromFocused(t *testing.T) {
	tm := WizardTheme()
	if tm == nil {
		t.Fatal("WizardTheme returned nil")
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
