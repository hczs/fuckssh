package wizard

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fuckssh/fuckssh/internal/i18n"
)

func TestHostField_rejectsEmptyInline(t *testing.T) {
	host := ""
	f := NewHostField(nil).Value(&host)
	f.textinput.SetValue("")
	model, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
	field := model.(*hostField)
	if cmd != nil {
		t.Fatal("want no advance on empty host")
	}
	if field.inlineMsg == "" {
		t.Fatal("want inline empty message")
	}
	if !strings.Contains(field.inlineMsg, i18n.T(i18n.KeyWizardErrEmpty)) {
		t.Fatalf("inlineMsg = %q", field.inlineMsg)
	}
}

func TestHostField_acceptsValidHost(t *testing.T) {
	host := ""
	f := NewHostField(nil).Value(&host)
	f.textinput.SetValue("203.0.113.1")
	model, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
	field := model.(*hostField)
	if cmd == nil {
		t.Fatal("want NextField on valid host")
	}
	if field.inlineMsg != "" {
		t.Fatalf("unexpected inline error: %q", field.inlineMsg)
	}
	if host != "203.0.113.1" {
		t.Errorf("host = %q", host)
	}
}

func TestHostField_upGoesPrev(t *testing.T) {
	host := ""
	f := NewHostField(nil).Value(&host)
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cmd == nil {
		t.Fatal("want PrevField on up")
	}
}

func TestHostField_downGoesNextWhenValid(t *testing.T) {
	host := ""
	f := NewHostField(nil).Value(&host)
	f.textinput.SetValue("203.0.113.1")
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd == nil {
		t.Fatal("want NextField on down when host valid")
	}
}

func TestHostField_tabDoesNotAdvance(t *testing.T) {
	host := ""
	f := NewHostField(nil).Value(&host)
	f.textinput.SetValue("203.0.113.1")
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Fatal("tab should not advance to next field")
	}
}
