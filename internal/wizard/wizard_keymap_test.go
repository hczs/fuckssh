package wizard

import (
	"testing"

	"github.com/charmbracelet/huh"
)

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
