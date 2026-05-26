package wizard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func TestAuthModeField_upGoesPrev(t *testing.T) {
	mode := ModePassword
	f := NewAuthModeField(&mode, nil)
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cmd == nil {
		t.Fatal("want PrevField on up")
	}
}

func TestAuthModeField_upDoesNotChangeMode(t *testing.T) {
	mode := ModeKey
	f := NewAuthModeField(&mode, nil)
	f.Update(tea.KeyMsg{Type: tea.KeyUp})
	if mode != ModeKey {
		t.Errorf("mode = %q, want key unchanged after up", mode)
	}
}

func TestAuthModeField_tabTogglesMode(t *testing.T) {
	mode := ModePassword
	f := NewAuthModeField(&mode, nil)
	f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if mode != ModeKey {
		t.Errorf("mode = %q, want key after tab", mode)
	}
	f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if mode != ModePassword {
		t.Errorf("mode = %q, want password after second tab", mode)
	}
}

func TestAuthModeField_leftRightDoNotToggleMode(t *testing.T) {
	mode := ModePassword
	f := NewAuthModeField(&mode, nil)
	f.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if mode != ModePassword {
		t.Errorf("mode = %q, want password unchanged after left", mode)
	}
	f.Update(tea.KeyMsg{Type: tea.KeyRight})
	if mode != ModePassword {
		t.Errorf("mode = %q, want password unchanged after right", mode)
	}
}

func TestAuthModeField_downGoesNext(t *testing.T) {
	mode := ModePassword
	f := NewAuthModeField(&mode, nil)
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd == nil {
		t.Fatal("want NextField on down")
	}
}

func TestAuthModeField_downDoesNotToggleMode(t *testing.T) {
	mode := ModePassword
	f := NewAuthModeField(&mode, nil)
	f.Update(tea.KeyMsg{Type: tea.KeyDown})
	// down 会 NextField，但调用前 mode 不应被切换；此处仅验证 Update 内未 toggle
	if mode != ModePassword {
		t.Errorf("mode = %q, should stay password before next field", mode)
	}
}

func TestAuthModeField_enterGoesNextWhenNotLastInForm(t *testing.T) {
	mode := ModePassword
	f := NewAuthModeField(&mode, nil)
	f.WithPosition(huh.FieldPosition{Field: 4, LastField: 6, Group: 4, LastGroup: 5})
	if !f.keymap.Next.Enabled() {
		t.Fatal("Next should be enabled when auth is not last")
	}
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("want NextField on enter when not last in form")
	}
}

func TestAuthModeField_enterGoesNext(t *testing.T) {
	mode := ModePassword
	f := NewAuthModeField(&mode, nil)
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("want NextField on enter")
	}
}

func TestAuthModeField_prevBindingIsUp(t *testing.T) {
	f := NewAuthModeField(new(ConnectionMode), nil)
	if keys := f.keymap.Prev.Keys(); len(keys) == 0 || keys[0] != "up" {
		t.Fatalf("prev keys = %v, want up", f.keymap.Prev.Keys())
	}
}

func TestAuthModeField_tabBindingOnLeftKey(t *testing.T) {
	f := NewAuthModeField(new(ConnectionMode), nil)
	if keys := f.keymap.Left.Keys(); len(keys) == 0 || keys[0] != "tab" {
		t.Fatalf("left keys = %v, want tab for switch", f.keymap.Left.Keys())
	}
}
